package dice

// Difficulty is the number of dice a Check rolls. Traveller assumes an Average
// (2D) check unless a rule states otherwise; Easy is 1D and Hard is 3D. Other
// dice counts are specified directly (e.g. a 5D or 10D check).
const (
	Easy    = 1
	Average = 2
	Hard    = 3
)

// A Check is a roll-low task resolution. Roll Dice D6, then apply the two
// kinds of adjustment T5 distinguishes:
//
//   - Mod adjusts the Target. A positive Mod raises the Target, making success
//     easier; it is an asset.
//   - DM adjusts the die roll. A positive DM raises the roll, making success
//     harder.
//
// The check succeeds when the adjusted roll is less than or equal to the
// adjusted Target. A Dice count of zero or less defaults to Average (2D). See Book 1,
// "Mods Versus DMs" (p. 19).
type Check struct {
	Dice   int
	Target int
	Mod    int
	DM     int
}

// A CheckResult reports the outcome of resolving a Check. Effect is the
// success margin: the adjusted Target minus the adjusted roll. It is >= 0 on
// success and negative on failure, and gives the degree of success or failure.
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
		n = Average
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
