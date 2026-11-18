package verify

import (
	"testing"

	"github.com/charmbracelet/crush/internal/envreport"
	"github.com/stretchr/testify/require"
)

func TestParseCoverageProfile(t *testing.T) {
	t.Parallel()

	profile := `mode: set
example.com/p/calc.go:3.29,3.46 1 1
example.com/p/calc.go:5.29,5.46 1 1
example.com/p/calc.go:7.29,7.46 1 0
example.com/p/util.go:3.1,5.2 2 0
`
	covered, total, err := parseCoverageProfile([]byte(profile))
	require.NoError(t, err)
	require.Equal(t, 2, covered)
	require.Equal(t, 5, total)
}

func TestParseCoverageProfileRejectsGarbage(t *testing.T) {
	t.Parallel()

	_, _, err := parseCoverageProfile([]byte("example.com/p/calc.go:3.29,3.46 1 1\n"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "mode line")

	_, _, err = parseCoverageProfile([]byte("mode: set\nnot a profile line\n"))
	require.Error(t, err)
}

func TestPlanCoverageForGo(t *testing.T) {
	t.Parallel()

	spec, reason := PlanCoverage([]envreport.Stack{{Name: "Go"}}, "/tmp/bundle")
	require.Empty(t, reason)
	require.Contains(t, spec.command, "go test -coverprofile=")
	require.Contains(t, spec.command, "./...")
	require.Equal(t, "/tmp/bundle/coverage.out", spec.profile)
}

func TestPlanCoverageExplainsMissingProvider(t *testing.T) {
	t.Parallel()

	_, reason := PlanCoverage([]envreport.Stack{{Name: "Node"}, {Name: "Python"}}, "/tmp/bundle")
	require.Contains(t, reason, "no coverage provider")
	require.Contains(t, reason, "Node, Python")

	_, reason = PlanCoverage(nil, "/tmp/bundle")
	require.Contains(t, reason, "no stack was detected")
}

func TestGateCoverage(t *testing.T) {
	t.Parallel()

	available := func(percent float64) *Coverage {
		return &Coverage{Available: true, Percent: percent, Threshold: 80, Passed: percent >= 80}
	}

	// No gate requested: verdicts pass through untouched.
	require.Equal(t, VerdictPass, gateCoverage(VerdictPass, nil))
	// A measured number at or above the threshold keeps the pass.
	require.Equal(t, VerdictPass, gateCoverage(VerdictPass, available(80)))
	require.Equal(t, VerdictPass, gateCoverage(VerdictPass, available(91.2)))
	// Below the threshold is its own verdict, not a plain test failure.
	require.Equal(t, VerdictUnderCovered, gateCoverage(VerdictPass, available(33.3)))
	// Not measurable is explicit rather than a zero or a silent pass.
	require.Equal(t, VerdictCoverageUnavailable, gateCoverage(VerdictPass, &Coverage{Threshold: 80, Reason: "no provider"}))
	// Failing steps are never relabelled by the gate.
	require.Equal(t, VerdictFail, gateCoverage(VerdictFail, available(10)))
	require.Equal(t, VerdictFlaky, gateCoverage(VerdictFlaky, &Coverage{Threshold: 80}))
}
