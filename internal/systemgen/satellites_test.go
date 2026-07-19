package systemgen

import (
	"slices"
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/uwp"
	"github.com/philoserf/t5/internal/worldgen"
)

// countThenSixes is a die source giving one leading face — the satellite-count
// die — and a 6 for everything after it. rollMoon's dice consumption varies with
// the world type it rolls, so a fixed script cannot keep several moons aligned;
// an unbounded source can.
func countThenSixes(first int) func() int {
	rolled := false

	return func() int {
		if !rolled {
			rolled = true

			return first
		}

		return 6
	}
}

// moonScript builds a rollMoon die script: the leading type die, then n sixes for
// everything the moon rolls after it.
func moonScript(typeDie, n int) []int {
	return append([]int{typeDie}, slices.Repeat([]int{6}, n)...)
}

func TestSatelliteCount(t *testing.T) {
	hz := 4
	// 1D roll of 6 with each zone's DM: GG 1D-1=5, inner 1D-5=1, HZ 1D-4=2,
	// outer 1D-3=3; belts always 0.
	cases := []struct {
		kind  OrbitKind
		orbit int
		want  int
	}{
		{KindGasGiant, 2, 5},
		{KindWorld, 2, 1},     // inner (HZ-2)
		{KindWorld, 3, 2},     // hospitable (HZ-1), not inner
		{KindMainworld, 4, 2}, // hospitable (HZ)
		{KindWorld, 5, 2},     // hospitable (HZ+1), not outer
		{KindWorld, 6, 3},     // outer (HZ+2)
		{KindWorld, 9, 3},     // outer
		{KindBelt, 4, 0},
	}
	for _, c := range cases {
		if got, rings := satelliteCount(
			dice.NewScripted(6),
			c.kind,
			c.orbit,
			hz,
			true,
		); got != c.want ||
			rings != 0 {
			t.Errorf(
				"satelliteCount(%s, orbit %d) = %d moons, %d rings, want %d moons, 0 rings",
				c.kind,
				c.orbit,
				got,
				rings,
				c.want,
			)
		}
	}
	// A negative roll is none (inner 1D-5 with 1D=1 -> -4 -> none).
	if got, rings := satelliteCount(
		dice.NewScripted(1),
		KindWorld,
		2,
		hz,
		true,
	); got != 0 ||
		rings != 0 {
		t.Errorf("satelliteCount(inner, 1D=1) = %d moons, %d rings, want 0, 0", got, rings)
	}
	// No habitable zone treats every world as outer.
	if got, _ := satelliteCount(dice.NewScripted(6), KindWorld, 2, hz, false); got != 3 {
		t.Errorf("satelliteCount(no HZ) = %d, want 3 (outer)", got)
	}
	// A gas giant rolling exactly 0 (1D=1 -> 1D-1=0) yields a Ring, then re-rolls
	// the count (1D=4 -> 3 moons).
	if moons, rings := satelliteCount(
		dice.NewScripted(1, 4),
		KindGasGiant,
		2,
		hz,
		true,
	); moons != 3 ||
		rings != 1 {
		t.Errorf("satelliteCount(GG ring) = %d moons, %d rings, want 3, 1", moons, rings)
	}
}

func TestRollSatellites(t *testing.T) {
	newSys := func() *System {
		s := &System{
			Primary: Star{Type: "F", Decimal: 8, Size: "V"}, // HZ 4
			Orbits:  []PlacedOrbit{{Orbit: 2, Kind: KindGasGiant}},
		}
		s.Mainworld.Profile.Population = 8

		return s
	}
	// A gas giant with moons: each is a real body (type + UWP + orbit letter), and
	// a gas giant's moon is never size-capped, so never a double planet.
	s := newSys()
	s.rollSatellites(dice.NewWithSeed(4))

	moons := s.Orbits[0].Satellites
	if len(moons) == 0 {
		t.Fatal("gas giant got no satellites for seed 4")
	}

	for i, m := range moons {
		if m.Ring {
			continue
		}

		if m.OrbitLetter == "" || m.Profile.String() == "" {
			t.Errorf("moon %d missing letter or UWP: %+v", i, m)
		}

		if m.DoublePlanet {
			t.Errorf("gas-giant moon %d flagged double planet: %+v", i, m)
		}
	}
	// Deterministic for a fixed seed.
	s2 := newSys()
	s2.rollSatellites(dice.NewWithSeed(4))

	if len(s2.Orbits[0].Satellites) != len(moons) {
		t.Errorf(
			"non-deterministic satellite count: %d vs %d",
			len(s2.Orbits[0].Satellites),
			len(moons),
		)
	}
}

