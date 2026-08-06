package chargen

import (
	"fmt"

	"github.com/philoserf/t5/internal/dice"
)

// Rank/branch machinery for the term engine (Book 1 pp.64-66, 81-86): the rank
// ladders and promotion/commission/tenure rolls, and the armed-forces Branch &
// Operations tables. Split from career.go (#331); all one package.

// A Rank is one rung of a career's rank ladder: a title and an optional skill
// granted automatically on reaching it (Book 1 p.82, "Automatic Skills by
// Rank").
type Rank struct {
	Title string
	Skill string // "" when the rank grants no automatic skill
}

// A PromotionRule is a rank-advancement roll: 2D at or under a characteristic,
// optionally raised by the character's Medal mods, or Publications. Wound Badges
// do NOT raise it — see promoted, which resolves that book conflict.
type PromotionRule struct {
	Char Characteristic
	// MedalMods raises the promotion target by the character's summed Medal mods
	// (Book 1 p.70). Wound Badges do NOT count despite three pages saying they do —
	// see promoted, which resolves that conflict against the worked example.
	MedalMods bool
	PubsMod   bool
}

// A TenureRule gates a Scholar's promotion beyond a rank until they earn Tenure
// (Book 1 p.76): with Education at EduMin, apply for Tenure each term at Rank by
// rolling 2D at or under Publications × PubsMult.
type TenureRule struct {
	Rank     int // the rank at which further promotion is gated (Scholar 3)
	EduMin   int // the Education needed to apply (10)
	PubsMult int // the tenure target is Publications × this (3)
}

// A Branch is one armed-forces branch of service: its officer Risk & Reward
// modifier and the DM it adds to Operations rolls (Book 1 pp.81-86).
type Branch struct {
	Name  string
	Mod   int
	OpsDM int
}

// A BranchOps holds an armed-forces career's Branch and Operations tables. The
// Branch is rolled once at career start (1D, +2 if Edu 10+); Operations are
// rolled four times per term (1D + the branch's OpsDM, +2 if Edu 10+), taking
// the highest mod. The combined mod (branch + best operations) makes Risk &
// Reward riskier and more rewarding — negative on the Risk roll, positive on the
// Reward roll.
//
// The Branch table can print two columns. The Soldier (p.82) and Marine (p.86)
// print one, which serves either status; the Spacer's NAVAL BRANCH (p.81)
// prints Officer and Enlisted columns that disagree on four of its eight rolls,
// so a character reads the column matching their status when they select a Branch
// (branchFor). Every character enters a career enlisted (Book 1 p.64), so the
// Enlisted column is the one a career start reads; the Officer column is reached
// at Commission (commissionBranch).
//
// WHEN Branch may change is a genuine conflict between two pages, resolved in
// favor of the general rule. The career pages say "Enlisted may select a new
// Branch upon Promotion" (p.81), but p.66 states the rule for the Armed Forces
// entire: "A non-officer character may change (reselect or reroll) Branch at the
// end of each Term. A character who receives a Commission may roll for Branch or
// keep his current Branch (for Spacers, Crew becomes Line). An Officer may not
// change Branch." The engine follows p.66 — end of every term, not only a
// promoted one — because p.81 is a career-page shorthand for a rule the general
// section spells out, and the wider trigger subsumes the narrower one.
//
// The choice itself is the policy's (Policy.RerollBranch, Policy.RerollBranchOnCommission),
// defaulting to keeping the current Branch: keeping rolls no die, so a default
// character's dice stream is exactly what it was before this rule existed.
//
// The p.66 option to *select* rather than roll a Branch is offered, and priced:
// "He must roll Soc or less to select Branch" (the p.72 worked example; the
// checklists print "Select Branch  Soc"). See selectBranch. Selection without
// that gate would hand a character their best Branch for nothing, which is why
// the gate and the option landed together.
type BranchOps struct {
	Branches [9]Branch // indexed 1-8 by the branch roll (index 0 unused)
	// EnlistedBranches is the separate Enlisted column, for a career that prints
	// one (the Spacer). Where it is nil the career prints a single Branch table
	// and Branches serves both officers and enlisted.
	EnlistedBranches *[9]Branch
	OpsMods          [10]int // indexed 1-9 by the operations roll (index 0 unused)
}

