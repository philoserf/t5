package survey

import (
	"slices"
	"strings"
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/sectorgen"
	"github.com/philoserf/t5/internal/systemgen"
	"github.com/philoserf/t5/internal/uwp"
	"github.com/philoserf/t5/internal/worldgen"
)

// recordWith builds a minimal Record whose mainworld has the given starport and
// importance, for exercising capital selection.
func recordWith(hex sectorgen.Hex, starport byte, ix int) Record {
	return Record{
		Hex: hex,
		System: systemgen.System{
			Mainworld: worldgen.World{
				Profile:    uwp.Profile{Starport: starport},
				Importance: ix,
			},
		},
	}
}

func TestMarkCapital(t *testing.T) {
	records := []Record{
		recordWith(sectorgen.Hex{Col: 1, Row: 1}, 'C', 5), // higher Ix but not Starport A
		recordWith(sectorgen.Hex{Col: 2, Row: 2}, 'A', 2), // Starport A
		recordWith(sectorgen.Hex{Col: 3, Row: 3}, 'A', 4), // Starport A, highest among A -> capital
		recordWith(sectorgen.Hex{Col: 4, Row: 4}, 'B', 9),
	}
	markCapital(records)
	if !slices.Contains(records[2].System.Mainworld.TradeCodes, "Cs") {
		t.Errorf("highest-Ix Starport-A world should be the capital: %+v", records[2].System.Mainworld.TradeCodes)
	}
	for i, rec := range records {
		if i != 2 && slices.Contains(rec.System.Mainworld.TradeCodes, "Cs") {
			t.Errorf("record %d should not be the capital", i)
		}
	}
}

func TestMarkCapitalNoStarportA(t *testing.T) {
	records := []Record{recordWith(sectorgen.Hex{Col: 1, Row: 1}, 'C', 9), recordWith(sectorgen.Hex{Col: 2, Row: 2}, 'B', 8)}
	markCapital(records)
	for _, rec := range records {
		if slices.Contains(rec.System.Mainworld.TradeCodes, "Cs") {
			t.Errorf("no Starport-A world, so no capital should be marked")
		}
	}
}

func TestSubsectorDeterministicAndSurveyed(t *testing.T) {
	a := Subsector(dice.NewWithSeed(7), sectorgen.Sparse, 'A')
	b := Subsector(dice.NewWithSeed(7), sectorgen.Sparse, 'A')
	if len(a) != len(b) || len(a) == 0 {
		t.Fatalf("expected a stable, non-empty survey: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i].SecondSurvey() != b[i].SecondSurvey() {
			t.Fatalf("record %d differs between identical seeds", i)
		}
		// Each line begins with the hex coordinate and carries the world name.
		line := a[i].SecondSurvey()
		if !strings.HasPrefix(line, a[i].Hex.String()+" ") || !strings.Contains(line, a[i].Name) {
			t.Errorf("survey line malformed: %q (hex %s, name %s)", line, a[i].Hex, a[i].Name)
		}
		// Every surveyed hex sits inside subsector A.
		if a[i].Hex.Subsector() != 'A' {
			t.Errorf("hex %s is not in subsector A", a[i].Hex)
		}
	}
}

func TestWorldName(t *testing.T) {
	name := worldName(dice.NewScripted(3, 4, 2, 5, 1, 6))
	if name == "" || name[0] < 'A' || name[0] > 'Z' {
		t.Errorf("worldName = %q, want a capitalized non-empty name", name)
	}
}
