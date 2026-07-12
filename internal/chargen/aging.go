package chargen

import "github.com/philoserf/t5/internal/dice"

// Aging (Book 1, "The Aging Process", p. 89). Once aging begins it strikes every
// four years: each applicable characteristic makes an Aging Check, and a
// character "wants to fail" it.

const (
	// physicalAgingAge is Life Stage 5 (Peak) — physical aging (Str, Dex, End)
	// begins here. mentalAgingAge is Life Stage 9 (Retirement) — mental aging
	// (Int) begins here.
	physicalAgingAge = 34
	mentalAgingAge   = 66
)

// lifeStage maps an age to a life stage: 0 for infancy (under 2), then one stage
// per eight years (2-9 Child = 1 … 34-41 Peak = 5 … 66+ Retirement = 9).
func lifeStage(age int) int {
	if age < 2 {
		return 0
	}
	return (age-2)/8 + 1
}

// An Illness is the fallout when characteristics are driven to zero by a single
// aging check.
type Illness int

const (
	NoIllness      Illness = iota
	MajorIllness           // two characteristics zeroed: four weeks R&R
	ExtremeIllness         // three characteristics zeroed: four months R&R
)

// agingCascade classifies the outcome of one aging check in which zeroed
// characteristics were driven to zero, given the count of prior extreme-illness
// events. It reports the illness level and whether the character dies — the
// second extreme illness is fatal.
func agingCascade(zeroed, priorExtreme int) (Illness, bool) {
	switch {
	case zeroed >= 3:
		return ExtremeIllness, priorExtreme+1 >= 2
	case zeroed == 2:
		return MajorIllness, false
	default:
		return NoIllness, false
	}
}

// AgingCheck resolves one aging check for a character at its current age: 2D per
// applicable characteristic, and on a roll *under* the life stage the
// characteristic drops by 1 (the character wanted to fail). A characteristic
// driven to 0 resets to 1; two zeroed is a major illness, three an extreme
// illness, and a second extreme illness kills. It has no effect before aging
// begins (age 34) or once the character is dead. R&R downtime is not modeled.
func AgingCheck(r *dice.Roller, c *Character) {
	if c.Dead || c.Age < physicalAgingAge {
		return
	}
	stage := lifeStage(c.Age)

	chars := []Characteristic{Strength, Dexterity, Endurance}
	if c.Age >= mentalAgingAge {
		chars = append(chars, Intelligence)
	}

	zeroed := 0
	for _, ch := range chars {
		if r.Dice(2) < stage { // success ages the characteristic
			if v := c.scores[ch] - 1; v <= 0 {
				c.scores[ch] = 1
				zeroed++
			} else {
				c.scores[ch] = v
			}
		}
	}

	illness, dead := agingCascade(zeroed, c.extremeAgings)
	if illness == ExtremeIllness {
		c.extremeAgings++
	}
	if dead {
		c.Dead = true
	}
}
