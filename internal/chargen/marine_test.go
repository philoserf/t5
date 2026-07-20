package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// TestGoldenMarine traces a complete two-term Marine, confirming the rank engine
// generalizes to a second armed-forces career whose promotion characteristics
// differ from the Soldier's (Enlisted Promotion vs Str, Officer Promotion vs
// Int). A medal each term, an enlisted promotion, then a Commission to 2nd
// Lieutenant. Rolls are 3,4 (= 7) unless noted. Starting scores: Str 8, Dex 7,
// End 8, Int 8, Edu 10, Soc 7 ("8788A7"); the final UPP is "9788A7" after the
// Str +1 muster benefit.
func TestGoldenMarine(t *testing.T) {
	seq := []int{
		// UPP: Str 8, Dex 7, End 8, Int 8, Edu 10(5,5), Soc 7. Edu 10 gives the +2
		// Branch/Operations bonus, used to reach the Technical branch and Garrison
		// operations so the net Branch/Ops mod is 0 and Risk & Reward are unchanged.
		4, 4, 3, 4, 4, 4, 4, 4, 5, 5, 3, 4,
		3, 4, // qualify vs Str 8: 7 <= 8, enters (Private, Fighter-1)
		5, // Branch: 5 + 2 = 7 -> Technical (mod 0, Ops DM 6)
		// Term 1: 4 Operations rolls, each 1 + 6 + 2 = 9 -> Garrison (mod 0); net 0.
		1, 1, 1, 1,
		3, 4, // risk survive -> XS badge (mods +1)
		3, 4, // reward: raw 7, enlisted -> Medals line 7 = XS (mods +2)
		5, 5, // Commission vs End 8: 10 > 8, fails
		4, 4, // Enlisted Promotion vs Str 8 + Medal mods 2 = 10: 8 <= 10, promote to Lance Corporal
		1, 1, 1, 1, 1, // 4 + 1 (promotion) skill rolls, Peacekeeper col row 1 = Vacc Suit
		3, 4, // continue vs Str 8: 7, policy wants term 2
		// Term 2: Operations again (net 0). Risk survive -> XS; Reward 7 -> XS.
		1, 1, 1, 1,
		3, 4, // risk
		3, 4, // reward: raw 7 -> XS (4 medals, mods +4)
		3, 4, // Commission vs End 8: 7 <= 8, commissioned -> 2nd Lieutenant (Leader-1)
		1, 1, 1, 1, 1, // Vacc Suit x5 (4 + 1 commission)
		3, 4, // continue: policy stops after term 2
		// Muster out: 2 rolls, Benefit column, DM +Officer Rank (=1, 2nd Lieutenant).
		1, // 1 + 1 = row 2 -> Str +1 (8 -> 9)
		2, // 2 + 1 = row 3 -> Wafer Jack
	}

	// goldenPolicy (scout_test.go) picks skill column 3; for the Marine grid that
	// column is Peacekeeper (Vacc Suit at row 1).
	c := GenerateCareered(dice.NewScripted(seq...), goldenPolicy{}, worldgen.World{}, MarineCareer)

	if got := c.UPP(); got != "9788A7" {
		t.Errorf("UPP = %q, want %q (Str 8 +1 muster benefit, Edu 10)", got, "9788A7")
	}

	// Four medals over two terms: an XS for each Risk held (Book 1 pp.81/86) and an
	// XS for each Reward passed (raw roll 7, enlisted, Medals table line 7).
	if c.MedalCount() != 4 || c.MedalMods() != 4 {
		t.Errorf("Medals = %d mods = %d, want 4 and 4", c.MedalCount(), c.MedalMods())
	}

	if c.Skills.Level("Fighter") != 1 || c.Skills.Level("Leader") != 1 ||
		c.Skills.Level("Vacc Suit") != 10 {
		t.Errorf("skills: Fighter=%d Leader=%d Vacc Suit=%d, want 1/1/10",
			c.Skills.Level("Fighter"), c.Skills.Level("Leader"), c.Skills.Level("Vacc Suit"))
	}

	rec := c.Careers[0]
	if rec.Career != Marine || rec.Terms != 2 || rec.Outcome != MusteredOut {
		t.Errorf("record = %+v, want Marine/2 terms/MusteredOut", rec)
	}

	if !rec.Officer || rec.Rank != 1 {
		t.Errorf(
			"rank = %d officer %v, want officer rank 1 (2nd Lieutenant)",
			rec.Rank,
			rec.Officer,
		)
	}

	if len(c.Benefits) != 1 || c.Benefits[0] != "Wafer Jack" {
		t.Errorf("Benefits = %v, want [Wafer Jack]", c.Benefits)
	}
}
