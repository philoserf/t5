package shipgen

import "testing"

func TestDrivePotential(t *testing.T) {
	// The Z1 formula, verified against the book's own worked cells (p.78).
	cases := []struct{ drive, hull, want int }{
		{1, 1, 2},   // any Drive = Hull -> 2
		{18, 5, 7},  // Drive-T in Hull-E = floor(2*18/5) = 7 (book example)
		{10, 10, 2}, // Jump-K in Hull-K = 2
		{1, 2, 1},   // Drive-A in Hull-B
		{1, 3, 0},   // Drive-A in Hull-C -> not possible
		{5, 1, 9},   // capped at 9
		{26, 12, 4}, // Maneuver-N2 (26) in Hull-M (12) = floor(52/12) = 4 (Kinunir)
		{24, 12, 4}, // Jump-Z (24) in Hull-M = 4 (Kinunir)
	}
	for _, c := range cases {
		if got := drivePotential(c.drive, c.hull, 100); got != c.want {
			t.Errorf("drivePotential(%d, %d) = %d, want %d", c.drive, c.hull, got, c.want)
		}
	}
}

func TestDriveForPotential(t *testing.T) {
	// Z2 inverse: Jump-6 in a 1000t Hull-K needs Jump-Drive-Q2 (ordinal 30).
	if got := DriveForPotential(6, 10); got != 30 || driveLabel(got) != "Q2" {
		t.Errorf("DriveForPotential(6, 10) = %d (%s), want 30 (Q2)", got, driveLabel(got))
	}

	if got := DriveForPotential(2, 1); got != 1 { // Potential-2 in Hull-A -> Drive-A
		t.Errorf("DriveForPotential(2, 1) = %d, want 1", got)
	}

	if DriveForPotential(9, 24) != 0 { // beyond Z2 range
		t.Errorf("unreachable potential should return 0")
	}
}

// Every ordinal the Z2 inverse hands back has to be a drive a yard can actually
// build: a lettered size (1..24) or a "letter2" gang (an even 26..48, Book 2
// p.63 Drive Nexi). An odd extended ordinal is neither — driveLabel renders it
// "?" and driveTonsBase's >24 branch prices it as a doubling it is not.
func TestDriveForPotentialIsBuildable(t *testing.T) {
	// Jump-5 in a Hull-K: ceil(5*10/2) = 25, which is no drive at all. The
	// smallest real one is 26 = N2, and it does deliver Potential-5.
	if got := DriveForPotential(5, 10); got != 26 || driveLabel(got) != "N2" {
		t.Errorf("DriveForPotential(5, 10) = %d (%s), want 26 (N2)", got, driveLabel(got))
	}

	for potential := 1; potential <= 9; potential++ {
		for hullOrd := 1; hullOrd <= maxLetter; hullOrd++ {
			ord := DriveForPotential(potential, hullOrd)
			if ord == 0 {
				continue
			}

			if driveLabel(ord) == "?" {
				t.Errorf("DriveForPotential(%d, %d) = %d, which is not a buildable size",
					potential, hullOrd, ord)
			}

			if got := drivePotential(ord, hullOrd, 100); got < potential {
				t.Errorf("DriveForPotential(%d, %d) = %d, which only yields Potential-%d",
					potential, hullOrd, ord, got)
			}
		}
	}
}

func TestDriveTonsBase(t *testing.T) {
	cases := []struct {
		kind DriveKind
		ord  int
		tons int
	}{
		{Maneuver, 1, 2},
		{Maneuver, 24, 47},
		{Maneuver, 26, 50},
		{Maneuver, 48, 94},
		{Jump, 1, 10},
		{Jump, 24, 125},
		{Jump, 26, 140},
		{Jump, 48, 250},
		{Power, 1, 4},
		{Power, 24, 73},
		{Power, 26, 80},
		{Power, 48, 146},
	}
	for _, c := range cases {
		if got := driveTonsBase(c.kind, c.ord); got != c.tons {
			t.Errorf("driveTonsBase(%s, %d) = %d, want %d", c.kind, c.ord, got, c.tons)
		}
	}
}

func TestAvailabilityMax(t *testing.T) {
	cases := []struct {
		kind DriveKind
		tl   int
		max  int
	}{
		{Maneuver, 8, 0},
		{Maneuver, 12, 7},
		{Maneuver, 13, 9},
		{Maneuver, 20, 9},
		{Jump, 9, 1},
		{Jump, 12, 3},
		{Jump, 18, 9},
		{Power, 8, 1},
		{Power, 12, 5},
		{Power, 16, 9},
	}
	for _, c := range cases {
		if got := availabilityMax(c.kind, c.tl); got != c.max {
			t.Errorf("availabilityMax(%s, TL%d) = %d, want %d", c.kind, c.tl, got, c.max)
		}
	}
}