// armedForces reports whether this is an armed-forces career — one that prints a
// Branch table (Book 1 pp.81-86). It is the discriminator for the rules that page
// set carries alone: the Branch/Operations mods, the Commission ladder, and the XS
// badge a held Risk earns.
//
// It is BranchOps, not RewardKind or hasRanks. RewardKind names what a *Reward*
// earns and coincides only because Soldier/Spacer/Marine are today's only
// RewardMedal careers; hasRanks is broader still, matching the Merchant, Scholar
// and Functionary, none of which print the XS line.
//
// Value receiver, like every other Career method, despite Career being 4384 bytes:
// this inlines at all three call sites, so the copy is elided. Verified with
// `go build -gcflags=-m`.
func (c Career) armedForces() bool { return c.BranchOps != nil }

// branchFor returns the Branch for a branch roll, read from the column matching
// the character's status at the moment they select a Branch (Book 1 p.81).
func (bo *BranchOps) branchFor(officer bool, roll int) Branch {
	return bo.column(officer)[roll]
}

// commissionBranch is the Branch a newly commissioned character keeps (Book 1
// p.66: "A character who receives a Commission may roll for Branch or keep his
// current Branch (for Spacers, Crew becomes Line)"). roll is the branch roll
// that chose their enlisted Branch.
//
// Keeping a Branch across the Commission means re-reading it from the Officer
// column, and the book names the one case where the columns cannot agree: an
// enlisted Spacer's Crew has no officer counterpart, so it becomes Line. That is
// derived here rather than special-cased — the Branch is matched by name in the
// Officer column, and only a name absent from it (Crew) falls back to the
// officer entry at the same roll, which on the p.81 table is Line. A career
// printing a single Branch table returns the same Branch it already had.
func (bo *BranchOps) commissionBranch(roll int) Branch {
	if bo.EnlistedBranches == nil {
		return bo.Branches[roll]
	}

	current := bo.EnlistedBranches[roll]
	if b, ok := bo.officerBranchNamed(current.Name); ok {
		return b
	}

	return bo.Branches[roll]
}

// eduBonus is the +2 Branch/Operations die modifier for a well-educated
// character (Book 1: "DM +2 if Edu 10+").
func eduBonus(c Character) int {
	if c.Score(Education) >= 10 {
		return 2
	}

	return 0
}

// column returns the Branch table the character reads: the Enlisted column where
// the career prints one and the character is not an officer, else Branches.
func (bo *BranchOps) column(officer bool) *[9]Branch {
	if !officer && bo.EnlistedBranches != nil {
		return bo.EnlistedBranches
	}

	return &bo.Branches
}

// officerBranchNamed locates a Branch by name in the Officer column. A name on
// more than one row (Infantry holds rows 1 and 2 of the Soldier table) resolves
// to the first, which is the lowest roll that reaches it.
func (bo *BranchOps) officerBranchNamed(name string) (Branch, bool) {
	for roll := 1; roll <= 8; roll++ {
		if bo.Branches[roll].Name == name {
			return bo.Branches[roll], true
		}
	}

	return Branch{}, false
}

// selectBranch attempts the "select" half of Book 1 p.66, "the character Begins
// in a service, select or roll for Branch". Selecting is not free: the Eneri
// Dinsha example prices it — "He must roll Soc or less to select Branch (roll 10
// or less; he rolls 7) and chooses Flight" — and the p.72 checklists print the
// same gate for all three armed-forces careers as "Select Branch  Soc".
//
// It reports whether a Branch was selected. A policy that does not wish to
// select rolls no dice at all, which is why DefaultPolicy's behaviour — and every
// existing golden — is unchanged. A FAILED Soc check falls back to rolling: p.66
// offers select and roll as alternatives, so failing to select leaves the roll,
// and the character is not left without a Branch.
func selectBranch(r *dice.Roller, p Policy, c *Character, run *careerRun, bo *BranchOps) bool {
	col := bo.column(run.officer)

	// A copy, not col[1:]: that slice aliases the package-level career table, so a
	// policy that sorted it to pick the best Mod would permanently reorder the rules
	// data for every character generated afterwards.
	available := append([]Branch(nil), col[1:]...)

	idx, want := p.SelectBranch(*c, available)
	if !want {
		return false
	}

	// The index is into the slice the hook was just handed, so anything outside it
	// is a policy bug, not a rule outcome — panic rather than quietly rolling and
	// leaving the character in a Branch nobody chose. This is the same treatment
	// awardSkillsN gives an out-of-range skill column.
	if idx < 0 || idx >= len(available) {
		panic(fmt.Sprintf("chargen: SelectBranch index %d out of range 0-%d", idx, len(available)-1))
	}

	// Checked before the roll, so a buggy policy fails without first perturbing the
	// dice stream.
	if !r.Resolve(dice.Check{Dice: 2, Target: c.Score(Social)}).Success {
		return false
	}

	// branchRoll is kept in step with the Branch, since a later Commission re-reads
	// the Officer column through it.
	run.branchRoll = idx + 1
	holdBranch(run, col[run.branchRoll])

	return true
}

