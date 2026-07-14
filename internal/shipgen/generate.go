package shipgen

import (
	"strings"

	"github.com/philoserf/t5/internal/dice"
)

// ConfigByLetter returns the Config for a QSP config letter (C/B/P/U/S/A/L,
// case-insensitive) and whether it was found. Only the first character is
// significant, so a full config name whose initial matches (e.g. "Cluster")
// also resolves.
func ConfigByLetter(letter string) (Config, bool) {
	if letter == "" {
		return 0, false
	}
	b := letter[0]
	if b >= 'a' && b <= 'z' {
		b -= 'a' - 'A'
	}
	for c := Cluster; c <= Lifting; c++ {
		if c.Letter() == b {
			return c, true
		}
	}
	return 0, false
}

// structureByName maps a hull-structure name (case-insensitive) to its value.
var structureByName = map[string]Structure{
	"plate": FramePlate, "frame-plate": FramePlate, "frameplate": FramePlate,
	"shell": Shell, "polymer": Polymer, "feni": FeNi, "organic": Organic, "charged": Charged,
}

// StructureByName returns the Structure for a name and whether it was found.
func StructureByName(name string) (Structure, bool) {
	s, ok := structureByName[strings.ToLower(name)]
	return s, ok
}

// Generate rolls a plausible ship and designs it. Ship design is deterministic,
// so this just chooses a feasible spec — a small hull, a modest tech level, and
// drives sized (via the Z2 inverse) to a low thrust/jump — and delegates to
// Design. It is the random entry point the other generators' Generate provides;
// the package's real work is Design.
func Generate(r *dice.Roller) Ship {
	hullOrd := 1 + r.Index(8) // Hull A-H (100-800t)
	targetG := 1 + r.Index(2) // 1-2 G
	targetJ := 1 + r.Index(2) // Jump 1-2
	spec := ShipSpec{
		Mission:     "X",
		TL:          12 + r.Index(4), // TL 12-15
		HullLetter:  hullOrd,
		Config:      Config(r.Index(int(Lifting) + 1)),
		ArmorLayers: 1 + r.Index(2), // 1-2 layers
		Maneuver:    &DriveSpec{Letter: DriveForPotential(targetG, hullOrd)},
		Jump:        &DriveSpec{Letter: DriveForPotential(targetJ, hullOrd)},
		Power:       &DriveSpec{Letter: DriveForPotential(max(targetG, targetJ), hullOrd)},
		FuelScoop:   true,
	}
	return Design(spec)
}
