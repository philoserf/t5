package dice

import (
	"slices"
	"testing"
)

// scripted is a package-local alias for NewScripted, kept for the many
// exact-outcome tests below.
func scripted(vals ...int) *Roller {
	return NewScripted(vals...)
}

// TestScriptedNeedsFaces covers the empty-script guard.
func TestScriptedNeedsFaces(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("NewScripted() with no faces did not panic")
		}
	}()

	NewScripted()
}

// TestScriptedRejectsNonFaces is #237: a script is validated eagerly, so a
// typo'd or off-by-one face fails at the offending line rather than flowing on
// as a legal-looking roll and silently shifting a golden's expected value.
func TestScriptedRejectsNonFaces(t *testing.T) {
	for _, bad := range [][]int{{0}, {7}, {-1}, {3, 4, 7}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("NewScripted(%v) accepted a non-die face", bad)
				}
			}()

			NewScripted(bad...)
		}()
	}
}

// TestScriptedPanicsWhenExhausted is #289: the script is exact, not cyclic.
// Wrapping would let a change in dice consumption pass every golden test
// unnoticed, which is precisely what a scripted Roller exists to prevent.
func TestScriptedPanicsWhenExhausted(t *testing.T) {
	r := scripted(1, 2)
	if got := r.Dice(2); got != 3 {
		t.Fatalf("Dice(2) = %d, want 3", got)
	}

	defer func() {
		if recover() == nil {
			t.Error("drawing past the end of the script did not panic")
		}
	}()

	r.Die()
}

// TestFixedRejectsNonFaces mirrors TestScriptedRejectsNonFaces: Fixed
// validates its face eagerly, at construction, not at first roll.
func TestFixedRejectsNonFaces(t *testing.T) {
	for _, bad := range []int{0, 7, -1} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Fixed(%d) accepted a non-die face", bad)
				}
			}()

			Fixed(bad)
		}()
	}
}

// TestFixedRepeatsForever distinguishes Fixed from NewScripted: unlike a
// script, it never exhausts.
func TestFixedRepeatsForever(t *testing.T) {
	r := Fixed(6)

	for range 100 {
		if got := r.Die(); got != 6 {
			t.Fatalf("Die() = %d, want 6", got)
		}
	}
}

func TestDie(t *testing.T) {
	r := scripted(4)
	if got := r.Die(); got != 4 {
		t.Fatalf("Die() = %d, want 4", got)
	}
}

func TestDice(t *testing.T) {
	r := scripted(1, 2, 3)
	if got := r.Dice(3); got != 6 {
		t.Fatalf("Dice(3) = %d, want 6", got)
	}

	if got := scripted(6).Dice(0); got != 0 {
		t.Fatalf("Dice(0) = %d, want 0", got)
	}
}

func TestFlux(t *testing.T) {
	if got := scripted(2, 5).Flux(); got != -3 {
		t.Fatalf("Flux() = %d, want -3", got)
	}

	if got := scripted(6, 1).Flux(); got != 5 {
		t.Fatalf("Flux() = %d, want 5", got)
	}
}

func TestGoodAndBadFlux(t *testing.T) {
	if got := scripted(2, 5).GoodFlux(); got != 3 {
		t.Fatalf("GoodFlux() = %d, want 3", got)
	}

	if got := scripted(4, 4).GoodFlux(); got != 0 {
		t.Fatalf("GoodFlux() equal dice = %d, want 0", got)
	}

	if got := scripted(2, 5).BadFlux(); got != -3 {
		t.Fatalf("BadFlux() = %d, want -3", got)
	}
}

func TestHalfDie(t *testing.T) {
	// Round up: 1,2->1  3,4->2  5,6->3
	want := map[int]int{1: 1, 2: 1, 3: 2, 4: 2, 5: 3, 6: 3}
	for face, exp := range want {
		if got := scripted(face).HalfDie(); got != exp {
			t.Errorf("HalfDie() face %d = %d, want %d", face, got, exp)
		}
	}
}

