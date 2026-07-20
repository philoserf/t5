package chargen

import (
	"slices"
	"strings"
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// TestGoldenSoldier traces a complete two-term Soldier, exercising the rank
// engine end-to-end: two medals each term (one for holding Risk, one for passing
// Reward), an enlisted promotion helped by their mods, then a Commission to the
// officer track. Rolls are 3,4 (= 7) unless
// noted. Starting scores: Str 8, Dex 7, End 8, Int 7, Edu 10, Soc 8 ("8787A8");
// the final UPP is "9787A8" after the Str +1 muster benefit.
func TestGoldenSoldier(t *testing.T) {
	seq := []int{
		// UPP: Str 8(4,4), Dex 7(3,4), End 8(4,4), Int 7(3,4), Edu 10(5,5), Soc 8(4,4).
		// Edu 10 gives the +2 Branch/Operations bonus, which is used below to reach
		// the Technical branch and Base operations so the net Branch/Ops mod is 0
		// and Risk & Reward are unchanged.
		4, 4, 3, 4, 4, 4, 3, 4, 5, 5, 4, 4,
		3, 4, // qualify vs Str 8: 7 <= 8, enters (begins Private, Fighter-1)
		5, // Branch: 5 + Edu bonus 2 = 7 -> Technical (mod 0, Ops DM 6)
		// Term 1: 4 Operations rolls, each 1 + 6 + 2 = 9 -> Base (mod 0); net mod 0.
		1, 1, 1, 1,
		3, 4, // risk 7 survive; net mod 0 -> XS badge (mods +1)
		3, 4, // reward: raw 7, enlisted -> Medals line 7 = XS (mods +2)
		5, 5, // Commission vs End 8: 10 > 8, fails
		4, 4, // Enlisted Promotion vs End 8 + Medal mods 2 = 10: 8 <= 10, promote to Corporal
		1, 1, 1, 1, 1, // 4 + 1 (promotion) skill rolls, Peacekeeper col row 1 = Admin
		3, 4, // continue vs End 8: 7, policy wants term 2
		// Term 2: Operations again (net mod 0). Risk 7 survive -> XS; Reward 7 -> XS.
		1, 1, 1, 1,
		3, 4, // risk
		3, 4, // reward: raw 7 -> XS (4 medals, mods +4)
		3, 4, // Commission vs End 8: 7 <= 8, commissioned -> 2nd Lieutenant (Leader-1)
		1, 1, 1, 1, 1, // Admin x5 (4 + 1 commission)
		3, 4, // continue: policy stops after term 2
		// Muster out: 2 rolls, Benefit column, DM +Officer Rank (=1, 2nd Lieutenant).
		1, // 1 + 1 = row 2 -> Str +1 (8 -> 9)
		2, // 2 + 1 = row 3 -> Wafer Jack
	}

	// goldenPolicy (scout_test.go) picks skill column 3; for the Soldier grid
	// that column is Peacekeeper (Admin at row 1), not the Scout's Exploration.
	c := GenerateCareered(dice.NewScripted(seq...), goldenPolicy{}, worldgen.World{}, SoldierCareer)

	if got := c.UPP(); got != "9787A8" {
		t.Errorf("UPP = %q, want %q (Str 8 +1 muster benefit, Edu 10)", got, "9787A8")
	}

	// Four medals over two terms: each term holds its Risk (an XS, Book 1 p.82
	// "Risk Success: Receive XS Exemplary Service Badge") and passes its Reward
	// (raw roll 7, enlisted, so Medals table line 7 — also an XS). All four are
	// XS, so the promotion mod is +4.
	if c.MedalCount() != 4 || c.MedalMods() != 4 {
		t.Errorf("Medals = %d mods = %d, want 4 and 4 (an XS for each Risk held and each Reward passed)",
			c.MedalCount(), c.MedalMods())
	}

	if c.Skills.Level("Fighter") != 1 || c.Skills.Level("Leader") != 1 ||
		c.Skills.Level("Admin") != 10 {
		t.Errorf("skills: Fighter=%d Leader=%d Admin=%d, want 1/1/10",
			c.Skills.Level("Fighter"), c.Skills.Level("Leader"), c.Skills.Level("Admin"))
	}

	if len(c.Careers) != 1 {
		t.Fatalf("careers = %+v, want one record", c.Careers)
	}

	rec := c.Careers[0]
	if rec.Career != Soldier || rec.Terms != 2 || rec.Outcome != MusteredOut {
		t.Errorf("record = %+v, want Soldier/2 terms/MusteredOut", rec)
	}

	// The Branch the trace above selected (roll 5 + Edu bonus 2 = 7). Recording it
	// is what makes the served branch recoverable after muster-out, and what the
	// deferred branch-keyed automatic skills will key on.
	if rec.Branch != "Technical" {
		t.Errorf("record Branch = %q, want %q", rec.Branch, "Technical")
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

func TestPromotedMedalMods(t *testing.T) {
	// One XS (+1) and one MCUF (+2): two medals, but a +3 mod. The Wound Badge is
	// deliberately present and deliberately not counted (Book 1 p.70, and the Eneri
	// Dinsha example promotes at "Soc plus Medal Mods (10 +1)" while holding one).
	c := Character{
		scores:      [count]int{7, 7, 7, 7, 7, 6},
		Medals:      []Medal{medalsTable[2], medalsTable[9]}, // an XS (+1) and an MCUF (+2)
		WoundBadges: 1,
	}
	// Soc 6 + Medal mods 3 = target 9; a roll of 8 succeeds.
	if !promoted(dice.NewScripted(4, 4), c, PromotionRule{Char: Social, MedalMods: true}) {
		t.Error("promotion with medal mods should succeed at 8 vs target 9")
	}
	// Without the mods the target is just Soc 6; 8 fails.
	if promoted(dice.NewScripted(4, 4), c, PromotionRule{Char: Social}) {
		t.Error("promotion without mods should fail at 8 vs target 6")
	}
	// The Wound Badge must not contribute: Soc 6 + mods 0, with two badges, is
	// still target 6 and 8 fails. A flat +WoundBadges model would make it 8 and pass.
	wounded := Character{scores: [count]int{7, 7, 7, 7, 7, 6}, WoundBadges: 2}
	if promoted(dice.NewScripted(4, 4), wounded, PromotionRule{Char: Social, MedalMods: true}) {
		t.Error("Wound Badges must not raise the promotion target (Book 1 p.70)")
	}
}

func TestResolveRankCommission(t *testing.T) {
	// An enlisted character whose Commission roll succeeds jumps to the officer
	// track at rank 1, gaining that rank's automatic skill.
	c := Character{scores: [count]int{7, 7, 8, 7, 7, 7}}
	run := careerRun{rank: 2}
	// Commission vs End 8: 7 <= 8. resolveRank reports the promotion (which earns
	// the term's extra skill).
	if !resolveRank(dice.NewScripted(3, 4), DefaultPolicy{}, &c, &run, SoldierCareer) {
		t.Error("resolveRank should report true on a successful commission")
	}

	if !run.officer || run.rank != 1 {
		t.Fatalf("after commission: officer %v rank %d, want officer rank 1", run.officer, run.rank)
	}

	if c.Skills.Level("Leader") != 1 {
		t.Errorf("2nd Lieutenant auto-skill Leader = %d, want 1", c.Skills.Level("Leader"))
	}
	// A failed commission and enlisted promotion report false (no extra skill).
	stuck := careerRun{rank: 2}
	// Two rolls of 12: the Commission fails, then the Enlisted Promotion.
	if resolveRank(dice.NewScripted(slices.Repeat([]int{6}, 4)...), DefaultPolicy{}, &c, &stuck, SoldierCareer) {
		t.Error("resolveRank should report false when neither commission nor promotion succeeds")
	}
}

// TestBranchOpsMod checks the combined Branch & Operations mod: a high-danger
// branch (Infantry: mod 1, low Ops DM) yields high Operations mods, while a
// Technical branch (mod 0, Ops DM 6) yields Base operations (mod 0).
func TestBranchOpsMod(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 7, 7, 7}} // Edu 7, no +2 bonus
	// Infantry branch (mod 1, Ops DM 1): four ops rolls of 1 -> index 2 -> Combat (2).
	run := careerRun{branchMod: 1, branchOpsDM: 1}
	if got := branchOpsMod(dice.NewScripted(slices.Repeat([]int{1}, 4)...), &c, &run, SoldierCareer); got != 3 {
		t.Errorf("Infantry Branch/Ops mod = %d, want 3 (branch 1 + Combat 2)", got)
	}
	// Technical branch (mod 0, Ops DM 6): ops rolls of 3 -> index 9 -> Base (0).
	tech := careerRun{branchMod: 0, branchOpsDM: 6}
	if got := branchOpsMod(dice.NewScripted(slices.Repeat([]int{3}, 4)...), &c, &tech, SoldierCareer); got != 0 {
		t.Errorf("Technical Branch/Ops mod = %d, want 0 (branch 0 + Base 0)", got)
	}
}

// selectingPolicy attempts to select a Branch by row rather than roll for one.
// row is the Branch table row it wants (1-8), which the hook reports as an index
// into the rows-1-8 slice it is handed.
type selectingPolicy struct {
	goldenPolicy

	row int
}

func (s selectingPolicy) SelectBranch(Character, []Branch) (int, bool) { return s.row - 1, true }

// TestSelectBranchSocCheck covers the "select" half of Book 1 p.66, priced by the
// Eneri Dinsha example: "He must roll Soc or less to select Branch (roll 10 or
// less; he rolls 7) and chooses Flight".
func TestSelectBranchSocCheck(t *testing.T) {
	// Soc 10, and a Soc check of 7: the selection succeeds and no Branch die is
	// drawn. Medical is row 8 of the Soldier table, so branchRoll follows it there
	// — a later Commission re-reads the Officer column through that row.
	c := Character{scores: [count]int{7, 7, 7, 7, 7, 10}}
	run := careerRun{}

	if !selectBranch(dice.NewScripted(3, 4), selectingPolicy{row: 8}, &c, &run, SoldierCareer.BranchOps) {
		t.Fatal("a Soc check of 7 against Soc 10 should select the Branch")
	}

	if run.branchName != "Medical" || run.branchRoll != 8 {
		t.Errorf("selected %q at roll %d, want Medical at 8", run.branchName, run.branchRoll)
	}
}

// TestSelectBranchFailedSocFallsBack covers the gap the book leaves: its example
// shows only a successful selection. p.66 offers select and roll as alternatives
// ("select or roll for Branch"), so a failed check leaves the roll — the
// character is never left without a Branch.
func TestSelectBranchFailedSocFallsBack(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 7, 7, 6}}
	run := careerRun{}
	// Soc 6 against a check of 12: the selection fails. chooseBranch then rolls,
	// and the script provides exactly one die for it (a 3 -> Artillery).
	chooseBranch(dice.NewScripted(6, 6, 3), selectingPolicy{row: 8}, &c, &run, SoldierCareer)

	if run.branchName != "Artillery" {
		t.Errorf("after a failed Soc check the Branch is %q, want the rolled Artillery",
			run.branchName)
	}
}

