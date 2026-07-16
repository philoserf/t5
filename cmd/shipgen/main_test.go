package main

import (
	"testing"

	"github.com/philoserf/t5/internal/shipgen"
)

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
	}
	for name, mutate := range cases {
		f := murphyFlags()
		mutate(&f)
		if _, err := specFromFlags(f); err == nil {
			t.Errorf("%s: should be rejected, got a spec", name)
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
