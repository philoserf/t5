package systemgen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
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
	// A gas giant in orbit 2 (1D-1). 1D=4 -> 3 moons. Each: 2D for Close/Far,
	// Flux for the orbit letter (Flux 0 -> index 6 -> "Gee" close / "Tee" far).
	// moon0 Close (2D=5), moon1 Far (2D=9), moon2 Close (2D=7).
	s := &System{
		Primary: Star{Type: "F", Decimal: 8, Size: "V"}, // HZ 4
		Orbits:  []PlacedOrbit{{Orbit: 2, Kind: KindGasGiant}},
	}
	s.rollSatellites(dice.NewScripted(4 /*count*/, 2, 3 /*Close*/, 3, 3 /*Flux 0*/, 4, 5 /*Far*/, 3, 3, 3, 4 /*Close*/, 3, 3))
	moons := s.Orbits[0].Satellites
	if len(moons) != 3 {
		t.Fatalf("got %d moons, want 3: %+v", len(moons), moons)
	}
	if moons[0].Far || moons[0].OrbitLetter != "Gee" {
		t.Errorf("moon 0 = %+v, want Close Gee", moons[0])
	}
	if !moons[1].Far || moons[1].OrbitLetter != "Tee" {
		t.Errorf("moon 1 = %+v, want Far Tee", moons[1])
	}
}
