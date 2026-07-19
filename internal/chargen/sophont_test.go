package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/sophont"
)

// humanSpecies is an all-2D, single-gender, casteless species — the sophont
// equivalent of a plain Human, used to lock the regression against Generate.
func humanSpecies() sophont.Species {
	var table [13]string
	for i := 2; i <= 12; i++ {
		table[i] = "Solo"
	}

	return sophont.Species{
		Chars: [6]sophont.CharSpec{
			{Name: sophont.Str, Dice: 2},
			{Name: sophont.Dex, Dice: 2},
			{Name: sophont.End, Dice: 2},
			{Name: sophont.Int, Dice: 2},
			{Name: sophont.Edu, Dice: 2},
			{Name: sophont.Soc, Dice: 2},
		},
		Gender: sophont.Gender{
			Structure: sophont.Solitaire,
			Genders:   []string{"Solo"},
			Table:     table,
		},
	}
}

// TestGenerateSophontCasteC6NoValue: for a Caste species, C6 is the caste — the
// slot has no characteristic value. GenerateSophont must not roll one (the die
// count is 0), so the sixth score stays at the 0 sentinel and the dice stream
// carries straight on to the gender and caste assignment rolls.
func TestGenerateSophontCasteC6NoValue(t *testing.T) {
	s := humanSpecies()
	s.Chars[5] = sophont.CharSpec{Name: sophont.Cas, Dice: 0}

	var casteTable [13]string
	for i := 2; i <= 12; i++ {
		casteTable[i] = "Muscle"
	}

	s.Caste = &sophont.Caste{Structure: sophont.Body, Table: casteTable}
	// Five 2D characteristic rolls (C6 rolls none), then the gender-assignment 2D
	// and the caste-assignment 2D. Scripted exactly: an extra C6 roll exhausts it.
	seq := []int{3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4}

	c := GenerateSophont(dice.NewScripted(seq...), s)
	if got := c.Score(Social); got != 0 {
		t.Errorf("C6 score = %d, want 0 — a Caste C6 holds no rolled value", got)
	}

	if c.Caste != "Muscle" {
		t.Errorf("Caste = %q, want Muscle", c.Caste)
	}
}

// TestGenerateSophontHumanRegression: an all-2D species rolled with the same
// dice as the human Generate yields the same UPP. The six characteristic rolls
// come first, so "777777" (each 2D = 3+4) matches, and the trailing 2D is the
// gender assignment (no effect for a single-gender species).
func TestGenerateSophontHumanRegression(t *testing.T) {
	seq := []int{3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4} // six 2D + gender-assign 2D
	c := GenerateSophont(dice.NewScripted(seq...), humanSpecies())

	if got := c.UPP(); got != "777777" {
		t.Errorf("UPP = %q, want 777777", got)
	}

	if c.Age != startingAge {
		t.Errorf("Age = %d, want %d", c.Age, startingAge)
	}

	if c.Gender != "Solo" {
		t.Errorf("Gender = %q, want Solo", c.Gender)
	}
}

// TestGenerateSophontBigStr checks the die-count scaling: a Str-6D species rolls
// 12 + 4D for Strength, so four 6s give Str 36 (eHex "Z0" not representable in a
// single digit — but the value is what matters), while a 2D characteristic caps
// far lower.
func TestGenerateSophontBigStr(t *testing.T) {
	sp := humanSpecies()
	sp.Chars[0].Dice = 6 // Strength rolled 6D -> 12 + 4D
	// Str: 6,6,6,6 -> 12+24 = 36. Remaining five chars: 2D each (3+4=7). Gender 2D.
	seq := []int{6, 6, 6, 6, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4}
	c := GenerateSophont(dice.NewScripted(seq...), sp)

	if got := c.Score(Strength); got != 36 {
		t.Errorf("Str = %d, want 36 (6D = 12+4D, all sixes)", got)
	}

	if got := c.Score(Endurance); got != 7 {
		t.Errorf("End = %d, want 7 (2D)", got)
	}
	// String/UPP is a renderer and must stay total: a Strength of 36 exceeds the
	// eHex range but must render (as "?") rather than panic.
	if got := c.String(); got != "?77777" {
		t.Errorf("String() = %q, want ?77777 (Str 36 out of eHex range)", got)
	}
}

// TestGenerateSophontGenderDifference confirms a non-base gender's difference is
// applied: a Male gender with a +2-across difference raises C1-C5 by 2.
func TestGenerateSophontGenderDifference(t *testing.T) {
	sp := humanSpecies()

	var table [13]string
	for i := 2; i <= 12; i++ {
		table[i] = "Male" // every 2D roll assigns Male
	}

	table[2] = "Female" // the base gender (no difference)
	sp.Gender = sophont.Gender{
		Structure: sophont.Dual,
		Genders:   []string{"Female", "Male"},
		Table:     table,
		Differences: map[string]sophont.Difference{
			"Male": {Mods: [5]int{2, 2, 2, 2, 2}},
		},
	}
	// Six 2D (all 7), then gender-assign 2D = 7 -> table[7] = "Male" -> +2 to C1-C5.
	seq := []int{3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4}
	c := GenerateSophont(dice.NewScripted(seq...), sp)

	if c.Gender != "Male" {
		t.Fatalf("Gender = %q, want Male", c.Gender)
	}

	if got := c.UPP(); got != "999997" { // C1-C5 raised 7->9, C6 (Soc) untouched
		t.Errorf("UPP = %q, want 999997 (Male +2 to C1-C5, Soc unchanged)", got)
	}
}

// TestGenerateSophontCasteDifference confirms a caste's difference is applied on
// top of the gender, including the Strength dice-add.
func TestGenerateSophontCasteDifference(t *testing.T) {
	sp := humanSpecies()
	sp.Chars[5].Name = sophont.Cas // C6 is Caste

	var ctable [13]string
	for i := 2; i <= 12; i++ {
		ctable[i] = "Warrior"
	}

	caste := sophont.Caste{
		Structure: sophont.Military,
		Table:     ctable,
		Differences: map[string]sophont.Difference{
			"Warrior": {C1Dice: 2, Mods: [5]int{0, 1, 1, 1, 1}}, // +2D Str, +1 C2-C5
		},
	}
	sp.Caste = &caste

	// Six 2D (all 7), gender-assign 2D (Solo, no diff), caste-assign 2D = 7 ->
	// Warrior -> +2D Str (2,2 = +4 -> 11) and +1 to C2-C5 (7->8).
	seq := []int{3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 2, 2}
	c := GenerateSophont(dice.NewScripted(seq...), sp)

	if c.Caste != "Warrior" {
		t.Fatalf("Caste = %q, want Warrior", c.Caste)
	}

	if got := c.Score(Strength); got != 11 {
		t.Errorf("Str = %d, want 11 (7 + 2D[2+2])", got)
	}

	if got := c.Score(Endurance); got != 8 {
		t.Errorf("End = %d, want 8 (7 + 1)", got)
	}
}
