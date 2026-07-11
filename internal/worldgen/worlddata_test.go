package worldgen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/uwp"
)

func TestNobilityRegina(t *testing.T) {
	// Regina (subsector capital, Ix 4, TCs Ph/Pa/Ri): Knight, Baronet (Pa),
	// Baron (Ri), Viscount (Ph), Duke (capital) = BcCeF.
	if got := Nobility([]string{"Ph", "Pa", "Ri"}, 4, true); got != "BcCeF" {
		t.Fatalf("Nobility(Regina) = %q, want BcCeF", got)
	}
}

func TestNobility(t *testing.T) {
	// Bare world: only the always-present Knight.
	if got := Nobility(nil, 0, false); got != "B" {
		t.Errorf("bare = %q, want B", got)
	}
	// Important non-capital gets Duke f (lowercase), not F.
	if got := Nobility(nil, 4, false); got != "Bf" {
		t.Errorf("important non-capital = %q, want Bf", got)
	}
	// Capital takes F even when also important.
	if got := Nobility(nil, 5, true); got != "BF" {
		t.Errorf("important capital = %q, want BF", got)
	}
	// Count (E) from In or Hi; Marquis (D) from Pi.
	if got := Nobility([]string{"Hi", "Pi"}, 0, false); got != "BDE" {
		t.Errorf("Hi+Pi = %q, want BDE (Marquis, Count)", got)
	}
}

func TestRollBases(t *testing.T) {
	// Starport A: Naval on 2D<=6, Scout on 2D<=4. Roll 2D=5 then 2D=3.
	naval, scout := RollBases(dice.NewScripted(2, 3 /*=5*/, 1, 2 /*=3*/), 'A')
	if !naval || !scout {
		t.Errorf("A with low rolls: naval=%v scout=%v, want both true", naval, scout)
	}
	// Starport A: Naval on 2D=7 fails (>6); Scout on 2D=4 passes.
	naval, scout = RollBases(dice.NewScripted(3, 4 /*=7*/, 2, 2 /*=4*/), 'A')
	if naval || !scout {
		t.Errorf("A with 7/4: naval=%v scout=%v, want false/true", naval, scout)
	}
	// Starport E: neither base possible regardless of roll.
	naval, scout = RollBases(dice.NewScripted(1, 1), 'E')
	if naval || scout {
		t.Errorf("E: naval=%v scout=%v, want both false", naval, scout)
	}
	// Starport C: no Naval, but Scout on 2D<=6.
	naval, scout = RollBases(dice.NewScripted(3, 3 /*=6*/), 'C')
	if naval || !scout {
		t.Errorf("C: naval=%v scout=%v, want false/true", naval, scout)
	}
}

func TestTravelZone(t *testing.T) {
	// Regina Gov 9 + Law 9 = 18 -> Green.
	if got := TravelZone(regina); got != 'G' {
		t.Errorf("Regina zone = %c, want G", got)
	}
	// Gov+Law 20 -> Amber; 22 -> Red.
	if got := TravelZone(uwp.Profile{Starport: 'A', Government: 11, Law: 9}); got != 'A' {
		t.Errorf("Gov+Law 20 = %c, want A", got)
	}
	if got := TravelZone(uwp.Profile{Starport: 'A', Government: 11, Law: 11}); got != 'R' {
		t.Errorf("Gov+Law 22 = %c, want R", got)
	}
	// Class-X starport is Red regardless of Gov/Law.
	if got := TravelZone(uwp.Profile{Starport: 'X', Government: 0, Law: 0}); got != 'R' {
		t.Errorf("X starport = %c, want R", got)
	}
}

func TestNativeStatus(t *testing.T) {
	cases := []struct {
		name string
		p    uwp.Profile
		want string
	}{
		{"Regina natives", regina, "Natives"},
		{"corporate override", uwp.Profile{Government: 1, Population: 8}, "Corporate"},
		{"colonists override", uwp.Profile{Government: 6, Population: 8}, "Colonists"},
		{"exotic natives", uwp.Profile{Population: 8, Atmosphere: 11, TechLevel: 9}, "Exotic Natives"},
		{"transplants", uwp.Profile{Population: 8, Atmosphere: 0, TechLevel: 9}, "Transplants"},
		{"settlers", uwp.Profile{Population: 5, Atmosphere: 6, TechLevel: 7}, "Settlers"},
		{"transients", uwp.Profile{Population: 2, Atmosphere: 6, TechLevel: 5}, "Transients"},
		{"catastrophic XN", uwp.Profile{Population: 0, Atmosphere: 6, TechLevel: 5}, "Catastrophic XN"},
		{"extinct natives", uwp.Profile{Population: 0, Atmosphere: 6, TechLevel: 0}, "Extinct Natives"},
		{"catastrophic EXN", uwp.Profile{Population: 0, Atmosphere: 11, TechLevel: 3}, "Catastrophic EXN"},
		{"vanished transplants", uwp.Profile{Population: 0, Atmosphere: 1, TechLevel: 5}, "Vanished Transplants"},
		{"barren none", uwp.Profile{Population: 0, Atmosphere: 0, TechLevel: 0}, "None"},
		{"pop-0 ignores gov override", uwp.Profile{Population: 0, Government: 1, Atmosphere: 6, TechLevel: 5}, "Catastrophic XN"},
	}
	for _, c := range cases {
		if got := NativeStatus(c.p); got != c.want {
			t.Errorf("%s: NativeStatus = %q, want %q", c.name, got, c.want)
		}
	}
}
