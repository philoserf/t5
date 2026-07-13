package systemgen

import (
	"slices"
	"sort"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// A PlacedOrbit is one occupied orbit in the system's orbit map: its orbit
// number around the primary, what occupies it, and — for a gas-giant or a
// detailed secondary-world orbit — the giant or world placed there.
type PlacedOrbit struct {
	Orbit int
	Kind  string // "Mainworld", "Gas Giant", "Belt", or "World"
	Giant *GasGiant
	World *OtherWorld
}

// placeOrbits builds the primary star's orbit map (Book 3 p.29 P1/P2). It places
// the mainworld at its rolled orbit, then each gas giant and belt at an
// HZ-relative P2 orbit, then the remaining "other" worlds at absolute P2 orbits
// (World1 for all but the last, World2 for the last). A duplicated or precluded
// target is nudged to the closest free orbit at or beyond the star's floor.
//
// Placement across secondary stars ("Rotate Placement Per Star") and their
// exclusion zones are deferred: everything is placed around the primary, which
// is exactly correct for a single-star system.
func (s *System) placeOrbits(r *dice.Roller) {
	floor := firstOrbit(s.Primary)
	occupied := map[int]bool{}
	var placed []PlacedOrbit

	claim := func(want int) int {
		o := resolveOrbit(want, floor, occupied)
		occupied[o] = true
		return o
	}

	if s.MainworldOrbit >= 0 {
		placed = append(placed, PlacedOrbit{Orbit: claim(s.MainworldOrbit), Kind: "Mainworld"})
	}

	// Gas giants and belts are placed relative to the habitable zone, so they
	// need one (a primary with no HZ has no place to hang them).
	hz, hasHZ := HZOrbit(s.Primary)
	if hasHZ {
		for i := range s.Giants {
			g := &s.Giants[i]
			placed = append(placed, PlacedOrbit{Orbit: claim(hz + p2(r.Dice(2)).ggOffset(g.Class)), Kind: "Gas Giant", Giant: g})
		}
		for range s.Belts {
			placed = append(placed, PlacedOrbit{Orbit: claim(hz + p2(r.Dice(2)).belt), Kind: "Belt"})
		}
	}

	// Other worlds fill the remaining world count (mainworld, giants, and belts
	// already account for the rest). Their orbits are absolute; each is then
	// detailed with a type, UWP, and trade codes given its zone and the system.
	mwPop := s.Mainworld.Profile.Population
	mwIndustrial := slices.Contains(s.Mainworld.TradeCodes, "In")
	others := max(s.Worlds-1-s.GasGiants-s.Belts, 0)
	for i := range others {
		row := p2(r.Dice(2))
		want := row.world1
		if i == others-1 {
			want = row.world2
		}
		o := claim(want)
		wt := otherWorldType(o, hz, hasHZ, r.Die())
		prof := worldgen.GenerateOtherWorld(r, wt, mwPop)
		tcs := worldgen.TradeClassificationsWithContext(prof, worldgen.WorldContext{
			InHZ:                hasHZ && o == hz,
			MainworldIndustrial: mwIndustrial,
		})
		placed = append(placed, PlacedOrbit{Orbit: o, Kind: "World", World: &OtherWorld{Type: wt, Profile: prof, TradeCodes: tcs}})
	}

	sort.Slice(placed, func(i, j int) bool { return placed[i].Orbit < placed[j].Orbit })
	s.Orbits = placed
}

// resolveOrbit returns the closest free orbit to want at or beyond floor (Book 3
// p.29: "if an orbit is duplicated or precluded, adjust to an adjacent or the
// closest possible orbit"). Orbits below the floor are precluded, so it clamps
// up first, then spirals out — inward before outward at equal distance.
func resolveOrbit(want, floor int, occupied map[int]bool) int {
	if want < floor {
		want = floor
	}
	for d := 0; ; d++ {
		if o := want - d; o >= floor && !occupied[o] {
			return o
		}
		if o := want + d; !occupied[o] {
			return o
		}
	}
}
