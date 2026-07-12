package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// TestGoldenMerchant traces a two-term Merchant: standard Risk & Reward where
// each Reward success is escalating Ship Shares (1 then 2 = 3 total), a Rating
// promotion (Temp -> Spacehand), then a Commission to Fourth Officer. Rolls are
// 3,4 (= 7) unless noted. Starting scores "887877" (final UPP 888877).
func TestGoldenMerchant(t *testing.T) {
	seq := []int{
		// UPP: Str 8, Dex 8, End 7, Int 8, Edu 7, Soc 7.
		4, 4, 4, 4, 3, 4, 4, 4, 3, 4, 3, 4,
		// No qualify — Merchant auto-begins as Temp (rank 1).
		// Term 1: CC = Str (8). Risk 7 survives; Reward 7 -> 1 Ship Share.
		3, 4, // risk
		3, 4, // reward -> Ship Shares 1
		5, 5, // Commission vs Int 8: 10 > 8, fails
		4, 4, // Rating Promotion vs Dex 8: 8 <= 8, promote to Spacehand (rank 2)
		1, 1, 1, 1, // 4 skill rolls, Trade col row 1 = Broker
		3, 4, // continue vs Str 8: 7, policy wants term 2
		// Term 2: CC = Dex (8). Reward 7 -> 2 more Ship Shares (total 3).
		3, 4, // risk
		3, 4, // reward -> Ship Shares +2 = 3
		3, 4, // Commission vs Int 8: 7 <= 8, commissioned -> Fourth Officer (Steward-1)
		1, 1, 1, 1, // Broker x4 (now 8)
		3, 4, // continue: policy stops after term 2
		// Muster out: 2 rolls, Benefit column, DM +Terms (=2).
		3, // 3 + 2 = row 5 -> End +1 (7 -> 8)
		1, // 1 + 2 = row 3 -> Wafer Jack
	}

	// goldenPolicy (scout_test.go) picks skill column 3; for the Merchant grid
	// that column is Trade (Broker at row 1).
	c := GenerateCareered(dice.NewScripted(seq...), goldenPolicy{}, worldgen.World{}, MerchantCareer)

	if got := c.UPP(); got != "888877" {
		t.Errorf("UPP = %q, want %q (End 7 +1 muster benefit)", got, "888877")
	}
	if c.ShipShares != 3 {
		t.Errorf("ShipShares = %d, want 3 (escalating: 1 + 2)", c.ShipShares)
	}
	if c.Skills.Level("Broker") != 8 || c.Skills.Level("Steward") != 1 {
		t.Errorf("skills: Broker=%d Steward=%d, want 8/1", c.Skills.Level("Broker"), c.Skills.Level("Steward"))
	}
	rec := c.Careers[0]
	if rec.Career != Merchant || rec.Terms != 2 || rec.Outcome != MusteredOut {
		t.Errorf("record = %+v, want Merchant/2 terms/MusteredOut", rec)
	}
	if !rec.Officer || rec.Rank != 1 {
		t.Errorf("rank = %d officer %v, want officer rank 1 (Fourth Officer)", rec.Rank, rec.Officer)
	}
}