func TestDesignDriveMurphy(t *testing.T) {
	// Murphy Scout drives: Maneuver-A/Jump-A/Power-A (Standard) in Hull-A, TL-12.
	m, p1 := designDrive(Maneuver, DriveSpec{Letter: 1}, 1, 12)
	if p1 != "" || m.Potential != 2 || m.Tons != 2 || m.Cost != 4_000_000 {
		t.Errorf("Maneuver-A = %+v (%q), want 2G/2t/MCr4", m, p1)
	}

	j, p2 := designDrive(Jump, DriveSpec{Letter: 1}, 1, 12)
	if p2 != "" || j.Potential != 2 || j.Tons != 10 || j.Cost != 10_000_000 {
		t.Errorf("Jump-A = %+v (%q), want J2/10t/MCr10", j, p2)
	}

	pw, p3 := designDrive(Power, DriveSpec{Letter: 1}, 1, 12)
	if p3 != "" || pw.Potential != 2 || pw.Tons != 4 || pw.Cost != 4_000_000 {
		t.Errorf("Power-A = %+v (%q), want pot2/4t/MCr4", pw, p3)
	}
}

// The p.127 worked column: every stage of a J-Drive-B in a 200-ton Hull-B,
// whose Standard Drive Potential is 2. The book prints the unrounded potential
// in parentheses and the whole number beside it, plus the fuel tonnage that
// follows from it: "Potential increases or decreases with Efficiency".
func TestDesignDriveStageEfficiency(t *testing.T) {
	cases := []struct {
		stage    Stage
		exact    string // the book's parenthetical
		pot      int
		fuelTons int
	}{
		{Experimental, "50%=1.0", 1, 40},
		{Prototype, "80%=1.6", 1, 24},
		{Early, "90%=1.8", 1, 22},
		{Standard, "100%=2.0", 2, 40},
		{Basic, "90%=1.8", 1, 22},
		{Alternate, "100%=2.0", 2, 40},
		{Improved, "110%=2.2", 2, 36},
		{Generic, "90%=1.8", 1, 22},
		{Modified, "110%=2.2", 2, 36},
		{Advanced, "120%=2.4", 2, 32},
		{Ultimate, "130%=2.6", 2, 28},
	}
	for _, c := range cases {
		// TL-18 keeps availability out of the way; the book's own TL column
		// varies with the stage precisely to hold the potential at its Z1 value.
		j, _ := designDrive(Jump, DriveSpec{Letter: 2, Stage: c.stage}, 2, 18)
		if j.Potential != c.pot {
			t.Errorf("%s J-Drive-B in Hull-B: Potential %d, want %d (%s)",
				c.stage, j.Potential, c.pot, c.exact)
		}

		if got := fuel(200, j, nil, false, false).Tons; got != c.fuelTons {
			t.Errorf("%s J-Drive-B in Hull-B: fuel %dt, want %dt", c.stage, got, c.fuelTons)
		}
	}
}

// The p.76 Table X footer states the rule in EP terms: "Standard Drive-C has 300
// EP; [Experimental] Drive-C outputs 150 EP; Advanced Drive-C outputs 360 EP."
// In a Hull-C that is P = (EP/Hull)*2 = 2, 1, and 2 (=2.4) respectively.
func TestDrivePotentialEfficiencyIsEPBased(t *testing.T) {
	cases := []struct{ drive, hull, eff, want int }{
		{3, 3, 100, 2}, // 300 EP
		{3, 3, 50, 1},  // 150 EP
		{3, 3, 120, 2}, // 360 EP -> 2.4
		// Book 2 p.62: an Experimental Jump Drive-K in a Hull-K, whose Standard
		// Potential is 2, "shows 50% efficiency = Calculated Drive Potential=1".
		{10, 10, 100, 2},
		{10, 10, 50, 1},
		// p.127: "Efficiencies round down (thus Early Jump-1 at 90% becomes
		// Jump-0)." Rounding is applied once, to the EP-derived value — floor
		// first and then scaling would report 1 here.
		{1, 2, 90, 0},
		// And the scaling reaches past a floor the old two-step could not:
		// 2*5/4 = 2.5, so an Ultimate drive delivers 3, not 2.
		{5, 4, 130, 3},
		{5, 4, 100, 2},
		{5, 1, 130, 9}, // still capped at 9 (p.63 "Maximum 9")
	}
	for _, c := range cases {
		if got := drivePotential(c.drive, c.hull, c.eff); got != c.want {
			t.Errorf("drivePotential(%d, %d, %d%%) = %d, want %d",
				c.drive, c.hull, c.eff, got, c.want)
		}
	}
}

func TestDesignDriveAvailabilityCap(t *testing.T) {
	// A Jump-E (Z1 rating 9) in Hull-A at TL-12, where availability caps Jump at
	// 3: the potential is reduced and a problem is recorded.
	j, problem := designDrive(Jump, DriveSpec{Letter: 5}, 1, 12)
	if j.Potential != 3 || problem == "" {
		t.Errorf("expected Jump capped to 3 with a problem, got %+v (%q)", j, problem)
	}
}
