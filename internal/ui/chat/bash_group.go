package chat

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/ui/common"
	"github.com/charmbracelet/crush/internal/ui/list"
	"github.com/charmbracelet/crush/internal/ui/styles"
	"github.com/charmbracelet/x/ansi"
)

// BashToolGroupItem collapses a run of consecutive Bash tool calls into a
// single summary row ("✓ Ran 4 commands"). Expanding it reveals every call
// and its output. The item owns the underlying tool items so their cached
// renders and status stay in sync with the agent.
type BashToolGroupItem struct {
	*list.Versioned
	*highlightableMessageItem
	*focusableMessageItem

	id       string
	items    []ToolMessageItem
	expanded bool
	sty      *styles.Styles
}

var (
	_ MessageItem             = (*BashToolGroupItem)(nil)
	_ Expandable              = (*BashToolGroupItem)(nil)
	_ NestedToolContainer     = (*BashToolGroupItem)(nil)
	_ KeyEventHandler         = (*BashToolGroupItem)(nil)
	_ list.MouseClickable     = (*BashToolGroupItem)(nil)
	_ MouseDoubleClickHandler = (*BashToolGroupItem)(nil)
)

// NewBashToolGroupItem creates a group from a run of Bash tool items.
func NewBashToolGroupItem(sty *styles.Styles, items []ToolMessageItem) MessageItem {
	v := list.NewVersioned()
	id := "bash-group"
	if len(items) > 0 {
		id += "-" + items[0].ToolCall().ID
	}
	return &BashToolGroupItem{
		Versioned:                v,
		highlightableMessageItem: defaultHighlighter(sty, v),
		focusableMessageItem:     newFocusableMessageItem(v),
		id:                       id,
		items:                    items,
		sty:                      sty,
	}
}

// ID implements MessageItem.
func (g *BashToolGroupItem) ID() string { return g.id }

// FilterValue implements list.Item.
func (g *BashToolGroupItem) FilterValue() string {
	if len(g.items) == 0 {
		return ""
	}
	return g.items[0].ToolCall().Name
}

// Finished implements list.Item. A group is finished once every call in it
// has finished and recorded its result.
func (g *BashToolGroupItem) Finished() bool {
	for _, item := range g.items {
		if !item.Finished() {
			return false
		}
	}
	return true
}

// ToggleExpanded implements Expandable. Expanding the group also expands the
// child calls so their full output is visible.
func (g *BashToolGroupItem) ToggleExpanded() bool {
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
func (g *BashToolGroupItem) NestedTools() []ToolMessageItem { return g.items }

// SetNestedTools implements NestedToolContainer.
func (g *BashToolGroupItem) SetNestedTools(items []ToolMessageItem) {
	if len(items) == 0 {
		return
	}
	g.items = items
	g.Bump()
}

// AddNestedTool implements NestedToolContainer.
func (g *BashToolGroupItem) AddNestedTool(item ToolMessageItem) {
	g.items = append(g.items, item)
	g.Bump()
}

// HandleMouseClick implements MouseClickable. A single click only selects
// the row; expansion is a deliberate double-click.
func (g *BashToolGroupItem) HandleMouseClick(btn ansi.MouseButton, x, y int) bool {
	return false
}

// HandleDoubleClick implements MouseDoubleClickHandler.
func (g *BashToolGroupItem) HandleDoubleClick(x, y int) bool {
	g.ToggleExpanded()
	return true
}

// HandleKeyEvent implements KeyEventHandler for copying the whole run.
func (g *BashToolGroupItem) HandleKeyEvent(key tea.KeyMsg) (bool, tea.Cmd) {
	switch key.String() {
	case "c", "y":
		var parts []string
		for _, item := range g.items {
			var params tools.BashParams
			_ = json.Unmarshal([]byte(item.ToolCall().Input), &params)
			parts = append(parts, "$ "+params.Command)
			if result, ok := item.(interface{ Result() *message.ToolResult }); ok {
				if r := result.Result(); r != nil && r.Content != "" {
					parts = append(parts, ansi.Strip(r.Content))
				}
			}
			parts = append(parts, "")
		}
		text := strings.TrimRight(strings.Join(parts, "\n"), "\n")
		return true, common.CopyToClipboard(text, "Commands copied to clipboard")
	}
	return false, nil
}

// Render implements MessageItem. It prefixes every line with the same
// focused/blurred treatment tool rows use.
func (g *BashToolGroupItem) Render(width int) string {
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
func (g *BashToolGroupItem) RawRender(width int) string {
	if g.expanded {
		return g.renderExpanded(width)
	}
	return g.renderSummary(width)
}

func (g *BashToolGroupItem) renderSummary(width int) string {
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
		label = fmt.Sprintf("Running %d command%s…", count, pluralSuffix(count))
	} else {
		label = fmt.Sprintf("Ran %d command%s", count, pluralSuffix(count))
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

func (g *BashToolGroupItem) renderExpanded(width int) string {
	parts := []string{g.renderSummary(width)}
	for _, item := range g.items {
		parts = append(parts, item.RawRender(width))
	}
	return strings.Join(parts, "\n\n")
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

// groupBashToolItems collapses runs of two or more consecutive finished Bash
// tool items into a single [BashToolGroupItem], preserving order.
func groupBashToolItems(sty *styles.Styles, items []MessageItem) []MessageItem {
	out := make([]MessageItem, 0, len(items))
	var run []ToolMessageItem

	flush := func() {
		if len(run) < 2 {
			for _, item := range run {
				out = append(out, item)
			}
			run = nil
			return
		}
		for _, item := range run {
			if !item.Finished() {
				for _, pending := range run {
					out = append(out, pending)
				}
				run = nil
				return
			}
		}
		out = append(out, NewBashToolGroupItem(sty, run))
		run = nil
	}

	for _, item := range items {
		if toolItem, ok := item.(ToolMessageItem); ok && toolItem.ToolCall().Name == tools.BashToolName {
			run = append(run, toolItem)
			continue
		}
		flush()
		out = append(out, item)
	}
	flush()
	return out
}
