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
)

// A Sense is one of the six senses (Book 1 p. 186): its ID letter, name, the
// human sense Constant, and whether it operates at range (rolling R=Range dice)
// or in contact (2D). Awareness and Perception are not available to humans (a
// human Constant of 0); a non-human sophont supplies its own Constant.
type Sense struct {
	ID       byte
	Name     string
	Constant int
	AtRange  bool
}

// The six senses with their human Constants (Book 1 p. 187). Awareness and
// Perception have no human value ("X"), modelled as Constant 0.
var (
	Vision     = Sense{'V', "Vision", 16, true}
	Hearing    = Sense{'H', "Hearing", 16, true}
	Smell      = Sense{'S', "Smell", 10, false}
	Touch      = Sense{'T', "Touch", 6, false}
	Awareness  = Sense{'A', "Awareness", 0, true}
	Perception = Sense{'P', "Perception", 0, true}
)

// String renders the sense in the book's ID-Constant form, e.g. "V-16".
func (s Sense) String() string {
	return fmt.Sprintf("%c-%d", s.ID, s.Constant)
}

// NoticeAtRange resolves an at-range sense Action (Vision/Hearing/Awareness/
// Perception; Book 1 p. 190). It rolls R=Range dice — Range 0 and the R/T
// sub-bands count as 1 — and succeeds on a roll at or under Constant plus the
// Benchmark (objectSize - Range) plus any situational mods (Master Mods; higher
// is better). A Benchmark that drives the target to zero or less can't be sensed.
func NoticeAtRange(r *dice.Roller, s Sense, objectSize, rng int, mods ...int) dice.CheckResult {
	return r.Resolve(dice.Check{
		Dice:   max(rng, 1),
		Target: s.Constant,
		Mod:    (objectSize - rng) + sum(mods),
	})
}

// NoticeInContact resolves an in-contact sense Action (Touch, and Smell by scent
// intensity; Book 1 p. 187): 2D at or under Constant + benchmark + mods.
func NoticeInContact(r *dice.Roller, s Sense, benchmark int, mods ...int) dice.CheckResult {
	return r.Resolve(dice.Check{
		Dice:   dice.Average,
		Target: s.Constant,
		Mod:    benchmark + sum(mods),
	})
}

// RangeBand returns the sense Range band (0-9) for a distance in meters, using
// the shared world-range ladder (Book 1 pp. 24, 188). The R and T reading/talking
// sub-bands both count as Range 1, per the sense rules.
func RangeBand(meters float64) int {
	switch code := rangeband.WorldForDistance(meters).Code; code {
	case "0":
		return 0
	case "R", "T":
		return 1
	default:
		n := 0
		fmt.Sscan(code, &n)
		return n
	}
}

func sum(xs []int) int {
	total := 0
	for _, x := range xs {
		total += x
	}
	return total
}