// TestSelectBranchRejectsOutOfRangeIndex guards the policy seam. An index is into
// the slice the hook was handed, so one outside it is a programming error, not a
// rule outcome — it panics rather than quietly rolling and leaving the character
// in a Branch nobody chose. The panic happens before the Soc check, so a buggy
// policy cannot perturb the dice stream on its way out.
func TestSelectBranchRejectsOutOfRangeIndex(t *testing.T) {
	defer func() {
		// Assert on the message, not merely that something panicked: an empty
		// NewScripted panics at construction, so a test that only checks "did it
		// panic" passes whether or not the bounds guard exists.
		r := recover()
		if r == nil {
			t.Fatal("an out-of-range SelectBranch index did not panic")
		}

		if msg, ok := r.(string); !ok || !strings.Contains(msg, "SelectBranch index 8") {
			t.Errorf("panicked with %v, want the SelectBranch bounds message", r)
		}
	}()

	c := Character{scores: [count]int{7, 7, 7, 7, 7, 12}}
	run := careerRun{}
	// row 9 is off the 1-8 table. Two faces are scripted, enough for a Soc check,
	// so reaching the check would NOT panic — only the bounds guard can.
	selectBranch(dice.NewScripted(3, 4), selectingPolicy{row: 9}, &c, &run, SoldierCareer.BranchOps)
}

