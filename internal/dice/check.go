package dice

// DefaultCheckDice is the dice count Resolve falls back to when a Check does not
// state one: Traveller assumes a 2D check unless a rule says otherwise.
//
// It is the only dice count this package names. Difficulty is task's word, not
// dice's — a Check field named Dice holds a *count*, while task.Difficulty is an
// *index* into the Book 1 p. 120 ladder, and the two numbering schemes disagree
// (task.Average is 1, but an Average check rolls 2D). This package once exported
// Easy/Average/Hard as counts, one import away from the ladder's identically
// named indices; callers wanting a difficulty's dice count ask the ladder for it
// with task.Difficulty.Dice().
const DefaultCheckDice = 2

// A Check is a roll-low task resolution. Roll Dice D6, then apply the two
// kinds of adjustment T5 distinguishes:
//
//   - Mod adjusts the Target. A positive Mod raises the Target, making success
//     easier; it is an asset.
//   - DM adjusts the die roll. A positive DM raises the roll, making success
//     harder.
//
// The check succeeds when the adjusted roll is less than or equal to the
// adjusted Target. Dice is a count of dice, not a difficulty level; zero or less
// defaults to DefaultCheckDice. See Book 1, "Mods Versus DMs" (p. 19).
type Check struct {
	Dice   int
	Target int
	Mod    int
	DM     int
}

// A CheckResult reports the outcome of resolving a Check.
//
// Success is the task's outcome after the Spectacular override (Book 1 p. 127):
// three or more ones force it true and three or more sixes force it false,
// regardless of the arithmetic. Compare Total against Target directly for the
// unoverridden comparison.
//
// Effect is the arithmetic success margin: the adjusted Target minus the
// adjusted roll. It is >= 0 when Total <= Target and negative otherwise, and
// gives the degree of success or failure. It is deliberately NOT rewritten by
// the Spectacular override — the book assigns no margin to a spectacular
// outcome, so Effect keeps reporting what the dice actually did. A caller that
// needs the two to agree should branch on Spectacular() itself.
type CheckResult struct {
	Roll    int // sum of the dice, before the DM
	Total   int // Roll + DM, compared against the target
	Target  int // Check.Target + Mod, the number to roll at or under
	Success bool
	Effect  int
	Faces   []int // the individual dice, for Spectacular detection
}

// Spectacular classifies the check's raw dice (Book 1 p. 127): three or more
// ones or sixes make a roll spectacular regardless of the pass/fail outcome.
func (c CheckResult) Spectacular() Spectacular {
	return Classify(c.Faces)
}

// Resolve rolls the check and reports the result.
func (r *Roller) Resolve(c Check) CheckResult {
	n := c.Dice
	if n <= 0 {
		n = DefaultCheckDice
	}

	faces := r.DiceFaces(n)

	roll := 0
	for _, f := range faces {
		roll += f
	}

	total := roll + c.DM
	target := c.Target + c.Mod

	return CheckResult{
		Roll:    roll,
		Total:   total,
		Target:  target,
		Success: applySpectacular(total <= target, faces),
		Effect:  target - total,
		Faces:   faces,
	}
}

// applySpectacular overrides an arithmetic pass/fail with the Spectacular
// result (Book 1 p. 127). Three ones make the task succeed "even if the result
// would otherwise be a failure"; three sixes make it "fail to produce the
// results desired".
//
// SpectacularlyInteresting (three ones AND three sixes at once, only reachable
// on 6D or more) is deliberately left alone: the book describes it as "a
// situation involving both Spectacular Success and Spectacular Failure (and a
// sign that the referee should make [the] situation a rousing, interesting
// event)" and never states whether the task itself succeeds. That is a referee
// call, so the arithmetic outcome stands and the caller can detect the case via
// CheckResult.Spectacular.
func applySpectacular(arithmetic bool, faces []int) bool {
	switch Classify(faces) {
	case SpectacularSuccess:
		return true
	case SpectacularFailure:
		return false
	default:
		// NotSpectacular, and SpectacularlyInteresting per the comment above.
		return arithmetic
	}
}
