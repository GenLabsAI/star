package search

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/charmbracelet/crush/internal/message"
)

// Result represents a search match.
type Result struct {
	SessionID    string
	SessionTitle string
	MessageID    string
	Role         string
	Snippet      string
	Rank         float64
	CreatedAt    int64
}

// SearchOpts contains filtering and pagination options for a search.
type SearchOpts struct {
	SessionID string
	Roles     []string
	Limit     int
	Offset    int
}

// Service provides cross-session message content search.
type Service interface {
	Search(ctx context.Context, query string, opts SearchOpts) ([]Result, error)
	Index(ctx context.Context, msg message.Message) error
	Remove(ctx context.Context, messageID string) error
	Reindex(ctx context.Context, progress func(done, total int)) error
	NeedsReindex(ctx context.Context) (bool, error)
}

type service struct {
	db       *sql.DB
	messages message.Service
}

// NewService creates a new search service.
func NewService(db *sql.DB, messages message.Service) Service {
	return &service{
		db:       db,
		messages: messages,
	}
}

func (s *service) Search(ctx context.Context, query string, opts SearchOpts) ([]Result, error) {
	// Limit default
	if opts.Limit <= 0 {
		opts.Limit = 50
	}

	var results []Result
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			mf.session_id,
			(SELECT title FROM sessions WHERE id = mf.session_id) as session_title,
			mf.message_id,
			mf.role,
			snippet(messages_fts, 0, '{{', '}}', '…', 48) AS snippet,
			bm25(messages_fts) as rank
		FROM messages_fts mf
		WHERE messages_fts MATCH ?
		ORDER BY rank
		LIMIT ? OFFSET ?;
	`, query, opts.Limit, opts.Offset)
	if err != nil {
		return nil, fmt.Errorf("search query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r Result
		var title sql.NullString
		if err := rows.Scan(&r.SessionID, &title, &r.MessageID, &r.Role, &r.Snippet, &r.Rank); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		r.SessionTitle = title.String
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return results, nil
}

func (s *service) Index(ctx context.Context, msg message.Message) error {
	text := ExtractSearchableText(msg.Parts)
	if text == "" {
		// If there's no searchable text, just remove any existing entry to keep it clean
		return s.Remove(ctx, msg.ID)
	}

	// Use an UPSERT-like pattern (SQLite INSERT OR REPLACE)
	_, err := s.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO messages_fts (message_id, session_id, role, content)
		VALUES (?, ?, ?, ?)
	`, msg.ID, msg.SessionID, string(msg.Role), text)
	return err
}

func (s *service) Remove(ctx context.Context, messageID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM messages_fts WHERE message_id = ?", messageID)
	return err
}

func (s *service) NeedsReindex(ctx context.Context) (bool, error) {
	var msgCount, ftsCount int
	
	err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM messages").Scan(&msgCount)
	if err != nil {
		return false, fmt.Errorf("count messages: %w", err)
	}

	err = s.db.QueryRowContext(ctx, "SELECT count(*) FROM messages_fts").Scan(&ftsCount)
	if err != nil {
		return false, fmt.Errorf("count messages_fts: %w", err)
	}

	return msgCount > 0 && ftsCount == 0, nil
}

func (s *service) Reindex(ctx context.Context, progress func(done, total int)) error {
	var total int
	err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM messages").Scan(&total)
	if err != nil {
		return fmt.Errorf("count messages: %w", err)
	}

	msgs, err := s.messages.ListAllMessages(ctx)
	if err != nil {
		return fmt.Errorf("list all messages: %w", err)
	}

	done := 0
	for _, msg := range msgs {
		if err := s.Index(ctx, msg); err != nil {
			return fmt.Errorf("index message %s: %w", msg.ID, err)
		}
		done++
		if progress != nil && done%10 == 0 {
			progress(done, total)
		}
	}
	if progress != nil {
		progress(total, total)
	}

	return nil
}
