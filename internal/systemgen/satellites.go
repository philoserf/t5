package systemgen

import "github.com/philoserf/t5/internal/dice"

// A Satellite is a moon of a placed world or gas giant: its orbit is Close
// (tidally near) or Far, named by an orbit letter (Book 3 p.29 S-table and the
// p.24 2C Close/Far letters).
type Satellite struct {
	Far         bool
	OrbitLetter string
}

// rollSatellites gives each placed world and gas giant its moons (Book 3 p.29
// "Number Of Satellites"): a gas giant rolls 1D-1, and a world rolls by its
// zone — 1D-5 inside the habitable zone, 1D-4 in it, 1D-3 outside (a belt has
// none). Each satellite is then Close (2D<=7) or Far (2D>=8) with a Flux-rolled
// orbit letter. Run as its own pass after placement so it never perturbs the
// orbit map's dice.
func (s *System) rollSatellites(r *dice.Roller) {
	for i := range s.Orbits {
		o := &s.Orbits[i]
		hz, hasHZ := HZOrbit(s.hostStar(o.Host))
		n := satelliteCount(r, o.Kind, o.Orbit, hz, hasHZ)
		for range n {
			far := r.Dice(2) >= 8
			idx := dice.FluxIndex(r.Flux())
			letter := closeOrbitLetters[idx]
			if far {
				letter = farOrbitLetters[idx]
			}
			o.Satellites = append(o.Satellites, Satellite{Far: far, OrbitLetter: letter})
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

// satelliteCount rolls a body's satellite count by kind and orbital zone (Book 3
// p.29): a gas giant 1D-1, and a world 1D-5 inner / 1D-4 hospitable / 1D-3
// outer. Belts carry none; a negative roll means none.
func satelliteCount(r *dice.Roller, kind OrbitKind, orbit, hz int, hasHZ bool) int {
	if kind == KindBelt {
		return 0
	}
	if kind == KindGasGiant {
		return max(r.Die()-1, 0)
	}
	dm := -4 // hospitable
	switch zoneOf(orbit, hz, hasHZ) {
	case innerZone:
		dm = -5
	case outerZone:
		dm = -3
	}
	return max(r.Die()+dm, 0)
}
