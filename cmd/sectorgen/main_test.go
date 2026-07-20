package main

import (
	"testing"

	"github.com/philoserf/t5/internal/clitest"
)

// command runs sectorgen end to end; see internal/clitest.
var command = clitest.Command{Name: "sectorgen", Main: main}

func TestMain(m *testing.M) { command.TestMain(m) }

// TestSelectView: the three views are exclusive and ordered — -hex wins over
// -sector, which wins over the default subsector listing.
func TestSelectView(t *testing.T) {
	cases := map[string]struct {
		hex       string
		sector    bool
		subsector string
	}{
		"hex beats sector":    {hex: "0436", sector: true, subsector: "A"},
		"hex beats subsector": {hex: "0436", subsector: "A"},
		"sector":              {sector: true, subsector: "A"},
		"subsector default":   {subsector: "P"},
	}
	for name, c := range cases {
		if v, err := selectView(c.hex, c.sector, c.subsector); err != nil || v == nil {
			t.Errorf("%s: got view %v, err %v; want a view", name, v, err)
		}
	}
}

// TestSelectViewRejects: a view's own argument is checked when the view is
// selected — before the survey runs, which is what lets the seed report wait
// until the whole command line is known good (#293).
func TestSelectViewRejects(t *testing.T) {
	cases := map[string]struct {
		hex       string
		sector    bool
		subsector string
	}{
		"hex out of range":   {hex: "9999"},
		"hex not four digit": {hex: "43"},
		"hex not numeric":    {hex: "abcd"},
		"subsector Q":        {subsector: "Q"},
		"subsector empty":    {subsector: ""},
		"subsector word":     {subsector: "AA"},
	}
	for name, c := range cases {
		if _, err := selectView(c.hex, c.sector, c.subsector); err == nil {
			t.Errorf("%s: should be rejected, got a view", name)
		}
	}
}

// TestMainSeedFollowsValidation is the end-to-end contract of #293: an unseeded
// run names the seed it drew only once its own flags check out. A rejected flag
// exits 2 with its reason on stderr and no seed line — the seed would name a run
// that surveyed nothing — and no records on stdout.
func TestMainSeedFollowsValidation(t *testing.T) {
	bad := map[string][]string{
		"unknown density":   {"-density", "qqq"},
		"invalid subsector": {"-subsector", "Q"},
		"invalid hex":       {"-hex", "9999"},
	}
	for name, args := range bad {
		t.Run(name, func(t *testing.T) {
			command.Run(t, args...).AssertRejected(t)
		})
	}

	// The other half: a good command line still names its drawn seed, on stderr,
	// with the records alone on stdout.
	t.Run("valid run reports its seed", func(t *testing.T) {
		command.Run(t, "-subsector", "A").AssertReportedSeed(t)
	})
}
