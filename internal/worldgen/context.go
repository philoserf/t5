package worldgen

import "github.com/philoserf/t5/internal/uwp"

// WorldContext carries the system facts a world's context-dependent trade codes
// need (Book 3 Chart D, p.26). These "Secondary" codes only attach to worlds
// other than the mainworld and depend on where the world sits and what the
// system mainworld is.
type WorldContext struct {
	IsMainworld         bool // the mainworld never earns the secondary codes
	InHZ                bool // the world sits in the habitable zone (Fa)
	MainworldIndustrial bool // the system mainworld carries In (enables Mi)
}

// TradeClassificationsWithContext returns the UWP-determinable trade codes plus
// the secondary codes that need system context: Fa (Farming), Mi (Mining), and
// Pe (Penal Colony) — Book 3 Chart D, p.26. All three are "not the mainworld"
// codes; Mr and the referee-assigned Politicals/Specials are out of scope.
func TradeClassificationsWithContext(p uwp.Profile, ctx WorldContext) []string {
	tcs := TradeClassifications(p)
	if ctx.IsMainworld {
		return tcs
	}
	// Fa Farming: in the habitable zone, Atm 4-9, Hyd 4-8, Pop 2-6.
	if ctx.InHZ && allows("456789", p.Atmosphere) && allows("45678", p.Hydrographics) && allows("23456", p.Population) {
		tcs = append(tcs, "Fa")
	}
	// Mi Mining: Pop 2-6 and the system mainworld is Industrial.
	if ctx.MainworldIndustrial && allows("23456", p.Population) {
		tcs = append(tcs, "Mi")
	}
	// Pe Penal Colony: Atm 23AB, Hyd 1-5, Pop 3-6, Gov 6, Law 6-9.
	if allows("23AB", p.Atmosphere) && allows("12345", p.Hydrographics) &&
		allows("3456", p.Population) && allows("6", p.Government) && allows("6789", p.Law) {
		tcs = append(tcs, "Pe")
	}
	return tcs
}
