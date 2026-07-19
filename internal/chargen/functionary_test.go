package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// TestGoldenFunctionary traces a two-term Functionary: Office Politics promotes
// on each Reward success (Clerk -> Supervisor -> Senior Supervisor) with no
// injury. Rolls are 3,4 (= 7) unless noted. Starting scores "787887" (final UPP
// 887987 after the two muster benefits).
func TestGoldenFunctionary(t *testing.T) {
	seq := []int{
		// UPP: Str 7, Dex 8, End 7, Int 8, Edu 8, Soc 7.
		3, 4, 4, 4, 3, 4, 4, 4, 4, 4, 3, 4,
		// No qualify — Functionary auto-begins (Clerk, rank 1, Bureaucrat-1).
		// Term 1: CC = Dex (8). Risk 7 keeps the job; Reward 7 -> promote to Supervisor.
		3, 4, // risk
		3, 4, // reward -> promote (rank 2)
		1, 1, 1, 1, 1, // 4 + 1 (promotion) skill rolls, General col row 1 = One Trade -> Biologics
		3, 4, // continue (Office Politics kept the job; policy wants term 2)
		// Term 2: CC = End (7). Risk 7 keeps; Reward 7 -> promote to Senior Supervisor (Admin-1).
		3, 4, // risk
		3, 4, // reward -> promote (rank 3)
		1, 1, 1, 1, 1, // Biologics x5 (4 + 1 promotion), now 10
		3, 4, // continue: policy stops after term 2
		// Muster out: 2 rolls, Benefit column. The DM is the rank number the p.87
		// ladder prints, and that ladder starts at F0 Clerk — two promotions from
		// F0 is F2 Senior Supervisor, so the DM is +2 (the engine's 1-based rank 3
		// minus one). This once read the internal index straight through as +3.
		2, // 2 + 2 = row 4 -> Str +1 (7 -> 8)
		5, // 5 + 2 = row 7 -> Int +1 (8 -> 9)
	}

	// goldenPolicy (scout_test.go) picks skill column 3; for the Functionary grid
	// that column is General (One Trade -> Biologics at row 1).
	c := GenerateCareered(
		dice.NewScripted(seq...),
		goldenPolicy{},
		worldgen.World{},
		FunctionaryCareer,
	)

	if got := c.UPP(); got != "887987" {
		t.Errorf("UPP = %q, want %q (Str 7->8, Int 8->9 muster benefits)", got, "887987")
	}

	if len(c.Benefits) != 0 {
		t.Errorf("Benefits = %v, want none (both muster rolls landed on characteristics)", c.Benefits)
	}

	if c.Skills.Level("Bureaucrat") != 1 || c.Skills.Level("Admin") != 1 ||
		c.Skills.Level("Biologics") != 10 {
		t.Errorf("skills: Bureaucrat=%d Admin=%d Biologics=%d, want 1/1/10",
			c.Skills.Level("Bureaucrat"), c.Skills.Level("Admin"), c.Skills.Level("Biologics"))
	}

	rec := c.Careers[0]
	if rec.Career != Functionary || rec.Terms != 2 || rec.Outcome != MusteredOut {
		t.Errorf("record = %+v, want Functionary/2 terms/MusteredOut", rec)
	}

	if rec.Officer || rec.Rank != 3 {
		t.Errorf(
			"rank = %d officer %v, want single-ladder rank 3 (Senior Supervisor)",
			rec.Rank,
			rec.Officer,
		)
	}
	// The muster DM is the number the p.87 ladder prints, not the engine's index:
	// the third rung of F0..F8 is F2.
	if got := benefitDM(FunctionaryCareer.BenefitDM, c, rec); got != 2 {
		t.Errorf("muster Benefit DM = %d, want 2 (F2 Senior Supervisor)", got)
	}
	// A Clerk has yet to be promoted: F0 is a DM of +0, the case the off-by-one
	// most obviously broke.
	clerk := CareerRecord{Career: Functionary, Rank: 1}
	if got := benefitDM(FunctionaryCareer.BenefitDM, c, clerk); got != 0 {
		t.Errorf("a Clerk's muster Benefit DM = %d, want 0 (F0)", got)
	}
}

// TestOfficePoliticsRiskFailureEndsCareer confirms a failed Risk ends the career
// with MusteredOut (job loss) and no injury — not Disabled.
func TestOfficePoliticsRiskFailureEndsCareer(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 7, 7, 7}}
	run := careerRun{rank: 1}
	// Risk 12 fails vs Dex 7; Reward 12 fails too; then 4 skill rolls.
	got := runPoliticsTerm(
		dice.NewScripted(6, 6, 6, 6, 1, 1, 1, 1),
		DefaultPolicy{},
		&c,
		&run,
		FunctionaryCareer,
		Dexterity,
	)
	if got != MusteredOut {
		t.Errorf("outcome = %v, want MusteredOut (job loss)", got)
	}

	if c.WoundBadges != 0 || run.rank != 1 {
		t.Errorf(
			"job loss should not injure or promote: wounds %d rank %d",
			c.WoundBadges,
			run.rank,
		)
	}
}
