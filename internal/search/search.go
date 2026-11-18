package search

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"unicode"

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

// matchQuery converts a raw user query into an FTS5 MATCH expression.
//
// FTS5 reads characters such as quotes, brackets, parentheses, asterisks
// and colons, and keywords such as AND, OR and NOT, as query syntax. Passing
// the raw query through makes anything containing them fail outright with a
// syntax error, so every whitespace-separated term is quoted as literal text
// (embedded quotes are doubled, FTS5's escape) and the terms are joined with
// AND so adding words narrows the results.
//
// Terms without any letters or numbers carry no searchable tokens and are
// dropped; a query made only of punctuation searches for nothing.
func matchQuery(query string) string {
	terms := make([]string, 0, len(query))
	for _, term := range strings.Fields(query) {
		if !hasSearchableRune(term) {
			continue
		}
		terms = append(terms, `"`+strings.ReplaceAll(term, `"`, `""`)+`"`)
	}
	return strings.Join(terms, " AND ")
}

// hasSearchableRune reports whether s contains a rune the FTS5 tokenizer
// keeps, i.e. a letter or a number.
func hasSearchableRune(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return true
		}
	}
	return false
}

func (s *service) Search(ctx context.Context, query string, opts SearchOpts) ([]Result, error) {
	match := matchQuery(query)
	if match == "" {
		return nil, nil
	}
	if opts.Limit <= 0 {
		opts.Limit = 50
	}
	if opts.Limit > 200 {
		opts.Limit = 200
	}
	if opts.Offset < 0 {
		opts.Offset = 0
	}

	var b strings.Builder
	b.WriteString(`
		SELECT
			mf.session_id,
			(SELECT title FROM sessions WHERE id = mf.session_id) AS session_title,
			mf.message_id,
			mf.role,
			snippet(messages_fts, 0, '{{', '}}', '…', 48) AS snippet,
			bm25(messages_fts) AS rank
		FROM messages_fts mf
		WHERE messages_fts MATCH ?`)
	args := []any{match}

	if opts.SessionID != "" {
		b.WriteString(" AND mf.session_id = ?")
		args = append(args, opts.SessionID)
	}
	if len(opts.Roles) > 0 {
		b.WriteString(" AND mf.role IN (")
		for i, role := range opts.Roles {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteByte('?')
			args = append(args, role)
		}
		b.WriteByte(')')
	}
	b.WriteString(" ORDER BY rank LIMIT ? OFFSET ?")
	args = append(args, opts.Limit, opts.Offset)

	rows, err := s.db.QueryContext(ctx, b.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("search query: %w", err)
	}
	defer rows.Close()

	var results []Result
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
	text := strings.TrimSpace(ExtractSearchableText(msg.Parts))

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin search index transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.ExecContext(ctx, "DELETE FROM messages_fts WHERE message_id = ?", msg.ID); err != nil {
		return fmt.Errorf("remove old search index entry: %w", err)
	}
	if text == "" {
		return tx.Commit()
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO messages_fts (message_id, session_id, role, content)
		VALUES (?, ?, ?, ?)
	`, msg.ID, msg.SessionID, string(msg.Role), text); err != nil {
		return fmt.Errorf("insert search index entry: %w", err)
	}

	return tx.Commit()
}

func (s *service) Remove(ctx context.Context, messageID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM messages_fts WHERE message_id = ?", messageID)
	return err
}

func (s *service) NeedsReindex(ctx context.Context) (bool, error) {
	var msgCount, ftsCount int

	if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM messages").Scan(&msgCount); err != nil {
		return false, fmt.Errorf("count messages: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM messages_fts").Scan(&ftsCount); err != nil {
		return false, fmt.Errorf("count messages_fts: %w", err)
	}

	return msgCount > 0 && ftsCount == 0, nil
}

func (s *service) Reindex(ctx context.Context, progress func(done, total int)) error {
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM messages").Scan(&total); err != nil {
		return fmt.Errorf("count messages: %w", err)
	}

	msgs, err := s.messages.ListAllMessages(ctx)
	if err != nil {
		return fmt.Errorf("list all messages: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, "DELETE FROM messages_fts"); err != nil {
		return fmt.Errorf("clear search index: %w", err)
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
