package shipgen

import (
	"strings"
	"testing"
)

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

// The p.127 X-table worked column, cell by cell: a J-Drive-B in a 200-ton
// Hull-B, "Variations based on Standard P-Plant-B TL-9 ... with Drive
// Potential=2", at MCr1 per ton against a Jump-B base of 15 tons. This is the
// column that settles both Book-2 drive-table conflicts at once, so it is locked
// whole — TL, Potential, Tons, and MCr for all eleven stages:
//
//   - The Modified cost cell is /2, not the x1 that pp.63 and 76 print. Modified
//     is 8 tons for MCr4; only 8 x 1/2 gives 4. The book contradicts itself here
//     — p.48's sample notes work the same stage at x1, in prose — so this row is
//     asserted deliberately, to hold the resolution stageData documents (#300).
//   - Tonnage rounds UP: Modified is printed "8 (=7.5)" and Ultimate "4 (=3.7)"
//     for 15/4 = 3.75, per the footer's "Round against advantage".
//   - Three of those rows (8, 5, and 4 tons) sit below the 10-ton Jump-A, so
//     p.77's "no drive may be smaller than the Drive-A of the class" is a floor
//     on the size letter, not on tonnage (see designDrive).
//
// The TL column is the base TL-9 shifted by the stage's TL delta; the MCr column
// is the book's own display rounding of the exact cost, which is why Basic and
// Generic print 8 for an exact MCr7.5.
// The Modified stage's COST is asserted in BOTH catalogs, which is the point of
// #300 having been settled: the stageData note cites p.127 and p.134 as two
// independent columns computing /2, so both must actually hold the cell. If only
// one did, reverting stageData to x1 would fail a single assertion while the note
// went on claiming two corroborated it.
//
// Book 2 prints that cell six times: x1 on pp.63 and 76, /2 on pp.104, 127, 134
// and 190. The worked columns on pp.127 and 134 compute with /2 — but Book 2
// p.48 states the opposite in prose, for two drives, and its figures reproduce
// exactly here:
//
//	"Standard Fusion Plant-S2 104 tons, MCr104 ... The Modified version is TL+2,
//	 half-tonnage, SAME PRICING PER TON, 90% fuel use: Modified Power Plant-S2,
//	 52 tons, MCr52."   (note 14; note 13 is Modified Jump Drive-Z, 125t, MCr125)
//
// So there are worked examples on both sides and no reading satisfies all of
// them. The code follows /2 on weight of evidence — four printings and two
// self-reconciling columns against two printings and two notes — as stageData's
// note argues at length. The p.48 side is recorded there rather than dismissed,
// because omitting it is how this cell was mis-resolved twice. The tonnage half of
// both pages still reproduces exactly, which is what corroborates the round-up
// ruling independently of this.
//
// Tracked in #300. Until it is settled, asserting either value here would dress
// a coin-flip as a golden.
func TestDesignDriveStageCatalogP127(t *testing.T) {
	cases := []struct {
		stage    Stage
		tl       int
		pot      int
		tons     int
		mcr      int
		parenSay string // the book's own unrounded tonnage, where it prints one
	}{
		{Experimental, 6, 1, 45, 450, ""},
		{Prototype, 7, 1, 30, 150, ""},
		{Early, 8, 1, 15, 30, ""},
		{Standard, 9, 2, 15, 15, ""},
		{Basic, 9, 1, 15, 8, ""},
		{Alternate, 9, 2, 15, 15, ""},
		{Improved, 10, 2, 15, 15, ""},
		{Generic, 10, 1, 15, 8, ""},
		{Modified, 11, 2, 8, 4, "=7.5"}, // 8t x 1/2 = MCr4, the /2 reading (#300; see stageData's note)
		{Advanced, 12, 2, 5, 10, ""},
		{Ultimate, 13, 2, 4, 12, "=3.7"},
	}

	const baseTL = 9

	for _, c := range cases {
		// TL-18 keeps availability out of the way; the book's TL column is the
		// stage delta, checked separately against the base TL-9.
		j, problem := designDrive(Jump, DriveSpec{Letter: 2, Stage: c.stage}, 2, 18)
		if problem != "" {
			t.Errorf("%s J-Drive-B in Hull-B: unexpected problem %q", c.stage, problem)
		}

		if got := baseTL + stageData[c.stage].tlDelta; got != c.tl {
			t.Errorf("%s J-Drive-B: TL %d, want %d", c.stage, got, c.tl)
		}

		if j.Potential != c.pot {
			t.Errorf("%s J-Drive-B: Potential %d, want %d", c.stage, j.Potential, c.pot)
		}

		if j.Tons != c.tons {
			t.Errorf("%s J-Drive-B: %d tons, want %d (book prints %q)",
				c.stage, j.Tons, c.tons, c.parenSay)
		}

		// The book's MCr column is the exact cost rounded for display, up like
		// every other figure in the table: Basic's exact MCr7.5 prints as 8.
		//
		// Asserted unconditionally. There used to be an "mcr == 0 skips this row"
		// escape hatch for the disputed Modified cell; #300 settled that, and the
		// hatch was worse than the dispute — it was a documented way to opt a row
		// out of the check, so re-opening #300 by zeroing the row would leave this
		// test green while asserting nothing about the cell it exists to lock.
		if got := ceilDiv(j.Cost, 1_000_000); got != c.mcr {
			t.Errorf("%s J-Drive-B: MCr%d (Cr%d), want MCr%d", c.stage, got, j.Cost, c.mcr)
		}
	}
}

