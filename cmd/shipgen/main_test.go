package main

import (
	"strings"
	"testing"

	"github.com/philoserf/t5/internal/clitest"
	"github.com/philoserf/t5/internal/shipgen"
)

// command runs shipgen end to end; see internal/clitest.
var command = clitest.Command{Name: "shipgen", Main: main}

func TestMain(m *testing.M) { command.TestMain(m) }

// murphyFlags are the Murphy-class Scout's design flags (Book 2 pp.42-43), the
// package's golden ship: 100t Hull-A, TL-12, Lifting Body, Maneuver-A + Jump-A +
// Power-A. `shipgen -hull A -tl 12 -config L` must reach exactly this spec.
func murphyFlags() flags {
	return flags{
		hull: "A", tl: 12, config: "L", structure: "shell", armor: 2,
		maneuver: "A", jump: "A", power: "A", mission: "S",
	}
}

func TestSpecFromFlags(t *testing.T) {
	spec, err := specFromFlags(murphyFlags())
	if err != nil {
		t.Fatalf("the Murphy's own flags were rejected: %v", err)
	}

	if spec.HullLetter != 1 || spec.TL != 12 || spec.ArmorLayers != 2 {
		t.Errorf(
			"hull/TL/armor = %d/%d/%d, want 1/12/2",
			spec.HullLetter,
			spec.TL,
			spec.ArmorLayers,
		)
	}

	if spec.Maneuver == nil || spec.Jump == nil || spec.Power == nil {
		t.Fatalf(
			"all three drives were requested, got %+v/%+v/%+v",
			spec.Maneuver,
			spec.Jump,
			spec.Power,
		)
	}

	if spec.Maneuver.Letter != 1 {
		t.Errorf("Maneuver-A ordinal = %d, want 1", spec.Maneuver.Letter)
	}
	// End to end: the flags a user actually types must design the golden ship.
	if qsp := shipgen.Design(spec).QSP(); qsp != "S-AL22" {
		t.Errorf("QSP = %q, want the Murphy's S-AL22", qsp)
	}
}

// TestSpecFromFlagsOmittedDrives: a blank drive letter means the ship has no such
// drive (nil), not a size-A one — a non-starship has no jump drive at all.
func TestSpecFromFlagsOmittedDrives(t *testing.T) {
	f := murphyFlags()
	f.jump = ""

	spec, err := specFromFlags(f)
	if err != nil {
		t.Fatalf("a blank jump drive is legal, got %v", err)
	}

	if spec.Jump != nil {
		t.Errorf("blank -jump should give no jump drive, got %+v", spec.Jump)
	}

	if spec.Maneuver == nil {
		t.Errorf("the other drives should survive an omitted jump drive")
	}

	// All three are independently omittable — blank is the only "none".
	for name, blank := range map[string]func(*flags){
		"maneuver": func(f *flags) { f.maneuver = "" },
		"jump":     func(f *flags) { f.jump = "" },
		"power":    func(f *flags) { f.power = "" },
	} {
		g := murphyFlags()
		blank(&g)

		if _, err := specFromFlags(g); err != nil {
			t.Errorf("a blank %s drive is legal, got %v", name, err)
		}
	}
}

