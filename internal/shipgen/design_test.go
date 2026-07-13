package shipgen

import (
	"strings"
	"testing"
)

// murphySpec is the design input for the Book 2 p.42 Murphy-class Scout.
func murphySpec() ShipSpec {
	return ShipSpec{
		Name: "Murphy-class Scout", Mission: "S", TL: 12,
		HullLetter: 1, Config: Lifting, Structure: Shell, ArmorLayers: 2,
		Maneuver: &DriveSpec{Letter: 1}, Jump: &DriveSpec{Letter: 1}, Power: &DriveSpec{Letter: 1},
		FuelScoop: true, FuelPurifier: true,
	}
}

func TestDesignMurphy(t *testing.T) {
	s := Design(murphySpec())

	if got := s.QSP(); got != "S-AL22" {
		t.Errorf("QSP = %q, want S-AL22", got)
	}
	if len(s.Problems) != 0 {
		t.Errorf("unexpected problems: %v", s.Problems)
	}
	// Hull.
	if s.Hull.Tons != 100 || s.Hull.Cost != 16_000_000 || s.Hull.BaseArmor != 6 {
		t.Errorf("hull = %+v", s.Hull)
	}
	// Drives: 2G / Jump-2 / power tier 2.
	if s.Maneuver.Potential != 2 || s.Maneuver.Tons != 2 || s.Jump.Potential != 2 || s.Jump.Tons != 10 || s.Power.Potential != 2 {
		t.Errorf("drives wrong: %+v %+v %+v", s.Maneuver, s.Jump, s.Power)
	}
	// Fuel and armor.
	if s.Fuel.Tons != 22 || s.Armor.AV != 6 || s.Armor.Tons != 2 {
		t.Errorf("fuel/armor: fuel %dt, armor AV %d %dt", s.Fuel.Tons, s.Armor.AV, s.Armor.Tons)
	}
	// Budget: 16t drives + 22t fuel + 2t fittings + 2t armor = 42t used, 58t
	// left for the (deferred) accommodations/weapons/cargo payload.
	if s.Tonnage.Used != 42 || s.Tonnage.Payload != 58 {
		t.Errorf("budget = %+v, want 42 used / 58 payload", s.Tonnage)
	}
	// Core cost = hull 16 + drives 18 + fuel 1.111 MCr. (The book's MCr 70.3
	// full-ship figure additionally includes the deferred staterooms, turret,
	// and sensors.)
	if s.Cost != 16_000_000+18_000_000+1_111_000 {
		t.Errorf("core cost = %d, want %d", s.Cost, 35_111_000)
	}
}

func TestDesignBeowulf(t *testing.T) {
	// Beowulf-class Free Trader (Book 2 p.45): 222t on a 200t Hull-B,
	// Streamlined, TL-10, Maneuver-A + Jump-A + Power-A -> 1G / Jump-1, QSP
	// A-BS11. Exercises the minor-overtonnage path (Agility -1, Stability 0).
	s := Design(ShipSpec{
		Name: "Beowulf-class Free Trader", Mission: "A", TL: 10,
		HullLetter: 2, Tons: 222, Config: Streamlined,
		Maneuver: &DriveSpec{Letter: 1}, Jump: &DriveSpec{Letter: 1}, Power: &DriveSpec{Letter: 1},
	})
	if got := s.QSP(); got != "A-BS11" {
		t.Errorf("QSP = %q, want A-BS11", got)
	}
	if s.Hull.Agility != -1 || s.Hull.Stability != 0 {
		t.Errorf("overtonnage agility/stability = %d/%d, want -1/0", s.Hull.Agility, s.Hull.Stability)
	}
	if s.Maneuver.Potential != 1 || s.Jump.Potential != 1 {
		t.Errorf("potentials = %dG / J%d, want 1G / J1", s.Maneuver.Potential, s.Jump.Potential)
	}
}

func TestDesignOverBudget(t *testing.T) {
	// A Jump-Z drive (125t) will not fit in a 100t Hull-A.
	s := Design(ShipSpec{
		Mission: "X", TL: 18, HullLetter: 1, Config: Streamlined,
		Jump: &DriveSpec{Letter: 24}, Power: &DriveSpec{Letter: 24},
	})
	if s.Tonnage.Payload >= 0 {
		t.Errorf("payload = %d, want negative (over budget)", s.Tonnage.Payload)
	}
	if !hasProblem(s, "over budget") {
		t.Errorf("expected an over-budget problem, got %v", s.Problems)
	}
}

func TestDesignUnderpoweredPlant(t *testing.T) {
	// Maneuver-G in Hull-C gives 4G, but Power-D only reaches tier 2.
	s := Design(ShipSpec{
		Mission: "X", TL: 13, HullLetter: 3, Config: Streamlined,
		Maneuver: &DriveSpec{Letter: 7}, Power: &DriveSpec{Letter: 4},
	})
	if s.Maneuver.Potential != 4 || s.Power.Potential != 2 {
		t.Fatalf("potentials = maneuver %d / power %d, want 4 / 2", s.Maneuver.Potential, s.Power.Potential)
	}
	if !hasProblem(s, "power plant potential") {
		t.Errorf("expected an underpowered-plant problem, got %v", s.Problems)
	}
}

func TestDesignNeedsPowerPlant(t *testing.T) {
	s := Design(ShipSpec{Mission: "X", TL: 12, HullLetter: 1, Config: Streamlined, Maneuver: &DriveSpec{Letter: 1}})
	if !hasProblem(s, "require a power plant") {
		t.Errorf("expected a missing-plant problem, got %v", s.Problems)
	}
}

func hasProblem(s Ship, substr string) bool {
	for _, p := range s.Problems {
		if strings.Contains(p, substr) {
			return true
		}
	}
	return false
}