// The p.134 X-table column, the same eleven stages for a P-Plant-B (base 7 tons,
// MCr1 per ton) in the same 200-ton Hull-B. It reproduces the tonnage column
// whole — including Modified 4 "(=3.5)", Advanced 3 "(=2.3)", and Ultimate 2
// "(=1.7)", two of which are below the 4-ton Power-A — and the cost column for
// every row whose tonnage came out a whole number.
//
// The other three costs are NOT lockable, and the reason is worth recording: the
// page computes them from the unrounded tonnage and then from its own truncated
// display of it. Modified prints MCr1.7 (3.5 x 1/2 = 1.75), Advanced MCr4.6
// (2.3 x 2, from a real 2.333), Ultimate MCr5.1 (1.7 x 3, from a real 1.75 whose
// honest product is 5.25). p.127 does the opposite and follows the shared
// footnote, "using final tonnage and cost multiplier" — Modified there is 8 x
// 1/2 = MCr4, not 7.5 x 1/2. The footnote and p.127 agree, so the code follows
// them and p.134's three trailing costs are left as the outlier they are.
func TestDesignDriveStageCatalogP134(t *testing.T) {
	cases := []struct {
		stage     Stage
		tl        int
		pot       int
		tons      int
		mcr       int // 0 where the page's own figure is not reproducible
		fractCost string
	}{
		{Experimental, 6, 1, 21, 210, ""},
		{Prototype, 7, 1, 14, 70, ""},
		{Early, 8, 1, 7, 14, ""},
		{Standard, 9, 2, 7, 7, ""},
		{Basic, 9, 1, 7, 4, "3.5"}, // exact Cr3,500,000
		{Alternate, 9, 2, 7, 7, ""},
		{Improved, 10, 2, 7, 7, ""},
		{Generic, 10, 1, 7, 4, "3.5"},
		{Modified, 11, 2, 4, 0, "1.7"}, // 3.5 unrounded x 1/2; see the exact-cost loop below
		{Advanced, 12, 2, 3, 0, "4.6"},
		{Ultimate, 13, 2, 2, 0, "5.1"},
	}

	const baseTL = 9

	for _, c := range cases {
		p, problem := designDrive(Power, DriveSpec{Letter: 2, Stage: c.stage}, 2, 18)
		if problem != "" {
			t.Errorf("%s P-Plant-B in Hull-B: unexpected problem %q", c.stage, problem)
		}

		if got := baseTL + stageData[c.stage].tlDelta; got != c.tl {
			t.Errorf("%s P-Plant-B: TL %d, want %d", c.stage, got, c.tl)
		}

		if p.Potential != c.pot {
			t.Errorf("%s P-Plant-B: Potential %d, want %d", c.stage, p.Potential, c.pot)
		}

		if p.Tons != c.tons {
			t.Errorf("%s P-Plant-B: %d tons, want %d", c.stage, p.Tons, c.tons)
		}

		if c.mcr != 0 {
			if got := ceilDiv(p.Cost, 1_000_000); got != c.mcr {
				t.Errorf("%s P-Plant-B: MCr%d (Cr%d), want MCr%d", c.stage, got, p.Cost, c.mcr)
			}

			continue
		}
		// mcr == 0 marks a row whose printed cost this page derives from its
		// unrounded tonnage rather than the shared footnote's "final tonnage"
		// (see the note below). Assert the divergence rather than skipping it
		// silently, so the row is pinned to a decision instead of to nothing —
		// and name the page's own figure, which fractCost carries.
		if got := ceilDiv(p.Cost, 1_000_000); got == 0 {
			t.Errorf("%s P-Plant-B: cost is zero; p.134 prints MCr%s", c.stage, c.fractCost)
		}
	}

	// The three that p.134 gets from unrounded tonnage: what the footnote's rule
	// actually yields, pinned so the divergence stays visible and deliberate.
	for _, c := range []struct {
		stage Stage
		cost  int
	}{
		// Modified is included, and must be: the stageData note cites this column
		// as the SECOND independent worked example for /2 (#300), so leaving it
		// unasserted would let a revert to x1 fail only p.127's catalog while the
		// note still claimed two columns backed it.
		{Modified, 2_000_000}, // p.134 prints MCr1.7 (3.5 unrounded x 1/2); 4 final tons x 1/2 = 2
		{Advanced, 6_000_000}, // p.134 prints MCr4.6 (2.3 x 2)
		{Ultimate, 6_000_000}, // p.134 prints MCr5.1 (1.7 x 3)
	} {
		p, _ := designDrive(Power, DriveSpec{Letter: 2, Stage: c.stage}, 2, 18)
		if p.Cost != c.cost {
			t.Errorf("%s P-Plant-B: Cr%d, want Cr%d (final tonnage x cost multiplier)",
				c.stage, p.Cost, c.cost)
		}
	}
}

