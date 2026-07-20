package main

import (
	"testing"

	"github.com/philoserf/t5/internal/clitest"
)

// command runs worldgen end to end; see internal/clitest.
var command = clitest.Command{Name: "worldgen", Main: main}

func TestMain(m *testing.M) { command.TestMain(m) }

// TestMainReportsSeedOnValidRun is #316 for worldgen: reportSeed is an obligation
// the compiler cannot enforce, so this is what fails if the call is ever dropped.
// The seed must reach stderr and stay off the record stream, which is the only
// reason a piped run of worldgen is reproducible at all.
func TestMainReportsSeedOnValidRun(t *testing.T) {
	command.Run(t, "-n", "3").AssertReportedSeed(t)
}

// TestMainRejectsBadFlags: worldgen defines no flags of its own, so the input it
// can reject is the shared -n (checked inside SeededRoller, before the roller is
// handed back) and anything flag itself refuses. Either way: exit 2, the reason on
// stderr, no records, and no seed named for a run that generated nothing.
func TestMainRejectsBadFlags(t *testing.T) {
	cases := map[string][]string{
		"zero count":     {"-n", "0"},
		"negative count": {"-n", "-4"},
		"unknown flag":   {"-nosuchflag"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			command.Run(t, args...).AssertRejected(t)
		})
	}
}

// TestMainSeededRunSaysNothing: with -seed the caller already knows the seed, so
// it is not read back — but the records still arrive on stdout.
func TestMainSeededRunSaysNothing(t *testing.T) {
	res := command.Run(t, "-seed", "1")

	if res.Code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr %q)", res.Code, res.Stderr)
	}

	if res.Stderr != "" {
		t.Errorf("a seeded run has nothing to report, got %q", res.Stderr)
	}

	if res.Stdout == "" {
		t.Error("the world record belongs on stdout, got nothing")
	}
}
