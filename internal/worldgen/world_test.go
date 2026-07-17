package worldgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/uwp"
)

func TestGenerateBeltWorld(t *testing.T) {
	// An asteroid-belt mainworld is Size 0 (so Atmosphere and Hydrographics are
	// 0 too) and carries the As trade code.
	w := GenerateBeltWorld(dice.NewWithSeed(3), 0, 0, false)
	if w.Profile.Size != 0 || w.Profile.Atmosphere != 0 || w.Profile.Hydrographics != 0 {
		t.Errorf("belt world = %s, want Size/Atm/Hyd 0", w.Profile)
	}

	if !slices.Contains(w.TradeCodes, "As") {
		t.Errorf("belt world lacks the As trade code: %v", w.TradeCodes)
	}
	// The belt path consumes the same dice as a normal world (size is rolled then
	// superseded), so a following roll lands identically for both.
	rb, rn := dice.NewWithSeed(3), dice.NewWithSeed(3)
	GenerateBeltWorld(rb, 0, 0, false)
	GenerateWorld(rn, 0, 0, false)

	if rb.Die() != rn.Die() {
		t.Errorf("belt and normal generation consumed different dice counts")
	}
}

func TestSecondSurveyRegina(t *testing.T) {
	w := World{
		Profile:      regina,
		TradeCodes:   []string{"Ph", "Pa", "Ri"},
		Importance:   4,
		Economic:     Economic{Resources: 13, Labor: 7, Infrastructure: 14, Efficiency: 4},
		Cultural:     Cultural{Heterogeneity: 9, Acceptance: 12, Strangeness: 6, Symbols: 13},
		Nobility:     "BcCeF",
		NavalBase:    true,
		ScoutBase:    true,
		Zone:         'G',
		NativeStatus: "Natives",
	}

	want := "A788899-C Ph Pa Ri {+4}(D7E+4)[9C6D] BcCeF NS -"
	if got := w.SecondSurvey(); got != want {
		t.Fatalf("SecondSurvey() =\n%q\nwant\n%q", got, want)
	}
}

func TestSecondSurveyFieldEdges(t *testing.T) {
	// A world with no trade codes and no bases. Every empty field is DASHED, not
	// dropped: this is a positional record, and a collapsed TC column silently
	// shifted the extensions, nobility, bases, and zone one place to the left.
	w := World{
		Profile:  regina,
		Zone:     'R',
		Nobility: "B",
	}

	want := "A788899-C - {0}(000+0)[0000] B - R"
	if got := w.SecondSurvey(); got != want {
		t.Fatalf("SecondSurvey() edges =\n%q\nwant\n%q", got, want)
	}
}

func TestSecondSurveyZeroImportance(t *testing.T) {
	// Ix of zero renders as "{0}", not "{+0}" (the book's Importance table
	// shows a bare 0).
	w := World{Profile: regina, Importance: 0, Nobility: "B"}

	want := "A788899-C - {0}(000+0)[0000] B - -"
	if got := w.SecondSurvey(); got != want {
		t.Fatalf("SecondSurvey() zero Ix =\n%q\nwant\n%q", got, want)
	}
}

func TestSetWayStation(t *testing.T) {
	w := World{Profile: regina, TradeCodes: []string{"Ph"}}
	base := Importance(w.Profile, w.TradeCodes, w.NavalBase, w.ScoutBase, false)
	w.Importance = base

	w.SetWayStation()

	if !w.WayStation {
		t.Errorf("WayStation flag not set")
	}

	if w.Importance != base+1 {
		t.Errorf("Importance = %d, want %d (base + Way Station bonus)", w.Importance, base+1)
	}

	if got := w.bases(); got != "W" {
		t.Errorf("bases() = %q, want W", got)
	}
	// Idempotent: a second call does not bump Importance again.
	w.SetWayStation()

	if w.Importance != base+1 {
		t.Errorf("second SetWayStation re-bumped Importance to %d", w.Importance)
	}
}

func TestPopulationDigit(t *testing.T) {
	if got := PopulationDigit(dice.NewScripted(1, 1), 0); got != 0 {
		t.Errorf("PopulationDigit(pop 0) = %d, want 0", got)
	}

	r := dice.NewWithSeed(1)
	for range 500 {
		if d := PopulationDigit(r, 8); d < 1 || d > 9 {
			t.Fatalf("PopulationDigit(pop 8) = %d, out of 1-9", d)
		}
	}
}

func TestGenerateWorld(t *testing.T) {
	for seed := uint64(1); seed <= 100; seed++ {
		a := GenerateWorld(dice.NewWithSeed(seed), 2, 1, false)
		// Reproducible and internally consistent.
		b := GenerateWorld(dice.NewWithSeed(seed), 2, 1, false)
		if a.SecondSurvey() != b.SecondSurvey() {
			t.Fatalf("seed %d not reproducible", seed)
		}

		if a.Profile.Population > 0 && (a.PopulationDigit < 1 || a.PopulationDigit > 9) {
			t.Fatalf(
				"seed %d: pop digit %d invalid for pop %d",
				seed,
				a.PopulationDigit,
				a.Profile.Population,
			)
		}

		if a.Profile.Population == 0 && a.PopulationDigit != 0 {
			t.Fatalf("seed %d: pop-0 world has pop digit %d", seed, a.PopulationDigit)
		}
	}
}

// TestSecondSurveyNoTradeCodes: a real world can match no trade code at all.
// C539700-8 is one — its atmosphere (3), hydrographics (9), and population (7)
// between them exclude every rule on Chart D. Its record must still have a trade-
// code column, because the record is positional: dropping the column shifted its
// extensions into the trade-code slot, where a reader (or a parser) would read
// "{0}(000+0)[0000]" as this world's trade classifications.
func TestSecondSurveyNoTradeCodes(t *testing.T) {
	p := uwp.Profile{
		Starport: 'C', Size: 5, Atmosphere: 3, Hydrographics: 9,
		Population: 7, Government: 0, Law: 0, TechLevel: 8,
	}
	if tcs := TradeClassifications(p); len(tcs) != 0 {
		t.Fatalf("this test needs a world with no trade codes; %s has %v", p, tcs)
	}

	w := World{Profile: p, Nobility: "B"}
	got := w.SecondSurvey()
	// Six fields, always: UWP, TCs, extensions, nobility, bases, zone.
	if n := len(strings.Fields(got)); n != 6 {
		t.Errorf("record has %d fields, want 6 — a dropped column shifts the rest:\n%s", n, got)
	}

	if !strings.HasPrefix(got, "C539700-8 - {") {
		t.Errorf("an empty trade-code column should be dashed, got %q", got)
	}
}

// TestBaseCodes: the base record codes are N Naval, S Scout, D Naval Depot, W Way
// Station (Book 3 Chart F, p.28). The Depot's code is "D", not the first letter of
// its name — that would collide with the Naval base.
func TestBaseCodes(t *testing.T) {
	w := World{NavalBase: true, ScoutBase: true, NavalDepot: true, WayStation: true}
	if got := w.bases(); got != "NSDW" {
		t.Errorf("bases = %q, want NSDW", got)
	}
	// A lone depot renders "D", not "N".
	depot := World{NavalDepot: true}
	if got := depot.bases(); got != "D" {
		t.Errorf("a lone Naval Depot renders %q, want D", got)
	}

	if names := depot.BaseNames(); len(names) != 1 || names[0] != "Naval Depot" {
		t.Errorf("BaseNames = %v, want [Naval Depot]", names)
	}
}