// The p.104 X-table column (M-Drive-B, base 3 tons at MCr2 per ton, in Hulls
// A-B-C) is NOT lockable as a whole, and this records exactly how far it goes.
//
// Its tonnage column reproduces on ten of eleven rows — Alternate prints 7 tons
// where every other 100%-efficiency, x1-tonnage stage on the page prints 3, an
// isolated typo. Its TL column is off by one from Basic downward (Standard is
// TL-10 as the caption says, but Basic through Ultimate are numbered from 9).
// Its cost column drops the cost multiplier entirely on the three
// tonnage-reduced stages: Advanced prints MCr2 where the page's own footnote
// ("Base cost of MCr2 per ton using final tonnage and cost multiplier") gives
// 1 ton x MCr2 x 2 = MCr4.
//
// What it does corroborate is the rounding: Modified 3/2 = 1.5 prints 2 and
// Ultimate 3/4 = 0.75 prints 1, both rounded up, and that Ultimate 1 is below
// the 2-ton Maneuver-A — a third page's worth of sub-Drive-A tonnage.
func TestDesignDriveStageCatalogP104(t *testing.T) {
	cases := []struct {
		stage Stage
		tons  int
	}{
		{Experimental, 9},
		{Prototype, 6},
		{Early, 3},
		{Standard, 3},
		{Basic, 3},
		// Alternate omitted: the page prints 7 tons, which no rule on it produces.
		{Improved, 3},
		{Generic, 3},
		{Modified, 2}, // 1.5, rounded up
		{Advanced, 1},
		{Ultimate, 1}, // 0.75, rounded up — and below the 2-ton Maneuver-A
	}
	for _, c := range cases {
		m, problem := designDrive(Maneuver, DriveSpec{Letter: 2, Stage: c.stage}, 2, 18)
		if problem != "" {
			t.Errorf("%s M-Drive-B in Hull-B: unexpected problem %q", c.stage, problem)
		}

		if m.Tons != c.tons {
			t.Errorf("%s M-Drive-B: %d tons, want %d", c.stage, m.Tons, c.tons)
		}
	}

	// The page's Alternate row is the same drive as Standard in every column that
	// matters, whatever its tonnage cell says.
	alt, _ := designDrive(Maneuver, DriveSpec{Letter: 2, Stage: Alternate}, 2, 18)
	std, _ := designDrive(Maneuver, DriveSpec{Letter: 2, Stage: Standard}, 2, 18)

	if alt.Tons != std.Tons || alt.Cost != std.Cost {
		t.Errorf("Alternate M-Drive-B = %dt/Cr%d, want Standard's %dt/Cr%d",
			alt.Tons, alt.Cost, std.Tons, std.Cost)
	}
}

// The p.76 Table X footer states the rule in EP terms. As printed: "Standard
// Drive-C has 300 EP; Early Drive-C outputs 150 EP; Advanced Drive-C outputs 360
// EP." Read "Early" as Experimental — 150 is 50% of 300, which is Experimental's
// efficiency; Early is 90% and would print 270 (see drivePotential's note).
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

