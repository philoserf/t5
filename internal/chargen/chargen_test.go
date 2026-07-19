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
	// An omitted dice count is the standard 2D (Book 1 p. 47, "Check 2D =<
	// Characteristic"), not the 1D floor task.ResolveDice would otherwise apply.
	if two := c.Check(dice.NewScripted(3, 4), Endurance, 0, 0); len(two.Faces) != 2 {
		t.Fatalf("Check with numDice 0 rolled %dD, want 2D", len(two.Faces))
	}
}

// TestHardCheckIsSpectacularEligible guards the #290/#291 bundle. The p. 127
// Spectacular override lives in internal/task, so it reaches a Check
// Characteristic only because Character.Check routes through that package. A
// Hard Check is 3D and Book 1 p. 47 calls a 3D Check a "very hard task", so
// three ones must override a failing total here exactly as they would for any
// other task. If this fails, Check has been unrouted back onto the dice
// primitives.
func TestHardCheckIsSpectacularEligible(t *testing.T) {
	c := Generate(dice.NewScripted(1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1)) // all 2s

	// A Hard (3D) Check against Strength 2: 1+1+1 = 3 beats the target of 2, so
	// the arithmetic fails — but three ones are a Spectacular Success.
	res := c.Check(dice.NewScripted(1, 1, 1), Strength, task.Difficult.Dice(), 0)
	if res.Total <= res.Target {
		t.Fatalf("total %d vs target %d should fail arithmetically", res.Total, res.Target)
	}

	if !res.Success {
		t.Errorf("three ones must override a failing Hard Check (Book 1 pp. 47, 127)")
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