// chooseBranch resolves a Branch at a selection point: the character may attempt
// to select one, and otherwise — or on a failed Soc check — rolls for it.
func chooseBranch(r *dice.Roller, p Policy, c *Character, run *careerRun, career Career) {
	if !selectBranch(r, p, c, run, career.BranchOps) {
		rollBranch(r, c, run, career)
	}
}

// rollBranch rolls a Branch (1D, +2 if Edu 10+; Book 1 pp.81-86) and makes it
// the run's current Branch, read from the column matching the character's
// status. It serves both the career-start selection and a non-officer's
// end-of-term reroll.
func rollBranch(r *dice.Roller, c *Character, run *careerRun, career Career) {
	run.branchRoll = min(r.Die()+eduBonus(*c), 8)
	holdBranch(run, career.BranchOps.branchFor(run.officer, run.branchRoll))
}

// holdBranch makes b the run's current Branch.
func holdBranch(run *careerRun, b Branch) {
	run.branchMod, run.branchOpsDM, run.branchName = b.Mod, b.OpsDM, b.Name
}

// rerollBranch offers a surviving non-officer the end-of-term Branch change of
// Book 1 p.66 ("A non-officer character may change (reselect or reroll) Branch
// at the end of each Term"). An officer may not change Branch, and a career
// without a Branch table has nothing to change. Keeping — the default policy —
// rolls no die.
func rerollBranch(r *dice.Roller, p Policy, c *Character, run *careerRun, career Career, rec CareerRecord) {
	if career.BranchOps == nil || run.officer {
		return
	}

	if p.RerollBranch(*c, rec) {
		// p.66 says "change (reselect or reroll)" — both halves are offered here.
		// #298 exposed only the reroll, deliberately: without the Soc gate a policy
		// could hand a character his best Branch free, every term. The gate is what
		// makes reselection a decision rather than a giveaway, so the two land
		// together.
		chooseBranch(r, p, c, run, career)
	}
}

// changeBranchOnCommission resolves the Branch of a character who has just been
// commissioned (Book 1 p.66: "A character who receives a Commission may roll
// for Branch or keep his current Branch"). The policy chooses; the default keeps
// it, which rolls no die but still re-reads the Branch from the Officer column
// (BranchOps.commissionBranch), since that is the column an officer serves in.
func changeBranchOnCommission(r *dice.Roller, p Policy, c *Character, run *careerRun, career Career) {
	if career.BranchOps == nil {
		return
	}

	if p.RerollBranchOnCommission(*c) {
		rollBranch(r, c, run, career) // run.officer is already true: the Officer column

		return
	}

	holdBranch(run, career.BranchOps.commissionBranch(run.branchRoll))
}

// branchOpsMod returns an armed-forces term's combined Branch & Operations mod
// (Book 1 p.82): the chosen Branch's mod plus the highest of four Operations
// rolls. It is 0 (and rolls nothing) for a career without Branch/Operations.
func branchOpsMod(r *dice.Roller, c *Character, run *careerRun, career Career) int {
	if career.BranchOps == nil {
		return 0
	}

	best := 0
	for range 4 {
		best = max(
			best,
			career.BranchOps.OpsMods[min(max(r.Die()+run.branchOpsDM+eduBonus(*c), 1), 9)],
		)
	}

	return run.branchMod + best
}