func TestSatelliteType(t *testing.T) {
	hz := 4
	// Inner/HZ satellites match the other-world inner table exactly.
	for roll := 1; roll <= 6; roll++ {
		if got, want := satelliteType(
			hz,
			hz,
			true,
			roll,
		), otherWorldType(
			hz,
			hz,
			true,
			roll,
		); got != want {
			t.Errorf("inner satelliteType(roll %d) = %v, want %v", roll, got, want)
		}
	}
	// Outer satellites match the outer other-world table except roll 4, which is a
	// StormWorld (Iceworld for an other world).
	for roll := 1; roll <= 6; roll++ {
		got := satelliteType(hz+3, hz, true, roll)

		want := otherWorldType(hz+3, hz, true, roll)
		if roll == 4 {
			want = worldgen.StormWorld

			if otherWorldType(hz+3, hz, true, 4) != worldgen.Iceworld {
				t.Errorf("fixture: outer other-world roll 4 should be Iceworld")
			}
		}

		if got != want {
			t.Errorf("outer satelliteType(roll %d) = %v, want %v", roll, got, want)
		}
	}
}

// TestRollMoonSizeCap covers the satellite-size rule (Book 3 p.21): an oversized
// moon is cut to its parent's size, and at equal size the pair is a double
// planet. A gas-giant parent never caps.
//
// All-6s dice make a BigWorld of Size 2D+7 = 19 with Flux 0, so uncapped it is
// Atm F / Hyd A — the profile a cap must not leave behind.
func TestRollMoonSizeCap(t *testing.T) {
	cases := []struct {
		name       string
		maxSize    int
		wantSize   int
		wantDouble bool
	}{
		{"oversized cut to parent", 5, 5, true},
		{"cut to a worldlet parent", 3, 3, true},
		{"gas-giant parent never caps", worldgen.NoSizeCap, 19, false},
		// A Size digit of 0 is the asteroid-belt code, not a dimension, so it is
		// not a cap: capping to it flattened every moon of a belt mainworld to
		// Y000000-0, atmosphere and tech level included.
		{"size-0 parent is a belt code, not a cap", 0, 19, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// The leading 3 is the type die: orbit 3 = HZ 3 is hospitable, and
			// the Inner/HZ Satellites table's 3 is a BigWorld. A full moon then
			// draws twenty more dice; every one is a 6.
			m := rollMoon(dice.NewScripted(moonScript(3, 20)...), newSatelliteOrbits(), moonSpec{
				Orbit: 3, HZOrbit: 3, HasHZ: true,
				MWPop: 8, MaxSize: c.maxSize,
			})
			if m.Profile.Size != c.wantSize {
				t.Errorf("Size = %d, want %d", m.Profile.Size, c.wantSize)
			}

			if m.DoublePlanet != c.wantDouble {
				t.Errorf("DoublePlanet = %v, want %v", m.DoublePlanet, c.wantDouble)
			}
		})
	}
}

// TestRollMoonCappedProfileIsConsistent is the regression for #213: capping a
// moon's Size used to overwrite that one field and leave Atmosphere and
// Hydrographics as rolled for the larger world, producing UWPs the World
// Creation chart forbids (p.24: "If Siz=0, Atm=0", "If Siz <2, Hyd =0") and
// trade codes classified off the inconsistent profile. The capped size must
// feed the characteristics that derive from it.
func TestRollMoonCappedProfileIsConsistent(t *testing.T) {
	// Size 1: Atmosphere follows Flux+Siz from the capped size, Hydrographics is
	// forced dry.
	m := rollMoon(dice.NewScripted(moonScript(3, 20)...), newSatelliteOrbits(), moonSpec{
		Orbit: 3, HZOrbit: 3, HasHZ: true,
		MWPop: 8, MaxSize: 1,
	})
	if m.Profile.Atmosphere != 1 {
		t.Errorf("Size-1 moon Atm = %d, want 1 (Flux 0 + Siz 1)", m.Profile.Atmosphere)
	}

	if m.Profile.Hydrographics != 0 {
		t.Errorf("Size-1 moon Hyd = %d, want 0 (p.24 If Siz <2, Hyd =0)",
			m.Profile.Hydrographics)
	}
}

