// Package dice implements the Traveller5 (T5) dice mechanics.
//
// Traveller uses only six-sided dice (D6). The notation nD means "roll n
// six-sided dice and sum them"; an optional +k or -k adjusts the total.
// Beyond plain sums, T5 defines Flux (D-D, ranging -5..+5), its Good and Bad
// variants, the half-die (D/2), and several "even distributions" that contort
// D6 results into ranges like 1-9 or 0-9.
//
// All randomness flows through a Roller so callers can seed it for
// reproducible results. See Book 1, "Traveller Uses Dice" (pp. 18-19) and the
// Dice Appendix (pp. 253-260).
package dice

import "math/rand/v2"

// A Roller is the single source of die rolls. Construct one with New (auto
// seeded) or NewWithSeed (deterministic, for tests and reproducible worlds).
// The zero value is not usable.
type Roller struct {
	// d6 returns a single die result in 1..6. Held as a func so tests can
	// substitute a scripted sequence.
	d6 func() int
}

// New returns a Roller seeded from the runtime's random source.
func New() *Roller {
	return NewWithSeed(rand.Uint64())
}

// NewWithSeed returns a Roller whose rolls are fully determined by seed.
// The same seed always produces the same sequence of rolls.
func NewWithSeed(seed uint64) *Roller {
	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	return NewSource(func() int { return rng.IntN(6) + 1 })
}

// NewSource returns a Roller that draws each die from next, which must return
// values in 1..6. It lets callers supply a custom die source — a replay log or
// a scripted sequence for deterministic testing — in place of the built-in
// generator.
func NewSource(next func() int) *Roller {
	return &Roller{d6: next}
}

// NewScripted returns a Roller that yields the given die faces in order,
// cycling back to the start once they are exhausted. It is a convenience over
// NewSource for deterministic tests and replay, and panics if no faces are
// given.
func NewScripted(faces ...int) *Roller {
	if len(faces) == 0 {
		panic("dice: NewScripted needs at least one face")
	}
	i := 0
	return NewSource(func() int {
		v := faces[i%len(faces)]
		i++
		return v
	})
}

// Die rolls a single D6, returning 1..6.
func (r *Roller) Die() int {
	return r.d6()
}

// Dice rolls nD: the sum of n six-sided dice, ranging n..6n. It returns 0 for
// n <= 0.
func (r *Roller) Dice(n int) int {
	sum := 0
	for range n {
		sum += r.d6()
	}
	return sum
}

// DiceFaces rolls nD and returns the individual faces in roll order (nil for
// n <= 0). The sum of DiceFaces(n) equals Dice(n); DiceFaces additionally
// preserves each face, which Spectacular detection (Book 1 p. 127) and the
// Genetics gene (the first face, Book 1 p. 102) need. Dice stays allocation-free
// for the hot paths that only want the sum.
func (r *Roller) DiceFaces(n int) []int {
	if n <= 0 {
		return nil
	}
	faces := make([]int, n)
	for i := range faces {
		faces[i] = r.d6()
	}
	return faces
}

// Flux rolls the Flux die: one die minus a second, ranging -5..+5. It is
// identical in output to 2D-7 and to D-D. By convention the first (light) die
// is positive and the second (dark) die is subtracted.
func (r *Roller) Flux() int {
	return r.d6() - r.d6() //nolint:staticcheck // SA4000: d6 is stateful; two distinct rolls
}

// GoodFlux rolls two dice and subtracts the smaller from the larger, ranging
// 0..+5 (0 when the dice are equal).
func (r *Roller) GoodFlux() int {
	a, b := r.d6(), r.d6()
	return max(a, b) - min(a, b)
}

// BadFlux rolls two dice and subtracts the larger from the smaller, ranging
// -5..0 (0 when the dice are equal).
func (r *Roller) BadFlux() int {
	return -r.GoodFlux()
}

// FluxIndex maps a Flux value (-6..+6) to a 0..12 table index, clamping
// out-of-range inputs. Tables indexed by Flux (13 entries, Flux -6 first) use it.
func FluxIndex(flux int) int {
	return min(max(flux+6, 0), 12)
}

// HalfDie rolls D/2, rounding up (always in the rolling player's favor), and
// so returns 1..3.
func (r *Roller) HalfDie() int {
	return (r.d6() + 1) / 2
}
