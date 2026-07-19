package systemgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

func TestHostClaim(t *testing.T) {
	h := &orbitHost{floor: 0, maxOrbit: 19, occupied: map[int]bool{4: true, 5: true}}
	// Free target stays put.
	if o, ok := h.claim(6); !ok || o != 6 {
		t.Errorf("claim(6) = %d,%v, want 6,true", o, ok)
	}
	// 4 and 5 are taken (and 6 now too); claim(4) nudges inward to 3.
	if o, ok := h.claim(4); !ok || o != 3 {
		t.Errorf("claim(4) = %d,%v, want 3,true", o, ok)
	}
	// Below the floor clamps up to the first free orbit at/after the floor.
	f := &orbitHost{floor: 4, maxOrbit: 19, occupied: map[int]bool{4: true, 5: true}}
	if o, ok := f.claim(-2); !ok || o != 6 {
		t.Errorf("claim(-2, floor 4) = %d,%v, want 6,true", o, ok)
	}
	// A star with no available orbits reports no room.
	tiny := &orbitHost{floor: 0, maxOrbit: -1, occupied: map[int]bool{}}
	if _, ok := tiny.claim(0); ok {
		t.Errorf("tiny host should have no room")
	}
}

func TestPlaceOrbitsMultiStar(t *testing.T) {
	// Regina-style rotation: a satellite mainworld rides the first gas giant on
	// the primary, and the remaining giants rotate Primary -> Far. Primary F8 V
	// has HZ 4; the Far G0 V (HZ 3) sits at orbit 16, so it holds orbits 0..13.
	far := Star{Type: "G", Decimal: 0, Size: "V"}
	s := &System{
		Primary:            Star{Type: "F", Decimal: 8, Size: "V"},
		Far:                &far,
		FarOrbit:           16,
		GasGiants:          3,
		Worlds:             4, // others = 4 - 1 - 3 - 0 = 0
		MainworldOrbit:     4,
		MainworldSatellite: MainworldSatellite{IsSatellite: true, Far: true},
		Giants: []GasGiant{
			{Size: 26, Class: LargeGasGiant}, // S — the mainworld rides this one
			{Size: 21, Class: SmallGasGiant}, // M — rotates to the Primary
			{Size: 24, Class: IceGiant},      // Q->IG — rotates to the Far star
		},
	}
	s.Mainworld.Profile.Population = 8
	// SGG on Primary: 2D=2 -> p2(2).sgg=-2 -> orbit 4-2=2.
	// IG on Far (HZ 3):  2D=2 -> p2(2).ig=+1 -> orbit 3+1=4.
	s.placeOrbits(dice.NewScripted(1, 1, 1, 1))

	want := []struct {
		host  string
		orbit int
		kind  OrbitKind
		size  int // gas-giant size, 0 if none
	}{
		{"Primary", 2, KindGasGiant, 21},  // M SGG
		{"Primary", 4, KindMainworld, 26}, // rides the S LGG
		{"Far", 4, KindGasGiant, 24},      // Q IG around the Far star
	}
	if len(s.Orbits) != len(want) {
		t.Fatalf("placed %d orbits, want %d: %+v", len(s.Orbits), len(want), s.Orbits)
	}

	for i, w := range want {
		o := s.Orbits[i]
		if o.Host != w.host || o.Orbit != w.orbit || o.Kind != w.kind {
			t.Errorf(
				"orbit %d = {%s %d %s}, want {%s %d %s}",
				i,
				o.Host,
				o.Orbit,
				o.Kind,
				w.host,
				w.orbit,
				w.kind,
			)
		}

		if o.Giant == nil || o.Giant.Size != w.size {
			t.Errorf("orbit %d giant = %v, want size %d", i, o.Giant, w.size)
		}
	}
}