// resolveRank runs a surviving ranked character's rank step (Book 1 p.64 for
// the mechanic, p.82 for the Soldier). An armed-forces enlisted character first
// rolls for Commission (success moves them to the officer track at Officer 1);
// failing that, they roll Enlisted Promotion, and an officer rolls Officer
// Promotion. A single-ladder career (no officer track, e.g. the Scholar) only
// rolls its one promotion. Promotion (not Commission) targets are raised by
// Medal mods or Publications — not Wound Badges (see promoted). Reaching a rank
// grants its auto-skill.
// It reports whether the character commissioned or promoted this term (which
// earns one extra skill, Book 1 p.82).
func resolveRank(r *dice.Roller, p Policy, c *Character, run *careerRun, career Career) bool {
	if !career.hasRanks() {
		return false
	}
	// A Scholar below the Education floor is an Amateur and cannot be promoted
	// (Book 1 p.76: "Promotion is possible only to those with Edu 8+").
	if career.amateur(c) {
		return false
	}
	// Tenure gate: at the tenure rank without Tenure, the term's advancement is a
	// Tenure application, not a promotion — promotion beyond is blocked until it
	// is earned (Book 1 p.76).
	if career.Tenure != nil && run.rank == career.Tenure.Rank && !c.Tenured {
		attemptTenure(r, c, *career.Tenure)

		return false
	}

	if run.officer {
		if run.rank < len(career.OfficerRanks) && promoted(r, *c, career.OfficerPromote) {
			promoteRank(c, run, career.OfficerRanks)

			return true
		}

		return false
	}
	// Commission applies only to careers with an officer track to rise into.
	if len(career.OfficerRanks) > 0 && promoted(r, *c, career.Commission) {
		run.officer = true
		run.rank = 1

		grantRankSkill(c, career.OfficerRanks, 1)
		changeBranchOnCommission(r, p, c, run, career)

		return true
	}

	if run.rank < len(career.EnlistedRanks) && promoted(r, *c, career.EnlistedPromote) {
		promoteRank(c, run, career.EnlistedRanks)

		return true
	}

	return false
}

// attemptTenure resolves a Scholar's Tenure application (Book 1 p.76): with
// Education at the rule's minimum, roll 2D at or under Publications × PubsMult.
// Success grants Tenure, the prerequisite for promotion beyond the tenure rank.
// A Scholar with too few Publications (target below 2) can never earn it.
func attemptTenure(r *dice.Roller, c *Character, rule TenureRule) {
	if c.Score(Education) < rule.EduMin {
		return // not yet eligible to apply
	}

	if r.Resolve(dice.Check{Dice: 2, Target: c.Publications * rule.PubsMult}).Success {
		c.Tenured = true
	}
}

// promoted resolves one promotion roll: 2D at or under the rule's characteristic,
// raised by the character's Medal mods, or Publications, when the rule allows.
//
// Wound Badges do NOT count, though three places in the book say they do. The
// Soldier/Spacer/Marine pages footnote their promotion lines "*+Medals and WB
// Mods" (pp.81/82/86) and the Master Chargen Checklists repeat it (p.72). Against
// them stand the Medals table's own footnote — "Medals (but not Wound Badges) are
// Mods for Soldier / Spacer / Marine Promotion" (p.70) — and, decisively, the only
// dice-traced example: Eneri Dinsha ends his first term holding a Wound Badge and
// an XS, and promotes against "Soc plus Medal Mods (10 +1) = 11" (p.72). The badge
// is not in that sum. A worked example beats a footnote, so the badge is excluded.
//
// WoundBadges is kept, but note it is now WRITE-ONLY in the engine: it is
// incremented on a Wounded result and printed by cmd/chargen, and nothing reads it
// mechanically. Book 1 p.91 gives it consequences this package has not implemented,
// so it is a record awaiting a consumer, not a live input.
//
// The same example fixes the mod's SIZE: his second-term promotion is "(10 +1+2)
// = 13", the first term's XS plus that term's MCUF. Medals contribute their table
// mod, not one point each — a flat count would have made it 12.
func promoted(r *dice.Roller, c Character, rule PromotionRule) bool {
	target := c.Score(rule.Char)
	if rule.MedalMods {
		target += c.MedalMods()
	}

	if rule.PubsMod {
		target += c.Publications
	}

	return r.Resolve(dice.Check{Dice: 2, Target: target}).Success
}

// grantRankSkill grants the automatic skill of a rank (Book 1 p.82), if any.
// rank is 1-based; a rank past the top of the ladder grants nothing.
func grantRankSkill(c *Character, ranks []Rank, rank int) {
	if rank >= 1 && rank <= len(ranks) && ranks[rank-1].Skill != "" {
		c.Skills.Raise(ranks[rank-1].Skill, 1)
	}
}

// promoteRank advances the character one rung up a rank ladder and grants that
// rung's automatic skill. The caller has already confirmed a rung remains.
func promoteRank(c *Character, run *careerRun, ranks []Rank) {
	run.rank++
	grantRankSkill(c, ranks, run.rank)
}
