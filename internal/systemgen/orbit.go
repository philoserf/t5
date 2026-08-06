package systemgen

// Orbital geometry primitives (Book 1 p.31 orbital distances, Book 3 p.29
// habitable-zone table). These are the pure lookups the per-world placement and
// climate steps build on. The stellar-surface / precluded-orbit table (Book 1 p.31
// "Sub-Orbit" columns), consumed only by the placement scheduler, is transcribed
// with that scheduler.

// hzOrbits maps a star's spectral type to its habitable-zone orbit for each
// luminosity size (Book 3 p.29), in size order Ia Ib II III IV V VI D. A -1
// entry is a type/size the table marks not applicable ("-").
var hzOrbits = map[string][8]int{
	//        Ia  Ib  II III  IV   V  VI   D
	"O": {15, 15, 14, 13, 12, 11, -1, 1},
	"B": {13, 13, 12, 11, 10, 9, -1, 0},
	"A": {12, 11, 9, 7, 7, 7, -1, 0},
	"F": {11, 10, 9, 6, 6, 4, 3, 0},
	"G": {12, 10, 9, 7, 5, 3, 2, 0},
	"K": {12, 10, 9, 8, 5, 2, 1, 0},
	"M": {12, 11, 10, 9, -1, 0, 0, 0},
}

// hzSizeIndex orders the luminosity sizes as the hzOrbits columns.
var hzSizeIndex = map[string]int{
	"Ia":  0,
	"Ib":  1,
	"II":  2,
	"III": 3,
	"IV":  4,
	"V":   5,
	"VI":  6,
	"D":   7,
}

// HZOrbit returns a star's habitable-zone orbit (Book 3 p.29) and whether it has
// one — false for a brown dwarf and for the table's "-" cells (e.g. size VI on an
// O/B/A star). Regina's F-V primary has HZ orbit 4.
func HZOrbit(star Star) (int, bool) {
	row, ok := hzOrbits[star.Type]
	if !ok {
		return 0, false // BD or an unknown type: no habitable zone
	}

	col, ok := hzSizeIndex[star.Size]
	if !ok {
		return 0, false
	}

	if orbit := row[col]; orbit >= 0 {
		return orbit, true
	}

	return 0, false
}

// orbitAU is the semi-major axis in AU for orbits 0-20 (Book 1 p.31, Chart 05).
var orbitAU = [21]float64{
	0.2, 0.4, 0.7, 1.0, 1.6, 2.8, 5.2, 10, 20, 40, 77,
	154, 308, 615, 1230, 2500, 4900, 9800, 19500, 39500, 78700,
}

// OrbitAU returns an orbit's distance from its star in AU (Book 1 p.31); orbits
// outside 0-20 return 0.
func OrbitAU(orbit int) float64 {
	if orbit < 0 || orbit >= len(orbitAU) {
		return 0
	}

	return orbitAU[orbit]
}
