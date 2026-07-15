package worldgen

import (
	"slices"
	"testing"

	"github.com/philoserf/t5/internal/uwp"
)

func TestTradeClassificationsWithContext(t *testing.T) {
	// Fa Farming: in HZ, Atm 4-9, Hyd 4-8, Pop 2-6.
	fa := uwp.Profile{Starport: 'F', Size: 7, Atmosphere: 6, Hydrographics: 5, Population: 4}
	if got := TradeClassificationsWithContext(fa, WorldContext{HasHZ: true, Orbit: 3, HZOrbit: 3}); !slices.Contains(got, "Fa") {
		t.Errorf("Fa world in HZ = %v, want Fa", got)
	}
	// Same world outside the HZ earns no Fa.
	if got := TradeClassificationsWithContext(fa, WorldContext{HasHZ: true, Orbit: 5, HZOrbit: 3}); slices.Contains(got, "Fa") {
		t.Errorf("Fa world outside HZ = %v, should not have Fa", got)
	}

	// Mi Mining: Pop 2-6 and the system mainworld is Industrial.
	mi := uwp.Profile{Starport: 'H', Size: 4, Population: 5}
	if got := TradeClassificationsWithContext(mi, WorldContext{MainworldIndustrial: true}); !slices.Contains(got, "Mi") {
		t.Errorf("Mi world (MW industrial) = %v, want Mi", got)
	}
	if got := TradeClassificationsWithContext(mi, WorldContext{MainworldIndustrial: false}); slices.Contains(got, "Mi") {
		t.Errorf("Mi world (MW not industrial) = %v, should not have Mi", got)
	}

	// Pe Penal Colony: Atm 23AB, Hyd 1-5, Pop 3-6, Gov 6, Law 6-9.
	pe := uwp.Profile{Starport: 'H', Size: 5, Atmosphere: 3, Hydrographics: 2, Population: 4, Government: 6, Law: 7}
	if got := TradeClassificationsWithContext(pe, WorldContext{}); !slices.Contains(got, "Pe") {
		t.Errorf("Pe world = %v, want Pe", got)
	}

	// The mainworld never earns the secondary codes.
	if got := TradeClassificationsWithContext(pe, WorldContext{IsMainworld: true}); slices.Contains(got, "Pe") {
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
