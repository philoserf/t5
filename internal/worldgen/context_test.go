package worldgen

import (
	"slices"
	"testing"

	"github.com/philoserf/t5/internal/uwp"
)

func TestTradeClassificationsWithContext(t *testing.T) {
	// Fa Farming: in HZ, Atm 4-9, Hyd 4-8, Pop 2-6.
	fa := uwp.Profile{Starport: 'F', Size: 7, Atmosphere: 6, Hydrographics: 5, Population: 4}
	if got := TradeClassificationsWithContext(fa, WorldContext{InHZ: true}); !slices.Contains(got, "Fa") {
		t.Errorf("Fa world in HZ = %v, want Fa", got)
	}
	// Same world outside the HZ earns no Fa.
	if got := TradeClassificationsWithContext(fa, WorldContext{InHZ: false}); slices.Contains(got, "Fa") {
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
