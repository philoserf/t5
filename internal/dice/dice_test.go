package dice

import "testing"

// scripted returns a Roller whose dice come from vals in order, cycling when
// exhausted, for exact-outcome tests.
func scripted(vals ...int) *Roller {
	i := 0
	return &Roller{d6: func() int {
		v := vals[i%len(vals)]
		i++
		return v
	}}
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
