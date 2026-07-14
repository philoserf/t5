package systemgen

import (
	"slices"
	"sort"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// OrbitKind is what occupies a placed orbit.
type OrbitKind int

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
func (s *System) placeOrbits(r *dice.Roller) {
	hosts := s.buildHosts()
	primary := hosts[0]
	var placed []PlacedOrbit

	giants := s.Giants
	if s.MainworldOrbit >= 0 {
		o, _ := primary.claim(s.MainworldOrbit)
		mw := PlacedOrbit{Host: "Primary", Orbit: o, Kind: KindMainworld}
		// A satellite mainworld needs a parent body in its orbit to accommodate
		// it (Book 3 p.21). It rides the system's first gas giant; that giant is
		// not rotated again. If the system has no gas giant, an accommodating
		// BigWorld is created to host it. (The book prints the parent's size as
		// "-2D+7", but that contradicts worldgen's own BigWorld (Siz 2D+7) and
		// would make the parent smaller than its moon in violation of the
		// satellite-size rule, so we create a standard BigWorld.)
		if s.MainworldSatellite.IsSatellite {
			switch {
			case len(giants) > 0:
				mw.Giant = &giants[0]
				giants = giants[1:]
			default:
				prof := worldgen.GenerateOtherWorld(r, worldgen.BigWorld, s.Mainworld.Profile.Population)
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
			placed = append(placed, PlacedOrbit{Host: h.label, Orbit: o, Kind: KindGasGiant, Giant: g})
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
	mwIndustrial := slices.Contains(s.Mainworld.TradeCodes, "In")
	worldRotate := rotator(hosts)
	others := max(s.Worlds-1-s.GasGiants-s.Belts, 0)
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
			wt := otherWorldType(placed[gi].Orbit, h.hz, h.hasHZ, r.Die())
			prof := worldgen.GenerateOtherWorld(r, wt, mwPop)
			far := r.Dice(2) >= 8
			idx := dice.FluxIndex(r.Flux())
			letter := closeOrbitLetters[idx]
			if far {
				letter = farOrbitLetters[idx]
			}
			placed[gi].Satellites = append(placed[gi].Satellites, Satellite{
				Far: far, OrbitLetter: letter, Type: wt, Profile: prof,
			})
			continue
		}
		o, ok := h.claim(want)
		if !ok {
			continue // no room on this star: ignore the excess world
		}
		wt := otherWorldType(o, h.hz, h.hasHZ, r.Die())
		prof := worldgen.GenerateOtherWorld(r, wt, mwPop)
		tcs := worldgen.TradeClassificationsWithContext(prof, worldgen.WorldContext{
			InHZ:                h.hasHZ && o == h.hz,
			MainworldIndustrial: mwIndustrial,
		})
		placed = append(placed, PlacedOrbit{Host: h.label, Orbit: o, Kind: KindWorld, World: &OtherWorld{Type: wt, Profile: prof, TradeCodes: tcs}})
	}

	sort.Slice(placed, func(i, j int) bool {
		if placed[i].Host != placed[j].Host {
			return hostRank[placed[i].Host] < hostRank[placed[j].Host]
		}
		return placed[i].Orbit < placed[j].Orbit
	})
	s.Orbits = placed
}
