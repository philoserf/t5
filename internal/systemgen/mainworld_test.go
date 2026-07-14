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

	// A temperate planet: HZ-variance Flux 0 -> orbit 4; type Flux +2 -> Planet
	// (not a satellite), so no climate/satellite code beyond the given Ph.
	mw := worldgen.World{Profile: uwp.Profile{Size: 7, Atmosphere: 8, Hydrographics: 8}, TradeCodes: []string{"Ph"}}
	orbit, sat := placeMainworld(dice.NewScripted(4, 4 /*HZVar 0*/, 5, 3 /*type Flux +2 -> Planet*/), primary, &mw)
	if orbit != 4 || sat.IsSatellite {
		t.Errorf("Regina placement: orbit %d satellite %v, want orbit 4, planet", orbit, sat.IsSatellite)
	}
	if !slices.Equal(mw.TradeCodes, []string{"Ph"}) {
		t.Errorf("temperate planet gained an extra code: %v", mw.TradeCodes)
	}

	// Hot placement: HZVar Flux -4 -> -1 -> orbit 3; type Flux +2 -> Planet;
	// temperate-band UWP -> Tr.
	mw2 := worldgen.World{Profile: uwp.Profile{Size: 7, Atmosphere: 6, Hydrographics: 5}}
	orbit2, _ := placeMainworld(dice.NewScripted(1, 5 /*-4*/, 5, 3 /*Planet*/), primary, &mw2)
	if orbit2 != 3 || !slices.Contains(mw2.TradeCodes, "Tr") {
		t.Errorf("hot placement: orbit %d, codes %v, want orbit 3 + Tr", orbit2, mw2.TradeCodes)
	}

	// Far-satellite mainworld: type Flux -4 -> Far Satellite, orbit Flux -2 ->
	// "Arr"; the mainworld earns Sa.
	mw4 := worldgen.World{Profile: uwp.Profile{Size: 7, Atmosphere: 8, Hydrographics: 8}}
	_, sat4 := placeMainworld(dice.NewScripted(4, 4 /*HZVar 0*/, 1, 5 /*type -4*/, 4, 6 /*orbit -2*/), primary, &mw4)
	if !sat4.IsSatellite || !sat4.Far || sat4.OrbitLetter != "Arr" {
		t.Errorf("far satellite = %+v, want Far/Arr", sat4)
	}
	if !slices.Contains(mw4.TradeCodes, "Sa") {
		t.Errorf("far-satellite mainworld should carry Sa: %v", mw4.TradeCodes)
	}

	// A primary with no habitable zone (size-VI O star): orbit -1.
	mw3 := worldgen.World{Profile: uwp.Profile{Size: 7}}
	orbit3, _ := placeMainworld(dice.NewScripted(4, 4, 5, 3), Star{Type: "O", Size: "VI"}, &mw3)
	if orbit3 != -1 {
		t.Errorf("no-HZ orbit = %d, want -1", orbit3)
	}

	// An asteroid-belt mainworld (Size 0) is placed via the P2 Belt column
	// without regard to HZ: 2D=7 -> Belt offset 4; no satellite, no codes.
	mw5 := worldgen.World{Profile: uwp.Profile{Size: 0}}
	orbit5, sat5 := placeMainworld(dice.NewScripted(3, 4 /*2D=7*/), primary, &mw5)
	if orbit5 != 4 || sat5.IsSatellite || len(mw5.TradeCodes) != 0 {
		t.Errorf("belt mainworld: orbit %d sat %v codes %v, want orbit 4, planet, no codes", orbit5, sat5.IsSatellite, mw5.TradeCodes)
	}
}