// TestBeltMainworldDoesNotFlattenItsMoons is the regression for the belt-code
// cap: an asteroid-belt mainworld carries UWP Size 0, and reading that digit as
// a satellite cap cut every one of its moons to Size 0 — losing Atmosphere,
// Hydrographics and Tech Level with it, so a Big World rendered Y000000-0 and
// every moon came back flagged a double planet with an As trade code.
func TestBeltMainworldDoesNotFlattenItsMoons(t *testing.T) {
	checked := 0

	for seed := uint64(1); seed <= 40; seed++ {
		s := GenerateForMap(dice.NewWithSeed(seed), true, true)
		if s.Mainworld.Profile.Size != 0 {
			continue // not a belt mainworld for this seed
		}

		for _, o := range s.Orbits {
			if o.Kind != KindMainworld {
				continue
			}

			for _, sat := range o.Satellites {
				if sat.Ring {
					continue
				}

				checked++

				// Worldlet (1D-3) and Planetoids roll Size 0 honestly; the types
				// below have a nonzero minimum (BigWorld 2D+7, StormWorld and
				// RadWorld 2D, Inferno 6+1D), so a Size-0 one can only be a cap
				// artifact.
				switch sat.Type {
				case worldgen.BigWorld, worldgen.StormWorld, worldgen.RadWorld, worldgen.Inferno:
					if sat.Profile.Size == 0 {
						t.Errorf(
							"seed %d: %v moon of a belt mainworld is Size 0 (%s); the belt's Size digit is a code, not a cap",
							seed,
							sat.Type,
							sat.Profile,
						)
					}
				default:
					// Hospitable/InnerWorld/Iceworld/Worldlet/Planetoids may roll 0.
				}

				if sat.DoublePlanet {
					t.Errorf("seed %d: %v moon (%s) flagged a double planet with a belt parent",
						seed, sat.Type, sat.Profile)
				}
			}
		}
	}

	if checked == 0 {
		t.Skip("no belt-mainworld moons in the sampled seeds")
	}

	t.Logf("checked %d moons of belt mainworlds", checked)
}

// TestBeltMainworldRollsNoSatellites is the regression for #309, the count half
// of the same belt-code defect TestBeltMainworldDoesNotFlattenItsMoons covers on
// the cap side. A belt's moons were capped correctly but still *counted*: a belt
// mainworld's orbit is KindMainworld, so satelliteCount's `kind == KindBelt`
// guard never saw it and the belt fell through to the world rule, rolling 1D-5 /
// 1D-4 / 1D-3 moons plus rings around a field of asteroids — and drawing dice for
// them, so every later system in the stream shifted too.
//
// Book 3 p.29 gives a satellite count only to worlds and gas giants; the three
// doc comments in satellites.go say so as well.
func TestBeltMainworldRollsNoSatellites(t *testing.T) {
	checked := 0

	for seed := uint64(1); seed <= 200; seed++ {
		s := GenerateForMap(dice.NewWithSeed(seed), true, true)
		if !s.Mainworld.Profile.IsBelt() {
			continue // not a belt mainworld for this seed
		}

		for _, o := range s.Orbits {
			// A satellite mainworld's orbit belongs to its parent body, whose own
			// moons are legitimate (see satelliteParent); only a belt sitting in
			// its own orbit is at issue.
			if o.Kind != KindMainworld || o.Giant != nil || o.Parent != nil {
				continue
			}

			checked++

			if len(o.Satellites) != 0 {
				t.Errorf("seed %d: belt mainworld in orbit %d has %d satellites (%s); a belt has no moons",
					seed, o.Orbit, len(o.Satellites), orbitLabel(o))
			}
		}
	}

	if checked == 0 {
		t.Fatal("no belt mainworld placed in its own orbit across 200 seeds; the regression is unexercised")
	}

	t.Logf("checked %d belt mainworlds", checked)
}

