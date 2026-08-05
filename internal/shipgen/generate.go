package shipgen

import "github.com/philoserf/t5/internal/dice"

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

// StructureByName returns the Structure for a name and whether it was found,
// matching case-, space-, and hyphen-insensitively against the structure's CLI
// short name (see StructureNames), its display name, or its alias — so
// "plate", "Frame-and-Plate", and "frame-plate" all resolve to FramePlate.
func StructureByName(name string) (Structure, bool) {
	key := squash(name)

	for i, d := range structureData {
		if squash(d.cliName) == key || squash(d.display) == key ||
			(d.alias != "" && squash(d.alias) == key) {
			return Structure(i), true
		}
	}

	return 0, false
}

// Generate rolls a plausible ship and designs it. Ship design is deterministic,
// so this just chooses a feasible spec — a small hull, a modest tech level, and
// drives sized (via the Z2 inverse) to a low thrust/jump — and delegates to
// Design. It is the random entry point the other generators' Generate provides;
// the package's real work is Design.
//
// "Feasible" is a promise, so the rolls are constrained rather than left to
// contradict each other and be reported: the thrust target is clamped to the
// rolled configuration's MaxG, because a Cluster hull is rated for 1G and a
// Braced for 3 however big the drive is (Book 2 p.71). That is not the clamping
// Design refuses to do. Design must never quietly build something other than
// what a naval architect asked for — hence specProblems — but Generate's
// "request" is its own dice, and there is no one to surprise. It already picks
// its hull, TL, and drive sizes to stay inside the rules; the config/thrust pair
// was the one place two independent rolls could disagree.
func Generate(r *dice.Roller) Ship {
	hullOrd := 1 + r.Index(8) // Hull A-H (100-800t)
	targetG := 1 + r.Index(2) // 1-2 G
	targetJ := 1 + r.Index(2) // Jump 1-2
	tl := 12 + r.Index(4)     // TL 12-15
	config := Config(r.Index(int(Lifting) + 1))
	maxG := configAttr[config].maxG
	targetG = min(targetG, maxG)
	// Clamping the target is not by itself enough, because thrust has a floor: a
	// drive's Potential is (EP/Hull)*2 and the smallest drive is Drive-A, so
	// Book 2 p.63's own worked case — "Jump Drive-A has 100 Energy Points.
	// Installed in a 100-ton hull, the drive produces Potential-2" — means a
	// 100-ton hull has no maneuver drive rated under 2G. A Cluster is rated for
	// 1G, so a 100-ton Cluster cannot carry a maneuver drive at all. Grow the
	// hull until the thrust the configuration allows is actually orderable.
	for hullOrd < maxLetter && drivePotential(DriveForPotential(targetG, hullOrd), hullOrd, 100) > maxG {
		hullOrd++
	}

	spec := ShipSpec{
		Mission:     "X",
		TL:          tl,
		HullLetter:  hullOrd,
		Config:      config,
		ArmorLayers: 1 + r.Index(2), // 1-2 layers
		Maneuver:    &DriveSpec{Letter: DriveForPotential(targetG, hullOrd)},
		Jump:        &DriveSpec{Letter: DriveForPotential(targetJ, hullOrd)},
		Power:       &DriveSpec{Letter: DriveForPotential(max(targetG, targetJ), hullOrd)},
		FuelScoop:   true,
	}

	return Design(spec)
}
