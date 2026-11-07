package dialog

import (
	"fmt"
	"image"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/crush/internal/ui/common"
	"github.com/charmbracelet/crush/internal/ui/list"
	"github.com/charmbracelet/crush/internal/ui/styles"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/sahilm/fuzzy"
)

// ForkFromID is the identifier for the fork point dialog.
const ForkFromID = "fork_from"

const forkDoubleClickThreshold = 400 * time.Millisecond

// ForkPoint is one message the user can fork the conversation from.
type ForkPoint struct {
	MessageID string
	Role      string
	Preview   string
	// Newest marks the most recent message. The dialog opens with it
	// selected and labels it so the default fork point is obvious.
	Newest bool
}

// ForkFrom lists the conversation's messages newest-first so the user can
// scroll back and fork from an earlier point.
type ForkFrom struct {
	com    *common.Common
	help   help.Model
	list   *list.FilterableList
	points []ForkPoint

	bodyArea      image.Rectangle
	mouseScrolled bool
	lastClickTime time.Time
	lastClickID   string

	keyMap struct {
		Select,
		Next,
		Previous,
		UpDown,
		PageUp,
		PageDown,
		Close key.Binding
	}
}

var _ Dialog = (*ForkFrom)(nil)

// NewForkFrom creates a new fork point dialog. points must be ordered newest
// first; the first point is selected by default.
func NewForkFrom(com *common.Common, points []ForkPoint) *ForkFrom {
	d := &ForkFrom{
		com:    com,
		points: points,
	}

	d.help = help.New()
	d.help.Styles = com.Styles.DialogHelpStyles()

	items := make([]list.FilterableItem, len(points))
	for i, p := range points {
		items[i] = NewForkPointItem(com.Styles, p)
	}
	d.list = list.NewFilterableList(items...)
	d.list.Focus()
	d.list.SetSelected(0)

	d.keyMap.Select = key.NewBinding(
		key.WithKeys("enter", "ctrl+y"),
		key.WithHelp("enter", "fork"),
	)
	d.keyMap.Next = key.NewBinding(
		key.WithKeys("down", "ctrl+n"),
		key.WithHelp("↓", "older"),
	)
	d.keyMap.Previous = key.NewBinding(
		key.WithKeys("up", "ctrl+p"),
		key.WithHelp("↑", "newer"),
	)
	d.keyMap.UpDown = key.NewBinding(
		key.WithKeys("up", "down"),
		key.WithHelp("↑↓", "choose"),
	)
	d.keyMap.PageUp = key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("pgup", "page up"),
	)
	d.keyMap.PageDown = key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("pgdown", "page down"),
	)
	d.keyMap.Close = CloseKey

	return d
}

// ID implements [Dialog].
func (d *ForkFrom) ID() string {
	return ForkFromID
}

// Select selects the fork point at index, scrolling it into view. It is used
// by the UI and visual harnesses to preselect an older message.
func (d *ForkFrom) Select(index int) {
	d.list.SetSelected(index)
	d.list.ScrollToSelected()
	d.mouseScrolled = false
}

// HandleMsg implements [Dialog].
func (d *ForkFrom) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, d.keyMap.Close):
			return ActionClose{}
		case key.Matches(msg, d.keyMap.Previous):
			d.mouseScrolled = false
			d.list.Focus()
			if d.list.IsSelectedFirst() {
				d.list.SelectLast()
			} else {
				d.list.SelectPrev()
			}
			d.list.ScrollToSelected()
		case key.Matches(msg, d.keyMap.Next):
			d.mouseScrolled = false
			d.list.Focus()
			if d.list.IsSelectedLast() {
				d.list.SelectFirst()
			} else {
				d.list.SelectNext()
			}
			d.list.ScrollToSelected()
		case key.Matches(msg, d.keyMap.PageUp):
			d.mouseScrolled = true
			d.list.ScrollBy(-max(1, d.list.Height()-1))
			d.list.SelectFirstInView()
		case key.Matches(msg, d.keyMap.PageDown):
			d.mouseScrolled = true
			d.list.ScrollBy(max(1, d.list.Height()-1))
			d.list.SelectLastInView()
		case key.Matches(msg, d.keyMap.Select):
			if point, ok := d.list.SelectedItem().(*ForkPointItem); ok {
				return ActionForkSession{MessageID: point.point.MessageID}
			}
		}
	case common.CoalescedWheelMsg:
		if image.Pt(msg.Mouse.X, msg.Mouse.Y).In(d.bodyArea) {
			d.mouseScrolled = true
			d.list.ScrollBy(int(msg.DeltaY))
			if msg.DeltaY < 0 {
				d.list.SelectFirstInView()
			} else {
				d.list.SelectLastInView()
			}
		}
	case tea.MouseClickMsg:
		return d.handleMouseClick(msg)
	}
	return nil
}

