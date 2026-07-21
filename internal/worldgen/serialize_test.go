package worldgen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

// TestWorldRecordRoundTrip is the #327 property for the world record line: every
// generated world's SecondSurvey line parses back and re-renders byte-identically.
// It exercises the full field set — UWP, trade codes (including the no-code "-"),
// the {Ix}(Ex)[Cx] Extensions, nobility, bases, and zone — across seeds, and the
// belt path, so a codec that dropped or transposed a field fails on some world.
func TestWorldRecordRoundTrip(t *testing.T) {
	for seed := uint64(1); seed <= 300; seed++ {
		for _, w := range []World{
			GenerateWorld(dice.NewWithSeed(seed), 2, 1, false),
			GenerateWorld(dice.NewWithSeed(seed), 0, 0, true), // capital: exercises the capital nobility
			GenerateBeltWorld(dice.NewWithSeed(seed), 1, 1, false),
		} {
			line := w.SecondSurvey()

			got, err := ParseWorld(line)
			if err != nil {
				t.Fatalf("seed %d: ParseWorld(%q) errored: %v", seed, line, err)
			}

			if re := got.SecondSurvey(); re != line {
				t.Fatalf("seed %d: round-trip\n got %q\nwant %q", seed, re, line)
			}
		}
	}
}

// TestWorldRecordRoundTripEdges pins the hand-built edge lines the sweep may not
// reach: the Regina golden, a world with no trade codes, and the Amber/Red zones
// and every base combination.
func TestWorldRecordRoundTripEdges(t *testing.T) {
	for _, line := range []string{
		"A788899-C Ph Pa Ri {+4}(D7E+4)[9C6D] BcCeF NS -", // Regina
		"C539700-8 - {0}(000+0)[0000] B - -",              // no trade codes
		"E410100-7 Lo Co {-3}(500-2)[1139] B - A",         // Amber zone
		"X000000-0 - {0}(000+0)[0000] B NSDW R",           // every base, Red zone
	} {
		got, err := ParseWorld(line)
		if err != nil {
			t.Fatalf("ParseWorld(%q) errored: %v", line, err)
		}

		if re := got.SecondSurvey(); re != line {
			t.Fatalf("round-trip\n got %q\nwant %q", re, line)
		}
	}
}

// TestParseWorldRejectsMalformed: the reader is strict.
func TestParseWorldRejectsMalformed(t *testing.T) {
	for _, s := range []string{
		"",                                       // empty
		"A788899-C",                              // no extensions/nobility/bases/zone
		"Zz88899-C Ph {0}(000+0)[0000] B - -",    // bad UWP
		"A788899-C Zz {0}(000+0)[0000] B - -",    // bad trade code
		"A788899-C - {0}(000+0)[0000] B SN -",    // non-canonical bases (out of order)
		"A788899-C - {0}(000+0)[0000] B Q -",     // bad base letter
		"A788899-C - {0}(000+0)[0000] B - X",     // bad zone
		"A788899-C - (000+0)[0000] B - -",        // missing {Ix}
		"A788899-C - {0}(000+0)[0000] B - - 233", // trailing system fields (world reader only)
	} {
		if _, err := ParseWorld(s); err == nil {
			t.Errorf("ParseWorld(%q) succeeded, want a malformed-record error", s)
		}
	}
}
