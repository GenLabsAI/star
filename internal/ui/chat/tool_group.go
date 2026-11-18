package chat

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/agent"
	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/ui/common"
	"github.com/charmbracelet/crush/internal/ui/list"
	"github.com/charmbracelet/crush/internal/ui/styles"
	"github.com/charmbracelet/x/ansi"
)

// ToolGroupItem collapses a run of consecutive tool calls into one summary
// row such as "Ran 4 commands, Read 2 files, Searched the web, Sourcegraph".
// Expanding it reveals every call and its output in order. File edits and
// subagent runs are never grouped: they break the run and keep their own row.
type ToolGroupItem struct {
	*list.Versioned
	*highlightableMessageItem
	*focusableMessageItem

	id       string
	items    []ToolMessageItem
	expanded bool
	sty      *styles.Styles
}

var (
	_ MessageItem             = (*ToolGroupItem)(nil)
	_ Expandable              = (*ToolGroupItem)(nil)
	_ NestedToolContainer     = (*ToolGroupItem)(nil)
	_ KeyEventHandler         = (*ToolGroupItem)(nil)
	_ list.MouseClickable     = (*ToolGroupItem)(nil)
	_ MouseDoubleClickHandler = (*ToolGroupItem)(nil)
)

// NewToolGroupItem creates a summary item from a run of tool items.
func NewToolGroupItem(sty *styles.Styles, items []ToolMessageItem) MessageItem {
	v := list.NewVersioned()
	id := "tool-group"
	if len(items) > 0 {
		id += "-" + items[0].ToolCall().ID
	}
	return &ToolGroupItem{
		Versioned:                v,
		highlightableMessageItem: defaultHighlighter(sty, v),
		focusableMessageItem:     newFocusableMessageItem(v),
		id:                       id,
		items:                    items,
		sty:                      sty,
	}
}

// GroupableTool reports whether calls to name can join a summary row. File
// edits and subagent runs stay separate so diffs and delegated work are never
// hidden behind a count.
func GroupableTool(name string) bool {
	switch name {
	case tools.EditToolName, tools.MultiEditToolName, tools.WriteToolName,
		tools.RenameToolName, tools.ReplaceSymbolToolName,
		agent.AgentToolName, tools.TeamToolName:
		return false
	}
	return true
}

// ID implements MessageItem.
func (g *ToolGroupItem) ID() string { return g.id }

// FilterValue implements list.Item.
func (g *ToolGroupItem) FilterValue() string {
	names := make([]string, 0, len(g.items))
	for _, item := range g.items {
		names = append(names, item.ToolCall().Name)
	}
	return strings.Join(names, " ")
}

// Finished implements list.Item. A group is finished once every call in it
// has finished and recorded its result.
func (g *ToolGroupItem) Finished() bool {
	for _, item := range g.items {
		if !item.Finished() {
			return false
		}
	}
	return true
}

// ToggleExpanded implements Expandable. Expanding the group also expands the
// child calls so their full output is visible.
func (g *ToolGroupItem) ToggleExpanded() bool {
	g.expanded = !g.expanded
	for _, item := range g.items {
		if expandable, ok := item.(interface{ SetExpanded(bool) }); ok {
			expandable.SetExpanded(g.expanded)
		}
	}
	g.Bump()
	return g.expanded
}

// NestedTools implements NestedToolContainer so child call IDs resolve to
// this group in the chat index.
func (g *ToolGroupItem) NestedTools() []ToolMessageItem { return g.items }

// SetNestedTools implements NestedToolContainer.
func (g *ToolGroupItem) SetNestedTools(items []ToolMessageItem) {
	if len(items) == 0 {
		return
	}
	g.items = items
	g.Bump()
}

// AddNestedTool implements NestedToolContainer.
func (g *ToolGroupItem) AddNestedTool(item ToolMessageItem) {
	g.items = append(g.items, item)
	g.Bump()
}

// HandleMouseClick implements MouseClickable. A single click only selects
// the row; expansion is a deliberate double-click.
func (g *ToolGroupItem) HandleMouseClick(btn ansi.MouseButton, x, y int) bool {
	return false
}

// HandleDoubleClick implements MouseDoubleClickHandler.
func (g *ToolGroupItem) HandleDoubleClick(x, y int) bool {
	g.ToggleExpanded()
	return true
}

// HandleKeyEvent implements KeyEventHandler for copying the whole run.
func (g *ToolGroupItem) HandleKeyEvent(key tea.KeyMsg) (bool, tea.Cmd) {
	switch key.String() {
	case "c", "y":
		var parts []string
		for _, item := range g.items {
			call := item.ToolCall()
			if call.Name == tools.BashToolName {
				var params tools.BashParams
				_ = json.Unmarshal([]byte(call.Input), &params)
				parts = append(parts, "$ "+params.Command)
			} else {
				parts = append(parts, strings.TrimSpace(call.Name+" "+call.Input))
			}
			if result, ok := item.(interface{ Result() *message.ToolResult }); ok {
				if r := result.Result(); r != nil && r.Content != "" {
					parts = append(parts, ansi.Strip(r.Content))
				}
			}
			parts = append(parts, "")
		}
		text := strings.TrimRight(strings.Join(parts, "\n"), "\n")
		return true, common.CopyToClipboard(text, "Tool content copied to clipboard")
	}
	return false, nil
}

