package rangeband

import (
	"math"
	"testing"
)

func TestWorldBand(t *testing.T) {
	b, ok := WorldBand("3")
	if !ok || b.Descriptor != "Medium" || b.Meters != 150 {
		t.Errorf("WorldBand(3) = %+v,%v, want Medium/150", b, ok)
	}

	if _, ok := WorldBand("Z"); ok {
		t.Errorf("WorldBand(Z) should not exist")
	}
}

func TestSpaceBand(t *testing.T) {
	if b, _ := SpaceBand("B"); b.Descriptor != "Boarding" || b.Combat != "B" || b.Meters != 1000 {
		t.Errorf("SpaceBand(B) = %+v, want Boarding/B/1000", b)
	}

	if b, _ := SpaceBand("13"); b.Descriptor != "Outer System" || b.Meters != 1.5e12 {
		t.Errorf("SpaceBand(13) = %+v, want Outer System/1.5e12", b)
	}
}

func TestForDistance(t *testing.T) {
	// Exact representative distances land on their band.
	if b := WorldForDistance(150); b.Code != "3" {
		t.Errorf("WorldForDistance(150m) = %q, want 3", b.Code)
	}

	if b := WorldForDistance(5_000); b.Code != "6" {
		t.Errorf("WorldForDistance(5km) = %q, want 6", b.Code)
	}

	if b := WorldForDistance(0); b.Code != "0" {
		t.Errorf("WorldForDistance(0) = %q, want 0 (Contact)", b.Code)
	}

	if b := SpaceForDistance(5_000_000); b.Code != "4" {
		t.Errorf("SpaceForDistance(5000km) = %q, want 4 (Far Orbit)", b.Code)
	}

	if b := SpaceForDistance(1.5e12); b.Code != "13" {
		t.Errorf("SpaceForDistance(1.5bn km) = %q, want 13", b.Code)
	}
}

func TestBandNumber(t *testing.T) {
	b3, _ := WorldBand("3")
	if n, ok := b3.Number(); n != 3 || !ok {
		t.Errorf("WorldBand(3).Number() = %d,%v, want 3,true", n, ok)
	}

	s13, _ := SpaceBand("13")
	if n, ok := s13.Number(); n != 13 || !ok {
		t.Errorf("SpaceBand(13).Number() = %d,%v, want 13,true", n, ok)
	}
	// The lettered Contact/Boarding sub-bands have no number.
	for _, b := range []Band{{Code: "R"}, {Code: "T"}, {Code: "B"}} {
		if _, ok := b.Number(); ok {
			t.Errorf("Band(%q).Number() should report false", b.Code)
		}
	}
}

func TestConversion(t *testing.T) {
	world := map[string]string{"3": "0", "4": "0", "5": "B", "6": "1", "9": "4"}
	for r, wantS := range world {
		if s, ok := WorldToSpace(r); !ok || s != wantS {
			t.Errorf("WorldToSpace(%q) = %q,%v, want %q", r, s, ok, wantS)
		}
	}

	space := map[string]string{"0": "0", "B": "5", "4": "9", "13": "18"}
	for s, wantR := range space {
		if r, ok := SpaceToWorld(s); !ok || r != wantR {
			t.Errorf("SpaceToWorld(%q) = %q,%v, want %q", s, r, ok, wantR)
		}
	}
}

// TestConversionRoundTrip pins that the two conversions agree over their shared
// domain (Book 1 p.29: R= runs to 18, so S=5..13 map to the extended R=10..18
// that lie beyond the World scale's named bands but on the same ladder).
func TestConversionRoundTrip(t *testing.T) {
	// Every S code survives S -> R -> S.
	for _, b := range spaceBands {
		r, ok := SpaceToWorld(b.Code)
		if !ok {
			t.Errorf("SpaceToWorld(%q) not ok", b.Code)

			continue
		}

		back, ok := WorldToSpace(r)
		if !ok {
			t.Errorf("WorldToSpace(%q) = _,false; SpaceToWorld(%q) produced it", r, b.Code)

			continue
		}

		if back != b.Code {
			t.Errorf("S=%q -> R=%q -> S=%q, want %q", b.Code, r, back, b.Code)
		}
	}
	// Named R codes convert; R<=4 all collapse to S=0, so only R>=5 round-trips.
	for _, b := range worldBands {
		s, ok := WorldToSpace(b.Code)
		if !ok {
			t.Errorf("WorldToSpace(%q) not ok", b.Code)

			continue
		}

		if n, isNum := b.Number(); !isNum || n < 5 {
			if s != "0" {
				t.Errorf("WorldToSpace(%q) = %q, want 0 (R<=4 collapses)", b.Code, s)
			}

			continue
		}

		if back, _ := SpaceToWorld(s); back != b.Code {
			t.Errorf("R=%q -> S=%q -> R=%q, want %q", b.Code, s, back, b.Code)
		}
	}
	// The extended codes the chart prints, spot-checked against p.29.
	for r, wantS := range map[string]string{"10": "5", "15": "10", "18": "13"} {
		if s, ok := WorldToSpace(r); !ok || s != wantS {
			t.Errorf("WorldToSpace(%q) = %q,%v, want %q", r, s, ok, wantS)
		}
	}
	// Beyond the chart's R= column there is no band.
	for _, r := range []string{"19", "25", "-1"} {
		if _, ok := WorldToSpace(r); ok {
			t.Errorf("WorldToSpace(%q) should not convert", r)
		}
	}
}

