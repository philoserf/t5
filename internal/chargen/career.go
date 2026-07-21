package chargen

import (
	"slices"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// Career resolution (Book 1 pp. 63-74): the term-engine spine. A character serves
// a career in four-year Terms; each term selects a Controlling Characteristic,
// resolves Risk & Reward, awards skills, and rolls to Continue. The per-variant
// terms live in variant.go, the rank/branch machinery in branch.go, mid-career
// awards in reward.go, and mustering-out in muster.go — all one package (#331).

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

// A TermVariant selects how a career resolves each of its terms (Book 1 careers).
// Most careers run the Standard Risk & Reward term; a handful replace it wholesale
// or run something alongside it, and this one field names which. It replaces eight
// exclusive-by-convention bools (#331): a career could set two of those and
// silently run whichever the dispatch ladder reached first, but it can hold only
// one TermVariant, so two variants are now unrepresentable rather than ordered.
type TermVariant int

// The term variants. StandardTerm is the zero value, so a career that names none
// runs Risk & Reward (Soldier, Marine, Spacer, Scholar, Merchant).
const (
	StandardTerm   TermVariant = iota // Risk & Reward
	FameTerm                          // Fame/Talent instead of R&R, and no Controlling Characteristic (Entertainer)
	CitizenTerm                       // benign Citizen Life instead of R&R (Citizen)
	CraftsmanTerm                     // a Masterpiece attempt instead of R&R (Craftsman)
	PoliticsTerm                      // Office Politics instead of R&R (Functionary)
	IntrigueTerm                      // Return & Intrigue instead of R&R (Noble)
	RogueTerm                         // a Rogue Scheme instead of R&R (Rogue)
	ScoutTerm                         // Courier (no R&R) or Explorer (R&R), the Scout's per-term choice
	UndercoverTerm                    // an Undercover Assignment alongside R&R (Agent)
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
	EligPerTerm      int         // number of skill rolls a surviving term grants
	BenefitDM        MusterDM    // die modifier the muster Benefit column adds (Money always adds +Terms)
	Term             TermVariant // how each term resolves; StandardTerm (Risk & Reward) unless a career varies it
	AutoBegin        bool        // the career is entered automatically, with no qualify roll (Citizen)
	AutoFailOn12     bool        // a natural 12 always fails, whatever the target (Rogue; see autoFails)
	RewardKind       RewardKind  // what a successful Reward roll earns
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
	if len(career.ControllingChars) == 0 && career.Term != FameTerm {
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

	if career.Term == FameTerm {
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
		// of any later term (rerollBranch). The character may SELECT rather than
		// roll, at the price of a Soc check (Book 1 p.66); chooseBranch offers that
		// and rolls when it is declined or failed.
		chooseBranch(r, p, c, run, career)
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

		// Keep the record's Branch in step with the run before any policy hook reads
		// it. rec is passed by value to RerollBranch and Continue, and assigning it
		// only at muster-out (below) left both seeing "" for the whole career — so a
		// policy meaning "keep Flight" rerolled every term, drawing an extra die and
		// landing the character somewhere he never chose.
		rec.Branch = run.branchName

		// The term is over and the character survived it: a non-officer may change
		// Branch (Book 1 p. 66, "at the end of each Term"). It runs before the
		// Continue roll — the change belongs to the term that just ended, and the
		// Continue roll reads no Branch, so the order is not observable.
		rerollBranch(r, p, c, run, career, rec)

		rec.Branch = run.branchName // rerollBranch may have just changed it

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
//
//nolint:cyclop // the term engine dispatches every career variant; irreducibly branchy
func runTerm(r *dice.Roller, p Policy, c *Character, run *careerRun, career Career) TermOutcome {
	if career.Term == FameTerm {
		return runFameTerm(r, p, c, run, career) // no CC — the Entertainer resolves Fame/Talent
	}

	cc := selectCC(p, *c, run, career)
	if career.Term == CitizenTerm {
		return runCitizenTerm(r, p, c, run, career, cc)
	}

	if career.Term == CraftsmanTerm {
		return runCraftsmanTerm(r, p, c, career, cc)
	}

	if career.Term == PoliticsTerm {
		return runPoliticsTerm(r, p, c, run, career, cc)
	}

	if career.Term == IntrigueTerm {
		return runIntrigueTerm(r, p, c, run, career, cc)
	}

	if career.Term == RogueTerm {
		return runRogueTerm(r, p, c, run, career, cc)
	}

	if career.Term == ScoutTerm && !p.ChooseExplorerDuty(*c) {
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

	if riskOK {
		// "Risk Success: Receive XS Exemplary Service Badge. Character is unharmed."
		// (Book 1 pp.81/82/86, printed on all three armed-forces pages.) Medals were
		// awarded only on Reward successes, so a character who held his Risk went
		// undercounted at every later promotion roll. No dice are drawn: the badge
		// follows from the Risk roll already made.
		if career.armedForces() {
			awardMedal(c, exemplaryService)
		}
	} else {
		outcome = resolveRiskInjury(r, c, cc, ccVal, mod, bo)
	}

	// Reward is rolled EVERY term, held Risk or lost — see grantReward's contract.
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
	if career.Term == UndercoverTerm {
		awardUndercover(r, p, c, career, riskOK)

		return outcome
	}

	// The rank roll is skipped for a Disabled character: he serves the term out and
	// earns its skills, but is not promoted on his way to the infirmary. Skipping it
	// also keeps his stream free of a draw whose result could never apply.
	elig := career.EligPerTerm
	if outcome == Ongoing && resolveRank(r, p, c, run, career) {
		elig++ // Book 1 p.82, "1 skill because he was promoted"
	}

	awardSkillsN(r, p, c, career, elig)

	return outcome
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
