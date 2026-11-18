package verify

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// writeScript writes an executable shell script into dir.
func writeScript(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(body), 0o755))
	return path
}

// A command that fails once and then passes is flaky, and a flaky run is
// not a pass: the verdict says so, every attempt is recorded, and the
// output of the failing attempt is kept even though a later one passed.
func TestRepeatDetectsFlakyCommand(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	counter := filepath.Join(dir, "counter")
	writeScript(t, dir, "flake.sh", `#!/bin/sh
if [ ! -f "$1" ]; then
  touch "$1"
  echo "flaked on the first run"
  exit 1
fi
echo "passed"
`)

	result, err := Run(context.Background(), dir, []Spec{{Kind: "test", Command: "./flake.sh " + counter}}, Options{
		OutDir: filepath.Join(dir, "bundle"),
		Repeat: 3,
	})
	require.NoError(t, err)

	require.Equal(t, VerdictFlaky, result.Verdict)
	require.True(t, result.Failed(), "a flaky run must not be treated as a pass")
	require.Len(t, result.Steps, 1)

	step := result.Steps[0]
	require.True(t, step.Flaky)
	require.Equal(t, "fail, pass, pass", step.Outcomes())
	require.Len(t, step.Attempts, 3)
	require.Equal(t, []int{1, 0, 0}, []int{
		step.Attempts[0].ExitCode,
		step.Attempts[1].ExitCode,
		step.Attempts[2].ExitCode,
	})
	require.Contains(t, step.OutputTail, "flaked on the first run", "the failing attempt's output must be quoted")

	// Every attempt gets its own log, all of them inside the bundle.
	for i, attempt := range step.Attempts {
		data, err := os.ReadFile(filepath.Join(result.Bundle, filepath.FromSlash(attempt.Log)))
		require.NoError(t, err, "attempt %d", i+1)
		require.Contains(t, string(data), "# exit:", attempt.Log)
	}
}

// A command that fails every time is a failure, not flakiness: repeating
// a run must not turn a broken project into an uncertain one.
func TestRepeatKeepsConsistentFailureFailed(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeScript(t, dir, "broken.sh", `#!/bin/sh
echo "always broken"
exit 2
`)

	result, err := Run(context.Background(), dir, []Spec{{Kind: "test", Command: "./broken.sh"}}, Options{
		OutDir: filepath.Join(dir, "bundle"),
		Repeat: 3,
	})
	require.NoError(t, err)

	require.Equal(t, VerdictFail, result.Verdict)
	require.False(t, result.Steps[0].Flaky)
	require.Equal(t, []int{2, 2, 2}, []int{
		result.Steps[0].Attempts[0].ExitCode,
		result.Steps[0].Attempts[1].ExitCode,
		result.Steps[0].Attempts[2].ExitCode,
	})
	require.Contains(t, result.Steps[0].OutputTail, "always broken")
}

// Repeating a stable passing command is still a pass, and the record
// keeps one attempt list rather than pretending it ran once.
func TestRepeatStablePass(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeScript(t, dir, "ok.sh", "#!/bin/sh\necho fine\n")

	result, err := Run(context.Background(), dir, []Spec{{Kind: "build", Command: "./ok.sh"}}, Options{
		OutDir: filepath.Join(dir, "bundle"),
		Repeat: 2,
	})
	require.NoError(t, err)

	require.Equal(t, VerdictPass, result.Verdict)
	require.False(t, result.Failed())
	require.Equal(t, "pass, pass", result.Steps[0].Outcomes())
	require.Contains(t, SummaryMarkdown(*result), "| build | `./ok.sh` | pass |")
}

// A repeated run has to disable Go's test cache, or the second run would
// replay the first run's verdict and a flaky Go test would look stable.
func TestRepeatDisablesGoTestCache(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeScript(t, dir, "cache.sh", `#!/bin/sh
case "$GOFLAGS" in
  *-count=1*) echo "cache disabled"; exit 0 ;;
esac
echo "GOFLAGS=$GOFLAGS"
exit 1
`)

	repeated, err := Run(context.Background(), dir, []Spec{{Kind: "test", Command: "./cache.sh"}}, Options{
		OutDir: filepath.Join(dir, "bundle"),
		Repeat: 3,
	})
	require.NoError(t, err)
	require.Equal(t, VerdictPass, repeated.Verdict, repeated.Steps[0].OutputTail)

	single, err := Run(context.Background(), dir, []Spec{{Kind: "test", Command: "./cache.sh"}}, Options{
		OutDir: filepath.Join(dir, "single"),
	})
	require.NoError(t, err)
	require.Equal(t, VerdictFail, single.Verdict, "a single run is left in the caller's environment")
	require.Contains(t, single.Steps[0].OutputTail, "GOFLAGS=")
}

// An existing GOFLAGS value is extended, not replaced.
func TestRepeatKeepsExistingGoFlags(t *testing.T) {
	t.Setenv("GOFLAGS", "-mod=mod")

	dir := t.TempDir()
	writeScript(t, dir, "flags.sh", `#!/bin/sh
case "$GOFLAGS" in
  *-mod=mod*-count=1*) exit 0 ;;
esac
echo "GOFLAGS=$GOFLAGS"
exit 1
`)

	result, err := Run(context.Background(), dir, []Spec{{Kind: "test", Command: "./flags.sh"}}, Options{
		OutDir: filepath.Join(dir, "bundle"),
		Repeat: 2,
	})
	require.NoError(t, err)
	require.Equal(t, VerdictPass, result.Verdict, result.Steps[0].OutputTail)
}
