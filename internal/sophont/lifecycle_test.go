package sophont

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

// TestHumanLifespan is the benchmark (Book 3 p.231): with every stage at Flux 0
// (2 terms each), a Human lives 2 years of infancy plus nine 8-year stages = 74.
func TestHumanLifespan(t *testing.T) {
	lc := rollLifeCycle(dice.NewScripted(fluxSeq(0, 0, 0, 0, 0, 0, 0, 0, 0)...))
	if lc.Lifespan != 74 {
		t.Errorf("Human lifespan = %d, want 74", lc.Lifespan)
	}
	for stage := 1; stage <= 9; stage++ {
		if lc.Terms[stage] != 2 {
			t.Errorf("stage %d terms = %d, want 2", stage, lc.Terms[stage])
		}
	}
}

// TestLifeStageDuration exercises chart 09A, including the irregular Flux -5 row
// where only stages 1, 5, and 9 last a term.
func TestLifeStageDuration(t *testing.T) {
	cases := []struct {
		flux, stage, want int
	}{
		{-5, 1, 1},
		{-5, 2, 0},
		{-5, 5, 1},
		{-5, 4, 0},
		{-5, 9, 1}, // short-lived row
		{-4, 3, 1},
		{-2, 7, 1},
		{-1, 4, 2},
		{0, 6, 2},
		{1, 2, 2},
		{2, 8, 3},
		{3, 1, 3},
		{4, 5, 4},
		{5, 9, 6},
	}
	for _, c := range cases {
		if got := lifeStageDuration(c.flux, c.stage); got != c.want {
			t.Errorf(
				"lifeStageDuration(flux %+d, stage %d) = %d, want %d",
				c.flux,
				c.stage,
				got,
				c.want,
			)
		}
	}
}

// TestShortLivedSpecies confirms the extreme case: an all-Flux-(-5) species
// lives only its infancy plus three single-term stages = 2 + 12 = 14 years.
func TestShortLivedSpecies(t *testing.T) {
	lc := rollLifeCycle(dice.NewScripted(fluxSeq(-5, -5, -5, -5, -5, -5, -5, -5, -5)...))
	if lc.Lifespan != 14 {
		t.Errorf("short-lived lifespan = %d, want 14", lc.Lifespan)
	}
}
