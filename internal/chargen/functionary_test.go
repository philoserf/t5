package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// TestGoldenFunctionary traces a two-term Functionary: Office Politics promotes
// on each Reward success (Clerk -> Supervisor -> Senior Supervisor) with no
// injury. Rolls are 3,4 (= 7) unless noted. Starting scores "787887" (final UPP
// 797887 after the muster benefit).
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
		// Muster out: 2 rolls, Benefit column, DM +Rank (=3, Senior Supervisor).
		2, // 2 + 3 = row 5 -> Dex +1 (8 -> 9)
		5, // 5 + 3 = row 8 -> Life Insurance
	}

	// goldenPolicy (scout_test.go) picks skill column 3; for the Functionary grid
	// that column is General (One Trade -> Biologics at row 1).
	c := GenerateCareered(
		dice.NewScripted(seq...),
		goldenPolicy{},
		worldgen.World{},
		FunctionaryCareer,
	)

	if got := c.UPP(); got != "797887" {
		t.Errorf("UPP = %q, want %q (Dex 8 +1 muster benefit)", got, "797887")
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
