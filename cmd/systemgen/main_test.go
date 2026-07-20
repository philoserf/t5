package main

import (
	"testing"

	"github.com/philoserf/t5/internal/clitest"
)

// command runs systemgen end to end; see internal/clitest.
var command = clitest.Command{Name: "systemgen", Main: main}

func TestMain(m *testing.M) { command.TestMain(m) }

// TestMainReportsSeedOnValidRun is #316 for systemgen: reportSeed is an obligation
// the compiler cannot enforce, so this is what fails if the call is ever dropped.
// A system record nobody can regenerate is the bug the seed report exists to
// prevent, and it only works if the seed lands on stderr and not in the records.
func TestMainReportsSeedOnValidRun(t *testing.T) {
	command.Run(t, "-n", "2").AssertReportedSeed(t)
}

// TestMainRejectsBadFlags: systemgen defines no flags of its own, so the input it
// can reject is the shared -n (checked inside SeededRoller, before the roller is
// handed back) and anything flag itself refuses.
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
