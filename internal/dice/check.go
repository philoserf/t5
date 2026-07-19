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
// Success is the plain arithmetic outcome, Total <= Target. It is NOT the Book 1
// p. 127 Spectacular override: the book calls that a property of a *task* result
// ("Sometimes the task result is Spectacular"), and pp. 120-131 belong to
// internal/task, so task.Resolve and task.ResolveDice apply it and this
// primitive tier does not. A caller resolving a task goes through task and gets
// the override; a caller rolling dice against a number gets arithmetic. Faces
// and Spectacular below make the classification available either way — dice
// reports what the dice did, task decides what it means.
//
// Effect is the arithmetic success margin: the adjusted Target minus the
// adjusted roll. It is >= 0 when Total <= Target and negative otherwise, and
// gives the degree of success or failure. The task layer leaves it alone when it
// overrides Success — the book assigns no margin to a spectacular outcome, so
// Effect keeps reporting what the dice actually did. A caller that needs the two
// to agree should branch on Spectacular() itself.
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

// Resolve rolls the check and reports the arithmetic result. It does not apply
// the p. 127 Spectacular override; task.Resolve does. See CheckResult.
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
		Success: total <= target,
		Effect:  target - total,
		Faces:   faces,
	}
}
