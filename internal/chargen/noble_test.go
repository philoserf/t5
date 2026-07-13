package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// TestGoldenNoble traces a two-term Noble. Return & Intrigue: term 1's Intrigue
// succeeds and the Elevation (roll-high vs Soc 10) succeeds, raising Soc to 11
// (a Land Grant); term 2's Intrigue succeeds but the Elevation vs Soc 11 fails.
// Rolls are 3,4 (= 7) unless noted. Starting scores "78788A" (final UPP 79788B).
func TestGoldenNoble(t *testing.T) {
	seq := []int{
		// UPP: Str 7, Dex 8, End 7, Int 8, Edu 8, Soc 10.
		3, 4, 4, 4, 3, 4, 4, 4, 4, 4, 5, 5,
		// No qualify — Noble auto-begins (not in Exile).
		// Term 1: CC = Dex (8). Intrigue 7 <= 8 succeeds; Elevation 2D >= Soc 10.
		3, 4, // intrigue
		5, 5, // elevation: 10 >= 10 -> Soc 10 -> 11, Land Grant
		1, 1, 1, 1, 1, 1, // 4 + 2 (When Elevated) skill rolls, General col row 1 = Advocate
		3, 4, // continue vs 7: 7, policy wants term 2
		// Term 2: CC = End (7). Intrigue 7 <= 7 succeeds; Elevation 2D >= Soc 11 fails.
		3, 4, // intrigue
		3, 4, // elevation: 7 < 11 -> no elevation (no bonus skills)
		1, 1, 1, 1, // Advocate x4 (now 10)
		3, 4, // continue: policy stops after term 2
		// Muster out: 2 rolls, Benefit column, DM +Terms (=2).
		4, // 4 + 2 = row 6 -> Dex +1 (8 -> 9)
		1, // 1 + 2 = row 3 -> Wafer Jack
	}

	// goldenPolicy (scout_test.go) picks skill column 3; for the Noble grid that
	// column is General (Advocate at row 1).
	c := GenerateCareered(dice.NewScripted(seq...), goldenPolicy{}, worldgen.World{}, NobleCareer)

	if got := c.UPP(); got != "79788B" {
		t.Errorf("UPP = %q, want %q (Soc 10 -> 11 by Elevation, Dex +1 muster)", got, "79788B")
	}
	if c.Score(Social) != 11 || c.LandGrants != 1 {
		t.Errorf("Soc = %d, LandGrants = %d, want 11 / 1 (one Elevation)", c.Score(Social), c.LandGrants)
	}
	if c.Skills.Level("Advocate") != 10 {
		t.Errorf("Advocate = %d, want 10 (6 elevated term 1 + 4 term 2)", c.Skills.Level("Advocate"))
	}
	rec := c.Careers[0]
	if rec.Career != Noble || rec.Terms != 2 || rec.Outcome != MusteredOut {
		t.Errorf("record = %+v, want Noble/2 terms/MusteredOut", rec)
	}
	if len(c.Benefits) != 1 || c.Benefits[0] != "Wafer Jack" {
		t.Errorf("Benefits = %v, want [Wafer Jack]", c.Benefits)
	}
}

// TestIntrigueFailureExiles confirms a failed Intrigue exiles the Noble (no
// injury, no Elevation), and a subsequent Return roll can end the Exile.
func TestIntrigueFailureExiles(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 7, 7, 10}}
	run := careerRun{}
	// Intrigue 12 fails vs Dex 7 -> Exile; then 4 skill rolls.
	runIntrigueTerm(dice.NewScripted(6, 6, 1, 1, 1, 1), goldenPolicy{}, &c, &run, NobleCareer, Dexterity)
	if !run.exiled || c.Score(Social) != 10 || c.WoundBadges != 0 {
		t.Fatalf("failed Intrigue: exiled %v Soc %d wounds %d, want exiled/10/0", run.exiled, c.Score(Social), c.WoundBadges)
	}
	// In Exile, a successful Return (7 <= Dex 7) ends it.
	runIntrigueTerm(dice.NewScripted(3, 4, 1, 1, 1, 1), goldenPolicy{}, &c, &run, NobleCareer, Dexterity)
	if run.exiled {
		t.Error("successful Return should end the Exile")
	}
}
