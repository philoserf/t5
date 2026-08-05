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

import (
	"fmt"
	"math/rand/v2"
)

// A Roller is the single source of die rolls. Construct one with New (auto
// seeded) or NewWithSeed (deterministic, for tests and reproducible worlds).
// The zero value is not usable.
type Roller struct {
	// d6 returns a single die result in 1..6. Held as a func so tests can
	// substitute a scripted sequence.
	d6 func() int

	// seed is the value the generator was built from, kept so that even an
	// auto-seeded Roller can name it; seeded is false for a Roller drawing from
	// a caller-supplied sequence, which has no seed to name.
	seed   uint64
	seeded bool
}

// New returns a Roller seeded from the runtime's random source. The drawn seed
// is kept and reported by Seed, so output from an auto-seeded Roller stays
// reproducible — never draw a seed and discard it.
func New() *Roller {
	return NewWithSeed(rand.Uint64())
}

// NewWithSeed returns a Roller whose rolls are fully determined by seed.
// The same seed always produces the same sequence of rolls.
func NewWithSeed(seed uint64) *Roller {
	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))

	r := NewSource(func() int { return rng.IntN(6) + 1 })
	r.seed, r.seeded = seed, true

	return r
}

// NewSource returns a Roller that draws each die from next, which must return
// values in 1..6. It lets callers supply a custom die source — a replay log or
// a scripted sequence for deterministic testing — in place of the built-in
// generator.
func NewSource(next func() int) *Roller {
	return &Roller{d6: next}
}

// NewScripted returns a Roller that yields the given die faces in order. It is
// a convenience over NewSource for deterministic tests and replay.
//
// The script is exact, not cyclic: drawing past its end panics rather than
// wrapping. A scripted Roller exists to pin the dice a golden test consumes, so
// a test that outruns its script no longer describes the rolls being made —
// silently serving it recycled faces would let a change in dice consumption
// pass unnoticed. For the same reason the faces are validated eagerly: every
// entry must be a real die face in 1..6, and a bad one panics at construction,
// pointing at the offending script rather than at whichever roll later drew it.
//
// It panics if no faces are given.
func NewScripted(faces ...int) *Roller {
	if len(faces) == 0 {
		panic("dice: NewScripted needs at least one face")
	}

	for i, f := range faces {
		if f < 1 || f > 6 {
			panic(fmt.Sprintf("dice: NewScripted face %d at index %d is not a die face (want 1..6)", f, i))
		}
	}

	i := 0

	return NewSource(func() int {
		if i >= len(faces) {
			panic(fmt.Sprintf("dice: NewScripted script exhausted after %d faces; "+
				"the test consumed more dice than it scripted, so the script no longer "+
				"describes the rolls being made — extend it to cover every roll", len(faces)))
		}

		v := faces[i]
		i++

		return v
	})
}

// Fixed returns a Roller where every die shows face, forever. It pins one
// variable in a property sweep — e.g. holding a die at 6 while another
// Roller varies the rest — where NewScripted's exact, finite script would be
// the wrong shape.
//
// It panics if face is not a real die face in 1..6.
func Fixed(face int) *Roller {
	if face < 1 || face > 6 {
		panic(fmt.Sprintf("dice: Fixed face %d is not a die face (want 1..6)", face))
	}

	return NewSource(func() int { return face })
}

// Seed reports the seed the Roller was built from and whether it has one. A
// Roller from NewSource or NewScripted draws from a supplied sequence rather
// than a seeded generator, so it reports ok false.
func (r *Roller) Seed() (uint64, bool) {
	return r.seed, r.seeded
}

// Derive returns an independent child Roller whose stream is a deterministic
// function of this Roller's seed and the given discriminators, and of nothing
// else — in particular, not of how many rolls this Roller has already made.
//
// It exists to give each entity its own substream. Keying a hex on its
// coordinates (r.Derive(uint64(col), uint64(row))), or a character on its index,
// makes that entity regenerable in isolation and immune to a draw-count change
// anywhere else built from the same parent seed: a rule fix that adds or removes
// a roll for one entity no longer shifts every entity generated after it. Two
// children built from the same discriminators are identical; children with
// different discriminators are independent.
//
// Derive requires a seeded parent (from New or NewWithSeed). A Roller built from
// NewSource or NewScripted has no seed to key from, so Derive panics on one — a
// scripted test that wants a substream should script that substream directly.
func (r *Roller) Derive(discriminators ...uint64) *Roller {
	if !r.seeded {
		panic("dice: Derive needs a seeded Roller (New/NewWithSeed), not a scripted or sourced one")
	}

	// splitmix64 finalizer, folded once per discriminator, so every bit of every
	// discriminator avalanches through the whole child seed and permutations like
	// (1,2) and (2,1) diverge.
	h := mix64(r.seed)
	for _, d := range discriminators {
		h = mix64(h ^ d)
	}

	return NewWithSeed(h)
}

// mix64 is the splitmix64 finalizer: a bijection on uint64 with strong
// avalanche, so a one-bit input change flips about half the output bits.
func mix64(z uint64) uint64 {
	z += 0x9e3779b97f4a7c15
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb

	return z ^ (z >> 31)
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
