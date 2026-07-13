package systemgen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestGasGiantDetail(t *testing.T) {
	cases := []struct {
		size     int
		diameter int
		skimG    float64
		class    GGClass
	}{
		{20, 20000, 0.2, SmallGasGiant},  // L, row 1
		{21, 30000, 0.3, SmallGasGiant},  // M (Neptune)
		{25, 70000, 0.7, SmallGasGiant},  // R (Saturn) — last Small
		{26, 80000, 0.8, LargeGasGiant},  // S — first Large
		{27, 90000, 0.9, LargeGasGiant},  // T (Jupiter)
		{32, 250000, 3.0, LargeGasGiant}, // Y, row 13 (brown dwarf)
	}
	for _, c := range cases {
		g := gasGiantDetail(c.size)
		if g.Diameter != c.diameter || g.SkimG != c.skimG || g.Class != c.class {
			t.Errorf("gasGiantDetail(%d) = %+v, want diameter %d skimG %g class %v",
				c.size, g, c.diameter, c.skimG, c.class)
		}
	}
}

func TestRollGasGiants(t *testing.T) {
	// Four gas giants, sized 2D each: SGG (23), LGG (27), SGG (25), SGG (23).
	// Every second Small Gas Giant becomes an Ice Giant, so the 3rd (the 2nd
	// Small) converts: SGG, LGG, IG, SGG.
	r := dice.NewScripted(2, 2 /*2D=4 -> 23*/, 4, 4 /*2D=8 -> 27*/, 3, 3 /*2D=6 -> 25*/, 2, 2 /*2D=4 -> 23*/)
	giants := rollGasGiants(r, 4)
	want := []struct {
		size  int
		class GGClass
	}{
		{23, SmallGasGiant},
		{27, LargeGasGiant},
		{25, IceGiant},
		{23, SmallGasGiant},
	}
	if len(giants) != len(want) {
		t.Fatalf("rollGasGiants returned %d giants, want %d", len(giants), len(want))
	}
	for i, w := range want {
		if giants[i].Size != w.size || giants[i].Class != w.class {
			t.Errorf("giant %d = %s (size %d), want size %d class %v",
				i, giants[i], giants[i].Size, w.size, w.class)
		}
	}
}

func TestGGClassString(t *testing.T) {
	cases := map[GGClass]string{SmallGasGiant: "SGG", LargeGasGiant: "LGG", IceGiant: "IG"}
	for c, want := range cases {
		if got := c.String(); got != want {
			t.Errorf("%d.String() = %q, want %q", c, got, want)
		}
	}
}
