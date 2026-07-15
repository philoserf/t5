package sophont

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

// fluxSeq builds a scripted die sequence producing the given Flux results in
// order (Flux = first die - second die).
func fluxSeq(vals ...int) []int {
	seq := make([]int, 0, 2*len(vals))
	for _, v := range vals {
		if v >= 0 {
			seq = append(seq, 1+v, 1)
		} else {
			seq = append(seq, 1, 1-v)
		}
	}
	return seq
}

// TestHumanCharacteristics is the reference golden (Book 3 p.228): with every
// Flux at 0 and no environmental modifier, a sophont comes out Human-standard —
// all six characteristics 2D with the Genetic Profile "SDEIES".
func TestHumanCharacteristics(t *testing.T) {
	env := Environment{Locomotion: Walker, Class: "Omnivore", Niche: "H/G"}
	// Nine Flux rolls, all 0: names C2/C3/C5/C6, then values C1/C2/C3/C4/C6
	// (C5 = Edu takes a flat 2D and rolls no value Flux).
	r := dice.NewScripted(fluxSeq(0, 0, 0, 0, 0, 0, 0, 0, 0)...)
	chars, gp := rollCharacteristics(r, env)

	if gp != "SDEIES" {
		t.Errorf("Genetic Profile = %q, want SDEIES", gp)
	}
	want := [6]CharSpec{{Str, 2}, {Dex, 2}, {End, 2}, {Int, 2}, {Edu, 2}, {Soc, 2}}
	if chars != want {
		t.Errorf("chars = %v, want %v", chars, want)
	}
}

// TestNameSlots exercises chart 06A: the variable name in each of slots C2/C3/C5/
// C6 across the Flux range, plus the Flyer locomotion DM shifting C2 to an analog.
func TestNameSlots(t *testing.T) {
	cases := []struct {
		slot, flux, mod int
		want            CharName
	}{
		{2, -5, 0, Agi}, {2, 0, 0, Dex}, {2, 5, 0, Gra}, {2, 0, -2, Agi}, // Flyer favors Agi
		{3, -3, 0, Sta}, {3, 0, 0, End}, {3, 3, 0, Vig},
		{5, -5, 0, Ins}, {5, 0, 0, Edu}, {5, 5, 0, Tra},
		{6, -5, 0, Cas}, {6, -2, 0, Soc}, {6, 0, 0, Soc}, {6, 5, 0, Cha},
	}
	for _, c := range cases {
		r := dice.NewScripted(fluxSeq(c.flux)...)
		if got := rollName(r, c.slot, c.mod); got != c.want {
			t.Errorf("slot %d flux %+d mod %+d = %v, want %v", c.slot, c.flux, c.mod, got, c.want)
		}
	}
}

// TestDieCountModifiers checks that the Environ DM scales Str's die count and
// the carnivore sub-niche shifts C3 (chart 06B).
func TestDieCountModifiers(t *testing.T) {
	// Environ DM +4 pushes Str (C1) from 2D to 6D at Flux 0.
	env := Environment{EnvironDM: 4}
	if got := rollDice(dice.NewScripted(fluxSeq(0)...), 0, Str, env); got != 6 {
		t.Errorf("Str dice at EnvironDM +4 = %d, want 6", got)
	}
	// A Chaser carnivore raises C3 from 2D to 3D; a Pouncer would drop it.
	chaser := Environment{Class: "Carnivore", Niche: "Chaser"}
	if got := rollDice(dice.NewScripted(fluxSeq(0)...), 2, End, chaser); got != 3 {
		t.Errorf("C3 dice for a Chaser = %d, want 3", got)
	}
	// A Pouncer drops a die at the boundary: at Flux -2 the End column is 2D,
	// but the Pouncer's -2 pushes the effective index to -4, giving 1D.
	pouncer := Environment{Class: "Carnivore", Niche: "Pouncer"}
	if got := rollDice(dice.NewScripted(fluxSeq(-2)...), 2, End, pouncer); got != 1 {
		t.Errorf("C3 dice for a Pouncer at Flux -2 = %d, want 1", got)
	}
}

// TestRollValueBounds locks the "Rolling Higher Characteristics" rule: 1-3 dice
// roll normally, 4+ dice roll 12 + (dice-2)D (Book 3 p.228 ranges).
func TestRollValueBounds(t *testing.T) {
	cases := []struct {
		name string
		seq  []int
		dice int
		want int
	}{
		{"2D normal", []int{4, 3}, 2, 7},
		{"6D min", []int{1, 1, 1, 1}, 6, 16},       // 12 + 4
		{"6D max", []int{6, 6, 6, 6}, 6, 36},       // 12 + 24
		{"8D max", []int{6, 6, 6, 6, 6, 6}, 8, 48}, // 12 + 36
	}
	for _, c := range cases {
		if got := RollValue(dice.NewScripted(c.seq...), c.dice); got != c.want {
			t.Errorf("%s: RollValue = %d, want %d", c.name, got, c.want)
		}
	}
}

// TestGeneticProfileLetters checks the analog GP letters, notably Caste = K.
func TestGeneticProfileLetters(t *testing.T) {
	cases := []struct {
		name CharName
		want byte
	}{{Str, 'S'}, {Agi, 'A'}, {Gra, 'G'}, {Vig, 'V'}, {Ins, 'I'}, {Cha, 'C'}, {Cas, 'K'}}
	for _, c := range cases {
		if got := c.name.gpLetter(); got != c.want {
			t.Errorf("%v GP letter = %c, want %c", c.name, got, c.want)
		}
	}
}