func (d *ForkFrom) handleMouseClick(msg tea.MouseClickMsg) Action {
	if msg.Button != tea.MouseLeft {
		d.resetMouseClick()
		return nil
	}
	area := d.bodyArea
	area.Max.X = min(area.Max.X, area.Min.X+d.list.Width())
	point := image.Pt(msg.X, msg.Y)
	if !point.In(area) {
		d.resetMouseClick()
		return nil
	}
	index, _ := d.list.ItemIndexAtPosition(point.X-area.Min.X, point.Y-area.Min.Y)
	if index < 0 {
		d.resetMouseClick()
		return nil
	}
	item, ok := d.list.ItemAt(index).(*ForkPointItem)
	if !ok {
		d.resetMouseClick()
		return nil
	}
	now := time.Now()
	if d.lastClickID == item.ID() && now.Sub(d.lastClickTime) <= forkDoubleClickThreshold {
		d.resetMouseClick()
		return ActionForkSession{MessageID: item.point.MessageID}
	}
	d.lastClickTime = now
	d.lastClickID = item.ID()
	d.list.SetSelected(index)
	d.mouseScrolled = true
	return nil
}

func (d *ForkFrom) resetMouseClick() {
	d.lastClickTime = time.Time{}
	d.lastClickID = ""
}

// Draw implements [Dialog].
func (d *ForkFrom) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := d.com.Styles
	d.bodyArea = image.Rectangle{}

	width := max(0, min(defaultDialogMaxWidth, area.Dx()-t.Dialog.View.GetHorizontalBorderSize()))
	height := max(0, min(defaultDialogHeight, area.Dy()-t.Dialog.View.GetVerticalBorderSize()))
	innerWidth := width - t.Dialog.View.GetHorizontalFrameSize()

	heightOffset := t.Dialog.Title.GetVerticalFrameSize() + titleContentHeight +
		t.Dialog.HelpView.GetVerticalFrameSize() +
		t.Dialog.View.GetVerticalFrameSize()
	listHeight := max(0, height-heightOffset)
	listTotalHeight := d.list.TotalHeight()

	scrollbarWidth := 0
	if listTotalHeight > listHeight {
		scrollbarWidth = 1
	}
	d.list.SetSize(max(0, innerWidth-scrollbarWidth), listHeight)

	// Keep the selection visible without fighting a deliberate wheel scroll.
	start, end := d.list.VisibleItemIndices()
	if !d.mouseScrolled && (d.list.Selected() < start || d.list.Selected() > end) {
		d.list.ScrollToSelected()
	}

	rc := NewRenderContext(t, width)
	rc.Title = "Fork From..."
	rc.TitleInfo = t.Dialog.SecondaryText.Render(fmt.Sprintf("%d messages", len(d.points)))
	bodyView := t.Dialog.List.Height(d.list.Height()).Render(d.list.Render())
	bodyView = joinScrollbar(t, bodyView, listHeight, listTotalHeight, listHeight, d.list.Offset())
	rc.AddPart(bodyView)
	rc.Help = renderDialogHelp(t, &d.help, d, innerWidth)

	view := rc.Render()
	d.updateBodyArea(area, view, bodyView, rc.Help, listHeight)
	DrawCenter(scr, area, view)
	return nil
}

