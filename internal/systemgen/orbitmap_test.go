package systemgen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestResolveOrbit(t *testing.T) {
	occupied := map[int]bool{4: true, 5: true}
	// Free target stays put.
	if got := resolveOrbit(6, 0, occupied); got != 6 {
		t.Errorf("resolveOrbit(6) = %d, want 6", got)
	}
	// Occupied target nudges inward first (closest free is 3).
	if got := resolveOrbit(4, 0, occupied); got != 3 {
		t.Errorf("resolveOrbit(4) = %d, want 3", got)
	}
	// Below the floor clamps up, then spirals out past occupied.
	if got := resolveOrbit(-2, 4, occupied); got != 6 {
		t.Errorf("resolveOrbit(-2, floor 4) = %d, want 6", got)
	}
}

func TestPlaceOrbits(t *testing.T) {
	// Primary F8 V: HZ orbit 4, floor 0. Mainworld at 4. Two giants (P/SGG,
	// T/LGG), one belt, and two other worlds (Worlds = 6 - accounts for MW +
	// 2 giants + 1 belt + 2 others).
	s := &System{
		Primary:        Star{Type: "F", Decimal: 8, Size: "V"},
		GasGiants:      2,
		Belts:          1,
		Worlds:         6,
		MainworldOrbit: 4,
		Giants: []GasGiant{
			{Size: 23, Class: SmallGasGiant}, // P
			{Size: 27, Class: LargeGasGiant}, // T
		},
	}
	// Rolls, in placement order: GG1 2D=6 (SGG +2 -> 6), GG2 2D=8 (LGG +3 -> 7),
	// Belt 2D=4 (+1 -> 5), World1 2D=5 (abs 4, taken -> nudged to 3),
	// World2 2D=10 (abs 9).
	s.placeOrbits(dice.NewScripted(3, 3, 4, 4, 2, 2, 2, 3, 5, 5))

	want := []struct {
		orbit int
		kind  string
	}{
		{3, "World"},
		{4, "Mainworld"},
		{5, "Belt"},
		{6, "Gas Giant"},
		{7, "Gas Giant"},
		{9, "World"},
	}
	if len(s.Orbits) != len(want) {
		t.Fatalf("placed %d orbits, want %d: %+v", len(s.Orbits), len(want), s.Orbits)
	}
	for i, w := range want {
		if s.Orbits[i].Orbit != w.orbit || s.Orbits[i].Kind != w.kind {
			t.Errorf("orbit %d = {%d %q}, want {%d %q}",
				i, s.Orbits[i].Orbit, s.Orbits[i].Kind, w.orbit, w.kind)
		}
	}
	// The gas giants keep their identity in the map.
	if g := s.Orbits[3].Giant; g == nil || g.Size != 23 {
		t.Errorf("orbit 6 giant = %v, want size 23 (P SGG)", s.Orbits[3].Giant)
	}
	if g := s.Orbits[4].Giant; g == nil || g.Size != 27 {
		t.Errorf("orbit 7 giant = %v, want size 27 (T LGG)", s.Orbits[4].Giant)
	}
}
