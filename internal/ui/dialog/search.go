package dialog

import (
	"context"
	"image"

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
	com  *common.Common
	help help.Model
	list *list.FilterableList

	input textinput.Model

	bodyArea image.Rectangle

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

func (s *SearchDialog) performSearch() {
	query := s.input.Value()
	if len(query) < 3 {
		s.list.SetItems()
		return
	}
	results, err := s.com.Workspace.SearchMessages(
		context.Background(), query, search.SearchOpts{Limit: 50},
	)
	if err != nil {
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
	s.list.SetSelected(0)
	s.list.ScrollToTop()
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
				si := item.(*SearchItem)
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
			value := s.input.Value()
			if value != prevValue {
				s.performSearch()
			}
			return ActionCmd{cmd}
		}

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
	s.bodyArea = image.Rectangle{}
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

	bodyView := t.Dialog.List.Height(s.list.Height()).Render(s.list.Render())
	bodyView = joinScrollbar(t, bodyView, listHeight, listTotalHeight, listHeight, s.list.Offset())
	rc.AddPart(bodyView)

	rc.Help = renderDialogHelp(t, &s.help, s, innerWidth)
	view := rc.Render()

	DrawCenterCursor(scr, area, view, cur)
	return cur
}
