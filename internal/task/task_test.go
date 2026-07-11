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
