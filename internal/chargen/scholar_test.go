package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// TestGoldenScholar traces a complete two-term Scholar: standard Risk & Reward
// (each Reward success is a Publication), and single-ladder promotion whose
// target rises with Publications. Rolls are 3,4 (= 7) unless noted. Starting
// scores "777887" (final UPP 877887 after the muster benefit).
func TestGoldenScholar(t *testing.T) {
	seq := []int{
		// UPP: Str 7, Dex 7, End 7, Int 8, Edu 8, Soc 7.
		3, 4, 3, 4, 3, 4, 4, 4, 4, 4, 3, 4,
		3, 4, // qualify vs Edu 8: 7 <= 8, enters (Lecturer, rank 1)
		// Term 1: CC = Str (7). Research 7 survives; Publication 7 -> Publication 1.
		3, 4, // risk (research)
		3, 4, // reward (publication) -> Publications 1
		4, 4, // Promotion vs Int 8 + Pubs 1 = 9: 8 <= 9, promote to Instructor (rank 2)
		1, 1, 1, 1, 1, // 4 + 1 (promotion) skill rolls, General col row 1 = Survey
		3, 4, // continue vs Edu 8 + Pubs 1 = 9: 7, policy wants term 2
		// Term 2: CC = Dex (7).
		3, 4, // research
		3, 4, // publication -> Publications 2
		4, 4, // Promotion vs Int 8 + Pubs 2 = 10: 8 <= 10, promote to Assistant Professor (rank 3)
		1, 1, 1, 1, 1, // Survey x5 (4 + 1 promotion), now 10
		3, 4, // continue: policy stops after term 2
		// Muster out: 2 rolls, Benefit column, DM +Terms (=2).
		3, // 3 + 2 = row 5 -> Str +1 (7 -> 8)
		1, // 1 + 2 = row 3 -> Wafer Jack
	}

	// goldenPolicy (scout_test.go) picks skill column 3; for the Scholar grid that
	// column is General (Survey at row 1).
	c := GenerateCareered(dice.NewScripted(seq...), goldenPolicy{}, worldgen.World{}, ScholarCareer)

	if got := c.UPP(); got != "877887" {
		t.Errorf("UPP = %q, want %q (Str 7 +1 muster benefit)", got, "877887")
	}
	if c.Publications != 2 {
		t.Errorf("Publications = %d, want 2 (a Publication each term)", c.Publications)
	}
	if c.Skills.Level("Survey") != 10 {
		t.Errorf("Survey = %d, want 10", c.Skills.Level("Survey"))
	}
	rec := c.Careers[0]
	if rec.Career != Scholar || rec.Terms != 2 || rec.Outcome != MusteredOut {
		t.Errorf("record = %+v, want Scholar/2 terms/MusteredOut", rec)
	}
	if rec.Officer || rec.Rank != 3 {
		t.Errorf("rank = %d officer %v, want single-ladder rank 3 (Assistant Professor)", rec.Rank, rec.Officer)
	}
	if len(c.Benefits) != 1 || c.Benefits[0] != "Wafer Jack" {
		t.Errorf("Benefits = %v, want [Wafer Jack]", c.Benefits)
	}
}