func TestPlaceOrbitsSatelliteMainworldNoGasGiant(t *testing.T) {
	// A satellite mainworld whose system has no gas giant gets an accommodating
	// BigWorld parent created in its orbit (Book 3 p.21).
	s := &System{
		Primary:            Star{Type: "F", Decimal: 8, Size: "V"},
		GasGiants:          0,
		Worlds:             1, // others = 1 - 1 - 0 - 0 = 0
		MainworldOrbit:     4,
		MainworldSatellite: MainworldSatellite{IsSatellite: true},
	}
	s.Mainworld.Profile.Population = 8
	s.placeOrbits(dice.NewWithSeed(1))

	if len(s.Orbits) != 1 {
		t.Fatalf("placed %d orbits, want 1: %+v", len(s.Orbits), s.Orbits)
	}

	mw := s.Orbits[0]
	if mw.Kind != KindMainworld || mw.Orbit != 4 {
		t.Errorf("mainworld orbit = {%s %d}, want {Mainworld 4}", mw.Kind, mw.Orbit)
	}

	if mw.Giant != nil {
		t.Errorf("system has no gas giant, but mainworld rides one: %v", mw.Giant)
	}

	if mw.Parent == nil || mw.Parent.Type != worldgen.BigWorld {
		t.Fatalf("mainworld parent = %+v, want a BigWorld", mw.Parent)
	}

	if sz := mw.Parent.Profile.Size; sz < 9 || sz > 19 {
		t.Errorf("BigWorld parent size = %d, want 2D+7 (9..19)", sz)
	}
}

// TestPlaceOrbitsBigWorldParentIsNotSmallerThanItsMoon is the regression for
// #215. The accommodating BigWorld parent rolls Size 2D+7 = 9..19, but a
// mainworld's Size reaches 15 (rollSize's 2D-2, rerolling a 10 as 1D+9), so the
// parent could come out smaller than the moon it exists to carry — the exact
// violation of Book 3 p.29's "a satellite is always smaller than its parent" that
// the adjacent comment invokes to reject the book's "-2D+7" printing.
func TestPlaceOrbitsBigWorldParentIsNotSmallerThanItsMoon(t *testing.T) {
	s := &System{
		Primary:            Star{Type: "F", Decimal: 8, Size: "V"},
		GasGiants:          0,
		Worlds:             1, // others = 1 - 1 - 0 - 0 = 0
		MainworldOrbit:     4,
		MainworldSatellite: MainworldSatellite{IsSatellite: true},
	}
	s.Mainworld.Profile.Population = 8
	s.Mainworld.Profile.Size = 15 // the largest a mainworld reaches
	// All 1s: the parent's Size rolls 2D+7 = 9, below its own moon.
	s.placeOrbits(dice.NewScripted(slices.Repeat([]int{1}, 20)...))

	parent := s.Orbits[0].Parent
	if parent == nil {
		t.Fatal("satellite mainworld with no giants got no BigWorld parent")
	}

	if parent.Profile.Size < s.Mainworld.Profile.Size {
		t.Errorf("BigWorld parent Size %d is smaller than its Size-%d satellite mainworld (%s)",
			parent.Profile.Size, s.Mainworld.Profile.Size, parent.Profile)
	}
}

// TestPlaceOrbitsEqualSizeHostIsADoublePlanet is the regression for #310. The
// accommodating BigWorld host's Size is floored at its satellite mainworld's
// *inclusively*, because equal sizes are the book's double planet (Book 3 p.21),
// so the pair can come out identically sized. That relationship has to be
// recorded, or the record disagrees with itself: an equal-size moon rolled by
// rollMoon around the same host prints the designation and the mainworld does not.
func TestPlaceOrbitsEqualSizeHostIsADoublePlanet(t *testing.T) {
	newSystem := func(mwSize int) *System {
		s := &System{
			Primary:            Star{Type: "F", Decimal: 8, Size: "V"},
			GasGiants:          0,
			Worlds:             1, // others = 1 - 1 - 0 - 0 = 0
			MainworldOrbit:     4,
			MainworldSatellite: MainworldSatellite{IsSatellite: true},
		}
		s.Mainworld.Profile.Population = 8
		s.Mainworld.Profile.Size = mwSize

		return s
	}

	// All 1s: the host's Size rolls 2D+7 = 9, so the floor decides the outcome. A
	// Size-15 mainworld pulls the host up to exactly 15 — the double planet. A
	// Size-3 mainworld leaves the rolled 9 standing, well clear of it.
	script := func() *dice.Roller { return dice.NewScripted(slices.Repeat([]int{1}, 20)...) }

	double := newSystem(15)
	double.placeOrbits(script())

	plain := newSystem(3)
	plain.placeOrbits(script())

	if got := double.Orbits[0].Parent.Profile.Size; got != 15 {
		t.Fatalf("host Size = %d, want 15 (floored to its equal-size mainworld)", got)
	}

	if !double.Orbits[0].DoublePlanet {
		t.Errorf("equal-size mainworld and host are not marked a double planet: %+v",
			double.Orbits[0])
	}

	if plain.Orbits[0].DoublePlanet {
		t.Errorf("Size-3 mainworld on a Size-%d host marked a double planet",
			plain.Orbits[0].Parent.Profile.Size)
	}

	if label := orbitLabel(double.Orbits[0]); !strings.Contains(label, " dp)") {
		t.Errorf("orbitLabel = %q, want moonList's double-planet designation", label)
	}

	if label := orbitLabel(plain.Orbits[0]); strings.Contains(label, " dp") {
		t.Errorf("orbitLabel = %q, want no double-planet designation", label)
	}
}

