package main

import (
	"testing"

	"github.com/philoserf/t5/internal/clitest"
)

// command runs sophont end to end; see internal/clitest.
var command = clitest.Command{Name: "sophont", Main: main}

func TestMain(m *testing.M) { command.TestMain(m) }

// TestMainReportsSeedOnValidRun is #316 for sophont: reportSeed is an obligation
// the compiler cannot enforce, so this is what fails if the call is ever dropped.
func TestMainReportsSeedOnValidRun(t *testing.T) {
	command.Run(t, "-n", "2").AssertReportedSeed(t)
}

// TestMainReportsSeedWithChar: -char adds a second generator to the same dice
// stream, so it is the other path through main and must report the seed too.
func TestMainReportsSeedWithChar(t *testing.T) {
	command.Run(t, "-char").AssertReportedSeed(t)
}

// TestMainRejectsBadFlags: -char is a bool, so sophont has nothing of its own to
// reject; what remains is the shared -n and anything flag itself refuses.
func TestMainRejectsBadFlags(t *testing.T) {
	cases := map[string][]string{
		"zero count":     {"-n", "0"},
		"negative count": {"-n", "-4"},
		"unknown flag":   {"-nosuchflag"},
		"non-bool -char": {"-char=maybe"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			command.Run(t, args...).AssertRejected(t)
		})
	}
}
