package worldgen

import (
	"reflect"
	"slices"
	"testing"

	"github.com/philoserf/t5/internal/tradecode"
	"github.com/philoserf/t5/internal/uwp"
)

// TestTradeClassificationsRegina checks the canonical worked example: Regina
// (A788899-C) is Pre-High, Pre-Agricultural, and Rich.
func TestTradeClassificationsRegina(t *testing.T) {
	reginaProfile := uwp.Profile{
		Starport:      'A',
		Size:          7,
		Atmosphere:    8,
		Hydrographics: 8,
		Population:    8,
		Government:    9,
		Law:           9,
		TechLevel:     12,
	}
	got := TradeClassifications(reginaProfile)

	want := []tradecode.Code{"Ph", "Pa", "Ri"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("TradeClassifications(Regina) = %v, want %v", got, want)
	}
}

func TestTradeClassificationsOutOfRangeSafe(t *testing.T) {
	// A hand-built Profile with an out-of-range value must not panic; the
	// value simply matches no constrained classification.
	got := TradeClassifications(
		uwp.Profile{Size: 7, Atmosphere: 8, Hydrographics: 8, Population: 40},
	)
	for _, code := range got {
		if code == "Ph" || code == "Hi" {
			t.Fatalf("out-of-range population matched a population class: %v", got)
		}
	}
}

func TestTradeClassifications(t *testing.T) {
	cases := []struct {
		name string
		p    uwp.Profile
		want []tradecode.Code
	}{
		{
			"asteroid belt",
			uwp.Profile{Size: 0, Atmosphere: 0, Hydrographics: 0, Population: 0},
			// As is a Chart-D UWP code (Siz0/Atm0/Hyd0, Book 3 p.26/p.16), correct for
			// a mainworld and a genuine belt. A secondary Size-0 Worldlet shares this
			// profile without being a belt, so TradeClassificationsWithContext strips
			// As there (#324); the pure classifier still emits it. Pop 0 matches no
			// population class.
			[]tradecode.Code{"As", "Va"},
		},
		{
			"vacuum low-pop rock",
			uwp.Profile{Size: 4, Atmosphere: 0, Hydrographics: 0, Population: 2},
			// Atm 0 -> Va and De needs Atm 2-9 (no); Pop 2 -> Lo.
			[]tradecode.Code{"Va", "Lo"},
		},
		{
			"water world",
			uwp.Profile{Size: 6, Atmosphere: 6, Hydrographics: 10, Population: 5},
			// Siz6/Atm6/HydA -> Wa; Pop5 -> Ni; Atm6/Pop5 -> Pr (Pre-Rich).
			[]tradecode.Code{"Wa", "Ni", "Pr"},
		},
		{
			"garden agricultural",
			uwp.Profile{Size: 7, Atmosphere: 6, Hydrographics: 6, Population: 6},
			// Siz7/Atm6/Hyd6 -> Ga; Pop6 -> Ni; Atm6/Hyd6/Pop6 -> Ag; Atm6/Pop6 -> Ri.
			[]tradecode.Code{"Ga", "Ni", "Ag", "Ri"},
		},
	}
	for _, c := range cases {
		if got := TradeClassifications(c.p); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: TradeClassifications = %v, want %v", c.name, got, c.want)
		}
	}
}

