package chargen

import (
	"fmt"

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
)

// CCMode controls how a career's Controlling Characteristic is chosen each term:
// rotated through the available set (the default) or fixed for the whole career
// (Rogue).
type CCMode int

const (
	RotateCC CCMode = iota
	FixedCC
)

// AdvanceRule controls how rank advancement rolls compare: roll-low (the T5
// default) or roll-high (Noble elevation).
type AdvanceRule int

const (
	RollLow AdvanceRule = iota
	RollHigh
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
	best := c.Score(q.Chars[0])
	for _, ch := range q.Chars[1:] {
		best = max(best, c.Score(ch))
	}
	return best + q.Mod
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
// optionally raised by the character's Medals and Wound Badges.
type PromotionRule struct {
	Char            Characteristic
	MedalsAndWounds bool
}

// A Career is the data for one career. It grows as later slices add the skill
// grid, ranks, and mustering-out table.
type Career struct {
	ID               CareerID
	Name             string
	Qualify          Qualification
	CCMode           CCMode
	ControllingChars []Characteristic
	Continue         ContinueRule
	Advance          AdvanceRule
	EligPerTerm      int  // number of skill rolls a surviving term grants
	MusterBenefitDMT bool // muster Benefit column adds +Terms (else +Fame/2, currently 0)
	AutoBegin        bool // the career is entered automatically, with no qualify roll (Citizen)
	CitizenLife      bool // the term uses benign Citizen Life instead of Risk & Reward (Citizen)
	FameCareer       bool // the term resolves Fame/Talent instead of Risk & Reward (Entertainer)
	Masterpiece      bool // the term attempts a Masterpiece instead of Risk & Reward (Craftsman)
	Skills           SkillGrid
	MusterOut        MusterTable

	// Rank ladders and promotion rules — set only for the armed-forces careers.
	// EnlistedRanks empty means the career has no rank (Book 1 p. 64).
	EnlistedRanks   []Rank
	OfficerRanks    []Rank
	Commission      PromotionRule // enlisted -> officer track
	EnlistedPromote PromotionRule
	OfficerPromote  PromotionRule
}

// hasRanks reports whether a career runs the rank/promotion machinery.
func (c Career) hasRanks() bool { return len(c.EnlistedRanks) > 0 }

// A TermOutcome is the result of a term or of a whole career. Ongoing marks a
// term the character survived; the others are how a character left a career.
type TermOutcome int

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
	Career  CareerID
	Terms   int
	Rank    int  // rank number within the current track (0 for a rankless career)
	Officer bool // whether Rank is on the officer track
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
}

// GenerateCareered generates a character on the given homeworld and runs one
// career on them. It follows the checklist order (Book 1 p. 72): roll the UPP
// (A), take homeworld skills (B), optionally run the education stage (C — ED5,
// College, or University), then attempt the career (D) and muster out (E). The character qualifies on 2D at or under
// the best of the career's qualifying characteristics; on success they serve
// terms and muster out, and on failure they enter no career and remain a fresh
// 18-year-old — but keep their homeworld skills and education either way.
func GenerateCareered(r *dice.Roller, p Policy, homeworld worldgen.World, career Career) Character {
	c := Generate(r)
	c.Homeworld = homeworld
	ApplyHomeworldSkills(&c, homeworld, p)
	if p.PursueEducation(c) {
		educate(r, p, &c)
	}
	// AutoBegin (the Citizen) enters with no qualify roll; the || short-circuits
	// so Qualify.target is never evaluated for such a career.
	if career.AutoBegin || r.Resolve(dice.Check{Dice: 2, Target: career.Qualify.target(c)}).Success {
		RunCareer(r, p, &c, career)
		if rec := c.Careers[len(c.Careers)-1]; rec.Outcome != Died {
			MusterOut(r, p, &c, rec, career)
		}
	}
	return c
}

// RunCareer runs the term loop of one career on a character, appending a
// CareerRecord.
func RunCareer(r *dice.Roller, p Policy, c *Character, career Career) {
	if len(career.ControllingChars) == 0 && !career.FameCareer {
		panic("chargen: career " + career.Name + " has no controlling characteristics")
	}
	run := careerRun{ccPool: append([]Characteristic(nil), career.ControllingChars...)}
	rec := CareerRecord{Career: career.ID}
	if career.hasRanks() {
		run.rank = 1 // armed forces begin at enlisted rank 1 (Book 1 p. 64)
		grantRankSkill(c, career.EnlistedRanks, 1)
	}
	if career.FameCareer {
		c.Fame = r.Dice(2) // initial Fame and Talent are one 2D roll (Book 1 p. 77)
		c.Talent = c.Fame
	}

	for {
		outcome := runTerm(r, p, c, &run, career)
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
		if !continues(r, p, *c, career, rec, &run) {
			rec.Outcome = MusteredOut
			break
		}
	}
	rec.Rank = run.rank
	rec.Officer = run.officer
	c.Careers = append(c.Careers, rec)
}