func (d *ForkFrom) updateBodyArea(
	area uv.Rectangle,
	view string,
	bodyView string,
	helpView string,
	listHeight int,
) {
	t := d.com.Styles
	viewWidth, viewHeight := lipgloss.Size(view)
	dialogArea := common.CenterRect(area, min(viewWidth, area.Dx()), min(viewHeight, area.Dy()))
	viewStyle := t.Dialog.View
	bodyStyle := t.Dialog.List

	bodyViewTop := dialogArea.Max.Y -
		viewStyle.GetMarginBottom() -
		viewStyle.GetBorderBottomSize() -
		viewStyle.GetPaddingBottom() -
		lipgloss.Height(helpView) -
		lipgloss.Height(bodyView)

	bodyMin := image.Pt(
		dialogArea.Min.X+
			viewStyle.GetMarginLeft()+
			viewStyle.GetBorderLeftSize()+
			viewStyle.GetPaddingLeft()+
			bodyStyle.GetMarginLeft()+
			bodyStyle.GetBorderLeftSize()+
			bodyStyle.GetPaddingLeft(),
		bodyViewTop+
			bodyStyle.GetMarginTop()+
			bodyStyle.GetBorderTopSize()+
			bodyStyle.GetPaddingTop(),
	)

	d.bodyArea = image.Rect(
		bodyMin.X,
		bodyMin.Y,
		bodyMin.X+d.list.Width(),
		bodyMin.Y+listHeight,
	).Intersect(dialogArea).Intersect(area)
}

// ShortHelp implements [help.KeyMap].
func (d *ForkFrom) ShortHelp() []key.Binding {
	return []key.Binding{
		d.keyMap.UpDown,
		d.keyMap.Select,
		d.keyMap.Close,
	}
}

// FullHelp implements [help.KeyMap].
func (d *ForkFrom) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{d.keyMap.Select, d.keyMap.Next, d.keyMap.Previous},
		{d.keyMap.PageUp, d.keyMap.PageDown, d.keyMap.Close},
	}
}

// ForkPointItem is one selectable message row in the fork dialog.
type ForkPointItem struct {
	*list.Versioned
	point   ForkPoint
	t       *styles.Styles
	cache   map[int]string
	focused bool
	m       fuzzy.Match
}

var _ ListItem = (*ForkPointItem)(nil)

// NewForkPointItem creates a fork point list item.
func NewForkPointItem(t *styles.Styles, point ForkPoint) *ForkPointItem {
	return &ForkPointItem{
		Versioned: list.NewVersioned(),
		point:     point,
		t:         t,
	}
}

// Finished implements list.Item.
func (f *ForkPointItem) Finished() bool {
	return true
}

// Filter implements list.FilterableItem.
func (f *ForkPointItem) Filter() string {
	return f.point.Role + " " + f.point.Preview
}

// ID implements ListItem.
func (f *ForkPointItem) ID() string {
	return f.point.MessageID
}

// SetFocused implements list.Focusable.
func (f *ForkPointItem) SetFocused(focused bool) {
	if f.focused == focused {
		return
	}
	f.cache = nil
	f.focused = focused
	if f.Versioned != nil {
		f.Bump()
	}
}

// SetMatch implements list.MatchSettable.
func (f *ForkPointItem) SetMatch(m fuzzy.Match) {
	if sameFuzzyMatch(f.m, m) {
		return
	}
	f.cache = nil
	f.m = m
	if f.Versioned != nil {
		f.Bump()
	}
}

// Render implements list.Item.
func (f *ForkPointItem) Render(width int) string {
	itemStyles := ListItemStyles{
		ItemBlurred:     f.t.Dialog.NormalItem,
		ItemFocused:     f.t.Dialog.SelectedItem,
		InfoTextBlurred: f.t.Dialog.ListItem.InfoBlurred,
		InfoTextFocused: f.t.Dialog.ListItem.InfoFocused,
	}
	preview := f.point.Preview
	if preview == "" {
		preview = "(empty message)"
	}
	title := f.point.Role + "  " + preview
	info := ""
	if f.point.Newest {
		info = "latest"
	}
	return renderItem(itemStyles, title, info, f.focused, false, width, f.cache, &f.m)
}
