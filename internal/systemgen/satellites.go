package systemgen

import (
	"slices"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/tradecode"
	"github.com/philoserf/t5/internal/uwp"
	"github.com/philoserf/t5/internal/worldgen"
)

// A Satellite is a moon (or ring) of a placed world or gas giant. A moon's orbit
// is Close (tidally near) or Far, named by an orbit letter (Book 3 p.29 S-table
// and the p.24 2C Close/Far letters), and it carries a generated world type and
// UWP (the p.29 Satellites tables). A moon of a world is never larger than its
// parent (Book 3 p.21): an oversized roll is cut to the parent's size and, at
// equal size, the pair is a double planet. A Ring has no orbit letter or body.
//
// As with OtherWorld, the type, profile, and trade codes are unexported and set
// together inside systemgen (which generates the profile from the type), so a moon
// cannot carry a Type disagreeing with its Profile (#330). Read them through
// Type/Profile/TradeCodes.
type Satellite struct {
	Ring         bool
	Far          bool
	OrbitLetter  string
	typ          worldgen.OtherWorldType
	profile      uwp.Profile
	DoublePlanet bool
	tradeCodes   []tradecode.Code
}

// Type reports the moon's body type. It is meaningless for a Ring.
func (s Satellite) Type() worldgen.OtherWorldType { return s.typ }

// Profile reports the moon's UWP. It is the zero Profile for a Ring.
func (s Satellite) Profile() uwp.Profile { return s.profile }

