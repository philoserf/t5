package systemgen

import (
	"slices"
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// script builds a rollMoon die script: the leading type die, then n sixes for
// everything the moon rolls after it.
func script(typeDie, n int) []int {
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
			m := rollMoon(dice.NewScripted(script(3, 20)...), moonSpec{
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
	m := rollMoon(dice.NewScripted(script(3, 20)...), moonSpec{
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

// TestSatellitesCarryTradeCodes: every generated non-ring satellite carries trade
// codes, where before a non-mainworld satellite carried none (the Sa/Lk logic
// itself is unit-tested in worldgen, where the assembler lives).
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
