package chat

import (
	"encoding/json"
	"testing"

	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/ui/styles"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func newTestBashItem(t *testing.T, sty *styles.Styles, id, command, output string, status ToolStatus) ToolMessageItem {
	t.Helper()

	input, err := json.Marshal(tools.BashParams{Command: command})
	require.NoError(t, err)

	tc := message.ToolCall{
		ID:       id,
		Name:     tools.BashToolName,
		Input:    string(input),
		Finished: true,
	}
	result := &message.ToolResult{ToolCallID: id, Name: tools.BashToolName, Content: output, IsError: status == ToolStatusError}
	item := NewBashToolMessageItem(sty, tc, result, false, "")
	item.SetStatus(status)
	return item
}

func TestBashToolGroupItem_CollapsedSummary(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	group := NewBashToolGroupItem(&sty, []ToolMessageItem{
		newTestBashItem(t, &sty, "b1", "git status --short", " M a.go", ToolStatusSuccess),
		newTestBashItem(t, &sty, "b2", "go build ./...", "", ToolStatusSuccess),
		newTestBashItem(t, &sty, "b3", "gofmt -l internal/auth", "internal/auth/a.go", ToolStatusSuccess),
	})

	out := ansi.Strip(group.RawRender(120))
	require.Contains(t, out, "Ran 3 commands")
	require.Contains(t, out, "double-click to expand")
	require.NotContains(t, out, "git status --short", "collapsed group must not list individual commands")
	require.NotContains(t, out, "gofmt", "collapsed group must not list individual commands")
}

func TestBashToolGroupItem_ExpandedShowsCallsAndOutput(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	group := NewBashToolGroupItem(&sty, []ToolMessageItem{
		newTestBashItem(t, &sty, "b1", "git status --short", " M a.go", ToolStatusSuccess),
		newTestBashItem(t, &sty, "b2", "go build ./...", "", ToolStatusSuccess),
	})

	require.True(t, group.(*BashToolGroupItem).ToggleExpanded())
	out := ansi.Strip(group.RawRender(120))
	require.Contains(t, out, "Ran 2 commands")
	require.Contains(t, out, "git status --short")
	require.Contains(t, out, " M a.go")
	require.Contains(t, out, "go build ./...")

	require.False(t, group.(*BashToolGroupItem).ToggleExpanded())
	out = ansi.Strip(group.RawRender(120))
	require.NotContains(t, out, "git status --short")
}

func TestBashToolGroupItem_FailureSummary(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	group := NewBashToolGroupItem(&sty, []ToolMessageItem{
		newTestBashItem(t, &sty, "b1", "go test ./...", "FAIL", ToolStatusError),
		newTestBashItem(t, &sty, "b2", "git status", "", ToolStatusSuccess),
	})

	out := ansi.Strip(group.RawRender(120))
	require.Contains(t, out, "Ran 2 commands")
	require.Contains(t, out, "1 failed")
}

func TestExtractMessageItems_GroupsConsecutiveBashCalls(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	bashInput := func(cmd string) string {
		in, err := json.Marshal(tools.BashParams{Command: cmd})
		require.NoError(t, err)
		return string(in)
	}

	msg := &message.Message{
		ID:   "a1",
		Role: message.Assistant,
		Parts: []message.ContentPart{
			message.TextContent{Text: "Running the checks."},
			message.ToolCall{ID: "b1", Name: tools.BashToolName, Input: bashInput("ls"), Finished: true},
			message.ToolCall{ID: "b2", Name: tools.BashToolName, Input: bashInput("pwd"), Finished: true},
			message.ToolCall{ID: "b3", Name: tools.BashToolName, Input: bashInput("go build ./..."), Finished: true},
		},
	}
	results := map[string]message.ToolResult{
		"b1": {ToolCallID: "b1", Content: "a.go"},
		"b2": {ToolCallID: "b2", Content: "/src"},
		"b3": {ToolCallID: "b3", Content: ""},
	}

	items := ExtractMessageItems(&sty, msg, results, "")
	require.Len(t, items, 2, "assistant text plus one grouped row")
	group, ok := items[1].(*BashToolGroupItem)
	require.True(t, ok, "consecutive bash calls must collapse into one group")
	require.Len(t, group.NestedTools(), 3)

	out := ansi.Strip(group.RawRender(120))
	require.Contains(t, out, "Ran 3 commands")
}

func TestExtractMessageItems_SingleBashCallNotGrouped(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	input, err := json.Marshal(tools.BashParams{Command: "ls"})
	require.NoError(t, err)

	msg := &message.Message{
		ID:   "a1",
		Role: message.Assistant,
		Parts: []message.ContentPart{
			message.TextContent{Text: "Checking."},
			message.ToolCall{ID: "b1", Name: tools.BashToolName, Input: string(input), Finished: true},
		},
	}
	results := map[string]message.ToolResult{"b1": {ToolCallID: "b1", Content: "a.go"}}

	items := ExtractMessageItems(&sty, msg, results, "")
	require.Len(t, items, 2)
	_, ok := items[1].(*BashToolMessageItem)
	require.True(t, ok, "a lone bash call keeps its own row")
}

func TestBashToolGroupItem_DoubleClickToggles(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	group := NewBashToolGroupItem(&sty, []ToolMessageItem{
		newTestBashItem(t, &sty, "b1", "git status --short", " M a.go", ToolStatusSuccess),
		newTestBashItem(t, &sty, "b2", "go build ./...", "", ToolStatusSuccess),
	})
	item := group.(*BashToolGroupItem)

	require.False(t, item.HandleMouseClick(ansi.MouseButton1, 4, 0),
		"a single click must not expand the group")
	require.True(t, item.HandleDoubleClick(4, 0))
	out := ansi.Strip(item.RawRender(120))
	require.Contains(t, out, "git status --short")
	require.Contains(t, out, "double-click to collapse")

	require.True(t, item.HandleDoubleClick(4, 0))
	out = ansi.Strip(item.RawRender(120))
	require.NotContains(t, out, "git status --short")
	require.Contains(t, out, "double-click to expand")
}
