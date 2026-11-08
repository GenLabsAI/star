package chat

import (
	"encoding/json"
	"testing"

	"github.com/charmbracelet/crush/internal/agent"
	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/ui/styles"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func bashTestInput(t *testing.T, command string) string {
	t.Helper()
	input, err := json.Marshal(tools.BashParams{Command: command})
	require.NoError(t, err)
	return string(input)
}

func newTestToolItem(t *testing.T, sty *styles.Styles, id, name, input, output string, isError bool) ToolMessageItem {
	t.Helper()
	tc := message.ToolCall{ID: id, Name: name, Input: input, Finished: true}
	result := &message.ToolResult{ToolCallID: id, Name: name, Content: output, IsError: isError}
	return NewToolMessageItem(sty, "msg-1", tc, result, false, "")
}

func bashTestItem(t *testing.T, sty *styles.Styles, id, command, output string, isError bool) ToolMessageItem {
	t.Helper()
	return newTestToolItem(t, sty, id, tools.BashToolName, bashTestInput(t, command), output, isError)
}

func TestToolGroupItem_SummaryListsActions(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	group := NewToolGroupItem(&sty, []ToolMessageItem{
		bashTestItem(t, &sty, "b1", "git status --short", " M a.go", false),
		bashTestItem(t, &sty, "b2", "go build ./...", "", false),
		newTestToolItem(t, &sty, "v1", tools.ViewToolName, `{"file_path":"a.go"}`, "package a", false),
		newTestToolItem(t, &sty, "v2", tools.ViewToolName, `{"file_path":"b.go"}`, "package b", false),
		newTestToolItem(t, &sty, "w1", tools.WebSearchToolName, `{"query":"oauth rotation"}`, "1. result", false),
		newTestToolItem(t, &sty, "s1", tools.SourcegraphToolName, `{"query":"repo:acme/app rotation"}`, "match", false),
	})

	out := ansi.Strip(group.RawRender(160))
	require.Contains(t, out, "Ran 2 commands, Read 2 files, Searched the web, Sourcegraph")
	require.Contains(t, out, "double-click to expand")
	require.NotContains(t, out, "git status --short", "summary must not list individual calls")
}

func TestToolGroupItem_ExpandedShowsCallsAndResults(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	group := NewToolGroupItem(&sty, []ToolMessageItem{
		bashTestItem(t, &sty, "b1", "git status --short", " M a.go", false),
		newTestToolItem(t, &sty, "v1", tools.ViewToolName, `{"file_path":"a.go"}`, "package a", false),
	})

	require.True(t, group.(*ToolGroupItem).ToggleExpanded())
	out := ansi.Strip(group.RawRender(160))
	require.Contains(t, out, "Ran 1 command, Read 1 file")
	require.Contains(t, out, "git status --short")
	require.Contains(t, out, " M a.go")
	require.Contains(t, out, "a.go")

	require.False(t, group.(*ToolGroupItem).ToggleExpanded())
	out = ansi.Strip(group.RawRender(160))
	require.NotContains(t, out, "git status --short")
}

func TestToolGroupItem_FailureSummary(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	group := NewToolGroupItem(&sty, []ToolMessageItem{
		bashTestItem(t, &sty, "b1", "go test ./...", "FAIL", true),
		bashTestItem(t, &sty, "b2", "git status", "", false),
	})

	out := ansi.Strip(group.RawRender(120))
	require.Contains(t, out, "Ran 2 commands")
	require.Contains(t, out, "1 failed")
}

func TestToolGroupItem_DoubleClickToggles(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	group := NewToolGroupItem(&sty, []ToolMessageItem{
		bashTestItem(t, &sty, "b1", "git status --short", " M a.go", false),
		bashTestItem(t, &sty, "b2", "go build ./...", "", false),
	})
	item := group.(*ToolGroupItem)

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

func TestExtractMessageItems_GroupsMixedToolRun(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	msg := &message.Message{
		ID:   "a1",
		Role: message.Assistant,
		Parts: []message.ContentPart{
			message.ToolCall{ID: "b1", Name: tools.BashToolName, Input: bashTestInput(t, "ls"), Finished: true},
			message.ToolCall{ID: "b2", Name: tools.ViewToolName, Input: `{"file_path":"a.go"}`, Finished: true},
			message.ToolCall{ID: "b3", Name: tools.WebSearchToolName, Input: `{"query":"x"}`, Finished: true},
		},
	}
	results := map[string]message.ToolResult{
		"b1": {ToolCallID: "b1", Content: "a.go"},
		"b2": {ToolCallID: "b2", Content: "package a"},
		"b3": {ToolCallID: "b3", Content: "result"},
	}

	items := ExtractMessageItems(&sty, msg, results, "")
	require.Len(t, items, 1, "a mixed run collapses into one summary")
	group, ok := items[0].(*ToolGroupItem)
	require.True(t, ok)
	require.Contains(t, ansi.Strip(group.RawRender(160)), "Ran 1 command, Read 1 file, Searched the web")
}

func TestExtractMessageItems_EditBreaksRun(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	msg := &message.Message{
		ID:   "a1",
		Role: message.Assistant,
		Parts: []message.ContentPart{
			message.ToolCall{ID: "b1", Name: tools.BashToolName, Input: bashTestInput(t, "git status"), Finished: true},
			message.ToolCall{ID: "e1", Name: tools.EditToolName, Input: `{"file_path":"a.go","old_string":"a","new_string":"b"}`, Finished: true},
			message.ToolCall{ID: "b2", Name: tools.BashToolName, Input: bashTestInput(t, "go build ./..."), Finished: true},
			message.ToolCall{ID: "v1", Name: tools.ViewToolName, Input: `{"file_path":"b.go"}`, Finished: true},
		},
	}
	results := map[string]message.ToolResult{
		"b1": {ToolCallID: "b1", Content: ""},
		"e1": {ToolCallID: "e1", Content: "Applied 1 edit"},
		"b2": {ToolCallID: "b2", Content: ""},
		"v1": {ToolCallID: "v1", Content: "package b"},
	}

	items := ExtractMessageItems(&sty, msg, results, "")
	require.Len(t, items, 3, "the edit must split the summary in two")
	_, ok := items[0].(*ToolGroupItem)
	require.True(t, ok)
	_, groupedAgain := items[1].(*ToolGroupItem)
	require.False(t, groupedAgain, "edit keeps its own row")
	require.Equal(t, "e1", items[1].ID())
	_, ok = items[2].(*ToolGroupItem)
	require.True(t, ok)
}

func TestExtractMessageItems_SubagentStaysSeparate(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	msg := &message.Message{
		ID:   "a1",
		Role: message.Assistant,
		Parts: []message.ContentPart{
			message.ToolCall{ID: "b1", Name: tools.BashToolName, Input: bashTestInput(t, "ls"), Finished: true},
			message.ToolCall{ID: "a2", Name: agent.AgentToolName, Input: `{"prompt":"review the diff"}`, Finished: true},
			message.ToolCall{ID: "b2", Name: tools.BashToolName, Input: bashTestInput(t, "pwd"), Finished: true},
		},
	}
	results := map[string]message.ToolResult{
		"b1": {ToolCallID: "b1", Content: ""},
		"a2": {ToolCallID: "a2", Content: "done"},
		"b2": {ToolCallID: "b2", Content: "/src"},
	}

	items := ExtractMessageItems(&sty, msg, results, "")
	require.Len(t, items, 3, "the subagent call must split the summary")
	_, ok := items[1].(*AgentToolMessageItem)
	require.True(t, ok, "subagent keeps its own row")
}