// Render implements MessageItem. It prefixes every line with the same
// focused/blurred treatment tool rows use.
func (g *ToolGroupItem) Render(width int) string {
	innerWidth := max(0, width-MessageLeftPaddingTotal)
	content := g.RawRender(innerWidth)

	var prefix string
	if g.focused {
		prefix = g.sty.Messages.ToolCallFocused.Render()
	} else {
		prefix = g.sty.Messages.ToolCallBlurred.Render()
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	out := strings.Join(lines, "\n")
	return g.renderHighlighted(out, width, len(lines))
}

// RawRender implements MessageItem.
func (g *ToolGroupItem) RawRender(width int) string {
	if g.expanded {
		return g.renderExpanded(width)
	}
	return g.renderSummary(width)
}

func (g *ToolGroupItem) renderSummary(width int) string {
	count := len(g.items)
	if count == 0 {
		return ""
	}

	failed := 0
	running := 0
	for _, item := range g.items {
		switch groupItemStatus(item) {
		case ToolStatusError:
			failed++
		case ToolStatusRunning, ToolStatusAwaitingPermission:
			running++
		}
	}

	status := ToolStatusSuccess
	switch {
	case running > 0:
		status = ToolStatusRunning
	case failed > 0:
		status = ToolStatusError
	}

	var label string
	if running > 0 {
		label = fmt.Sprintf("Running %d tool%s…", count, pluralSuffix(count))
	} else {
		label = summarizeToolActions(g.items)
	}

	line := toolIcon(g.sty, status) + " " + g.sty.Tool.NameNormal.Render(label)
	if failed > 0 {
		line += g.sty.Tool.ErrorMessage.Render(fmt.Sprintf(" · %d failed", failed))
	}

	hint := "(double-click to expand)"
	if g.expanded {
		hint = "(double-click to collapse)"
	}
	line += " " + g.sty.Messages.ShellGroupExpandHint.Render(hint)

	if width > 0 {
		line = ansi.Truncate(line, width, "…")
	}
	return line
}

func (g *ToolGroupItem) renderExpanded(width int) string {
	parts := []string{g.renderSummary(width)}
	for _, item := range g.items {
		parts = append(parts, item.RawRender(width))
	}
	return strings.Join(parts, "\n\n")
}

type toolActionCount struct {
	key      string
	singular string
	plural   string
	count    int
}

// summarizeToolActions turns a run of tools into an ordered, comma-separated
// tally, e.g. "Ran 4 commands, Read 2 files, Searched the web, Sourcegraph".
func summarizeToolActions(items []ToolMessageItem) string {
	parts := make([]toolActionCount, 0, len(items))
	index := make(map[string]int, len(items))
	for _, item := range items {
		key, singular, plural := toolAction(item.ToolCall().Name)
		if i, ok := index[key]; ok {
			parts[i].count++
			continue
		}
		index[key] = len(parts)
		parts = append(parts, toolActionCount{
			key:      key,
			singular: singular,
			plural:   plural,
			count:    1,
		})
	}

	labels := make([]string, 0, len(parts))
	for _, part := range parts {
		switch {
		case part.count == 1:
			labels = append(labels, part.singular)
		case part.plural != "":
			labels = append(labels, fmt.Sprintf(part.plural, part.count))
		default:
			labels = append(labels, fmt.Sprintf("%s ×%d", part.singular, part.count))
		}
	}
	return strings.Join(labels, ", ")
}

// toolAction maps a tool name to the phrase used in a group summary.
func toolAction(name string) (key, singular, plural string) {
	switch name {
	case tools.BashToolName, tools.JobOutputToolName, tools.JobKillToolName:
		return "bash", "Ran 1 command", "Ran %d commands"
	case tools.ViewToolName:
		return "view", "Read 1 file", "Read %d files"
	case tools.WebSearchToolName:
		return "web_search", "Searched the web", ""
	case tools.SourcegraphToolName:
		return "sourcegraph", "Sourcegraph", ""
	case tools.GrepToolName, tools.GlobToolName, tools.LSToolName:
		return "search", "Searched the codebase", ""
	case tools.FetchToolName, tools.WebFetchToolName, tools.AgenticFetchToolName:
		return "fetch", "Fetched the web", ""
	case tools.DiagnosticsToolName:
		return "diagnostics", "Checked diagnostics", ""
	case tools.TodosToolName:
		return "todos", "Updated todos", ""
	case tools.QuestionToolName:
		return "question", "Asked a question", ""
	default:
		label := strings.ReplaceAll(name, "_", " ")
		if label != "" {
			label = strings.ToUpper(label[:1]) + label[1:]
		}
		return name, label, ""
	}
}

// groupItemStatus derives a child's display status from its result when one
// is recorded, matching computeStatus on the base tool item.
func groupItemStatus(item ToolMessageItem) ToolStatus {
	if result, ok := item.(interface{ Result() *message.ToolResult }); ok {
		if r := result.Result(); r != nil {
			if r.IsError {
				return ToolStatusError
			}
			return ToolStatusSuccess
		}
	}
	return item.Status()
}

// groupToolItems collapses runs of consecutive groupable, finished tools into
// a single [ToolGroupItem], preserving order. Edits, writes, and subagent
// calls flush the current run and render on their own.
func groupToolItems(sty *styles.Styles, items []MessageItem) []MessageItem {
	out := make([]MessageItem, 0, len(items))
	var run []ToolMessageItem

	flush := func() {
		if len(run) == 0 {
			return
		}
		allFinished := true
		for _, item := range run {
			if !item.Finished() {
				allFinished = false
				break
			}
		}
		if allFinished {
			out = append(out, NewToolGroupItem(sty, run))
		} else {
			for _, item := range run {
				out = append(out, item)
			}
		}
		run = nil
	}

	for _, item := range items {
		if toolItem, ok := item.(ToolMessageItem); ok && GroupableTool(toolItem.ToolCall().Name) {
			run = append(run, toolItem)
			continue
		}
		flush()
		out = append(out, item)
	}
	flush()
	return out
}