// TestSpecFromFlagsRejects: every bad value is rejected rather than silently
// coerced. The command turns each of these into a stderr message and exit 2.
func TestSpecFromFlagsRejects(t *testing.T) {
	cases := map[string]func(*flags){
		"empty hull":      func(f *flags) { f.hull = "" },
		"non-letter hull": func(f *flags) { f.hull = "?" },
		"multi-char hull": func(f *flags) { f.hull = "AB" },
		// I and O are not eHex letters: "I" must not slide to some other hull.
		"hull I":            func(f *flags) { f.hull = "I" },
		"hull O":            func(f *flags) { f.hull = "O" },
		"unknown config":    func(f *flags) { f.config = "Q" },
		"unknown structure": func(f *flags) { f.structure = "adamantium" },
		// A bad drive letter is a typo, not a request for no drive: only a blank
		// means "none" (see TestSpecFromFlagsOmittedDrives). Silently dropping the
		// drive would hand back a driveless ship and exit 0.
		"non-letter maneuver": func(f *flags) { f.maneuver = "3" },
		"multi-char jump":     func(f *flags) { f.jump = "AB" },
		"jump I":              func(f *flags) { f.jump = "I" },
		"power O":             func(f *flags) { f.power = "O" },
		"non-letter power":    func(f *flags) { f.power = "?" },
		// Layer 1 is integral to the hull, so the library floors a low count — but
		// a typed 0 or negative is a mistake, not a request for the minimum.
		"zero armor":     func(f *flags) { f.armor = 0 },
		"negative armor": func(f *flags) { f.armor = -3 },
		// A TL outside the design system's range is not a Tech Level: it renders
		// "TL--5" and a malformed QSP on a record stream reserved for real ships.
		"negative TL":  func(f *flags) { f.tl = -5 },
		"TL above max": func(f *flags) { f.tl = shipgen.MaxTL + 1 },
		// The QSP mission is one to three letters (Book 2 p.51), so a freeform
		// string must not ride into the profile as "not-a-code-AS22".
		"hyphenated mission": func(f *flags) { f.mission = "not-a-code" },
		"empty mission":      func(f *flags) { f.mission = "" },
		"overlong mission":   func(f *flags) { f.mission = "ABCD" },
		"numeric mission":    func(f *flags) { f.mission = "1" },
		"leading digit":      func(f *flags) { f.mission = "2A" },
		"digit past two":     func(f *flags) { f.mission = "A4" },
		"digit in a triple":  func(f *flags) { f.mission = "AB2" },
	}
	for name, mutate := range cases {
		f := murphyFlags()
		mutate(&f)

		if _, err := specFromFlags(f); err == nil {
			t.Errorf("%s: should be rejected, got a spec", name)
		}
	}
}

// TestSpecFromFlagsTLBounds: the whole legal range is accepted, not just the
// values the goldens happen to use — 0 is a real Tech Level and 21 is the design
// system's ceiling (Book 2 p.51), so neither edge may be rejected.
func TestSpecFromFlagsTLBounds(t *testing.T) {
	for tl := shipgen.MinTL; tl <= shipgen.MaxTL; tl++ {
		f := murphyFlags()
		f.tl = tl

		spec, err := specFromFlags(f)
		if err != nil {
			t.Errorf("TL %d is in range, got %v", tl, err)

			continue
		}

		if spec.TL != tl {
			t.Errorf("TL %d survived as %d", tl, spec.TL)
		}
	}
}

