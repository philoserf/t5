package task

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestDifficultyDice(t *testing.T) {
	want := map[Difficulty]int{
		Easy: 1, Average: 2, Difficult: 3, Formidable: 4,
		Staggering: 5, Hopeless: 6, Impossible: 7, BeyondImpossible: 8,
	}
	for d, n := range want {
		if got := d.Dice(); got != n {
			t.Errorf("%s.Dice() = %d, want %d", d, got, n)
		}
	}
}

// Issue #242: a Difficulty outside the ladder has no dice count. Dice must not
// invent one — a non-positive count would fall through to dice.Resolve's 2D
// default and silently become an ordinary Average check.
func TestDiceOffLadderPanics(t *testing.T) {
	for _, d := range []Difficulty{-1, BeyondImpossible + 1, 99} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Difficulty(%d).Dice() = %d, want panic", int(d), d.Dice())
				}
			}()

			_ = d.Dice()
		}()
	}
}

// Hasty and Cautious are built on Dice, so they inherit the domain check.
func TestPaceOffLadderPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Difficulty(99).Hasty() did not panic")
		}
	}()

	_ = Difficulty(99).Hasty()
}

// Issue #261: task.Difficulty is the only difficulty vocabulary. The dice
// package must speak dice counts alone — its former Easy/Average/Hard constants
// held counts (1/2/3) under names this ladder uses for indices (0/1/2), so
// dice.Check{Dice: int(task.Average)} silently rolled 1D. dice keeps exactly one
// named count, the default Resolve falls back to, and it must agree with the
// ladder's Average.
func TestDiceVocabularyDoesNotCollide(t *testing.T) {
	if dice.DefaultCheckDice != Average.Dice() {
		t.Errorf("dice.DefaultCheckDice = %d, want Average.Dice() = %d",
			dice.DefaultCheckDice, Average.Dice())
	}
}

func TestPace(t *testing.T) {
	// Hasty is +1D; Cautious is -1D, floored at 1D (Book 1 p. 129 columns).
	cases := []struct {
		d               Difficulty
		hasty, cautious int
	}{
		{Easy, 2, 1},      // cautious cannot drop below 1D
		{Average, 3, 1},   // 2D -> cautious 1D
		{Difficult, 4, 2}, // 3D -> cautious 2D
		{BeyondImpossible, 9, 7},
	}
	for _, c := range cases {
		if got := c.d.Hasty(); got != c.hasty {
			t.Errorf("%s.Hasty() = %d, want %d", c.d, got, c.hasty)
		}

		if got := c.d.Cautious(); got != c.cautious {
			t.Errorf("%s.Cautious() = %d, want %d", c.d, got, c.cautious)
		}

		if got := c.d.ExtraHasty(); got != c.d.Dice()+2 {
			t.Errorf("%s.ExtraHasty() = %d, want %d", c.d, got, c.d.Dice()+2)
		}
	}
}

func TestString(t *testing.T) {
	if got := Formidable.String(); got != "Formidable" {
		t.Errorf("Formidable.String() = %q", got)
	}

	if got := BeyondImpossible.String(); got != "Beyond Impossible" {
		t.Errorf("BeyondImpossible.String() = %q", got)
	}

	if got := Difficulty(99).String(); got != "?" {
		t.Errorf("out-of-range = %q, want ?", got)
	}
}

func TestResolve(t *testing.T) {
	// Book 1 example: an Average (2D) task against Edu+Comm = 7+2 = 9. A roll of
	// 8 succeeds (8 <= 9).
	res := Resolve(dice.NewScripted(3, 5), Average, 7, 2)
	if !res.Success || res.Target != 9 || res.Roll != 8 {
		t.Fatalf("Resolve = %+v, want success roll 8 target 9", res)
	}
	// A Difficult (3D) task against the same 9, rolling 12, fails.
	fail := Resolve(dice.NewScripted(4, 4, 4), Difficult, 9)
	if fail.Success || fail.Roll != 12 {
		t.Fatalf("Resolve = %+v, want failure roll 12", fail)
	}
}

func TestResolveDice(t *testing.T) {
	// A cautious Average task uses 1D (see Average.Cautious()).
	res := ResolveDice(dice.NewScripted(6), Average.Cautious(), 5)
	if res.Roll != 6 || res.Success {
		t.Fatalf("cautious = %+v, want roll 6 failure (6 > 5)", res)
	}
	// A count below 1 is treated as 1D, not the 2D roll-low default.
	if got := ResolveDice(dice.NewScripted(4), 0, 5); got.Roll != 4 {
		t.Fatalf("ResolveDice(0) rolled %d dice worth, want 1D (roll 4)", got.Roll)
	}
}
