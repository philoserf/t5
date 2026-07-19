package task

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

// TestSpectacularOverride locks the Book 1 p. 127 override, including its
// boundaries: exactly two ones or sixes is ordinary, and the both-at-once
// Spectacularly Interesting case leaves the arithmetic outcome standing.
//
// The rule lives here rather than in dice because p. 127 states it about tasks
// ("Sometimes the task result is Spectacular") and this package owns
// pp. 120-131. dice.Resolve reports arithmetic; these cases are what the task
// layer adds on top of it.
func TestSpectacularOverride(t *testing.T) {
	cases := []struct {
		name        string
		faces       []int
		target      int
		wantArith   bool // what Total <= Target says on its own
		wantSuccess bool // what Success must report after the override
	}{
		// Book 1 p. 127, Charles "Buzz" Van Sickle 596B77 Computer-3, a
		// Difficult (3D) computer search. Edu 7 + Computer-3 = target 10.
		{"buzz rolls 1-1-1", []int{1, 1, 1}, 10, true, true},
		{"buzz rolls 6-6-6", []int{6, 6, 6}, 10, false, false},
		// The override matters only when it disagrees with the arithmetic.
		{"three ones over target", []int{1, 1, 1, 6, 6}, 5, false, true},
		{"three sixes under target", []int{6, 6, 6, 1, 1}, 30, true, false},
		// Boundary: exactly two is not spectacular, so arithmetic rules.
		{"two ones failing", []int{1, 1, 6}, 5, false, false},
		{"two sixes passing", []int{6, 6, 1}, 30, true, true},
		// Six dice can show both. The book assigns no pass/fail, so the
		// arithmetic stands either way.
		{"interesting, arithmetic pass", []int{1, 1, 1, 6, 6, 6}, 30, true, true},
		{"interesting, arithmetic fail", []int{1, 1, 1, 6, 6, 6}, 5, false, false},
	}
	for _, c := range cases {
		res := ResolveDice(dice.NewScripted(c.faces...), len(c.faces), c.target)
		if arith := res.Total <= res.Target; arith != c.wantArith {
			t.Errorf("%s: arithmetic %d<=%d = %v, want %v",
				c.name, res.Total, res.Target, arith, c.wantArith)
		}

		if res.Success != c.wantSuccess {
			t.Errorf("%s: Success = %v, want %v", c.name, res.Success, c.wantSuccess)
		}
		// Effect stays arithmetic even when Success is overridden.
		if want := c.target - res.Total; res.Effect != want {
			t.Errorf("%s: Effect = %d, want arithmetic %d", c.name, res.Effect, want)
		}
	}
}

// TestSpectacularHadwonUniverse2 locks the p. 127 Gazelle example, the book's
// clearest statement that the override beats the arithmetic: Astrogator The
// Hadwon 7777A7 Astrogator-3 needs 13 or less on 5D (Staggering). He rolls
// 1,1,5,6 and the referee's Uncertain die is 1 — a total of 14, which "look[s]
// like a failure", but the three ones make it a Spectacular Success.
//
// Resolved through the Difficulty ladder the book names, so the golden exercises
// Staggering.Dice() == 5 rather than restating the count.
func TestSpectacularHadwonUniverse2(t *testing.T) {
	res := Resolve(dice.NewScripted(1, 1, 5, 6, 1), Staggering, 13)
	if res.Total != 14 {
		t.Fatalf("Total = %d, want 14", res.Total)
	}

	if res.Total <= res.Target {
		t.Fatalf("total 14 vs target 13 should fail arithmetically")
	}

	if got := res.Spectacular(); got != dice.SpectacularSuccess {
		t.Errorf("Spectacular() = %v, want SpectacularSuccess", got)
	}

	if !res.Success {
		t.Errorf("Success = false, want true: three ones override the failing total")
	}
}
