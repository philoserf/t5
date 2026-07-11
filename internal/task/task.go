// Package task implements Traveller5's Universal Task Format (Book 1, Tasks
// pp. 120-131). A task rolls a number of dice set by its Difficulty and
// succeeds when the total is at or under a target number built from a
// characteristic, a skill, and modifiers — the roll-low resolution the dice
// engine already provides.
package task

import (
	"github.com/philoserf/t5/internal/dice"
)

// Difficulty is a task's difficulty level (Book 1 p. 129). It sets how many
// dice are rolled: more dice make a low roll harder, so higher difficulty is
// harder. The ladder runs Easy (1D) through Beyond Impossible (8D).
type Difficulty int

const (
	Easy Difficulty = iota
	Average
	Difficult
	Formidable
	Staggering
	Hopeless
	Impossible
	BeyondImpossible
)

var difficultyNames = [...]string{
	"Easy", "Average", "Difficult", "Formidable",
	"Staggering", "Hopeless", "Impossible", "Beyond Impossible",
}

// Dice returns the number of dice a task at this difficulty rolls (Easy 1D …
// Beyond Impossible 8D).
func (d Difficulty) Dice() int {
	return int(d) + 1
}

// Hasty returns the dice count when rushing the task: one level harder (+1D).
func (d Difficulty) Hasty() int {
	return d.Dice() + 1
}

// Cautious returns the dice count when taking extra care: one level easier
// (-1D), never below 1D.
func (d Difficulty) Cautious() int {
	return max(d.Dice()-1, 1)
}

// String returns the difficulty's name.
func (d Difficulty) String() string {
	if d < Easy || d > BeyondImpossible {
		return "?"
	}
	return difficultyNames[d]
}

// Resolve rolls a task: roll the difficulty's dice and succeed when the total
// is at or under target + the sum of mods, where target is typically a
// characteristic plus a skill. The returned CheckResult carries the roll and
// the success margin (Effect).
func Resolve(r *dice.Roller, d Difficulty, target int, mods ...int) dice.CheckResult {
	return r.Resolve(dice.Check{Dice: d.Dice(), Target: target + sum(mods)})
}

// ResolveDice rolls a task with an explicit dice count, for callers using a
// Hasty or Cautious pace (see Difficulty.Hasty and Difficulty.Cautious).
func ResolveDice(r *dice.Roller, numDice, target int, mods ...int) dice.CheckResult {
	return r.Resolve(dice.Check{Dice: numDice, Target: target + sum(mods)})
}

func sum(xs []int) int {
	total := 0
	for _, x := range xs {
		total += x
	}
	return total
}
