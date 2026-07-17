package chargen

import (
	"slices"

	"github.com/philoserf/t5/internal/dice"
)

// Pre-career education (Book 1, "Pre-Career Education", pp. 59-60). Before a
// career, a character may attend an educational institution: they Apply (a
// characteristic Check, with an optional Waiver on failure), then make one
// Pass/Fail Check per year, and on completing every year Graduate for a degree
// and an Education bump.
//
// This slice implements the remedial ED5, the vocational Trade School, the
// undergraduate College and University, and the post-graduate Masters and
// Professors ladder. Deferred: the professional and military institutions
// (Service Academy, Medical/Law, Honors, OTC/NOTC, Flight School), the full
// Available-Skills matrix, and the Tra-based training path.

const (
	ed5MaxEdu   = 4 // ED5 admits a character of Edu 4 or less
	ed5RaisesTo = 5 // and raises their Edu to 5

	tradeSchoolPreReqEdu = 5 // Trade School admits a character of Edu 5+
	tradeSchoolMajorBump = 2 // and, on completion, raises a vocational Major +2
)

// An academicProgram is a degree course (Book 1 p. 60): apply, one Pass/Fail
// Check per year, then graduate for a degree and an Edu bump. All five higher-
// education programs — College, University, Service Academy, Masters, and
// Professors — share one merged "Provides" cell on p.60: Major+1 per Pass and
// Minor+1 per 2 Passes. (The Masters and Professors rows were once read as awarding
// only the Minor / nothing; the cell spans them too. The undergraduate goldens
// never reach the post-grad ladder, so nothing caught it.) Post-graduate programs
// are gated on a prior degree rather than an Edu level.
type academicProgram struct {
	name         string
	years        int
	awardsMajor  bool   // raise a Major each pass
	awardsMinor  bool   // raise a Minor every second pass
	preReqEdu    int    // minimum Edu to enroll (0 when preReqDegree is the gate)
	preReqDegree string // a prior degree required to enroll ("" for undergraduate programs)
	gradEdu      int
	degree       string // the degree conferred on graduation
}

var (
	college = academicProgram{
		name:        "College",
		years:       4,
		awardsMajor: true,
		awardsMinor: true,
		preReqEdu:   5,
		gradEdu:     8,
		degree:      "BA",
	}
	university = academicProgram{
		name:        "University",
		years:       4,
		awardsMajor: true,
		awardsMinor: true,
		preReqEdu:   7,
		gradEdu:     9,
		degree:      "BA",
	}
	masters = academicProgram{
		name:         "Masters",
		years:        2,
		awardsMajor:  true,
		awardsMinor:  true,
		preReqDegree: "BA",
		gradEdu:      9,
		degree:       "MA",
	}
	professors = academicProgram{
		name:         "Professors",
		years:        2,
		awardsMajor:  true,
		awardsMinor:  true,
		preReqDegree: "MA",
		gradEdu:      12,
		degree:       "Professor",
	}
)

// academicMajors is a representative list of College Major/Minor subjects (the
// full Available-Skills matrix is deferred). A character's Minor must differ
// from their Major.
var academicMajors = []string{
	"Psychology", "Robotics", "History", "Physics", "Chemistry",
	"Biology", "Economics", "Philosophy", "Astronomy", "Sophontology",
}

// educate runs a character's pre-career education (Book 1 pp. 59-60). It rolls
// dice, so callers gate it on Policy.PursueEducation to keep generation
// deterministic where no education is wanted. A character below the College
// prerequisite first attempts the remedial ED5, then attends the most
// prestigious academic program they qualify for.
func educate(r *dice.Roller, p Policy, c *Character) {
	if c.Score(Education) < college.preReqEdu {
		attemptED5(r, c)
	}
	// A character who elects the vocational path attends a Trade School instead of
	// the academic College/University (they share the Edu 5+ prerequisite).
	if p.ChooseTradeSchool(*c) && c.Score(Education) >= tradeSchoolPreReqEdu {
		attendTradeSchool(r, p, c)

		return
	}

	switch {
	case c.Score(Education) >= university.preReqEdu:
		attendAcademic(r, p, c, university)
	case c.Score(Education) >= college.preReqEdu:
		attendAcademic(r, p, c, college)
	}
	// A graduate — with a BA — may climb the post-graduate ladder: a Masters, then
	// (having earned the MA) a terminal Professors program.
	if p.PursueGraduateSchool(*c) && c.hasDegree(masters.preReqDegree) {
		attendAcademic(r, p, c, masters)

		if c.hasDegree(professors.preReqDegree) {
			attendAcademic(r, p, c, professors)
		}
	}
}

// hasDegree reports whether the character has earned a given degree.
func (c Character) hasDegree(degree string) bool {
	return slices.Contains(c.Degrees, degree)
}