func TestPlaceOrbitsWorldCapturedByGiant(t *testing.T) {
	// Primary F8 V (HZ 4, floor 0). One SGG and one other world whose World2
	// target lands on the giant's orbit, so the world becomes the giant's moon
	// instead of being nudged aside.
	s := &System{
		Primary:        Star{Type: "F", Decimal: 8, Size: "V"},
		GasGiants:      1,
		Worlds:         3, // others = 3 - 1 - 1 - 0 = 1
		MainworldOrbit: 4,
		Giants:         []GasGiant{{Size: 23, Class: SmallGasGiant}},
	}
	s.Mainworld.Profile.Population = 8
	// GG 2D=8 -> SGG offset +4 -> orbit 8. World 2D=11 -> World2 col = 8 (the
	// giant's orbit) -> capture. Capture rolls: type 1D=5 (outer -> RadWorld),
	// RadWorld UWP (size 2D, atm Flux, hyd Flux, spaceport 1D), Close/Far 2D=5
	// (Close), letter Flux 0 -> "Gee".
	s.placeOrbits(dice.NewScripted(4, 4, 5, 6, 5, 3, 3, 3, 3, 3, 3, 4, 2, 3, 3, 3))

	if len(s.Orbits) != 2 {
		t.Fatalf(
			"got %d orbits, want 2 (mainworld + giant, no standalone world): %+v",
			len(s.Orbits),
			s.Orbits,
		)
	}

	for _, o := range s.Orbits {
		if o.Kind == KindWorld {
			t.Errorf("captured world should not occupy a standalone orbit: %+v", o)
		}
	}

	giant := s.Orbits[1] // sorted: mainworld@4, giant@8
	if giant.Kind != KindGasGiant || giant.Orbit != 8 {
		t.Fatalf("expected the giant at orbit 8: %+v", giant)
	}

	if len(giant.Satellites) != 1 {
		t.Fatalf(
			"giant should have 1 captured moon, has %d: %+v",
			len(giant.Satellites),
			giant.Satellites,
		)
	}

	m := giant.Satellites[0]
	if m.Type != worldgen.RadWorld || m.OrbitLetter != "Gee" || m.DoublePlanet {
		t.Errorf("captured moon = %+v, want RadWorld, Gee, not a double planet", m)
	}
}

// TestPlaceOrbitsCapturedWorldUsesSatelliteTable pins the one cell where Book 3
// p.29's four type tables disagree: an outer-zone 1D=4 is an Iceworld on Outer
// Worlds but a Stormworld on Outer Satellites. A world whose target orbit a gas
// giant already holds is created as that giant's satellite — it takes a Close/Far
// determination, an orbit letter, and the parent-size rule — so it is typed from
// the Satellites table, like every other moon.
//
// TestPlaceOrbitsWorldCapturedByGiant scripts a type roll of 5 (RadWorld in both
// tables), which is why the divergence went unnoticed; this test scripts the 4.
func TestPlaceOrbitsCapturedWorldUsesSatelliteTable(t *testing.T) {
	s := &System{
		Primary:        Star{Type: "F", Decimal: 8, Size: "V"},
		GasGiants:      1,
		Worlds:         3, // others = 3 - 1 - 1 - 0 = 1
		MainworldOrbit: 4,
		Giants:         []GasGiant{{Size: 23, Class: SmallGasGiant}},
	}
	s.Mainworld.Profile.Population = 8
	// As above: GG 2D=8 -> orbit 8 (outer, HZ 4), World 2D=11 -> World2 col = 8 ->
	// capture. The type die is then a 4.
	script := append([]int{4, 4, 5, 6, 4}, slices.Repeat([]int{3}, 30)...)
	s.placeOrbits(dice.NewScripted(script...))

	giant := s.Orbits[1]
	if len(giant.Satellites) != 1 {
		t.Fatalf("giant should have 1 captured moon, has %d: %+v",
			len(giant.Satellites), giant.Satellites)
	}

	if got := giant.Satellites[0].Type; got != worldgen.StormWorld {
		t.Errorf("captured moon type = %v, want %v (Outer Satellites 1D=4; Outer Worlds says %v)",
			got, worldgen.StormWorld, worldgen.Iceworld)
	}
}

