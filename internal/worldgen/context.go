package worldgen

import (
	"slices"

	"github.com/philoserf/t5/internal/uwp"
)

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
//
// The Special-section zone codes (Da/Pz/Fo) are appended last, matching Chart D's
// order; they are pure-UWP, so every world earns them (the mainworld gets them the
// same way in generateWorld).
func TradeClassificationsWithContext(p uwp.Profile, ctx WorldContext) []string {
	tcs := TradeClassifications(p)
	if ctx.IsMainworld {
		return append(tcs, ZoneCodes(p)...)
	}
	// Px (Prison/Exile Camp) is a mainworld-only code (Book 3 Chart D p.26, "MW");
	// a non-mainworld with that same profile is a Pe (Penal Colony) instead, so
	// strip any Px the base classifier emitted before considering Pe below.
	tcs = slices.DeleteFunc(tcs, func(code string) bool { return code == "Px" })
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
	return append(tcs, ZoneCodes(p)...)
}
