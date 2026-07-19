package chargen

import (
	"fmt"
	"slices"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// Career resolution (Book 1 pp. 63-74). A character serves a career in four-year
// Terms; each term selects a Controlling Characteristic, resolves Risk & Reward,
// awards skills, and rolls to Continue. This file is the shared engine and its
// control flow; the per-term risk, skill, and mustering-out steps arrive in
// later slices, and per-career data (Scout first) in its own file.

// termYears is the length of one career term.
const termYears = 4

// A CareerID names a career.
type CareerID int

// The implemented careers, in checklist order.
const (
	Scout       CareerID = iota // the first implemented career
	Rogue                       // a fixed-CC career (Book 1 p. 84)
	Soldier                     // the first armed-forces (ranked) career (Book 1 p. 82)
	Marine                      // a second armed-forces career (Book 1 p. 86)
	Spacer                      // the naval armed-forces career (Book 1 p. 81)
	Agent                       // a rankless career (Book 1 p. 83)
	Citizen                     // an auto-begin career using Citizen Life (Book 1 p. 78)
	Entertainer                 // a Fame/Talent career (Book 1 p. 77)
	Craftsman                   // a Masterpiece career (Book 1 p. 75)
	Scholar                     // a single-ladder rank career with Publications (Book 1 p. 76)
	Functionary                 // an Office Politics career (Book 1 p. 87)
	Noble                       // a Return & Intrigue career (Book 1 p. 85)
	Merchant                    // a dual-track (Rating/Officer) career with Ship Shares (Book 1 p. 80)
)

// CCMode controls how a career's Controlling Characteristic is chosen each term:
// rotated through the available set (the default) or fixed for the whole career
// (Rogue).
type CCMode int

// Controlling-Characteristic selection modes for the term engine.
const (
	RotateCC CCMode = iota
	FixedCC
)

// A Qualification is a career's entry gate: roll 2D at or under the best of the
// listed characteristics (plus a modifier).
type Qualification struct {
	Chars []Characteristic
	Mod   int
}

// target returns the qualification target: the highest of the characteristics,
// plus the modifier. It panics on an empty characteristic set — a career that
// cannot be qualified for is a data error, not a 0-target impossibility to
// puzzle out later (compare RunCareer's controlling-characteristics check).
func (q Qualification) target(c Character) int {
	if len(q.Chars) == 0 {
		panic("chargen: qualification has no characteristics")
	}

	return c.Score(bestChar(c, q.Chars...)) + q.Mod
}

// A ContinueRule gives the target of a career's Continue roll — a fixed number,
// a named characteristic, or the career's Controlling Characteristic (UseCC, for
// a fixed-CC career like the Rogue) — plus a modifier.
type ContinueRule struct {
	UseChar   bool
	UseCC     bool   // target is the term's Controlling Characteristic (resolved in continues)
	UseFame   bool   // target is the character's Fame (the Entertainer)
	UseSkill  string // target is SkillMult times this skill's level (the Craftsman: Craftsman x2)
	SkillMult int
	TermsMod  bool // add the number of terms served to the target (Book 1: "Mod +Terms")
	PubsMod   bool // add the character's Publications to the target (the Scholar)
	Char      Characteristic
	Fixed     int
	Mod       int
}

// target resolves the Continue target for a character. UseCC is resolved by the
// caller (it needs the term's Controlling Characteristic), so this treats it
// like a fixed rule of value Mod.
func (rule ContinueRule) target(c Character) int {
	if rule.UseChar {
		return c.Score(rule.Char) + rule.Mod
	}

	return rule.Fixed + rule.Mod
}

// A Rank is one rung of a career's rank ladder: a title and an optional skill
// granted automatically on reaching it (Book 1 p. 82, "Automatic Skills by
// Rank").
type Rank struct {
	Title string
	Skill string // "" when the rank grants no automatic skill
}

// A PromotionRule is a rank-advancement roll: 2D at or under a characteristic,
// optionally raised by the character's Medals and Wound Badges, or Publications.
type PromotionRule struct {
	Char            Characteristic
	MedalsAndWounds bool
	PubsMod         bool
}

// A TenureRule gates a Scholar's promotion beyond a rank until they earn Tenure
// (Book 1 p.76): with Education at EduMin, apply for Tenure each term at Rank by
// rolling 2D at or under Publications × PubsMult.
type TenureRule struct {
	Rank     int // the rank at which further promotion is gated (Scholar 3)
	EduMin   int // the Education needed to apply (10)
	PubsMult int // the tenure target is Publications × this (3)
}

// A MusterDM selects the die modifier a career's Benefit muster-out column adds
// to each 1D roll (Book 1 pp. 67-70, each career page's muster DM line).
type MusterDM int

// Sources of the muster-out Benefit-column die modifier.
const (
	DMNone        MusterDM = iota // no modifier (the zero value)
	DMTerms                       // + terms served
	DMOfficerRank                 // + rank, only while on the officer track (armed forces, Merchant)
	DMRank                        // + rank on any track, ladder numbered from 1 (the Scholar)
	DMRankF0                      // + rank on a ladder numbered from 0 (the Functionary; see below)
	DMFameHalf                    // + Fame/2 (the Scout)
	DMCommends                    // + Commendations (the Agent)
)

// benefitDM returns the value of a Benefit-column muster DM for a character.
//
// A rank DM is the rank *number the career prints*, and two careers number their
// ladders differently. The Scholar begins at Scholar1 (Book 1 p.65, "Scholars
// begin with formal rank (Scholar = Scholar1)"), so its first rung is 1. The
// Functionary's ladder runs F0 Clerk … F8 Secretary (p.87), so its first rung is
// 0 — a Clerk's "+Officer Rank" muster DM is +0, not +1. run.rank is always a
// 1-based ladder index, so DMRankF0 subtracts the difference.
func benefitDM(dm MusterDM, c Character, rec CareerRecord) int {
	switch dm {
	case DMTerms:
		return rec.Terms
	case DMOfficerRank:
		if rec.Officer {
			return rec.Rank
		}

		return 0
	case DMRank:
		return rec.Rank
	case DMRankF0:
		return max(rec.Rank-1, 0)
	case DMFameHalf:
		return c.Fame / 2
	case DMCommends:
		return rec.Commendations
	default:
		return 0
	}
}

// A Branch is one armed-forces branch of service: its officer Risk & Reward
// modifier and the DM it adds to Operations rolls (Book 1 pp. 81-86).
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
// The Branch table can print two columns. The Soldier (p. 82) and Marine (p. 86)
// print one, which serves either status; the Spacer's NAVAL BRANCH (p. 81)
// prints Officer and Enlisted columns that disagree on four of its eight rolls,
// so a character reads the column matching their status when they select a Branch
// (branchFor). Every character enters a career enlisted (Book 1 p. 64), so the
// Enlisted column is the one a career start reads; the Officer column is reached
// at Commission (commissionBranch).
//
// WHEN Branch may change is a genuine conflict between two pages, resolved in
// favor of the general rule. The career pages say "Enlisted may select a new
// Branch upon Promotion" (p. 81), but p. 66 states the rule for the Armed Forces
// entire: "A non-officer character may change (reselect or reroll) Branch at the
// end of each Term. A character who receives a Commission may roll for Branch or
// keep his current Branch (for Spacers, Crew becomes Line). An Officer may not
// change Branch." The engine follows p. 66 — end of every term, not only a
// promoted one — because p. 81 is a career-page shorthand for a rule the general
// section spells out, and the wider trigger subsumes the narrower one.
//
// The choice itself is the policy's (Policy.RerollBranch, Policy.RerollBranchOnCommission),
// defaulting to keeping the current Branch: keeping rolls no die, so a default
// character's dice stream is exactly what it was before this rule existed.
//
// The p. 66 option to *select* rather than reroll a Branch is not offered. The
// book gates selection on a roll — "He must roll Soc or less to select Branch"
// (p. 66's worked example; the p. 72 checklists print "Select Branch Soc") — and
// that gate is not implemented anywhere, at career start included, where the
// engine simply rolls. Offering free selection without the gate would hand a
// character their best Branch for nothing; the gate and the option belong in one
// change.
type BranchOps struct {
	Branches [9]Branch // indexed 1-8 by the branch roll (index 0 unused)
	// EnlistedBranches is the separate Enlisted column, for a career that prints
	// one (the Spacer). Where it is nil the career prints a single Branch table
	// and Branches serves both officers and enlisted.
	EnlistedBranches *[9]Branch
	OpsMods          [10]int // indexed 1-9 by the operations roll (index 0 unused)
}

// branchFor returns the Branch for a branch roll, read from the column matching
// the character's status at the moment they select a Branch (Book 1 p. 81).
func (bo *BranchOps) branchFor(officer bool, roll int) Branch {
	if !officer && bo.EnlistedBranches != nil {
		return bo.EnlistedBranches[roll]
	}

	return bo.Branches[roll]
}

// commissionBranch is the Branch a newly commissioned character keeps (Book 1
// p. 66: "A character who receives a Commission may roll for Branch or keep his
// current Branch (for Spacers, Crew becomes Line)"). roll is the branch roll
// that chose their enlisted Branch.
//
// Keeping a Branch across the Commission means re-reading it from the Officer
// column, and the book names the one case where the columns cannot agree: an
// enlisted Spacer's Crew has no officer counterpart, so it becomes Line. That is
// derived here rather than special-cased — the Branch is matched by name in the
// Officer column, and only a name absent from it (Crew) falls back to the
// officer entry at the same roll, which on the p. 81 table is Line. A career
// printing a single Branch table returns the same Branch it already had.
func (bo *BranchOps) commissionBranch(roll int) Branch {
	if bo.EnlistedBranches == nil {
		return bo.Branches[roll]
	}

	current := bo.EnlistedBranches[roll]
	for i := 1; i < len(bo.Branches); i++ {
		if bo.Branches[i].Name == current.Name {
			return bo.Branches[i]
		}
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

// A RewardKind is the token a successful Reward roll earns for a career.
type RewardKind int

// Reward tokens a career grants on a successful Reward roll.
const (
	RewardNone         RewardKind = iota // the reward benefit is deferred (Rogue via its Scheme, …)
	RewardMedal                          // armed forces: a Medal
	RewardPublication                    // the Scholar: a Publication
	RewardShipShares                     // the Merchant: escalating Ship Shares (Nth reward = N shares)
	RewardDiscovery                      // the Scout: a Discovery — a Land Grant and Fame +1 (Book 1 p. 79)
	RewardCommendation                   // the Agent: an official Commendation (Book 1 p. 83)
)

// A Career is the data for one career. It grows as later slices add the skill
// grid, ranks, and mustering-out table.
type Career struct {
	ID               CareerID
	Name             string
	Qualify          Qualification
	CCMode           CCMode
	ControllingChars []Characteristic
	Continue         ContinueRule
	EligPerTerm      int        // number of skill rolls a surviving term grants
	BenefitDM        MusterDM   // die modifier the muster Benefit column adds (Money always adds +Terms)
	AutoBegin        bool       // the career is entered automatically, with no qualify roll (Citizen)
	CitizenLife      bool       // the term uses benign Citizen Life instead of Risk & Reward (Citizen)
	FameCareer       bool       // the term resolves Fame/Talent instead of Risk & Reward (Entertainer)
	Masterpiece      bool       // the term attempts a Masterpiece instead of Risk & Reward (Craftsman)
	OfficePolitics   bool       // the term resolves Office Politics instead of Risk & Reward (Functionary)
	ReturnIntrigue   bool       // the term resolves Return & Intrigue instead of Risk & Reward (Noble)
	ScoutDuty        bool       // the term picks Courier (no R&R, 4 skills) or Explorer (R&R, EligPerTerm skills) (Scout)
	SchemeCareer     bool       // the term masterminds a Rogue Scheme instead of Risk & Reward (Rogue)
	AutoFailOn12     bool       // a natural 12 always fails, whatever the target (Rogue; see autoFails)
	UndercoverCareer bool       // the term runs an Undercover Assignment alongside Risk & Reward (Agent)
	RewardKind       RewardKind // what a successful Reward roll earns
	Skills           SkillGrid
	MusterOut        MusterTable

	// AutoBenefits are the mustering-out awards a career grants unconditionally,
	// outside its muster-out rolls — the "Automatic:" lines some career pages
	// print under the table (Book 1 p. 87's Gold Watch). It is a function because
	// such an award can be valued from the record (100 x Terms). nil for the
	// careers whose pages print no automatic award.
	AutoBenefits func(rec CareerRecord) []Benefit

	// Rank ladders and promotion rules — set only for the armed-forces careers.
	// EnlistedRanks empty means the career has no rank (Book 1 p. 64).
	EnlistedRanks   []Rank
	OfficerRanks    []Rank
	Commission      PromotionRule // enlisted -> officer track
	EnlistedPromote PromotionRule
	OfficerPromote  PromotionRule

	BranchOps *BranchOps // armed-forces Branch/Operations R&R modifiers (nil for the rest)

	PromoteEduMin int         // minimum Education to hold rank 1+ and to promote (Scholar 8); 0 = no gate
	Tenure        *TenureRule // gates promotion beyond a rank until Tenure is earned (Scholar); nil for the rest
}

// held reports whether a career roll succeeded, applying the career's
// automatic-failure rule.
//
// The Rogue's box (Book 1 p. 84) prints three targets — "To Begin CC", "Risk &
// Reward CC*", "Continue CC*" — and then two footnotes under the block:
// "*Mod +Terms." and "But, 12 is always automatic failure." The first is already
// read as covering every starred line; the second sits at the same level and is
// unstarred, so it covers the block entire — Begin, Risk, Reward, and Continue.
// It has to: "Mod +Terms" pushes those targets past 12 with experience, and
// without the footnote a veteran Rogue would become literally unfailable.
func (c Career) held(res dice.CheckResult) bool {
	return res.Success && (!c.AutoFailOn12 || res.Roll != 12)
}

// hasRanks reports whether a career runs the rank/promotion machinery.
func (c Career) hasRanks() bool { return len(c.EnlistedRanks) > 0 }

// amateur reports whether a character falls below the career's Education floor
// for holding rank (Book 1 p.76's "Edu 8+"). An Amateur Scholar begins at rank 0
// — the sentinel below the EnlistedRanks ladder — and cannot be promoted.
func (c Career) amateur(ch *Character) bool {
	return c.PromoteEduMin > 0 && ch.Score(Education) < c.PromoteEduMin
}

// A TermOutcome is the result of a term or of a whole career. Ongoing marks a
// term the character survived; the others are how a character left a career.
type TermOutcome int

// Term outcomes from the term engine.
const (
	Ongoing     TermOutcome = iota // survived the term; the career continues
	MusteredOut                    // left the career (voluntary or failed to continue)
	Disabled                       // a career injury forced early muster-out
	Died                           // killed by a mishap or by aging
)

// String names the outcome for display.
func (o TermOutcome) String() string {
	switch o {
	case Ongoing:
		return "Ongoing"
	case MusteredOut:
		return "MusteredOut"
	case Disabled:
		return "Disabled"
	case Died:
		return "Died"
	default:
		return "Unknown"
	}
}

// A CareerRecord is the durable history of one career served.
type CareerRecord struct {
	Career        CareerID
	Terms         int
	Rank          int  // rank number within the current track (0 for a rankless career)
	Officer       bool // whether Rank is on the officer track
	Commendations int  // Commendations earned in this career (drive its muster rolls/DM)
	// Branch is the armed-forces Branch of service the character last held in this
	// career ("" for a career with no Branch table). It is recorded per career
	// rather than on the Character because a character may serve several, and it
	// is the LAST branch held rather than the first: Book 1 p. 66 lets a non-officer
	// reselect or reroll at the end of each Term, and the Commission re-reads it.
	Branch  string
	Outcome TermOutcome
}

// careerRun is the transient bookkeeping for one career, live only during
// generation (kept out of Character, like systemgen's orbit bookkeeping).
type careerRun struct {
	ccPool      []Characteristic // Controlling Characteristics not yet used this cycle
	fixed       Characteristic   // the chosen Controlling Characteristic under FixedCC
	fixedChosen bool             // whether fixed has been selected yet
	rank        int              // current rank number (1-based) for a rank career
	officer     bool             // whether rank is on the officer track
	citizenWins int              // Citizen Life successes so far (drives the Job/Hobby schedule)
	job, hobby  string           // the Citizen's Job and Hobby skills, set on the 1st/2nd success
	exiled      bool             // the Noble is currently in Exile (Book 1 p. 85)
	rewards     int              // Reward successes so far (the Merchant's escalating Ship Shares)
	commends    int              // Commendations earned this career (per-career muster rolls/DM)
	branchMod   int              // the chosen armed-forces Branch's R&R mod
	branchOpsDM int              // the chosen Branch's DM on Operations rolls
	branchName  string           // the chosen Branch's name, for the career record
	branchRoll  int              // the roll that chose the current Branch (re-read at Commission)
	terms       int              // terms served before the current one (the Rogue's "Mod +Terms")
	inPrison    bool             // the Rogue serves the coming term in prison (Book 1 p. 84)
}

// GenerateCareered generates a character on the given homeworld and runs their
// careers. It follows the checklist order (Book 1 p. 72): roll the UPP (A), take
// homeworld skills (B), optionally run the education stage (C — ED5, Trade
// School, or the College→Professors ladder), then attempt careers (D) and muster
// out (E).
//
// Each chosen career's Begin is rolled once (Book 1 p.65: "Some Careers allow
// Retry", a per-career property shown on the career's own box — no career in this
// edition grants a Begin retry, so none is offered; the code once gave the *first*
// career a retry, which p.63/p.65 do not support). A character refused by every
// chosen career falls back to the Citizen life, whose Begin is automatic — T5's
// replacement for the classic draft, so no one ends up careerless.
func GenerateCareered(r *dice.Roller, p Policy, homeworld worldgen.World, career Career) Character {
	c := Generate(r)
	c.Homeworld = homeworld
	ApplyHomeworldSkills(&c, homeworld, p)

	if p.PursueEducation(c) {
		educate(r, p, &c)
	}

	entered := serveCareer(r, p, &c, career)
	// A character may serve more than one career (Book 1), so long as they live.
	for !c.Dead {
		next, ok := p.NextCareer(c)
		if !ok {
			break
		}

		if serveCareer(r, p, &c, next) {
			entered = true
		}
	}

	if !entered && !c.Dead {
		serveCareer(r, p, &c, CitizenCareer) // fall back to the auto-begin Citizen life
	}

	computeEntitlements(&c)

	return c
}

// serveCareer attempts one career on a character: on a successful (or automatic)
// begin it runs the term loop and, unless the character died, musters out. It
// reports whether the character entered the career.
// The run is created here rather than inside the term loop because a fixed-CC
// career's Controlling Characteristic is chosen at Begin and then serves the
// whole career (Book 1 p.84) — Begin and the terms must share one run.
func serveCareer(r *dice.Roller, p Policy, c *Character, career Career) bool {
	run := newCareerRun(career)
	if !beginCareer(r, p, c, &run, career) {
		return false
	}

	runCareer(r, p, c, &run, career)

	if rec := c.Careers[len(c.Careers)-1]; rec.Outcome != Died {
		MusterOut(r, p, c, rec, career)
	}

	return true
}

// newCareerRun starts the per-career scratch state for one term loop.
func newCareerRun(career Career) careerRun {
	return careerRun{ccPool: append([]Characteristic(nil), career.ControllingChars...)}
}

// beginCareer resolves a career's Begin (Book 1 p. 63): an AutoBegin career (the
// Citizen) enters unconditionally; otherwise the character rolls 2D at or under
// the best qualifying characteristic (Book 1 p.65). The roll is made once: Begin
// retry is a per-career property ("Some Careers allow Retry", shown on the career's
// box), and no career in this edition grants one, so none is offered.
//
// A refusal costs time: "Each failed attempt (both Begin or Retry) takes one
// year" (Book 1 p.65). Only a rolled refusal does — an automatic entry makes no
// attempt to fail.
//
// A fixed-CC career Begins against its own Controlling Characteristic, not
// against a Qualification set: the Rogue's box reads "To Begin CC", and the CC
// he picks is "then used throughout his career (not just in the current Term)"
// (Book 1 p.84). Choosing it here — through the same selectCC the terms use, on
// the run they share — is what makes Begin, Risk & Reward, and Continue all read
// the one characteristic the book says they do.
func beginCareer(r *dice.Roller, p Policy, c *Character, run *careerRun, career Career) bool {
	if career.AutoBegin {
		return true
	}
	// An Education-floored career auto-admits those who meet the floor (Book 1
	// p.76: "To Begin (Edu 8+) Automatic"); a character below it must roll to
	// Begin ("To Begin Edu or Tra"), entering as an Amateur on success.
	if career.PromoteEduMin > 0 && c.Score(Education) >= career.PromoteEduMin {
		return true
	}

	var target int
	if career.CCMode == FixedCC {
		target = c.Score(selectCC(p, *c, run, career))
	} else {
		target = career.Qualify.target(*c)
	}

	if career.held(r.Resolve(dice.Check{Dice: 2, Target: target})) {
		return true
	}

	c.Age++

	return false
}

// RunCareer runs the term loop of one career on a character, appending a
// CareerRecord. It starts a fresh run, so a fixed-CC career picks its
// Controlling Characteristic on the first term; serveCareer instead picks it at
// Begin and hands the same run to runCareer (Book 1 p.84).
func RunCareer(r *dice.Roller, p Policy, c *Character, career Career) {
	run := newCareerRun(career)
	runCareer(r, p, c, &run, career)
}

// runCareer is RunCareer over a caller-owned run.
func runCareer(r *dice.Roller, p Policy, c *Character, run *careerRun, career Career) {
	if len(career.ControllingChars) == 0 && !career.FameCareer {
		panic("chargen: career " + career.Name + " has no controlling characteristics")
	}

	rec := CareerRecord{Career: career.ID}
	if career.hasRanks() {
		run.rank = 1 // armed forces begin at enlisted rank 1 (Book 1 p. 64)
		if career.amateur(c) {
			run.rank = 0 // an Amateur Scholar (Book 1 p.76): serves but cannot be promoted
		} else {
			grantRankSkill(c, career.EnlistedRanks, 1)
		}
	}

	if career.FameCareer {
		// One 2D roll opens the career (Book 1 p. 77: "roll initial Fame and Talent
		// (with one 2D roll; they are equal)"). Returning to it is a Comeback, and
		// the p.64/p.77 Fame-and-Talent table says what a Comeback does with the
		// same roll: "Comeback: Reset Fame to 2D; Talent is unchanged." A career of
		// Talent earned across earlier terms is not un-learned by walking away —
		// only the audience's memory resets.
		roll := r.Dice(2)
		if c.Talent == 0 {
			c.Talent = roll // a first Entertainer career: Fame and Talent are equal
		}

		c.Fame = roll
	}

	if career.BranchOps != nil {
		// Branch is chosen at career start — where every character is still
		// enlisted (rank R1 above), so run.officer is false and a two-column table
		// is read from its Enlisted side. A non-officer may reselect it at the end
		// of any later term (rerollBranch).
		rollBranch(r, c, run, career)
	}

	for {
		run.terms = rec.Terms // terms served before this one, for the Rogue's "Mod +Terms"
		outcome := runTerm(r, p, c, run, career)
		rec.Terms++
		c.Age += termYears
		AgingCheck(r, c) // no-op before age 34; may set c.Dead

		if outcome == Died || c.Dead {
			rec.Outcome = Died

			break
		}

		if outcome == Disabled {
			rec.Outcome = Disabled

			break
		}

		if outcome == MusteredOut {
			rec.Outcome = MusteredOut // a term forced the career to end (Office Politics job loss)

			break
		}

		// The term is over and the character survived it: a non-officer may change
		// Branch (Book 1 p. 66, "at the end of each Term"). It runs before the
		// Continue roll — the change belongs to the term that just ended, and the
		// Continue roll reads no Branch, so the order is not observable.
		rerollBranch(r, p, c, run, career, rec)

		if !continues(r, p, *c, career, rec, run) {
			rec.Outcome = MusteredOut

			break
		}
	}

	rec.Rank = run.rank
	rec.Officer = run.officer
	rec.Commendations = run.commends
	rec.Branch = run.branchName
	c.Careers = append(c.Careers, rec)
}

// resolveRiskInjury applies a failed Risk roll's consequence and reports the
// term's verdict. The CC drops by any negative (bravery) mod and the Branch/
// Operations mod, then Flux. This resolution comes before the Reward roll because
// it is the Risk roll's own outcome (Book 1 p.65, and the Eneri Dinsha example on
// p.72 applies the injury Flux before rolling Reward).
func resolveRiskInjury(
	r *dice.Roller,
	c *Character,
	cc Characteristic,
	ccVal, mod, bo int,
) TermOutcome {
	negMods := bo
	if mod < 0 {
		negMods += -mod
	}

	injury, newVal := classifyInjury(ccVal, negMods, r.Flux())
	switch injury {
	case Unharmed:
		// the Flux compensated for the mods: no injury
	case Wounded:
		c.scores[cc] = newVal
		c.WoundBadges++
	case Disabling:
		c.scores[cc] = newVal

		return Disabled
	case Fatal:
		c.scores[cc] = max(newVal, 0)
		c.Dead = true

		return Died
	}

	return Ongoing
}

// runTerm resolves one term (Book 1 p. 64). It selects the Controlling
// Characteristic, rolls Risk (2D <= CC + mod) and, on survival, Reward
// (2D <= CC - mod, mods flipped). A failed Risk injures the character; a
// surviving armed-forces character then resolves rank. It returns Ongoing,
// Disabled, or Died.
func runTerm(r *dice.Roller, p Policy, c *Character, run *careerRun, career Career) TermOutcome {
	if career.FameCareer {
		return runFameTerm(r, p, c, run, career) // no CC — the Entertainer resolves Fame/Talent
	}

	cc := selectCC(p, *c, run, career)
	if career.CitizenLife {
		return runCitizenTerm(r, p, c, run, career, cc)
	}

	if career.Masterpiece {
		return runCraftsmanTerm(r, p, c, career, cc)
	}

	if career.OfficePolitics {
		return runPoliticsTerm(r, p, c, run, career, cc)
	}

	if career.ReturnIntrigue {
		return runIntrigueTerm(r, p, c, run, career, cc)
	}

	if career.SchemeCareer {
		return runRogueTerm(r, p, c, run, career, cc)
	}

	if career.ScoutDuty && !p.ChooseExplorerDuty(*c) {
		awardSkillsN(r, p, c, career, courierElig) // Courier duty avoids Risk & Reward

		return Ongoing
	}

	ccVal := c.Score(cc)
	mod := p.RiskMod(*c, ccVal) // caution (+), bravery (-), or 0
	bo := branchOpsMod(r, c, run, career)

	// Branch & Operations mods are negative on Risk (riskier) and positive on
	// Reward (Book 1 p. 82); Caution/Bravery keep their usual signs.
	riskOK := r.Resolve(dice.Check{Dice: 2, Target: ccVal + mod - bo}).Success

	// outcome is the term's verdict, decided by the Risk roll but not returned
	// until after the Reward roll, which happens either way.
	outcome := Ongoing

	awardExemplaryService(c, career, riskOK)

	if !riskOK {
		outcome = resolveRiskInjury(r, c, cc, ccVal, mod, bo)
	}

	// Reward is rolled EVERY term, held Risk or lost — see rollReward's contract.
	if reward := r.Resolve(dice.Check{Dice: 2, Target: ccVal - mod + bo}); reward.Success && outcome != Died {
		grantReward(c, run, career, reward, ccVal)
	}

	// A character the Risk roll killed stops here; everyone else, including a
	// character it disabled, finishes the term below.
	//
	// A Disabled character serves out the term he was disabled in: Book 1 p.65 has
	// him "Muster Out at the end of the Term", which is the term COMPLETING rather
	// than aborting, so the skill eligibilities he earned by serving it are his. He
	// takes no rank roll — a promotion for a man being invalided out has no support
	// in the text, and the book states no rule either way, so the narrower reading
	// is taken and recorded here rather than inferred at each reading.
	if outcome == Died {
		return outcome
	}

	// An Agent runs an Undercover Assignment; everyone else takes their term
	// skills, with one extra on a term they commission or promote (Book 1 p. 82:
	// "1 skill because he was promoted").
	if career.UndercoverCareer {
		awardUndercover(r, p, c, career, riskOK)

		return outcome
	}

	awardSkillsN(r, p, c, career, termEligibility(r, p, c, run, career, outcome))

	return outcome
}

// termEligibility is how many skills a term grants: the career's base, plus one
// for a term the character commissioned or promoted in (Book 1 p.82, "1 skill
// because he was promoted").
//
// The rank roll happens here, so it is skipped for a Disabled character — he
// serves the term out and earns its skills, but is not promoted on his way to
// the infirmary. Skipping it also keeps his dice stream free of a roll whose
// result could never apply.
func termEligibility(
	r *dice.Roller,
	p Policy,
	c *Character,
	run *careerRun,
	career Career,
	outcome TermOutcome,
) int {
	if outcome == Ongoing && resolveRank(r, p, c, run, career) {
		return career.EligPerTerm + 1
	}

	return career.EligPerTerm
}

// A Medal is one award from the Imperial Medals table (Book 1 p.70). Mod is the
// bonus it adds to a Soldier/Spacer/Marine promotion target.
type Medal struct {
	Code string
	Name string
	Mod  int
}

// medalsTable is the p.70 Medals table, indexed by the line the book names: the
// *unmodified* successful Reward roll, +1 if the character is an Officer. 2D
// gives 2..12, so the officer bump makes 13 the top line; index 0 and 1 are unused.
//
// The Eneri Dinsha worked example (p.72) locks both ends of the lookup: a raw
// Reward roll of 3 with the Officer +1 reaches "Medals table line 4" for an XS,
// and a raw 9 with the same +1 reaches "line 10" for an MCUF.
var medalsTable = [14]Medal{
	2:  {"XS", "Exemplary Service", 1},
	3:  {"XS", "Exemplary Service", 1},
	4:  {"XS", "Exemplary Service", 1},
	5:  {"XS", "Exemplary Service", 1},
	6:  {"XS", "Exemplary Service", 1},
	7:  {"XS", "Exemplary Service", 1},
	8:  {"XS", "Exemplary Service", 1},
	9:  {"MCUF", "Meritorious Conduct Under Fire", 2},
	10: {"MCUF", "Meritorious Conduct Under Fire", 2},
	11: {"MCG", "Medal for Conspicuous Gallantry", 3},
	12: {"SEH", "Starburst for Extreme Heroism", 4},
	13: {"*SEH*", "SEH With Diamonds", 5},
}

// exemplaryService is the XS badge a held Risk earns ("Risk Success: Receive XS
// Exemplary Service Badge", Book 1 pp.82/86 and the Spacer's p.81). It is the
// table's own bottom entry, not a separate award.
var exemplaryService = medalsTable[2]

// awardMedal records one medal: the count drives the muster-out rolls, the mod
// sum drives promotion targets. They are kept apart because the book uses them
// apart — a character with one MCUF has one medal but a +2 promotion mod.
func awardMedal(c *Character, m Medal) {
	c.Medals++
	c.MedalMods += m.Mod
}

// awardExemplaryService grants the XS badge a held Risk earns: "Risk Success:
// Receive XS Exemplary Service Badge. Character is unharmed." (Book 1 pp.81/82/86,
// printed on all three armed-forces pages). Medals were awarded only on Reward
// successes, so a character who held his Risk went undercounted at every later
// promotion roll.
//
// The gate is RewardMedal because the XS is an armed-forces award — a Scholar who
// holds a Risk earns no badge. No dice are drawn: the badge follows from the Risk
// roll already made, so this does not move the stream.
func awardExemplaryService(c *Character, career Career, riskHeld bool) {
	if riskHeld && career.RewardKind == RewardMedal {
		awardMedal(c, exemplaryService)
	}
}

// medalFor reads the award off the p.70 table for a successful Reward roll.
// The roll is the raw 2D, before any Risk & Reward mods: "Rew= Successful
// unmodified Reward Roll. If Officer, increase +1".
func medalFor(rawRoll int, officer bool) Medal {
	line := rawRoll
	if officer {
		line++
	}

	return medalsTable[min(max(line, 2), 13)]
}

// grantReward awards the career's reward token for a successful Reward roll
// (Book 1 p.65). ccVal is the term's original Controlling Characteristic value,
// which the Scholar's Award-Winning threshold is measured against.
//
// The Reward is rolled every term, whether the Risk was held or lost (p.65: "The
// Character rolls for Risk ... and determines the outcome. He then rolls again for
// Reward ... and determines the consequences"). The Eneri Dinsha worked example
// (p.72) fails Risk in both of his terms — taking a Wound Badge in the first — and
// still rolls Reward and takes a Medal each time. The target keeps the ORIGINAL
// Controlling Characteristic: his second-term Reward is "10 +2 +1 -2" against his
// pre-injury Dexterity-10.
//
// The roll happens either way, so the dice stream does not depend on the Risk
// outcome — but a character the Risk roll killed collects nothing: the book has him
// "determine the consequences", and a corpse has none. Without that guard a Merchant
// killed in his term still banks Ship Shares and a Scout still records a Discovery,
// both of which feed muster-out and the character sheet.
func grantReward(
	c *Character,
	run *careerRun,
	career Career,
	reward dice.CheckResult,
	ccVal int,
) {
	switch career.RewardKind {
	case RewardNone:
		// the reward is deferred (e.g. the Rogue's Scheme); nothing to grant
	case RewardMedal:
		awardMedal(c, medalFor(reward.Roll, run.officer))
	case RewardPublication:
		c.Publications++
		if reward.Roll <= ccVal-4 {
			c.Publications++ // Award-Winning (Book 1 p.76): a Publication 4 under the CC counts as two
		}
	case RewardShipShares:
		run.rewards++ // the Nth Reward success is worth N Ship Shares
		c.ShipShares += run.rewards
	case RewardDiscovery:
		c.Discoveries++ // a valuable new world or feature: a Land Grant and Fame +1
		c.LandGrants++
		c.Fame++
	case RewardCommendation:
		c.Commendations++
		run.commends++
	}
}

// rollBranch rolls a Branch (1D, +2 if Edu 10+; Book 1 pp. 81-86) and makes it
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
// Book 1 p. 66 ("A non-officer character may change (reselect or reroll) Branch
// at the end of each Term"). An officer may not change Branch, and a career
// without a Branch table has nothing to change. Keeping — the default policy —
// rolls no die.
func rerollBranch(r *dice.Roller, p Policy, c *Character, run *careerRun, career Career, rec CareerRecord) {
	if career.BranchOps == nil || run.officer {
		return
	}

	if p.RerollBranch(*c, rec) {
		rollBranch(r, c, run, career)
	}
}

// changeBranchOnCommission resolves the Branch of a character who has just been
// commissioned (Book 1 p. 66: "A character who receives a Commission may roll
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
// (Book 1 p. 82): the chosen Branch's mod plus the highest of four Operations
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

// citizenJobs is a representative list of Citizen Job/Hobby skills; the book's
// full Citizen Skills-and-Knowledges table (Book 1 p. 78) is deferred.
var citizenJobs = []string{
	"Admin", "Broker", "Trader", "Computer", "Steward", "Liaison", "Counsellor", "Driver",
}

// runCitizenTerm resolves a Citizen's term (Book 1 p. 78). Citizen Life replaces
// Risk & Reward: a benign 2D roll at or under the Controlling Characteristic —
// success grants a Job or Hobby skill, failure does nothing (no injury). The
// character always survives the career step; only aging can end it.
func runCitizenTerm(
	r *dice.Roller,
	p Policy,
	c *Character,
	run *careerRun,
	career Career,
	cc Characteristic,
) TermOutcome {
	if c.Check(r, cc, 2, 0).Success {
		awardCitizenLife(p, c, run)
	}

	awardSkills(r, p, c, career)

	return Ongoing
}

// runFameTerm resolves an Entertainer's term (Book 1 p. 77). At the start of the
// term, events shift Fame by a Flux roll; if Fame increases the character gains
// Talent +1 and two extra skill rolls. There is no injury; only aging can end
// the career. (The optional second and third Flux rolls are deferred.)
func runFameTerm(
	r *dice.Roller,
	p Policy,
	c *Character,
	run *careerRun,
	career Career,
) TermOutcome {
	elig := career.EligPerTerm
	// Term 1 keeps the initial 2D Fame/Talent; the "+F" Flux begins at Term 2
	// (Book 1 p.77). run.terms is 0 on the first served term.
	if run.terms > 0 {
		before := c.Fame

		c.Fame = max(c.Fame+r.Flux(), 0)
		if c.Fame > before {
			c.Talent++
			elig += 2 // "If Fame Increases: 2 [skills] and Talent+1"
		}
	}

	awardSkillsN(r, p, c, career, elig)

	return Ongoing
}

// masterpieceMinimum is the fewest Master Points that allow a Masterpiece
// attempt (Book 1 p. 75).
const masterpieceMinimum = 40

// masterpieceValue returns a Masterpiece's sale value (Book 1 p. 75):
// Cr150,000 plus Cr10,000 per Master Point over 40, doubled for a Perfect
// Masterpiece (55+ Master Points).
func masterpieceValue(points int) int {
	v := 150_000 + (points-masterpieceMinimum)*10_000
	if points >= 55 {
		v *= 2
	}

	return v
}

// courierElig is the skill eligibility of a Scout's Courier duty (Book 1 p. 79);
// Explorer duty grants the career's full EligPerTerm.
const courierElig = 4

// runCraftsmanTerm resolves a Craftsman's term (Book 1 p. 75). The Craftsman
// attempts a Masterpiece instead of Risk & Reward: Master Points total the
// Controlling Characteristic, the Craftsman skill, and up to five other skills
// at level 6+. With at least 40 points a 9D roll at or under the total creates a
// Masterpiece; otherwise the attempt fails. Either way the Craftsman skill rises
// +1 (learning); success grants three extra skill rolls, failure one. There is
// no injury.
func runCraftsmanTerm(
	r *dice.Roller,
	p Policy,
	c *Character,
	career Career,
	cc Characteristic,
) TermOutcome {
	points := c.Score(
		cc,
	) + c.Skills.Level(
		"Craftsman",
	) + c.Skills.TopLevels(
		5,
		6,
		"Craftsman",
		"Language",
	)
	elig := career.EligPerTerm

	if points >= masterpieceMinimum && r.Dice(9) <= points {
		c.Masterpieces++
		c.MasterpieceValue += masterpieceValue(points)
		elig += 3
	} else {
		elig++
	}

	c.Skills.Raise("Craftsman", 1) // learning from the work, success or failure
	awardSkillsN(r, p, c, career, elig)

	return Ongoing
}

// runPoliticsTerm resolves a Functionary's term (Book 1 p. 87). Office Politics
// replaces Risk & Reward with two unmodified rolls against the Controlling
// Characteristic: a failed Risk ends the career (job loss — the character
// musters out normally, uninjured), and a successful Reward is a promotion. It
// returns MusteredOut on a lost job, otherwise Ongoing.
func runPoliticsTerm(
	r *dice.Roller,
	p Policy,
	c *Character,
	run *careerRun,
	career Career,
	cc Characteristic,
) TermOutcome {
	riskKept := c.Check(r, cc, 2, 0).Success

	elig := career.EligPerTerm
	if c.Check(r, cc, 2, 0).Success &&
		run.rank < len(career.EnlistedRanks) {
		promoteRank(c, run, career.EnlistedRanks)

		elig++ // one extra skill on promotion (Book 1 p. 82)
	}

	awardSkillsN(r, p, c, career, elig)

	if !riskKept {
		return MusteredOut
	}

	return Ongoing
}

// runIntrigueTerm resolves a Noble's term (Book 1 p. 85). Return & Intrigue
// replaces Risk & Reward, with no injury. A Noble in Exile rolls Return (success
// ends the Exile); one not in Exile rolls Intrigue (failure begins an Exile;
// success offers an Elevation — a roll-high check of 2D at or over Social
// Standing that, on success, raises Soc by 1 and awards a Land Grant).
func runIntrigueTerm(
	r *dice.Roller,
	p Policy,
	c *Character,
	run *careerRun,
	career Career,
	cc Characteristic,
) TermOutcome {
	elevated := false

	switch {
	case run.exiled:
		if c.Check(r, cc, 2, 0).Success {
			run.exiled = false // Return from Exile
		}
	case c.Check(r, cc, 2, 0).Success:
		if r.Dice(2) >= c.Score(Social) && c.scores[Social] < maxCharacteristic {
			c.scores[Social]++ // Elevation to the next Noble rank
			c.LandGrants++
			elevated = true
		}
	default:
		run.exiled = true // Exile
	}

	elig := career.EligPerTerm
	if elevated {
		elig += 2 // "When Elevated 2" (Book 1 p.85)
	}

	awardSkillsN(r, p, c, career, elig)

	return Ongoing
}

// A schemeValue is a Rogue Scheme's payoff (Book 1 p. 84): a credit value or,
// for the Scout and Merchant schemes, one Ship Share.
type schemeValue struct {
	credits int
	share   bool
}

// rogueSchemes is the Rogue Schemes table (Book 1 p. 84), indexed by Flux + 6
// (Flux runs -6..+6). A raw Flux roll reaches only -5..+5; the ±1 after-roll
// modification that extends it to the -6/+6 rows is deferred, as is a Rogue's
// option to select any previous career in place of the roll.
var rogueSchemes = [13]schemeValue{
	{credits: 200_000}, // -6 Craftsman
	{credits: 100_000}, // -5 Scholar
	{credits: 300_000}, // -4 Entertainer
	{credits: 50_000},  // -3 Citizen
	{share: true},      // -2 Scout
	{share: true},      // -1 Merchant
	{credits: 100_000}, //  0 Spacer
	{credits: 50_000},  // +1 Soldier
	{credits: 100_000}, // +2 Agent
	{credits: 100_000}, // +3 Rogue
	{credits: 500_000}, // +4 Noble
	{credits: 50_000},  // +5 Marine
	{credits: 100_000}, // +6 Functionary
}

// Rogue skill eligibility (Book 1 p. 84 B block): base Per Term 2, plus 4 for a
// Successful Scheme (Risk held) or 1 for a Failed Scheme (Risk lost); a term
// served in prison instead grants 2 In-Prison skills and nothing else.
const (
	rogueSuccessElig = 6 // Per Term 2 + Successful Scheme 4
	rogueFailElig    = 3 // Per Term 2 + Failed Scheme 1
	roguePrisonElig  = 2 // In Prison 2 (Personal or Academic column only, no Term or Scheme skills)
)

// runRogueTerm resolves a Rogue's term (Book 1 p. 84). The Rogue masterminds a
// Scheme — a plan to amass wealth at others' expense — rolled from the Rogue
// Schemes table for a Value V. Risk and Reward are two rolls against the
// Controlling Characteristic (Mod +Terms; a Risk roll of 12 always fails): a
// successful Reward pays V x (1 + CC - R + Mods), halved when the Risk failed,
// while a failed Risk sends the Rogue to prison next term and earns Fame +1
// (actually Infamy). A prison term draws only In-Prison skills (Personal or
// Academic) and no Scheme payoff.
//
// Deferred: the exact prison sentence (0-4 years at the start of the next term
// is simplified to "any positive sentence costs the following term"), the ±1
// Flux modification, and selecting a previous career in place of the roll. The
// Scheme carries no injury — only aging can end the career.
func runRogueTerm(
	r *dice.Roller,
	p Policy,
	c *Character,
	run *careerRun,
	career Career,
	cc Characteristic,
) TermOutcome {
	if run.inPrison {
		run.inPrison = false

		awardPrisonSkills(r, p, c, career, roguePrisonElig)

		return Ongoing
	}

	scheme := rogueSchemes[dice.FluxIndex(r.Flux())]
	ccVal := c.Score(cc)
	riskMod := p.RiskMod(*c, ccVal) // caution (+), bravery (-), or 0

	// "Mod +Terms": experience eases both rolls; Caution/Bravery flips sign for
	// the Reward (Book 1 p. 84, "opposite sign Mods").
	risk := r.Resolve(dice.Check{Dice: 2, Target: ccVal + riskMod + run.terms})
	riskOK := career.held(risk)
	rewardMods := -riskMod + run.terms

	reward := r.Resolve(dice.Check{Dice: 2, Target: ccVal + rewardMods})
	if career.held(reward) {
		payScheme(c, scheme, ccVal+rewardMods, reward.Roll, riskOK)
	}

	elig := rogueSuccessElig

	if !riskOK {
		c.Fame++ // Infamy
		elig = rogueFailElig

		negMods := 0
		if riskMod < 0 {
			negMods = -riskMod
		}

		if min(max(negMods+r.Flux(), 0), 4) > 0 { // a positive sentence means prison next term
			run.inPrison = true
		}
	}

	awardSkillsN(r, p, c, career, elig)

	return Ongoing
}

// payScheme applies a successful Rogue Scheme Reward (Book 1 p. 84): a Ship
// Share Scheme grants one share; a credit Scheme pays V x (1 + CC - R + Mods),
// which — since the Reward succeeded (R <= rewardTarget = CC + Mods) — is at
// least V, and is halved when the Risk failed.
func payScheme(c *Character, s schemeValue, rewardTarget, rewardRoll int, riskOK bool) {
	if s.share {
		c.ShipShares++

		return
	}

	payoff := s.credits * (1 + rewardTarget - rewardRoll)
	if !riskOK {
		payoff /= 2 // "Payoff (if any) is halved"
	}

	c.Credits += payoff
}

// awardPrisonSkills grants a Rogue's In-Prison skill rolls (Book 1 p. 84), drawn
// only from the Personal and Academic columns (grid columns 0-1); the policy's
// column choice is clamped into that range.
func awardPrisonSkills(r *dice.Roller, p Policy, c *Character, career Career, n int) {
	// In-Prison skills come from Personal (col 0) or Academic (col 1) only (Book 1
	// p. 84). Prefer Academic when the character has a Major/Minor there worth
	// raising; otherwise the always-productive Personal column (characteristic
	// bumps), so every prison roll lands on a real award.
	const personal, academic = 0, 1

	col := personal
	if productive, _ := columnScore(*c, career.Skills[academic]); productive {
		col = academic
	}

	for range n {
		applyCell(p, c, career.Skills[col][r.Die()-1])
	}
}

// undercoverAssignments maps an Agent's Undercover Assignment roll — die A
// (1-3) then die B (1-6) — to the career whose skill tables the Agent draws
// from that term (Book 1 p. 83). "Army" is the Soldier, "Navy" the Spacer; the
// Enlisted/Officer split and the specific-skill C column are flavor the Agent
// overrides by selecting the skill.
var undercoverAssignments = [4][7]CareerID{
	1: {1: Soldier, 2: Soldier, 3: Marine, 4: Marine, 5: Spacer, 6: Spacer},
	2: {1: Scholar, 2: Scholar, 3: Entertainer, 4: Entertainer, 5: Citizen, 6: Citizen},
	3: {1: Merchant, 2: Merchant, 3: Scout, 4: Scout, 5: Noble, 6: Functionary},
}

// undercoverAssignment rolls an Agent's Undercover Assignment career (Book 1 p.
// 83): die A rerolled until it is 1-3, then die B (1-6).
func undercoverAssignment(r *dice.Roller) CareerID {
	a := r.Die()
	for a > 3 {
		a = r.Die()
	}

	return undercoverAssignments[a][r.Die()]
}

// awardUndercover grants an Agent's term skills (Book 1 p. 83): the two Per-Term
// skills from the Agent's own tables, one Undercover skill selected from the
// tables of a rolled Undercover Assignment career, and — on a Successful Mission
// (a held Risk) — four more. The "select (not roll)" Undercover skill is modelled
// as the first skill of a policy-chosen column of the borrowed grid.
func awardUndercover(r *dice.Roller, p Policy, c *Character, career Career, missionOK bool) {
	borrowed := CareerByID(undercoverAssignment(r)).Skills
	col := min(max(p.ChooseSkillColumn(*c, borrowed), 0), len(borrowed)-1)
	applyCell(p, c, borrowed[col][0]) // Undercover 1: the selected skill

	elig := career.EligPerTerm
	if missionOK {
		elig += 4 // Successful Mission
	}

	awardSkillsN(r, p, c, career, elig)
}

// awardCitizenLife applies one Citizen Life success on the Job/Hobby schedule
// (Book 1 p. 78): the 1st success sets the Job at level 4, the 2nd sets the
// Hobby at level 2, and later successes alternate Job/Hobby at +1 (odd = Job,
// even = Hobby). Job and Hobby, once chosen, do not change.
func awardCitizenLife(p Policy, c *Character, run *careerRun) {
	run.citizenWins++
	switch {
	case run.citizenWins == 1:
		run.job = p.ChooseSkill(*c, citizenJobs)
		c.Skills.Raise(run.job, 4)
	case run.citizenWins == 2:
		run.hobby = p.ChooseSkill(*c, without(citizenJobs, run.job))
		c.Skills.Raise(run.hobby, 2)
	case run.citizenWins%2 == 1:
		c.Skills.Raise(run.job, 1)
	default:
		c.Skills.Raise(run.hobby, 1)
	}
}

// resolveRank runs a surviving ranked character's rank step (Book 1 p. 64 for
// the mechanic, p. 82 for the Soldier). An armed-forces enlisted character first
// rolls for Commission (success moves them to the officer track at Officer 1);
// failing that, they roll Enlisted Promotion, and an officer rolls Officer
// Promotion. A single-ladder career (no officer track, e.g. the Scholar) only
// rolls its one promotion. Promotion (not Commission) targets are raised by
// Medals, Wound Badges, or Publications. Reaching a rank grants its auto-skill.
// It reports whether the character commissioned or promoted this term (which
// earns one extra skill, Book 1 p. 82).
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
// is not in that sum. A worked example beats a footnote, so the badge is excluded
// and the field kept (it is still a muster-out and Fame input).
//
// The same example fixes the mod's SIZE: his second-term promotion is "(10 +1+2)
// = 13", the first term's XS plus that term's MCUF. Medals contribute their table
// mod, not one point each — a flat count would have made it 12.
func promoted(r *dice.Roller, c Character, rule PromotionRule) bool {
	target := c.Score(rule.Char)
	if rule.MedalsAndWounds {
		target += c.MedalMods
	}

	if rule.PubsMod {
		target += c.Publications
	}

	return r.Resolve(dice.Check{Dice: 2, Target: target}).Success
}

// grantRankSkill grants the automatic skill of a rank (Book 1 p. 82), if any.
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

// awardSkills grants the term's skill eligibility (Book 1 p. 65).
func awardSkills(r *dice.Roller, p Policy, c *Character, career Career) {
	awardSkillsN(r, p, c, career, career.EligPerTerm)
}

// awardSkillsN grants n skill rolls: for each the policy picks a column of the
// career's skill grid and 1D selects the row.
func awardSkillsN(r *dice.Roller, p Policy, c *Character, career Career, n int) {
	for range n {
		col := p.ChooseSkillColumn(*c, career.Skills)
		if col < 0 || col >= len(career.Skills) {
			panic(
				fmt.Sprintf(
					"chargen: skill column %d out of range 0-%d",
					col,
					len(career.Skills)-1,
				),
			)
		}

		applyCell(p, c, career.Skills[col][r.Die()-1])
	}
}

// applyCell applies one skill-grid cell: raise a skill (cascade skills grant a
// knowledge via the K-K-S progression), bump a characteristic (capped at the
// human maximum), or resolve a player choice among options.
// cascadeParent returns the cascade parent this cell awards a Knowledge under,
// or "" when the cell is not a cascade cell. Two shapes qualify: an AwardSkill
// carrying a Knowledge, and an AwardChoice whose Options are knowledges under a
// named parent (what the cascade() grid helper builds).
//
// It lives here, beside applyCell, because applyCell's routing and the career
// grids' cascade-cell test must agree about what a cascade cell IS. Encoding
// that twice — once in the engine, once in a test — means a third shape added
// here would silently stop being covered there.
func (cell Cell) cascadeParent() string {
	switch cell.Kind {
	case AwardSkill:
		if cell.Knowledge != "" {
			return cell.Skill
		}
	case AwardChoice:
		return cell.Skill // "" for a plain choice among flat skills
	case NoAward, AwardBump, AwardMajor, AwardMinor:
	}

	return ""
}

func applyCell(p Policy, c *Character, cell Cell) {
	switch cell.Kind {
	case NoAward:
		// an empty cell: nothing to apply
	case AwardSkill:
		if parent := cell.cascadeParent(); parent != "" {
			c.Skills.GrantCascade(parent, cell.Knowledge)
		} else {
			c.Skills.Raise(cell.Skill, 1)
		}
	case AwardBump:
		c.scores[cell.Char] = min(c.scores[cell.Char]+1, maxCharacteristic)
	case AwardChoice:
		if len(cell.Options) == 0 {
			panic("chargen: AwardChoice cell has no options")
		}

		chosen := p.ChooseSkill(*c, cell.Options)
		if parent := cell.cascadeParent(); parent != "" {
			// A cascade choice: the options are knowledges under the parent skill
			// (e.g. Language/Galanglic), granted via the K-K-S progression.
			c.Skills.GrantCascade(parent, chosen)
		} else {
			c.Skills.Raise(chosen, 1)
		}
	case AwardMajor:
		if c.Major != "" { // lost if the character has no Major (never went to college)
			c.Skills.Raise(c.Major, 1)
		}
	case AwardMinor:
		if c.Minor != "" {
			c.Skills.Raise(c.Minor, 1)
		}
	}
}

// maxCharacteristic is the human cap on a characteristic raised in play (eHex F).
const maxCharacteristic = 15

// A CellKind identifies what a skill-grid cell awards.
type CellKind int

// Skill-grid cell kinds.
const (
	NoAward     CellKind = iota // an empty cell
	AwardSkill                  // raise Skill (with Knowledge for a cascade skill)
	AwardBump                   // raise the characteristic Char
	AwardChoice                 // grant one skill the policy picks from Options
	AwardMajor                  // raise the character's College Major (lost if none)
	AwardMinor                  // raise the character's College Minor (lost if none)
)

// A Cell is one entry in a career's skill grid.
type Cell struct {
	Kind CellKind
	// Skill names the skill for AwardSkill; for AwardChoice it optionally names a
	// cascade parent, in which case Options are knowledges granted under it.
	Skill     string
	Knowledge string         // AwardSkill: the knowledge, for a cascade skill
	Char      Characteristic // AwardBump
	Options   []string       // AwardChoice: the skills (or knowledges) to pick among
}

// A SkillGrid is a career's skill table: seven columns of six rows. The column
// is chosen (see Policy.ChooseSkillColumn); the row is a 1D roll.
type SkillGrid [7][6]Cell

// A MusterColumn selects which column of a muster-out row a benefit comes from.
type MusterColumn int

// Muster-out table columns.
const (
	MoneyColumn MusterColumn = iota
	BenefitColumn
)

// A BenefitKind identifies a mustering-out award.
type BenefitKind int

// Muster-out benefit kinds.
const (
	Cash         BenefitKind = iota // Value credits
	CharBump                        // +Value to characteristic Char
	FameBump                        // +Value to Fame (the Entertainer's "Fame +1")
	Named                           // a named benefit (Ship Share, TAS Fellowship, …)
	Knighted                        // a Knighthood: raises Soc (Book 1 p.68); Name is still shown
	PensionX2                       // doubles the character's Pension (Book 1 pp.68-69)
	RetirementX2                    // doubles the character's Retirement Pay (Book 1 pp.68-69)
)

// A Benefit is one mustering-out award.
type Benefit struct {
	Kind  BenefitKind
	Value int
	Char  Characteristic
	Name  string
}

// A MusterRow is one row of a muster-out table: a Money award and a Benefit
// award; the character takes one per roll.
type MusterRow struct {
	Money   Benefit
	Benefit Benefit
}

// A MusterTable is a career's mustering-out table, indexed 1-12 by the muster
// roll (1D plus a column DM — see MusterOut).
type MusterTable [13]MusterRow // index 1-12 used

// MusterOut resolves a character's mustering-out benefits (Book 1 pp. 67-70).
// The character rolls once per term served (doubled when disabled); each roll is
// 1D plus the column DM. The Money column adds +Terms; the Benefit column adds
// the career's BenefitDM. The policy chooses the column.
func MusterOut(r *dice.Roller, p Policy, c *Character, rec CareerRecord, career Career) {
	rolls := musterRollCount(*c, rec)
	for range rolls {
		col := p.MusterColumn(*c, rec)

		fullDM := rec.Terms // Money column DM
		if col == BenefitColumn {
			fullDM = benefitDM(career.BenefitDM, *c, rec)
		}

		award := rollMusterAward(r, p, career, col, fullDM)
		// A result duplicating an unusable named benefit (a second Wafer Jack) may
		// be rerolled until different (Book 1 p.69). The cap keeps a degenerate
		// table from looping; if it is hit with the award still a duplicate, the
		// roll grants nothing rather than a useless repeat.
		for tries := 0; isDuplicateBenefit(*c, award) && tries < musterRerollCap; tries++ {
			award = rollMusterAward(r, p, career, col, fullDM)
		}

		if !isDuplicateBenefit(*c, award) {
			applyBenefit(c, award, career, rec)
		}
	}

	// The career's automatic awards land last, after the rolled ones. Order is
	// not a rule — the page prints them beside the table, not in it — but taking
	// them last keeps an automatic named award out of the rolled awards'
	// duplicate check, so adding one cannot shift a single die.
	if career.AutoBenefits != nil {
		for _, award := range career.AutoBenefits(rec) {
			applyBenefit(c, award, career, rec)
		}
	}
}

// musterRerollCap bounds duplicate-benefit rerolls (Book 1 p.69's "until a
// different benefit is received") so a table of all-identical benefits cannot
// loop forever.
const musterRerollCap = 12

// rollMusterAward rolls one mustering-out award: 1D plus a column DM, read off
// the chosen column. The DM is optional and partial (Book 1 p.68), so a policy
// that randomizes it selects any value from 0 to the full DM; otherwise the full
// DM applies.
func rollMusterAward(
	r *dice.Roller,
	p Policy,
	career Career,
	col MusterColumn,
	fullDM int,
) Benefit {
	dm := fullDM
	if fullDM > 0 && p.RandomizeMusterDM() {
		dm = r.Index(fullDM + 1)
	}

	row := min(max(r.Die()+dm, 1), 12)
	if col == BenefitColumn {
		return career.MusterOut[row].Benefit
	}

	return career.MusterOut[row].Money
}

// rerollableDuplicates are the named benefits that are useless when held twice
// (Book 1 p.69's "unwanted or unusable" examples — Wafer Jack, TAS Member — plus
// Life Insurance). Others accumulate, so a repeat is kept, not rerolled: Ship
// Share, Forbidden Knowledge, passages, and StarPass all stack, and a second
// Knighthood still grants Soc +1 (p.68, applyKnighthood), so it too is applied
// rather than rerolled. Deciding which benefits stack is a judgment the book,
// which lists Knighthood as a reroll example yet gives it a repeat effect, leaves
// genuinely open.
var rerollableDuplicates = map[string]bool{
	"Wafer Jack":     true,
	"TAS Fellowship": true,
	"Life Insurance": true,
}

// isDuplicateBenefit reports whether an award repeats a single-instance named
// benefit the character already holds. Cash, characteristic bumps, Fame, and the
// pension doublings carry no name and never count as duplicates.
func isDuplicateBenefit(c Character, b Benefit) bool {
	return rerollableDuplicates[b.Name] && slices.Contains(c.Benefits, b.Name)
}

// musterRollCount is the number of mustering-out rolls (Book 1 p.67): one per
// term served, doubled when disabled, plus one per Commendation and one if Fame
// is 19+. The book's "per MCG/SEH medal" extra rolls are deferred — the engine
// tracks a flat Medal count, not the specific top-tier medals the rule keys on.
func musterRollCount(c Character, rec CareerRecord) int {
	rolls := rec.Terms
	if rec.Outcome == Disabled {
		rolls *= 2
	}

	rolls += rec.Commendations
	if c.Fame >= 19 {
		rolls++
	}

	return rolls
}

// applyBenefit applies one mustering-out award. career and rec supply the
// context a Knighthood needs (armed-forces careers knight only officers).
func applyBenefit(c *Character, b Benefit, career Career, rec CareerRecord) {
	switch b.Kind {
	case Cash:
		c.Credits += b.Value
	case CharBump:
		// A Characteristic Improvement that would raise a characteristic above 15 is
		// lost, not clamped (Book 1 p.68: "If a benefit elevates a characteristic
		// above 15, that benefit is lost"). Clamping quietly banked the partial gain.
		if c.scores[b.Char]+b.Value <= maxCharacteristic {
			c.scores[b.Char] += b.Value
		}
	case FameBump:
		c.Fame += b.Value
	case Named:
		c.Benefits = append(c.Benefits, b.Name)
	case Knighted:
		applyKnighthood(c, career, rec)
		c.Benefits = append(c.Benefits, b.Name) // the title still shows on the sheet
	case PensionX2:
		c.pensionDoublings++
	case RetirementX2:
		c.retirementDoublings++
	}
}

// applyKnighthood raises Social Standing per Book 1 p.68: a Knighthood raises
// any Soc to B (11), or +1 if already 11+. In the armed forces (the Branch/Ops
// careers Spacer/Soldier/Marine) Knighthood is officer-only; a non-officer
// receives Soc +1 instead.
func applyKnighthood(c *Character, career Career, rec CareerRecord) {
	if career.BranchOps != nil && !rec.Officer {
		c.scores[Social] = min(c.scores[Social]+1, maxCharacteristic)

		return
	}

	if c.scores[Social] < 11 {
		c.scores[Social] = 11
	} else {
		c.scores[Social] = min(c.scores[Social]+1, maxCharacteristic)
	}
}

// An Injury classifies the result of a failed Risk roll.
type Injury int

// Injury outcomes from a failed Risk roll.
const (
	Unharmed  Injury = iota // the characteristic ends at or above its original value
	Wounded                 // reduced by 1-3 (a Wound Badge)
	Disabling               // reduced by 4+ (forced muster-out with double benefits)
	Fatal                   // reduced to 0 or below (the character dies)
)

// classifyInjury applies a failed Risk roll: the Controlling Characteristic,
// reduced by negMods (from negative mods) and shifted by Flux, is compared to
// its original value. It returns the injury and the characteristic's new value
// (unchanged when unharmed).
func classifyInjury(original, negMods, flux int) (Injury, int) {
	injured := original - negMods + flux
	switch reduction := original - injured; {
	case injured <= 0:
		return Fatal, injured
	case reduction >= 4:
		return Disabling, injured
	case reduction >= 1:
		return Wounded, injured
	default:
		return Unharmed, original
	}
}

// selectCC picks the term's Controlling Characteristic. Under RotateCC a
// characteristic cannot be reused until the whole set has been used; under
// FixedCC the policy chooses one characteristic on the first term and it serves
// the entire career.
func selectCC(p Policy, c Character, run *careerRun, career Career) Characteristic {
	if career.CCMode == FixedCC {
		if !run.fixedChosen {
			run.fixed = p.ChooseCC(c, career.ControllingChars)
			run.fixedChosen = true
		}

		return run.fixed
	}

	if len(run.ccPool) == 0 {
		run.ccPool = append(run.ccPool, career.ControllingChars...)
	}

	cc := p.ChooseCC(c, run.ccPool)
	run.ccPool = removeChar(run.ccPool, cc)

	return cc
}

// continues resolves the end-of-term Continue decision: a natural 2 forces
// another term (Mandatory Continue); otherwise the character continues only if
// the policy wishes to and the 2D Continue roll succeeds. A UseCC rule targets
// the career's fixed Controlling Characteristic (Book 1 p. 84, the Rogue).
func continues(
	r *dice.Roller,
	p Policy,
	c Character,
	career Career,
	rec CareerRecord,
	run *careerRun,
) bool {
	target := career.Continue.target(c)
	if career.Continue.UseCC {
		target = c.Score(run.fixed) + career.Continue.Mod
	}

	if career.Continue.UseFame {
		target = c.Fame + career.Continue.Mod
	}

	if career.Continue.UseSkill != "" {
		target = c.Skills.Level(
			career.Continue.UseSkill,
		)*career.Continue.SkillMult + career.Continue.Mod
	}

	if career.Continue.TermsMod {
		target += rec.Terms // more experience makes staying in easier (Book 1 p. 83, Agent)
	}

	if career.Continue.PubsMod {
		target += c.Publications // the Scholar continues more easily as they publish
	}

	res := r.Resolve(dice.Check{Dice: 2, Target: target})
	if res.Roll == 2 {
		return true // Mandatory Continue
	}

	return p.Continue(c, rec) && career.held(res)
}

// removeChar returns chars without the first occurrence of ch.
func removeChar(chars []Characteristic, ch Characteristic) []Characteristic {
	if i := slices.Index(chars, ch); i >= 0 {
		return slices.Delete(chars, i, i+1)
	}

	return chars
}
