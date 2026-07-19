package sophont

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

// TestGenderAyFixture reproduces the "Ay" specimen (Book 3 p.219 Gender Example
// Rolls): a Dual species rolled at constant Flux 0. Entry 2 is Female, entry 3
// is Male, and every other entry is Female — giving a 34-female / 2-male census.
func TestGenderAyFixture(t *testing.T) {
	// Flux 0 for the structure, then Flux 0 for entries 4-12 (9 rolls), then the
	// Male gender's five difference rolls, C1..C5. Flux 0 structure -> Dual.
	seq := fluxSeq(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	g := rollGender(dice.NewScripted(seq...))

	if g.Structure != Dual {
		t.Fatalf("structure = %v, want Dual", g.Structure)
	}

	if g.Table[2] != "Female" || g.Table[3] != "Male" {
		t.Errorf("entries 2/3 = %q/%q, want Female/Male", g.Table[2], g.Table[3])
	}

	for entry := 4; entry <= 12; entry++ {
		if g.Table[entry] != "Female" {
			t.Errorf("entry %d = %q, want Female", entry, g.Table[entry])
		}
	}
	// Female is the base gender (no difference); Male rolled Flux 0 = no change.
	if _, ok := g.Differences["Female"]; ok {
		t.Error("base gender Female should carry no difference")
	}

	if d := g.Differences["Male"]; d != (Difference{}) {
		t.Errorf("Male difference at Flux 0 = %+v, want zero", d)
	}
}

// TestGroupEntryThreeIsRolled: chart 08A auto-fills entry 3 only for the
// three-name structures — "If Dual, FMN, or EAB, enter Gender 2 (Male,
// Activator) on entry line 3" (Book 3 p.230). A Group structure rolls entry 3
// like every other entry, which is what makes the book's own Group example
// reachable: the Rem "have a gender structure 36145 ... Note that Gender Two has
// evolutionarily dropped out" (p.219). Pinning "Two" at entry 3 would give it a
// floor of 2/36 and no Group species could ever lose it.
func TestGroupEntryThreeIsRolled(t *testing.T) {
	// Flux +4 for the structure -> Group; then Flux 0 for entries 3-12 (ten
	// rolls) -> "One"; then five difference rolls (C1..C5) for each of the five
	// non-base genders.
	seq := fluxSeq(append([]int{4}, make([]int, 35)...)...)
	g := rollGender(dice.NewScripted(seq...))

	if g.Structure != Group {
		t.Fatalf("structure = %v, want Group", g.Structure)
	}

	if g.Table[3] != "One" {
		t.Errorf("entry 3 = %q, want One (rolled, not auto-filled with Gender 2)", g.Table[3])
	}

	for entry := 2; entry <= 12; entry++ {
		if g.Table[entry] == "Two" {
			t.Errorf("entry %d = Two; an all-Flux-0 Group must be able to drop Gender Two", entry)
		}
	}
}

// TestGenderDifferenceIsUncorrelated is the same rule seen through rollGender:
// a Dual species whose Male difference draws five deliberately unequal Fluxes
// must record all five, not one row replicated. This is the shape a single
// per-gender roll can never produce.
func TestGenderDifferenceIsUncorrelated(t *testing.T) {
	// Flux 0 structure -> Dual; nine Flux-0 entry rolls; then Male's C1..C5.
	seq := fluxSeq(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, -4, 0, 2, 5)
	g := rollGender(dice.NewScripted(seq...))

	want := Difference{C1Dice: 1, Mods: [5]int{0, -4, 0, 2, 5}}
	if got := g.Differences["Male"]; got != want {
		t.Errorf("Male difference = %+v, want %+v", got, want)
	}
}

// TestGenderStructureSelection checks the Structure column of chart 08A.
func TestGenderStructureSelection(t *testing.T) {
	cases := []struct {
		flux int
		want GenderStructure
	}{{-5, Solitaire}, {-3, EAB}, {0, Dual}, {2, FMN}, {5, Group}}
	for _, c := range cases {
		if got := structureByFlux[clamp(c.flux, -5, 5)+5]; got != c.want {
			t.Errorf("structure at Flux %+d = %v, want %v", c.flux, got, c.want)
		}
	}
}

// TestGenderC1Column locks the C1 column of chart 08B, including the flat-to-
// dice transition at +3 ("+1D", "+2D", "+3D").
func TestGenderC1Column(t *testing.T) {
	cases := []struct{ flux, wantDice, wantMod int }{
		{-5, 0, -5}, {-3, 0, -3}, {-2, 0, -2},
		{-1, 0, 0}, {0, 0, 0},
		{1, 0, 1}, {2, 0, 2},
		{3, 1, 0}, {4, 2, 0}, {5, 3, 0},
	}
	for _, c := range cases {
		gotDice, gotMod := genderC1(c.flux)
		if gotDice != c.wantDice || gotMod != c.wantMod {
			t.Errorf("genderC1(%+d) = %dD/%+d, want %dD/%+d",
				c.flux, gotDice, gotMod, c.wantDice, c.wantMod)
		}
	}
}

// TestDifferenceModColumn locks the C2..C5 columns shared by charts 07C and 08B.
func TestDifferenceModColumn(t *testing.T) {
	cases := []struct{ flux, want int }{
		{-5, -5}, {-2, -2}, {-1, 0}, {0, 0}, {1, 0}, {2, 2}, {5, 5},
	}
	for _, c := range cases {
		if got := differenceMod(c.flux); got != c.want {
			t.Errorf("differenceMod(%+d) = %+d, want %+d", c.flux, got, c.want)
		}
	}
}

// TestGenderDifferencePerCharacteristic: chart 08B says "Roll once within each
// Gender for each Characteristic" (Book 3 p.230) — five independent Flux rolls,
// not one row applied across the board. A single roll would make every
// difference perfectly correlated, a distribution the rule cannot produce.
func TestGenderDifferencePerCharacteristic(t *testing.T) {
	r := dice.NewScripted(fluxSeq(3, -4, 0, 2, 5)...) // C1..C5, deliberately unequal
	got := rollGenderDifference(r)
	want := Difference{C1Dice: 1, Mods: [5]int{0, -4, 0, 2, 5}}

	if got != want {
		t.Errorf("rollGenderDifference = %+v, want %+v", got, want)
	}
}
