package sophont

import (
	"math"
	"testing"
)

var humanChars = [6]CharSpec{{Str, 2}, {Dex, 2}, {End, 2}, {Int, 2}, {Edu, 2}, {Soc, 2}}

// TestHumanSize is the benchmark (Book 3 p.236): an all-2D Human is Size 72.
func TestHumanSize(t *testing.T) {
	if got := Size(humanChars); got != 72 {
		t.Errorf("Human Size = %d, want 72", got)
	}
}

// TestSizeWeightings checks the physical-characteristic weightings (chart 14A):
// Stamina doubles, Grace/Agility/Vigor halve.
func TestSizeWeightings(t *testing.T) {
	cases := []struct {
		name string
		c2   CharName
		c3   CharName
		want int
	}{
		{"baseline Dex/End", Dex, End, 72}, // (2+2+2)*12 = 72
		{"double Stamina", Dex, Sta, 96},   // (2+2+4)*12 = 96
		{"half Vigor", Dex, Vig, 60},       // (2+2+1)*12 = 60
		{"half Grace", Gra, End, 60},       // (2+1+2)*12 = 60
	}
	for _, c := range cases {
		chars := [6]CharSpec{{Str, 2}, {c.c2, 2}, {c.c3, 2}, {Int, 2}, {Edu, 2}, {Soc, 2}}
		if got := Size(chars); got != c.want {
			t.Errorf("%s: Size = %d, want %d", c.name, got, c.want)
		}
	}
}

// TestBulkMultiplier checks the super-linear Bulk scaling for a big-Str species
// (chart 14C): Str 5D / Dex 3D / End 2D totals 10 dice at the 5D multiplier ×60
// = 600. (This follows the Bulk table; the page's contradictory prose figure of
// 1200 is a known book error.)
func TestBulkMultiplier(t *testing.T) {
	chars := [6]CharSpec{{Str, 5}, {Dex, 3}, {End, 2}, {Int, 2}, {Edu, 2}, {Soc, 2}}
	if got := Size(chars); got != 600 {
		t.Errorf("Bulk Size = %d, want 600", got)
	}
}

// TestHeight locks the height formula against the book's precomputed grid
// (Book 3 p.237): a Size-72 Human is ~1.8 m at BFP 9, ~1.2 m at BFP 5, ~2.5 m at
// BFP 15; a Size-180 sophont is ~2.4 m at BFP 9.
func TestHeight(t *testing.T) {
	cases := []struct {
		size int
		bfp  float64
		want float64
	}{
		{72, 9, 1.8}, {72, 5, 1.2}, {72, 15, 2.5}, {180, 9, 2.4},
	}
	for _, c := range cases {
		if got := Height(c.size, c.bfp); math.Abs(got-c.want) > 0.05 {
			t.Errorf("Height(%d, %g) = %.2f, want ~%.1f", c.size, c.bfp, got, c.want)
		}
	}
}
