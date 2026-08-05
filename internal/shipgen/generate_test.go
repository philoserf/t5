package shipgen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestGenerateDeterministic(t *testing.T) {
	a := Generate(dice.NewWithSeed(1))

	b := Generate(dice.NewWithSeed(1))
	if a.QSP() != b.QSP() {
		t.Errorf("same seed produced different ships: %q vs %q", a.QSP(), b.QSP())
	}
	// A generated ship is a valid design: a small hull with the three drives.
	if a.Hull.Tons < 100 || a.Hull.Tons > 800 {
		t.Errorf("generated hull %dt out of range", a.Hull.Tons)
	}

	if a.Maneuver == nil || a.Jump == nil || a.Power == nil {
		t.Errorf("generated ship missing drives: %+v", a)
	}
}

// Generate promises a feasible spec, so a generated ship carries no Problems —
// it is the one caller whose "request" is its own dice, with nobody to surprise
// by keeping it inside the rules. A Cluster hull is capped at 1G (Book 2 p.71),
// and the maneuver target used to be rolled independently of the config, so
// about one ship in fourteen came back flagged.
func TestGenerateIsAlwaysFeasible(t *testing.T) {
	seen := map[Config]int{}

	for seed := range uint64(500) {
		s := Generate(dice.NewWithSeed(seed))
		seen[s.Hull.Config]++

		if len(s.Problems) != 0 {
			t.Errorf("seed %d generated %s with problems: %v", seed, s.QSP(), s.Problems)
		}

		if s.Maneuver.Potential > s.Hull.MaxG {
			t.Errorf("seed %d: %dG maneuver in a %s hull capped at %dG",
				seed, s.Maneuver.Potential, s.Hull.Config, s.Hull.MaxG)
		}
	}
	// The clamp must not have collapsed the config roll to the roomy ones.
	if seen[Cluster] == 0 || seen[Braced] == 0 {
		t.Errorf("expected Cluster and Braced hulls in 500 rolls, saw %v", seen)
	}
}

func TestConfigByLetter(t *testing.T) {
	if c, ok := ConfigByLetter("L"); !ok || c != Lifting {
		t.Errorf("ConfigByLetter(L) = %v,%v, want Lifting", c, ok)
	}

	if c, ok := ConfigByLetter("S"); !ok || c != Streamlined {
		t.Errorf("ConfigByLetter(S) = %v,%v, want Streamlined", c, ok)
	}

	if _, ok := ConfigByLetter("Q"); ok {
		t.Errorf("ConfigByLetter(Q) should be unknown")
	}
}

func TestStructureByName(t *testing.T) {
	if s, ok := StructureByName("Shell"); !ok || s != Shell {
		t.Errorf("StructureByName(Shell) = %v,%v, want Shell", s, ok)
	}

	if s, ok := StructureByName("plate"); !ok || s != FramePlate {
		t.Errorf("StructureByName(plate) = %v,%v, want FramePlate", s, ok)
	}

	// The display name resolves too, case/space/hyphen-insensitively — not just
	// the short CLI form.
	if s, ok := StructureByName("Frame-and-Plate"); !ok || s != FramePlate {
		t.Errorf("StructureByName(Frame-and-Plate) = %v,%v, want FramePlate", s, ok)
	}

	if s, ok := StructureByName("frameandplate"); !ok || s != FramePlate {
		t.Errorf("StructureByName(frameandplate) = %v,%v, want FramePlate", s, ok)
	}

	// "frame-plate"/"frameplate" were reachable under the old two-table version
	// (via a live "frameplate" map key) — preserved here as an explicit alias.
	if s, ok := StructureByName("frame-plate"); !ok || s != FramePlate {
		t.Errorf("StructureByName(frame-plate) = %v,%v, want FramePlate", s, ok)
	}

	if _, ok := StructureByName("bogus"); ok {
		t.Errorf("StructureByName(bogus) should be unknown")
	}
}

// TestStructureNamesMatchesString checks StructureNames and String stay in
// step: every name StructureNames advertises must itself resolve back via
// StructureByName to the structure whose position it holds.
func TestStructureNamesMatchesString(t *testing.T) {
	names := StructureNames()
	if len(names) != int(Charged)+1 {
		t.Fatalf("StructureNames() has %d entries, want %d", len(names), int(Charged)+1)
	}

	for i, name := range names {
		s, ok := StructureByName(name)
		if !ok || s != Structure(i) {
			t.Errorf("StructureByName(%q) = %v,%v, want %v,true", name, s, ok, Structure(i))
		}
	}
}
