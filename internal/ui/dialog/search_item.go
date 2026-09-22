package dialog

import (
	"strings"

	"github.com/charmbracelet/crush/internal/search"
	"github.com/charmbracelet/crush/internal/ui/list"
	"github.com/charmbracelet/crush/internal/ui/styles"
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
	// Strip FTS5 highlight markers.
	snippet := strings.ReplaceAll(s.Result.Snippet, "{{", "")
	snippet = strings.ReplaceAll(snippet, "}}", "")

	info := s.Result.SessionTitle
	return renderItem(ListItemStyles{
		ItemBlurred:     s.t.Dialog.NormalItem,
		ItemFocused:     s.t.Dialog.SelectedItem,
		InfoTextBlurred: s.t.Dialog.Sessions.InfoBlurred,
		InfoTextFocused: s.t.Dialog.Sessions.InfoFocused,
	}, snippet, info, s.focused, false, width, s.cache, nil)
}

// SetFocused sets the focus state.
func (s *SearchItem) SetFocused(v bool) {
	s.focused = v
}

// Finished implements list.Item.
func (s *SearchItem) Finished() bool {
	return true
}
