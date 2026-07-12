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
}

// A TermOutcome is how a character left a career.
type TermOutcome int

const (
	MusteredOut TermOutcome = iota // left the career (voluntary or failed to continue)
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
	run := careerRun{ccPool: append([]Characteristic(nil), career.ControllingChars...)}
	rec := CareerRecord{Career: career.ID}

	for {
		runTerm(p, c, &run, career)
		rec.Terms++
		c.Age += termYears
		AgingCheck(r, c) // no-op before age 34
		if c.Dead {
			rec.Outcome = Died
			break
		}
		if !continues(r, p, *c, career, rec) {
			rec.Outcome = MusteredOut
			break
		}
	}
	c.Careers = append(c.Careers, rec)
}

// runTerm resolves one term. For now it only advances the Controlling
// Characteristic rotation; Risk & Reward and skills arrive in later slices.
func runTerm(p Policy, c *Character, run *careerRun, career Career) {
	_ = selectCC(p, *c, run, career)
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
