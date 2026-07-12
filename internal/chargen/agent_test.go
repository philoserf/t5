package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// TestGoldenAgent traces a complete two-term Agent — a rankless career, so no
// rank step runs and no medals accrue. Rolls are 3,4 (= 7) unless noted.
// Starting scores "888777" (final UPP 988777 after the Str +1 muster benefit).
func TestGoldenAgent(t *testing.T) {
	seq := []int{
		// UPP: Str 8, Dex 8, End 8, Int 7, Edu 7, Soc 7.
		4, 4, 4, 4, 4, 4, 3, 4, 3, 4, 3, 4,
		3, 4, // qualify vs End 8: 7 <= 8, enters
		// Term 1: CC = Str (8). Risk survive; Reward (no medal — rankless).
		3, 4, // risk
		3, 4, // reward
		1, 1, // 2 skill rolls, Mission col row 1 = Survey
		3, 4, // continue vs Str 8 + Terms 1 = 9: 7, policy wants term 2
		// Term 2: CC = Dex (8).
		3, 4, // risk
		3, 4, // reward
		1, 1, // Survey x2 again
		3, 4, // continue vs Str 8 + Terms 2 = 10: 7, policy stops after term 2
		// Muster out: 2 rolls, Benefit column.
		5, // row 5 -> Str +1 (8 -> 9)
		1, // row 1 -> Ship Share
	}

	// goldenPolicy (scout_test.go) picks skill column 3; for the Agent grid that
	// column is Mission (Survey at row 1).
	c := GenerateCareered(dice.NewScripted(seq...), goldenPolicy{}, worldgen.World{}, AgentCareer)

	if got := c.UPP(); got != "988777" {
		t.Errorf("UPP = %q, want %q (Str 8 +1 muster benefit)", got, "988777")
	}
	if c.Medals != 0 {
		t.Errorf("Medals = %d, want 0 (rankless career earns no medals)", c.Medals)
	}
	if c.Skills.Level("Survey") != 4 {
		t.Errorf("Survey = %d, want 4 (2 rolls x 2 terms)", c.Skills.Level("Survey"))
	}
	rec := c.Careers[0]
	if rec.Career != Agent || rec.Terms != 2 || rec.Outcome != MusteredOut {
		t.Errorf("record = %+v, want Agent/2 terms/MusteredOut", rec)
	}
	if rec.Rank != 0 || rec.Officer {
		t.Errorf("rank = %d officer %v, want rankless (0, false)", rec.Rank, rec.Officer)
	}
	if len(c.Benefits) != 1 || c.Benefits[0] != "Ship Share" {
		t.Errorf("Benefits = %v, want [Ship Share]", c.Benefits)
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
