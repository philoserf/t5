package systemgen

import (
	"sort"

	"github.com/philoserf/t5/internal/dice"
)

// A PlacedOrbit is one occupied orbit in the system's orbit map: its orbit
// number around the primary, what occupies it, and — for a gas-giant orbit —
// the giant placed there.
type PlacedOrbit struct {
	Orbit int
	Kind  string // "Mainworld", "Gas Giant", "Belt", or "World"
	Giant *GasGiant
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

	place := func(want int, kind string, g *GasGiant) {
		o := resolveOrbit(want, floor, occupied)
		occupied[o] = true
		placed = append(placed, PlacedOrbit{Orbit: o, Kind: kind, Giant: g})
	}

	if s.MainworldOrbit >= 0 {
		place(s.MainworldOrbit, "Mainworld", nil)
	}

	// Gas giants and belts are placed relative to the habitable zone, so they
	// need one (a primary with no HZ has no place to hang them).
	if hz, ok := HZOrbit(s.Primary); ok {
		for i := range s.Giants {
			g := &s.Giants[i]
			place(hz+p2(r.Dice(2)).ggOffset(g.Class), "Gas Giant", g)
		}
		for range s.Belts {
			place(hz+p2(r.Dice(2)).belt, "Belt", nil)
		}
	}

	// Other worlds fill the remaining world count (mainworld, giants, and belts
	// already account for the rest). Their orbits are absolute.
	others := max(s.Worlds-1-s.GasGiants-s.Belts, 0)
	for i := range others {
		row := p2(r.Dice(2))
		orbit := row.world1
		if i == others-1 {
			orbit = row.world2
		}
		place(orbit, "World", nil)
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
