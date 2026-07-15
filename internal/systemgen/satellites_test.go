package systemgen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

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
		if got, rings := satelliteCount(dice.NewScripted(6), c.kind, c.orbit, hz, true); got != c.want || rings != 0 {
			t.Errorf("satelliteCount(%s, orbit %d) = %d moons, %d rings, want %d moons, 0 rings", c.kind, c.orbit, got, rings, c.want)
		}
	}
	// A negative roll is none (inner 1D-5 with 1D=1 -> -4 -> none).
	if got, rings := satelliteCount(dice.NewScripted(1), KindWorld, 2, hz, true); got != 0 || rings != 0 {
		t.Errorf("satelliteCount(inner, 1D=1) = %d moons, %d rings, want 0, 0", got, rings)
	}
	// No habitable zone treats every world as outer.
	if got, _ := satelliteCount(dice.NewScripted(6), KindWorld, 2, hz, false); got != 3 {
		t.Errorf("satelliteCount(no HZ) = %d, want 3 (outer)", got)
	}
	// A gas giant rolling exactly 0 (1D=1 -> 1D-1=0) yields a Ring, then re-rolls
	// the count (1D=4 -> 3 moons).
	if moons, rings := satelliteCount(dice.NewScripted(1, 4), KindGasGiant, 2, hz, true); moons != 3 || rings != 1 {
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
		t.Errorf("non-deterministic satellite count: %d vs %d", len(s2.Orbits[0].Satellites), len(moons))
	}
}

func TestSatelliteType(t *testing.T) {
	hz := 4
	// Inner/HZ satellites match the other-world inner table exactly.
	for roll := 1; roll <= 6; roll++ {
		if got, want := satelliteType(hz, hz, true, roll), otherWorldType(hz, hz, true, roll); got != want {
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

func TestCapSatelliteSize(t *testing.T) {
	cases := []struct {
		sat, parent int
		capped      bool
		wantSize    int
		wantDouble  bool
	}{
		{10, 5, true, 5, true},    // oversized -> cut to parent, double planet
		{5, 5, true, 5, true},     // equal -> double planet
		{3, 5, true, 3, false},    // smaller -> unchanged
		{10, 0, false, 10, false}, // gas-giant parent: never capped
	}
	for _, c := range cases {
		size, double := capSatelliteSize(c.sat, c.parent, c.capped)
		if size != c.wantSize || double != c.wantDouble {
			t.Errorf("capSatelliteSize(%d, %d, %v) = %d,%v, want %d,%v",
				c.sat, c.parent, c.capped, size, double, c.wantSize, c.wantDouble)
		}
	}
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