// The stage TL delta shifts the availability lookup DOWN, not up: a stage-d
// drive at a TL-t yard reaches what a Standard drive reaches at TL t-d. Book 2
// p.76 says so twice in worked prose — "Standard TL-10 Maneuver-F ... Advanced
// (TL-13) Maneuver Drive-F" and "Standard TL-12 Jump-F ... Modified Jump
// Drive-F (available at TL-14)" — an advanced stage RAISES the TL a yard needs.
// p.76 also puts the other end: "Early, Prototype, and Experimental mechanisms
// are available locally", i.e. below the standard TL.
func TestDesignDriveStageShiftsAvailabilityDown(t *testing.T) {
	// Experimental (-3) at a TL-6 yard reads availability at TL-9, where Jump
	// reaches Potential-1. Its 50% efficiency rates it at 1, so it is buildable.
	j, problem := designDrive(Jump, DriveSpec{Letter: 2, Stage: Experimental}, 2, 6)
	if j.Potential != 1 || problem != "" {
		t.Errorf("Experimental J-Drive-B at TL-6 = Potential %d (%q), want 1 and no problem",
			j.Potential, problem)
	}

	// Advanced (+3) at a TL-9 yard reads availability at TL-6, where no jump
	// drive exists at all. Its 120% efficiency rates it at 2; the yard cannot
	// build it.
	j, problem = designDrive(Jump, DriveSpec{Letter: 2, Stage: Advanced}, 2, 9)
	if j.Potential != 0 || problem == "" {
		t.Errorf("Advanced J-Drive-B at TL-9 = Potential %d (%q), want 0 with a problem",
			j.Potential, problem)
	}

	// The problem names the shifted TL that did the capping, not just the yard's.
	if !strings.Contains(problem, "TL-9") || !strings.Contains(problem, "TL-6") {
		t.Errorf("problem %q should name both the yard TL-9 and the shifted TL-6", problem)
	}
}

// A drive whose Potential computes to 0 cannot function (Book 2 p.63, and p.127
// names the case: "Early Jump-1 at 90% becomes Jump-0"). That is a design
// failure and has to be reported, not billed in silence.
func TestDesignDrivePotentialZeroIsReported(t *testing.T) {
	// Too small for the hull: Jump-A in a Hull-C is 2*1/3 = 0.
	j, problem := designDrive(Jump, DriveSpec{Letter: 1}, 3, 15)
	if j.Potential != 0 || problem == "" {
		t.Errorf("Jump-A in Hull-C = Potential %d (%q), want 0 with a problem",
			j.Potential, problem)
	}

	// Rounded away by efficiency: the book's own Early Jump-1 case.
	j, problem = designDrive(Jump, DriveSpec{Letter: 1, Stage: Early}, 2, 15)
	if j.Potential != 0 || problem == "" {
		t.Errorf("Early Jump-A in Hull-B = Potential %d (%q), want 0 with a problem",
			j.Potential, problem)
	}

	// One mistake, one message: a Potential-0 drive is not also complained about
	// for the TL cap (design.go's `aboard` policy).
	if strings.Count(problem, ";") != 0 {
		t.Errorf("expected a single problem, got %q", problem)
	}
}

// DriveForPotential answers the Standard-stage question only — the book's Z2 is
// a plain Potential-by-Hull grid with no efficiency dimension, and p.76 says the
// Drive Potential Tables "show standard tech levels". This pins both halves of
// that contract: it holds at 100%, and it demonstrably does not hold elsewhere,
// which is what the doc used to claim without qualification.
func TestDriveForPotentialIsStandardStageOnly(t *testing.T) {
	for hullOrd := 1; hullOrd <= maxLetter; hullOrd++ {
		for potential := 1; potential <= 9; potential++ {
			ord := DriveForPotential(potential, hullOrd)
			if ord == 0 {
				continue
			}

			if got := drivePotential(ord, hullOrd, 100); got < potential {
				t.Errorf("DriveForPotential(%d, %d) = %d yields only Potential-%d at Standard",
					potential, hullOrd, ord, got)
			}
		}
	}

	// Below 100% the named size falls short, and no caller may assume otherwise.
	// Jump-2 in a Hull-B is Drive-B; at Early's 90% that is 2*2*90/200 = 1.
	if ord := DriveForPotential(2, 2); ord != 2 || drivePotential(ord, 2, 90) != 1 {
		t.Errorf("DriveForPotential(2, 2) = %d, yielding %d at 90%% — expected 2 yielding 1",
			ord, drivePotential(ord, 2, 90))
	}

	// Above 100% it overshoots, which is harmless but is also not "at least".
	if got := drivePotential(DriveForPotential(2, 2), 2, 130); got != 2 {
		t.Errorf("Drive-B in Hull-B at 130%% = Potential %d, want 2", got)
	}
}
