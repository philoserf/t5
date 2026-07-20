package worldgen

import (
	"slices"
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/uwp"
)

func TestTradeClassificationsWithContext(t *testing.T) {
	// Fa Farming: in HZ, Atm 4-9, Hyd 4-8, Pop 2-6.
	fa := uwp.Profile{Starport: 'F', Size: 7, Atmosphere: 6, Hydrographics: 5, Population: 4}
	if got := TradeClassificationsWithContext(
		fa,
		WorldContext{HasHZ: true, Orbit: 3, HZOrbit: 3},
	); !slices.Contains(
		got,
		"Fa",
	) {
		t.Errorf("Fa world in HZ = %v, want Fa", got)
	}
	// Same world outside the HZ earns no Fa.
	if got := TradeClassificationsWithContext(
		fa,
		WorldContext{HasHZ: true, Orbit: 5, HZOrbit: 3},
	); slices.Contains(
		got,
		"Fa",
	) {
		t.Errorf("Fa world outside HZ = %v, should not have Fa", got)
	}

	// Mi Mining: Pop 2-6 and the system mainworld is Industrial.
	mi := uwp.Profile{Starport: 'H', Size: 4, Population: 5}
	if got := TradeClassificationsWithContext(
		mi,
		WorldContext{MainworldIndustrial: true},
	); !slices.Contains(
		got,
		"Mi",
	) {
		t.Errorf("Mi world (MW industrial) = %v, want Mi", got)
	}

	if got := TradeClassificationsWithContext(
		mi,
		WorldContext{MainworldIndustrial: false},
	); slices.Contains(
		got,
		"Mi",
	) {
		t.Errorf("Mi world (MW not industrial) = %v, should not have Mi", got)
	}

	// Pe Penal Colony: Atm 23AB, Hyd 1-5, Pop 3-6, Gov 6, Law 6-9.
	pe := uwp.Profile{
		Starport:      'H',
		Size:          5,
		Atmosphere:    3,
		Hydrographics: 2,
		Population:    4,
		Government:    6,
		Law:           7,
	}
	if got := TradeClassificationsWithContext(pe, WorldContext{}); !slices.Contains(got, "Pe") {
		t.Errorf("Pe world = %v, want Pe", got)
	}

	// The mainworld never earns the secondary codes.
	if got := TradeClassificationsWithContext(
		pe,
		WorldContext{IsMainworld: true},
	); slices.Contains(
		got,
		"Pe",
	) {
		t.Errorf("mainworld = %v, should not have Pe", got)
	}
}

// TestSatelliteCodes: a moon earns Sa (far) or Lk (close), and only a moon does —
// they are the "is this a satellite, and how far" codes (Book 3 Chart D p.26).
func TestSatelliteCodes(t *testing.T) {
	p := uwp.Profile{Size: 3, Atmosphere: 0}

	far := TradeClassificationsWithContext(p, WorldContext{Satellite: true, SatelliteFar: true})
	if !slices.Contains(far, "Sa") || slices.Contains(far, "Lk") {
		t.Errorf("a far moon = %v, want Sa (not Lk)", far)
	}

	near := TradeClassificationsWithContext(p, WorldContext{Satellite: true})
	if !slices.Contains(near, "Lk") || slices.Contains(near, "Sa") {
		t.Errorf("a close moon = %v, want Lk (not Sa)", near)
	}
	// A world that is not a moon earns neither.
	planet := TradeClassificationsWithContext(p, WorldContext{})
	if slices.Contains(planet, "Sa") || slices.Contains(planet, "Lk") {
		t.Errorf("a non-moon = %v, want no satellite code", planet)
	}
}

