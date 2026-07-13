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
