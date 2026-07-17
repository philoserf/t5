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
	// Male gender's difference roll (1). Flux 0 structure column -> Dual.
	seq := fluxSeq(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
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

// TestGenderDifferences locks chart 08B, including the C1 dice-vs-flat split.
func TestGenderDifferences(t *testing.T) {
	cases := []struct {
		flux int
		want Difference
	}{
		{-3, Difference{Mods: [5]int{-3, -3, -3, -3, -3}}},
		{0, Difference{}},
		{1, Difference{Mods: [5]int{1, 0, 0, 0, 0}}},
		{2, Difference{Mods: [5]int{2, 2, 2, 2, 2}}},
		{3, Difference{C1Dice: 1, Mods: [5]int{0, 3, 3, 3, 3}}},
		{5, Difference{C1Dice: 3, Mods: [5]int{0, 5, 5, 5, 5}}},
	}
	for _, c := range cases {
		if got := genderDifference(c.flux); got != c.want {
			t.Errorf("genderDifference(%+d) = %+v, want %+v", c.flux, got, c.want)
		}
	}
}
