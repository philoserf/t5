package systemgen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestGasGiants(t *testing.T) {
	// 2D/2 - 2, floored at 0.
	cases := map[int]int{2: 0, 4: 0, 8: 2, 12: 4} // 2D total -> gas giants
	for total, want := range cases {
		// Two dice summing to total (use total/2 twice, adjusting for odd).
		a := total / 2
		b := total - a
		if got := gasGiants(scriptedRoller(a, b)); got != want {
			t.Errorf("gasGiants(2D=%d) = %d, want %d", total, got, want)
		}
	}
}

func TestBelts(t *testing.T) {
	cases := map[int]int{1: 0, 3: 0, 4: 1, 6: 3} // 1D -> belts
	for die, want := range cases {
		if got := belts(scriptedRoller(die)); got != want {
			t.Errorf("belts(1D=%d) = %d, want %d", die, got, want)
		}
	}
}

// TestGenerateDeterministic confirms a seeded system is reproducible and
// internally consistent: the primary always exists and world/count invariants
// hold across many seeds.
func TestGenerateDeterministic(t *testing.T) {
	for seed := uint64(1); seed <= 200; seed++ {
		a := Generate(dice.NewWithSeed(seed))
		b := Generate(dice.NewWithSeed(seed))
		if a.String() != b.String() {
			t.Fatalf("seed %d not reproducible:\n%s\n---\n%s", seed, a, b)
		}
		if a.Primary.Type == "" {
			t.Fatalf("seed %d: primary has no spectral type", seed)
		}
		if a.GasGiants < 0 || a.GasGiants > 4 {
			t.Fatalf("seed %d: gas giants %d out of range 0-4", seed, a.GasGiants)
		}
		if a.Belts < 0 || a.Belts > 3 {
			t.Fatalf("seed %d: belts %d out of range 0-3", seed, a.Belts)
		}
		// Worlds = 1 + GG + Belts + 2D, so at least 1+GG+Belts+2.
		if min := 1 + a.GasGiants + a.Belts + 2; a.Worlds < min {
			t.Fatalf("seed %d: worlds %d below floor %d", seed, a.Worlds, min)
		}
		// A companion only exists when its star does.
		if a.CloseCompanion != nil && a.Close == nil {
			t.Fatalf("seed %d: close companion without a close star", seed)
		}
	}
}

// TestPrimaryNeverBDorOB documents that a primary's single Flux (-5..+5) can
// never reach the OB (Flux -6) or BD (Flux +6..+8) rows.
func TestPrimaryNeverBDorOB(t *testing.T) {
	for seed := uint64(1); seed <= 500; seed++ {
		p := Generate(dice.NewWithSeed(seed)).Primary
		if p.Type == "BD" || p.Type == "O" || p.Type == "B" {
			t.Fatalf("seed %d: primary is %s, which should be unreachable", seed, p.Type)
		}
	}
}
