package changes

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func gitCmd(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Star Test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Star Test",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	return string(out)
}

func initRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()
	gitCmd(t, dir, "init", "-q", "-b", "main")
	return dir
}

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func TestCollectOutsideRepository(t *testing.T) {
	t.Parallel()

	summary, err := Collect(t.TempDir())
	require.NoError(t, err)
	require.False(t, summary.Repository)
	require.Empty(t, summary.Files)
}

func TestCollectCleanRepository(t *testing.T) {
	t.Parallel()

	dir := initRepo(t)
	write(t, dir, "README.md", "hello\n")
	gitCmd(t, dir, "add", ".")
	gitCmd(t, dir, "commit", "-q", "-m", "init")

	summary, err := Collect(dir)
	require.NoError(t, err)
	require.True(t, summary.Repository)
	require.Equal(t, "main", summary.Branch)
	require.True(t, summary.Clean())
	require.Zero(t, summary.PatchBytes)
}

func TestCollectModifiedAndUntracked(t *testing.T) {
	t.Parallel()

	dir := initRepo(t)
	write(t, dir, "app.go", "package app\n")
	gitCmd(t, dir, "add", ".")
	gitCmd(t, dir, "commit", "-q", "-m", "init")

	write(t, dir, "app.go", "package app\n\nfunc Run() {}\n")
	write(t, dir, "notes.txt", "scratch\n")

	summary, err := Collect(dir)
	require.NoError(t, err)
	require.False(t, summary.Clean())
	require.Len(t, summary.Files, 2)
	require.Contains(t, summary.Files, File{Path: "app.go", Status: " M"})
	require.Contains(t, summary.Files, File{Path: "notes.txt", Status: "??"})
	require.Contains(t, summary.Stat, "app.go")
	require.Positive(t, summary.PatchBytes)

	patch, err := Diff(dir)
	require.NoError(t, err)
	require.Contains(t, string(patch), "func Run() {}")
	// Untracked files are listed but are not part of a git diff.
	require.NotContains(t, string(patch), "scratch")
}

func TestCollectStagedAddition(t *testing.T) {
	t.Parallel()

	dir := initRepo(t)
	write(t, dir, "one.txt", "1\n")
	gitCmd(t, dir, "add", ".")
	gitCmd(t, dir, "commit", "-q", "-m", "init")

	write(t, dir, "two.txt", "2\n")
	gitCmd(t, dir, "add", "two.txt")

	summary, err := Collect(dir)
	require.NoError(t, err)
	require.Equal(t, []File{{Path: "two.txt", Status: "A "}}, summary.Files)
}

func TestCollectWithIgnoresPath(t *testing.T) {
	t.Parallel()

	dir := initRepo(t)
	write(t, dir, "app.go", "package app\n")
	gitCmd(t, dir, "add", ".")
	gitCmd(t, dir, "commit", "-q", "-m", "init")

	write(t, dir, "app.go", "package app\n\nfunc Run() {}\n")
	write(t, dir, ".crush/verify/logs/01-build.log", "ran\n")
	write(t, dir, "notes.txt", "scratch\n")

	summary, err := CollectWith(dir, Options{Ignore: []string{filepath.Join(dir, ".crush")}})
	require.NoError(t, err)
	require.Equal(t, []File{
		{Path: "app.go", Status: " M"},
		{Path: "notes.txt", Status: "??"},
	}, summary.Files)
}

func TestParsePorcelainHandlesRenames(t *testing.T) {
	t.Parallel()

	// "R  new\0old\0" and a plain modified entry.
	data := "R  new.txt\x00old.txt\x00 M app.go\x00?? extra.txt\x00"
	files := parsePorcelain(data, "", nil)

	require.Equal(t, []File{
		{Path: "new.txt", Status: "R ", Original: "old.txt"},
		{Path: "app.go", Status: " M"},
		{Path: "extra.txt", Status: "??"},
	}, files)
}

func TestParsePorcelainIgnoresEmptyEntries(t *testing.T) {
	t.Parallel()

	require.Empty(t, parsePorcelain("", "", nil))
	require.Empty(t, parsePorcelain("\x00", "", nil))
}

func TestSummaryClean(t *testing.T) {
	t.Parallel()

	require.True(t, Summary{}.Clean())
	require.False(t, Summary{Files: []File{{Path: "a", Status: " M"}}}.Clean())
}