// TestTradeClassificationsUWPCodes covers the Gov/Law/Starport-gated codes added
// beyond the original 21 (Book 3 Chart D, p.26): Dieback, Barren, Prison/Exile,
// and Reserve.
func TestTradeClassificationsUWPCodes(t *testing.T) {
	// Di Dieback: Pop0/Gov0/Law0 with a working starport (A-D).
	di := uwp.Profile{
		Starport:      'C',
		Size:          5,
		Atmosphere:    4,
		Hydrographics: 3,
		Population:    0,
		Government:    0,
		Law:           0,
	}
	if got := TradeClassifications(di); !slices.Contains(got, "Di") || slices.Contains(got, "Ba") {
		t.Errorf("Dieback = %v, want Di and not Ba", got)
	}
	// Ba Barren: the same core with Starport E or X.
	for _, sp := range []byte{'E', 'X'} {
		ba := uwp.Profile{
			Starport:      sp,
			Size:          5,
			Atmosphere:    4,
			Hydrographics: 3,
			Population:    0,
			Government:    0,
			Law:           0,
		}
		if got := TradeClassifications(ba); !slices.Contains(got, "Ba") || slices.Contains(got, "Di") {
			t.Errorf("Barren (starport %c) = %v, want Ba and not Di", sp, got)
		}
	}
	// Di/Ba are mutually exclusive across every starport (and the belt's 0x00).
	for _, sp := range []byte{'A', 'B', 'C', 'D', 'E', 'X', 0} {
		got := TradeClassifications(uwp.Profile{Starport: sp, Population: 0, Government: 0, Law: 0})
		if slices.Contains(got, "Di") && slices.Contains(got, "Ba") {
			t.Errorf("starport %q emitted both Di and Ba: %v", sp, got)
		}
	}

	// Px Prison/Exile: Atm 2/3/A/B, Hyd 1-5, Pop 3-6, Law 6-9.
	px := uwp.Profile{
		Starport:      'B',
		Atmosphere:    10,
		Hydrographics: 3,
		Population:    4,
		Government:    5,
		Law:           7,
	}
	if got := TradeClassifications(px); !slices.Contains(got, "Px") {
		t.Errorf("Prison = %v, want Px", got)
	}

	if got := TradeClassifications(
		uwp.Profile{
			Starport:      'B',
			Atmosphere:    10,
			Hydrographics: 3,
			Population:    4,
			Government:    5,
			Law:           5,
		},
	); slices.Contains(
		got,
		"Px",
	) {
		t.Errorf("Law-5 world should not be Px: %v", got)
	}

	// Re Reserve: Pop 0-4, Gov 6, Law 0/4/5.
	re := uwp.Profile{Starport: 'C', Population: 2, Government: 6, Law: 4}
	if got := TradeClassifications(re); !slices.Contains(got, "Re") {
		t.Errorf("Reserve = %v, want Re", got)
	}

	if got := TradeClassifications(
		uwp.Profile{Starport: 'C', Population: 2, Government: 5, Law: 4},
	); slices.Contains(
		got,
		"Re",
	) {
		t.Errorf("Gov-5 world should not be Re: %v", got)
	}
}

// TestOrderTradeCodes: codes render in Book 3 Chart D section order regardless of
// the order they were accumulated in — Planetary (incl. Sa/Lk), Population,
// Economic, Climate, Secondary, Political, Special. A moon's codes are built base →
// climate → satellite → zone, which is not the book's order; OrderTradeCodes fixes
// that at render.
func TestOrderTradeCodes(t *testing.T) {
	// As accumulated for a satellite: He (Planetary), a zone code, a climate code,
	// then Sa (Planetary) last.
	got := OrderTradeCodes([]tradecode.Code{"He", "Da", "Tz", "Sa"})

	want := []tradecode.Code{"He", "Sa", "Tz", "Da"} // Planetary, Climate, Special
	if !slices.Equal(got, want) {
		t.Errorf("OrderTradeCodes = %v, want %v", got, want)
	}
	// A capital (Political) sorts before the Special zone codes, wherever it was
	// stamped in.
	capacity := OrderTradeCodes([]tradecode.Code{"Fo", "Cs", "Ni"})
	if !slices.Equal(capacity, []tradecode.Code{"Ni", "Cs", "Fo"}) { // Population, Political, Special
		t.Errorf("capital ordering = %v, want [Ni Cs Fo]", capacity)
	}
	// Regina's codes are already in order and stay put.
	reg := OrderTradeCodes([]tradecode.Code{"Ph", "Pa", "Ri"})
	if !slices.Equal(reg, []tradecode.Code{"Ph", "Pa", "Ri"}) {
		t.Errorf("Regina order = %v, want [Ph Pa Ri]", reg)
	}
	// It does not mutate its input.
	in := []tradecode.Code{"Da", "He"}

	_ = OrderTradeCodes(in)
	if !slices.Equal(in, []tradecode.Code{"Da", "He"}) {
		t.Errorf("OrderTradeCodes mutated its input: %v", in)
	}
}