// TestSatelliteBodyReadsTheBeltCode pins the single decision the #309 fix rests
// on: kind and cap come from one read of the parent's UWP, so a body whose Size
// digit is the asteroid-belt code cannot be classified a world by one half and a
// belt by the other.
func TestSatelliteBodyReadsTheBeltCode(t *testing.T) {
	for _, tc := range []struct {
		name string
		size int
		kind OrbitKind
		cap  int
	}{
		{"belt code", 0, KindBelt, worldgen.NoSizeCap},
		{"ordinary world", 7, KindWorld, 7},
		{"smallest real world", 1, KindWorld, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			kind, maxSize := satelliteBody(uwp.Profile{Size: tc.size})
			if kind != tc.kind || maxSize != tc.cap {
				t.Errorf("satelliteBody(Size %d) = %v, %d; want %v, %d",
					tc.size, kind, maxSize, tc.kind, tc.cap)
			}
		})
	}
}

// TestSatellitesCarryTradeCodes: every generated non-ring satellite carries trade
// codes, where before a non-mainworld satellite carried none (the Sa/Lk logic
// itself is unit-tested in worldgen, where the assembler lives).
// TestSatelliteMainworldOrbitRollsForItsParent is the regression for #214. When
// the mainworld is itself a satellite, the body occupying the orbit is its
// parent — Book 3 p.21 places a gas giant (or a BigWorld) in the mainworld's
// orbit to accommodate it, and p.29 rolls satellites "for each world in the
// system" against that body. So the orbit's moons are the mainworld's siblings
// around the parent: counted by the parent's rule, capped by the parent's size.
//
// Before the fix the orbit was rolled as the mainworld's own: a gas-giant parent
// got the world zone count (1D-4 here, not 1D-1) and its moons were capped to the
// mainworld — a fellow moon.
func TestSatelliteMainworldOrbitRollsForItsParent(t *testing.T) {
	newSys := func() *System {
		s := &System{
			Primary: Star{Type: "F", Decimal: 8, Size: "V"}, // HZ 4
			Orbits: []PlacedOrbit{{
				Host: "Primary", Orbit: 4, Kind: KindMainworld,
				Giant: &GasGiant{Size: 26, Class: LargeGasGiant},
			}},
			MainworldSatellite: MainworldSatellite{IsSatellite: true, Far: true},
		}
		s.Mainworld.Profile.Population = 8
		s.Mainworld.Profile.Size = 3

		return s
	}
	// Orbit 4 is the habitable zone. A count die of 5 is 1D-1 = 4 moons under the
	// gas giant's rule, but 1D-4 = 1 under the world rule the orbit used to take.
	// Every later die is a 6: type 6 is Hospitable, whose Size is rollSize's 2D-2 =
	// 10 rerolled as 1D+9 = 15 — uncapped, or cut to the Size-3 mainworld (and
	// flagged a double planet) under the old parent.
	s := newSys()
	s.rollSatellites(dice.NewSource(countThenSixes(5)))

	moons := s.Orbits[0].Satellites
	if len(moons) != 4 {
		t.Fatalf("satellite mainworld orbit rolled %d moons, want 4 (the giant\u2019s 1D-1)", len(moons))
	}

	for i, m := range moons {
		if m.Profile.Size != 15 {
			t.Errorf("moon %d Size = %d, want 15 (a gas giant caps nothing)", i, m.Profile.Size)
		}

		if m.DoublePlanet {
			t.Errorf("moon %d is flagged a double planet with a gas-giant parent", i)
		}
	}
}

// TestSatelliteMainworldBigWorldParentCapsItsMoons is the other half of #214: with
// no gas giant in the system the mainworld rides an accommodating BigWorld, so
// that BigWorld — not the mainworld — is the parent whose size caps the orbit's
// other moons (Book 3 p.29).
func TestSatelliteMainworldBigWorldParentCapsItsMoons(t *testing.T) {
	s := &System{
		Primary: Star{Type: "F", Decimal: 8, Size: "V"}, // HZ 4
		Orbits: []PlacedOrbit{{
			Host: "Primary", Orbit: 4, Kind: KindMainworld,
			Parent: &OtherWorld{
				Type:    worldgen.BigWorld,
				Profile: uwp.Profile{Size: 11},
			},
		}},
		MainworldSatellite: MainworldSatellite{IsSatellite: true},
	}
	s.Mainworld.Profile.Population = 8
	s.Mainworld.Profile.Size = 3
	// Count die 5 with the world rule 1D-4 = 1 moon (the BigWorld parent is a
	// world, so the zone count is unchanged), then an all-6s Hospitable moon: Size
	// 15 uncapped, cut to the parent's 11.
	s.rollSatellites(dice.NewSource(countThenSixes(5)))

	moons := s.Orbits[0].Satellites
	if len(moons) != 1 {
		t.Fatalf("rolled %d moons, want 1", len(moons))
	}

	if got := moons[0].Profile.Size; got != 11 {
		t.Errorf("moon Size = %d, want 11 (capped to the BigWorld parent, not the Size-3 mainworld)", got)
	}

	if !moons[0].DoublePlanet {
		t.Error("a moon at exactly its parent\u2019s size is a double planet (Book 3 p.29)")
	}
}

