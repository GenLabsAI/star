package search

import (
	"context"
	"testing"
	"time"

	"github.com/charmbracelet/crush/internal/db"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/stretchr/testify/require"
)

func TestIndexAndSearchAcrossSessions(t *testing.T) {
	t.Cleanup(db.ResetPool)

	dataDir := t.TempDir()
	conn, err := db.Connect(context.Background(), dataDir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Release(dataDir) })
	require.NoError(t, db.MigrationWarning(dataDir))

	ctx := context.Background()
	now := time.Now().Unix()
	_, err = conn.ExecContext(ctx, `
		INSERT INTO sessions (
			id, title, message_count, prompt_tokens, completion_tokens,
			cost, updated_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, "s1", "Migration work", 0, 0, 0, 0, now, now)
	require.NoError(t, err)

	svc := NewService(conn, nil)
	msg := message.Message{
		ID:        "m1",
		SessionID: "s1",
		Role:      message.User,
		Parts: []message.ContentPart{
			message.TextContent{Text: "The migration happy path and search index are working."},
		},
	}

	require.NoError(t, svc.Index(ctx, msg))
	require.NoError(t, svc.Index(ctx, msg), "reindexing the same message must be idempotent")

	var count int
	require.NoError(t, conn.QueryRowContext(ctx,
		`SELECT count(*) FROM messages_fts WHERE message_id = ?`, msg.ID,
	).Scan(&count))
	require.Equal(t, 1, count)

	results, err := svc.Search(ctx, "migration", SearchOpts{})
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, msg.ID, results[0].MessageID)
	require.Equal(t, "s1", results[0].SessionID)
	require.Equal(t, "Migration work", results[0].SessionTitle)
	require.Contains(t, results[0].Snippet, "migration")

	results, err = svc.Search(ctx, "happy", SearchOpts{SessionID: "s2"})
	require.NoError(t, err)
	require.Empty(t, results, "session filtering must exclude other sessions")

	require.NoError(t, svc.Remove(ctx, msg.ID))
	results, err = svc.Search(ctx, "migration", SearchOpts{})
	require.NoError(t, err)
	require.Empty(t, results)
}

// TestSearchAcceptsQuotesAndBrackets guards that a query is searched as
// literal text rather than FTS5 syntax. Quotes, brackets, parentheses,
// asterisks and boolean keywords used to make SQLite reject the whole MATCH,
// which the search dialog surfaced as "Search unavailable".
func TestSearchAcceptsQuotesAndBrackets(t *testing.T) {
	t.Cleanup(db.ResetPool)

	dataDir := t.TempDir()
	conn, err := db.Connect(context.Background(), dataDir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Release(dataDir) })

	ctx := context.Background()
	now := time.Now().Unix()
	_, err = conn.ExecContext(ctx, `
		INSERT INTO sessions (
			id, title, message_count, prompt_tokens, completion_tokens,
			cost, updated_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, "s1", "Migration work", 0, 0, 0, 0, now, now)
	require.NoError(t, err)

	svc := NewService(conn, nil)
	msg := message.Message{
		ID:        "m1",
		SessionID: "s1",
		Role:      message.User,
		Parts: []message.ContentPart{
			message.TextContent{Text: `The migration "happy path" handles [brackets] and (parens) too.`},
		},
	}
	require.NoError(t, svc.Index(ctx, msg))

	for _, query := range []string{
		`[brackets]`,
		`migration "happy`,
		`"happy path"`,
		`(parens)`,
		`migration AND`,
		`*`,
		`foo"bar`,
		`NEAR(migration brackets)`,
	} {
		_, err := svc.Search(ctx, query, SearchOpts{})
		require.NoError(t, err, "query %q must not fail", query)
	}

	results, err := svc.Search(ctx, `[brackets]`, SearchOpts{})
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, msg.ID, results[0].MessageID)

	results, err = svc.Search(ctx, `MIGRATION`, SearchOpts{})
	require.NoError(t, err)
	require.Len(t, results, 1, "queries stay case-insensitive")

	results, err = svc.Search(ctx, `[]`, SearchOpts{})
	require.NoError(t, err)
	require.Empty(t, results, "a query with no searchable characters matches nothing")
}
