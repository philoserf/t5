package chargen

import (
	"slices"
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/task"
)

func TestGenerateUPP(t *testing.T) {
	// Six 2D rolls in order: 7,7,7,8,9,10 -> eHex "77789A".
	rolls := []int{
		3, 4, // Str 7
		4, 3, // Dex 7
		5, 2, // End 7
		4, 4, // Int 8
		4, 5, // Edu 9
		5, 5, // Soc 10 (A)
	}

	c := Generate(dice.NewScripted(rolls...))
	if got := c.UPP(); got != "77789A" {
		t.Fatalf("UPP() = %q, want 77789A", got)
	}

	if c.Score(Endurance) != 7 || c.Score(Social) != 10 {
		t.Fatalf("Score mismatch: End=%d Soc=%d", c.Score(Endurance), c.Score(Social))
	}
}

func TestCharacteristicString(t *testing.T) {
	want := []string{"Str", "Dex", "End", "Int", "Edu", "Soc"}
	for i, w := range want {
		if got := Characteristic(i).String(); got != w {
			t.Errorf("Characteristic(%d) = %q, want %q", i, got, w)
		}
	}

	if got := Characteristic(99).String(); got != "?" {
		t.Errorf("out-of-range = %q, want ?", got)
	}
}

func TestScorePanicsOutOfRange(t *testing.T) {
	c := Generate(dice.NewScripted(slices.Repeat([]int{4}, 12)...)) // six 2D characteristics

	for _, ch := range []Characteristic{-1, count, 99} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Score(%d) did not panic", ch)
				}
			}()

			c.Score(ch)
		}()
	}
}

func TestCheck(t *testing.T) {
	// Endurance 8; a 2D roll of 7 succeeds (7 <= 8).
	c := Generate(dice.NewScripted(4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4)) // all 8s

	res := c.Check(dice.NewScripted(3, 4), Endurance, task.Average.Dice(), 0)
	if !res.Success || res.Target != 8 {
		t.Fatalf("Check = %+v, want success against target 8", res)
	}
	// A +2 mod raises the target; an Easy (1D) check uses one die.
	easy := c.Check(dice.NewScripted(6), Strength, task.Easy.Dice(), 2)
	if !easy.Success || easy.Target != 10 {
		t.Fatalf("Easy check = %+v, want success against target 10", easy)
	}
}

func TestGenerateRangeAndDeterminism(t *testing.T) {
	for seed := uint64(1); seed <= 300; seed++ {
		a := Generate(dice.NewWithSeed(seed))

		b := Generate(dice.NewWithSeed(seed))
		if a.UPP() != b.UPP() {
			t.Fatalf("seed %d not reproducible: %s vs %s", seed, a, b)
		}

		for ch := Strength; ch <= Social; ch++ {
			if v := a.Score(ch); v < 2 || v > 12 {
				t.Fatalf("seed %d: %s = %d out of 2D range", seed, ch, v)
			}
		}
	}
}
