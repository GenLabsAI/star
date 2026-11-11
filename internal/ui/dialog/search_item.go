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

// SearchItem wraps a search.Result to implement the list item interfaces.
type SearchItem struct {
	*list.Versioned
	search.Result
	t       *styles.Styles
	focused bool
	cache   map[int]string
}

var (
	_ list.FilterableItem = (*SearchItem)(nil)
	_ list.Focusable      = (*SearchItem)(nil)
)

// newSearchItems builds the list rows for the given results, keeping one row
// per distinct visible result. The same text regularly ends up indexed in many
// messages (forks, repeated tool output, retries), and dozens of rows reading
// exactly the same are useless to pick from.
func newSearchItems(sty *styles.Styles, results []search.Result) []list.FilterableItem {
	items := make([]list.FilterableItem, 0, len(results))
	seen := make(map[string]struct{}, len(results))
	for _, result := range results {
		key := sanitizeLine(result.SessionTitle) + "\x00" + sanitizeLine(result.Snippet)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		items = append(items, &SearchItem{
			Versioned: list.NewVersioned(),
			Result:    result,
			t:         sty,
			cache:     make(map[int]string),
		})
	}
	return items
}

// Filter returns the filterable value.
func (s *SearchItem) Filter() string {
	return s.Result.Snippet
}

// ID returns the unique identifier.
func (s *SearchItem) ID() string {
	return s.Result.MessageID
}

// sanitizeLine makes text safe to draw on a single dialog row.
//
// Tabs matter most: terminals expand them to the next tab stop, but width
// measurement counts them differently, so a row containing one overflows the
// width it was truncated to and wraps. Newlines split a row in two, other
// control characters corrupt it, and ANSI escapes from indexed tool output
// would leak colors into the dialog, so all of them are stripped or folded
// into single spaces.
func sanitizeLine(text string) string {
	text = ansi.Strip(text)
	text = strings.Map(func(r rune) rune {
		switch {
		case r == '\t' || r == '\n' || r == '\r' || r == '\v' || r == '\f':
			return ' '
		case r < 0x20 || r == 0x7f:
			return -1
		}
		return r
	}, text)
	return strings.Join(strings.Fields(text), " ")
}

// Render returns the string representation of the search item.
func (s *SearchItem) Render(width int) string {
	snippet := sanitizeLine(s.Result.Snippet)
	info := sanitizeLine(s.Result.SessionTitle)

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

	marked := strings.ReplaceAll(snippet, "{{", "\x1b[33m")
	marked = strings.ReplaceAll(marked, "}}", "\x1b[39m")
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