// TestSpaceCombatBands pins the Band column of the p.29 Space Ranges chart, the
// last untested column. Each combat letter spans two S= rows — SR on 4-5, AR on
// 6-7, LR on 8-9, DS on 11-12 — which is how the chart stacks the two-word band
// names ("Short"/"Range") across the pair. The empties are pinned too: S=0, 3,
// 10, and 13 carry no combat letter.
func TestSpaceCombatBands(t *testing.T) {
	want := map[string]string{
		"0": "", "B": "B", "1": "F1", "2": "F2",
		"3": "", "4": "SR", "5": "SR", "6": "AR",
		"7": "AR", "8": "LR", "9": "LR", "10": "",
		"11": "DS", "12": "DS", "13": "",
	}
	if len(want) != len(spaceBands) {
		t.Fatalf("pinned %d rows, spaceBands has %d", len(want), len(spaceBands))
	}

	for code, combat := range want {
		b, ok := SpaceBand(code)
		if !ok {
			t.Errorf("SpaceBand(%q) not found", code)

			continue
		}

		if b.Combat != combat {
			t.Errorf("SpaceBand(%q).Combat = %q, want %q", code, b.Combat, combat)
		}
	}
}

// TestWorldForDistanceWidths pins WorldForDistance against the Range Band Width
// column printed on Book 1 p.24, band by band: each band's lower edge, an
// interior point, and its printed upper bound (which is inclusive here — the
// book prints shared edges, e.g. Vshort "3 m to 25 m" and Short "25 m to 100 m").
func TestWorldForDistanceWidths(t *testing.T) {
	cases := []struct {
		meters float64
		want   string
	}{
		{0, "0"}, {0.1, "0"}, {0.25, "0"}, // Contact: zero point to 25 cm
		{0.3, "R"}, {0.5, "R"}, {1, "R"}, // Reading: 25 cm to 1 m
		{1.1, "T"}, {1.5, "T"}, {3, "T"}, // Talking: 1 m to about 3 m
		{4, "1"}, {20, "1"}, {25, "1"}, // Vshort: 3 m to 25 m
		{30, "2"}, {90, "2"}, {100, "2"}, // Short: 25 m to 100 m
		{150, "3"}, {300, "3"}, // Medium: 100 m to 300 m
		{500, "4"}, {750, "4"}, // Long: 300 m to 750 m
		{1_000, "5"}, {2_500, "5"}, {3_000, "5"}, // Vlong: 750 m to 3 km
		{5_000, "6"}, {25_000, "6"}, // Distant: 3 km to 25 km
		{50_000, "7"}, {250_000, "7"}, // Vdistant: 25 km to 250 km
		{500_000, "8"}, {2_500_000, "8"}, // Orbit: 250 km to 2,500 km
		{5_000_000, "9"}, {25_000_000, "9"}, // Far Orbit: 2,500 km to 25,000 km
		{1e12, "9"}, // past the top of the ladder, Far Orbit still
	}
	for _, c := range cases {
		if got := WorldForDistance(c.meters); got.Code != c.want {
			t.Errorf("WorldForDistance(%gm) = %q (%s), want %q",
				c.meters, got.Code, got.Descriptor, c.want)
		}
	}
}

func TestWorldSubBand(t *testing.T) {
	// A representative distance lands exactly on its band.
	if got := WorldSubBand(150); got != 3 {
		t.Errorf("WorldSubBand(150m) = %g, want 3", got)
	}
	// Below R=1 clamps to 1, at/above R=9 clamps to 9.
	if got := WorldSubBand(1); got != 1 {
		t.Errorf("WorldSubBand(1m) = %g, want 1", got)
	}

	if got := WorldSubBand(1e9); got != 9 {
		t.Errorf("WorldSubBand(1e9) = %g, want 9", got)
	}
	// 8 km sits just into band 6 (the book's gas-giant 6.x sub-bands).
	if got := WorldSubBand(8_000); math.Abs(got-6.2) > 0.05 {
		t.Errorf("WorldSubBand(8km) = %g, want ~6.2", got)
	}
}

// TestSpaceDescriptors pins the descriptor of every space band against Book 1
// p.29 — the column that was untested, which let "Missile" sit on S=6 (it belongs
// on S=7; S=6 is the Attack Range band). A blank-descriptor row takes its band's
// name (S=5 Short Range, S=8/9 Long Range, S=11/12 Deep Space); a row with its own
// descriptor keeps it (S=7 Missile, S=10 Siege).
func TestSpaceDescriptors(t *testing.T) {
	want := map[string]string{
		"0": "Contact", "B": "Boarding", "1": "Close Fighter", "2": "Fighter",
		"3": "Orbit", "4": "Far Orbit", "5": "Short Range", "6": "Attack Range",
		"7": "Missile", "8": "Long Range", "9": "Long Range", "10": "Siege",
		"11": "Deep Space", "12": "Deep Space", "13": "Outer System",
	}
	for code, desc := range want {
		b, ok := SpaceBand(code)
		if !ok {
			t.Errorf("SpaceBand(%q) not found", code)

			continue
		}

		if b.Descriptor != desc {
			t.Errorf("SpaceBand(%q).Descriptor = %q, want %q", code, b.Descriptor, desc)
		}
	}
}
