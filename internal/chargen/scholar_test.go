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
		// Begin is automatic (Edu 8 >= 8): enters at Lecturer (rank 1), no roll.
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
	if c.Tenured {
		t.Error("an Edu-8 Scholar cannot earn Tenure (needs Edu 10+)")
	}
}

// TestResolveRankTenureGate: at rank 3 an untenured Scholar applies for Tenure
// (not a promotion); once tenured, a promotion roll advances past rank 3.
func TestResolveRankTenureGate(t *testing.T) {
	c := &Character{scores: [count]int{7, 7, 7, 10, 10, 7}, Publications: 3} // Int 10, Edu 10
	run := &careerRun{rank: 3}

	// Tenure application: 2D=4 <= Publications×3 = 9 -> tenured, no promotion.
	if resolveRank(dice.NewScripted(2, 2), c, run, ScholarCareer) {
		t.Error("a Tenure application is not a promotion")
	}
	if !c.Tenured || run.rank != 3 {
		t.Errorf("after tenure: Tenured=%v rank=%d, want true/3", c.Tenured, run.rank)
	}
	// Now tenured: 2D=4 <= Int 10 + Pubs 3 = 13 -> promote to Associate Professor.
	if !resolveRank(dice.NewScripted(2, 2), c, run, ScholarCareer) {
		t.Error("a tenured Scholar should promote past rank 3")
	}
	if run.rank != 4 {
		t.Errorf("rank = %d, want 4 (Associate Professor)", run.rank)
	}
}

// TestScholarBegin: Edu 8+ auto-begins the Scholar; Edu 7- must roll 2D <= Edu
// (Book 1 p.76's "To Begin (Edu 8+) Automatic" / "To Begin Edu or Tra").
func TestScholarBegin(t *testing.T) {
	// Edu 8: automatic (the dummy die is left unconsumed).
	edu8 := &Character{scores: [count]int{7, 7, 7, 7, 8, 7}}
	if !beginCareer(dice.NewScripted(1), edu8, ScholarCareer) {
		t.Error("Edu 8 should auto-begin the Scholar")
	}
	// Edu 5, roll 4 (<= 5): begins as an Amateur.
	edu5 := &Character{scores: [count]int{7, 7, 7, 7, 5, 7}}
	if !beginCareer(dice.NewScripted(2, 2), edu5, ScholarCareer) {
		t.Error("Edu 5 rolling 4 (<= 5) should begin as an Amateur")
	}
	// Edu 5, roll 11 (> 5): fails to begin (falls through to the next career).
	if beginCareer(dice.NewScripted(5, 6), edu5, ScholarCareer) {
		t.Error("Edu 5 rolling 11 (> 5) should fail to begin the Scholar")
	}
}

// TestResolveRankAmateur: a Scholar below Edu 8 is an Amateur who cannot promote.
func TestResolveRankAmateur(t *testing.T) {
	c := &Character{scores: [count]int{7, 7, 7, 12, 6, 7}} // Edu 6, high Int
	run := &careerRun{rank: 0}
	if resolveRank(dice.NewScripted(2, 2), c, run, ScholarCareer) || run.rank != 0 {
		t.Errorf("an Amateur promoted to rank %d (Edu 6 < 8 should bar promotion)", run.rank)
	}
}

// TestAwardWinningPublication: a Publication rolled 4 under the CC counts as two
// (Book 1 p.76).
func TestAwardWinningPublication(t *testing.T) {
	c := &Character{scores: [count]int{10, 7, 7, 7, 6, 7}} // Str 10 (the term's CC), Edu 6 (Amateur, no promotion roll)
	run := &careerRun{ccPool: []Characteristic{Strength}}
	// Risk 7 survives; Reward 2D=4 <= Str 10 and 4 <= Str-4=6 -> Award-Winning; 4 skills.
	runTerm(dice.NewScripted(3, 4, 2, 2, 1, 1, 1, 1), goldenPolicy{}, c, run, ScholarCareer)
	if c.Publications != 2 {
		t.Errorf("Publications = %d, want 2 (Award-Winning doubles)", c.Publications)
	}
}

// TestAttemptTenure covers the Tenure roll (Book 1 p.76): Edu 10+ required to
// apply, then 2D <= Publications×3.
func TestAttemptTenure(t *testing.T) {
	rule := TenureRule{Rank: 3, EduMin: 10, PubsMult: 3}
	cases := []struct {
		name string
		edu  int
		pubs int
		roll []int
		want bool
	}{
		{"Edu 9 cannot apply", 9, 5, []int{2, 2}, false},
		{"Edu 10, 3 pubs (target 9), roll 4 passes", 10, 3, []int{2, 2}, true},
		{"Edu 10, 1 pub (target 3), roll 7 fails", 10, 1, []int{3, 4}, false},
		{"Edu 12, 4 pubs (target 12), roll 10 passes", 12, 4, []int{5, 5}, true},
	}
	for _, tc := range cases {
		c := &Character{scores: [count]int{7, 7, 7, 7, tc.edu, 7}, Publications: tc.pubs}
		attemptTenure(dice.NewScripted(tc.roll...), c, rule)
		if c.Tenured != tc.want {
			t.Errorf("%s: Tenured = %v, want %v", tc.name, c.Tenured, tc.want)
		}
	}
}