// Range tests over a seeded roller guard the statistical bounds of each
// primitive without depending on a specific sequence.
func TestPrimitiveRanges(t *testing.T) {
	r := NewWithSeed(20260710)
	for range 5000 {
		if v := r.Die(); v < 1 || v > 6 {
			t.Fatalf("Die() out of range: %d", v)
		}

		if v := r.Dice(3); v < 3 || v > 18 {
			t.Fatalf("Dice(3) out of range: %d", v)
		}

		if v := r.Flux(); v < -5 || v > 5 {
			t.Fatalf("Flux() out of range: %d", v)
		}

		if v := r.GoodFlux(); v < 0 || v > 5 {
			t.Fatalf("GoodFlux() out of range: %d", v)
		}

		if v := r.BadFlux(); v < -5 || v > 0 {
			t.Fatalf("BadFlux() out of range: %d", v)
		}

		if v := r.HalfDie(); v < 1 || v > 3 {
			t.Fatalf("HalfDie() out of range: %d", v)
		}
	}
}

func TestSeedIsDeterministic(t *testing.T) {
	a, b := NewWithSeed(42), NewWithSeed(42)
	for range 100 {
		if x, y := a.Dice(2), b.Dice(2); x != y {
			t.Fatalf("same seed diverged: %d != %d", x, y)
		}
	}
}

// TestSeedIsRecoverable: a Roller can always name the seed it was built from,
// so output from an auto-seeded run can be reproduced. Only a Roller drawing
// from a caller-supplied sequence has no seed to name.
func TestSeedIsRecoverable(t *testing.T) {
	if got, ok := NewWithSeed(42).Seed(); !ok || got != 42 {
		t.Errorf("NewWithSeed(42).Seed() = %d, %v; want 42, true", got, ok)
	}

	// New draws its own seed, and replaying it reproduces the rolls.
	auto := New()

	seed, ok := auto.Seed()
	if !ok {
		t.Fatal("New().Seed() reported no seed; an auto-seeded run must stay reproducible")
	}

	replay := NewWithSeed(seed)
	for i := range 100 {
		if x, y := auto.Dice(2), replay.Dice(2); x != y {
			t.Fatalf("replay of seed %d diverged at roll %d: %d != %d", seed, i, x, y)
		}
	}

	// A scripted or sourced Roller has no seed and must say so rather than
	// reporting a plausible zero.
	if _, ok := NewScripted(1).Seed(); ok {
		t.Error("NewScripted().Seed() claimed a seed")
	}

	if _, ok := NewSource(func() int { return 1 }).Seed(); ok {
		t.Error("NewSource().Seed() claimed a seed")
	}
}

// TestDeriveIsDeterministicAndOrderIndependentOfParentDraws is the core
// substream contract: a child stream is a function of (parent seed,
// discriminators) alone, not of how many rolls the parent has taken. Deriving
// after 0, 1, or 50 parent draws yields byte-identical children.
func TestDeriveIsDeterministicAndOrderIndependentOfParentDraws(t *testing.T) {
	const seed = 42

	want := seq(NewWithSeed(seed).Derive(3, 7), 30)

	for _, warm := range []int{0, 1, 50} {
		parent := NewWithSeed(seed)
		for range warm {
			parent.Die()
		}

		if got := seq(parent.Derive(3, 7), 30); !slices.Equal(got, want) {
			t.Errorf("child after %d parent draws diverged from child after 0", warm)
		}
	}
}

// TestDeriveDiscriminatorsSeparateStreams: distinct discriminators (including
// permutations) give independent streams, so per-entity keys don't collide.
func TestDeriveDiscriminatorsSeparateStreams(t *testing.T) {
	base := NewWithSeed(42)
	keys := [][]uint64{{1, 1}, {1, 2}, {2, 1}, {2, 2}, {1}, {}}

	seqs := make([][]int, len(keys))
	for i, k := range keys {
		seqs[i] = seq(base.Derive(k...), 40)
	}

	for i := range seqs {
		for j := i + 1; j < len(seqs); j++ {
			if slices.Equal(seqs[i], seqs[j]) {
				t.Errorf("Derive(%v) and Derive(%v) produced identical streams", keys[i], keys[j])
			}
		}
	}
}

// TestDerivePanicsOnUnseeded: a scripted/sourced Roller has no seed to key from.
func TestDerivePanicsOnUnseeded(t *testing.T) {
	for name, r := range map[string]*Roller{
		"scripted": NewScripted(1),
		"sourced":  NewSource(func() int { return 1 }),
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s Roller.Derive did not panic", name)
				}
			}()

			r.Derive(1)
		}()
	}
}

func seq(r *Roller, n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = r.Die()
	}

	return out
}
