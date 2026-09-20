package dialog

import (
	"context"
	"image"
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

// SearchDialog is a dialog for searching across sessions.
type SearchDialog struct {
	com           *common.Common
	help          help.Model
	list          *list.FilterableList
	input         textinput.Model
	
	bodyArea      image.Rectangle
	mouseScrolled bool
	lastClickTime time.Time
	lastClickID   string

	debounceTimer *time.Timer

	keyMap struct {
		Select   key.Binding
		Next     key.Binding
		Previous key.Binding
		UpDown   key.Binding
		Close    key.Binding
	}
}

func (s *SearchDialog) ShortHelp() []key.Binding {
	return []key.Binding{s.keyMap.Select, s.keyMap.UpDown, s.keyMap.Close}
}

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
	s.input.Placeholder = "Search across sessions..."
	s.input.SetStyles(com.Styles.TextInput)
	s.input.Focus()

	s.keyMap.Select = key.NewBinding(
		key.WithKeys("enter", "tab", "ctrl+y"),
		key.WithHelp("enter", "jump to message"),
	)
	s.keyMap.Next = key.NewBinding(
		key.WithKeys("down", "ctrl+n"),
		key.WithHelp("↓", "next item"),
	)
	s.keyMap.Previous = key.NewBinding(
		key.WithKeys("up", "ctrl+p"),
		key.WithHelp("↑", "previous item"),
	)
	s.keyMap.UpDown = key.NewBinding(
		key.WithKeys("up", "down"),
		key.WithHelp("↑↓", "choose"),
	)
	s.keyMap.Close = CloseKey

	return s, nil
}

func (s *SearchDialog) ID() string {
	return SearchID
}

func (s *SearchDialog) performSearch() {
	query := s.input.Value()
	if query == "" {
		s.list.SetItems()
		return
	}
	results, err := s.com.Workspace.SearchMessages(context.Background(), query, search.SearchOpts{Limit: 50})
	if err != nil {
		// Log error or ignore; we just render empty on failure for now.
		return
	}

	items := make([]list.FilterableItem, len(results))
	for i, r := range results {
		items[i] = &SearchItem{
			Versioned: list.NewVersioned(),
			Result:    r,
			t:         s.com.Styles,
			cache:     make(map[int]string),
		}
	}
	s.list.SetItems(items...)
}

func (s *SearchDialog) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keyMap.Close):
			return ActionClose{}
		case key.Matches(msg, s.keyMap.Select):
			if item := s.list.SelectedItem(); item != nil {
				si := item.(*SearchItem)
				return ActionSearchJumpToMessage{
					SessionID: si.Result.SessionID,
					MessageID: si.Result.MessageID,
				}
			}
		case key.Matches(msg, s.keyMap.Next):
			s.list.SelectNext()
		case key.Matches(msg, s.keyMap.Previous):
			s.list.SelectPrev()
		default:
			var cmd tea.Cmd
			before := s.input.Value()
			s.input, cmd = s.input.Update(msg)
			if before != s.input.Value() {
				if s.debounceTimer != nil {
					s.debounceTimer.Stop()
				}
				s.debounceTimer = time.AfterFunc(150*time.Millisecond, s.performSearch)
			}
			return ActionCmd{Cmd: cmd}
		}

	case tea.MouseMsg:
		// basic mouse support similar to sessions dialog
		return nil
	}
	return nil
}

func (s *SearchDialog) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := s.com.Styles
	width := max(0, min(defaultDialogMaxWidth+20, area.Dx()-t.Dialog.View.GetHorizontalBorderSize()))
	height := max(0, min(defaultDialogHeight+10, area.Dy()-t.Dialog.View.GetVerticalBorderSize()))
	innerWidth := width - t.Dialog.View.GetHorizontalFrameSize()

	s.input.SetWidth(dialogInputTextWidth(t, s.input, innerWidth))
	listHeight, listTotalHeight, _ := sizeDialogList(t, s.list, innerWidth, height)

	rc := NewRenderContext(t, width)
	rc.Title = "Search"
	rc.AddPart(t.Dialog.InputPrompt.Render(s.input.View()))

	bodyView := t.Dialog.List.Height(s.list.Height()).Render(s.list.Render())
	bodyView = joinScrollbar(t, bodyView, listHeight, listTotalHeight, listHeight, s.list.Offset())
	rc.AddPart(bodyView)

	rc.Help = renderDialogHelp(t, &s.help, s, innerWidth)
	view := rc.Render()

	DrawCenterCursor(scr, area, view, InputCursor(t, s.input.Cursor()))
	return InputCursor(t, s.input.Cursor())
}
