package survey

import (
	"slices"
	"testing"

	"github.com/philoserf/t5/internal/sectorgen"
	"github.com/philoserf/t5/internal/systemgen"
	"github.com/philoserf/t5/internal/tradecode"
	"github.com/philoserf/t5/internal/uwp"
	"github.com/philoserf/t5/internal/worldgen"
)

func TestColonyProfile(t *testing.T) {
	for _, c := range []struct {
		name          string
		pop, gov, law int
		want          bool
	}{
		{"lower bound", 5, 6, 0, true},
		{"upper bound", 10, 6, 3, true},
		{"population low", 4, 6, 3, false},
		{"population high", 11, 6, 3, false},
		{"government", 7, 5, 3, false},
		{"law", 7, 6, 4, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			w := surveyWorld(0, c.pop, c.gov, c.law, 8)
			if got := colonyProfile(w); got != c.want {
				t.Errorf("colonyProfile(Pop %d Gov %d Law %d) = %v, want %v",
					c.pop, c.gov, c.law, got, c.want)
			}
		})
	}
}

func TestMarkColoniesChoosesCanonicalOwner(t *testing.T) {
	colony := Record{Hex: sectorgen.Hex{Col: 4, Row: 4}, System: systemgen.System{
		Mainworld: surveyWorld(0, 7, 6, 2, 8),
	}}
	records := []Record{
		colony,
		// Importance wins over Population and TL.
		{Hex: sectorgen.Hex{Col: 1, Row: 1}, System: systemgen.System{
			Mainworld: surveyWorld(3, 5, 0, 0, 5),
		}},
		{Hex: sectorgen.Hex{Col: 2, Row: 2}, System: systemgen.System{
			Mainworld: surveyWorld(4, 4, 0, 0, 4),
		}},
		// With equal Importance, Population wins; with equal Population, TL wins.
		{Hex: sectorgen.Hex{Col: 3, Row: 3}, System: systemgen.System{
			Mainworld: surveyWorld(4, 8, 0, 0, 7),
		}},
		{Hex: sectorgen.Hex{Col: 5, Row: 5}, System: systemgen.System{
			Mainworld: surveyWorld(4, 8, 0, 0, 12),
		}},
	}

	got := markColonies(records)
	if len(got) != 1 || got[0] != (Ownership{
		Colony: sectorgen.Hex{Col: 4, Row: 4},
		Owner:  sectorgen.Hex{Col: 5, Row: 5},
	}) {
		t.Fatalf("ownerships = %+v, want 0404 owned by 0505", got)
	}

	if !slices.Contains(records[0].System.Mainworld.TradeCodes, tradecode.Cy) {
		t.Errorf("colony was not stamped Cy: %v", records[0].System.Mainworld.TradeCodes)
	}
}

func TestColonyOwnerRangeAndTie(t *testing.T) {
	records := []Record{
		{Hex: sectorgen.Hex{Col: 10, Row: 10}, System: systemgen.System{
			Mainworld: surveyWorld(0, 7, 6, 2, 8),
		}},
		{Hex: sectorgen.Hex{Col: 9, Row: 10}, System: systemgen.System{
			Mainworld: surveyWorld(2, 8, 0, 0, 10),
		}},
		{Hex: sectorgen.Hex{Col: 11, Row: 10}, System: systemgen.System{
			Mainworld: surveyWorld(2, 8, 0, 0, 10),
		}},
		{Hex: sectorgen.Hex{Col: 20, Row: 20}, System: systemgen.System{
			Mainworld: surveyWorld(9, 15, 0, 0, 15),
		}},
	}

	if got := colonyOwner(records, 0); got != 1 {
		t.Errorf("equal owners chose record %d (%s), want lower CCRR record 1 (0910)", got, records[got].Hex)
	}
}

func surveyWorld(ix, pop, gov, law, tl int) worldgen.World {
	return worldgen.NewWorld(worldgen.World{Profile: uwp.Profile{
		Starport: 'C', Size: 5, Atmosphere: 5, Hydrographics: 5,
		Population: pop, Government: gov, Law: law, TechLevel: tl,
	}}, ix, worldgen.Economic{}, worldgen.Cultural{}, "B")
}
