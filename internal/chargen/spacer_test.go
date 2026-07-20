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
		3, 4, // risk survive -> XS badge (mods +1)
		3, 4, // reward: raw 7, enlisted -> Medals line 7 = XS (mods +2)
		5, 5, // Commission vs Dex 8: 10 > 8, fails
		4, 4, // Rating Promotion vs Dex 8 + Medal mods 2 = 10: 8 <= 10, promote to Able Spacer
		1, 1, 1, 1, 1, // 4 + 1 (promotion) skill rolls, Patrol/Strike col row 1 = Astrogation
		3, 4, // continue vs Str 8: 7, policy wants term 2
		// Term 2: Operations again (net 0). Risk survive -> XS; Reward 7 -> XS.
		1, 1, 1, 1,
		3, 4, // risk
		3, 4, // reward: raw 7 -> XS (4 medals, mods +4)
		3, 4, // Commission vs Dex 8: 7 <= 8, commissioned -> Ensign (Astrogation-1)
		1, 1, 1, 1, 1, // Astrogation x5 (4 + 1 commission)
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

	// Four medals over two terms: an XS for each Risk held (Book 1 pp.81/86) and an
	// XS for each Reward passed (raw roll 7, enlisted, Medals table line 7).
	if c.MedalCount() != 4 || c.MedalMods() != 4 {
		t.Errorf("Medals = %d mods = %d, want 4 and 4", c.MedalCount(), c.MedalMods())
	}
	// Astrogation: 5 (term 1 grid, 4 + 1 promotion) + 1 (Ensign auto-skill) +
	// 5 (term 2 grid, 4 + 1 commission) = 11.
	if c.Skills.Level("Fighter") != 1 || c.Skills.Level("Astrogation") != 11 {
		t.Errorf("skills: Fighter=%d Astrogation=%d, want 1/11",
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

// TestSpacerNavalBranchColumns locks both columns of the NAVAL BRANCH table
// (Book 1 p. 81), which prints "1D Officer Mod Enlisted Mod". The two columns
// disagree on rolls 1, 2, 3, and 6; every other roll is shared.
func TestSpacerNavalBranchColumns(t *testing.T) {
	for _, tc := range []struct {
		roll                    int
		officer, enlisted       string
		officerMod, enlistedMod int
	}{
		{1, "Line", "Crew", 1, 1},           // differs: name only
		{2, "Line", "Crew", 1, 1},           // differs: name only
		{3, "Line", "Engineer", 1, 0},       // differs: name and mod
		{4, "Engineer", "Engineer", 0, 0},   // agrees
		{5, "Gunnery", "Gunnery", 1, 1},     // agrees
		{6, "Flight", "Gunnery", 2, 1},      // differs: name and mod
		{7, "Technical", "Technical", 0, 0}, // agrees
		{8, "Medical", "Medical", 0, 0},     // agrees
	} {
		off := spacerBranchOps.branchFor(true, tc.roll)
		if off.Name != tc.officer || off.Mod != tc.officerMod {
			t.Errorf("branchFor(officer, %d) = %s/%d, want %s/%d",
				tc.roll, off.Name, off.Mod, tc.officer, tc.officerMod)
		}

		enl := spacerBranchOps.branchFor(false, tc.roll)
		if enl.Name != tc.enlisted || enl.Mod != tc.enlistedMod {
			t.Errorf("branchFor(enlisted, %d) = %s/%d, want %s/%d",
				tc.roll, enl.Name, enl.Mod, tc.enlisted, tc.enlistedMod)
		}
	}
}

// TestArmedForcesSingleBranchColumn guards the other two armed-forces Branch
// tables: the Soldier (p. 82) and Marine (p. 86) print a single Branch column,
// so both statuses read the same row and neither carries an Enlisted column.
func TestArmedForcesSingleBranchColumn(t *testing.T) {
	for _, career := range []Career{SoldierCareer, MarineCareer} {
		if career.BranchOps.EnlistedBranches != nil {
			t.Errorf("%s: EnlistedBranches set, but the book prints one Branch column", career.Name)
		}

		for roll := 1; roll <= 8; roll++ {
			if career.BranchOps.branchFor(true, roll) != career.BranchOps.branchFor(false, roll) {
				t.Errorf("%s: branch %d differs by status", career.Name, roll)
			}
		}
	}
}

// oneTermPolicy is goldenPolicy that leaves after a single term.
type oneTermPolicy struct{ goldenPolicy }

func (oneTermPolicy) Continue(Character, CareerRecord) bool { return false }

// spacerBranchScript builds a one-term Spacer whose only variable is the Branch
// roll. Every characteristic is 7 (so Edu is under 10 and the Branch/Operations
// +2 never applies), the four Operations rolls all land on Siege (mod 0) so the
// term's combined mod IS the Branch mod, and Risk is rolled low enough to hold
// under either column. That leaves the Reward roll — target CC + mod, so a
// bigger Branch mod makes it EASIER — as the observable: reward is scripted to
// miss the Enlisted target by exactly one, so reading the Officer column instead
// hands the character a Medal they did not earn.
func spacerBranchScript(branch int, risk, reward [2]int) []int {
	seq := []int{
		3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, // UPP: 777777
		3, 4, // Begin vs Int 7: 7 <= 7, enters (Spacehand, Fighter-1)
		branch,     // Branch (no Edu bonus, so the die is the row)
		3, 3, 3, 3, // 4 Operations rolls -> Siege (mod 0); best of four is 0
	}
	seq = append(seq, risk[0], risk[1])
	seq = append(seq, reward[0], reward[1])

	return append(
		seq,
		5, 5, // Commission vs Dex 7: 10 > 7, fails
		5, 5, // Rating Promotion vs Dex 7: 10 > 7, fails
		1, 1, 1, 1, // 4 skill rolls, Patrol/Strike col row 1 = Astrogation
		5, 5, // Continue vs Str 7: 10 > 7, and the policy leaves anyway
		1, // muster out: 1 term -> 1 roll, Benefit column, DM 0 -> row 1
	)
}

// TestSpacerEnlistedBranchRow3 covers NAVAL BRANCH row 3, where the columns
// disagree: Officer "Line 1" but Enlisted "Engineer 0". A Spacer selects their
// Branch at career start, where they are always an enlisted Rating (R1), so the
// Enlisted mod 0 applies. Reward target is CC 7 + 0 = 7 and the roll is 8: no
// Reward medal (only the XS the held Risk earns). Under the Officer column the mod would be 1, target 8, and the same
// roll would earn one.
func TestSpacerEnlistedBranchRow3(t *testing.T) {
	seq := spacerBranchScript(3, [2]int{2, 3}, [2]int{4, 4})

	c := GenerateCareered(dice.NewScripted(seq...), oneTermPolicy{}, worldgen.World{}, SpacerCareer)

	if c.WoundBadges != 0 {
		t.Fatalf("WoundBadges = %d, want 0 (Risk holds under either column)", c.WoundBadges)
	}

	// One medal, not two: the held Risk earns an XS (Book 1 p.81), and the Reward
	// misses. The Reward is still the discriminator — a SECOND medal would mean the
	// Officer column (Line mod 1) had been read, putting the target at 8.
	if c.MedalCount() != 1 {
		t.Errorf("Medals = %d, want 1 (the Risk XS alone): branch 3 is Enlisted Engineer mod 0, "+
			"so Reward is 8 vs 7 and misses; 2 Medals means the Officer column (Line mod 1) was read",
			c.MedalCount())
	}
}

// TestSpacerEnlistedBranchRow6 covers NAVAL BRANCH row 6, where the columns
// disagree: Officer "Flight 2" but Enlisted "Gunnery 1". The enlisted mod 1 puts
// the Reward target at 8 against a roll of 9: no Reward medal. The Officer column's mod
// 2 would put it at 9 and earn one.
func TestSpacerEnlistedBranchRow6(t *testing.T) {
	seq := spacerBranchScript(6, [2]int{2, 3}, [2]int{3, 6})

	c := GenerateCareered(dice.NewScripted(seq...), oneTermPolicy{}, worldgen.World{}, SpacerCareer)

	if c.WoundBadges != 0 {
		t.Fatalf("WoundBadges = %d, want 0 (Risk holds under either column)", c.WoundBadges)
	}

	// As in row 3: the held Risk earns an XS, so one medal means the Reward missed.
	if c.MedalCount() != 1 {
		t.Errorf("Medals = %d, want 1 (the Risk XS alone): branch 6 is Enlisted Gunnery mod 1, "+
			"so Reward is 9 vs 8 and misses; 2 Medals means the Officer column (Flight mod 2) was read",
			c.MedalCount())
	}
}

// TestSpacerBranchRow5Agrees is the control: row 5 is Gunnery mod 1 in BOTH
// columns, so the same script that distinguishes rows 3 and 6 must behave
// identically here whichever column is read — a Reward of 9 against target 8
// misses. It guards against a "fix" that shifts every row rather than the
// three that actually differ.
func TestSpacerBranchRow5Agrees(t *testing.T) {
	seq := spacerBranchScript(5, [2]int{2, 3}, [2]int{3, 6})

	c := GenerateCareered(dice.NewScripted(seq...), oneTermPolicy{}, worldgen.World{}, SpacerCareer)

	if c.MedalCount() != 1 {
		t.Errorf("Medals = %d, want 1 — the Risk XS alone (branch 5 is Gunnery mod 1 in both columns)",
			c.MedalCount())
	}
}

// TestMedalsTableEneri locks the Medals table lookup to the book's only
// dice-traced example, "Eneri Dinsha 9AB58A Serves The Empire" (Book 1 p.72).
// It is the fixture that decides both halves of the medal model, so it asserts
// the book's own arithmetic rather than the engine's.
func TestMedalsTableEneri(t *testing.T) {
	// Term 1: "He applies the raw Reward roll (3) (and Officer Mod +1) to Medals
	// table line 4 and receives an XS Exemplary Service."
	if got := medalFor(3, true); got.Code != "XS" || got.Mod != 1 {
		t.Errorf("medalFor(3, officer) = %+v, want XS +1 (p.72, table line 4)", got)
	}

	// Term 2: "He applies the raw Reward roll (9) Mod +1 (he is an Officer) to the
	// Medals table line 10 and receives MCUF Meritorious Conduct Under Fire."
	if got := medalFor(9, true); got.Code != "MCUF" || got.Mod != 2 {
		t.Errorf("medalFor(9, officer) = %+v, want MCUF +2 (p.72, table line 10)", got)
	}

	// The promotion targets those two medals produce. Eneri is Soc 10 throughout.
	// Term 1: "Soc plus Medal Mods (10 +1) = 11" — one XS, and the Wound Badge he
	// took that same term is NOT in the sum.
	eneri := Character{scores: [count]int{9, 10, 11, 5, 8, 10}, WoundBadges: 1}
	awardMedal(&eneri, medalFor(3, true))

	if got := eneri.Score(Social) + eneri.MedalMods(); got != 11 {
		t.Errorf("term-1 promotion target = %d, want 11 (Soc 10 +1; the Wound Badge does not count)", got)
	}

	// Term 2: "Soc plus Medal Mods (10 +1+2) = 13" — the first term's XS plus this
	// term's MCUF. A flat one-point-per-medal model yields 12 and fails here.
	awardMedal(&eneri, medalFor(9, true))

	if got := eneri.Score(Social) + eneri.MedalMods(); got != 13 {
		t.Errorf("term-2 promotion target = %d, want 13 (Soc 10 +1+2)", got)
	}

	if eneri.MedalCount() != 2 {
		t.Errorf("Medals = %d, want 2 (two awards, worth +3 between them)", eneri.MedalCount())
	}
}
