package chargen

import "github.com/philoserf/t5/internal/dice"

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

// A Qualification is a career's entry gate: roll 2D at or under a characteristic
// (plus a modifier).
type Qualification struct {
	Char Characteristic
	Mod  int
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
	q := career.Qualify
	if r.Resolve(dice.Check{Dice: 2, Target: c.Score(q.Char), Mod: q.Mod}).Success {
		RunCareer(r, p, &c, career)
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
