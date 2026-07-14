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
