package systemgen

import (
	"slices"
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
