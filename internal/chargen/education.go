package chargen

import "github.com/philoserf/t5/internal/dice"

// Pre-career education (Book 1, "Pre-Career Education", pp. 59-60). Before a
// career, a character may attend an educational institution: they Apply (a
// characteristic Check, with an optional Waiver on failure), then make one
// Pass/Fail Check per year, and on completing every year Graduate for a degree
// and an Education bump.
//
// This slice implements the academic programs (College and University) plus the
// remedial ED5. Deferred: Trade School, the higher and military institutions,
// the full Available-Skills matrix, Honors, and the Tra-based training path.

const (
	academicYears = 4 // College and University each require four Pass/Fail Checks
	ed5MaxEdu     = 4 // ED5 admits a character of Edu 4 or less
	ed5RaisesTo   = 5 // and raises their Edu to 5
)

// An academicProgram is a four-year degree course (Book 1 p. 60): College and
// University share the process and differ only in prerequisite and the Edu they
// confer on graduation.
type academicProgram struct {
	preReqEdu int
	gradEdu   int
}

var (
	college    = academicProgram{preReqEdu: 5, gradEdu: 8}
	university = academicProgram{preReqEdu: 7, gradEdu: 9}
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
	switch {
	case c.Score(Education) >= university.preReqEdu:
		attendAcademic(r, p, c, university)
	case c.Score(Education) >= college.preReqEdu:
		attendAcademic(r, p, c, college)
	}
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

// attendAcademic runs a character through a College or University program (Book
// 1 p. 60). A character below the program's Edu prerequisite cannot attend and
// is left unchanged.
func attendAcademic(r *dice.Roller, p Policy, c *Character, prog academicProgram) {
	if c.Score(Education) < prog.preReqEdu {
		return
	}
	priorWaivers := 0
	// Apply for admission: Check the better of Int or Edu, Waiver on failure.
	if !admitted(r, p, c, bestChar(*c, Intelligence, Education), &priorWaivers) {
		return
	}
	// One Pass/Fail Check per year; a failure ends attendance unless Waived.
	passCh := bestChar(*c, Intelligence, Education)
	passes := 0
	for range academicYears {
		if r.Resolve(dice.Check{Dice: 2, Target: c.Score(passCh)}).Success {
			passes++
			awardAcademicPass(c, p, passes)
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

// awardAcademicPass applies one passed year: Major +1 (declared on the first
// pass), plus Minor +1 on every second pass (declared on the first — Book 1 p.
// 60: "Major+1 per Pass" and "Minor+1 per 2 Passes").
func awardAcademicPass(c *Character, p Policy, passNum int) {
	if c.Major == "" {
		c.Major = p.ChooseSkill(*c, academicMajors)
	}
	c.Skills.Raise(c.Major, 1)
	if passNum%2 == 0 {
		if c.Minor == "" {
			c.Minor = p.ChooseSkill(*c, without(academicMajors, c.Major))
		}
		c.Skills.Raise(c.Minor, 1)
	}
}

// graduate applies a program's graduation benefit: Edu rises to the program's
// graduation level, or +1 if already there (Book 1 p. 60 note), and a BA is
// recorded.
func graduate(c *Character, prog academicProgram) {
	if c.scores[Education] < prog.gradEdu {
		c.scores[Education] = prog.gradEdu
	} else {
		c.scores[Education] = min(c.scores[Education]+1, maxCharacteristic)
	}
	c.Degrees = append(c.Degrees, "BA")
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

// without returns list with every occurrence of x removed.
func without(list []string, x string) []string {
	out := make([]string, 0, len(list))
	for _, s := range list {
		if s != x {
			out = append(out, s)
		}
	}
	return out
}
