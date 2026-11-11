package dialog

import (
	"context"
	"image"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/search"
	"github.com/charmbracelet/crush/internal/ui/common"
	"github.com/charmbracelet/crush/internal/ui/list"
	uv "github.com/charmbracelet/ultraviolet"
)

// SearchID is the identifier for the search dialog.
const SearchID = "search"

const (
	// searchMinQueryLen is the shortest query worth running.
	searchMinQueryLen = 3
	// searchDebounce is how long the query must sit still before it is
	// searched, so typing a word runs one search instead of one per key.
	searchDebounce = 150 * time.Millisecond
	// searchResultLimit caps the results fetched per search.
	searchResultLimit = 50
)

// searchDebounceMsg fires once the query has been quiet for
// [searchDebounce].
type searchDebounceMsg struct {
	generation uint64
	query      string
}

// searchResultsMsg carries the outcome of a search run off the Update
// goroutine.
type searchResultsMsg struct {
	generation uint64
	results    []search.Result
	err        error
}

// messageSearcher is implemented by workspaces that can search message
// history. It is intentionally optional so that workspaces without search
// support do not have to implement it.
type messageSearcher interface {
	SearchMessages(ctx context.Context, query string, opts search.SearchOpts) ([]search.Result, error)
}

// SearchDialog is a dialog for searching across sessions.
type SearchDialog struct {
	com  *common.Common
	help help.Model
	list *list.FilterableList

	input textinput.Model

	// searchGeneration identifies the latest query. Results and debounce
	// ticks carrying an older generation are stale, so they are dropped
	// instead of overwriting the newer query's results.
	searchGeneration uint64

	bodyArea image.Rectangle
	errText  string

	keyMap struct {
		Select   key.Binding
		Next     key.Binding
		Previous key.Binding
		UpDown   key.Binding
		Close    key.Binding
	}
}

// ShortHelp implements help.KeyMap.
func (s *SearchDialog) ShortHelp() []key.Binding {
	return []key.Binding{s.keyMap.Select, s.keyMap.UpDown, s.keyMap.Close}
}

// FullHelp implements help.KeyMap.
func (s *SearchDialog) FullHelp() [][]key.Binding {
	return [][]key.Binding{s.ShortHelp()}
}

var _ Dialog = (*SearchDialog)(nil)

// NewSearch creates a new search dialog.
func NewSearch(com *common.Common) (*SearchDialog, error) {
	s := &SearchDialog{
		com: com,
	}

	helpModel := help.New()
	helpModel.Styles = com.Styles.DialogHelpStyles()
	s.help = helpModel

	s.list = list.NewFilterableList()
	s.list.Focus()

	s.input = textinput.New()
	s.input.SetVirtualCursor(false)
	s.input.Placeholder = "Search across sessions (min 3 chars)..."
	s.input.SetStyles(com.Styles.TextInput)
	s.input.Focus()

	s.keyMap.Select = key.NewBinding(
		key.WithKeys("enter", "ctrl+y"),
		key.WithHelp("enter", "open"),
	)
	s.keyMap.Next = key.NewBinding(
		key.WithKeys("down", "ctrl+n"),
		key.WithHelp("↓", "next"),
	)
	s.keyMap.Previous = key.NewBinding(
		key.WithKeys("up", "ctrl+p"),
		key.WithHelp("↑", "prev"),
	)
	s.keyMap.UpDown = key.NewBinding(
		key.WithKeys("up", "down"),
		key.WithHelp("↑↓", "choose"),
	)
	s.keyMap.Close = CloseKey

	return s, nil
}

// ID implements Dialog.
func (s *SearchDialog) ID() string {
	return SearchID
}

// scheduleSearch invalidates any in-flight search and starts a debounced one
// for the current query. It returns nil when the query is too short to run.
func (s *SearchDialog) scheduleSearch() tea.Cmd {
	s.searchGeneration++
	generation := s.searchGeneration

	query := strings.TrimSpace(s.input.Value())
	if len(query) < searchMinQueryLen {
		s.errText = ""
		s.list.SetItems()
		return nil
	}

	return tea.Tick(searchDebounce, func(time.Time) tea.Msg {
		return searchDebounceMsg{generation: generation, query: query}
	})
}

