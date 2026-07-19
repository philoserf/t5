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

// TestCheckResultSpectacular checks that a CheckResult reports its dice's
// classification. Resolve classifies but does not act: the p. 127 override is
// the task layer's (see internal/task), so Success here stays arithmetic even
// when the faces are spectacular.
func TestCheckResultSpectacular(t *testing.T) {
	// A 3D check rolling 1,1,1 classifies as a Spectacular Success, but the
	// arithmetic (total 3 vs target 0) fails and dice.Resolve leaves it failing.
	res := scripted(1, 1, 1).Resolve(Check{Dice: 3, Target: 0})
	if res.Total <= res.Target {
		t.Errorf("total %d vs target %d should fail arithmetically", res.Total, res.Target)
	}

	if res.Success {
		t.Errorf("dice.Resolve must report arithmetic; the p. 127 override is task's")
	}

	if got := res.Spectacular(); got != SpectacularSuccess {
		t.Errorf("Spectacular() = %v, want SpectacularSuccess", got)
	}
	// Three sixes: classified a Spectacular Failure, but the target is easy and
	// the arithmetic pass stands at this tier.
	res6 := scripted(6, 6, 6).Resolve(Check{Dice: 3, Target: 20})
	if res6.Total > res6.Target {
		t.Errorf("total %d vs target %d should pass arithmetically", res6.Total, res6.Target)
	}

	if !res6.Success {
		t.Errorf("dice.Resolve must report arithmetic; the p. 127 override is task's")
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
