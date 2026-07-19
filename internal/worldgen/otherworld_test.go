package worldgen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestSpaceport(t *testing.T) {
	cases := map[int]byte{5: 'F', 4: 'F', 3: 'G', 2: 'H', 1: 'H', 0: 'Y', -3: 'Y'}
	for popLess1D, want := range cases {
		if got := spaceport(popLess1D); got != want {
			t.Errorf("spaceport(%d) = %q, want %q", popLess1D, got, want)
		}
	}
}

// TestCapSizeTreatsEveryNegativeAsNoCap pins the sentinel's whole domain rather
// than its one canonical value. capSize tested `maxSize == NoSizeCap`, so every
// *other* negative fell past the guard into min(size, maxSize). Unreachable from
// inside the package — systemgen's satelliteBody and rollMoon both normalize —
// but GenerateSatelliteWorld is exported and takes maxSize straight from a
// caller.
//
// This is the shape of #200: a sentinel sharing a field with a domain, tested by
// equality when values outside the domain can reach past it. A sentinel meaning
// "no bound" has to be tested as a bound.
func TestCapSizeTreatsEveryNegativeAsNoCap(t *testing.T) {
	for _, maxSize := range []int{NoSizeCap, -2, -7, -100} {
		if got := capSize(9, maxSize); got != 9 {
			t.Errorf("capSize(9, %d) = %d, want 9; a negative cap is no cap", maxSize, got)
		}
	}

	// A real cap still caps, and 0 is still the caller's question to answer
	// (capSize's doc: whether a Size digit of 0 is a cap or the belt code is not
	// this function's decision).
	if got := capSize(9, 4); got != 4 {
		t.Errorf("capSize(9, 4) = %d, want 4", got)
	}

	if got := capSize(9, 0); got != 0 {
		t.Errorf("capSize(9, 0) = %d, want 0", got)
	}
}

// TestGenerateSatelliteWorldNegativeCapIsNotABelt is the same guarantee at the
// exported boundary, and it is where the hole actually bites. A negative Size
// never escapes — generateOtherWorld floors at minSize 0 — so nothing renders
// "?"; what a caller gets instead is Size 0, which is uwp.BeltSize. A stray
// negative cap silently turned a Big World into an asteroid belt, the same
// code-read-as-a-dimension confusion as #213 and #309 running the other way.
func TestGenerateSatelliteWorldNegativeCapIsNotABelt(t *testing.T) {
	fours := func() *dice.Roller { return dice.NewSource(func() int { return 4 }) }

	want := GenerateSatelliteWorld(fours(), BigWorld, 8, NoSizeCap)
	got := GenerateSatelliteWorld(fours(), BigWorld, 8, -3)

	if got.IsBelt() {
		t.Errorf("GenerateSatelliteWorld with maxSize -3 gave %s, an asteroid belt", got)
	}

	if got != want {
		t.Errorf("maxSize -3 gave %s, want %s (every negative cap is NoSizeCap)", got, want)
	}
}

func TestGenerateOtherWorldInferno(t *testing.T) {
	// YSB0000-0, Siz = 6+1D: 1D=3 -> size 9, exotic (B) atmosphere, no spaceport.
	p := GenerateOtherWorld(dice.NewScripted(3), Inferno, 8)
	if p.Starport != 'Y' || p.Size != 9 || p.Atmosphere != 11 {
		t.Errorf("Inferno = %s, want starport Y, size 9, atm 11", p)
	}

	if p.Population != 0 || p.Government != 0 || p.Law != 0 || p.TechLevel != 0 {
		t.Errorf("Inferno should be uninhabited: %s", p)
	}
}

func TestGenerateOtherWorldStormWorld(t *testing.T) {
	// Siz=2D(4), Atm=2D+4(8), Hyd=2D-4(6), Pop=2D-6(2) capped by mwPop-1=7,
	// Gov=Flux+Pop(2), Law=Flux+Gov(2), spaceport Pop-1D=2-1=H, TL=1D(3)+size4+pop2.
	r := dice.NewScripted(2, 2, 2, 2, 5, 5, 4, 4, 4, 4, 4, 4, 1, 3)

	p := GenerateOtherWorld(r, StormWorld, 8)
	if got := p.String(); got != "H486222-5" {
		t.Errorf("StormWorld = %q, want %q", got, "H486222-5")
	}
}

func TestGenerateOtherWorldRadWorld(t *testing.T) {
	// StSAH000-0: Siz=2D(5), Atm=Flux+Siz(5), Hyd=Flux+Atm(5); Pop/Gov/Law/TL 0;
	// Pop 0 forces a Y spaceport.
	r := dice.NewScripted(2, 3, 3, 3, 3, 3, 4)

	p := GenerateOtherWorld(r, RadWorld, 8)
	if p.Starport != 'Y' || p.Size != 5 || p.Atmosphere != 5 || p.Hydrographics != 5 {
		t.Errorf("RadWorld = %s, want Y555...", p)
	}

	if p.Population != 0 || p.TechLevel != 0 {
		t.Errorf("RadWorld should be uninhabited: %s", p)
	}
}

func TestGenerateOtherWorldPopCap(t *testing.T) {
	// Iceworld Pop = 2D-6 rolls 6, but mwPop=2 caps it at 1.
	r := dice.NewScripted(2, 2, 3, 3, 3, 3, 6, 6, 3, 3, 3, 3, 2, 2)

	p := GenerateOtherWorld(r, Iceworld, 2)
	if p.Population != 1 {
		t.Errorf("Iceworld pop = %d, want 1 (capped at mwPop-1)", p.Population)
	}
}
