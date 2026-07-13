package senses

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestNoticeAtRangeReginaExample(t *testing.T) {
	// Book 1 p.190 worked example: Human V-16 spots a Size-6 cargo mover at
	// Range 6 moving Very Fast (+2). Target = 16 + (6-6) + 2 = 18, rolled on 6D.
	res := NoticeAtRange(dice.NewScripted(2, 2, 2, 2, 2, 2), Vision, 6, 6, 2) // 6D = 12
	if res.Target != 18 {
		t.Errorf("target = %d, want 18", res.Target)
	}
	if res.Roll != 12 || !res.Success {
		t.Errorf("roll 12 vs 18 should succeed: %+v", res)
	}
	// A roll of 24 (6x4) exceeds 18 and fails.
	if res := NoticeAtRange(dice.NewScripted(4, 4, 4, 4, 4, 4), Vision, 6, 6, 2); res.Success {
		t.Errorf("roll 24 vs 18 should fail: %+v", res)
	}

	// Second example: Size-6 mover at Range 5, motionless. Target = 16 + 1 = 17
	// on 5D.
	res2 := NoticeAtRange(dice.NewScripted(3, 3, 3, 3, 3), Vision, 6, 5) // 5D = 15
	if res2.Target != 17 || res2.Roll != 15 || !res2.Success {
		t.Errorf("Range-5 example: target/roll = %d/%d, want 17/15 success", res2.Target, res2.Roll)
	}
}

func TestRangeDiceFloor(t *testing.T) {
	// Range 0 (contact) and the R/T sub-bands roll a single die, not zero.
	res := NoticeAtRange(dice.NewScripted(6), Vision, 3, 0)
	if len(res.Faces) != 1 {
		t.Errorf("Range 0 rolled %d dice, want 1", len(res.Faces))
	}
}

func TestNoticeInContact(t *testing.T) {
	// Touch (Constant 6): 2D under 6 + benchmark.
	res := NoticeInContact(dice.NewScripted(2, 2), Touch, 3) // 2D=4 vs 6+3=9
	if res.Target != 9 || len(res.Faces) != 2 || !res.Success {
		t.Errorf("touch = %+v, want target 9, 2 dice, success", res)
	}
}

func TestRangeBand(t *testing.T) {
	cases := map[float64]int{5: 1, 50: 2, 200: 3, 1000: 5, 50_000: 7, 0: 0}
	for meters, want := range cases {
		if got := RangeBand(meters); got != want {
			t.Errorf("RangeBand(%gm) = %d, want %d", meters, got, want)
		}
	}
}

func TestSenseString(t *testing.T) {
	if Vision.String() != "V-16" || Touch.String() != "T-6" {
		t.Errorf("Sense.String = %q/%q, want V-16/T-6", Vision.String(), Touch.String())
	}
}
