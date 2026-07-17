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

	if _, ok := StructureByName("bogus"); ok {
		t.Errorf("StructureByName(bogus) should be unknown")
	}
}
