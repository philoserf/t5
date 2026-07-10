package systemgen

import (
	"reflect"
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/uwp"
)

func TestSystemString(t *testing.T) {
	closeStar := Star{Type: "K", Decimal: 8, Size: "VI"}
	companion := Star{Type: "M", Decimal: 6, Size: "V"}
	s := System{
		Primary:        Star{Type: "F", Decimal: 8, Size: "V"},
		Close:          &closeStar,
		CloseOrbit:     3,
		CloseCompanion: &companion,
		GasGiants:      1,
		Belts:          2,
		Worlds:         7,
		Mainworld:      uwp.Profile{Starport: 'A', Size: 7, Atmosphere: 8, Hydrographics: 8, Population: 8, Government: 9, Law: 9, TechLevel: 12},
	}
	want := "Primary: F8 V\n" +
		"Close: K8 VI (Orbit 3)\n" + // numbered orbit shown for a secondary
		"Close Companion: M6 V\n" + // companion has no orbit number
		"Gas Giants: 1  Belts: 2  Worlds: 7\n" +
		"Mainworld: A788899-C"
	if got := s.String(); got != want {
		t.Fatalf("String() =\n%q\nwant\n%q", got, want)
	}
}

func TestOrbitFormulas(t *testing.T) {
	// Close = 1D-1 (0-5), Near = 5+1D (6-11), Far = 11+1D (12-17).
	if got := closeOrbit(dice.NewScripted(1)); got != 0 {
		t.Errorf("closeOrbit(1D=1) = %d, want 0", got)
	}
	if got := closeOrbit(dice.NewScripted(6)); got != 5 {
		t.Errorf("closeOrbit(1D=6) = %d, want 5", got)
	}
	if got := nearOrbit(dice.NewScripted(1)); got != 6 {
		t.Errorf("nearOrbit(1D=1) = %d, want 6", got)
	}
	if got := nearOrbit(dice.NewScripted(6)); got != 11 {
		t.Errorf("nearOrbit(1D=6) = %d, want 11", got)
	}
	if got := farOrbit(dice.NewScripted(1)); got != 12 {
		t.Errorf("farOrbit(1D=1) = %d, want 12", got)
	}
	// Regina's worked example: Far star in Orbit 11 + 1D(=5) = 16.
	if got := farOrbit(dice.NewScripted(5)); got != 16 {
		t.Errorf("farOrbit(1D=5) = %d, want 16 (Regina)", got)
	}
}

func TestGasGiants(t *testing.T) {
	// 2D/2 - 2, floored at 0.
	cases := map[int]int{2: 0, 4: 0, 8: 2, 12: 4} // 2D total -> gas giants
	for total, want := range cases {
		// Two dice summing to total (use total/2 twice, adjusting for odd).
		a := total / 2
		b := total - a
		if got := gasGiants(dice.NewScripted(a, b)); got != want {
			t.Errorf("gasGiants(2D=%d) = %d, want %d", total, got, want)
		}
	}
}

func TestBelts(t *testing.T) {
	cases := map[int]int{1: 0, 3: 0, 4: 1, 6: 3} // 1D -> belts
	for die, want := range cases {
		if got := belts(dice.NewScripted(die)); got != want {
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
		// Compare the generated values, not their formatting, so a future
		// String() tweak can't fail a determinism test.
		if !reflect.DeepEqual(a, b) {
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
		// Present secondary stars sit in their orbit bands.
		if a.Close != nil && (a.CloseOrbit < 0 || a.CloseOrbit > 5) {
			t.Fatalf("seed %d: close orbit %d out of band 0-5", seed, a.CloseOrbit)
		}
		if a.Near != nil && (a.NearOrbit < 6 || a.NearOrbit > 11) {
			t.Fatalf("seed %d: near orbit %d out of band 6-11", seed, a.NearOrbit)
		}
		if a.Far != nil && (a.FarOrbit < 12 || a.FarOrbit > 17) {
			t.Fatalf("seed %d: far orbit %d out of band 12-17", seed, a.FarOrbit)
		}
		// A companion only exists when its star does.
		if a.CloseCompanion != nil && a.Close == nil {
			t.Fatalf("seed %d: close companion without a close star", seed)
		}
		if a.NearCompanion != nil && a.Near == nil {
			t.Fatalf("seed %d: near companion without a near star", seed)
		}
		if a.FarCompanion != nil && a.Far == nil {
			t.Fatalf("seed %d: far companion without a far star", seed)
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
