package survey

import (
	"slices"
	"strings"
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/route"
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

// TestBestCapital: a capital must be a Starport-A world, and among those the most
// Important (Book 3 Chart D p.26) — a higher-Ix world at a lesser port loses.
func TestBestCapital(t *testing.T) {
	records := []Record{
		recordWith(sectorgen.Hex{Col: 1, Row: 1}, 'C', 5), // higher Ix but not Starport A
		recordWith(sectorgen.Hex{Col: 2, Row: 2}, 'A', 2), // Starport A
		recordWith(sectorgen.Hex{Col: 3, Row: 3}, 'A', 4), // Starport A, highest among A
		recordWith(sectorgen.Hex{Col: 4, Row: 4}, 'B', 9),
	}
	if got := bestCapital(records, indices(records)); got != 2 {
		t.Errorf("bestCapital = %d, want 2 (the highest-Ix Starport-A world)", got)
	}
	// With no Starport-A world, no capital qualifies.
	none := []Record{
		recordWith(sectorgen.Hex{Col: 1, Row: 1}, 'C', 9),
		recordWith(sectorgen.Hex{Col: 2, Row: 2}, 'B', 8),
	}
	if got := bestCapital(none, indices(none)); got != -1 {
		t.Errorf("bestCapital with no Starport A = %d, want -1", got)
	}
}

func TestMarkSectorCapitals(t *testing.T) {
	records := []Record{
		recordWith(sectorgen.Hex{Col: 1, Row: 1}, 'A', 5), // subsector A, region-best -> Cx
		recordWith(sectorgen.Hex{Col: 2, Row: 2}, 'A', 3), // subsector A, lower -> nothing
		recordWith(sectorgen.Hex{Col: 9, Row: 1}, 'A', 4), // subsector B -> its capital Cs
	}
	markSectorCapitals(records)

	tcs := func(i int) []string { return records[i].System.Mainworld.TradeCodes }
	if !slices.Contains(tcs(0), "Cx") || slices.Contains(tcs(0), "Cs") {
		t.Errorf("region-best world should be Cx only: %v", tcs(0))
	}
	if !slices.Contains(tcs(2), "Cs") {
		t.Errorf("subsector-B best should be Cs: %v", tcs(2))
	}
	if len(tcs(1)) != 0 {
		t.Errorf("non-capital world should carry no capital code: %v", tcs(1))
	}
}

func TestPlaceWayStations(t *testing.T) {
	a := sectorgen.Hex{Col: 1, Row: 1}
	b := sectorgen.Hex{Col: 2, Row: 1}
	c := sectorgen.Hex{Col: 3, Row: 1}
	records := []Record{recordWith(a, 'A', 4), recordWith(b, 'A', 4), recordWith(c, 'A', 4)}
	// Total route length 100 pc -> 2 Way Stations. All three worlds have degree 2,
	// so the tie breaks to the two lowest-CCRR hubs: A and B.
	links := []route.Link{{From: a, To: b, Jump: 30}, {From: b, To: c, Jump: 40}, {From: a, To: c, Jump: 30}}
	placeWayStations(records, links)

	if !records[0].System.Mainworld.WayStation || !records[1].System.Mainworld.WayStation {
		t.Errorf("expected Way Stations at hubs A and B")
	}
	if records[2].System.Mainworld.WayStation {
		t.Errorf("C should not have a Way Station (only 2 stations for 100 pc)")
	}

	// Under 50 pc of route: no Way Stations.
	short := []Record{recordWith(a, 'A', 4), recordWith(b, 'A', 4)}
	placeWayStations(short, []route.Link{{From: a, To: b, Jump: 4}})
	if short[0].System.Mainworld.WayStation || short[1].System.Mainworld.WayStation {
		t.Errorf("no Way Station expected under %d pc of route", wayStationSpacing)
	}
}

func TestSectorDeterministic(t *testing.T) {
	a := Sector(dice.NewWithSeed(3), sectorgen.Sparse)
	b := Sector(dice.NewWithSeed(3), sectorgen.Sparse)
	if len(a.Records) != len(b.Records) || len(a.Routes) != len(b.Routes) {
		t.Fatalf("non-deterministic: records %d/%d, routes %d/%d",
			len(a.Records), len(b.Records), len(a.Routes), len(b.Routes))
	}
	// Every route endpoint is a surveyed hex.
	present := map[sectorgen.Hex]bool{}
	for _, rec := range a.Records {
		present[rec.Hex] = true
	}
	for _, l := range a.Routes {
		if !present[l.From] || !present[l.To] {
			t.Errorf("route %s-%s references an unsurveyed hex", l.From, l.To)
		}
	}
}

func TestSurveyString(t *testing.T) {
	sv := Sector(dice.NewWithSeed(5), sectorgen.Sparse)
	out := sv.String()
	if len(sv.Records) == 0 {
		t.Fatal("expected a non-empty survey")
	}
	// Every world line carries the per-world traffic annotation.
	if !strings.Contains(out, "/wk]") {
		t.Errorf("survey output missing traffic annotation: %q", out[:min(120, len(out))])
	}
	// The first record's hex appears in the output.
	if !strings.Contains(out, sv.Records[0].Hex.String()) {
		t.Errorf("survey output missing first hex %s", sv.Records[0].Hex)
	}
	// If routes exist they are listed under a Trade Routes heading.
	if len(sv.Routes) > 0 && !strings.Contains(out, "Trade Routes") {
		t.Errorf("routes present but no Trade Routes section rendered")
	}
}

// TestSectorDeterministicAndSurveyed: the same seed yields the same sector, and
// every record renders a well-formed Second Survey line. Views select from one
// sector survey, so this covers what the subsector and single-hex views print.
func TestSectorDeterministicAndSurveyed(t *testing.T) {
	a := Sector(dice.NewWithSeed(7), sectorgen.Sparse)
	b := Sector(dice.NewWithSeed(7), sectorgen.Sparse)
	if len(a.Records) != len(b.Records) || len(a.Records) == 0 {
		t.Fatalf("expected a stable, non-empty survey: %d vs %d", len(a.Records), len(b.Records))
	}
	inA := 0
	for i := range a.Records {
		if a.Records[i].SecondSurvey() != b.Records[i].SecondSurvey() {
			t.Fatalf("record %d differs between identical seeds", i)
		}
		// Each line begins with the hex coordinate and carries the world name.
		line := a.Records[i].SecondSurvey()
		if !strings.HasPrefix(line, a.Records[i].Hex.String()+" ") || !strings.Contains(line, a.Records[i].Name) {
			t.Errorf("survey line malformed: %q (hex %s, name %s)", line, a.Records[i].Hex, a.Records[i].Name)
		}
		if a.Records[i].Hex.Subsector() == 'A' {
			inA++
		}
	}
	// Selecting a subsector out of the sector is how the default view works.
	if inA == 0 {
		t.Errorf("no surveyed hex fell in subsector A")
	}
}

func TestWorldName(t *testing.T) {
	name := worldName(dice.NewScripted(3, 4, 2, 5, 1, 6))
	if name == "" || name[0] < 'A' || name[0] > 'Z' {
		t.Errorf("worldName = %q, want a capitalized non-empty name", name)
	}
}
