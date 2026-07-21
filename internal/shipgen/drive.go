package shipgen

import "fmt"

// Drive design (Book 2 pp. 54-56, 76-78). Drive Potential — the core performance
// number — turns out to be a clean formula, not the irregular grid it appears
// to be: the p.78 Z1 table is exactly Potential = floor(2 * driveOrdinal /
// hullOrdinal), capped at 9, with 0 meaning "not possible". Verified against the
// book's worked cells (Drive-T in Hull-E = floor(2*18/5) = 7; any Drive = Hull
// gives 2; the floor-to-zero "no" cutoffs all match).

// drivePotential is the Z1 Drive Potential for a drive size in a hull size, at a
// stage efficiency in percent (Book 2 pp.63, 76, 78). It means thrust in Gs
// (Maneuver), jump number (Jump), or the EP tier (Power). A result of 0 means
// the combination cannot function.
//
// Efficiency is not a separate step applied to the Z1 number: it "modifies the
// EP Energy Point output of a Drive", and Potential is P = (EP/Hull)*2,
// "Maximum 9. Round Down" (p.63). So the scaling happens inside the division and
// there is exactly one floor, at the end — p.127 puts it as "Efficiencies round
// down (thus Early Jump-1 at 90% becomes Jump-0)", which flooring the Z1 value
// first could not produce. Drive EP is the size ordinal x 100 (Drive-A = 100 EP,
// Drive-K = 1000) and hull tons is the hull ordinal x 100, so both hundreds
// cancel and only the percentage is left to divide out.
//
// Book conflict, p.76 Table X footer. As printed:
//
//	"Efficiency modifies the EP Energy Point output of a Drive. Standard
//	 Drive-C has 300 EP; Early Drive-C outputs 150 EP; Advanced Drive-C
//	 outputs 360 EP."
//
// Read "Early" as Experimental. 150 is 50% of 300, and 50% is Experimental's
// efficiency in the same table's own column; Early is 90%, which would print
// 270. Advanced checks out as printed (360 = 120% of 300). The footer is
// illustration, not a rule input — the efficiency column beside it is what this
// function is given — so nothing here changes either way; the note exists so
// the next reader comparing code to p.76 finds the discrepancy explained.
func drivePotential(driveOrd, hullOrd, effPct int) int {
	if driveOrd < 1 || hullOrd < 1 {
		return 0
	}

	return min(2*driveOrd*effPct/(hullOrd*100), 9)
}

// DriveForPotential is the Z2 inverse: the smallest drive size ordinal that
// yields at least the desired Potential in the given hull at STANDARD stage
// (Book 2 p.78), or 0 if no drive (up to Z2 = 48) can. E.g. Jump-6 in a Hull-K
// needs ordinal 30 = Q2.
//
// Standard stage is the whole contract, and deliberately so: the book's Z2 is a
// plain Potential-by-Hull grid with no efficiency dimension, the exact transpose
// of Z1, and p.76 says "The Drive Potential Tables show standard tech levels."
// Stage efficiency is a Table X effect layered on top, not part of this table.
//
// So this is the inverse of drivePotential at effPct=100 only. Below 100%
// (Experimental 50%, Prototype 80%, Early/Basic/Generic 90%) the size it names
// falls short of the Potential asked for; above it (Improved/Modified 110%,
// Advanced 120%, Ultimate 130%) it overshoots. The shortfall cannot be undone by
// scaling the argument — the efficiency divides inside drivePotential's single
// floor — so a stage-aware caller must check the answer with drivePotential at
// its own efficiency and step the size up until it satisfies. No caller needs
// that today: Generate builds Standard drives and passes effPct=100 explicitly.
//
// The result is always a size a yard can build. Past Drive-Z the only sizes this
// package models are the "letter2" gangs — two identical drives joined by a
// Nexus (Book 2 p.63) — so the extended ordinals are exactly the even 26..48,
// and an odd product is rounded up to the next one. (The book's Nexi are more
// general: any m x letter x n up to 9Z9, so ordinal 25 is really buildable as
// E5. Nothing else here models those, and driveLabel/driveTonsBase's >24 branch
// are both written for the doubling, so the inverse stays inside that set rather
// than naming a size the rest of the package would misprice.)
func DriveForPotential(potential, hullOrd int) int {
	if potential < 1 || hullOrd < 1 {
		return 0
	}

	ord := (potential*hullOrd + 1) / 2 // ceil(potential*hull/2)
	if ord > maxLetter && ord%2 == 1 {
		ord++ // extended sizes are letter2 doublings — even only
	}

	if ord > 2*maxLetter { // past Z2
		return 0
	}

	return ord
}

