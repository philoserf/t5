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
	"github.com/philoserf/t5/internal/skill"
	"github.com/philoserf/t5/internal/worldgen"
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

// A Character is a Traveller character: the six human characteristic scores,
// age, and the results of any careers served.
type Character struct {
	scores [count]int
	Age    int  // years; a freshly generated character starts at 18
	Dead   bool // set when aging (or a career mishap) kills the character

	Homeworld        worldgen.World // the world the character was raised on (Book 1 p. 56)
	Gender           string         // the individual's gender, for a non-human sophont (Book 3 p. 230); "" for a plain human
	Caste            string         // the individual's caste, for a casted sophont (Book 3 p. 229); "" if none
	Major            string         // College Major subject, if educated
	Minor            string         // College Minor subject, if educated
	Degrees          []string       // academic degrees earned (BA, …)
	Skills           skill.Set      // skills and knowledges gained in careers
	Careers          []CareerRecord // one record per career served
	Credits          int            // cash mustered out
	Benefits         []string       // named mustering-out benefits (Ship Share, TAS, …)
	WoundBadges      int            // career injuries survived
	Medals           int            // awards earned on successful Reward rolls (armed forces)
	Fame             int            // reputation, the Entertainer's measure of success (Book 1 p. 77)
	Talent           int            // the Entertainer's performance ability
	Masterpieces     int            // works of art the Craftsman has created (Book 1 p. 75)
	MasterpieceValue int            // total Cr value of those masterpieces
	Publications     int            // research the Scholar has published (Book 1 p. 76)
	Commendations    int            // official recognitions the Agent has earned (Book 1 p. 83)
	LandGrants       int            // fiefs granted to the Noble on each Elevation, and to the Scout on each Discovery (Book 1 pp. 79, 85)
	Discoveries      int            // valuable worlds or features the Scout has found (Book 1 p. 79)
	ShipShares       int            // ownership shares the Merchant accumulates (Book 1 p. 80)

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

// UPP renders the six characteristics as an eHex string, e.g. "777777". It is a
// renderer, so it must be total: a sophont characteristic can exceed the eHex
// range (an 8D Strength reaches 48), and String — an fmt.Stringer — must not
// panic on it. ehex.Format renders such a value as "?"; the characteristic bounds
// that keep a human in range are enforced in the mutation paths (the
// maxCharacteristic clamps), not here. cmd/sophont carries a non-lossy renderer.
func (c Character) UPP() string {
	var b strings.Builder
	b.Grow(count)
	for _, v := range c.scores {
		b.WriteString(ehex.Format(v))
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
