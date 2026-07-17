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
		if got := drivePotential(c.drive, c.hull); got != c.want {
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

func TestDesignDriveAvailabilityCap(t *testing.T) {
	// A Jump-E (Z1 rating 9) in Hull-A at TL-12, where availability caps Jump at
	// 3: the potential is reduced and a problem is recorded.
	j, problem := designDrive(Jump, DriveSpec{Letter: 5}, 1, 12)
	if j.Potential != 3 || problem == "" {
		t.Errorf("expected Jump capped to 3 with a problem, got %+v (%q)", j, problem)
	}
}
