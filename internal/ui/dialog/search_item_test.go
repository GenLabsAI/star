package dialog

import (
	"strings"
	"testing"

	"github.com/charmbracelet/crush/internal/search"
	"github.com/charmbracelet/crush/internal/ui/list"
	"github.com/charmbracelet/crush/internal/ui/styles"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

// TestSearchItemRenderStaysOnOneLine guards the search dialog rows against
// the two-line wrapping bug. A tab in a snippet is expanded by the terminal
// but measured as narrower, so it pushed the row past its width; a newline or
// tab in a session title split the row outright.
func TestSearchItemRenderStaysOnOneLine(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	const width = 40

	tests := []struct {
		name   string
		result search.Result
	}{
		{
			name:   "tab in snippet",
			result: search.Result{Snippet: "func\tmain the migration", SessionTitle: "Migration work"},
		},
		{
			name:   "control characters in snippet",
			result: search.Result{Snippet: "the\x07 migration\x0b path", SessionTitle: "Migration work"},
		},
		{
			name:   "newline in title",
			result: search.Result{Snippet: "the migration", SessionTitle: "Migration\nwork"},
		},
		{
			name:   "tab in title",
			result: search.Result{Snippet: "the migration", SessionTitle: "Migration\twork"},
		},
		{
			name:   "ansi snippet",
			result: search.Result{Snippet: "\x1b[31mthe migration\x1b[0m path", SessionTitle: "Migration work"},
		},
		{
			name:   "wide snippet",
			result: search.Result{Snippet: strings.Repeat("日本語", 20), SessionTitle: "Wide"},
		},
		{
			name:   "long plain snippet",
			result: search.Result{Snippet: strings.Repeat("migration ", 20), SessionTitle: "Migration work"},
		},
		{
			name:   "long title and snippet",
			result: search.Result{Snippet: strings.Repeat("migration ", 20), SessionTitle: strings.Repeat("title ", 20)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			item := &SearchItem{
				Versioned: list.NewVersioned(),
				Result:    tt.result,
				t:         &sty,
				cache:     make(map[int]string),
			}
			out := item.Render(width)

			require.Equal(t, 0, strings.Count(out, "\n"), "row must render on a single line")
			require.NotContains(t, ansi.Strip(out), "\t", "row must not contain tabs")
			require.LessOrEqual(t, ansi.StringWidth(out), width, "row must not exceed its width")
		})
	}
}

// TestNewSearchItemsCollapsesIdenticalRows guards that results whose rows
// would read exactly the same are listed once instead of dozens of times.
func TestNewSearchItemsCollapsesIdenticalRows(t *testing.T) {
	t.Parallel()

	sty := styles.CharmtonePantera()
	results := []search.Result{
		{SessionID: "s1", SessionTitle: "Migration work", MessageID: "m1", Snippet: "the {{migration}} path"},
		{SessionID: "s2", SessionTitle: "Fork of Migration work", MessageID: "m2", Snippet: "the {{migration}} path"},
		{SessionID: "s3", SessionTitle: "Migration work", MessageID: "m3", Snippet: "the {{migration}} path"},
		{SessionID: "s4", SessionTitle: "Migration work", MessageID: "m4", Snippet: "another {{migration}}"},
	}

	items := newSearchItems(&sty, results)

	require.Len(t, items, 3)
	require.Equal(t, "m1", items[0].(*SearchItem).Result.MessageID)
	require.Equal(t, "m2", items[1].(*SearchItem).Result.MessageID)
	require.Equal(t, "m4", items[2].(*SearchItem).Result.MessageID)
}
