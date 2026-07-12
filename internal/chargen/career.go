package chargen

import (
	"fmt"

	"github.com/philoserf/t5/internal/dice"
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
	Scout CareerID = iota // the first implemented career
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

// target returns the qualification target: the highest of the characteristics.
func (q Qualification) target(c Character) int {
	best := 0
	for _, ch := range q.Chars {
		best = max(best, c.Score(ch))
	}
	return best + q.Mod
}

// A ContinueRule gives the target of a career's Continue roll — either a fixed
// number or a characteristic value — plus a modifier.
type ContinueRule struct {
	UseChar bool
	Char    Characteristic
	Fixed   int
	Mod     int
}

// target resolves the Continue target for a character.
func (rule ContinueRule) target(c Character) int {
	if rule.UseChar {
		return c.Score(rule.Char) + rule.Mod
	}
	return rule.Fixed + rule.Mod
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
	EligPerTerm      int // number of skill rolls a surviving term grants
	Skills           SkillGrid
	MusterOut        MusterTable
}

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
	Rank    int
	Outcome TermOutcome
}

// careerRun is the transient bookkeeping for one career, live only during
// generation (kept out of Character, like systemgen's orbit bookkeeping).
type careerRun struct {
	ccPool []Characteristic // Controlling Characteristics not yet used this cycle
}

// GenerateCareered generates a character and runs one career on them. The
// character qualifies (2D at or under the career's characteristic); on failure
// they enter no career and remain a fresh 18-year-old.
func GenerateCareered(r *dice.Roller, p Policy, career Career) Character {
	c := Generate(r)
	if r.Resolve(dice.Check{Dice: 2, Target: career.Qualify.target(c)}).Success {
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
	if len(career.ControllingChars) == 0 {
		panic("chargen: career " + career.Name + " has no controlling characteristics")
	}
	run := careerRun{ccPool: append([]Characteristic(nil), career.ControllingChars...)}
	rec := CareerRecord{Career: career.ID}

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
		if !continues(r, p, *c, career, rec) {
			rec.Outcome = MusteredOut
			break
		}
	}
	c.Careers = append(c.Careers, rec)
}

// runTerm resolves one term (Book 1 p. 64). It selects the Controlling
// Characteristic, rolls Risk (2D <= CC + mod) and, on survival, Reward
// (2D <= CC - mod, mods flipped). A failed Risk injures the character. It
// returns Ongoing, Disabled, or Died. (Skills and the per-career reward benefit
// arrive in later slices.)
func runTerm(r *dice.Roller, p Policy, c *Character, run *careerRun, career Career) TermOutcome {
	cc := selectCC(p, *c, run, career)
	ccVal := c.Score(cc)
	mod := p.RiskMod(*c, ccVal) // caution (+), bravery (-), or 0

	if r.Resolve(dice.Check{Dice: 2, Target: ccVal + mod}).Success {
		// Survived. Reward roll (its benefit is applied where career data lands).
		r.Resolve(dice.Check{Dice: 2, Target: ccVal - mod})
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

	awardSkills(r, p, c, career) // a surviving (even wounded) character gains skills
	return Ongoing
}

// awardSkills grants the term's skill eligibility: for each roll the policy picks
// a column of the career's skill grid and 1D selects the row (Book 1 p. 65).
func awardSkills(r *dice.Roller, p Policy, c *Character, career Career) {
	for range career.EligPerTerm {
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
		c.Skills.Raise(p.ChooseSkill(*c, cell.Options), 1)
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
	AwardChoice                 // raise one skill the policy picks from Options
)

// A Cell is one entry in a career's skill grid.
type Cell struct {
	Kind      CellKind
	Skill     string         // AwardSkill: the skill (a cascade parent, if cascade)
	Knowledge string         // AwardSkill: the knowledge, for a cascade skill
	Char      Characteristic // AwardBump
	Options   []string       // AwardChoice: the skills to pick among
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
// 1D plus the column DM — the Money column adds +Terms, the Benefit column
// +Fame/2 (Fame is not yet tracked, so 0) — and the policy chooses the column.
func MusterOut(r *dice.Roller, p Policy, c *Character, rec CareerRecord, career Career) {
	rolls := rec.Terms
	if rec.Outcome == Disabled {
		rolls *= 2 // double the number of benefit rolls
	}
	for range rolls {
		col := p.MusterColumn(*c, rec)
		dm := 0
		if col == MoneyColumn {
			dm = rec.Terms // Money DM = + Terms
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
// FixedCC the same characteristic serves the entire career.
func selectCC(p Policy, c Character, run *careerRun, career Career) Characteristic {
	if career.CCMode == FixedCC {
		return career.ControllingChars[0]
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
// the policy wishes to and the 2D Continue roll succeeds.
func continues(r *dice.Roller, p Policy, c Character, career Career, rec CareerRecord) bool {
	res := r.Resolve(dice.Check{Dice: 2, Target: career.Continue.target(c)})
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