// TestDefaultPolicyDrawsNoSocCheck is why every existing golden is undisturbed:
// a policy that does not select rolls no Soc check at all. NewScripted panics on
// exhaustion, so a stray 2D here would fail loudly.
func TestDefaultPolicyDrawsNoSocCheck(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 7, 7, 7}}
	run := careerRun{}
	// Exactly one die: the Branch roll. A Soc check would need two more.
	chooseBranch(dice.NewScripted(3), DefaultPolicy{}, &c, &run, SoldierCareer)

	if run.branchName != "Artillery" {
		t.Errorf("Branch = %q, want Artillery (rolled, unselected)", run.branchName)
	}
}

// reselectPolicy takes the p.66 end-of-term "reselect" — the half #298 left out
// because it needs the Soc gate to be a decision rather than a giveaway.
type reselectPolicy struct {
	goldenPolicy

	row int
}

func (reselectPolicy) RerollBranch(Character, CareerRecord) bool      { return true }
func (s reselectPolicy) SelectBranch(Character, []Branch) (int, bool) { return s.row - 1, true }

// TestRerollBranchTakesTheSelectPath covers the end-of-term reselect, which is
// the headline of the Soc-gate change and was reachable only through a policy
// answering true to both RerollBranch and SelectBranch — a combination no test
// built. It also pins where the Soc check lands in the stream: rerollBranch runs
// after the term and before Continue.
func TestRerollBranchTakesTheSelectPath(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 7, 7, 12}}
	run := careerRun{branchName: "Infantry"}
	// Soc 12, and a check of 7 passes, so the selection lands and no branch die is
	// drawn. Exactly two faces are scripted: a third draw would panic.
	rerollBranch(dice.NewScripted(3, 4), reselectPolicy{row: 8}, &c, &run, SoldierCareer, CareerRecord{})

	if run.branchName != "Medical" || run.branchRoll != 8 {
		t.Errorf("after reselect: %q at roll %d, want Medical at 8", run.branchName, run.branchRoll)
	}
}