// attendTradeSchool runs a one-year Trade School (Book 1 p. 60): apply with a
// Check Int (Waiver on failure), then a single Pass/Fail Check (the better of
// Int or Edu, Waiver on failure). Completing the year declares a vocational
// Major and raises it +2. Unlike an academic program it grants no Minor, no Edu
// bump, and no degree.
func attendTradeSchool(r *dice.Roller, p Policy, c *Character) {
	if c.Score(Education) < tradeSchoolPreReqEdu {
		return
	}

	priorWaivers := 0
	if !admitted(r, p, c, Intelligence, &priorWaivers) {
		return
	}

	passCh := bestChar(*c, Intelligence, Education)
	if !r.Resolve(dice.Check{Dice: 2, Target: c.Score(passCh)}).Success &&
		!waiverGranted(r, p, c, &priorWaivers) {
		return // failed the year out — no Major
	}

	c.Major = p.ChooseSkill(*c, theTrades)
	c.Skills.Raise(c.Major, tradeSchoolMajorBump)
}

// attemptED5 runs the ED5 remedial program (Book 1 p. 60): a character of Edu 4
// or less may Check Int once to raise their Edu to 5, reaching the College
// prerequisite. DefaultPolicy does not pursue education this low, so ED5 serves
// policies that deliberately educate a low-Edu character.
func attemptED5(r *dice.Roller, c *Character) {
	if c.Score(Education) > ed5MaxEdu {
		return
	}

	if r.Resolve(dice.Check{Dice: 2, Target: c.Score(Intelligence)}).Success {
		c.scores[Education] = ed5RaisesTo
	}
}

// attendAcademic runs a character through an academic program (Book 1 p. 60). A
// character who fails to meet the program's prerequisite — an Edu level, or a
// prior degree for a post-graduate program — cannot attend and is left
// unchanged.
func attendAcademic(r *dice.Roller, p Policy, c *Character, prog academicProgram) {
	if prog.preReqDegree != "" && !c.hasDegree(prog.preReqDegree) {
		return
	}

	if c.Score(Education) < prog.preReqEdu {
		return
	}

	priorWaivers := 0
	// Admission and each yearly Pass/Fail Check use the better of Int or Edu.
	passCh := bestChar(*c, Intelligence, Education)
	if !admitted(r, p, c, passCh, &priorWaivers) {
		return
	}

	passes := 0

	for range prog.years {
		if r.Resolve(dice.Check{Dice: 2, Target: c.Score(passCh)}).Success {
			passes++
			awardAcademicPass(c, p, prog, passes)
		} else if !waiverGranted(r, p, c, &priorWaivers) {
			return // failed out — no graduation
		}
	}

	graduate(c, prog)
}

// admitted resolves the admission Check (2D at or under the characteristic),
// falling back to a Waiver on failure.
func admitted(r *dice.Roller, p Policy, c *Character, ch Characteristic, priorWaivers *int) bool {
	if r.Resolve(dice.Check{Dice: 2, Target: c.Score(ch)}).Success {
		return true
	}

	return waiverGranted(r, p, c, priorWaivers)
}

// waiverGranted attempts an Educational Waiver (Book 1 p. 59): Check Social, at a
// cumulative -1 per prior waiver attempt (successful or not). It returns whether
// the character may continue, and counts the attempt.
func waiverGranted(r *dice.Roller, p Policy, c *Character, priorWaivers *int) bool {
	if !p.TakeWaiver(*c, *priorWaivers) {
		return false
	}

	res := r.Resolve(dice.Check{Dice: 2, Target: c.Score(Social), Mod: -*priorWaivers})
	(*priorWaivers)++

	return res.Success
}

// awardAcademicPass applies one passed year (Book 1 p. 60: "Major+1 per Pass" and
// "Minor+1 per 2 Passes"): the Major rises each pass, and every second pass raises
// the Minor. All five higher-education programs share this rule.
func awardAcademicPass(c *Character, p Policy, prog academicProgram, passNum int) {
	if prog.awardsMajor {
		if c.Major == "" {
			c.Major = p.ChooseSkill(*c, academicMajors)
		}

		c.Skills.Raise(c.Major, 1)
	}

	if prog.awardsMinor && passNum%2 == 0 {
		if c.Minor == "" {
			c.Minor = p.ChooseSkill(*c, without(academicMajors, c.Major))
		}

		c.Skills.Raise(c.Minor, 1)
	}
}

// graduate applies a program's graduation benefit: Edu rises to the program's
// graduation level, or +1 if already there (Book 1 p. 60 note), and the
// program's degree is recorded.
func graduate(c *Character, prog academicProgram) {
	if c.scores[Education] < prog.gradEdu {
		c.scores[Education] = prog.gradEdu
	} else {
		c.scores[Education] = min(c.scores[Education]+1, maxCharacteristic)
	}

	c.Degrees = append(c.Degrees, prog.degree)
}

// bestChar returns the highest-scoring of the given characteristics.
func bestChar(c Character, chars ...Characteristic) Characteristic {
	best := chars[0]
	for _, ch := range chars[1:] {
		if c.Score(ch) > c.Score(best) {
			best = ch
		}
	}

	return best
}

// without returns a copy of list with every occurrence of x removed.
func without(list []string, x string) []string {
	return slices.DeleteFunc(slices.Clone(list), func(s string) bool { return s == x })
}