// ceilDiv divides and rounds up, for non-negative operands. The stage tables
// round "against advantage (up for tonnage; down for potential)" (Book 2 pp.104,
// 127, 134), and Go's integer division rounds down.
func ceilDiv(n, d int) int {
	return (n + d - 1) / d
}

// driveTonsBase is a drive's base tonnage before stage effects (Book 2 p.77 Y
// table). Each drive type is a piecewise-linear function of the size ordinal —
// verified against the p.77 grid (e.g. Jump-A = 10, Jump-Z = 125; Power-A = 4).
func driveTonsBase(kind DriveKind, ord int) int {
	switch kind {
	case Maneuver:
		switch {
		case ord == 1:
			return 2
		case ord <= 24:
			return 2*ord - 1
		default:
			return 2*ord - 2
		}
	case Jump:
		if ord <= 24 {
			return 5*ord + 5
		}

		return 5*ord + 10
	case Power:
		if ord <= 24 {
			return 3*ord + 1
		}

		return 3*ord + 2
	}

	return 0
}

// driveCrPerTon is a drive type's cost per ton in Cr (Book 2 p.77 footer):
// Maneuver MCr2.0, Jump MCr1.0, Power MCr1.0.
func driveCrPerTon(kind DriveKind) int {
	if kind == Maneuver {
		return 2_000_000
	}

	return 1_000_000
}

// driveAvail is each drive type's Tech Level availability breakpoints (Book 2
// p.76 W): the maximum Drive Potential a standard drive reaches at each TL.
// availabilityMax reads the highest breakpoint at or below a TL.
var driveAvail = map[DriveKind][]struct{ tl, max int }{
	Maneuver: {{9, 1}, {10, 3}, {11, 5}, {12, 7}, {13, 9}},
	Jump:     {{9, 1}, {11, 2}, {12, 3}, {13, 4}, {14, 5}, {15, 6}, {16, 7}, {17, 8}, {18, 9}},
	Power:    {{8, 1}, {9, 2}, {10, 3}, {11, 4}, {12, 5}, {13, 6}, {14, 7}, {15, 8}, {16, 9}},
}

// availabilityMax is the highest Potential a standard drive of the kind can
// reach at the given tech level (Book 2 p.76 W); 0 means no drive of that kind
// is available that low. driveAvail covers every DriveKind, so 0 is only ever
// the "TL too low" answer — the scan's unguarded low end is not also standing in
// for an unknown kind. Callers may pass a TL below any breakpoint (designDrive
// shifts the yard's TL down by the stage delta), and 0 is the right answer there.
func availabilityMax(kind DriveKind, tl int) int {
	m := 0

	for _, bp := range driveAvail[kind] {
		if tl >= bp.tl {
			m = bp.max
		}
	}

	return m
}

// Stage is a drive's TL stage effect (Book 2 p.76 X): a build quality that
// trades tech-level delta, cost, efficiency, fuel, and tonnage. Standard is the
// baseline (zero value).
type Stage int

// TL stages for a drive (Book 2 p. 76).
const (
	Standard Stage = iota
	Experimental
	Prototype
	Early
	Basic
	Alternate
	Improved
	Generic
	Modified
	Advanced
	Ultimate
)

