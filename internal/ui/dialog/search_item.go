package dialog

import (
	"github.com/charmbracelet/crush/internal/search"
	"github.com/charmbracelet/crush/internal/ui/list"
	"github.com/charmbracelet/crush/internal/ui/styles"
)

type SearchItem struct {
	*list.Versioned
	search.Result
	t       *styles.Styles
	focused bool
	cache   map[int]string
}

var _ list.FilterableItem = &SearchItem{}
var _ list.Focusable = &SearchItem{}

func (s *SearchItem) Filter() string {
	return s.Result.Snippet
}

func (s *SearchItem) ID() string {
	return s.Result.MessageID
}

func (s *SearchItem) Render(width int) string {
	// Use existing list item rendering pattern
	// Replace FTS5 highlight markers with ANSI
	// Simplistic for now, actual implementation would map FTS5 {{ and }} to proper styling
	snippet := s.Result.Snippet

	info := s.Result.SessionTitle
	return renderItem(ListItemStyles{
		ItemBlurred: s.t.Dialog.NormalItem,
		ItemFocused: s.t.Dialog.SelectedItem,
		InfoTextBlurred: s.t.Dialog.Sessions.InfoBlurred,
		InfoTextFocused: s.t.Dialog.Sessions.InfoFocused,
	}, snippet, info, s.focused, false, width, s.cache, nil)
}

func (s *SearchItem) SetFocused(v bool) {
	s.focused = v
}

func (s *SearchItem) Finished() bool {
	return true
}
