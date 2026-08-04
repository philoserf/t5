package task

import (
	"slices"
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestResolveUncertainHadwonUniverse2(t *testing.T) {
	// Book 1 p.127: the player sees 1,1,5,6 and assumes the hidden die is 3,
	// apparently failing 16 vs 13. The referee's hidden 1 makes an actual 14
	// with three ones: an invisible Spectacular Success.
	result := ResolveUncertainDice(dice.NewScripted(1, 1, 5, 6, 1), 5, 1, 13)
	if result.ApparentRoll != 16 || result.ApparentSuccess {
		t.Errorf("apparent = roll %d success %v, want 16 false", result.ApparentRoll, result.ApparentSuccess)
	}

	if result.Actual.Roll != 14 || !result.Actual.Success || result.Actual.Spectacular() != dice.SpectacularSuccess {
		t.Errorf("actual = %+v, want invisible Spectacular Success at 14", result.Actual)
	}

	if !slices.Equal(result.PlayerFaces, []int{1, 1, 5, 6}) || !slices.Equal(result.RefereeFaces, []int{1}) {
		t.Errorf("split faces = %v / %v", result.PlayerFaces, result.RefereeFaces)
	}
}

func TestResolveUncertainVisibleSpectacular(t *testing.T) {
	result := ResolveUncertainDice(dice.NewScripted(1, 1, 1, 6, 6), 5, 2, 2)
	if result.VisibleSpectacular != dice.SpectacularSuccess || !result.ApparentSuccess {
		t.Errorf("visible result = %v, success %v; want Spectacular Success",
			result.VisibleSpectacular, result.ApparentSuccess)
	}
}

func TestResolveUncertainKnownBounds(t *testing.T) {
	// Four visible dice total 6. Even a hidden 6 cannot exceed target 12.
	knownSuccess := ResolveUncertainDice(dice.NewScripted(1, 1, 2, 2, 6), 5, 1, 12)
	if !knownSuccess.ApparentSuccess || !knownSuccess.Actual.Success {
		t.Errorf("known success = %+v", knownSuccess)
	}

	// Four visible dice total 12. Even a hidden 1 exceeds target 12.
	knownFailure := ResolveUncertainDice(dice.NewScripted(3, 3, 3, 3, 1), 5, 1, 12)
	if knownFailure.ApparentSuccess || knownFailure.Actual.Success {
		t.Errorf("known failure = %+v", knownFailure)
	}
}

func TestUncertaintyPace(t *testing.T) {
	cases := []struct{ base, levels, taskDice, want int }{
		{1, 0, 4, 1}, // normal or Cautious: unchanged
		{1, 1, 3, 2}, // Hasty +1
		{1, 2, 4, 3}, // Extra Hasty +2
		{3, 2, 4, 4}, // cannot exceed the task dice
	}

	for _, tc := range cases {
		if got := UncertaintyDice(tc.base, tc.levels, tc.taskDice); got != tc.want {
			t.Errorf("UncertaintyDice(%d, %d, %d) = %d, want %d", tc.base, tc.levels, tc.taskDice, got, tc.want)
		}
	}
}

func TestResolveUncertainRejectsImpossibleSplit(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("uncertainty greater than task dice did not panic")
		}
	}()

	ResolveUncertainDice(dice.NewScripted(1), 2, 3, 7)
}