// TradeCodes reports the moon's trade classifications.
func (s Satellite) TradeCodes() []tradecode.Code { return s.tradeCodes }

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

		parentKind, maxSize := s.satelliteParent(o)

		moons, rings := satelliteCount(r, parentKind, o.Orbit, hz, hasHZ)
		for range rings {
			o.Satellites = append(o.Satellites, Satellite{Ring: true})
		}

		// The parent may already have moons: placeOrbits appends a captured world
		// before this pass runs, and a satellite mainworld holds one of its own
		// parent's orbit letters. Both are taken before any new moon claims one.
		orbits := newSatelliteOrbits()
		for _, sat := range o.Satellites {
			orbits.take(sat.OrbitLetter)
		}

		if o.Kind == KindMainworld && (o.Giant != nil || o.Parent != nil) {
			orbits.take(s.MainworldSatellite.OrbitLetter)
		}

		for range moons {
			o.Satellites = append(o.Satellites, rollMoon(r, orbits, moonSpec{
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

// moonSpec is everything rollMoon needs that it does not roll: its parent's
// orbit and habitable zone, the mainworld context its trade codes read, and the
// parent-size cap. MaxSize is worldgen's own sentinel rather than a size/flag
// pair, so "uncapped" has one representation. A cap of zero is not a real cap —
// Size 0 is the asteroid-belt code, not a dimension — so rollMoon reads any
// non-positive MaxSize as absent, which also makes this struct's zero value safe.
//
// The world type is deliberately not a field. It was one, and the two call sites
// filled it from different book tables (Other Worlds vs Satellites, which
// disagree at outer-zone 1D=4), so the type roll moved inside rollMoon — where
// there is exactly one of it.
type moonSpec struct {
	Orbit      int
	HZOrbit    int
	HasHZ      bool
	MWPop      int
	Industrial bool
	MaxSize    int
}

// rollMoon builds one satellite: its type (the Book 3 p.29 Satellites tables,
// read at its parent's orbital zone), its UWP (size-capped to its parent, p.21),
// then Close/Far (2D, 7- is Close) and the Flux-rolled orbit letter (p.24 table
// 2C), then its trade codes in satellite context. It is the single moon-assembly
// path — both the satellite pass and the orbit map's captured world (a world
// whose orbit a gas giant already holds) go through it, so neither the dice order
// (type, UWP, 2D far, Flux letter) nor the tables they read can drift apart.
func rollMoon(r *dice.Roller, orbits *satelliteOrbits, spec moonSpec) Satellite {
	wt := satelliteType(spec.Orbit, spec.HZOrbit, spec.HasHZ, r.Die())

	// Production callers arrive normalized — satelliteBody resolves the
	// asteroid-belt code to NoSizeCap, and the orbit map's captured-world site
	// passes the sentinel outright — but this stays as the zero-value guard:
	// moonSpec{} with an unset MaxSize must mean uncapped, not "flatten every
	// moon to Size 0". Below, the cap is asked as a bound rather than compared
	// to the sentinel, matching capSize.
	maxSize := spec.MaxSize
	if maxSize <= 0 {
		maxSize = worldgen.NoSizeCap
	}

	prof := worldgen.GenerateSatelliteWorld(r, wt, spec.MWPop, maxSize)
	double := maxSize > worldgen.NoSizeCap && prof.Size == maxSize

	far := r.Dice(2) >= 8
	letter := orbits.claim(dice.FluxIndex(r.Flux()), far)

	return Satellite{
		Far:          far,
		OrbitLetter:  letter,
		typ:          wt,
		profile:      prof,
		DoublePlanet: double,
		tradeCodes: worldgen.TradeClassificationsWithContext(prof, worldgen.WorldContext{
			MainworldIndustrial: spec.Industrial,
			Belt:                wt.IsBelt(),
			Orbit:               spec.Orbit, HZOrbit: spec.HZOrbit, HasHZ: spec.HasHZ,
			Satellite: true, SatelliteFar: far,
		}),
	}
}

// mainworldIndustrial reports whether the system mainworld carries In, which the
// Mining (Mi) secondary code keys off (Book 3 Chart D p.26).
func (s *System) mainworldIndustrial() bool {
	return slices.Contains(s.Mainworld.TradeCodes, tradecode.In)
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

// satelliteParent identifies the body an orbit's moons belong to: the kind that
// selects their count rule (Book 3 p.29 "Number Of Satellites") and the Size that
// caps them (p.29, "a satellite is always smaller than its parent"), or
// worldgen.NoSizeCap when nothing caps them. Gas-giant moons are never capped — a
// giant's size code far exceeds any world size — and belts have no moons.
//
// The body is not always the orbit's Kind. When the mainworld is itself a
// satellite, p.21 places a gas giant (or, with no giant in the system, a
// BigWorld) in the mainworld's orbit to accommodate it, so the body in that orbit
// is the parent and the orbit's moons are the mainworld's siblings around it.
// Rolling them as the mainworld's own gave a gas-giant parent the world zone
// count instead of its 1D-1 and capped its moons to a fellow moon.
//
// The book does not spell out satellite handling for a satellite mainworld; this
// is the reading consistent with its own parent-body semantics. It leaves one
// point genuinely open: whether the mainworld is itself one of the parent's 1D-1
// satellites or is additional to them. It is treated as additional, since the
// count is rolled after the mainworld is already placed.
//
// A body with a UWP is classified by satelliteBody, never by the orbit's Kind:
// a mainworld's Kind says which world it is, not what kind of body it is, and an
// asteroid-belt mainworld is a belt sitting in a KindMainworld orbit.
func (s *System) satelliteParent(o *PlacedOrbit) (OrbitKind, int) {
	switch {
	case o.Kind == KindMainworld && o.Giant != nil:
		return KindGasGiant, worldgen.NoSizeCap
	case o.Kind == KindMainworld && o.Parent != nil:
		// The accommodating host is a BigWorld (Siz 2D+7), never a belt.
		return satelliteBody(o.Parent.Profile(), false)
	case o.Kind == KindMainworld:
		// A mainworld's belt-ness is its Belt field, a body fact set at generation —
		// not its Size digit, which a tiny Size-0 world shares without being a belt.
		return satelliteBody(s.Mainworld.Profile, s.Mainworld.Belt)
	case o.Kind == KindWorld && o.World != nil:
		// A secondary world's belt-ness is its TYPE, not its Size: Planetoids is
		// the belt (St000PGL-T), while a Size-0 Worldlet is a very small world.
		return satelliteBody(o.World.Profile(), o.World.Type().IsBelt())
	default:
		// A gas giant or a belt orbit: uncapped, and a belt rolls no moons at all.
		return o.Kind, worldgen.NoSizeCap
	}
}

// satelliteBody reads a UWP as the parent of satellites, answering both halves
// of the question at once: what count rule the body takes, and what caps its
// moons. The two halves must agree, and the only way to guarantee that is to
// decide them together from one read of the profile — the count rule and the cap
// were separate decisions, and the cap resolved the asteroid-belt code while the
// count did not, which is exactly #309.
//
// A Size digit of 0 is uwp.BeltSize, a code rather than a dimension. A belt is
// not a small world: it takes no satellite count (Book 3 p.29 gives one only to
// worlds and gas giants), and it caps nothing — capping to Size 0 would cut every
// moon to Size 0, taking its Atmosphere, Hydrographics and Tech Level with it and
// rendering a Big World as Y000000-0 (#213).
func satelliteBody(p uwp.Profile, isBelt bool) (OrbitKind, int) {
	if isBelt {
		return KindBelt, worldgen.NoSizeCap
	}
	// A non-belt can still be Size 0 — a Worldlet rolls max(1D-3, 0), and the
	// 2D-2 types roll 0 on snake eyes. That is a genuinely tiny world, not a
	// belt, so it keeps the world count rule; it simply has no dimension for a
	// satellite to be smaller than, so nothing caps its moons.
	if p.Size <= uwp.BeltSize {
		return KindWorld, worldgen.NoSizeCap
	}

	return KindWorld, p.Size
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

// satelliteOrbits tracks the orbit letters already taken around one parent body.
// The letters are orbit *names* (Book 3 p.24 table 2C), so two moons of one
// parent can no more share one than two worlds can share an orbit number; the
// book's own resolution applies — "If an orbit is duplicated or precluded, adjust
// to an adjacent or the closest possible orbit" (p.29 chart P2).
//
// Close and Far are separate tables of thirteen letters each, and a moon nudges
// only within its own: Close and Far describe different physical distances, so a
// Far moon cannot resolve a collision by moving into a Close orbit.
type satelliteOrbits struct {
	taken map[string]bool
}

func newSatelliteOrbits() *satelliteOrbits {
	return &satelliteOrbits{taken: map[string]bool{}}
}

// take marks a letter as occupied by a body this type did not place — a moon the
// orbit map already captured, or the satellite mainworld itself. The empty string
// (a Ring, or a non-satellite mainworld) marks nothing.
func (o *satelliteOrbits) take(letter string) {
	if letter != "" {
		o.taken[letter] = true
	}
}

// claim returns the orbit letter at the rolled Flux index, or — when that orbit
// is already occupied — the nearest free letter in the same table, spiralling
// inward before outward exactly as orbitHost.claim does for world orbits.
//
// The Flux roll itself is untouched: the nudge reads an already-rolled index and
// draws no dice, so resolving a collision cannot shift the system's dice stream.
func (o *satelliteOrbits) claim(idx int, far bool) string {
	table := closeOrbitLetters
	if far {
		table = farOrbitLetters
	}

	for d := range table {
		if i := idx - d; i >= 0 && !o.taken[table[i]] {
			o.taken[table[i]] = true

			return table[i]
		}

		if i := idx + d; i < len(table) && !o.taken[table[i]] {
			o.taken[table[i]] = true

			return table[i]
		}
	}

	// All thirteen orbits of this table are occupied. A parent with fourteen
	// moons in one hemisphere is not something the book contemplates; keep the
	// rolled letter rather than dropping a body that was generated.
	return table[idx]
}