// TestRerollBranchNotOfferedToOfficers is the guard the select path must not
// erode: Book 1 p.66 gives the end-of-term change to a non-officer only ("A
// non-officer character may change (reselect or reroll) Branch"). An officer must
// reach neither the Soc check nor the roll, so no dice are scripted at all.
func TestRerollBranchNotOfferedToOfficers(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 7, 7, 12}}
	run := careerRun{branchName: "Infantry", officer: true}

	// One face: too few for the Soc check (which would panic) and exactly enough
	// for a branch roll (which would change the Branch). Either wrong path is caught.
	rerollBranch(dice.NewScripted(5), reselectPolicy{row: 8}, &c, &run, SoldierCareer, CareerRecord{})

	if run.branchName != "Infantry" {
		t.Errorf("an officer's Branch changed to %q, want Infantry unchanged", run.branchName)
	}
}

// TestRiskXSRaisesTheSameTermPromotion pins an ordering the book's only traced
// example cannot: Eneri Dinsha fails his Risk in both terms, so it never shows an
// XS earned and spent in the same term. Book 1 p.65 runs a term Risk -> Reward ->
// Promotion (the p.72 narrative follows exactly that order), so a badge earned by
// holding the Risk is in hand before the promotion roll and raises its target.
// Neither armed-forces golden distinguishes it — both clear their targets either
// way — so without this test the ordering is unpinned.
func TestRiskXSRaisesTheSameTermPromotion(t *testing.T) {
	// End 8, no prior medals, and a Technical-shaped branch (Ops DM 6) so the four
	// Operations rolls land on Base and the net Branch/Ops mod is 0, as the goldens
	// arrange. The Enlisted Promotion rolls 10: it clears End 8 + both of this
	// term's XS badges (+2) but not End 8 + the Reward badge alone (+1). That is the
	// whole point — a one-point difference decides the rank.
	c := Character{scores: [count]int{8, 8, 8, 8, 8, 8}}
	run := careerRun{ccPool: []Characteristic{Endurance}, rank: 1, branchOpsDM: 6}

	outcome := runTerm(dice.NewScripted(
		3, 3, 3, 3, // 4 Operations rolls, each 3+6 = 9 -> Base (mod 0)
		3, 4, // risk 7 vs End 8: held -> XS (mods +1)
		3, 4, // reward 7, enlisted -> Medals line 7 = XS (mods +2)
		6, 6, // commission 12 vs End 8: fails
		4, 6, // enlisted promotion 10 vs End 8 + 2 = 10: promotes
		1, 1, 1, 1, 1, // 4 term skills + 1 for promoting
	), stopAfter{}, &c, &run, SoldierCareer)

	if outcome != Ongoing {
		t.Fatalf("outcome = %v, want Ongoing", outcome)
	}

	if c.MedalCount() != 2 || c.MedalMods() != 2 {
		t.Fatalf("medals = %d mods = %d, want 2 and 2 (an XS for the Risk, an XS for the Reward)",
			c.MedalCount(), c.MedalMods())
	}

	if run.rank != 2 {
		t.Errorf("rank = %d, want 2 — the term's own Risk XS raises its promotion target",
			run.rank)
	}
}
