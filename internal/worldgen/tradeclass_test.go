package worldgen

import (
	"reflect"
	"testing"

	"github.com/philoserf/t5/internal/uwp"
)

// TestTradeClassificationsRegina checks the canonical worked example: Regina
// (A788899-C) is Pre-High, Pre-Agricultural, and Rich.
func TestTradeClassificationsRegina(t *testing.T) {
	regina := uwp.Profile{Starport: 'A', Size: 7, Atmosphere: 8, Hydrographics: 8, Population: 8, Government: 9, Law: 9, TechLevel: 12}
	got := TradeClassifications(regina)
	want := []string{"Ph", "Pa", "Ri"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("TradeClassifications(Regina) = %v, want %v", got, want)
	}
}

func TestTradeClassifications(t *testing.T) {
	cases := []struct {
		name string
		p    uwp.Profile
		want []string
	}{
		{
			"asteroid belt",
			uwp.Profile{Size: 0, Atmosphere: 0, Hydrographics: 0, Population: 0},
			// Size/Atm/Hyd 0 -> As; Pop 0 matches no population class here.
			[]string{"As", "Va"},
		},
		{
			"vacuum low-pop rock",
			uwp.Profile{Size: 4, Atmosphere: 0, Hydrographics: 0, Population: 2},
			// Atm 0 -> Va and De needs Atm 2-9 (no); Pop 2 -> Lo.
			[]string{"Va", "Lo"},
		},
		{
			"water world",
			uwp.Profile{Size: 6, Atmosphere: 6, Hydrographics: 10, Population: 5},
			// Siz6/Atm6/HydA -> Wa; Pop5 -> Ni; Atm6/Pop5 -> Pr (Pre-Rich).
			[]string{"Wa", "Ni", "Pr"},
		},
		{
			"garden agricultural",
			uwp.Profile{Size: 7, Atmosphere: 6, Hydrographics: 6, Population: 6},
			// Siz7/Atm6/Hyd6 -> Ga; Pop6 -> Ni; Atm6/Hyd6/Pop6 -> Ag; Atm6/Pop6 -> Ri.
			[]string{"Ga", "Ni", "Ag", "Ri"},
		},
	}
	for _, c := range cases {
		if got := TradeClassifications(c.p); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: TradeClassifications = %v, want %v", c.name, got, c.want)
		}
	}
}
