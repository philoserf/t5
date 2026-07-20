package main

import (
	"strings"
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

// TestMainRejectsFlagsTheViewIgnores covers the losing view's flag. The views are
// exclusive, so a flag belonging to a view that did not win is input this run
// cannot honor — and discarding it silently is worse here than it was in shipgen
// (#315), because "-subsector Q" is a value sectorgen's OWN validator rejects on
// the default path. It printed a hex at exit 0 while swallowing input it would
// otherwise refuse.
func TestMainRejectsFlagsTheViewIgnores(t *testing.T) {
	cases := map[string][]string{
		"subsector under hex":    {"-hex", "0436", "-subsector", "Q"},
		"sector under hex":       {"-hex", "0436", "-sector"},
		"subsector under sector": {"-sector", "-subsector", "A"},
	}

	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			command.Run(t, append(args, "-seed", "1")...).AssertRejected(t)
		})
	}
}

// TestMainAcceptsTheWinningView is the control: the same views run clean when
// nothing else is set, so the rejection above is about the discarded flag and not
// about the view itself.
//
// -sector and -subsector always survey worlds, so they assert the full good-run
// contract: records on stdout, the drawn seed on stderr and only there.
//
// -hex asserts the exit code alone, and only -hex. A hex holding no system is a
// true-but-empty result — cli.Notef says so on stderr and exits 0 with nothing on
// the record stream — so requiring records there made the test depend on whether a
// randomly drawn seed happened to populate that one hex, and it failed about a
// third of the time. Weakening the other two to match would have thrown away the
// stdout/stderr split this harness exists to enforce.
func TestMainAcceptsTheWinningView(t *testing.T) {
	for name, args := range map[string][]string{
		"sector":    {"-sector"},
		"subsector": {"-subsector", "A"},
	} {
		t.Run(name, func(t *testing.T) {
			command.Run(t, args...).AssertReportedSeed(t)
		})
	}

	t.Run("hex", func(t *testing.T) {
		res := command.Run(t, "-hex", "0436")
		if res.Code != 0 {
			t.Errorf("exit code = %d, want 0 (stderr %q)", res.Code, res.Stderr)
		}

		if strings.Contains(res.Stderr, "panic:") {
			t.Errorf("panicked: %q", res.Stderr)
		}
	})
}
