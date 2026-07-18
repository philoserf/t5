package worldgen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

// The World Creation chart's structural rules (Book 3 p.24) are conditions on
// Size, not on the formula that produced the value: "If Atm<0 or Siz=0, Atm=0"
// and "If Siz <2, Hyd =0". A world type carrying its own Atm/Hyd formula — one
// that never reads Size — must still pass through them, which is what
// sizedAtmosphere and sizedHydrographics are for.

func TestSizedAtmosphereSizeRule(t *testing.T) {
	// Size 0 is airless whatever the type's formula rolled.
	for _, atm := range []int{0, 5, 12, maxAtmosphere} {
		if got := sizedAtmosphere(atm, 0); got != 0 {
			t.Errorf("sizedAtmosphere(%d, 0) = %d, want 0 (Siz=0, Atm=0)", atm, got)
		}
	}

	// Above Size 0 the value passes through, floored at 0 and capped at F.
	cases := []struct{ atm, size, want int }{
		{6, 1, 6},
		{-3, 4, 0},
		{maxAtmosphere + 4, 4, maxAtmosphere},
	}
	for _, c := range cases {
		if got := sizedAtmosphere(c.atm, c.size); got != c.want {
			t.Errorf("sizedAtmosphere(%d, %d) = %d, want %d", c.atm, c.size, got, c.want)
		}
	}
}

func TestSizedHydrographicsSizeRule(t *testing.T) {
	// A world smaller than Size 2 holds no water, whatever the formula rolled.
	for size := range 2 {
		for _, hyd := range []int{0, 4, maxHydrographics} {
			if got := sizedHydrographics(hyd, size); got != 0 {
				t.Errorf("sizedHydrographics(%d, %d) = %d, want 0 (Siz <2, Hyd =0)", hyd, size, got)
			}
		}
	}

	// From Size 2 up the value passes through, floored at 0 and capped at A.
	cases := []struct{ hyd, size, want int }{
		{5, 2, 5},
		{-2, 7, 0},
		{maxHydrographics + 3, 7, maxHydrographics},
	}
	for _, c := range cases {
		if got := sizedHydrographics(c.hyd, c.size); got != c.want {
			t.Errorf("sizedHydrographics(%d, %d) = %d, want %d", c.hyd, c.size, got, c.want)
		}
	}
}

// TestInnerWorldSizeZeroHasNoHydrographics is the #210 regression: an Inner
// World rolls its own Hyd (2D-4) without consulting Size, so before the shared
// guard a Size-0 Inner World could come out an airless asteroid with oceans.
// rollSize is 2D-2, so scripted 1s give Size 0.
func TestInnerWorldSizeZeroHasNoHydrographics(t *testing.T) {
	p := GenerateOtherWorld(dice.NewScripted(1), InnerWorld, 8)

	if p.Size != 0 {
		t.Fatalf("Size = %d, want 0 (2D-2 on scripted 1s)", p.Size)
	}

	if p.Atmosphere != 0 {
		t.Errorf("Atmosphere = %d, want 0 (Book 3 p.24: Siz=0, Atm=0)", p.Atmosphere)
	}

	if p.Hydrographics != 0 {
		t.Errorf("Hydrographics = %d, want 0 (Book 3 p.24: Siz <2, Hyd =0)", p.Hydrographics)
	}
}