// stageData is the p.76 X table. Cost and tons are stored as fractions
// (numerator/denominator) to stay exact; fuelPct and eff are percentages.
//
// Book conflict, the Modified cost cell (#300). Book 2 prints the same X table
// six times and does not agree with itself:
//
//   - /2 on pp.104, 127, 134, 190
//   - x1 on pp.63, 76
//
// Two worked columns compute /2 and reconcile every row they print. p.127's
// Jump-B column: base 15t at MCr1/t, so Modified is 8t for MCr4 = 8 x 1/2,
// against Advanced's 5t for MCr10 = 5 x 2 and Ultimate's 4t for MCr12 = 4 x 3.
// p.134's Power-B column independently prints "Mod P-Plant-B ... 4 (=3.5) 1.7",
// and 3.5 x 1/2 = 1.75. (p.104's column is broken on its own terms — it honours
// the multiplier for the eight non-divisor rows and ignores it for Mod/Adv/Ult —
// so it is evidence for neither reading.)
//
// The x1 side is NOT unsupported, and pretending otherwise is how this cell got
// mis-resolved twice. Book 2 p.48's sample-ship notes work it the other way, one
// of them in prose: note 14 has "Standard Fusion Plant-S2 104 tons, MCr104 ...
// The Modified version is TL+2, half-tonnage, same pricing per ton ... 52 tons,
// MCr52", and note 13 does the same for a Jump Drive-Z (250t/MCr250 ->
// 125t/MCr125, MCr1 per ton throughout). Those are worked, explicit, and flatly
// incompatible with /2.
//
// So no reading satisfies the whole book. This code follows /2 on weight of
// evidence: four printings against two, and two independent worked columns that
// each reconcile all eleven of their rows, against two sample notes. A reader who
// prices a Modified drive from p.48 and gets twice this figure has found the
// book's inconsistency, not a bug here. TestDesignDriveStageCatalogP127 asserts
// the cell, so the resolution cannot drift back silently.
var stageData = [...]struct {
	name             string
	tlDelta          int
	costNum, costDen int
	eff, fuelPct     int
	tonsNum, tonsDen int
}{
	Standard:     {"Standard", 0, 1, 1, 100, 100, 1, 1},
	Experimental: {"Experimental", -3, 10, 1, 50, 200, 3, 1},
	Prototype:    {"Prototype", -2, 5, 1, 80, 120, 2, 1},
	Early:        {"Early", -1, 2, 1, 90, 110, 1, 1},
	Basic:        {"Basic", 0, 1, 2, 90, 110, 1, 1},
	Alternate:    {"Alternate", 0, 1, 1, 100, 100, 1, 1},
	Improved:     {"Improved", 1, 1, 1, 110, 90, 1, 1},
	Generic:      {"Generic", 1, 1, 2, 90, 110, 1, 1},
	Modified:     {"Modified", 2, 1, 2, 110, 90, 1, 2},
	Advanced:     {"Advanced", 3, 2, 1, 120, 80, 1, 3},
	Ultimate:     {"Ultimate", 4, 3, 1, 130, 70, 1, 4},
}

func (s Stage) String() string {
	if s < Standard || int(s) >= len(stageData) {
		return "?"
	}

	return stageData[s].name
}

// driveLabel renders a drive size ordinal as its letter, or "letter2" for an
// extended size (an even ordinal 26..48, where the "2" doubles the base letter's
// ordinal — e.g. 26 = N x 2 = "N2").
// buildableDriveOrd reports whether an ordinal names a drive this package can
// build. It is driveLabel's domain stated as a predicate, and the two must agree:
// an ordinal driveLabel renders as "?" is one no table row backs, so a spec
// carrying it should be reported rather than silently designed.
//
// A..Z is 1..24. Past Z the book gangs drives with a Nexus (p.63), and only the
// x2 gang is modelled — so the extended ordinals are exactly the EVEN 26..48. An
// odd ordinal up there (25, 27, ...) is buildable in the book as an m-by-n gang
// but is not modelled here, and previously designed silently as "Drive-?".
func buildableDriveOrd(ord int) bool {
	return driveLabel(ord) != "?"
}

func driveLabel(ord int) string {
	switch {
	case ord >= 1 && ord <= maxLetter:
		return string(ordinalLetter(ord))
	case ord >= 26 && ord%2 == 0 && ord/2 <= maxLetter:
		return string(ordinalLetter(ord/2)) + "2"
	default:
		return "?"
	}
}

