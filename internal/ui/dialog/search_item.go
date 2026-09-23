package dialog

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/crush/internal/search"
	"github.com/charmbracelet/crush/internal/ui/list"
	"github.com/charmbracelet/crush/internal/ui/styles"
	"github.com/charmbracelet/x/ansi"
)

// SearchItem wraps a search.Result to implement the ListItem interface.
type SearchItem struct {
	*list.Versioned
	search.Result
	t       *styles.Styles
	focused bool
	cache   map[int]string
}

var _ list.FilterableItem = &SearchItem{}
var _ list.Focusable = &SearchItem{}

// Filter returns the filterable value.
func (s *SearchItem) Filter() string {
	return s.Result.Snippet
}

// ID returns the unique identifier.
func (s *SearchItem) ID() string {
	return s.Result.MessageID
}

// Render returns the string representation of the search item.
func (s *SearchItem) Render(width int) string {
	snippet := strings.ReplaceAll(s.Result.Snippet, "\n", " ")
	snippet = strings.ReplaceAll(snippet, "\r", " ")
	snippet = strings.TrimSpace(snippet)

	info := s.Result.SessionTitle

	style := s.t.Dialog.NormalItem
	if s.focused {
		style = s.t.Dialog.SelectedItem
	}

	lineWidth := max(0, width-style.GetHorizontalFrameSize())
	var infoWidth int
	var infoText string

	if len(info) > 0 {
		if maxInfo := lineWidth / 2; lipgloss.Width(info)+2 > maxInfo {
			info = ansi.Truncate(info, max(0, maxInfo-2), "…")
		}
		infoText = fmt.Sprintf(" %s ", info)
		if s.focused {
			infoText = s.t.Dialog.Sessions.InfoFocused.Render(infoText)
		} else {
			infoText = s.t.Dialog.Sessions.InfoBlurred.Render(infoText)
		}
		infoWidth = lipgloss.Width(infoText)
	}

	// Use yellow for match highlighting
	marked := strings.ReplaceAll(snippet, "{{", "\x1b[33m")
	marked = strings.ReplaceAll(marked, "}}", "\x1b[39m")

	// Truncate keeping ANSI intact
	marked = ansi.Truncate(marked, max(0, lineWidth-infoWidth), "…")

	markedWidth := ansi.StringWidth(marked)
	gap := strings.Repeat(" ", max(0, lineWidth-markedWidth-infoWidth))

	return style.Render(marked + gap + infoText)
}

// SetFocused sets the focus state.
func (s *SearchItem) SetFocused(v bool) {
	s.focused = v
}

// Finished implements list.Item.
func (s *SearchItem) Finished() bool {
	return true
}
