package permission_test

import (
	"sync"
	"testing"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/hooks"
	"github.com/charmbracelet/crush/internal/permission"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissionService_PermissionRequestHook(t *testing.T) {
	t.Parallel()

	t.Run("hook allow auto-approves", func(t *testing.T) {
		t.Parallel()
		hookRunner := hooks.NewRunner(hooks.Static(map[string][]config.HookConfig{"PermissionRequest": []config.HookConfig{{Command: `echo '{"decision":"allow"}'`}}}), "/tmp", "/tmp")
		service := permission.NewPermissionService("/tmp", false, nil, hookRunner)

		granted, err := service.Request(t.Context(), permission.CreatePermissionRequest{
			SessionID:  "s1",
			ToolCallID: "call-1",
			ToolName:   "bash",
			Action:     "execute",
			Path:       "/tmp",
		})
		require.NoError(t, err)
		assert.True(t, granted, "hook-approved call should return granted")
	})

	t.Run("hook deny auto-denies", func(t *testing.T) {
		t.Parallel()
		hookRunner := hooks.NewRunner(hooks.Static(map[string][]config.HookConfig{"PermissionRequest": []config.HookConfig{{Command: `echo '{"decision":"deny"}'`}}}), "/tmp", "/tmp")
		service := permission.NewPermissionService("/tmp", false, nil, hookRunner)

		granted, err := service.Request(t.Context(), permission.CreatePermissionRequest{
			SessionID:  "s1",
			ToolCallID: "call-2",
			ToolName:   "bash",
			Action:     "execute",
			Path:       "/tmp",
		})
		require.NoError(t, err)
		assert.False(t, granted, "hook-denied call should return denied")
	})

	t.Run("hook none falls back to normal prompt", func(t *testing.T) {
		t.Parallel()
		hookRunner := hooks.NewRunner(hooks.Static(map[string][]config.HookConfig{"PermissionRequest": []config.HookConfig{{Command: `echo '{}'`}}}), "/tmp", "/tmp")
		service := permission.NewPermissionService("/tmp", false, nil, hookRunner)

		// Because it falls back, we need a subscriber to resolve it
		events := service.Subscribe(t.Context())
		var (
			wg      sync.WaitGroup
			granted bool
			err     error
		)
		wg.Go(func() {
			granted, err = service.Request(t.Context(), permission.CreatePermissionRequest{
				SessionID:  "s1",
				ToolCallID: "call-3",
				ToolName:   "bash",
				Action:     "execute",
				Path:       "/tmp",
			})
		})

		event := <-events
		service.Grant(event.Payload)
		wg.Wait()
		require.NoError(t, err)
		assert.True(t, granted)
	})
}
