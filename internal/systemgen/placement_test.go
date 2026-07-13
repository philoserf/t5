package systemgen

import "testing"

func TestSurfaceOrbit(t *testing.T) {
	cases := []struct {
		star Star
		want int
	}{
		{Star{Type: "A", Decimal: 0, Size: "Ia"}, 4},  // A0-F5 band
		{Star{Type: "G", Decimal: 0, Size: "Ia"}, 5},  // G0
		{Star{Type: "M", Decimal: 9, Size: "Ia"}, 9},  // M5-M9
		{Star{Type: "A", Decimal: 0, Size: "Ib"}, 1},  // A0
		{Star{Type: "M", Decimal: 9, Size: "Ib"}, 8},  // M5-M9
		{Star{Type: "A", Decimal: 0, Size: "II"}, 0},  // A0-F5
		{Star{Type: "M", Decimal: 9, Size: "II"}, 7},  // M9
		{Star{Type: "K", Decimal: 0, Size: "III"}, 0}, // A0-K0
		{Star{Type: "K", Decimal: 5, Size: "III"}, 1}, // K5
		{Star{Type: "M", Decimal: 9, Size: "III"}, 6}, // M9
		{Star{Type: "F", Decimal: 8, Size: "V"}, -1},  // main sequence: precludes none
		{Star{Type: "G", Decimal: 2, Size: "IV"}, -1}, // subgiant: none
		{Star{Type: "M", Decimal: 0, Size: "VI"}, -1}, // subdwarf: none
		{Star{Type: "K", Decimal: -1, Size: "D"}, -1}, // white dwarf: none
	}
	for _, c := range cases {
		if got := surfaceOrbit(c.star); got != c.want {
			t.Errorf("surfaceOrbit(%s) = %d, want %d", c.star, got, c.want)
		}
	}
	// firstOrbit is one beyond the surface; Regina's F8 V primary starts at 0.
	if got := firstOrbit(Star{Type: "F", Decimal: 8, Size: "V"}); got != 0 {
		t.Errorf("firstOrbit(F8 V) = %d, want 0", got)
	}
	if got := firstOrbit(Star{Type: "A", Decimal: 0, Size: "Ia"}); got != 5 {
		t.Errorf("firstOrbit(A0 Ia) = %d, want 5", got)
	}
}

func TestP2Chart(t *testing.T) {
	// Spot-check the anchors: HZ entries are 0, and the extremes.
	if r := p2(5); r.lgg != 0 { // roll 5: LGG at HZ
		t.Errorf("p2(5).lgg = %d, want 0", r.lgg)
	}
	if r := p2(7); r.world1 != 0 || r.world2 != 12 {
		t.Errorf("p2(7) world1/world2 = %d/%d, want 0/12", r.world1, r.world2)
	}
	if r := p2(1); r.belt != -2 || r.ig != 0 {
		t.Errorf("p2(1) belt/ig = %d/%d, want -2/0", r.belt, r.ig)
	}
	if r := p2(12); r.ig != 11 || r.sgg != 8 || r.lgg != 7 {
		t.Errorf("p2(12) ig/sgg/lgg = %d/%d/%d, want 11/8/7", r.ig, r.sgg, r.lgg)
	}
	// Clamping: out-of-range rolls saturate to rows 1 and 12.
	if p2(0) != p2Chart[1] || p2(99) != p2Chart[12] {
		t.Errorf("p2 clamp failed")
	}
	// ggOffset picks the class column.
	row := p2(1)
	if row.ggOffset(LargeGasGiant) != -4 || row.ggOffset(SmallGasGiant) != -3 || row.ggOffset(IceGiant) != 0 {
		t.Errorf("p2(1) ggOffset LGG/SGG/IG = %d/%d/%d, want -4/-3/0",
			row.ggOffset(LargeGasGiant), row.ggOffset(SmallGasGiant), row.ggOffset(IceGiant))
	}
}
