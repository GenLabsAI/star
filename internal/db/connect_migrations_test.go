package db

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

// TestConnect_AppliesMigrationsBehindCurrentVersion reproduces a database
// migrated by an older build that had already moved past a later migration
// (20260920000000_add_messages_fts, since removed from the tree) before the
// Prism and role-index migrations were added here. Goose reports those two as
// missing and refuses to start unless out-of-order migrations are allowed.
func TestConnect_AppliesMigrationsBehindCurrentVersion(t *testing.T) {
	t.Cleanup(ResetPool)

	dataDir := t.TempDir()
	dbPath := filepath.Join(dataDir, "crush.db")

	raw, err := openDB(dbPath)
	require.NoError(t, err)
	require.NoError(t, initGoose())
	require.NoError(t, goose.UpTo(raw, "migrations", 20260127000000))

	// The version an older build recorded for a migration file that no longer
	// exists. It puts the database ahead of the migrations that follow.
	_, err = raw.Exec(
		`INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, ?)`,
		int64(20260920000000), true,
	)
	require.NoError(t, err)
	require.NoError(t, raw.Close())

	conn, err := Connect(context.Background(), dataDir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = Release(dataDir) })
	require.NoError(t, MigrationWarning(dataDir))

	ctx := context.Background()

	var prismColumns int
	require.NoError(t, conn.QueryRowContext(ctx,
		`SELECT count(*) FROM pragma_table_info('messages')
		 WHERE name IN ('prism_model_id', 'prism_model_name',
		                'prism_hypercredit_savings', 'prism_dollar_savings')`,
	).Scan(&prismColumns))
	require.Equal(t, 4, prismColumns, "the Prism migration must be applied")

	var roleIndex int
	require.NoError(t, conn.QueryRowContext(ctx,
		`SELECT count(*) FROM sqlite_master
		 WHERE type = 'index' AND name = 'idx_messages_role_created_at'`,
	).Scan(&roleIndex))
	require.Equal(t, 1, roleIndex, "the role index migration must be applied")

	var channelColumn int
	require.NoError(t, conn.QueryRowContext(ctx,
		`SELECT count(*) FROM pragma_table_info('sessions') WHERE name = 'channel'`,
	).Scan(&channelColumn))
	require.Equal(t, 1, channelColumn, "migrations after the gap must still apply")

	var mcpTables int
	require.NoError(t, conn.QueryRowContext(ctx,
		`SELECT count(*) FROM sqlite_master
		 WHERE type = 'table' AND name IN ('mcp_disabled_servers', 'mcp_enabled_servers')`,
	).Scan(&mcpTables))
	require.Equal(t, 2, mcpTables, "migrations after the gap must still apply")
}

// TestConnect_StartsWithExistingSchemaWhenMigrationsFail ensures a broken
// migration cannot stop the app from starting. A duplicate column makes the
// next migration fail; Connect must still return a usable connection and
// report the problem through MigrationWarning.
func TestConnect_StartsWithExistingSchemaWhenMigrationsFail(t *testing.T) {
	t.Cleanup(ResetPool)

	dataDir := t.TempDir()
	dbPath := filepath.Join(dataDir, "crush.db")

	raw, err := openDB(dbPath)
	require.NoError(t, err)
	require.NoError(t, initGoose())
	require.NoError(t, goose.UpTo(raw, "migrations", 20260127000000))

	// The Prism migration adds this column; pre-adding it makes that
	// migration fail with a duplicate column error.
	_, err = raw.Exec(`ALTER TABLE messages ADD COLUMN prism_model_id TEXT`)
	require.NoError(t, err)
	require.NoError(t, raw.Close())

	conn, err := Connect(context.Background(), dataDir)
	require.NoError(t, err, "a failed migration must not prevent startup")
	t.Cleanup(func() { _ = Release(dataDir) })

	require.Error(t, MigrationWarning(dataDir), "the failed migration should be reported")
	require.NoError(t, conn.PingContext(context.Background()))

	// The pre-existing schema is still usable.
	var sessions int
	require.NoError(t, conn.QueryRowContext(context.Background(),
		`SELECT count(*) FROM sessions`,
	).Scan(&sessions))
}
