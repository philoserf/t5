package systemgen

import (
	"sort"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// OrbitKind is what occupies a placed orbit.
type OrbitKind int

// The kinds of body an orbit can hold.
const (
	KindMainworld OrbitKind = iota
	KindGasGiant
	KindBelt
	KindWorld
)

// String names the orbit kind for display.
func (k OrbitKind) String() string {
	switch k {
	case KindGasGiant:
		return "Gas Giant"
	case KindBelt:
		return "Belt"
	case KindWorld:
		return "World"
	default:
		return "Mainworld"
	}
}

// A PlacedOrbit is one occupied orbit in the system's orbit map: the star it
// orbits (Host), its orbit number around that star, what occupies it, and — for
// a gas-giant or a detailed secondary-world orbit — the giant or world there. A
// Mainworld with a non-nil Giant is a satellite riding that gas giant; one with
// a non-nil Parent rides an accommodating BigWorld created because the system
// has no gas giant (Book 3 p.21).
type PlacedOrbit struct {
	Host       string // "Primary", "Close", "Near", or "Far"
	Orbit      int
	Kind       OrbitKind
	Giant      *GasGiant
	Parent     *OtherWorld // BigWorld hosting a satellite mainworld when no gas giant exists
	World      *OtherWorld
	Satellites []Satellite
}

// hostRank orders the host stars for a stable orbit-map sort.
var hostRank = map[string]int{"Primary": 0, "Close": 1, "Near": 2, "Far": 3}

// primaryMaxOrbit is the outermost orbit the primary star may hold (Book 3 p.21).
const primaryMaxOrbit = 19

// an orbitHost is a star that holds worlds, with its own habitable zone, floor,
// outer limit, and occupied-orbit set (Book 3 p.21).
type orbitHost struct {
	label    string
	star     Star
	hz       int
	hasHZ    bool
	floor    int
	maxOrbit int
	occupied map[int]bool
}

func newHost(label string, star Star, maxOrbit int) *orbitHost {
	hz, hasHZ := HZOrbit(star)

	return &orbitHost{
		label: label, star: star, hz: hz, hasHZ: hasHZ,
		floor: firstOrbit(star), maxOrbit: maxOrbit, occupied: map[int]bool{},
	}
}

// claim reserves the closest free orbit to want within [floor, maxOrbit],
// spiralling inward before outward (Book 3 p.29 collision/precluded resolution).
// It reports false when the star has no room left (an overflow world is dropped,
// per "if the supply of worlds exceeds the available orbits, ignore excess").
func (h *orbitHost) claim(want int) (int, bool) {
	if h.maxOrbit < h.floor {
		return 0, false
	}

	want = clamp(want, h.floor, h.maxOrbit)
	for d := 0; want-d >= h.floor || want+d <= h.maxOrbit; d++ {
		if o := want - d; o >= h.floor && !h.occupied[o] {
			h.occupied[o] = true

			return o, true
		}

		if o := want + d; o <= h.maxOrbit && !h.occupied[o] {
			h.occupied[o] = true

			return o, true
		}
	}

	return 0, false
}

// claimMainworldOrbit claims the mainworld's concrete orbit on the primary and
// makes the rest of the system agree with it. The orbit placeMainworld rolled is
// only what the mainworld wants: claim may nudge it off an orbit a secondary star
// has reserved, or clamp it to the star's precluded floor. The claimed orbit is
// therefore written back to s.MainworldOrbit and used to key the mainworld's
// climate codes, so the one-line record, the field, and the orbit map cannot
// describe three different orbits.
//
// It reports false when the mainworld has no orbit to claim.
func (s *System) claimMainworldOrbit(primary *orbitHost) (int, bool) {
	if s.MainworldOrbit < 0 {
		return 0, false // the primary has no habitable zone to place it against
	}

	o, ok := primary.claim(s.MainworldOrbit)
	if !ok {
		// Unreachable today: nothing has been placed yet, the primary's floor is at
		// most 10 (Book 1 p.31 sub-orbits) against a maxOrbit of 19, and at most
		// three secondary-star orbits are reserved, so at least six orbits are free.
		// Should a change to primaryMaxOrbit, the reservation pass, or the sub-orbit
		// floors ever break that, drop the mainworld the way an excess world is
		// dropped and say so in the field, rather than recording it at orbit 0 —
		// an orbit claim never took, and which a star may already hold.
		s.MainworldOrbit = -1

		return 0, false
	}

	s.MainworldOrbit = o
	tagMainworldClimate(&s.Mainworld, o, primary.hz, primary.hasHZ)

	return o, true
}

// giantAt returns the index in placed of a gas giant occupying the given host's
// orbit, or -1 if none does.
func giantAt(placed []PlacedOrbit, host string, orbit int) int {
	for i := range placed {
		if placed[i].Kind == KindGasGiant && placed[i].Host == host && placed[i].Orbit == orbit {
			return i
		}
	}

	return -1
}

// buildHosts lists the stars that hold worlds, in rotation order: the Primary,
// then each present secondary (Close/Near/Far). A secondary at primary-orbit N
// may hold worlds only out to its own Orbit N-3 (Book 3 p.21), so one too close
// to hold anything is omitted. Companions orbit inside their star and never host.
func (s *System) buildHosts() []*orbitHost {
	hosts := []*orbitHost{newHost("Primary", s.Primary, primaryMaxOrbit)}
	add := func(label string, star *Star, primaryOrbit int) {
		if star == nil {
			return
		}

		if h := newHost(label, *star, primaryOrbit-3); h.maxOrbit >= h.floor {
			hosts = append(hosts, h)
		}
	}
	add("Close", s.Close, s.CloseOrbit)
	add("Near", s.Near, s.NearOrbit)
	add("Far", s.Far, s.FarOrbit)

	return hosts
}

// placeOrbits builds the system's orbit map (Book 3 p.21/p.29 P1/P2). The
// mainworld is placed first on the primary (riding a gas giant when it is a
// satellite); then the remaining gas giants, the belts, and the other worlds
// are each placed by "Rotate Placement Per Star" — the first around the primary,
// the next around the Close star, then Near, then Far, repeating. Gas giants and
// belts sit at habitable-zone-relative P2 orbits, other worlds at absolute ones,
// each within its host star's own orbit space.
func (s *System) placeOrbits(r *dice.Roller) { //nolint:gocognit,cyclop,funlen // per-star orbit map, Book 3 pp.16-17
	hosts := s.buildHosts()
	primary := hosts[0]

	var placed []PlacedOrbit

	// A secondary star fills its own orbit around the primary (Book 3 p.21), so no
	// world may share it. Reserve those orbits before anything is placed; a body
	// aimed at one is nudged aside by claim, as with any other collision.
	for _, sl := range s.Stars() {
		if !sl.Companion && sl.Orbit >= 0 {
			primary.occupied[sl.Orbit] = true
		}
	}

	giants := s.Giants
	if o, ok := s.claimMainworldOrbit(primary); ok {
		mw := PlacedOrbit{Host: "Primary", Orbit: o, Kind: KindMainworld}
		// A satellite mainworld needs a parent body in its orbit to accommodate
		// it (Book 3 p.21). It rides the system's first gas giant; that giant is
		// not rotated again. If the system has no gas giant, an accommodating
		// BigWorld is created to host it. (The book prints the parent's size as
		// "-2D+7", but that contradicts worldgen's own BigWorld (Siz 2D+7) and
		// would make the parent smaller than its moon in violation of the
		// satellite-size rule, so we create a standard BigWorld.)
		//
		// A standard BigWorld is not enough on its own: 2D+7 bottoms out at 9
		// while a mainworld's Size reaches 15, so the parent's Size is floored at
		// its own moon's (Book 3 p.29, "a satellite is always smaller than its
		// parent"; equal sizes are the book's double planet). The mainworld is not
		// shrunk to fit instead — its UWP is already generated and already carries
		// the system's trade codes and extensions — so the adjustment falls on the
		// body being created for it.
		if s.MainworldSatellite.IsSatellite {
			switch {
			case len(giants) > 0:
				mw.Giant = &giants[0]
				giants = giants[1:]
			default:
				// Size is read as a dimension here — the floor the host must clear.
				// That is only sound because a belt mainworld never reaches this
				// branch: placeMainworld returns an empty MainworldSatellite for one
				// (mainworld.go), so IsSatellite is false. Ask rather than rely on
				// that invariant holding from another file, because reading the
				// belt code as a dimension is exactly how this package's last three
				// bugs happened.
				floor := s.Mainworld.Profile.Size
				if s.Mainworld.Profile.IsBelt() {
					floor = 0 // a belt has no dimension to clear
				}

				prof := worldgen.GenerateHostWorld(
					r,
					worldgen.BigWorld,
					s.Mainworld.Profile.Population,
					floor,
				)
				mw.Parent = &OtherWorld{Type: worldgen.BigWorld, Profile: prof}
			}
		}

		placed = append(placed, mw)
	}

	// rotator hands out hosts round-robin from a pool; each placement category
	// gets its own rotator so it starts fresh at the pool's first star.
	rotator := func(pool []*orbitHost) func() *orbitHost {
		i := 0

		return func() *orbitHost {
			h := pool[i%len(pool)]
			i++

			return h
		}
	}

	// Gas giants and belts are placed relative to the habitable zone, so they
	// rotate only over stars that have one — a no-HZ secondary hosts worlds but
	// not giants/belts (which would otherwise be silently dropped). If no star
	// has an HZ at all, fall back to the primary anchored at its floor.
	hzHosts := make([]*orbitHost, 0, len(hosts))
	for _, h := range hosts {
		if h.hasHZ {
			hzHosts = append(hzHosts, h)
		}
	}

	if len(hzHosts) == 0 {
		hzHosts = hosts[:1]
	}

	anchor := func(h *orbitHost) int {
		if h.hasHZ {
			return h.hz
		}

		return h.floor
	}

	ggRotate := rotator(hzHosts)

	for i := range giants {
		g := &giants[i]

		h := ggRotate()
		if o, ok := h.claim(anchor(h) + p2(r.Dice(2)).ggOffset(g.Class)); ok {
			placed = append(
				placed,
				PlacedOrbit{Host: h.label, Orbit: o, Kind: KindGasGiant, Giant: g},
			)
		}
	}

	beltRotate := rotator(hzHosts)
	for range s.Belts {
		h := beltRotate()
		if o, ok := h.claim(anchor(h) + p2(r.Dice(2)).belt); ok {
			placed = append(placed, PlacedOrbit{Host: h.label, Orbit: o, Kind: KindBelt})
		}
	}

	// Other worlds fill the remaining world count. Their orbits are absolute, and
	// each is detailed with a type (by its zone), a UWP, and trade codes — but an
	// unplaceable (overflow) world is dropped before it is rolled, so it does not
	// advance the dice stream ("ignore excess worlds", Book 3 p.21).
	mwPop := s.Mainworld.Profile.Population
	mwIndustrial := s.mainworldIndustrial()
	worldRotate := rotator(hosts)

	others := max(s.Worlds-1-s.GasGiants-s.Belts, 0)
	// A captured world claims an orbit letter around its captor, and a giant can
	// capture more than one, so the claims are kept per giant. rollSatellites
	// re-reads them from the placed satellites when it adds the rest of the moons.
	captorOrbits := map[int]*satelliteOrbits{}

	for i := range others {
		h := worldRotate()
		row := p2(r.Dice(2))

		want := row.world1
		if i == others-1 {
			want = row.world2
		}
		// A world whose target orbit is already held by a gas giant becomes that
		// giant's moon rather than being nudged to a free orbit (Book 3 p.21).
		if gi := giantAt(placed, h.label, clamp(want, h.floor, h.maxOrbit)); gi >= 0 {
			// The captor is a gas giant, whose size code exceeds any world's, so
			// the moon takes no size cap (Book 3 p.21).
			//
			// rollMoon types it from the Satellites tables, not the Other Worlds
			// ones. Book 3 p.29 prints four type tables — Inner/HZ and Outer for
			// worlds in orbits, Inner/HZ and Outer for satellites — and p.21
			// directs that "similar tables direct Satellite creation as
			// necessary". A captured world is created as a satellite in every
			// other respect (Close/Far, an orbit letter, the parent-size rule),
			// so it is typed as one. The two Outer tables differ in exactly one
			// cell, 1D=4: Iceworld for a world, Stormworld for a satellite.
			if captorOrbits[gi] == nil {
				captorOrbits[gi] = newSatelliteOrbits()
			}

			placed[gi].Satellites = append(placed[gi].Satellites, rollMoon(r, captorOrbits[gi], moonSpec{
				Orbit:      placed[gi].Orbit,
				HZOrbit:    h.hz,
				HasHZ:      h.hasHZ,
				MWPop:      mwPop,
				Industrial: mwIndustrial,
				MaxSize:    worldgen.NoSizeCap,
			}))

			continue
		}

		o, ok := h.claim(want)
		if !ok {
			continue // no room on this star: ignore the excess world
		}

		wt := otherWorldType(o, h.hz, h.hasHZ, r.Die())
		prof := worldgen.GenerateOtherWorld(r, wt, mwPop)
		tcs := worldgen.TradeClassificationsWithContext(prof, worldgen.WorldContext{
			MainworldIndustrial: mwIndustrial,
			Orbit:               o,
			HZOrbit:             h.hz,
			HasHZ:               h.hasHZ,
		})
		placed = append(
			placed,
			PlacedOrbit{
				Host:  h.label,
				Orbit: o,
				Kind:  KindWorld,
				World: &OtherWorld{Type: wt, Profile: prof, TradeCodes: tcs},
			},
		)
	}

	sort.Slice(placed, func(i, j int) bool {
		if placed[i].Host != placed[j].Host {
			return hostRank[placed[i].Host] < hostRank[placed[j].Host]
		}

		return placed[i].Orbit < placed[j].Orbit
	})
	s.Orbits = placed
}
