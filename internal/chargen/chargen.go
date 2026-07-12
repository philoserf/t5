// Package chargen creates Traveller5 characters. This first stage generates a
// character's Universal Personality Profile (UPP) — the six human
// characteristics — and supports the Check Characteristic mechanic. Careers
// (qualification, terms, skills, mustering out) are a later stage.
//
// See Book 1, "Characteristics" (pp. 47+): the six human characteristics are
// each rolled as 2D and recorded as an eHex UPP string such as "777777".
package chargen

import (
	"fmt"
	"strings"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/ehex"
)

// A Characteristic identifies one of the six human characteristics, in UPP
// order (physical, then mental, then social).
type Characteristic int

const (
	Strength Characteristic = iota
	Dexterity
	Endurance
	Intelligence
	Education
	Social
)

// count is the number of characteristics in a human UPP.
const count = 6

// abbrev holds the three-letter abbreviation of each characteristic.
var abbrev = [count]string{"Str", "Dex", "End", "Int", "Edu", "Soc"}

// String returns the characteristic's abbreviation, e.g. "Str".
func (c Characteristic) String() string {
	if c < 0 || c >= count {
		return "?"
	}
	return abbrev[c]
}

// A Character is a Traveller character: the six human characteristic scores
// plus age. Careers and aging will extend it further.
type Character struct {
	scores [count]int
	Age    int  // years; a freshly generated character starts at 18
	Dead   bool // set when aging (or, later, a career mishap) kills the character

	extremeAgings int // aging checks that zeroed 3+ characteristics; the 2nd is fatal
}

// startingAge is the age of a freshly generated character — Young Adult (Life
// Stage 3), where adventuring traditionally begins.
const startingAge = 18

// Generate rolls a character's six characteristics, each 2D, in UPP order, at
// the starting age.
func Generate(r *dice.Roller) Character {
	c := Character{Age: startingAge}
	for i := range c.scores {
		c.scores[i] = r.Dice(2)
	}
	return c
}

// LifeStage returns the character's current life stage (Book 1 p. 89): 0 Infant,
// 3 Young Adult (18), 5 Peak (34), 9 Retirement (66).
func (c Character) LifeStage() int {
	return lifeStage(c.Age)
}

// Score returns the value of one characteristic. It panics on a Characteristic
// outside the defined set — an invalid constant is a programming error, and a
// clear panic beats masking the bug behind a plausible-looking zero score.
func (c Character) Score(ch Characteristic) int {
	if ch < 0 || ch >= count {
		panic(fmt.Sprintf("chargen: invalid characteristic %d", int(ch)))
	}
	return c.scores[ch]
}

// UPP renders the six characteristics as an eHex string, e.g. "777777".
func (c Character) UPP() string {
	var b strings.Builder
	b.Grow(count)
	for _, v := range c.scores {
		b.WriteByte(ehex.Digit(v))
	}
	return b.String()
}

// String returns the UPP.
func (c Character) String() string {
	return c.UPP()
}

// Check resolves a Check Characteristic: roll numDice (0 defaults to the
// standard 2D; use dice.Easy or dice.Hard for 1D or 3D) against the
// characteristic, succeeding on a roll of the score or less. mod adjusts the
// target the usual way (a positive mod makes success easier).
//
// The rulebook's escalating penalty for reusing a characteristic before two
// others are used is session state, left to the caller.
func (c Character) Check(r *dice.Roller, ch Characteristic, numDice, mod int) dice.CheckResult {
	return r.Resolve(dice.Check{Dice: numDice, Target: c.Score(ch), Mod: mod})
}
