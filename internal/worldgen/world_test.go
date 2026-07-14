package worldgen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

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
	// A world with no trade codes and no bases: the TC slot collapses, and the
	// bases slot shows a dash. Zero Extensions render as "{0}(000+0)[0000]".
	w := World{
		Profile:  regina,
		Zone:     'R',
		Nobility: "B",
	}
	want := "A788899-C {0}(000+0)[0000] B - R"
	if got := w.SecondSurvey(); got != want {
		t.Fatalf("SecondSurvey() edges =\n%q\nwant\n%q", got, want)
	}
}

func TestSecondSurveyZeroImportance(t *testing.T) {
	// Ix of zero renders as "{0}", not "{+0}" (the book's Importance table
	// shows a bare 0).
	w := World{Profile: regina, Importance: 0, Nobility: "B"}
	want := "A788899-C {0}(000+0)[0000] B - -"
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
			t.Fatalf("seed %d: pop digit %d invalid for pop %d", seed, a.PopulationDigit, a.Profile.Population)
		}
		if a.Profile.Population == 0 && a.PopulationDigit != 0 {
			t.Fatalf("seed %d: pop-0 world has pop digit %d", seed, a.PopulationDigit)
		}
	}
}
