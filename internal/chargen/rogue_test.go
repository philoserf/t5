package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// TestGoldenRogue traces a complete two-term Rogue, exercising the fixed-CC seam
// end-to-end. The character is all 7s, so the policy's fixed CC (the first
// controlling characteristic, Strength) and the qualify target (best of all six)
// coincide at 7. Every roll is 3,4 (= 7) unless noted.
func TestGoldenRogue(t *testing.T) {
	seq := []int{
		// UPP: six 2D, each 7 -> "777777"; qualify target = best of all six = 7.
		3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4,
		3, 4, // qualify: 7 <= 7, enters the career
		// Term 1: fixed CC = Strength (7). Risk 7 survive; Reward 7.
		3, 4, // risk
		3, 4, // reward
		// EligPerTerm 2 skill rolls from Space Travel (col 3), row 1 = Starship
		// Skill -> policy picks Pilot (flat).
		1, 1,
		3, 4, // continue (UseCC -> vs Strength 7): 7, policy wants term 2
		// Term 2: fixed CC still Strength.
		3, 4, // risk
		3, 4, // reward
		1, 1, // Pilot x2 again
		3, 4, // continue: policy stops after term 2
		// Muster out: 2 rolls (2 terms), Benefit column. Rogue's Benefit column
		// also gets the +Terms DM (=2).
		3, // 3 + 2 = row 5 -> End +1 (7 -> 8)
		4, // 4 + 2 = row 6 -> Life Insurance
	}

	c := GenerateCareered(dice.NewScripted(seq...), goldenPolicy{}, worldgen.World{}, RogueCareer)

	if got := c.UPP(); got != "778777" {
		t.Errorf("UPP = %q, want %q (End 7 +1 muster benefit)", got, "778777")
	}
	if c.Age != 26 {
		t.Errorf("Age = %d, want 26 (18 + 2 terms)", c.Age)
	}
	if got := c.Skills.Level("Pilot"); got != 4 {
		t.Errorf("Pilot = %d, want 4 (2 rolls x 2 terms)", got)
	}
	if len(c.Careers) != 1 {
		t.Fatalf("careers = %+v, want one record", c.Careers)
	}
	rec := c.Careers[0]
	if rec.Career != Rogue || rec.Terms != 2 || rec.Outcome != MusteredOut {
		t.Errorf("record = %+v, want Rogue/2 terms/MusteredOut", rec)
	}
	if len(c.Benefits) != 1 || c.Benefits[0] != "Life Insurance" {
		t.Errorf("Benefits = %v, want [Life Insurance]", c.Benefits)
	}
	if c.Credits != 0 {
		t.Errorf("Credits = %d, want 0 (both muster rolls took benefits)", c.Credits)
	}
}

func TestRogueFixedCCChosenOnce(t *testing.T) {
	// Under FixedCC the policy selects the CC once (highest score) and it is
	// reused every term.
	c := Character{scores: [count]int{6, 6, 9, 6, 6, 6}} // Endurance highest
	run := careerRun{}
	first := selectCC(DefaultPolicy{}, c, &run, RogueCareer)
	if first != Endurance {
		t.Fatalf("first CC = %v, want Endurance (highest)", first)
	}
	// Even if scores later shift, the fixed choice holds.
	c.scores[Strength] = 15
	if again := selectCC(DefaultPolicy{}, c, &run, RogueCareer); again != Endurance {
		t.Errorf("second CC = %v, want the same fixed Endurance", again)
	}
}
