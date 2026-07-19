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
// the reader to spot the Cs/Cp trade code, and pins each code to its own title:
// the sheet once labelled a Sector Capital "Subsector" after the codes were
// corrected in markSectorCapitals but not in the renderer. "Sector Capital" and
// "Subsector Capital" differ only in case at the seam, so the case-sensitive
// Contains is what tells the two apart.
func TestSheetMarksCapital(t *testing.T) {
	sv := Sector(dice.NewWithSeed(3), sectorgen.Dense)
	titles := map[string]string{"Cs": "Sector Capital", "Cp": "Subsector Capital"}
	seen := map[string]int{}

	var plain string

	for _, rec := range sv.Records {
		code := rec.System.Mainworld.CapitalCode()
		if code == "" {
			if plain == "" {
				plain = rec.Sheet()
			}

			continue
		}

		if slices.Contains(rec.System.Mainworld.TradeCodes, "Cx") {
			t.Errorf("%s carries Cx; the Imperial Capital is not generated", rec.Hex)
		}

		seen[code]++

		want, ok := titles[code]
		if !ok {
			t.Errorf("%s carries unexpected capital code %q", rec.Hex, code)

			continue
		}

		if sheet := rec.Sheet(); !strings.Contains(sheet, want) {
			t.Errorf("the %s world %s does not name itself %q:\n%s", code, rec.Hex, want, sheet)
		}
	}
	// Vacuity guards: a Dense sector holding no Starport-A world at all would be a
	// red flag in its own right, and without these the loop above would pass on a
	// survey that stamped nothing. markSectorCapitals marks exactly one sector
	// capital and one per remaining subsector, so Cs is exactly 1 and Cp at least 1.
	if seen["Cs"] != 1 {
		t.Errorf("sector capitals = %d, want exactly 1", seen["Cs"])
	}

	if seen["Cp"] < 1 {
		t.Errorf("subsector capitals = %d, want at least 1", seen["Cp"])
	}
	// And a world that is no capital claims no title, so the assertions above are
	// not passing on a header the renderer prints for everyone.
	if plain == "" {
		t.Fatal("every world in this sector is a capital")
	}

	for _, title := range titles {
		if strings.Contains(plain, title) {
			t.Errorf("a non-capital world's sheet claims %q:\n%s", title, plain)
		}
	}
}
