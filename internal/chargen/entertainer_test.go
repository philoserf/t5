package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// TestGoldenEntertainer traces a complete two-term Entertainer. Fame replaces
// Risk & Reward: term 1's Flux raises Fame (so Talent +1 and two extra skills),
// term 2's Flux lowers it (no bonus). Flux is d6-d6. Other rolls are 3,4 (= 7)
// unless noted. Starting scores "787777" (final UPP 887777 after muster).
func TestGoldenEntertainer(t *testing.T) {
	seq := []int{
		// UPP: Str 7, Dex 8, End 7, Int 7, Edu 7, Soc 7.
		3, 4, 4, 4, 3, 4, 3, 4, 3, 4, 3, 4,
		3, 4, // Begin (Actor) vs best(Dex 8, End 7) = 8: 7 <= 8, enters
		3, 4, // initial Fame and Talent = 2D = 7
		// Term 1: no Flux this edition (Term 1 = 2D); 4 base skills.
		1, 1, 1, 1, // Survey x4
		3, 4, // continue vs Fame 7: 7, policy wants term 2
		// Term 2: Flux 5-3 = +2 -> Fame 9, Talent 8, +2 skills (6 total).
		5, 3, // flux
		1, 1, 1, 1, 1, 1, // Survey x6 (Survey now 10)
		3, 4, // continue vs Fame 9: 7, policy stops after term 2
		// Muster out: 2 rolls, Benefit column, DM +Terms (=2).
		3, // 3 + 2 = row 5 -> Str +1 (7 -> 8)
		1, // 1 + 2 = row 3 -> Wafer Jack
	}

	// goldenPolicy (scout_test.go) picks skill column 3; for the Entertainer grid
	// that column is General (Survey at row 1).
	c := GenerateCareered(
		dice.NewScripted(seq...),
		goldenPolicy{},
		worldgen.World{},
		EntertainerCareer,
	)

	if got := c.UPP(); got != "887777" {
		t.Errorf("UPP = %q, want %q (Str 7 +1 muster benefit)", got, "887777")
	}

	if c.Fame != 9 || c.Talent != 8 {
		t.Errorf(
			"Fame/Talent = %d/%d, want 9/8 (start 7, no Flux term1, +2 Flux term2; Talent +1 on the rise)",
			c.Fame,
			c.Talent,
		)
	}

	if c.Skills.Level("Survey") != 10 {
		t.Errorf("Survey = %d, want 10 (6 skills term1 + 4 skills term2)", c.Skills.Level("Survey"))
	}

	rec := c.Careers[0]
	if rec.Career != Entertainer || rec.Terms != 2 || rec.Outcome != MusteredOut {
		t.Errorf("record = %+v, want Entertainer/2 terms/MusteredOut", rec)
	}

	if len(c.Benefits) != 1 || c.Benefits[0] != "Wafer Jack" {
		t.Errorf("Benefits = %v, want [Wafer Jack]", c.Benefits)
	}
}

// TestFameFallDoesNotBonus confirms a Fame drop grants no Talent and only the
// base skill eligibility, with no injury.
func TestFameFallDoesNotBonus(t *testing.T) {
	c := Character{Fame: 8, Talent: 8}
	// A later term (run.terms > 0) applies Flux: 1-6 = -5 lowers Fame; then 4 base
	// skill rolls (Survey).
	runFameTerm(
		dice.NewScripted(1, 6, 1, 1, 1, 1),
		goldenPolicy{},
		&c,
		&careerRun{terms: 1},
		EntertainerCareer,
	)

	if c.Fame != 3 || c.Talent != 8 {
		t.Errorf("Fame/Talent = %d/%d, want 3/8 (fall grants no Talent)", c.Fame, c.Talent)
	}

	if c.Skills.Level("Survey") != 4 {
		t.Errorf("Survey = %d, want 4 (base eligibility, no +2)", c.Skills.Level("Survey"))
	}
}
