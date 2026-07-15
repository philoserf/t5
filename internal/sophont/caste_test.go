package sophont

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

// TestCasteAyFixture reproduces the "Ay" specimen (Book 3 p.218 Caste Example
// Rolls): a Body caste rolled at constant Flux 0. Every entry is the Common caste
// Muscle, except the auto-inserted Unique caste Brain at entry 12.
func TestCasteAyFixture(t *testing.T) {
	// 1D structure = 1 (Body); nine Flux-0 entry rolls; one Flux-0 difference roll
	// for the single non-Common caste (Brain).
	seq := append([]int{1}, fluxSeq(0, 0, 0, 0, 0, 0, 0, 0, 0, 0)...)
	c := rollCaste(dice.NewScripted(seq...), Gender{})

	if c.Structure != Body {
		t.Fatalf("structure = %v, want Body", c.Structure)
	}
	for entry := 2; entry <= 11; entry++ {
		if c.Table[entry] != "Muscle" {
			t.Errorf("entry %d = %q, want Muscle", entry, c.Table[entry])
		}
	}
	if c.Table[12] != "Brain" {
		t.Errorf("entry 12 = %q, want Brain (Unique)", c.Table[12])
	}
	// Muscle is Common (no difference); Brain rolled Flux 0 = no change.
	if _, ok := c.Differences["Muscle"]; ok {
		t.Error("Common caste Muscle should carry no difference")
	}
	if d := c.Differences["Brain"]; d != (Difference{}) {
		t.Errorf("Brain difference at Flux 0 = %+v, want zero", d)
	}
}

// TestCasteGenderSubstitution checks that a =Gender cell (Body structure, Flux
// -4) copies the parallel gender at the same 2D slot.
func TestCasteGenderSubstitution(t *testing.T) {
	gender := Gender{}
	gender.Table[4] = "Bearer" // the gender occupying slot 4
	// Structure die 1 (Body). Entries 2 and 3 roll Flux 0 (Muscle); entry 4 rolls
	// Flux -4 (=Gender) -> copies gender.Table[4]; entries 5,6,8..11 roll Flux 0.
	seq := []int{1}
	seq = append(seq, fluxSeq(0, 0)...)       // entries 2, 3
	seq = append(seq, fluxSeq(-4)...)         // entry 4 -> =Gender
	seq = append(seq, fluxSeq(0, 0)...)       // entries 5, 6
	seq = append(seq, fluxSeq(0, 0, 0, 0)...) // entries 8, 9, 10, 11
	// Distinct non-Common castes are Bearer then (nothing else); Brain at 12. Two
	// difference rolls (Bearer, Brain).
	seq = append(seq, fluxSeq(0, 0)...)
	c := rollCaste(dice.NewScripted(seq...), gender)
	if c.Table[4] != "Bearer" {
		t.Errorf("entry 4 = %q, want Bearer (=Gender substitution)", c.Table[4])
	}
}

// TestSkilledCasteDeferred confirms a Skilled structure records only its
// structure (per-member skills are deferred to Chart 12).
func TestSkilledCasteDeferred(t *testing.T) {
	c := rollCaste(dice.NewScripted(6), Gender{}) // 1D = 6 -> Skilled
	if c.Structure != Skilled {
		t.Fatalf("structure = %v, want Skilled", c.Structure)
	}
	if c.Table != ([13]string{}) || len(c.Differences) != 0 {
		t.Error("Skilled caste should leave Table and Differences empty")
	}
}

// TestCasteDifferences locks chart 07C.
func TestCasteDifferences(t *testing.T) {
	cases := []struct {
		flux int
		want Difference
	}{
		{-4, Difference{Mods: [5]int{-4, -4, -4, -4, -4}}},
		{-1, Difference{}},
		{1, Difference{}},
		{2, Difference{C1Dice: 2, Mods: [5]int{0, 2, 2, 2, 2}}},
		{5, Difference{C1Dice: 5, Mods: [5]int{0, 5, 5, 5, 5}}},
	}
	for _, c := range cases {
		if got := casteDifference(c.flux); got != c.want {
			t.Errorf("casteDifference(%+d) = %+v, want %+v", c.flux, got, c.want)
		}
	}
}