// runTerm resolves one term (Book 1 p. 64). It selects the Controlling
// Characteristic, rolls Risk (2D <= CC + mod) and, on survival, Reward
// (2D <= CC - mod, mods flipped). A failed Risk injures the character; a
// surviving armed-forces character then resolves rank. It returns Ongoing,
// Disabled, or Died.
func runTerm(r *dice.Roller, p Policy, c *Character, run *careerRun, career Career) TermOutcome {
	if career.FameCareer {
		return runFameTerm(r, p, c, career) // no CC — the Entertainer resolves Fame/Talent
	}
	cc := selectCC(p, *c, run, career)
	if career.CitizenLife {
		return runCitizenTerm(r, p, c, run, career, cc)
	}
	if career.Masterpiece {
		return runCraftsmanTerm(r, p, c, career, cc)
	}
	ccVal := c.Score(cc)
	mod := p.RiskMod(*c, ccVal) // caution (+), bravery (-), or 0

	if r.Resolve(dice.Check{Dice: 2, Target: ccVal + mod}).Success {
		// Survived. Reward roll; for an armed-forces career a success is a Medal.
		if r.Resolve(dice.Check{Dice: 2, Target: ccVal - mod}).Success && career.hasRanks() {
			c.Medals++
		}
	} else {
		// Risk failed: the CC drops by any negative (bravery) mod, then Flux.
		negMods := 0
		if mod < 0 {
			negMods = -mod
		}
		injury, newVal := classifyInjury(ccVal, negMods, r.Flux())
		switch injury {
		case Fatal:
			c.scores[cc] = max(newVal, 0)
			c.Dead = true
			return Died
		case Disabling:
			c.scores[cc] = newVal
			return Disabled
		case Wounded:
			c.scores[cc] = newVal
			c.WoundBadges++
		}
	}

	resolveRank(r, c, run, career) // promotion / commission for a surviving term
	awardSkills(r, p, c, career)   // a surviving (even wounded) character gains skills
	return Ongoing
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
func runCitizenTerm(r *dice.Roller, p Policy, c *Character, run *careerRun, career Career, cc Characteristic) TermOutcome {
	if r.Resolve(dice.Check{Dice: 2, Target: c.Score(cc)}).Success {
		awardCitizenLife(p, c, run)
	}
	awardSkills(r, p, c, career)
	return Ongoing
}

// runFameTerm resolves an Entertainer's term (Book 1 p. 77). At the start of the
// term, events shift Fame by a Flux roll; if Fame increases the character gains
// Talent +1 and two extra skill rolls. There is no injury; only aging can end
// the career. (The optional second and third Flux rolls are deferred.)
func runFameTerm(r *dice.Roller, p Policy, c *Character, career Career) TermOutcome {
	before := c.Fame
	c.Fame = max(c.Fame+r.Flux(), 0)
	elig := career.EligPerTerm
	if c.Fame > before {
		c.Talent++
		elig += 2 // "If Fame Increases: 2 [skills] and Talent+1"
	}
	awardSkillsN(r, p, c, career, elig)
	return Ongoing
}

// masterpieceMinimum is the fewest Master Points that allow a Masterpiece
// attempt (Book 1 p. 75).
const masterpieceMinimum = 40

// runCraftsmanTerm resolves a Craftsman's term (Book 1 p. 75). The Craftsman
// attempts a Masterpiece instead of Risk & Reward: Master Points total the
// Controlling Characteristic, the Craftsman skill, and up to five other skills
// at level 6+. With at least 40 points a 9D roll at or under the total creates a
// Masterpiece; otherwise the attempt fails. Either way the Craftsman skill rises
// +1 (learning); success grants three extra skill rolls, failure one. There is
// no injury.
func runCraftsmanTerm(r *dice.Roller, p Policy, c *Character, career Career, cc Characteristic) TermOutcome {
	points := c.Score(cc) + c.Skills.Level("Craftsman") + c.Skills.TopLevels(5, 6, "Craftsman", "Language")
	elig := career.EligPerTerm
	if points >= masterpieceMinimum && r.Dice(9) <= points {
		c.Masterpieces++
		elig += 3
	} else {
		elig++
	}
	c.Skills.Raise("Craftsman", 1) // learning from the work, success or failure
	awardSkillsN(r, p, c, career, elig)
	return Ongoing
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

// resolveRank runs a surviving armed-forces character's rank step (Book 1 p. 64
// for the mechanic, p. 82 for the Soldier's targets and ladders). An enlisted
// character first rolls for Commission (success moves them to
// the officer track at Officer 1); failing that, they roll Enlisted Promotion.
// An officer rolls Officer Promotion. Promotion (not Commission) targets are
// raised by Medals and Wound Badges. Reaching a rank grants its automatic skill.
func resolveRank(r *dice.Roller, c *Character, run *careerRun, career Career) {
	if !career.hasRanks() {
		return
	}
	if run.officer {
		if run.rank < len(career.OfficerRanks) && promoted(r, *c, career.OfficerPromote) {
			run.rank++
			grantRankSkill(c, career.OfficerRanks, run.rank)
		}
		return
	}
	if promoted(r, *c, career.Commission) {
		run.officer = true
		run.rank = 1
		grantRankSkill(c, career.OfficerRanks, 1)
		return
	}
	if run.rank < len(career.EnlistedRanks) && promoted(r, *c, career.EnlistedPromote) {
		run.rank++
		grantRankSkill(c, career.EnlistedRanks, run.rank)
	}
}

// promoted resolves one promotion roll: 2D at or under the rule's characteristic,
// raised by Medals and Wound Badges when the rule allows.
func promoted(r *dice.Roller, c Character, rule PromotionRule) bool {
	target := c.Score(rule.Char)
	if rule.MedalsAndWounds {
		target += c.Medals + c.WoundBadges
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
			panic(fmt.Sprintf("chargen: skill column %d out of range 0-%d", col, len(career.Skills)-1))
		}
		applyCell(p, c, career.Skills[col][r.Die()-1])
	}
}

// applyCell applies one skill-grid cell: raise a skill (cascade skills grant a
// knowledge via the K-K-S progression), bump a characteristic (capped at the
// human maximum), or resolve a player choice among options.
func applyCell(p Policy, c *Character, cell Cell) {
	switch cell.Kind {
	case AwardSkill:
		if cell.Knowledge != "" {
			c.Skills.GrantCascade(cell.Skill, cell.Knowledge)
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
		if cell.Skill != "" {
			// A cascade choice: the options are knowledges under the parent skill
			// (e.g. Language/Galanglic), granted via the K-K-S progression.
			c.Skills.GrantCascade(cell.Skill, chosen)
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

const (
	MoneyColumn MusterColumn = iota
	BenefitColumn
)

// A BenefitKind identifies a mustering-out award.
type BenefitKind int

const (
	Cash     BenefitKind = iota // Value credits
	CharBump                    // +Value to characteristic Char
	Named                       // a named benefit (Ship Share, TAS Fellowship, …)
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
// +Terms when the career's MusterBenefitDMT is set (e.g. the Rogue), otherwise
// +Fame/2 (Fame is not yet tracked, so 0). The policy chooses the column.
func MusterOut(r *dice.Roller, p Policy, c *Character, rec CareerRecord, career Career) {
	rolls := rec.Terms
	if rec.Outcome == Disabled {
		rolls *= 2 // double the number of benefit rolls
	}
	for range rolls {
		col := p.MusterColumn(*c, rec)
		dm := 0
		if col == MoneyColumn || career.MusterBenefitDMT {
			dm = rec.Terms
		}
		row := min(max(r.Die()+dm, 1), 12)
		award := career.MusterOut[row].Money
		if col == BenefitColumn {
			award = career.MusterOut[row].Benefit
		}
		applyBenefit(c, award)
	}
}

// applyBenefit applies one mustering-out award.
func applyBenefit(c *Character, b Benefit) {
	switch b.Kind {
	case Cash:
		c.Credits += b.Value
	case CharBump:
		c.scores[b.Char] = min(c.scores[b.Char]+b.Value, maxCharacteristic)
	case Named:
		c.Benefits = append(c.Benefits, b.Name)
	}
}

// An Injury classifies the result of a failed Risk roll.
type Injury int

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
func continues(r *dice.Roller, p Policy, c Character, career Career, rec CareerRecord, run *careerRun) bool {
	target := career.Continue.target(c)
	if career.Continue.UseCC {
		target = c.Score(run.fixed) + career.Continue.Mod
	}
	if career.Continue.UseFame {
		target = c.Fame + career.Continue.Mod
	}
	if career.Continue.UseSkill != "" {
		target = c.Skills.Level(career.Continue.UseSkill)*career.Continue.SkillMult + career.Continue.Mod
	}
	if career.Continue.TermsMod {
		target += rec.Terms // more experience makes staying in easier (Book 1 p. 83, Agent)
	}
	res := r.Resolve(dice.Check{Dice: 2, Target: target})
	if res.Roll == 2 {
		return true // Mandatory Continue
	}
	return p.Continue(c, rec) && res.Success
}

// removeChar returns chars without the first occurrence of ch.
func removeChar(chars []Characteristic, ch Characteristic) []Characteristic {
	for i, x := range chars {
		if x == ch {
			return append(chars[:i], chars[i+1:]...)
		}
	}
	return chars
}