// TestBeltMainworldIsAlwaysPlaced: an asteroid-belt mainworld is placed from the
// P2 Belt column as an absolute orbit (Book 3 p.21, "If the Mainworld is an
// Asteroid Belt, it is placed using the Belt Column of the Basic Placement Chart
// without regard to Habitable Zone"). That column runs negative at its low end —
// on 2D=2 it is -1 — and an orbit is never negative. The raw offset used to reach
// s.MainworldOrbit, where -1 already meant "the primary has no habitable zone, do
// not place", so the belt mainworld was silently dropped from the orbit map.
func TestBeltMainworldIsAlwaysPlaced(t *testing.T) {
	for seed := uint64(1); seed <= 1000; seed++ {
		s := GenerateForMap(dice.NewWithSeed(seed), false, true)
		if !slices.ContainsFunc(s.Orbits, func(o PlacedOrbit) bool {
			return o.Kind == KindMainworld
		}) {
			t.Fatalf(
				"seed %d: belt mainworld dropped from the orbit map (MainworldOrbit %d)\n%s",
				seed, s.MainworldOrbit, s,
			)
		}
	}
}

// climateVocabulary is every code ClimateCodes can emit. No other classifier
// produces one, so a mainworld's climate codes can be recovered from its trade
// codes by this filter.
var climateVocabulary = []string{"Ho", "Co", "Tr", "Tu", "Fr", "Tz"}

func climateOf(codes []string) []string {
	got := []string{}

	for _, c := range codes {
		if slices.Contains(climateVocabulary, c) {
			got = append(got, c)
		}
	}

	slices.Sort(got)

	return got
}

// TestMainworldOrbitAgreesWithOrbitMap: the mainworld's orbit is claimed on the
// primary, and claim may nudge it — a secondary star's orbit is reserved, and the
// want is clamped to the star's precluded floor. The system must not then disagree
// with itself: s.MainworldOrbit, the orbit the mainworld occupies in s.Orbits, and
// the orbit its climate codes describe are all the same orbit.
func TestMainworldOrbitAgreesWithOrbitMap(t *testing.T) {
	for seed := uint64(1); seed <= 2000; seed++ {
		s := Generate(dice.NewWithSeed(seed))

		i := slices.IndexFunc(s.Orbits, func(o PlacedOrbit) bool { return o.Kind == KindMainworld })
		if i < 0 {
			if s.MainworldOrbit >= 0 {
				t.Fatalf("seed %d: MainworldOrbit %d but no mainworld in the orbit map",
					seed, s.MainworldOrbit)
			}

			continue
		}

		placed := s.Orbits[i].Orbit
		if s.MainworldOrbit != placed {
			t.Fatalf("seed %d: MainworldOrbit %d but the orbit map places it at %d",
				seed, s.MainworldOrbit, placed)
		}

		if s.Mainworld.Profile.Size == 0 {
			continue // a belt mainworld takes no climate codes
		}

		hz, hasHZ := HZOrbit(s.Primary)

		want := climateOf(worldgen.ClimateCodes(s.Mainworld.Profile, placed, hz, hasHZ))
		if got := climateOf(s.Mainworld.TradeCodes); !slices.Equal(got, want) {
			t.Fatalf(
				"seed %d: mainworld at orbit %d (HZ %d) carries climate %v, want %v\n%s",
				seed, placed, hz, got, want, s,
			)
		}
	}
}

