// Package task implements Traveller5's Universal Task Format (Book 1, Tasks
// pp. 120-131). A task rolls a number of dice set by its Difficulty and
// succeeds when the total is at or under a target number built from a
// characteristic, a skill, and modifiers — the roll-low resolution the dice
// engine already provides.
package task

import (
	"fmt"

	"github.com/philoserf/t5/internal/dice"
)

// Difficulty is a task's difficulty level (Book 1 p. 129). It sets how many
// dice are rolled: more dice make a low roll harder, so higher difficulty is
// harder. The ladder runs Easy (1D) through Beyond Impossible (8D).
type Difficulty int

// Task difficulties (Book 1 p. 120).
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
// Beyond Impossible 8D). It panics on a Difficulty outside the ladder.
//
// This is the strict path, not a display one. The ladder is a closed set of
// eight constants and the book defines no task above or below it, so a value
// outside it is a programming error. Returning a count anyway would conceal
// that error rather than surface it: a non-positive count falls through to
// dice.Resolve's 2D default, so an off-ladder difficulty would quietly resolve
// as an ordinary Average check. Compare ehex.Digit and Character.Score, which
// panic for the same reason, against the display paths (String below,
// ehex.Format) that render an out-of-domain value as a visible "?".
func (d Difficulty) Dice() int {
	if d < Easy || d > BeyondImpossible {
		panic(fmt.Sprintf("task: difficulty %d out of range %d..%d",
			int(d), int(Easy), int(BeyondImpossible)))
	}

	return int(d) + 1
}

// Hasty returns the dice count when rushing the task: one level harder (+1D).
func (d Difficulty) Hasty() int {
	return d.Dice() + 1
}

// ExtraHasty returns the dice count for a task rushed ahead of every competing
// non-Extra-Hasty task: two levels harder (+2D, Book 1 p. 125).
func (d Difficulty) ExtraHasty() int {
	return d.Dice() + 2
}

// Cautious returns the dice count when taking extra care: one level easier
// (-1D), never below 1D.
func (d Difficulty) Cautious() int {
	return max(d.Dice()-1, 1)
}

// String returns the difficulty's name, or "?" for a value outside the ladder.
// Unlike Dice it never panics: it is a display path, and an fmt.Stringer must be
// total so that logging a bad value can report it rather than crash.
func (d Difficulty) String() string {
	if d < Easy || d > BeyondImpossible {
		return "?"
	}

	return difficultyNames[d]
}

// Resolve rolls a task: roll the difficulty's dice and succeed when the total
// is at or under target + the sum of mods, where target is typically a
// characteristic plus a skill. The returned CheckResult carries the roll and
// the success margin (Effect), with Success after the Spectacular override.
func Resolve(r *dice.Roller, d Difficulty, target int, mods ...int) dice.CheckResult {
	// One resolution path, so the Spectacular override, the mod arithmetic and the
	// Check construction cannot drift apart between the two entry points. Dice()
	// panics off-ladder and otherwise yields 1..8, so ResolveDice's floor is a
	// no-op here.
	return ResolveDice(r, d.Dice(), target, mods...)
}

// ResolveDice rolls a task with an explicit dice count, for callers using a
// Hasty or Cautious pace (see Difficulty.Hasty and Difficulty.Cautious, which
// always return at least 1). A count below 1 is treated as 1D rather than
// falling through to the roll-low default. Like Resolve, it applies the
// Spectacular override.
func ResolveDice(r *dice.Roller, numDice, target int, mods ...int) dice.CheckResult {
	return applySpectacular(r.Resolve(dice.Check{Dice: max(numDice, 1), Target: target + sum(mods)}))
}

// applySpectacular overrides a task's arithmetic pass/fail with its Spectacular
// result (Book 1 p. 127). Three ones make the task succeed "even if the result
// would otherwise be a failure"; three sixes make it "fail to produce the
// results desired".
//
// This lives here, not in dice, because p. 127 states the rule about tasks —
// "Sometimes the task result is Spectacular" — and this package owns
// pp. 120-131. dice keeps what is genuinely a dice observation (counting ones
// and sixes on the faces, dice.Classify) and this layer keeps the consequence.
// A caller rolling a dice.Check that is not a task therefore gets arithmetic,
// with no flag to remember.
//
// Effect is deliberately left arithmetic: the book assigns no margin to a
// spectacular outcome, so Effect keeps reporting what the dice did.
//
// SpectacularlyInteresting (three ones AND three sixes at once, only reachable
// on 6D or more) is deliberately left alone: the book describes it as "a
// situation involving both Spectacular Success and Spectacular Failure (and a
// sign that the referee should make [the] situation a rousing, interesting
// event)" and never states whether the task itself succeeds. That is a referee
// call, so the arithmetic outcome stands and the caller can detect the case via
// CheckResult.Spectacular.
func applySpectacular(res dice.CheckResult) dice.CheckResult {
	switch res.Spectacular() {
	case dice.SpectacularSuccess:
		res.Success = true
	case dice.SpectacularFailure:
		res.Success = false
	case dice.NotSpectacular, dice.SpectacularlyInteresting:
		// Arithmetic stands, per the comment above.
	}

	return res
}

func sum(xs []int) int {
	total := 0
	for _, x := range xs {
		total += x
	}

	return total
}
