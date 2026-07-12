package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestLifeStage(t *testing.T) {
	cases := map[int]int{
		0: 0, 1: 0, 2: 1, 9: 1, 10: 2, 17: 2, 18: 3, 33: 4,
		34: 5, 41: 5, 42: 6, 65: 8, 66: 9, 71: 9,
	}
	for age, want := range cases {
		if got := lifeStage(age); got != want {
			t.Errorf("lifeStage(%d) = %d, want %d", age, got, want)
		}
	}
}

func TestGenerateStartsAtEighteen(t *testing.T) {
	c := Generate(dice.NewScripted(4, 4))
	if c.Age != 18 || c.LifeStage() != 3 {
		t.Fatalf("Generate age = %d (stage %d), want 18 (stage 3)", c.Age, c.LifeStage())
	}
}

func TestAgingCascade(t *testing.T) {
	cases := []struct {
		zeroed, prior int
		wantIllness   Illness
		wantDead      bool
	}{
		{0, 0, NoIllness, false},
		{1, 0, NoIllness, false},
		{2, 0, MajorIllness, false},
		{3, 0, ExtremeIllness, false}, // first extreme illness survives
		{3, 1, ExtremeIllness, true},  // the second is fatal
	}
	for _, c := range cases {
		gotIll, gotDead := agingCascade(c.zeroed, c.prior)
		if gotIll != c.wantIllness || gotDead != c.wantDead {
			t.Errorf("agingCascade(%d,%d) = (%v,%v), want (%v,%v)",
				c.zeroed, c.prior, gotIll, gotDead, c.wantIllness, c.wantDead)
		}
	}
}

func TestAgingCheckBeforePeakIsNoOp(t *testing.T) {
	c := Character{scores: [count]int{8, 8, 8, 8, 8, 8}, Age: 30}
	AgingCheck(dice.NewScripted(2), &c) // would age everything if it ran
	if c.scores != [count]int{8, 8, 8, 8, 8, 8} {
		t.Fatalf("aging before 34 changed scores: %v", c.scores)
	}
}

func TestAgingCheckDecrements(t *testing.T) {
	// Age 34 (stage 5): every 2D of 4 (< 5) ages the three physical chars by 1;
	// Intelligence (mental) is untouched until 66.
	c := Character{scores: [count]int{8, 8, 8, 8, 8, 8}, Age: 34}
	AgingCheck(dice.NewScripted(2), &c) // Dice(2) = 4 each
	if c.scores != [count]int{7, 7, 7, 8, 8, 8} {
		t.Fatalf("after aging = %v, want [7 7 7 8 8 8]", c.scores)
	}
	// A high roll (12) never ages.
	c2 := Character{scores: [count]int{8, 8, 8, 8, 8, 8}, Age: 34}
	AgingCheck(dice.NewScripted(6), &c2)
	if c2.scores != [count]int{8, 8, 8, 8, 8, 8} {
		t.Fatalf("high rolls aged the character: %v", c2.scores)
	}
}

func TestAgingCheckResetsZeroToOne(t *testing.T) {
	// Strength 1 ages to 0 and resets to 1; only one characteristic zeroed, so
	// no illness or death.
	c := Character{scores: [count]int{1, 8, 8, 8, 8, 8}, Age: 34}
	// Str rolls 4 (ages), Dex and End roll 12 (safe).
	AgingCheck(dice.NewScripted(2, 2, 6, 6, 6, 6), &c)
	if c.scores[Strength] != 1 || c.Dead || c.extremeAgings != 0 {
		t.Fatalf("reset case = %v dead=%v extreme=%d, want Str 1, alive, 0 extreme", c.scores, c.Dead, c.extremeAgings)
	}
}

func TestAgingCheckDeathOnSecondExtreme(t *testing.T) {
	// Three physical characteristics at 1: each aging check zeroes all three.
	c := Character{scores: [count]int{1, 1, 1, 8, 8, 8}, Age: 34}
	AgingCheck(dice.NewScripted(2), &c) // first extreme illness
	if c.Dead || c.extremeAgings != 1 {
		t.Fatalf("after first extreme: dead=%v extreme=%d, want alive, 1", c.Dead, c.extremeAgings)
	}
	AgingCheck(dice.NewScripted(2), &c) // second extreme illness kills
	if !c.Dead {
		t.Fatalf("second extreme illness did not kill")
	}
	// A dead character is not aged further.
	AgingCheck(dice.NewScripted(2), &c)
}
