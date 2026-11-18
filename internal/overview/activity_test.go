package overview

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/crush/internal/ports"
	"github.com/stretchr/testify/require"
)

func TestScheduleStatsReadsTheDurableStore(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	require.Equal(t, 0, mustCount(t, dataDir))

	store := filepath.Join(dataDir, "wakeups", "schedules.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(store), 0o755))
	require.NoError(t, os.WriteFile(store, []byte(`{
	  "version": 1,
	  "schedules": [
	    {"id": "timed", "next_fire_unix": 4102444800},
	    {"id": "event-only", "event": "pr-review"},
	    {"id": "sooner", "next_fire_unix": 4102444700}
	  ]
	}`), 0o644))

	count, next := scheduleStats(dataDir)
	require.Equal(t, 3, count)
	require.NotNil(t, next)
	require.Equal(t, time.Unix(4102444700, 0).UTC(), *next)
}

func TestScheduleStatsIgnoresBrokenStore(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	store := filepath.Join(dataDir, "wakeups", "schedules.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(store), 0o755))
	require.NoError(t, os.WriteFile(store, []byte("{not json"), 0o644))

	count, next := scheduleStats(dataDir)
	require.Equal(t, 0, count)
	require.Nil(t, next)
}

func TestSessionStatsWithoutADatabase(t *testing.T) {
	t.Parallel()

	count, last := sessionStats(t.TempDir())
	require.Equal(t, 0, count)
	require.Nil(t, last)
}

func TestPreviewsOfAddressesListeners(t *testing.T) {
	t.Parallel()

	previews := previewsOf([]ports.Listener{
		{Port: 3000, Address: "0.0.0.0", Process: "node"},
		{Port: 8080, Address: "127.0.0.1", Process: "python3"},
		{Port: 9090, Address: "::"},
	})
	require.Len(t, previews, 3)
	// Sorted by port, and a wildcard bind is reached on loopback.
	require.Equal(t, 3000, previews[0].Port)
	require.Equal(t, "http://127.0.0.1:3000", previews[0].URL)
	require.Equal(t, "http://127.0.0.1:8080", previews[1].URL)
	require.Equal(t, "http://127.0.0.1:9090", previews[2].URL)
}

func mustCount(t *testing.T, dataDir string) int {
	t.Helper()
	count, _ := scheduleStats(dataDir)
	return count
}
