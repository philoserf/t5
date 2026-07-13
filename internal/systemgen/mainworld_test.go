package systemgen

import (
	"slices"
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/uwp"
	"github.com/philoserf/t5/internal/worldgen"
)

func TestHZVarFromFlux(t *testing.T) {
	cases := map[int]int{-7: -2, -6: -2, -5: -1, -3: -1, -2: 0, 0: 0, 2: 0, 3: 1, 5: 1, 6: 2, 8: 2}
	for flux, want := range cases {
		if got := hzVarFromFlux(flux); got != want {
			t.Errorf("hzVarFromFlux(%d) = %d, want %d", flux, got, want)
		}
	}
}

func TestHZVarDM(t *testing.T) {
	cases := map[string]int{"M": 2, "O": -2, "B": -2, "F": 0, "G": 0}
	for typ, want := range cases {
		if got := hzVarDM(Star{Type: typ}); got != want {
			t.Errorf("hzVarDM(%s) = %d, want %d", typ, got, want)
		}
	}
}

func TestPlaceMainworld(t *testing.T) {
	primary := Star{Type: "F", Decimal: 7, Size: "V"} // HZ orbit 4 (Regina primary)

	// Regina: HZ-variance Flux 0 -> Temperate, orbit 4, no climate code added.
	mw := worldgen.World{Profile: uwp.Profile{Size: 7, Atmosphere: 8, Hydrographics: 8}, TradeCodes: []string{"Ph"}}
	if orbit := placeMainworld(dice.NewScripted(4, 4), primary, &mw); orbit != 4 {
		t.Errorf("Regina orbit = %d, want 4", orbit)
	}
	if !slices.Equal(mw.TradeCodes, []string{"Ph"}) {
		t.Errorf("temperate mainworld gained a climate code: %v", mw.TradeCodes)
	}

	// Hot placement: Flux 1-5 = -4 -> HZVar -1 -> orbit 3; temperate-band UWP -> Tr.
	mw2 := worldgen.World{Profile: uwp.Profile{Size: 7, Atmosphere: 6, Hydrographics: 5}}
	if orbit := placeMainworld(dice.NewScripted(1, 5), primary, &mw2); orbit != 3 || !slices.Contains(mw2.TradeCodes, "Tr") {
		t.Errorf("hot placement: orbit %d, codes %v, want orbit 3 + Tr", orbit, mw2.TradeCodes)
	}

	// A primary with no habitable zone (size-VI O star): orbit -1, no climate.
	var mw3 worldgen.World
	if orbit := placeMainworld(dice.NewScripted(4, 4), Star{Type: "O", Size: "VI"}, &mw3); orbit != -1 {
		t.Errorf("no-HZ orbit = %d, want -1", orbit)
	}
}
