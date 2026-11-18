package chat

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/ui/styles"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

// TestShellGroupItem_CollapsedRender verifies that the shell group shows
// "✓ Bash N commands" when collapsed with successful commands.
func TestShellGroupItem_CollapsedRender(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()

	commands := []ShellCommand{
		{Command: "ls -la", Output: "total 0", ExitCode: 0},
		{Command: "pwd", Output: "/home/user", ExitCode: 0},
		{Command: "echo hello", Output: "hello", ExitCode: 0},
	}

	group := NewShellGroupItem(&sty, commands).(*ShellGroupItem)
	require.Equal(t, 3, len(group.commands))
	require.False(t, group.expanded)

	rendered := ansi.Strip(group.RawRender(80))
	require.Contains(t, rendered, "✓ Bash 3 commands")
	require.Contains(t, rendered, "click to expand")
}

// TestShellGroupItem_ExpandedRender verifies that the shell group shows
// all commands when expanded.
func TestShellGroupItem_ExpandedRender(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()

	commands := []ShellCommand{
		{Command: "ls -la", Output: "total 0", ExitCode: 0},
		{Command: "pwd", Output: "/home/user", ExitCode: 0},
	}

	group := NewShellGroupItem(&sty, commands).(*ShellGroupItem)
	group.expanded = true

	rendered := ansi.Strip(group.RawRender(80))
	require.Contains(t, rendered, "$ ls -la")
	require.Contains(t, rendered, "total 0")
	require.Contains(t, rendered, "$ pwd")
	require.Contains(t, rendered, "/home/user")
}

// TestShellGroupItem_SingleCommandCollapsed verifies that a single command
// shows "$ 1 command" format.
func TestShellGroupItem_SingleCommandCollapsed(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()

	group := NewShellGroupItem(&sty, []ShellCommand{
		{Command: "ls -la", Output: "total 0", ExitCode: 0},
	}).(*ShellGroupItem)

	rendered := ansi.Strip(group.RawRender(80))
	require.Contains(t, rendered, "$ 1 command")
	require.NotContains(t, rendered, "Bash")
}

// TestShellGroupItem_ExpandToggle verifies that toggling expands and
// collapses the group.
func TestShellGroupItem_ExpandToggle(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()

	group := NewShellGroupItem(&sty, []ShellCommand{
		{Command: "ls", Output: "file1", ExitCode: 0},
		{Command: "pwd", Output: "/", ExitCode: 0},
	}).(*ShellGroupItem)
	require.False(t, group.expanded)

	require.True(t, group.ToggleExpanded())
	require.True(t, group.expanded)

	require.False(t, group.ToggleExpanded())
	require.False(t, group.expanded)
}

// TestShellGroupItem_CopyCommand verifies that copy key copies all commands.
func TestShellGroupItem_CopyCommand(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()

	group := NewShellGroupItem(&sty, []ShellCommand{
		{Command: "ls", Output: "file1", ExitCode: 0},
		{Command: "pwd", Output: "/", ExitCode: 0},
	}).(*ShellGroupItem)

	handled, _ := group.HandleKeyEvent(tea.KeyPressMsg{Code: 'c', Text: "c"})
	require.True(t, handled)
}

// TestShellGroupItem_ExitCodeDisplay verifies that failed commands show
// their exit code.
func TestShellGroupItem_ExitCodeDisplay(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()

	group := NewShellGroupItem(&sty, []ShellCommand{
		{Command: "false", Output: "", ExitCode: 1},
		{Command: "true", Output: "", ExitCode: 0},
	}).(*ShellGroupItem)
	group.expanded = true

	rendered := ansi.Strip(group.RawRender(80))
	require.Contains(t, rendered, "(exit 1)")
	require.NotContains(t, rendered, "(exit 0)")
}

// TestShellGroupItem_EmptyCommands verifies behavior with no commands.
func TestShellGroupItem_EmptyCommands(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	group := NewShellGroupItem(&sty, nil).(*ShellGroupItem)

	require.Equal(t, "", group.RawRender(80))
}

// TestShellGroupItem_Pluralization verifies correct pluralization.
func TestShellGroupItem_Pluralization(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()

	single := NewShellGroupItem(&sty, []ShellCommand{{Command: "ls"}}).(*ShellGroupItem)
	require.Contains(t, ansi.Strip(single.RawRender(80)), "$ 1 command")

	multiple := NewShellGroupItem(&sty, []ShellCommand{
		{Command: "ls"},
		{Command: "pwd"},
	}).(*ShellGroupItem)
	require.Contains(t, ansi.Strip(multiple.RawRender(80)), "✓ Bash 2 commands")
}

// TestShellGroupItem_FilterValue verifies FilterValue returns first command.
func TestShellGroupItem_FilterValue(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()

	group := NewShellGroupItem(&sty, []ShellCommand{
		{Command: "first command"},
		{Command: "second command"},
	}).(*ShellGroupItem)
	require.Equal(t, "first command", group.FilterValue())
}

// TestShellGroupItem_ID ensures unique IDs are generated.
func TestShellGroupItem_ID(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()

	group1 := NewShellGroupItem(&sty, []ShellCommand{{Command: "cmd1"}})
	group2 := NewShellGroupItem(&sty, []ShellCommand{{Command: "cmd2"}})

	require.NotEqual(t, group1.ID(), group2.ID())
	require.True(t, strings.HasPrefix(group1.ID(), "shell-group-"))
}
