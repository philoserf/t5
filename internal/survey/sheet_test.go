package survey

import (
	"slices"
	"strings"
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/sectorgen"
)

func TestSurveyAt(t *testing.T) {
	sv := Sector(dice.NewWithSeed(3), sectorgen.Dense)
	if len(sv.Records) == 0 {
		t.Fatal("expected a populated sector")
	}

	want := sv.Records[0].Hex

	rec, ok := sv.At(want)
	if !ok || rec.Hex != want {
		t.Errorf("At(%s) = %v,%v, want that record", want, rec.Hex, ok)
	}
	// A hex the sector actually left empty is not found.
	held := map[sectorgen.Hex]bool{}
	for _, rec := range sv.Records {
		held[rec.Hex] = true
	}

	var empty sectorgen.Hex
	for col := 1; col <= sectorgen.Columns && empty == (sectorgen.Hex{}); col++ {
		for row := 1; row <= sectorgen.Rows; row++ {
			if h := (sectorgen.Hex{Col: col, Row: row}); !held[h] {
				empty = h

				break
			}
		}
	}

	if empty == (sectorgen.Hex{}) {
		t.Skip("this sector has a system in every hex")
	}

	if _, found := sv.At(empty); found {
		t.Errorf("At(%s) found a record in an empty hex", empty)
	}
}

func TestParseHexRoundTrip(t *testing.T) {
	h, ok := sectorgen.ParseHex("0436")
	if !ok || h.Col != 4 || h.Row != 36 || h.String() != "0436" {
		t.Errorf("ParseHex(0436) = %v,%v, want col 4 row 36", h, ok)
	}
	// Malformed, out-of-range, and — the trap — signed halves, which strconv would
	// otherwise accept ("+436" silently parsing as hex 0436).
	for _, bad := range []string{"", "436", "04366", "zzzz", "0041", "3341", "0000", "+436", "04+3", "-436", " 436", "04 3"} {
		if _, ok := sectorgen.ParseHex(bad); ok {
			t.Errorf("ParseHex(%q) should be rejected", bad)
		}
	}
}

// TestSheetSurfacesHiddenDetail locks the point of the sheet: it renders the
// derived data the one-line Second Survey record has no room for — Resource
// Units, starport facilities, and native status — none of which any other
// renderer emits.
func TestSheetSurfacesHiddenDetail(t *testing.T) {
	sv := Sector(dice.NewWithSeed(3), sectorgen.Dense)

	sheet := sv.Records[0].Sheet()
	for _, want := range []string{
		"Mainworld", "Extensions", "RU ", "Traffic",
		"Starport", "Travel Zone", "Bases",
		"Stars", "Primary",
		"Orbits",
	} {
		if !strings.Contains(sheet, want) {
			t.Errorf("sheet missing %q:\n%s", want, sheet)
		}
	}
}

// TestSheetRendersOrbitTree confirms the orbit map reaches the bodies the sector
// view never shows: secondary worlds with their own UWPs, and their moons.
func TestSheetRendersOrbitTree(t *testing.T) {
	sv := Sector(dice.NewWithSeed(3), sectorgen.Dense)

	var withMoon, withGiant string

	for _, rec := range sv.Records {
		for _, o := range rec.System.Orbits {
			if len(o.Satellites) > 0 && withMoon == "" {
				withMoon = rec.Sheet()
			}

			if o.Giant != nil && withGiant == "" {
				withGiant = rec.Sheet()
			}
		}

		if withMoon != "" && withGiant != "" {
			break
		}
	}

	if withMoon == "" ||
		!strings.Contains(withMoon, "moon ") && !strings.Contains(withMoon, "Ring") {
		t.Errorf("a system with satellites rendered no moon or ring:\n%s", withMoon)
	}

	if withGiant == "" || !strings.Contains(withGiant, "Gas Giant") {
		t.Errorf("a system with a gas giant did not render it:\n%s", withGiant)
	}
}

// TestSheetMarksCapital shows a capital's title in the header rather than leaving
// the reader to spot the Cx/Cs trade code.
func TestSheetMarksCapital(t *testing.T) {
	sv := Sector(dice.NewWithSeed(3), sectorgen.Dense)
	for _, rec := range sv.Records {
		if slices.Contains(rec.System.Mainworld.TradeCodes, "Cx") {
			if !strings.Contains(rec.Sheet(), "Sector Capital") {
				t.Errorf("the Cx world's sheet does not name it the sector capital")
			}

			return
		}
	}

	t.Skip("no sector capital in this sector")
}
