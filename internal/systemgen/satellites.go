package systemgen

import (
	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/uwp"
	"github.com/philoserf/t5/internal/worldgen"
)

// A Satellite is a moon (or ring) of a placed world or gas giant. A moon's orbit
// is Close (tidally near) or Far, named by an orbit letter (Book 3 p.29 S-table
// and the p.24 2C Close/Far letters), and it carries a generated world Type and
// UWP (the p.29 Satellites tables). A moon of a world is never larger than its
// parent (Book 3 p.21): an oversized roll is cut to the parent's size and, at
// equal size, the pair is a double planet. A Ring has no orbit letter or body.
type Satellite struct {
	Ring         bool
	Far          bool
	OrbitLetter  string
	Type         worldgen.OtherWorldType
	Profile      uwp.Profile
	DoublePlanet bool
}

// rollSatellites gives each placed world and gas giant its moons (Book 3 p.29
// "Number Of Satellites"): a gas giant rolls 1D-1, and a world rolls by its
// zone — 1D-5 inside the habitable zone, 1D-4 in it, 1D-3 outside (a belt has
// none). A roll of exactly zero is a Ring plus a re-rolled count; a negative
// roll is none. Each moon then gets a type from the Satellites tables and a UWP,
// is capped to its parent's size (double planet at equal size), and is placed
// Close (2D<=7) or Far (2D>=8) with a Flux-rolled orbit letter. Run as its own
// pass after placement so it never perturbs the orbit map's dice.
func (s *System) rollSatellites(r *dice.Roller) {
	mwPop := s.Mainworld.Profile.Population
	for i := range s.Orbits {
		o := &s.Orbits[i]
		hz, hasHZ := HZOrbit(s.hostStar(o.Host))
		moons, rings := satelliteCount(r, o.Kind, o.Orbit, hz, hasHZ)
		for range rings {
			o.Satellites = append(o.Satellites, Satellite{Ring: true})
		}
		parentSize, capped := s.satelliteParentSize(o)
		for range moons {
			wt := satelliteType(o.Orbit, hz, hasHZ, r.Die())
			prof := worldgen.GenerateOtherWorld(r, wt, mwPop)
			size, double := capSatelliteSize(prof.Size, parentSize, capped)
			prof.Size = size
			far := r.Dice(2) >= 8
			idx := dice.FluxIndex(r.Flux())
			letter := closeOrbitLetters[idx]
			if far {
				letter = farOrbitLetters[idx]
			}
			o.Satellites = append(o.Satellites, Satellite{
				Far:          far,
				OrbitLetter:  letter,
				Type:         wt,
				Profile:      prof,
				DoublePlanet: double,
			})
		}
	}
}

// hostStar returns the star an orbit's host label names; an unknown or empty
// label (single-star maps) resolves to the primary.
func (s *System) hostStar(label string) Star {
	switch label {
	case "Close":
		return *s.Close
	case "Near":
		return *s.Near
	case "Far":
		return *s.Far
	default:
		return s.Primary
	}
}

// satelliteParentSize returns the parent body's UWP size and whether its moons
// are size-capped (Book 3 p.21: a satellite is never larger than its parent).
// Gas-giant moons are never capped — a giant's size code far exceeds any world
// size — and belts have no moons.
func (s *System) satelliteParentSize(o *PlacedOrbit) (size int, capped bool) {
	switch o.Kind {
	case KindMainworld:
		return s.Mainworld.Profile.Size, true
	case KindWorld:
		if o.World != nil {
			return o.World.Profile.Size, true
		}
	}
	return 0, false
}

// capSatelliteSize enforces the satellite-size rule (Book 3 p.21): a moon of a
// world that rolls at least as large as its parent is cut to the parent's size
// and flagged a double planet. Uncapped parents (gas giants) pass the size
// through unchanged.
func capSatelliteSize(satSize, parentSize int, capped bool) (int, bool) {
	if capped && satSize >= parentSize {
		return parentSize, true
	}
	return satSize, false
}

// satelliteCount rolls a body's moon count and ring count by kind and orbital
// zone (Book 3 p.29): a gas giant 1D-1, and a world 1D-5 inner / 1D-4 hospitable
// / 1D-3 outer. A roll of exactly zero yields a ring and re-rolls the count; a
// negative roll is none. Belts carry neither.
func satelliteCount(r *dice.Roller, kind OrbitKind, orbit, hz int, hasHZ bool) (moons, rings int) {
	if kind == KindBelt {
		return 0, 0
	}
	dm := -1 // gas giant: 1D-1
	if kind != KindGasGiant {
		dm = -4 // hospitable
		switch zoneOf(orbit, hz, hasHZ) {
		case innerZone:
			dm = -5
		case outerZone:
			dm = -3
		}
	}
	for {
		switch roll := r.Die() + dm; {
		case roll > 0:
			return roll, rings
		case roll < 0:
			return 0, rings
		default:
			rings++ // a ring, then re-roll the count
		}
	}
}
