package shipgen

import "fmt"

// Drive design (Book 2 pp. 54-56, 76-78). Drive Potential — the core performance
// number — turns out to be a clean formula, not the irregular grid it appears
// to be: the p.78 Z1 table is exactly Potential = floor(2 * driveOrdinal /
// hullOrdinal), capped at 9, with 0 meaning "not possible". Verified against the
// book's worked cells (Drive-T in Hull-E = floor(2*18/5) = 7; any Drive = Hull
// gives 2; the floor-to-zero "no" cutoffs all match).

// drivePotential is the Z1 Drive Potential for a drive size in a hull size
// (Book 2 p.78). It means thrust in Gs (Maneuver), jump number (Jump), or the
// EP tier (Power). A result of 0 means the combination cannot function.
func drivePotential(driveOrd, hullOrd int) int {
	if driveOrd < 1 || hullOrd < 1 {
		return 0
	}

	return min(2*driveOrd/hullOrd, 9)
}

// DriveForPotential is the Z2 inverse: the smallest drive size ordinal that
// yields at least the desired Potential in the given hull (Book 2 p.78), or 0 if
// no drive (up to Z2 = 48) can. E.g. Jump-6 in a Hull-K needs ordinal 30 = Q2.
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
// reach at the given tech level (Book 2 p.76 W); 0 if none is available.
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
	Modified:     {"Modified", 2, 1, 1, 110, 90, 1, 2},
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
// Z1 (capped by TL availability, shifted by the stage's TL delta), and its
// stage-adjusted tonnage and cost. It returns the drive and a problem string
// when TL availability caps the Potential below the drive's Z1 rating. The
// stage's fuel/efficiency values are carried on the Drive for the fuel phase.
func designDrive(kind DriveKind, spec DriveSpec, hullOrd, tl int) (*Drive, string) {
	st := stageData[stageIndex(spec.Stage)]
	raw := drivePotential(spec.Letter, hullOrd)

	pot, problem := raw, ""
	if capMax := availabilityMax(kind, tl+st.tlDelta); capMax < raw {
		pot = capMax
		problem = fmt.Sprintf("%s-%s rated %d but TL-%d caps it at %d",
			kind, driveLabel(spec.Letter), raw, tl, capMax)
	}

	tons := driveTonsBase(kind, spec.Letter) * st.tonsNum / st.tonsDen
	if floor := driveTonsBase(kind, 1); tons < floor {
		tons = floor // no drive is smaller than the class Drive-A (Book 2 p.77)
	}

	cost := tons * driveCrPerTon(kind) * st.costNum / st.costDen

	return &Drive{
		Kind: kind, Letter: spec.Letter,
		Potential: pot, Stage: spec.Stage, Tons: tons, Cost: cost,
	}, problem
}
