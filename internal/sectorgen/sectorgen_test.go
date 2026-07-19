package sectorgen

import (
	"strings"
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

func TestHexValidAndOffMapRendering(t *testing.T) {
	for _, h := range []Hex{{1, 1}, {32, 40}, {8, 10}} {
		if !h.Valid() {
			t.Errorf("%v should be a valid hex", h)
		}
	}
	// Off-map hexes: past either edge, at or below zero, and negative.
	offMap := map[Hex]string{
		{33, 1}:  "????", // one column past the sector
		{9, 41}:  "????", // one row past the sector
		{0, 0}:   "????", // zero is not a coordinate; CCRR starts at 0101
		{100, 5}: "????", // five digits would not survive ParseHex
		{-1, -1}: "????",
	}
	for h, want := range offMap {
		if h.Valid() {
			t.Errorf("%d,%d should not be a valid hex", h.Col, h.Row)
		}

		if got := h.String(); got != want {
			t.Errorf("Hex{%d,%d}.String() = %q, want %q", h.Col, h.Row, got, want)
		}
	}
}

// TestSubsectorPanicsOffMap locks the strict half of the split: Subsector's result
// is a filing decision (survey groups its records by it), so an off-map hex must
// fail loudly rather than hand back a plausible letter for the wrong subsector.
func TestSubsectorPanicsOffMap(t *testing.T) {
	for _, h := range []Hex{{33, 1}, {9, 41}, {0, 0}, {-1, 5}, {100, 5}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Hex{%d,%d}.Subsector() did not panic", h.Col, h.Row)
				}
			}()

			t.Errorf("Hex{%d,%d}.Subsector() = %c, want a panic", h.Col, h.Row, h.Subsector())
		}()
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

func TestHexDistance(t *testing.T) {
	// Distance to self is 0.
	if d := (Hex{5, 5}).Distance(Hex{5, 5}); d != 0 {
		t.Errorf("distance to self = %d, want 0", d)
	}
	// All six neighbors of an interior hex are 1 parsec away (even-q layout:
	// even columns shifted down, so 0505's cross-column neighbors are rows 4-5).
	for _, n := range []Hex{{5, 4}, {5, 6}, {4, 4}, {4, 5}, {6, 4}, {6, 5}} {
		if d := (Hex{5, 5}).Distance(n); d != 1 {
			t.Errorf("%s neighbor distance = %d, want 1", n, d)
		}
	}
	// A few known spans, and symmetry.
	cases := []struct {
		a, b Hex
		want int
	}{
		{Hex{1, 1}, Hex{3, 1}, 2},    // two columns, same row
		{Hex{1, 1}, Hex{1, 4}, 3},    // same column, three rows
		{Hex{1, 1}, Hex{32, 40}, 55}, // opposite corners
	}
	for _, c := range cases {
		if d := c.a.Distance(c.b); d != c.want {
			t.Errorf("%s-%s distance = %d, want %d", c.a, c.b, d, c.want)
		}

		if d := c.b.Distance(c.a); d != c.want {
			t.Errorf("%s-%s distance = %d, want %d (asymmetric)", c.b, c.a, d, c.want)
		}
	}
}

func TestDensityByName(t *testing.T) {
	// Case- and space-insensitive.
	cases := map[string]Density{
		"standard":       Standard,
		"Core":           Core,
		"extragalactic":  ExtraGalactic,
		"Extra Galactic": ExtraGalactic,
	}
	for name, want := range cases {
		if d, ok := DensityByName(name); !ok || d != want {
			t.Errorf("DensityByName(%q) = %v,%v, want %v", name, d, ok, want)
		}
	}

	if _, ok := DensityByName("bogus"); ok {
		t.Errorf("DensityByName(bogus) should not be found")
	}

	if names := DensityNames(); len(names) != 8 || names[0] != "Extra Galactic" ||
		names[7] != "Core" {
		t.Errorf("DensityNames() = %v", names)
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
	// Core = 2D <= 10 (see densityInfo): 10 present, 11 absent.
	if !SystemPresent(dice.NewScripted(4, 6), Core) {
		t.Errorf("Core 2D=10 should be present")
	}

	if SystemPresent(dice.NewScripted(5, 6), Core) {
		t.Errorf("Core 2D=11 should be absent")
	}
	// Rift = 2D <= 2 (a natural 2 only).
	if !SystemPresent(dice.NewScripted(1, 1), Rift) || SystemPresent(dice.NewScripted(1, 2), Rift) {
		t.Errorf("Rift should require a natural 2")
	}
	// An out-of-range density yields no system rather than panicking.
	if SystemPresent(dice.NewScripted(1), Density(99)) {
		t.Errorf("out-of-range density should report no system")
	}

	if GenerateSector(dice.NewScripted(1), Density(-1)) != nil {
		t.Errorf("out-of-range density should generate no systems")
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

func TestParseSubsector(t *testing.T) {
	for _, s := range []string{"A", "a", "P", "p"} {
		if letter, ok := ParseSubsector(s); !ok || letter != strings.ToUpper(s)[0] {
			t.Errorf("ParseSubsector(%q) = %c,%v, want the upper-case letter", s, letter, ok)
		}
	}
	// Out of range, empty, and — the trap a bare first-byte read would miss —
	// anything longer than one letter.
	for _, bad := range []string{"", "Q", "z", "1", "AB", "Applesauce"} {
		if _, ok := ParseSubsector(bad); ok {
			t.Errorf("ParseSubsector(%q) should be rejected", bad)
		}
	}
}

func TestParseHexRoundTrip(t *testing.T) {
	h, ok := ParseHex("0436")
	if !ok || h.Col != 4 || h.Row != 36 || h.String() != "0436" {
		t.Errorf("ParseHex(0436) = %v,%v, want col 4 row 36", h, ok)
	}
	// Malformed, out-of-range, and — the trap — signed halves, which strconv would
	// otherwise accept ("+436" silently parsing as hex 0436).
	for _, bad := range []string{"", "436", "04366", "zzzz", "0041", "3341", "0000", "+436", "04+3", "-436", " 436", "04 3"} {
		if _, ok := ParseHex(bad); ok {
			t.Errorf("ParseHex(%q) should be rejected", bad)
		}
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
	if sparse >= len(a) || len(a) >= core {
		t.Errorf("density ordering wrong: sparse %d, standard %d, core %d", sparse, len(a), core)
	}
}
