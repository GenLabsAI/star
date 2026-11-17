package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNextCronOccurrence(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, 11, 17, 10, 7, 30, 0, time.UTC)

	t.Run("every thirty minutes", func(t *testing.T) {
		t.Parallel()
		next, err := NextCronOccurrence("*/30 * * * *", from)
		require.NoError(t, err)
		require.Equal(t, time.Date(2026, 11, 17, 10, 30, 0, 0, time.UTC), next)
	})

	t.Run("daily at nine", func(t *testing.T) {
		t.Parallel()
		next, err := NextCronOccurrence("0 9 * * *", from)
		require.NoError(t, err)
		// 09:00 already passed today, so tomorrow.
		require.Equal(t, time.Date(2026, 11, 18, 9, 0, 0, 0, time.UTC), next)
	})

	t.Run("later today", func(t *testing.T) {
		t.Parallel()
		next, err := NextCronOccurrence("0 11 * * *", from)
		require.NoError(t, err)
		require.Equal(t, time.Date(2026, 11, 17, 11, 0, 0, 0, time.UTC), next)
	})

	t.Run("weekday names", func(t *testing.T) {
		t.Parallel()
		// 2026-11-17 is a Tuesday.
		next, err := NextCronOccurrence("0 9 * * mon", from)
		require.NoError(t, err)
		require.Equal(t, time.Monday, next.Weekday())
		require.Equal(t, 9, next.Hour())
	})

	t.Run("every duration", func(t *testing.T) {
		t.Parallel()
		next, err := NextCronOccurrence("@every 30m", from)
		require.NoError(t, err)
		require.Equal(t, from.Add(30*time.Minute), next)
	})

	t.Run("hourly and daily shorthands", func(t *testing.T) {
		t.Parallel()
		next, err := NextCronOccurrence("@hourly", from)
		require.NoError(t, err)
		require.Equal(t, time.Date(2026, 11, 17, 11, 0, 0, 0, time.UTC), next)

		next, err = NextCronOccurrence("@daily", from)
		require.NoError(t, err)
		require.Equal(t, time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC), next)
	})

	t.Run("invalid specs", func(t *testing.T) {
		t.Parallel()
		for _, spec := range []string{"", "0 9 * *", "61 * * * *", "*/0 * * * *", "9-5 * * * *", "@every banana", "0 9 * * someday"} {
			_, err := NextCronOccurrence(spec, from)
			require.Error(t, err, "spec %q", spec)
		}
	})
}

func TestNextDailyAt(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, 11, 17, 10, 7, 30, 0, time.UTC)

	next, err := NextDailyAt("08:00", from)
	require.NoError(t, err)
	require.Equal(t, time.Date(2026, 11, 18, 8, 0, 0, 0, time.UTC), next)

	next, err = NextDailyAt("12:30", from)
	require.NoError(t, err)
	require.Equal(t, time.Date(2026, 11, 17, 12, 30, 0, 0, time.UTC), next)

	for _, bad := range []string{"8", "25:00", "08:60", "morning"} {
		_, err := NextDailyAt(bad, from)
		require.Error(t, err, "dailyAt %q", bad)
	}

	cron, err := DailyAtCron("08:00")
	require.NoError(t, err)
	require.Equal(t, "0 8 * * *", cron)
	next, err = NextCronOccurrence(cron, from)
	require.NoError(t, err)
	require.Equal(t, time.Date(2026, 11, 18, 8, 0, 0, 0, time.UTC), next)
}