func TestOtherWorldType(t *testing.T) {
	hz := 4
	// Inner/HZ table (orbit <= hz+1): rolls 1..6.
	inner := map[int]string{
		1: "Inferno",
		2: "Inner World",
		3: "Big World",
		4: "Storm World",
		5: "Rad World",
		6: "Hospitable",
	}
	for roll, want := range inner {
		if got := otherWorldType(hz, hz, true, roll).String(); got != want {
			t.Errorf("otherWorldType(HZ, roll %d) = %q, want %q", roll, got, want)
		}
	}
	// Outer table (orbit > hz+1): rolls 1..6.
	outer := map[int]string{
		1: "Worldlet",
		2: "Iceworld",
		3: "Big World",
		4: "Iceworld",
		5: "Rad World",
		6: "Iceworld",
	}
	for roll, want := range outer {
		if got := otherWorldType(hz+3, hz, true, roll).String(); got != want {
			t.Errorf("otherWorldType(outer, roll %d) = %q, want %q", roll, got, want)
		}
	}
	// No habitable zone falls through to the outer table.
	if got := otherWorldType(4, 4, false, 1).String(); got != "Worldlet" {
		t.Errorf("otherWorldType(no HZ) = %q, want Worldlet", got)
	}
}

func TestPlaceOrbits(t *testing.T) {
	// Primary F8 V: HZ orbit 4, floor 0. Mainworld at 4. Two giants (P/SGG,
	// T/LGG), one belt, and one other world (Worlds = 5 = MW + 2 giants + 1 belt
	// + 1 other).
	s := &System{
		Primary:        Star{Type: "F", Decimal: 8, Size: "V"},
		GasGiants:      2,
		Belts:          1,
		Worlds:         5,
		MainworldOrbit: 4,
		Giants: []GasGiant{
			{Size: 23, Class: SmallGasGiant}, // P
			{Size: 27, Class: LargeGasGiant}, // T
		},
	}
	s.Mainworld.Profile.Population = 8

	// Placement rolls: GG1 2D=6 (SGG +2 -> 6), GG2 2D=8 (LGG +3 -> 7),
	// Belt 2D=4 (+1 -> 5), the lone world 2D=10 (World2 abs 9). At orbit 9
	// (outer) a type 1D=2 gives an Iceworld, then GenerateOtherWorld rolls its
	// UWP (size 2D-2=2, atm/hyd 2, pop 2D-6=6, F spaceport, TL 6) -> F222666-6.
	// That UWP (Atm 2, Hyd 2, Pop 6, Gov 6, Law 6) earns Pe on a non-mainworld.
	s.placeOrbits(
		dice.NewScripted(3, 3, 4, 4, 2, 2, 5, 5, 2, 2, 2, 3, 3, 3, 3, 6, 6, 3, 3, 3, 3, 2, 3),
	)

	want := []struct {
		orbit int
		kind  OrbitKind
	}{
		{4, KindMainworld},
		{5, KindBelt},
		{6, KindGasGiant},
		{7, KindGasGiant},
		{9, KindWorld},
	}
	if len(s.Orbits) != len(want) {
		t.Fatalf("placed %d orbits, want %d: %+v", len(s.Orbits), len(want), s.Orbits)
	}

	for i, w := range want {
		if s.Orbits[i].Orbit != w.orbit || s.Orbits[i].Kind != w.kind {
			t.Errorf("orbit %d = {%d %s}, want {%d %s}",
				i, s.Orbits[i].Orbit, s.Orbits[i].Kind, w.orbit, w.kind)
		}
	}

	if g := s.Orbits[2].Giant; g == nil || g.Size != 23 {
		t.Errorf("orbit 6 giant = %v, want size 23 (P SGG)", s.Orbits[2].Giant)
	}
	// The other world is detailed with its type, UWP, and context trade codes —
	// here the non-mainworld Pe (Penal Colony) code (closes catalog #1).
	w := s.Orbits[4].World
	if w == nil || w.Type != worldgen.Iceworld || w.Profile.String() != "F222666-6" {
		t.Errorf("orbit 9 world = %+v, want Iceworld F222666-6", w)
	}

	if !slices.Contains(w.TradeCodes, "Pe") {
		t.Errorf("orbit 9 world codes = %v, want Pe", w.TradeCodes)
	}
}