// TestMissionCode: the QSP mission's shape is checked, its meaning is not. The
// book composes codes from Service/Activity/Type/Qualifier and says their
// "actual meanings are subject to common sense, and may be ambiguous", so there
// is no closed vocabulary — any well-shaped code a referee invents is accepted.
func TestMissionCode(t *testing.T) {
	cases := map[string]string{
		"S":   "S",   // the Murphy Scout's own code
		"X":   "X",   // the flag default
		"C":   "C",   // Cruiser
		"SDB": "SDB", // System Defense Boat: Modifier-Modifier-Mission
		"A2":  "A2",  // the digit shorthand for AA
		"A3":  "A3",  // ...and for AAA
		"s":   "S",   // case-insensitive, like the hull and drive letters
		"sdb": "SDB",
	}
	for in, want := range cases {
		got, err := missionCode(in)
		if err != nil {
			t.Errorf("missionCode(%q) = %v, want %q", in, err, want)

			continue
		}

		if got != want {
			t.Errorf("missionCode(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestLetterOrdinal: hull and drive letters are the eHex letters used as an
// ordinal 1..24 — A=1 (Hull A = 100t) through Z=24, skipping I and O. This is
// NOT an eHex value (where A=10), which is the easy way to get it wrong.
func TestLetterOrdinal(t *testing.T) {
	cases := map[string]int{
		"A": 1, "B": 2, "H": 8,
		"J": 9,  // I is skipped, so J follows H
		"N": 13, // ...and O is skipped after N
		"P": 14,
		"Z": 24, // the last hull, 2400t
		"a": 1,  // letters are case-insensitive
		"z": 24,
		// I and O are not eHex letters, so they name no hull — they must be
		// rejected outright, not slid onto a neighboring size.
		"I": 0, "O": 0,
		"": 0, "?": 0, "AB": 0,
	}
	for s, want := range cases {
		if got := letterOrdinal(s); got != want {
			t.Errorf("letterOrdinal(%q) = %d, want %d", s, got, want)
		}
	}
}

// TestMainRejectsBadFlags is the end-to-end half of TestSpecFromFlagsRejects: a
// rejected flag must reach the caller as the shared CLI contract says — the
// message on stderr, nothing on stdout (a piped record stream stays clean), exit
// cli.FailureCode, and no seed named for a run that designed nothing (#293).
func TestMainRejectsBadFlags(t *testing.T) {
	cases := map[string][]string{
		"invalid jump drive":     {"-jump", "I"},
		"multi-char jump drive":  {"-jump", "AB"},
		"numeric maneuver drive": {"-maneuver", "3"},
		"zero armor":             {"-armor", "0"},
		"negative armor":         {"-armor", "-3"},
		// #299's own reproduction: this printed a "TL--5" ship card and the
		// malformed QSP "X-AS00" to stdout at exit 0.
		"negative tech level": {"-tl", "-5"},
		"tech level past max": {"-tl", "22"},
		"freeform mission":    {"-mission", "not-a-code"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			// -hull puts these on the design path, which is where they are read.
			command.Run(t, append([]string{"-hull", "A"}, args...)...).AssertRejected(t)
		})
	}
}

// TestMainRejectsDesignFlagsWithoutHull is #315: without -hull the ship is rolled,
// not designed, so a design flag has nothing to act on. Each of the ten used to be
// discarded in silence — "-tl 99" printed a well-formed TL-14 ship at exit 0 — and
// must now be refused like any other input the command cannot honor.
func TestMainRejectsDesignFlagsWithoutHull(t *testing.T) {
	cases := map[string][]string{
		"tech level": {"-tl", "99"},
		// A design flag set to its own default value is still input the caller
		// typed and the random path still cannot honor, so "was it set" is asked
		// of flag.Visit, not of the value.
		"tech level at the flag default": {"-tl", "12"},
		// ...and 0 is a legal Tech Level, so it is not "unset" by any reading.
		"tech level zero": {"-tl", "0"},
		"config":          {"-config", "L"},
		"structure":       {"-structure", "shell"},
		"armor":           {"-armor", "2"},
		"maneuver drive":  {"-maneuver", "A"},
		"jump drive":      {"-jump", "A"},
		"power plant":     {"-power", "A"},
		"mission":         {"-mission", "S"},
		"weapon":          {"-weapon", "beamlaser"},
		"defense":         {"-defense", "blackglobe"},
		// The message should name the flags, so several at once still resolves.
		"several at once": {"-tl", "99", "-config", "L", "-armor", "2"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			command.Run(t, args...).AssertRejected(t)
		})
	}
}

// TestMainRandomShipStillRuns: the fix rejects design flags, not the random path
// itself — with only the shared flags, an unseeded run still rolls a ship and
// names the seed that would reproduce it.
func TestMainRandomShipStillRuns(t *testing.T) {
	res := command.Run(t, "-n", "2")
	res.AssertReportedSeed(t)

	if !strings.Contains(res.Stdout, "Hull:") {
		t.Errorf("the ship record belongs on stdout, got %q", res.Stdout)
	}
}

// TestMainReportsSeedOnValidRun is the other half of #293: a good command line
// still names its drawn seed, on stderr, leaving the ship record alone on stdout.
func TestMainReportsSeedOnValidRun(t *testing.T) {
	res := command.Run(t, "-hull", "A")
	res.AssertReportedSeed(t)

	if !strings.Contains(res.Stdout, "Hull:") {
		t.Errorf("the ship record belongs on stdout, got %q", res.Stdout)
	}
}
