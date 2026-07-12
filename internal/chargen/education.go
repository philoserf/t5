package chargen

import "github.com/philoserf/t5/internal/dice"

// Pre-career education (Book 1, "Pre-Career Education", pp. 59-60). Before a
// career, a character may attend an educational institution: they Apply (a
// characteristic Check, with an optional Waiver on failure), then make one
// Pass/Fail Check per year, and on completing every year Graduate for a degree
// and an Education bump.
//
// This slice implements College — the worked example on p. 60 (Eneri Dinsha).
// Deferred: ED5, University, Trade School, the higher and military institutions,
// the full Available-Skills matrix, Honors, and the Tra-based training path.

const (
	collegePreReqEdu = 5 // College requires Edu 5+
	collegeYears     = 4 // four Pass/Fail Checks
	collegeGradEdu   = 8 // Graduation raises Edu to 8 (a BA)
)

// academicMajors is a representative list of College Major/Minor subjects (the
// full Available-Skills matrix is deferred). A character's Minor must differ
// from their Major.
var academicMajors = []string{
	"Psychology", "Robotics", "History", "Physics", "Chemistry",
	"Biology", "Economics", "Philosophy", "Astronomy", "Sophontology",
}

// AttendCollege runs a character through College (Book 1 p. 60). It rolls dice,
// so callers gate it on Policy.PursueEducation to keep generation deterministic
// where no education is wanted. A character below the Edu prerequisite cannot
// attend and is left unchanged.
func AttendCollege(r *dice.Roller, p Policy, c *Character) {
	if c.Score(Education) < collegePreReqEdu {
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
	for range collegeYears {
		if r.Resolve(dice.Check{Dice: 2, Target: c.Score(passCh)}).Success {
			passes++
			awardCollegePass(c, p, passes)
		} else if !waiverGranted(r, p, c, &priorWaivers) {
			return // failed out — no graduation
		}
	}
	graduateCollege(c)
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
	*priorWaivers++
	return res.Success
}

// awardCollegePass applies one passed year: Major +1 (declared on the first
// pass), plus Minor +1 on every second pass (declared on the first — Book 1 p.
// 60: "Major+1 per Pass" and "Minor+1 per 2 Passes").
func awardCollegePass(c *Character, p Policy, passNum int) {
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

// graduateCollege applies the College graduation benefit: Edu rises to 8, or +1
// if already there (Book 1 p. 60 note), and a BA is recorded.
func graduateCollege(c *Character) {
	if c.scores[Education] < collegeGradEdu {
		c.scores[Education] = collegeGradEdu
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

// without returns list with the first occurrence of x removed.
func without(list []string, x string) []string {
	out := make([]string, 0, len(list))
	for _, s := range list {
		if s != x {
			out = append(out, s)
		}
	}
	return out
}
