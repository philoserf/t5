package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// TestGoldenSpacer traces a complete two-term Spacer, confirming the naval
// Rating ladder and its Dex-based promotion flow through the rank engine. A
// medal each term, a Rating promotion, then a Commission to Ensign (whose auto
// skill Astrogation stacks with the grid). Rolls are 3,4 (= 7) unless noted.
// Starting scores "887877" (final UPP 987877 after the Str +1 muster benefit).
func TestGoldenSpacer(t *testing.T) {
	seq := []int{
		// UPP: Str 8, Dex 8, End 7, Int 8, Edu 10(5,5), Soc 7. Edu 10 gives the +2
		// Branch/Operations bonus, used to reach the Technical branch and Siege
		// operations so the net Branch/Ops mod is 0 and Risk & Reward are unchanged.
		4, 4, 4, 4, 3, 4, 4, 4, 5, 5, 3, 4,
		3, 4, // qualify vs Int 8: 7 <= 8, enters (Spacehand, Fighter-1)
		5, // Branch: 5 + 2 = 7 -> Technical (mod 0, Ops DM 0)
		// Term 1: 4 Operations rolls, each 1 + 0 + 2 = 3 -> Siege (mod 0); net 0.
		1, 1, 1, 1,
		3, 4, // risk survive; Reward -> Medal (1)
		3, 4, // reward -> Medal 1
		5, 5, // Commission vs Dex 8: 10 > 8, fails
		4, 4, // Rating Promotion vs Dex 8 + Medal 1 = 9: 8 <= 9, promote to Able Spacer
		1, 1, 1, 1, // 4 skill rolls, Patrol/Strike col row 1 = Astrogation
		3, 4, // continue vs Str 8: 7, policy wants term 2
		// Term 2: Operations again (net 0). Risk survive; Reward -> Medal (2).
		1, 1, 1, 1,
		3, 4, // risk
		3, 4, // reward -> Medal 2
		3, 4, // Commission vs Dex 8: 7 <= 8, commissioned -> Ensign (Astrogation-1)
		1, 1, 1, 1, // Astrogation x4 again
		3, 4, // continue: policy stops after term 2
		// Muster out: 2 rolls, Benefit column, DM +Officer Rank (=1, Ensign).
		1, // 1 + 1 = row 2 -> Str +1 (8 -> 9)
		2, // 2 + 1 = row 3 -> Wafer Jack
	}

	// goldenPolicy (scout_test.go) picks skill column 3; for the Spacer grid that
	// column is Patrol/Strike (Astrogation at row 1).
	c := GenerateCareered(dice.NewScripted(seq...), goldenPolicy{}, worldgen.World{}, SpacerCareer)

	if got := c.UPP(); got != "9878A7" {
		t.Errorf("UPP = %q, want %q (Str 8 +1 muster benefit, Edu 10)", got, "9878A7")
	}
	if c.Medals != 2 {
		t.Errorf("Medals = %d, want 2", c.Medals)
	}
	// Astrogation: 4 (term 1 grid) + 1 (Ensign auto-skill) + 4 (term 2 grid) = 9.
	if c.Skills.Level("Fighter") != 1 || c.Skills.Level("Astrogation") != 9 {
		t.Errorf("skills: Fighter=%d Astrogation=%d, want 1/9",
			c.Skills.Level("Fighter"), c.Skills.Level("Astrogation"))
	}
	rec := c.Careers[0]
	if rec.Career != Spacer || rec.Terms != 2 || rec.Outcome != MusteredOut {
		t.Errorf("record = %+v, want Spacer/2 terms/MusteredOut", rec)
	}
	if !rec.Officer || rec.Rank != 1 {
		t.Errorf("rank = %d officer %v, want officer rank 1 (Ensign)", rec.Rank, rec.Officer)
	}
	if len(c.Benefits) != 1 || c.Benefits[0] != "Wafer Jack" {
		t.Errorf("Benefits = %v, want [Wafer Jack]", c.Benefits)
	}
}
