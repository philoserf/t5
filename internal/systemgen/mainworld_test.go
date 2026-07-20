package systemgen

import (
	"slices"
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/tradecode"
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
	mw := worldgen.World{
		Profile:    uwp.Profile{Size: 7, Atmosphere: 8, Hydrographics: 8},
		TradeCodes: []tradecode.Code{"Ph"},
	}

	orbit, sat := placeMainworld(
		dice.NewScripted(4, 4 /*HZVar 0*/, 5, 3 /*type Flux +2 -> Planet*/),
		primary,
		&mw,
	)
	if orbit != 4 || sat.IsSatellite {
		t.Errorf(
			"Regina placement: orbit %d satellite %v, want orbit 4, planet",
			orbit,
			sat.IsSatellite,
		)
	}

	if !slices.Equal(mw.TradeCodes, []tradecode.Code{"Ph"}) {
		t.Errorf("temperate planet gained an extra code: %v", mw.TradeCodes)
	}

	// Hot placement: HZVar Flux -4 -> -1 -> orbit 3. The climate codes that orbit
	// earns are tagged later, against the claimed orbit (see TestTagMainworldClimate).
	mw2 := worldgen.World{Profile: uwp.Profile{Size: 7, Atmosphere: 6, Hydrographics: 5}}

	orbit2, _ := placeMainworld(dice.NewScripted(1, 5 /*-4*/, 5, 3 /*Planet*/), primary, &mw2)
	if orbit2 != 3 || len(mw2.TradeCodes) != 0 {
		t.Errorf("hot placement: orbit %d, codes %v, want orbit 3 and no codes yet",
			orbit2, mw2.TradeCodes)
	}

	// Far-satellite mainworld: one Flux of -4 (Chart 2C) gives both the Far
	// Satellite type and, from the same row's Far column, the orbit letter "Pee".
	// It is one roll, not two, so no separate orbit Flux is consumed.
	mw4 := worldgen.World{Profile: uwp.Profile{Size: 7, Atmosphere: 8, Hydrographics: 8}}

	_, sat4 := placeMainworld(dice.NewScripted(4, 4 /*HZVar 0*/, 1, 5 /*type -4*/), primary, &mw4)
	if !sat4.IsSatellite || !sat4.Far || sat4.OrbitLetter != "Pee" {
		t.Errorf("far satellite = %+v, want Far/Pee", sat4)
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
	mw5 := worldgen.World{Profile: uwp.Profile{Size: 0}, Belt: true}

	orbit5, sat5 := placeMainworld(dice.NewScripted(3, 4 /*2D=7*/), primary, &mw5)
	if orbit5 != 4 || sat5.IsSatellite || len(mw5.TradeCodes) != 0 {
		t.Errorf(
			"belt mainworld: orbit %d sat %v codes %v, want orbit 4, planet, no codes",
			orbit5,
			sat5.IsSatellite,
			mw5.TradeCodes,
		)
	}
	// The Belt column goes negative at its low end (2D=2 -> -1), which is not an
	// orbit and collided with the "no orbit" sentinel. It is floored at the star's
	// innermost legal orbit: 0 for the main-sequence F7 V above, and 3 for an M0 III
	// giant whose surface engulfs orbits 0-2 (Book 1 p.31 sub-orbits).
	mw6 := worldgen.World{Profile: uwp.Profile{Size: 0}, Belt: true}

	if orbit6, _ := placeMainworld(dice.NewScripted(1, 1 /*2D=2*/), primary, &mw6); orbit6 != 0 {
		t.Errorf("belt mainworld at 2D=2: orbit %d, want 0 (the F7 V floor)", orbit6)
	}

	mw7 := worldgen.World{Profile: uwp.Profile{Size: 0}, Belt: true}
	giant := Star{Type: "M", Decimal: 0, Size: "III"}

	if orbit7, _ := placeMainworld(dice.NewScripted(1, 1 /*2D=2*/), giant, &mw7); orbit7 != 3 {
		t.Errorf("belt mainworld at 2D=2 around M0 III: orbit %d, want 3 (its floor)", orbit7)
	}
}

// TestTagMainworldClimate: the climate codes are keyed to the orbit the mainworld
// was finally placed in, and an asteroid-belt mainworld — placed without regard to
// the habitable zone — takes none.
func TestTagMainworldClimate(t *testing.T) {
	// A temperate-band world one orbit inside the HZ is Hot and Tropic (Chart D).
	mw := worldgen.World{Profile: uwp.Profile{Size: 7, Atmosphere: 6, Hydrographics: 5}}

	tagMainworldClimate(&mw, 3, 4, true)

	if !slices.Equal(mw.TradeCodes, []tradecode.Code{"Ho", "Tr"}) {
		t.Errorf("orbit 3 of HZ 4 = %v, want [Ho Tr]", mw.TradeCodes)
	}
	// The same world placed in the habitable zone itself earns nothing — which is
	// the divergence this split fixes: the codes follow the claimed orbit, so a
	// mainworld nudged from HZ-1 to HZ stops claiming to be Hot.
	mw2 := worldgen.World{Profile: uwp.Profile{Size: 7, Atmosphere: 6, Hydrographics: 5}}

	tagMainworldClimate(&mw2, 4, 4, true)

	if len(mw2.TradeCodes) != 0 {
		t.Errorf("orbit 4 of HZ 4 = %v, want no codes", mw2.TradeCodes)
	}
	// A belt mainworld takes no climate codes even in a coded orbit.
	belt := worldgen.World{Profile: uwp.Profile{Size: 0}, Belt: true}

	tagMainworldClimate(&belt, 3, 4, true)

	if len(belt.TradeCodes) != 0 {
		t.Errorf("belt mainworld gained climate codes: %v", belt.TradeCodes)
	}
}
