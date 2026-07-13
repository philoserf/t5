package sectorgen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestHexString(t *testing.T) {
	cases := map[Hex]string{
		{1, 1}:   "0101",
		{8, 3}:   "0803",
		{32, 40}: "3240",
		{10, 5}:  "1005",
	}
	for h, want := range cases {
		if got := h.String(); got != want {
			t.Errorf("%v.String() = %q, want %q", h, got, want)
		}
	}
}

func TestSubsector(t *testing.T) {
	// The sixteen subsectors are lettered A-P left-to-right, top-to-bottom.
	cases := map[Hex]byte{
		{1, 1}:   'A', // upper-left
		{8, 10}:  'A', // lower-right of A
		{9, 1}:   'B',
		{25, 1}:  'D',
		{1, 11}:  'E', // second row band
		{25, 21}: 'L',
		{1, 31}:  'M',
		{32, 40}: 'P', // lower-right of the sector
	}
	for h, want := range cases {
		if got := h.Subsector(); got != want {
			t.Errorf("%s.Subsector() = %c, want %c", h, got, want)
		}
	}
}

func TestSystemPresent(t *testing.T) {
	// Standard = 1D <= 3.
	if !SystemPresent(dice.NewScripted(3), Standard) {
		t.Errorf("Standard 1D=3 should be present")
	}
	if SystemPresent(dice.NewScripted(4), Standard) {
		t.Errorf("Standard 1D=4 should be absent")
	}
	// Core = 2D <= 11: 11 present, 12 absent.
	if !SystemPresent(dice.NewScripted(5, 6), Core) {
		t.Errorf("Core 2D=11 should be present")
	}
	if SystemPresent(dice.NewScripted(6, 6), Core) {
		t.Errorf("Core 2D=12 should be absent")
	}
	// Rift = 2D <= 2 (a natural 2 only).
	if !SystemPresent(dice.NewScripted(1, 1), Rift) || SystemPresent(dice.NewScripted(1, 2), Rift) {
		t.Errorf("Rift should require a natural 2")
	}
}

func TestRollContents(t *testing.T) {
	// Gas giant on 2D<=8, asteroid mainworld on 2D<=2.
	sh := rollContents(dice.NewScripted(4, 4, 1, 1), Hex{8, 3}) // GG 2D=8 (yes), Ast 2D=2 (yes)
	if !sh.GasGiant || !sh.AsteroidMainworld || sh.Hex.String() != "0803" {
		t.Errorf("contents = %+v, want GG + asteroid at 0803", sh)
	}
	sh2 := rollContents(dice.NewScripted(5, 4, 3, 4), Hex{1, 1}) // GG 2D=9 (no), Ast 2D=7 (no)
	if sh2.GasGiant || sh2.AsteroidMainworld {
		t.Errorf("contents = %+v, want no GG, no asteroid", sh2)
	}
}

func TestGenerateSubsector(t *testing.T) {
	// Every system generated for subsector C lands in subsector C, and there are
	// at most 80 of them.
	systems := GenerateSubsector(dice.NewWithSeed(7), Dense, 'C')
	if len(systems) == 0 || len(systems) > 80 {
		t.Fatalf("subsector C has %d systems, want 1..80", len(systems))
	}
	for _, s := range systems {
		if s.Hex.Subsector() != 'C' {
			t.Errorf("hex %s is in subsector %c, not C", s.Hex, s.Hex.Subsector())
		}
	}
	// An out-of-range letter yields nothing.
	if GenerateSubsector(dice.NewWithSeed(7), Dense, 'Z') != nil {
		t.Errorf("subsector Z should yield no systems")
	}
}

func TestGenerateSectorDeterministicAndDense(t *testing.T) {
	a := GenerateSector(dice.NewWithSeed(42), Standard)
	b := GenerateSector(dice.NewWithSeed(42), Standard)
	if len(a) != len(b) {
		t.Fatalf("same seed gave different counts: %d vs %d", len(a), len(b))
	}
	// Standard density is 50%, so roughly half of 1280 hexes hold a system.
	if len(a) < 500 || len(a) > 780 {
		t.Errorf("Standard sector has %d systems, want ~640", len(a))
	}
	// Denser regions hold more systems than sparser ones.
	sparse := len(GenerateSector(dice.NewWithSeed(42), Sparse))
	core := len(GenerateSector(dice.NewWithSeed(42), Core))
	if !(sparse < len(a) && len(a) < core) {
		t.Errorf("density ordering wrong: sparse %d, standard %d, core %d", sparse, len(a), core)
	}
}
