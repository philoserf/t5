package combat

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestCombatNumbers(t *testing.T) {
	if ShootingNumber(8, 2, 1) != 11 || MeleeNumber(7, 0, 0) != 7 || ImpactNumber(3) != 3 {
		t.Errorf("combat numbers wrong")
	}
}

func TestTargetSize(t *testing.T) {
	if TargetSize(5, 5, Standing) != 0 || TargetSize(6, 5, Standing) != 1 {
		t.Errorf("target size wrong")
	}

	if TargetSize(5, 3, Prone) != 0 { // 5-3-2
		t.Errorf("prone target size wrong")
	}
}

func TestImpactReginaExample(t *testing.T) {
	// Book 1 p.202: a runaway groundcar Speed=3. Roberto C2=9 must roll 9-3=6 or
	// less to dodge; he rolls 5 and tumbles to safety (Success = avoided).
	if res := Impact(dice.NewScripted(2, 3), 9, 3); res.Target != 6 || !res.Success {
		t.Errorf("Roberto = %+v, want target 6, success (dodged)", res)
	}
	// Landor C2=8 must roll 8-3=5 or less; he rolls 8 and is hit (Success false).
	if res := Impact(dice.NewScripted(4, 4), 8, 3); res.Target != 5 || res.Success {
		t.Errorf("Landor = %+v, want target 5, failure (hit)", res)
	}
	// Speed 3 inflicts 3^2 = 9 dice of damage.
	if ImpactDamageDice(3) != 9 {
		t.Errorf("ImpactDamageDice(3) = %d, want 9", ImpactDamageDice(3))
	}
}

func TestImpactDamageDiceClampsNegativeSpeed(t *testing.T) {
	// Speed is a magnitude: a negative input deals zero dice, not the square
	// of its sign-stripped value.
	if got := ImpactDamageDice(-3); got != 0 {
		t.Errorf("ImpactDamageDice(-3) = %d, want 0", got)
	}

	if got := ImpactDamageDice(0); got != 0 {
		t.Errorf("ImpactDamageDice(0) = %d, want 0", got)
	}
}

func TestMeleeCayneExample(t *testing.T) {
	// Book 1 p.203: Corbett (DMN 7) allocates all 5 Dexterity to his AMN 7 ->
	// target 7+5-7 = 5 on 2D; he rolls 7 and fails.
	if res := Melee(dice.NewScripted(4, 3), 7, 7, 5); res.Target != 5 || res.Success {
		t.Errorf("Corbett = %+v, want target 5, failure (rolled 7)", res)
	}
	// A roll of 4 (under 5) would hit.
	if res := Melee(dice.NewScripted(2, 2), 7, 7, 5); !res.Success {
		t.Errorf("roll 4 vs 5 should hit: %+v", res)
	}
}

func TestRangedThisIsHard(t *testing.T) {
	// SN 10, Size 5 at Range 2 (target size 3), skill 3: no "This Is Hard!" ->
	// 2D at or under 10+3 = 13.
	res := Ranged(dice.NewScripted(3, 3), 10, 3, TargetSize(5, 2, Standing), 2)
	if res.Target != 13 || len(res.Faces) != 2 {
		t.Errorf("ranged = %+v, want target 13 on 2D", res)
	}
	// Range 4 with skill 2: the 4D roll exceeds skill, so +1D -> 5D.
	res2 := Ranged(dice.NewScripted(1, 1, 1, 1, 1), 10, 2, TargetSize(5, 4, Standing), 4)
	if len(res2.Faces) != 5 {
		t.Errorf("This Is Hard! rolled %d dice, want 5", len(res2.Faces))
	}
	// Range 0 (Contact Range) with skill 0. The floor makes this a 1D roll
	// (p.204: ranges below 1 "are all treated as 1D (since there is always some
	// slight chance ... that an unskilled attacker may miss)"), and TIH! is
	// stated in dice, not range: "If a task requires more dice than the
	// character has applicable skill levels, then increase the difficulty one
	// level" (p.128). 1D > skill-0, so the unskilled attacker rolls 2D — which
	// is exactly the miss chance the 1D floor exists to preserve.
	res3 := Ranged(dice.NewScripted(1, 1), 10, 0, TargetSize(5, 0, Standing), 0)
	if len(res3.Faces) != 2 {
		t.Errorf("Range 0 / Skill 0 rolled %d dice, want 2 (This Is Hard!)", len(res3.Faces))
	}
}

func TestRangedUnhittableTarget(t *testing.T) {
	// A prone Size-1 target at Range 1 has target size 1-1-2 = -2 and cannot be
	// attacked: the attack automatically fails regardless of the roll.
	ts := TargetSize(1, 1, Prone)
	if ts >= 0 {
		t.Fatalf("target size = %d, want negative", ts)
	}

	if res := Ranged(dice.NewScripted(1), 10, 3, ts, 1); res.Success {
		t.Errorf("attack on unhittable target should fail: %+v", res)
	}
}

func TestAbsorbAndWound(t *testing.T) {
	// BA-11 armor (Ar=32) absorbs 32 hits; 40 hits leave 8.
	if Absorb(40, 32) != 8 || Absorb(20, 32) != 0 {
		t.Errorf("Absorb wrong")
	}
	// Landor's 989765: physicals 9/8/9 (=26) take 25 hits, cascading to leave a
	// single Hit Point in Endurance (Book 1 p.202).
	if got := Wound([3]int{9, 8, 9}, 25); got != [3]int{0, 0, 1} {
		t.Errorf("Wound = %v, want [0 0 1]", got)
	}
}
