package systemgen

import (
	"slices"

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
	TradeCodes   []string
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

	mwIndustrial := s.mainworldIndustrial() // invariant for the system; hoisted, like placeOrbits
	for i := range s.Orbits {
		o := &s.Orbits[i]
		hz, hasHZ := HZOrbit(s.hostStar(o.Host))

		moons, rings := satelliteCount(r, o.Kind, o.Orbit, hz, hasHZ)
		for range rings {
			o.Satellites = append(o.Satellites, Satellite{Ring: true})
		}

		maxSize := s.satelliteMaxSize(o)
		for range moons {
			wt := satelliteType(o.Orbit, hz, hasHZ, r.Die())
			o.Satellites = append(o.Satellites, rollMoon(r, moonSpec{
				Type:       wt,
				Orbit:      o.Orbit,
				HZOrbit:    hz,
				HasHZ:      hasHZ,
				MWPop:      mwPop,
				Industrial: mwIndustrial,
				MaxSize:    maxSize,
			}))
		}
	}
}

// moonSpec is everything rollMoon needs that it does not roll: the moon's
// already-determined world type, its parent's orbit and habitable zone, the
// mainworld context its trade codes read, and the parent-size cap. MaxSize is
// worldgen's own sentinel rather than a size/flag pair, so "uncapped" has one
// representation and a capped moon cannot be spelled as a cap of zero.
type moonSpec struct {
	Type       worldgen.OtherWorldType
	Orbit      int
	HZOrbit    int
	HasHZ      bool
	MWPop      int
	Industrial bool
	MaxSize    int
}

// rollMoon builds one satellite: its UWP (size-capped to its parent, Book 3
// p.21), then Close/Far (2D, 7- is Close) and the Flux-rolled orbit letter
// (p.24 table 2C), then its trade codes in satellite context. It is the single
// moon-assembly path — both the satellite pass and the orbit map's captured
// world (a world whose orbit a gas giant already holds) go through it, so the
// dice order (UWP, 2D far, Flux letter) cannot drift between them.
func rollMoon(r *dice.Roller, spec moonSpec) Satellite {
	// No parent caps its moons at Size 0 — satelliteMaxSize resolves the
	// asteroid-belt code to NoSizeCap — so treat any non-positive cap as absent
	// rather than flattening the moon. This also keeps moonSpec's zero value
	// safe: an unset MaxSize means uncapped, not "cap everything to Size 0".
	maxSize := spec.MaxSize
	if maxSize <= 0 {
		maxSize = worldgen.NoSizeCap
	}

	prof := worldgen.GenerateSatelliteWorld(r, spec.Type, spec.MWPop, maxSize)
	double := maxSize != worldgen.NoSizeCap && prof.Size == maxSize

	far := r.Dice(2) >= 8
	idx := dice.FluxIndex(r.Flux())

	letter := closeOrbitLetters[idx]
	if far {
		letter = farOrbitLetters[idx]
	}

	return Satellite{
		Far:          far,
		OrbitLetter:  letter,
		Type:         spec.Type,
		Profile:      prof,
		DoublePlanet: double,
		TradeCodes: worldgen.TradeClassificationsWithContext(prof, worldgen.WorldContext{
			MainworldIndustrial: spec.Industrial,
			Orbit:               spec.Orbit, HZOrbit: spec.HZOrbit, HasHZ: spec.HasHZ,
			Satellite: true, SatelliteFar: far,
		}),
	}
}

// mainworldIndustrial reports whether the system mainworld carries In, which the
// Mining (Mi) secondary code keys off (Book 3 Chart D p.26).
func (s *System) mainworldIndustrial() bool {
	return slices.Contains(s.Mainworld.TradeCodes, "In")
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

// satelliteMaxSize returns the size cap a body's moons take (Book 3 p.21: a
// satellite is never larger than its parent), or worldgen.NoSizeCap when they
// take none. Gas-giant moons are never capped — a giant's size code far exceeds
// any world size — and belts have no moons.
//
// A Size digit of 0 is not a cap. In a UWP it marks an asteroid belt (the same
// convention PortFacilities reads to site a Beltport), so it is a code rather
// than a dimension: capping to it would cut every moon of a belt mainworld to
// Size 0, taking its Atmosphere, Hydrographics and Tech Level with it and
// rendering a Big World as Y000000-0.
func (s *System) satelliteMaxSize(o *PlacedOrbit) int {
	switch o.Kind {
	case KindMainworld:
		return sizeCapOf(s.Mainworld.Profile.Size)
	case KindWorld:
		if o.World != nil {
			return sizeCapOf(o.World.Profile.Size)
		}
	default:
		// KindGasGiant and KindBelt are uncapped; handled by the return below.
	}

	return worldgen.NoSizeCap
}

// sizeCapOf turns a parent's UWP Size digit into a satellite cap, treating 0 —
// the asteroid-belt code — as no cap at all.
func sizeCapOf(parentSize int) int {
	if parentSize <= 0 {
		return worldgen.NoSizeCap
	}

	return parentSize
}

// satelliteCount rolls a body's moon count and ring count by kind and orbital
// zone (Book 3 p.29): a gas giant 1D-1, and a world 1D-5 inner / 1D-4 hospitable
// / 1D-3 outer. A roll of exactly zero yields a ring and re-rolls the count; a
// negative roll is none. Belts carry neither.
func satelliteCount(r *dice.Roller, kind OrbitKind, orbit, hz int, hasHZ bool) (int, int) {
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
		default:
			// hospitableZone keeps the -4 set above.
		}
	}

	rings := 0

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
