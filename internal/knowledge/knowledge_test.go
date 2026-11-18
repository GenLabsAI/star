package knowledge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddAndGetVerbatim(t *testing.T) {
	t.Parallel()

	store := NewStore(t.TempDir())
	body := "# CI quirk\n\nCI runs with -count=1, so cached results will not hide a failure.\n"
	entry, err := store.Add("ci-quirk", body)
	require.NoError(t, err)
	require.Equal(t, "ci-quirk", entry.Name)
	require.Equal(t, int64(len(body)), entry.Size)
	require.Equal(t, "CI quirk", entry.Summary)

	_, got, err := store.Get("ci-quirk")
	require.NoError(t, err)
	require.Equal(t, body, got, "content must come back exactly as written")
}

func TestAddNormalizesTrailingNewlineOnly(t *testing.T) {
	t.Parallel()

	store := NewStore(t.TempDir())
	_, err := store.Add("note", "line one\nline two\n\n\n")
	require.NoError(t, err)

	_, got, err := store.Get("note")
	require.NoError(t, err)
	require.Equal(t, "line one\nline two\n", got)
}

func TestAddReplacesExistingEntry(t *testing.T) {
	t.Parallel()

	store := NewStore(t.TempDir())
	_, err := store.Add("db", "port 5432")
	require.NoError(t, err)
	_, err = store.Add("db", "port 5433 after the container moved")
	require.NoError(t, err)

	_, got, err := store.Get("db")
	require.NoError(t, err)
	require.Equal(t, "port 5433 after the container moved\n", got)
}

func TestListSortedWithSummaries(t *testing.T) {
	t.Parallel()

	store := NewStore(t.TempDir())
	_, err := store.Add("gamma", "third\nmore")
	require.NoError(t, err)
	_, err = store.Add("alpha", "# First line\nsecond")
	require.NoError(t, err)

	entries, err := store.List()
	require.NoError(t, err)
	require.Len(t, entries, 2)
	require.Equal(t, "alpha", entries[0].Name)
	require.Equal(t, "First line", entries[0].Summary)
	require.Equal(t, "gamma", entries[1].Name)
	require.Equal(t, "third", entries[1].Summary)
}

func TestEmptyStoreListsNothing(t *testing.T) {
	t.Parallel()

	entries, err := NewStore(t.TempDir()).List()
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestGetMissingEntry(t *testing.T) {
	t.Parallel()

	_, _, err := NewStore(t.TempDir()).Get("nope")
	require.Error(t, err)
	require.Contains(t, err.Error(), "No knowledge entry")
}

func TestRejectsPathsAsNames(t *testing.T) {
	t.Parallel()

	store := NewStore(t.TempDir())
	for _, name := range []string{"../escape", "a/b", "", "Upper", "sneaky name"} {
		_, err := store.Add(name, "x")
		require.Error(t, err, "name %q should be rejected", name)
		require.Error(t, ValidateName(name))
	}
}

func TestValidateNameAcceptsSlugs(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"ci-quirk", "db.port", "v2_notes", "a1"} {
		require.NoError(t, ValidateName(name), name)
	}
}

func TestRejectsEmptyAndOversizedBody(t *testing.T) {
	t.Parallel()

	store := NewStore(t.TempDir())
	_, err := store.Add("blank", "   \n\t\n")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty")

	_, err = store.Add("huge", strings.Repeat("x", MaxBody+1))
	require.Error(t, err)
	require.Contains(t, err.Error(), "larger than")
}

func TestRemove(t *testing.T) {
	t.Parallel()

	store := NewStore(t.TempDir())
	_, err := store.Add("temp", "goes away")
	require.NoError(t, err)
	require.NoError(t, store.Remove("temp"))

	_, _, err = store.Get("temp")
	require.Error(t, err)
	require.Error(t, store.Remove("temp"))
}

func TestEntryFilesArePlainMarkdownInTheDataDir(t *testing.T) {
	t.Parallel()

	project := t.TempDir()
	store := NewStore(project)
	_, err := store.Add("quirk", "body")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(project, ".crush", "knowledge"), store.Dir())

	data, err := os.ReadFile(filepath.Join(store.Dir(), "quirk.md"))
	require.NoError(t, err)
	require.Equal(t, "body\n", string(data))
}
