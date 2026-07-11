package worldgen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/uwp"
)

// regina is the canonical worked example, UWP A788899-C.
var regina = uwp.Profile{Starport: 'A', Size: 7, Atmosphere: 8, Hydrographics: 8, Population: 8, Government: 9, Law: 9, TechLevel: 12}

func TestImportanceRegina(t *testing.T) {
	// Book 3's breakdown: Starport A +1, TL A +1, Rich +1, Pre-Ag +1 = +4.
	if got := Importance(regina, []string{"Ph", "Pa", "Ri"}, false, false, false); got != 4 {
		t.Fatalf("Importance(Regina) = %d, want 4", got)
	}
}

func TestImportanceModifiers(t *testing.T) {
	// Naval and Scout together add +1 (both required).
	if got := Importance(regina, []string{"Ph", "Pa", "Ri"}, true, true, false); got != 5 {
		t.Errorf("with both bases = %d, want 5", got)
	}
	if got := Importance(regina, []string{"Ph", "Pa", "Ri"}, true, false, false); got != 4 {
		t.Errorf("with only naval base = %d, want 4 (needs both)", got)
	}
	// A poor backwater: Starport X (-1), TL 3 (<=8, -1), no trade codes, Pop 2 (-1).
	poor := uwp.Profile{Starport: 'X', TechLevel: 3, Population: 2}
	if got := Importance(poor, nil, false, false, false); got != -3 {
		t.Errorf("poor world = %d, want -3", got)
	}
	// A high-TL hub: Starport A (+1), TL G=16 (>=G +1, >=A +1), In+Ri (+2), Pop 9.
	hub := uwp.Profile{Starport: 'A', TechLevel: 16, Population: 9}
	if got := Importance(hub, []string{"In", "Ri"}, false, false, true); got != 6 {
		// A+1, G+1, A+1, In+1, Ri+1, WayStation+1 = 6
		t.Errorf("hub = %d, want 6", got)
	}
}

func TestRollEconomicRegina(t *testing.T) {
	// Resources 2D=10 (+GG 3 +Belts 0 = 13); Infrastructure 2D=10 (+Ix 4 = 14);
	// Efficiency Flux = +4. Labor = Pop-1 = 7.
	r := dice.NewScripted(5, 5 /*R*/, 5, 5 /*I*/, 6, 2 /*E=+4*/)
	ex := RollEconomic(r, regina, 4, 3, 0)
	if ex != (Economic{Resources: 13, Labor: 7, Infrastructure: 14, Efficiency: 4}) {
		t.Fatalf("RollEconomic(Regina) = %+v", ex)
	}
	if got := ex.String(); got != "(D7E+4)" {
		t.Fatalf("Economic.String() = %q, want (D7E+4)", got)
	}
	if got := ex.RU(); got != 13*7*14*4 {
		t.Fatalf("RU() = %d, want %d", got, 13*7*14*4)
	}
}

func TestRollEconomicBands(t *testing.T) {
	// Low-TL world: gas giants/belts are NOT added to Resources.
	lowTL := uwp.Profile{Population: 5, TechLevel: 5}
	ex := RollEconomic(dice.NewScripted(3, 3 /*R=6*/, 4 /*I 1D=4*/, 3, 3 /*E=0*/), lowTL, 2, 4, 2)
	if ex.Resources != 6 {
		t.Errorf("low-TL Resources = %d, want 6 (no GG/belts)", ex.Resources)
	}
	if ex.Infrastructure != 6 { // Pop 4-6 -> 1D(4) + Ix(2)
		t.Errorf("Infrastructure = %d, want 6", ex.Infrastructure)
	}
	// Pop 0 -> Labor and Infrastructure floor at 0.
	ex0 := RollEconomic(dice.NewScripted(4, 4, 3, 3), uwp.Profile{Population: 0, TechLevel: 5}, 5, 0, 0)
	if ex0.Labor != 0 || ex0.Infrastructure != 0 {
		t.Errorf("pop-0 = %+v, want Labor 0 Infrastructure 0", ex0)
	}
}

func TestRollCulturalRegina(t *testing.T) {
	// H = Pop+Flux(+1) = 9; A = Pop+Ix = 12; S = Flux(+1)+5 = 6; Sy = Flux(+1)+TL = 13.
	r := dice.NewScripted(2, 1 /*H flux +1*/, 2, 1 /*S flux +1*/, 2, 1 /*Sy flux +1*/)
	cx := RollCultural(r, regina, 4)
	if cx != (Cultural{Heterogeneity: 9, Acceptance: 12, Strangeness: 6, Symbols: 13}) {
		t.Fatalf("RollCultural(Regina) = %+v", cx)
	}
	if got := cx.String(); got != "[9C6D]" {
		t.Fatalf("Cultural.String() = %q, want [9C6D]", got)
	}
}

func TestRollCulturalClampAndZero(t *testing.T) {
	// Values below 1 clamp to 1 (Strangeness = Flux(-5)+5 = 0 -> 1).
	cx := RollCultural(dice.NewScripted(1, 1 /*H flux 0*/, 1, 6 /*S flux -5*/, 1, 1 /*Sy flux 0*/),
		uwp.Profile{Population: 1, TechLevel: 1}, 0)
	if cx.Strangeness != 1 {
		t.Errorf("Strangeness = %d, want 1 (clamped)", cx.Strangeness)
	}
	// Pop 0 -> all zero.
	if got := RollCultural(dice.NewScripted(1, 1), uwp.Profile{Population: 0}, 5); got != (Cultural{}) {
		t.Fatalf("pop-0 Cultural = %+v, want all zero", got)
	}
}