// TestAssemblerReturnsChartDOrder: the assembler sorts its output into Chart D
// order, so a caller stores and renders it directly — no renderer has to remember
// to sort. The codes accumulate out of order (a satellite appends its Planetary
// Sa/Lk code after the climate codes), so this is a real reordering, not a no-op.
func TestAssemblerReturnsChartDOrder(t *testing.T) {
	// A far satellite one orbit inside the HZ: earns a climate code (Ho, Climate
	// section) and Sa (Planetary section). Accumulation puts Ho before Sa; Chart D
	// puts the Planetary Sa first.
	p := uwp.Profile{Size: 7, Atmosphere: 6, Hydrographics: 5}

	got := TradeClassificationsWithContext(p, WorldContext{
		Orbit: 3, HZOrbit: 4, HasHZ: true, Satellite: true, SatelliteFar: true,
	})
	if i, j := slices.Index(got, "Sa"), slices.Index(got, "Ho"); i < 0 || j < 0 || i > j {
		t.Errorf("codes not in Chart D order: %v (Sa must precede Ho)", got)
	}
	// The result is exactly what OrderTradeCodes would produce.
	if !slices.Equal(got, OrderTradeCodes(got)) {
		t.Errorf("assembler output is not fully ordered: %v", got)
	}
}

// TestAsIsBeltContext is #324/#328: As is a Chart-D UWP code, but for a SECONDARY
// world it means a belt only when the body is a Planetoids belt (ctx.Belt). A
// mainworld with a Size-0 profile is always a belt (Book 3 p.16), so it keeps As
// regardless; a secondary keeps As only with Belt, and a Size-0 Worldlet (no Belt)
// has its spurious As stripped.
func TestAsIsBeltContext(t *testing.T) {
	belt := uwp.Profile{Size: 0, Atmosphere: 0, Hydrographics: 0}

	// Mainworld: As stands whether or not Belt is set — Size 0 is a belt.
	for _, b := range []bool{true, false} {
		mw := TradeClassificationsWithContext(belt, WorldContext{IsMainworld: true, Belt: b})
		if !slices.Contains(mw, "As") {
			t.Errorf("mainworld Size-0 profile (Belt=%v) = %v, want As (a Size-0 mainworld is a belt)", b, mw)
		}
	}

	// Secondary belt (Planetoids): As kept.
	if sec := TradeClassificationsWithContext(belt, WorldContext{Belt: true}); !slices.Contains(sec, "As") {
		t.Errorf("secondary Planetoids belt = %v, want As", sec)
	}

	// Secondary non-belt (Worldlet-shaped): the pure classifier's As is stripped.
	if sec := TradeClassificationsWithContext(belt, WorldContext{Belt: false}); slices.Contains(sec, "As") {
		t.Errorf("secondary non-belt Size-0 = %v, must not have As", sec)
	}
}

// TestSizeZeroWorldletCarriesNoAs is the acceptance test for #324/#328: a real
// generated Worldlet that rolls Size 0 renders St000... — the exact profile the
// retired digit-rule stamped As on — but it is a tiny solid world, not a belt,
// so with a Worldlet's (non-Planetoids) belt context it earns no As. The absence
// of this test is what let the phantom-moons belt reach stdout.
func TestSizeZeroWorldletCarriesNoAs(t *testing.T) {
	// An all-1s roller: typeSize(Worldlet) = max(1D-3, 0) = 0.
	wl := GenerateOtherWorld(dice.NewSource(func() int { return 1 }), Worldlet, 5)
	if wl.Size != uwp.BeltSize {
		t.Fatalf("expected a Size-0 Worldlet, got %s (Size %d)", wl, wl.Size)
	}

	// Belt-ness comes from the type, exactly as systemgen sources it: a Worldlet
	// is not Planetoids, so Worldlet.IsBelt() is false and it earns no As.
	tcs := TradeClassificationsWithContext(wl, WorldContext{Belt: Worldlet.IsBelt()})
	if slices.Contains(tcs, "As") {
		t.Errorf("a Size-0 Worldlet must not be classified As (it is a tiny world, not a belt): %v", tcs)
	}
}
