package systemgen

import (
	"slices"
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
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

func TestOtherWorldType(t *testing.T) {
	hz := 4
	// Inner/HZ table (orbit <= hz+1): rolls 1..6.
	inner := map[int]string{1: "Inferno", 2: "Inner World", 3: "Big World", 4: "Storm World", 5: "Rad World", 6: "Hospitable"}
	for roll, want := range inner {
		if got := otherWorldType(hz, hz, true, roll).String(); got != want {
			t.Errorf("otherWorldType(HZ, roll %d) = %q, want %q", roll, got, want)
		}
	}
	// Outer table (orbit > hz+1): rolls 1..6.
	outer := map[int]string{1: "Worldlet", 2: "Iceworld", 3: "Big World", 4: "Iceworld", 5: "Rad World", 6: "Iceworld"}
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
	s.placeOrbits(dice.NewScripted(3, 3, 4, 4, 2, 2, 5, 5, 2, 2, 2, 3, 3, 3, 3, 6, 6, 3, 3, 3, 3, 2, 3))

	want := []struct {
		orbit int
		kind  string
	}{
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
