package worldgen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestStarport(t *testing.T) {
	cases := map[int]byte{2: 'A', 4: 'A', 5: 'B', 6: 'B', 7: 'C', 8: 'C', 9: 'D', 10: 'E', 11: 'E', 12: 'X'}
	for roll, want := range cases {
		if got := starport(roll); got != want {
			t.Errorf("starport(%d) = %q, want %q", roll, got, want)
		}
	}
}

func TestAtmosphere(t *testing.T) {
	cases := []struct{ flux, size, want int }{
		{1, 7, 8},   // Regina
		{0, 0, 0},   // size 0 -> atmosphere 0
		{-5, 0, 0},  // size 0 dominates
		{-5, 3, 0},  // negative clamps to 0
		{5, 12, 15}, // clamps to F
	}
	for _, c := range cases {
		if got := atmosphere(c.flux, c.size); got != c.want {
			t.Errorf("atmosphere(flux=%d,size=%d) = %d, want %d", c.flux, c.size, got, c.want)
		}
	}
}

func TestHydrographics(t *testing.T) {
	cases := []struct{ flux, atm, size, want int }{
		{0, 8, 8, 8},  // Regina
		{0, 5, 1, 0},  // size < 2 -> 0
		{0, 1, 7, 0},  // thin atmosphere: -4 mod, clamps to 0
		{2, 1, 7, 0},  // flux 2 + atm 1 - 4 = -1 -> 0
		{0, 10, 7, 6}, // dense atmosphere (>9): 10 - 4 = 6
		{5, 9, 7, 10}, // clamps to A
	}
	for _, c := range cases {
		if got := hydrographics(c.flux, c.atm, c.size); got != c.want {
			t.Errorf("hydrographics(flux=%d,atm=%d,size=%d) = %d, want %d", c.flux, c.atm, c.size, got, c.want)
		}
	}
}

func TestGovernmentAndLaw(t *testing.T) {
	if got := government(1, 8); got != 9 { // Regina
		t.Errorf("government(1,8) = %d, want 9", got)
	}
	if got := government(-5, 2); got != 0 { // clamps to 0
		t.Errorf("government(-5,2) = %d, want 0", got)
	}
	if got := government(5, 14); got != 15 { // clamps to F
		t.Errorf("government(5,14) = %d, want 15", got)
	}
	if got := law(0, 9); got != 9 { // Regina
		t.Errorf("law(0,9) = %d, want 9", got)
	}
	if got := law(5, 15); got != 18 { // clamps to J
		t.Errorf("law(5,15) = %d, want 18", got)
	}
}

func TestTechLevel(t *testing.T) {
	// Regina: 1D=6, Starport A (+6), no other mods apply.
	if got := techLevel(6, 'A', 7, 8, 8, 8, 9); got != 12 {
		t.Errorf("techLevel Regina = %d, want 12", got)
	}
	// Every mod bucket at once: 1D=1, Starport A (+6), Size 1 (+2), Atm 2 (+1),
	// Hyd 10 (+2), Pop 3 (+1), Gov 5 (+1) = 14.
	if got := techLevel(1, 'A', 1, 2, 10, 3, 5); got != 14 {
		t.Errorf("techLevel stacked = %d, want 14", got)
	}
	// Negative mods can drive it toward zero: 1D=1, Starport X (-4), Gov D (-2).
	if got := techLevel(1, 'X', 7, 8, 8, 8, 13); got != 0 {
		t.Errorf("techLevel floored = %d, want 0", got)
	}
	// Each starport letter's TL modifier, with a neutral world (size 6, atm 6,
	// hyd 6, pop 6, gov 6 contribute nothing) so only the starport shows.
	starportTL := map[byte]int{'A': 6, 'B': 4, 'C': 2, 'D': 0, 'E': 0, 'X': -4}
	for sp, mod := range starportTL {
		want := max(3+mod, 0) // 1D=3 baseline
		if got := techLevel(3, sp, 6, 6, 6, 6, 6); got != want {
			t.Errorf("techLevel starport %q = %d, want %d", sp, got, want)
		}
	}
}

func TestRollSizeReroll(t *testing.T) {
	// 2D = 12 -> would be 10, so reroll as 1D+9. 1D = 3 -> size 12.
	if got := rollSize(dice.NewScripted(6, 6, 3)); got != 12 {
		t.Errorf("rollSize reroll = %d, want 12", got)
	}
	// 2D = 9 -> size 7, no reroll.
	if got := rollSize(dice.NewScripted(4, 5)); got != 7 {
		t.Errorf("rollSize = %d, want 7", got)
	}
}

func TestRollPopulationReroll(t *testing.T) {
	// 2D = 12 -> would be 10, so reroll as 2D+3. 2D = 4 -> population 7.
	if got := rollPopulation(dice.NewScripted(6, 6, 2, 2)); got != 7 {
		t.Errorf("rollPopulation reroll = %d, want 7", got)
	}
	// 2D = 10 -> population 8, no reroll.
	if got := rollPopulation(dice.NewScripted(5, 5)); got != 8 {
		t.Errorf("rollPopulation = %d, want 8", got)
	}
}

// TestGenerateRegina feeds the exact die rolls from Book 3's Regina worked
// example, in checklist order, and asserts the canonical UWP A788899-C.
func TestGenerateRegina(t *testing.T) {
	rolls := []int{
		2, 2, // Starport 2D = 4 -> A
		4, 5, // Size 2D = 9 -> 7
		2, 1, // Atmosphere Flux = +1 -> 8
		3, 3, // Hydrographics Flux = 0 -> 8
		5, 5, // Population 2D = 10 -> 8
		2, 1, // Government Flux = +1 -> 9
		3, 3, // Law Flux = 0 -> 9
		6, // Tech Level 1D = 6, +6 (Starport A) -> 12
	}
	if got := Generate(dice.NewScripted(rolls...)).String(); got != "A788899-C" {
		t.Fatalf("Generate (Regina) = %q, want A788899-C", got)
	}
}
