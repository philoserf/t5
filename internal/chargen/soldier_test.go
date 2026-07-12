package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// TestGoldenSoldier traces a complete two-term Soldier, exercising the rank
// engine end-to-end: a medal each term, an enlisted promotion (helped by that
// medal), then a Commission to the officer track. Rolls are 3,4 (= 7) unless
// noted. Scores: Str 8, Dex 7, End 8, Int 7, Edu 7, Soc 8 ("878778").
func TestGoldenSoldier(t *testing.T) {
	seq := []int{
		// UPP: Str 8(4,4), Dex 7(3,4), End 8(4,4), Int 7(3,4), Edu 7(3,4), Soc 8(4,4).
		4, 4, 3, 4, 4, 4, 3, 4, 3, 4, 4, 4,
		3, 4, // qualify vs Str 8: 7 <= 8, enters (begins Private, Fighter-1)
		// Term 1: CC = Str (7... value 8). Risk 7 survive; Reward 7 -> Medal (1).
		3, 4, // risk
		3, 4, // reward -> Medal 1
		5, 5, // Commission vs End 8: 10 > 8, fails
		4, 4, // Enlisted Promotion vs End 8 + Medal 1 = 9: 8 <= 9, promote to Corporal
		1, 1, 1, 1, // 4 skill rolls, Peacekeeper col row 1 = Admin
		3, 4, // continue vs End 8: 7, policy wants term 2
		// Term 2: CC = End. Risk 7 survive; Reward 7 -> Medal (2).
		3, 4, // risk
		3, 4, // reward -> Medal 2
		3, 4, // Commission vs End 8: 7 <= 8, commissioned -> 2nd Lieutenant (Leader-1)
		1, 1, 1, 1, // Admin x4 again
		3, 4, // continue: policy stops after term 2
		// Muster out: 2 rolls, Benefit column (Soldier's Benefit DM is 0 here).
		2, // row 2 -> Str +1 (8 -> 9)
		3, // row 3 -> Wafer Jack
	}

	// goldenPolicy (scout_test.go) picks skill column 3; for the Soldier grid
	// that column is Peacekeeper (Admin at row 1), not the Scout's Exploration.
	c := GenerateCareered(dice.NewScripted(seq...), goldenPolicy{}, worldgen.World{}, SoldierCareer)

	if got := c.UPP(); got != "978778" {
		t.Errorf("UPP = %q, want %q (Str 8 +1 muster benefit)", got, "978778")
	}
	if c.Medals != 2 {
		t.Errorf("Medals = %d, want 2 (a Reward success each term)", c.Medals)
	}
	if c.Skills.Level("Fighter") != 1 || c.Skills.Level("Leader") != 1 || c.Skills.Level("Admin") != 8 {
		t.Errorf("skills: Fighter=%d Leader=%d Admin=%d, want 1/1/8",
			c.Skills.Level("Fighter"), c.Skills.Level("Leader"), c.Skills.Level("Admin"))
	}
	if len(c.Careers) != 1 {
		t.Fatalf("careers = %+v, want one record", c.Careers)
	}
	rec := c.Careers[0]
	if rec.Career != Soldier || rec.Terms != 2 || rec.Outcome != MusteredOut {
		t.Errorf("record = %+v, want Soldier/2 terms/MusteredOut", rec)
	}
	if !rec.Officer || rec.Rank != 1 {
		t.Errorf("rank = %d officer %v, want officer rank 1 (2nd Lieutenant)", rec.Rank, rec.Officer)
	}
	if len(c.Benefits) != 1 || c.Benefits[0] != "Wafer Jack" {
		t.Errorf("Benefits = %v, want [Wafer Jack]", c.Benefits)
	}
}

func TestPromotedMedalsAndWounds(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 7, 7, 6}, Medals: 2, WoundBadges: 1}
	// Soc 6 + Medals 2 + Wound Badges 1 = target 9; a roll of 8 succeeds.
	if !promoted(dice.NewScripted(4, 4), c, PromotionRule{Char: Social, MedalsAndWounds: true}) {
		t.Error("promotion with medal/wound mods should succeed at 8 vs target 9")
	}
	// Without the mods the target is just Soc 6; 8 fails.
	if promoted(dice.NewScripted(4, 4), c, PromotionRule{Char: Social}) {
		t.Error("promotion without mods should fail at 8 vs target 6")
	}
}

func TestResolveRankCommission(t *testing.T) {
	// An enlisted character whose Commission roll succeeds jumps to the officer
	// track at rank 1, gaining that rank's automatic skill.
	c := Character{scores: [count]int{7, 7, 8, 7, 7, 7}}
	run := careerRun{rank: 2}
	resolveRank(dice.NewScripted(3, 4), &c, &run, SoldierCareer) // Commission vs End 8: 7 <= 8
	if !run.officer || run.rank != 1 {
		t.Fatalf("after commission: officer %v rank %d, want officer rank 1", run.officer, run.rank)
	}
	if c.Skills.Level("Leader") != 1 {
		t.Errorf("2nd Lieutenant auto-skill Leader = %d, want 1", c.Skills.Level("Leader"))
	}
}
