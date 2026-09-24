package dialog

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func TestWakeupToastCancelKeys(t *testing.T) {
	t.Parallel()

	toast := NewWakeupToast(nil, "session-id", "", time.Time{})

	testCases := map[string]tea.KeyPressMsg{
		"enter":  {Code: tea.KeyEnter},
		"space":  {Code: tea.KeySpace, Text: " "},
		"escape": {Code: tea.KeyEscape},
		"c":      {Code: 'c', Text: "c"},
	}
	for name, msg := range testCases {
		t.Run(name, func(t *testing.T) {
			action, ok := toast.HandleMsg(msg).(ActionCancelWakeup)
			require.True(t, ok)
			require.Equal(t, "session-id", action.SessionID)
		})
	}
}
