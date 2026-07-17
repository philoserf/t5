package dice

// The Many-Dice fast methods (Book 1 p. 260) resolve very large dice pools
// without rolling every die, for damage rolls that would otherwise be tens or
// hundreds of dice. They are intended for pools of at least ManyDiceMin; the
// functions themselves are permissive and return 0 for n <= 0. The caller
// chooses the method and rounds the averaging results as the situation requires.

// ManyDiceMin is the pool size at or above which the Many-Dice methods apply
// (Book 1 p. 260: "at least 11 Dice").
const ManyDiceMin = 11

// ManyDice10 sums nD by rolling ten dice once and reusing them cyclically —
// die 1 also serves rolls 11, 21, … (Book 1 p. 260, "Many Dice 10"). Exact.
func (r *Roller) ManyDice10(n int) int {
	if n <= 0 {
		return 0
	}

	return cycleSum(r.DiceFaces(10), n)
}

// ManyDice2D sums nD by rolling k = 2D dice as a subsample and reusing those k
// faces cyclically over n (Book 1 p. 260, "Many Dice 2D"). Exact.
func (r *Roller) ManyDice2D(n int) int {
	if n <= 0 {
		return 0
	}

	return cycleSum(r.DiceFaces(r.Dice(2)), n)
}

// cycleSum totals n dice by reusing the given faces cyclically (face i%len).
func cycleSum(faces []int, n int) int {
	sum := 0
	for i := range n {
		sum += faces[i%len(faces)]
	}

	return sum
}

// Average35 returns n × 3.5, the expected total of nD (Book 1 p. 260, "Many Dice
// 3.5"). No dice are rolled.
func Average35(n int) float64 {
	if n <= 0 {
		return 0
	}

	return float64(n) * 3.5
}

// ManyDice35Flux returns n times the Flux-adjusted per-die average (7+Flux)/2
// (Book 1 p. 260, "Many Dice 3.5 Flux"): Flux −5 → ×1 … 0 → ×3.5 … +5 → ×6.
func (r *Roller) ManyDice35Flux(n int) float64 {
	if n <= 0 {
		return 0
	}

	return (7.0 + float64(r.Flux())) / 2.0 * float64(n)
}
