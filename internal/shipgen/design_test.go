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
	if s.Maneuver.Potential != 2 || s.Maneuver.Tons != 2 || s.Jump.Potential != 2 ||
		s.Jump.Tons != 10 ||
		s.Power.Potential != 2 {
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
		t.Errorf(
			"overtonnage agility/stability = %d/%d, want -1/0",
			s.Hull.Agility,
			s.Hull.Stability,
		)
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
		t.Fatalf(
			"potentials = maneuver %d / power %d, want 4 / 2",
			s.Maneuver.Potential,
			s.Power.Potential,
		)
	}

	if !hasProblem(s, "power plant potential") {
		t.Errorf("expected an underpowered-plant problem, got %v", s.Problems)
	}
}

func TestDesignNeedsPowerPlant(t *testing.T) {
	s := Design(
		ShipSpec{
			Mission:    "X",
			TL:         12,
			HullLetter: 1,
			Config:     Streamlined,
			Maneuver:   &DriveSpec{Letter: 1},
		},
	)
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

// Design is total (design.go): every out-of-range enum in a ShipSpec has to come
// back as a Problem, never a panic. These are the inputs that used to index the
// stage and config tables raw (GitHub #199, #219).
func TestDesignOutOfRangeSpecFields(t *testing.T) {
	base := func() ShipSpec {
		return ShipSpec{
			Mission: "X", TL: 12, HullLetter: 1, Config: Streamlined,
			Maneuver: &DriveSpec{Letter: 1}, Power: &DriveSpec{Letter: 1},
		}
	}

	tests := []struct {
		name    string
		mutate  func(*ShipSpec)
		problem string
	}{
		{"stage above the table", func(s *ShipSpec) {
			s.Maneuver.Stage = 99
		}, "stage 99"},
		{"negative stage", func(s *ShipSpec) {
			s.Maneuver.Stage = -1
		}, "stage -1"},
		{"jump stage above the table", func(s *ShipSpec) {
			s.Jump = &DriveSpec{Letter: 1, Stage: 42}
		}, "stage 42"},
		{"power stage below the table", func(s *ShipSpec) {
			s.Power.Stage = -7
		}, "stage -7"},
		{"config above the table", func(s *ShipSpec) {
			s.Config = Config(7)
		}, "configuration 7"},
		{"negative config", func(s *ShipSpec) {
			s.Config = Config(-1)
		}, "configuration -1"},
		{"hull letter above Z", func(s *ShipSpec) {
			s.HullLetter = 25
		}, "hull size 25"},
		{"hull letter of zero", func(s *ShipSpec) {
			s.HullLetter = 0
		}, "hull size 0"},
		{"negative hull letter", func(s *ShipSpec) {
			s.HullLetter = -3
		}, "hull size -3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := base()
			tt.mutate(&spec)

			s := Design(spec) // must not panic
			if !hasProblem(s, tt.problem) {
				t.Errorf("expected a %q problem, got %v", tt.problem, s.Problems)
			}
		})
	}
}

// A clamped-away bad Stage must not quietly buy the caller a Standard-stage
// drive that looks clean: the substitution is reported, and the ship it yields
// is the Standard one.
func TestDesignBadStageClampsToStandard(t *testing.T) {
	good := Design(ShipSpec{
		Mission: "X", TL: 12, HullLetter: 1, Config: Streamlined,
		Maneuver: &DriveSpec{Letter: 1, Stage: Standard},
	})
	bad := Design(ShipSpec{
		Mission: "X", TL: 12, HullLetter: 1, Config: Streamlined,
		Maneuver: &DriveSpec{Letter: 1, Stage: 99},
	})

	if bad.Maneuver.Tons != good.Maneuver.Tons || bad.Maneuver.Cost != good.Maneuver.Cost {
		t.Errorf("bad-stage drive = %+v, want the Standard %+v", bad.Maneuver, good.Maneuver)
	}

	if len(good.Problems) == len(bad.Problems) {
		t.Errorf("the substitution went unreported: %v", bad.Problems)
	}
}

// The fuel phase reads the same stage table as the drive phase (fuelarmor.go).
func TestFuelMulOutOfRangeStage(t *testing.T) {
	for _, stage := range []Stage{-1, 99} {
		if got := fuelMul(100, stage); got != 100 {
			t.Errorf("fuelMul(100, %d) = %d, want the Standard 100", stage, got)
		}
	}
}

// hull is total for the same reason Design is: it is the first thing Design calls.
func TestHullOutOfRangeConfig(t *testing.T) {
	for _, config := range []Config{-1, 7, 99} {
		h := hull(12, 1, 0, config, Shell) // must not panic
		if h.MaxG != configAttr[Cluster].maxG {
			t.Errorf("hull(config %d).MaxG = %d, want the Cluster %d",
				config, h.MaxG, configAttr[Cluster].maxG)
		}
	}
}

// TestRefusedComponentIsNotCharged locks the three accountings together: a
// component the designer refused claims no mount point, spends no tonnage, and
// costs nothing. install may already have priced it before the refusal was
// recorded, so the guard is load-bearing rather than a formality — a Nuclear
// Damper refused for Long Range was adding MCr1.6 to the ship.
func TestRefusedComponentIsNotCharged(t *testing.T) {
	bare := ShipSpec{TL: 12, HullLetter: 1}
	base := Design(bare)

	withRefused := bare
	withRefused.Defenses = []DefenseSpec{
		{Model: NuclearDamper, Mount: SingleTurret, Range: LongRange},
	}

	got := Design(withRefused)
	if len(got.Defenses) != 1 || aboard(got.Defenses[0].Problems) {
		t.Fatalf("fixture: want one refused defense, got %+v", got.Defenses)
	}

	if got.Cost != base.Cost {
		t.Errorf("refused defense added Cr%d to the ship's cost", got.Cost-base.Cost)
	}

	if got.Defenses[0].Installed() {
		t.Error("a refused defense reports itself installed, so it renders a full line")
	}
}

// TestSubADriveDoesNotRefundTonnage locks the letter floor. Book 2 p.77's "no
// drive may be smaller than the Drive-A of the class" is read as a floor on the
// size LETTER — that is what reconciles it with the worked columns on pp.104,
// 127 and 134, where stage-reduced rows print below their class's Drive-A
// tonnage while staying a Drive-B.
//
// Reading it that way removed the tonnage floor that had been enforcing it by
// accident, and nothing replaced it: driveTonsBase runs negative below A, so a
// Letter of 0 produced -1 tons and -MCr2 — a drive that ADDS budget and frees
// hull space. specProblems now reports the bad ordinal and designDrive prices it
// as the smallest real drive.
func TestSubADriveDoesNotRefundTonnage(t *testing.T) {
	for _, letter := range []int{0, -1, -5, 2 * maxLetter} {
		spec := ShipSpec{TL: 12, HullLetter: 1, Maneuver: &DriveSpec{Letter: letter}}

		got := Design(spec)
		if got.Tonnage.Used < 0 {
			t.Errorf("drive letter %d: used tonnage %d is negative", letter, got.Tonnage.Used)
		}

		if got.Cost < 0 {
			t.Errorf("drive letter %d: cost Cr%d is negative", letter, got.Cost)
		}

		if letter < 1 && !hasProblem(got, "names no drive") {
			t.Errorf("drive letter %d: no problem reported, got %v", letter, got.Problems)
		}
	}
}
