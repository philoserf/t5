package dice

import (
	"slices"
	"testing"
)

func TestDiceFaces(t *testing.T) {
	if got := scripted(1, 2, 3).DiceFaces(3); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("DiceFaces(3) = %v, want [1 2 3]", got)
	}

	if got := scripted(1).DiceFaces(0); got != nil {
		t.Errorf("DiceFaces(0) = %v, want nil", got)
	}
	// sum(DiceFaces(n)) == Dice(n) for the same scripted sequence.
	faces := scripted(2, 3, 4).DiceFaces(3)

	sum := 0
	for _, f := range faces {
		sum += f
	}

	if want := scripted(2, 3, 4).Dice(3); sum != want {
		t.Errorf("sum(DiceFaces) = %d, want Dice = %d", sum, want)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		name  string
		faces []int
		want  Spectacular
	}{
		{"three ones", []int{1, 1, 1}, SpectacularSuccess},
		{"three sixes", []int{6, 6, 6}, SpectacularFailure},
		{"four sixes", []int{6, 6, 6, 6, 2}, SpectacularFailure},
		{"both", []int{1, 1, 1, 6, 6, 6}, SpectacularlyInteresting},
		{"two ones", []int{1, 1, 4}, NotSpectacular},
		{"too few dice", []int{1, 1}, NotSpectacular},
		{"empty", nil, NotSpectacular},
	}
	for _, c := range cases {
		if got := Classify(c.faces); got != c.want {
			t.Errorf("%s: Classify(%v) = %v, want %v", c.name, c.faces, got, c.want)
		}
	}
}

func TestCheckResultSpectacular(t *testing.T) {
	// A 3D check rolling 1,1,1 is a Spectacular Success even though the
	// arithmetic (total 3 vs target 0) fails. Per Book 1 p. 127 the override
	// makes Success true; Total/Target still report the raw comparison.
	res := scripted(1, 1, 1).Resolve(Check{Dice: 3, Target: 0})
	if res.Total <= res.Target {
		t.Errorf("total %d vs target %d should fail arithmetically", res.Total, res.Target)
	}

	if !res.Success {
		t.Errorf("three ones must force Success (Book 1 p. 127)")
	}

	if got := res.Spectacular(); got != SpectacularSuccess {
		t.Errorf("Spectacular() = %v, want SpectacularSuccess", got)
	}
	// Three sixes: Spectacular Failure regardless of the (easy) target.
	res6 := scripted(6, 6, 6).Resolve(Check{Dice: 3, Target: 20})
	if res6.Total > res6.Target {
		t.Errorf("total %d vs target %d should pass arithmetically", res6.Total, res6.Target)
	}

	if res6.Success {
		t.Errorf("three sixes must force failure (Book 1 p. 127)")
	}

	if got := res6.Spectacular(); got != SpectacularFailure {
		t.Errorf("Spectacular() = %v, want SpectacularFailure", got)
	}
	// A 2D check can never be spectacular.
	if got := scripted(
		1,
		1,
	).Resolve(Check{Dice: 2, Target: 12}).
		Spectacular(); got != NotSpectacular {
		t.Errorf("2D Spectacular() = %v, want NotSpectacular", got)
	}
}

// TestSpectacularOverride locks the Book 1 p. 127 override, including its
// boundaries: exactly two ones or sixes is ordinary, and the both-at-once
// Spectacularly Interesting case leaves the arithmetic outcome standing.
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
		res := scripted(c.faces...).Resolve(Check{Dice: len(c.faces), Target: c.target})
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
func TestSpectacularHadwonUniverse2(t *testing.T) {
	res := scripted(1, 1, 5, 6, 1).Resolve(Check{Dice: 5, Target: 13})
	if res.Total != 14 {
		t.Fatalf("Total = %d, want 14", res.Total)
	}

	if res.Total <= res.Target {
		t.Fatalf("total 14 vs target 13 should fail arithmetically")
	}

	if got := res.Spectacular(); got != SpectacularSuccess {
		t.Errorf("Spectacular() = %v, want SpectacularSuccess", got)
	}

	if !res.Success {
		t.Errorf("Success = false, want true: three ones override the failing total")
	}
}
