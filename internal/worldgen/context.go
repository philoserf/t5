package worldgen

import (
	"slices"

	"github.com/philoserf/t5/internal/uwp"
)

// WorldContext carries the system facts a world's context-dependent trade codes
// need (Book 3 Chart D, p.26): where the world sits, what the system mainworld is,
// and whether it is a moon. It carries the raw orbit rather than a derived InHZ so
// that TradeClassificationsWithContext is the single place that assembles a
// non-mainworld's complete code list — the habitable-zone test, the climate codes,
// and the satellite code all follow from these fields, and were once appended by
// hand at three separate call sites.
type WorldContext struct {
	IsMainworld         bool // the mainworld never earns the secondary codes
	MainworldIndustrial bool // the system mainworld carries In (enables Mi)

	Orbit   int  // the world's orbit number, for the habitable-zone and climate codes
	HZOrbit int  // the star's habitable-zone orbit
	HasHZ   bool // the star has a habitable zone at all

	Satellite    bool // the world is a moon (earns Sa or Lk)
	SatelliteFar bool // ...a far moon (Sa) rather than a close/locked one (Lk)
}

// TradeClassificationsWithContext assembles a world's complete trade-code list
// from its profile and its place in the system (Book 3 Chart D, p.26). It is the
// single assembler for a non-mainworld: the base UWP codes, the Secondary codes
// that need system context (Fa Farming, Mi Mining, Pe Penal Colony), the climate
// codes its orbit earns, the satellite code if it is a moon, and the pure-UWP zone
// codes — in Chart D order. A caller passes a WorldContext and gets the finished
// list, rather than appending climate and Sa/Lk by hand as three call sites once
// did.
//
// The mainworld is the one exception: it is generated before it is placed, so it
// takes only the base and zone codes here (IsMainworld, no orbit) and earns its
// climate and satellite codes later, at placement (see systemgen/mainworld.go).
func TradeClassificationsWithContext(p uwp.Profile, ctx WorldContext) []string {
	tcs := TradeClassifications(p)
	// The Secondary codes apply only to non-mainworlds (Book 3 Chart D p.26).
	if !ctx.IsMainworld {
		inHZ := ctx.HasHZ && ctx.Orbit == ctx.HZOrbit
		// Px (Prison/Exile Camp) is a mainworld-only code, "MW"; a non-mainworld
		// with that same profile is a Pe (Penal Colony) instead, so strip any Px the
		// base classifier emitted before considering Pe below.
		tcs = slices.DeleteFunc(tcs, func(code string) bool { return code == "Px" })
		// Fa Farming: in the habitable zone, Atm 4-9, Hyd 4-8, Pop 2-6.
		if inHZ && allows("456789", p.Atmosphere) && allows("45678", p.Hydrographics) && allows("23456", p.Population) {
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
		// Climate codes from the orbit (Book 3 p.26 / Chart B).
		if ctx.HasHZ {
			tcs = append(tcs, ClimateCodes(p, ctx.Orbit, ctx.HZOrbit)...)
		}
		// The satellite code: Sa for a far moon, Lk for a close (locked) one — pure
		// "is this a moon, and how far" codes with no UWP constraints.
		if ctx.Satellite {
			if ctx.SatelliteFar {
				tcs = append(tcs, "Sa")
			} else {
				tcs = append(tcs, "Lk")
			}
		}
	}
	// The Special-section zone codes (Da/Pz/Fo) are pure-UWP, so every world earns
	// them, last, in Chart D order.
	return append(tcs, ZoneCodes(p)...)
}
