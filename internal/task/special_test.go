package task

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestCooperativeTargets(t *testing.T) {
	if got := CooperativeTargetSkill(9, 3, 2, 1); got != 15 {
		t.Errorf("skill cooperation target = %d, want 15", got)
	}

	if got := CooperativeTargetCharacteristic(0, 7, 8, 9); got != 24 {
		t.Errorf("characteristic cooperation target = %d, want 24", got)
	}
}

func checkResult(roll int, success bool) dice.CheckResult {
	return dice.CheckResult{Roll: roll, Success: success}
}

func TestOpposedWinner(t *testing.T) {
	cases := []struct {
		name    string
		results []dice.CheckResult
		winner  int
		ok      bool
	}{
		{
			"lowest successful",
			[]dice.CheckResult{checkResult(9, true), checkResult(7, true), checkResult(12, false)},
			1,
			true,
		},
		{"failed low roll ignored", []dice.CheckResult{checkResult(4, false), checkResult(8, true)}, 1, true},
		{"tie has no winner", []dice.CheckResult{checkResult(7, true), checkResult(7, true)}, 0, false},
		{"no success", []dice.CheckResult{checkResult(9, false), checkResult(12, false)}, 0, false},
	}

	for _, tc := range cases {
		got, ok := OpposedWinner(tc.results)
		if ok != tc.ok || (ok && got != tc.winner) {
			t.Errorf("%s: OpposedWinner = %d, %v; want %d, %v", tc.name, got, ok, tc.winner, tc.ok)
		}
	}
}

func TestOpposedLoser(t *testing.T) {
	if got, ok := OpposedLoser([]dice.CheckResult{
		checkResult(8, true), checkResult(13, false), checkResult(10, false),
	}); !ok || got != 1 {
		t.Errorf("OpposedLoser = %d, %v; want 1, true", got, ok)
	}

	if _, ok := OpposedLoser([]dice.CheckResult{checkResult(8, true), checkResult(7, true)}); ok {
		t.Error("all-success round selected a loser")
	}
}

func TestArcaneOwnership(t *testing.T) {
	r := dice.NewScripted(3, 4, 5)
	if result, ok := ResolveArcane(r, Difficult, 12, false); ok || result.Roll != 0 || result.Faces != nil {
		t.Errorf("non-owner = %+v, %v; want zero, false", result, ok)
	}

	result, ok := ResolveArcane(r, Difficult, 12, true)
	if !ok || !result.Success || result.Roll != 12 {
		t.Errorf("owner = %+v, %v; want successful roll 12", result, ok)
	}
}

func TestThisIsHard(t *testing.T) {
	cases := []struct {
		dice, skill, jot, want int
	}{
		{3, 2, 0, 4}, // p.123: Skill-2 Difficult becomes Formidable.
		{3, 2, 1, 3}, // JOT shields but is not an asset.
		{2, 0, 0, 3}, // default Skill-0 Average becomes Difficult.
		{1, 1, 0, 1},
		{0, 0, 0, 2}, // invalid count floors to 1D, then TIH applies.
	}

	for _, tc := range cases {
		if got := ThisIsHardDice(tc.dice, tc.skill, tc.jot); got != tc.want {
			t.Errorf("ThisIsHardDice(%d, %d, %d) = %d, want %d", tc.dice, tc.skill, tc.jot, got, tc.want)
		}
	}
}

func TestPhantomAssetsAndCertification(t *testing.T) {
	if PhantomSkill != 3 || PhantomCharacteristic != 7 {
		t.Errorf("phantom assets = C%d/S%d, want C7/S3", PhantomCharacteristic, PhantomSkill)
	}

	if rank, ok := CertificationRank(checkResult(4, true)); !ok || rank != 4 {
		t.Errorf("successful rank = %d, %v; want 4, true", rank, ok)
	}

	if _, ok := CertificationRank(checkResult(18, false)); ok {
		t.Error("failed test conferred certification")
	}
}

func TestSubstituteAssets(t *testing.T) {
	for level, want := range map[int]int{0: 0, 1: 0, 2: 0, 3: 1, 6: 4} {
		if got := JackOfAllTradesSkill(level); got != want {
			t.Errorf("JackOfAllTradesSkill(%d) = %d, want %d", level, got, want)
		}
	}

	for value, want := range map[int]int{0: 0, 1: 1, 6: 3, 7: 4, 8: 4} {
		if got := AnalogCharacteristic(value); got != want {
			t.Errorf("AnalogCharacteristic(%d) = %d, want %d", value, got, want)
		}
	}
}