// TestSiblingMoonsGetDistinctOrbitLetters is the regression for #216. A moon's
// orbit letter is an orbit *name* (Book 3 p.24 table 2C), rolled independently
// per moon, so a parent with several moons could roll the same letter twice and
// render two bodies in one named orbit. Book 3 p.29's placement note — "If an
// orbit is duplicated or precluded, adjust to an adjacent or the closest possible
// orbit" — resolves it, so a duplicate is nudged to the nearest free letter.
//
// Every die is a 6: a gas giant rolls 1D-1 = 5 moons, and each one is Far (2D=12)
// with Flux 0, which is farOrbitLetters[6] = "Tee" for all five.
func TestSiblingMoonsGetDistinctOrbitLetters(t *testing.T) {
	s := &System{
		Primary: Star{Type: "F", Decimal: 8, Size: "V"},
		Orbits:  []PlacedOrbit{{Host: "Primary", Orbit: 2, Kind: KindGasGiant}},
	}
	s.Mainworld.Profile.Population = 8
	s.rollSatellites(dice.NewSource(func() int { return 6 }))

	moons := s.Orbits[0].Satellites
	if len(moons) != 5 {
		t.Fatalf("rolled %d moons, want 5", len(moons))
	}

	seen := map[string]bool{}

	for _, m := range moons {
		if seen[m.OrbitLetter] {
			t.Errorf("orbit letter %q used twice around one parent: %v",
				m.OrbitLetter, letters(moons))
		}

		seen[m.OrbitLetter] = true

		if !m.Far {
			t.Errorf("moon %q should be Far (2D=12)", m.OrbitLetter)
		}
	}
	// The rolled letter is kept for the first moon; the rest spiral outward from
	// it, inward before outward, as orbitHost.claim does for world orbits.
	if got, want := letters(moons), []string{"Tee", "Ess", "Yu", "Arr", "Vee"}; !slices.Equal(got, want) {
		t.Errorf("letters = %v, want %v", got, want)
	}
}

func letters(moons []Satellite) []string {
	got := make([]string, len(moons))
	for i, m := range moons {
		got[i] = m.OrbitLetter
	}

	return got
}

// TestNoParentHasDuplicateOrbitLetters covers both moon-creation sites at once:
// the satellite pass and the orbit map's gas-giant-captured world append to the
// same parent, so they must not collide with each other either.
func TestNoParentHasDuplicateOrbitLetters(t *testing.T) {
	for seed := uint64(1); seed <= 500; seed++ {
		s := Generate(dice.NewWithSeed(seed))
		for _, o := range s.Orbits {
			seen := map[string]bool{}
			// A satellite mainworld shares its parent with the orbit's moons, so
			// its own letter is taken too.
			if o.Kind == KindMainworld && (o.Giant != nil || o.Parent != nil) {
				seen[s.MainworldSatellite.OrbitLetter] = true
			}

			for _, m := range o.Satellites {
				if m.Ring {
					continue
				}

				if seen[m.OrbitLetter] {
					t.Errorf("seed %d: %s orbit %d has two moons in orbit %q: %v",
						seed, o.Kind, o.Orbit, m.OrbitLetter, letters(o.Satellites))
				}

				seen[m.OrbitLetter] = true
			}
		}
	}
}

func TestSatellitesCarryTradeCodes(t *testing.T) {
	sys := Generate(dice.NewWithSeed(11))
	found := false

	for _, o := range sys.Orbits {
		for _, sat := range o.Satellites {
			if sat.Ring {
				continue
			}

			found = true

			if len(sat.TradeCodes) == 0 {
				t.Errorf("satellite %s carries no trade codes", sat.OrbitLetter)
			}
		}
	}

	if !found {
		t.Skip("this seed produced no non-ring satellites")
	}
}
