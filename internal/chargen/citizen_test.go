package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// TestGoldenCitizen traces a complete two-term Citizen — an auto-begin career
// whose Citizen Life roll grants a Job (level 4) on the first success and a
// Hobby (level 2) on the second, with no career injury. Rolls are 3,4 (= 7)
// unless noted. Starting scores "777777" (final UPP 877777 after the Str +1
// muster benefit).
func TestGoldenCitizen(t *testing.T) {
	seq := []int{
		// UPP: "777777".
		3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4,
		// No qualify roll — Citizen auto-begins.
		// Term 1: CC = Str (7). Citizen Life 7 <= 7 succeeds -> Job (Admin-4).
		3, 4, // citizen life
		5, 5, 5, 5, // 4 skill rolls, General col row 5 = Bureaucrat (avoids the Admin Job)
		3, 4, // continue vs 10: 7, policy wants term 2
		// Term 2: CC = Dex (7). Citizen Life succeeds -> Hobby (Broker-2).
		3, 4, // citizen life
		5, 5, 5, 5, // Bureaucrat x4 again
		3, 4, // continue: policy stops after term 2
		// Muster out: 2 rolls, Benefit column, Benefit DM +Terms (=2).
		3, // 3 + 2 = row 5 -> Str +1 (7 -> 8)
		1, // 1 + 2 = row 3 -> Wafer Jack
	}

	// goldenPolicy (scout_test.go) picks skill column 3; for the Citizen grid that
	// column is General. Its ChooseSkill takes the first option, so the Job is
	// Admin and the Hobby is Broker.
	c := GenerateCareered(dice.NewScripted(seq...), goldenPolicy{}, worldgen.World{}, CitizenCareer)

	if got := c.UPP(); got != "877777" {
		t.Errorf("UPP = %q, want %q (Str 7 +1 muster benefit)", got, "877777")
	}

	if c.Skills.Level("Admin") != 4 || c.Skills.Level("Broker") != 2 {
		t.Errorf("Job/Hobby: Admin=%d Broker=%d, want 4/2",
			c.Skills.Level("Admin"), c.Skills.Level("Broker"))
	}

	if c.Skills.Level("Bureaucrat") != 8 {
		t.Errorf("Bureaucrat = %d, want 8 (grid skills)", c.Skills.Level("Bureaucrat"))
	}

	rec := c.Careers[0]
	if rec.Career != Citizen || rec.Terms != 2 || rec.Outcome != MusteredOut {
		t.Errorf("record = %+v, want Citizen/2 terms/MusteredOut", rec)
	}

	if len(c.Benefits) != 1 || c.Benefits[0] != "Wafer Jack" {
		t.Errorf("Benefits = %v, want [Wafer Jack]", c.Benefits)
	}
}

// TestCitizenLifeFailureIsHarmless confirms a failed Citizen Life inflicts no
// injury and awards no Job/Hobby skill, and the character survives the term.
func TestCitizenLifeFailureIsHarmless(t *testing.T) {
	c := Character{scores: [count]int{5, 5, 5, 5, 5, 5}}
	run := careerRun{}
	// Citizen Life vs Str 5 with a roll of 12 fails; no skills, no wound.
	got := runCitizenTerm(
		dice.NewScripted(6, 6, 1, 1, 1, 1),
		DefaultPolicy{},
		&c,
		&run,
		CitizenCareer,
		Strength,
	)
	if got != Ongoing {
		t.Errorf("outcome = %v, want Ongoing", got)
	}

	if run.job != "" || c.WoundBadges != 0 {
		t.Errorf(
			"failed Citizen Life gave job %q or wounds %d, want none/0",
			run.job,
			c.WoundBadges,
		)
	}
}
