package systemgen

// Placement geometry for the system's orbit map (Book 3 p.29 P2 chart and Book
// 1 p.31 sub-orbit floors). These are the pure tables the rotate scheduler
// consumes; the scheduler itself (round-robin placement across stars with
// collision resolution) builds on them.

// spectralRank orders the spectral types hot-to-cool; combined with the decimal
// it gives a monotonic temperature index (O0=0 .. M9=69). Cooler giants have
// larger radii, so they sit at higher sub-orbit surface orbits.
var spectralRank = map[string]int{"O": 0, "B": 1, "A": 2, "F": 3, "G": 4, "K": 5, "M": 6}

// spectralIndex maps a star to its temperature index. Stars without a decimal
// (size D, brown dwarfs) never reach the sub-orbit table, so their -1 decimal
// is irrelevant here.
func spectralIndex(s Star) int {
	return spectralRank[s.Type]*10 + s.Decimal
}

// suborbitBand pairs the largest spectral index in a band with the orbit at
// which such a star's surface sits.
type suborbitBand struct{ maxIndex, surface int }

// suborbits is the Sub-Orbit table (Book 1 p.31): for each luminosity size, the
// orbit at which a star of a given spectral type has its surface. Only giants
// (Ia/Ib/II/III) engulf inner orbits; sizes IV/V/VI/D and brown dwarfs preclude
// none. Bands are ordered by ascending index; a star sits in the first band
// whose maxIndex covers its spectral index.
var suborbits = map[string][]suborbitBand{
	"Ia":  {{35, 4}, {40, 5}, {50, 6}, {55, 7}, {60, 8}, {69, 9}},
	"Ib":  {{20, 1}, {40, 2}, {45, 4}, {50, 5}, {55, 6}, {60, 7}, {69, 8}},
	"II":  {{35, 0}, {45, 1}, {50, 2}, {55, 4}, {60, 5}, {65, 6}, {69, 7}},
	"III": {{50, 0}, {55, 1}, {60, 2}, {65, 5}, {69, 6}},
}

// surfaceOrbit returns the orbit at which the star's surface sits (Book 1 p.31):
// orbits 0..surfaceOrbit are precluded (inside the star). A star that precludes
// nothing — size IV/V/VI/D or a brown dwarf — returns -1.
func surfaceOrbit(s Star) int {
	bands, ok := suborbits[s.Size]
	if !ok {
		return -1
	}
	idx := spectralIndex(s)
	for _, b := range bands {
		if idx <= b.maxIndex {
			return b.surface
		}
	}
	return bands[len(bands)-1].surface
}

// firstOrbit is the innermost orbit a world may occupy around the star: one
// beyond its precluded surface, or orbit 0 when it precludes none (Book 1 p.31).
func firstOrbit(s Star) int {
	return surfaceOrbit(s) + 1
}

// p2Row is one row of the Basic Placement Chart (Book 3 p.29): the orbit offset
// for a gas giant of each class and for a planetoid belt — all relative to the
// habitable-zone orbit — plus the absolute orbit for an "other" world (World1
// for every world but the last, World2 for the last).
type p2Row struct {
	lgg, sgg, ig, belt int
	world1, world2     int
}

// p2Chart is indexed by roll 1..12 (index 0 unused). Gas-giant and belt offsets
// add to the habitable-zone orbit; world orbits are absolute (Book 3 p.29).
var p2Chart = [13]p2Row{
	1:  {-4, -3, 0, -2, 11, 18},
	2:  {-3, -2, 1, -1, 10, 17},
	3:  {-2, -1, 2, 0, 8, 16},
	4:  {-1, 0, 3, 1, 6, 15},
	5:  {0, 1, 4, 2, 4, 14},
	6:  {1, 2, 5, 3, 2, 13},
	7:  {2, 3, 6, 4, 0, 12},
	8:  {3, 4, 7, 5, 1, 11},
	9:  {4, 5, 8, 6, 3, 10},
	10: {5, 6, 9, 7, 5, 9},
	11: {6, 7, 10, 8, 7, 8},
	12: {7, 8, 11, 9, 9, 7},
}

// clamp constrains v to the inclusive range [lo, hi].
func clamp(v, lo, hi int) int {
	return min(max(v, lo), hi)
}

// p2 returns the placement row for a 2D roll, clamped to the chart's 1..12 range.
func p2(roll int) p2Row {
	return p2Chart[clamp(roll, 1, 12)]
}

// ggOffset is the HZ-relative orbit offset for a gas giant of the given class
// on this placement row (Book 3 p.29).
func (row p2Row) ggOffset(class GGClass) int {
	switch class {
	case LargeGasGiant:
		return row.lgg
	case IceGiant:
		return row.ig
	default:
		return row.sgg
	}
}
