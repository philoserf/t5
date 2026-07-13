// Package senses implements Traveller5's six senses (Book 1 pp. 186-199). A
// sense Action is a roll-low task with no skill: the character rolls dice and
// succeeds on a roll at or under Constant + Benchmark + Mods. At-range senses
// (Vision, Hearing, Awareness, Perception) roll dice equal to the R= Range band;
// in-contact senses (Touch, and Smell by scent intensity) roll 2D.
package senses

import (
	"fmt"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/rangeband"
	"github.com/philoserf/t5/internal/task"
)

// A Sense is one of the six senses (Book 1 p. 186): its ID letter and the sense
// Constant. A Constant of 0 or less marks a sense the sophont does not have — for
// humans, Awareness and Perception ("X") — and its Actions automatically fail; a
// non-human sophont supplies its own positive Constant.
type Sense struct {
	ID       byte
	Constant int
}

// The six senses with their human Constants (Book 1 p. 187). Vision and Hearing
// operate at range, Touch and Smell in contact; Awareness and Perception have no
// human value ("X"), modelled as Constant 0.
var (
	Vision     = Sense{'V', 16}
	Hearing    = Sense{'H', 16}
	Smell      = Sense{'S', 10}
	Touch      = Sense{'T', 6}
	Awareness  = Sense{'A', 0}
	Perception = Sense{'P', 0}
)

// String renders the sense in the book's ID-Constant form, e.g. "V-16".
func (s Sense) String() string {
	return fmt.Sprintf("%c-%d", s.ID, s.Constant)
}

// NoticeAtRange resolves an at-range sense Action (Vision/Hearing/Awareness/
// Perception; Book 1 p. 190). It rolls R=Range dice — Range 0 and the R/T
// sub-bands count as 1 — and succeeds on a roll at or under Constant plus the
// Benchmark (objectSize - Range) plus any situational mods (Master Mods; higher
// is better). A sense the sophont lacks (Constant 0 or less) automatically fails.
func NoticeAtRange(r *dice.Roller, s Sense, objectSize, rng int, mods ...int) dice.CheckResult {
	if s.Constant <= 0 {
		return dice.CheckResult{} // an absent sense (e.g. human Awareness) never succeeds
	}
	return task.ResolveDice(r, rng, s.Constant+(objectSize-rng), mods...)
}

// NoticeInContact resolves an in-contact sense Action (Touch, and Smell by scent
// intensity; Book 1 p. 187): 2D at or under Constant + benchmark + mods. A sense
// the sophont lacks (Constant 0 or less) automatically fails.
func NoticeInContact(r *dice.Roller, s Sense, benchmark int, mods ...int) dice.CheckResult {
	if s.Constant <= 0 {
		return dice.CheckResult{}
	}
	return task.ResolveDice(r, dice.Average, s.Constant+benchmark, mods...)
}

// RangeBand returns the sense Range band (0-9) for a distance in meters, using
// the shared world-range ladder (Book 1 pp. 24, 188). The R and T reading/talking
// sub-bands both count as Range 1, per the sense rules.
func RangeBand(meters float64) int {
	n, ok := rangeband.WorldForDistance(meters).Number()
	if !ok {
		return 1 // the R and T reading/talking sub-bands both count as Range 1
	}
	return n
}
