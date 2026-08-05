package chargen

import (
	"slices"
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// TestMasterpieceValue pins masterpieceValue's formula (Book 1 p.75:
// Cr150,000 plus Cr10,000 per Master Point over 40, doubled at 55+) against
// literals computed independently of the function itself — a Copilot review
// on #370 correctly flagged that TestCraftsmanMasterpiece below now builds
// its expectation by CALLING masterpieceValue, so a bug in the function's own
// arithmetic needs coverage that doesn't also call it.
func TestMasterpieceValue(t *testing.T) {
	cases := []struct {
		points int
		want   int
	}{
		{40, 150_000},                   // the minimum: no points over 40
		{48, 230_000},                   // 150,000 + 8*10,000
		{54, 290_000},                   // just under Perfect
		{55, (150_000 + 15*10_000) * 2}, // Perfect threshold: doubling begins here
		{60, (150_000 + 20*10_000) * 2},
	}

	for _, tc := range cases {
		if got := masterpieceValue(tc.points); got != tc.want {
			t.Errorf("masterpieceValue(%d) = %d, want %d", tc.points, got, tc.want)
		}
	}
}

// TestGoldenCraftsman traces a two-term fresh Craftsman. From scratch the
// character cannot reach 40 Master Points, so both Masterpiece attempts fail
// (no 9D roll), each raising the Craftsman skill +1 and granting one extra
// skill. Rolls are 3,4 (= 7) unless noted. Starting scores "777777" (final UPP
// 877777 after the muster benefit).
func TestGoldenCraftsman(t *testing.T) {
	seq := []int{
		// UPP: "777777".
		3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4,
		// No qualify — Craftsman auto-begins.
		// Term 1: CC = Str. Master Points ~7 (< 40) -> failed attempt, Craftsman -> 1.
		// EligPerTerm 4 + 1 (failure) = 5 skill rolls, General col row 1 = Animals.
		1, 1, 1, 1, 1,
		1, 1, // continue vs Craftsman(1) x2 = 2: a natural 2 mandates another term
		// Term 2: CC = Dex. Master Points still < 40 -> fail, Craftsman -> 2.
		1, 1, 1, 1, 1, // Animals x5 again (now 10)
		5, 6, // continue vs Craftsman(2) x2 = 4: 11 > 4 fails, policy stops
		// Muster out: 2 rolls, Benefit column, DM +Terms (=2).
		3, // 3 + 2 = row 5 -> Str +1 (7 -> 8)
		1, // 1 + 2 = row 3 -> Wafer Jack
	}

	// goldenPolicy (scout_test.go) picks skill column 3; for the Craftsman grid
	// that column is General (Animals at row 1).
	c := GenerateCareered(
		dice.NewScripted(seq...),
		goldenPolicy{},
		worldgen.World{},
		CraftsmanCareer,
	)

	if got := c.UPP(); got != "877777" {
		t.Errorf("UPP = %q, want %q (Str 7 +1 muster benefit)", got, "877777")
	}

	if c.Masterpieces != 0 {
		t.Errorf(
			"Masterpieces = %d, want 0 (a fresh Craftsman cannot reach 40 Master Points)",
			c.Masterpieces,
		)
	}

	if c.Skills.Level("Craftsman") != 2 {
		t.Errorf(
			"Craftsman = %d, want 2 (+1 per term from failed attempts)",
			c.Skills.Level("Craftsman"),
		)
	}

	if c.Skills.Level("Animals") != 10 {
		t.Errorf("Animals = %d, want 10 (5 skills x 2 terms)", c.Skills.Level("Animals"))
	}

	rec := c.Careers[0]
	if rec.Career != Craftsman || rec.Terms != 2 || rec.Outcome != MusteredOut {
		t.Errorf("record = %+v, want Craftsman/2 terms/MusteredOut", rec)
	}
}

// TestCraftsmanMasterpiece confirms a skilled Craftsman (40+ Master Points)
// creates a Masterpiece on a passing 9D roll.
func TestCraftsmanMasterpiece(t *testing.T) {
	c := Character{scores: [count]int{8, 7, 7, 7, 7, 7}}
	c.Skills.Raise("Craftsman", 10)

	for _, s := range []string{"Pilot", "Gunner", "Computer", "Admin", "Trader"} {
		c.Skills.Raise(s, 6) // five skills at 6+
	}
	// Master Points = Str 8 + Craftsman 10 + (5 x 6) = 48. A 9D of nine 5s = 45 <= 48.
	// The term draws sixteen dice: the 9D Masterpiece attempt, then the term's
	// skill and Continue rolls. Every one is a 5.
	const points = 8 + 10 + 5*6

	runCraftsmanTerm(dice.NewScripted(slices.Repeat([]int{5}, 16)...), DefaultPolicy{}, &c, CraftsmanCareer, Strength)

	if c.Masterpieces != 1 {
		t.Errorf("Masterpieces = %d, want 1", c.Masterpieces)
	}
	// Derived from the same formula the term runner calls, not hand-traced
	// separately: a points-vs-formula mismatch on both sides would otherwise
	// pass unnoticed (#357).
	if want := masterpieceValue(points); c.MasterpieceValue != want {
		t.Errorf("MasterpieceValue = %d, want %d (masterpieceValue(%d))", c.MasterpieceValue, want, points)
	}

	if c.Skills.Level("Craftsman") != 11 {
		t.Errorf(
			"Craftsman = %d, want 11 (+1 from the successful term)",
			c.Skills.Level("Craftsman"),
		)
	}
}
