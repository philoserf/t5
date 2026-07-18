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
	if got := bestCapital(records, nil); got != 2 {
		t.Errorf("bestCapital = %d, want 2 (the highest-Ix Starport-A world)", got)
	}
	// With no Starport-A world, no capital qualifies.
	none := []Record{
		recordWith(sectorgen.Hex{Col: 1, Row: 1}, 'C', 9),
		recordWith(sectorgen.Hex{Col: 2, Row: 2}, 'B', 8),
	}
	if got := bestCapital(none, nil); got != -1 {
		t.Errorf("bestCapital with no Starport A = %d, want -1", got)
	}
}

func TestMarkSectorCapitals(t *testing.T) {
	records := []Record{
		recordWith(sectorgen.Hex{Col: 1, Row: 1}, 'A', 5), // subsector A, region-best -> Cx
		recordWith(sectorgen.Hex{Col: 2, Row: 2}, 'A', 3), // subsector A, lower -> nothing
		recordWith(sectorgen.Hex{Col: 9, Row: 1}, 'A', 4), // subsector B -> its capital Cp
	}
	markSectorCapitals(records)

	// Book 3 Chart D (p.26): Cp is a Subsector Capital, Cs is a Sector Capital, and
	// Cx is the Imperial Capital. These were swapped, so every generated sector
	// promoted each subsector capital to a sector capital, and its sector capital to
	// the capital of the Imperium.
	tcs := func(i int) []string { return records[i].System.Mainworld.TradeCodes }
	if !slices.Contains(tcs(0), "Cs") || slices.Contains(tcs(0), "Cp") {
		t.Errorf("the sector's best world is its Sector Capital (Cs) only: %v", tcs(0))
	}

	if !slices.Contains(tcs(2), "Cp") {
		t.Errorf("subsector-B's best world is a Subsector Capital (Cp): %v", tcs(2))
	}
	// Cx is the Imperial Capital: there is one, and it is not a fact about a sector.
	for i := range records {
		if slices.Contains(tcs(i), "Cx") {
			t.Errorf("record %d was marked the Imperial Capital (Cx): %v", i, tcs(i))
		}
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
	links := []route.Link{
		{From: a, To: b, Jump: 30},
		{From: b, To: c, Jump: 40},
		{From: a, To: c, Jump: 30},
	}
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

	for i, rec := range a.Records {
		line := rec.SecondSurvey()
		if line != b.Records[i].SecondSurvey() {
			t.Fatalf("record %d differs between identical seeds", i)
		}
		// Each line begins with the hex coordinate and carries the world name.
		if !strings.HasPrefix(line, rec.Hex.String()+" ") || !strings.Contains(line, rec.Name) {
			t.Errorf("survey line malformed: %q (hex %s, name %s)", line, rec.Hex, rec.Name)
		}
	}
}

// TestViewsAgree is the invariant the whole design exists to hold: the views are
// selections from one survey, so a subsector's record and the same hex's record
// are the same world. Surveying a subsector on its own would break this — which
// is why nothing can.
//
// It walks all sixteen subsectors and checks they partition the survey, so it
// cannot quietly become a no-op if any one of them happens to come up empty.
func TestViewsAgree(t *testing.T) {
	sv := Sector(dice.NewWithSeed(7), sectorgen.Sparse)
	if len(sv.Records) == 0 {
		t.Fatal("expected a populated sector")
	}

	selected := 0

	for letter := byte('A'); letter <= 'P'; letter++ {
		for _, rec := range sv.Subsector(letter) {
			selected++

			if rec.Hex.Subsector() != letter {
				t.Errorf(
					"Subsector(%c) returned hex %s, which is in %c",
					letter,
					rec.Hex,
					rec.Hex.Subsector(),
				)
			}
			// The same hex, reached the other way, is the same world.
			byHex, found := sv.At(rec.Hex)
			if !found || byHex.SecondSurvey() != rec.SecondSurvey() {
				t.Errorf("subsector view and hex view disagree at %s:\n %s\n %s",
					rec.Hex, rec.SecondSurvey(), byHex.SecondSurvey())
			}
		}
	}
	// Every record belongs to exactly one subsector, so the selections partition
	// the survey — no world is dropped by the views, and none is double-counted.
	if selected != len(sv.Records) {
		t.Errorf("the sixteen subsectors selected %d of %d records", selected, len(sv.Records))
	}

	// A lower-case letter selects the same subsector; a letter outside A-P selects
	// nothing (callers validate with sectorgen.ParseSubsector first).
	if len(sv.Subsector('a')) != len(sv.Subsector('A')) {
		t.Errorf("Subsector('a') and Subsector('A') disagree")
	}

	if got := sv.Subsector('Q'); got != nil {
		t.Errorf("Subsector('Q') selected %d records, want none", len(got))
	}
}

func TestWorldName(t *testing.T) {
	// A name draws one die for the syllable count and two per syllable, and each
	// of those is an even-distribution index costing more than one die: nine in all.
	name := worldName(dice.NewScripted(3, 4, 2, 5, 1, 6, 3, 4, 2))
	if name == "" || name[0] < 'A' || name[0] > 'Z' {
		t.Errorf("worldName = %q, want a capitalized non-empty name", name)
	}
}