// designDrive builds one drive for a hull (Book 2 pp. 76-78): its Potential from
// Z1 — scaled by the stage's efficiency, then capped by TL availability at the
// stage's shifted TL — and its stage-adjusted tonnage and cost. It returns the
// drive and a problem string when the combination cannot function or TL
// availability caps the Potential below the drive's rating. The stage's fuel
// multiplier is applied later, in the fuel phase, from Drive.Stage.
//
// The stage shifts the availability lookup DOWN by its TL delta. An advanced
// stage RAISES the tech level a yard needs, so a yard at TL t builds a stage-d
// drive only as far as a Standard drive reaches at TL t-d. Book 2 p.76 states
// it twice in worked prose — "Standard TL-10 Maneuver-F ... Advanced (TL-13)
// Maneuver Drive-F" (+3: 10 -> 13) and "Standard TL-12 Jump-F ... Modified Jump
// Drive-F (available at TL-14)" (+2: 12 -> 14) — and puts the other end as
// "Early, Prototype, and Experimental mechanisms are available locally", i.e.
// buildable below the standard TL.
//
// Note the sign is opposite to the weapon/defense side, where stageTL ADDS the
// same delta: there it shifts the device's own TL rating upward, here it shifts
// the yard's reach downward. The shared stageData.tlDelta column is correct for
// both; only the direction of use differs, so this is corrected here rather
// than in the table.
func designDrive(kind DriveKind, spec DriveSpec, hullOrd, tl int) (*Drive, Problem) {
	st := stageData[stageIndex(spec.Stage)]
	raw := drivePotential(spec.Letter, hullOrd, st.eff)
	availTL := tl - st.tlDelta

	pot := raw

	var problem Problem

	switch {
	case raw == 0:
		// The combination cannot function at all (Book 2 p.63): the drive is too
		// small for the hull, or the stage's efficiency rounded its Potential
		// away — p.127's own "Early Jump-1 at 90% becomes Jump-0". Reported
		// instead of the TL cap, never alongside it: capMax < 0 is impossible, so
		// one mistake still yields one message (see design.go's `aboard` note).
		problem = Problem{Kind: DrivePotentialZero, Detail: fmt.Sprintf(
			"%s-%s in a Hull-%s yields Potential 0: too small to drive this hull",
			kind, driveLabel(spec.Letter), driveLabel(hullOrd),
		)}
	case availabilityMax(kind, availTL) < raw:
		capMax := availabilityMax(kind, availTL)
		pot = capMax
		detail := fmt.Sprintf("%s-%s rated %d but TL-%d caps it at %d",
			kind, driveLabel(spec.Letter), raw, tl, capMax)

		if st.tlDelta != 0 {
			detail += fmt.Sprintf(" (%s reads availability at TL-%d)", spec.Stage, availTL)
		}

		problem = Problem{Kind: DriveAboveTL, Detail: detail}
	}

	// A Potential-0 drive is still built, weighed, and billed. Design is total
	// and reports infeasibility rather than correcting it (see the package doc):
	// the designer asked for this drive, and a phantom that occupied no tonnage
	// would quietly free up hull space and could turn an over-budget ship into
	// one that appears to fit — trading a reported mistake for a hidden one. The
	// record shows the useless drive; Problems says why it is useless.
	// Stage tonnage rounds UP, and there is no tonnage floor.
	//
	// Book conflict, p.77 versus the rounding footer printed on pp.104, 127, and
	// 134: "Round against advantage (up for tonnage; down for potential)" against
	// "No drive may be smaller than the Drive-A of the class, even after
	// modifications like Stage Effects." Read as a floor on TONNAGE the second
	// rule contradicts the worked tables in seven printed rows — p.127 prints a
	// Modified/Advanced/Ultimate Jump-B at 8/5/4 tons where Jump-A is 10, p.134 a
	// Power-B at 3/2 where Power-A is 4, p.104 a Maneuver-B at 1 where
	// Maneuver-A is 2.
	//
	// The reconciling reading is p.77's own words: smaller than "the Drive-A of
	// the class" is a floor on the size LETTER, not on tonnage. No drive is
	// specified below A — and every worked example above is still a Drive-B,
	// shrunk by its stage but not demoted. Under that reading all thirty-three
	// printed rows across the three tables reproduce and no source is a typo, so
	// tonnage simply rounds up here.
	//
	// The letter floor itself is enforced in specProblems, not here — a sub-A
	// ordinal makes driveTonsBase run negative, and a drive with negative tons
	// and cost would add budget rather than spend it. That check has to sit
	// where the spec is validated; this function is total and must build
	// something for whatever it is handed.
	// A sub-A ordinal is reported by specProblems, but this function is total and
	// still has to build something. Price it as the smallest real drive rather
	// than letting driveTonsBase run negative: negative tonnage would ADD budget
	// and free hull space, turning an over-budget ship into one that appears to
	// fit — the same phantom the Potential-0 case is deliberately billed to
	// avoid. Clamp and report, as stageIndex and configIndex do for their fields.
	letter := max(spec.Letter, 1)

	tons := ceilDiv(driveTonsBase(kind, letter)*st.tonsNum, st.tonsDen)
	cost := ceilDiv(tons*driveCrPerTon(kind)*st.costNum, st.costDen)

	return &Drive{
		Kind: kind, Letter: spec.Letter,
		Potential: pot, Stage: spec.Stage, Tons: tons, Cost: cost,
	}, problem
}

// hasProblem reports whether p carries a reported failure. designDrive returns a
// zero Problem (empty Detail) when the drive is buildable.
func (p Problem) reported() bool { return p.Detail != "" }