// searchCmd searches the workspace off the Update goroutine and reports the
// results back as a [searchResultsMsg].
func (s *SearchDialog) searchCmd(searcher messageSearcher, generation uint64, query string) tea.Cmd {
	return func() tea.Msg {
		results, err := searcher.SearchMessages(
			context.Background(), query, search.SearchOpts{Limit: searchResultLimit},
		)
		return searchResultsMsg{generation: generation, results: results, err: err}
	}
}

// Cursor returns the cursor position relative to the dialog.
func (s *SearchDialog) Cursor() *tea.Cursor {
	return InputCursor(s.com.Styles, s.input.Cursor())
}

// HandleMsg implements Dialog.
func (s *SearchDialog) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, s.keyMap.Close):
			return ActionClose{}
		case key.Matches(msg, s.keyMap.Select):
			if item := s.list.SelectedItem(); item != nil {
				si, ok := item.(*SearchItem)
				if !ok {
					return nil
				}
				return ActionSearchJumpToMessage{
					SessionID: si.Result.SessionID,
					MessageID: si.Result.MessageID,
				}
			}
		case key.Matches(msg, s.keyMap.Previous):
			s.list.Focus()
			if s.list.IsSelectedFirst() {
				s.list.SelectLast()
			} else {
				s.list.SelectPrev()
			}
			s.list.ScrollToSelected()
		case key.Matches(msg, s.keyMap.Next):
			s.list.Focus()
			if s.list.IsSelectedLast() {
				s.list.SelectFirst()
			} else {
				s.list.SelectNext()
			}
			s.list.ScrollToSelected()
		default:
			prevValue := s.input.Value()
			var cmd tea.Cmd
			s.input, cmd = s.input.Update(msg)
			if s.input.Value() != prevValue {
				cmd = tea.Batch(cmd, s.scheduleSearch())
			}
			return ActionCmd{Cmd: cmd}
		}

	case searchDebounceMsg:
		if msg.generation != s.searchGeneration {
			return nil
		}
		searcher, ok := s.com.Workspace.(messageSearcher)
		if !ok {
			s.errText = "Search is unavailable in this workspace."
			s.list.SetItems()
			return nil
		}
		return ActionCmd{Cmd: s.searchCmd(searcher, msg.generation, msg.query)}

	case searchResultsMsg:
		if msg.generation != s.searchGeneration {
			return nil
		}
		if msg.err != nil {
			s.errText = "Search unavailable: " + msg.err.Error()
			s.list.SetItems()
			return nil
		}
		s.errText = ""
		s.list.SetItems(newSearchItems(s.com.Styles, msg.results)...)
		if s.list.Len() > 0 {
			s.list.SetSelected(0)
		}
		s.list.ScrollToTop()

	case common.CoalescedWheelMsg:
		if image.Pt(msg.Mouse.X, msg.Mouse.Y).In(s.bodyArea) {
			s.list.ScrollBy(int(msg.DeltaY))
		}
	}
	return nil
}

// Draw implements Dialog.
func (s *SearchDialog) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := s.com.Styles
	s.bodyArea = area
	width := max(0, min(defaultDialogMaxWidth+20, area.Dx()-t.Dialog.View.GetHorizontalBorderSize()))
	height := max(0, min(defaultDialogHeight+10, area.Dy()-t.Dialog.View.GetVerticalBorderSize()))
	innerWidth := width - t.Dialog.View.GetHorizontalFrameSize()

	s.input.SetWidth(dialogInputTextWidth(t, s.input, innerWidth))
	listHeight, listTotalHeight, _ := sizeDialogList(t, s.list, innerWidth, height)

	cur := s.Cursor()
	rc := NewRenderContext(t, width)
	rc.Title = "Search"

	inputView := t.Dialog.InputPrompt.Render(s.input.View())
	rc.AddPart(inputView)

	if s.errText != "" {
		rc.AddPart(t.Dialog.NormalItem.Render(s.errText))
	}

	bodyView := t.Dialog.List.Height(s.list.Height()).Render(s.list.Render())
	bodyView = joinScrollbar(t, bodyView, listHeight, listTotalHeight, listHeight, s.list.Offset())
	rc.AddPart(bodyView)

	rc.Help = renderDialogHelp(t, &s.help, s, innerWidth)
	view := rc.Render()

	DrawCenterCursor(scr, area, view, cur)
	return cur
}
