package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// TestGoldenAgent traces a complete two-term Agent, exercising the Undercover
// Assignment: each term borrows a rolled career's skill and, on a Successful
// Mission (a held Risk), grants Per Term 2 + Successful Mission 4 = 6 skills,
// while a successful Reward is a Commendation. Rolls are 3,4 (= 7) unless noted.
// Starting scores "888777" (final UPP 988777 after the Str +1 muster benefit).
func TestGoldenAgent(t *testing.T) {
	seq := []int{
		// UPP: Str 8, Dex 8, End 8, Int 7, Edu 7, Soc 7.
		4, 4, 4, 4, 4, 4, 3, 4, 3, 4, 3, 4,
		3, 4, // qualify vs End 8: 7 <= 8, enters
		// Term 1: CC = Str (8). Risk survive (Successful Mission); Reward -> Commendation.
		3, 4, // risk
		3, 4, // reward -> Commendation 1
		1, 1, // Undercover Assignment A=1,B=1 -> Soldier; select col 3 row 1 = Admin
		1, 1, 1, 1, 1, 1, // Per Term 2 + Successful Mission 4 = 6, Mission col row 1 = Survey
		3, 4, // continue vs Str 8 + Terms 1 = 9: 7, policy wants term 2
		// Term 2: CC = Dex (8).
		3, 4, // risk
		3, 4, // reward -> Commendation 2
		1, 1, // Undercover -> Soldier again; Admin
		1, 1, 1, 1, 1, 1, // Survey x6 again -> 12
		3, 4, // continue vs Str 8 + Terms 2 = 10: 7, policy stops after term 2
		// Muster out: 4 rolls (2 terms + 2 Commendations, Book 1 p.67), Benefit
		// column, DM +Commendations (=2).
		3, // 3 + 2 = row 5 -> Str +1 (8 -> 9)
		1, // 1 + 2 = row 3 -> Wafer Jack
		6, // 6 + 2 = row 8 -> Ship Share
		6, // 6 + 2 = row 8 -> Ship Share
	}

	// goldenPolicy (scout_test.go) picks skill column 3; for the Agent grid that
	// column is Mission (Survey at row 1), and for the borrowed Soldier grid it
	// is Peacekeeper (Admin at row 1).
	c := GenerateCareered(dice.NewScripted(seq...), goldenPolicy{}, worldgen.World{}, AgentCareer)

	if got := c.UPP(); got != "988777" {
		t.Errorf("UPP = %q, want %q (Str 8 +1 muster benefit)", got, "988777")
	}

	if c.MedalCount() != 0 {
		t.Errorf("Medals = %d, want 0 (the Agent earns Commendations, not Medals)", c.MedalCount())
	}

	if c.Commendations != 2 {
		t.Errorf("Commendations = %d, want 2 (a Reward success each term)", c.Commendations)
	}

	if c.Skills.Level("Survey") != 12 {
		t.Errorf("Survey = %d, want 12 (6 rolls x 2 terms)", c.Skills.Level("Survey"))
	}

	if c.Skills.Level("Admin") != 2 {
		t.Errorf(
			"Admin = %d, want 2 (one Undercover skill borrowed from the Soldier each term)",
			c.Skills.Level("Admin"),
		)
	}

	rec := c.Careers[0]
	if rec.Career != Agent || rec.Terms != 2 || rec.Outcome != MusteredOut {
		t.Errorf("record = %+v, want Agent/2 terms/MusteredOut", rec)
	}

	if rec.Rank != 0 || rec.Officer {
		t.Errorf("rank = %d officer %v, want rankless (0, false)", rec.Rank, rec.Officer)
	}

	if got := c.Benefits; len(got) != 3 || got[0] != "Wafer Jack" || got[1] != "Ship Share" ||
		got[2] != "Ship Share" {
		t.Errorf("Benefits = %v, want [Wafer Jack, Ship Share, Ship Share]", c.Benefits)
	}
}

func TestUndercoverAssignment(t *testing.T) {
	// The A/B table maps a roll to the borrowed career.
	if got := undercoverAssignment(dice.NewScripted(1, 1)); got != Soldier {
		t.Errorf("A1 B1 = %v, want Soldier", got)
	}

	if got := undercoverAssignment(dice.NewScripted(2, 3)); got != Entertainer {
		t.Errorf("A2 B3 = %v, want Entertainer", got)
	}
	// Die A rerolls while it exceeds 3: 5 (reroll) then 3, with B = 5 -> Noble.
	if got := undercoverAssignment(dice.NewScripted(5, 3, 5)); got != Noble {
		t.Errorf("A(5->reroll->3) B5 = %v, want Noble", got)
	}
}

func TestAwardUndercoverFailedMission(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 7, 7, 7}}
	// A failed Mission grants only Per Term 2 plus the 1 Undercover skill, not the
	// Successful Mission 4. Undercover A=1,B=1 -> Soldier, borrow col 3 row 1 =
	// Admin; then 2 Survey from the Agent's own Mission column.
	awardUndercover(dice.NewScripted(1, 1, 1, 1), goldenPolicy{}, &c, AgentCareer, false)

	if c.Skills.Level("Admin") != 1 {
		t.Errorf("Admin = %d, want 1 (the borrowed Undercover skill)", c.Skills.Level("Admin"))
	}

	if c.Skills.Level("Survey") != 2 {
		t.Errorf(
			"Survey = %d, want 2 (Per Term 2, no Successful Mission bonus)",
			c.Skills.Level("Survey"),
		)
	}
}

func TestContinueTermsMod(t *testing.T) {
	// A "Str, Mod +Terms" Continue: at Str 6 with 2 terms served the target is 8,
	// so a roll of 7 continues — where a plain Str-6 rule would fail.
	c := Character{scores: [count]int{6, 7, 7, 7, 7, 7}}
	rule := ContinueRule{UseChar: true, Char: Strength, TermsMod: true}
	career := Career{Continue: rule, Name: "T"}

	rec := CareerRecord{Terms: 2}
	if !continues(dice.NewScripted(3, 4), alwaysContinue{}, c, career, rec, &careerRun{}) {
		t.Error("+Terms Continue should succeed at 7 vs target Str6 + 2 terms = 8")
	}
	// Without the +Terms bonus the target is just Str 6, and 7 fails.
	career.Continue = ContinueRule{UseChar: true, Char: Strength}
	if continues(dice.NewScripted(3, 4), alwaysContinue{}, c, career, rec, &careerRun{}) {
		t.Error("plain Str Continue should fail at 7 vs target 6")
	}
}

// alwaysContinue is a policy that always wishes to continue.
type alwaysContinue struct{ DefaultPolicy }

func (alwaysContinue) Continue(Character, CareerRecord) bool { return true }
