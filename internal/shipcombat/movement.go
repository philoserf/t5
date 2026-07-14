package shipcombat

// Movement and ranging (Book 2 pp. 195, 200-201).

// Agility is a ship's combat agility (Book 2 p.200): its maximum (or Power Plant
// Potential) Gs minus the Gs currently used for other maneuvering. Greater
// Agility resolves attacks first but moves last.
func Agility(maxGs, usedGs int) int {
	return maxGs - usedGs
}

// RammingHits is the damage a rammed ship takes (Book 2 p.200): the rammer's
// number of Compartments squared, applied at a Flux-rolled location. Ramming is
// only possible at Boarding Range, and the target may avoid it with a Check
// Double Pilot (2D).
func RammingHits(rammerCompartments int) int {
	return rammerCompartments * rammerCompartments
}

// rangeChangeRounds is the number of Combat Rounds to change between a Space
// Range band and the adjacent inner band at 1G..7+G (Book 2 p.200), for the
// tactical bands 1-8 where transit is measured in rounds. Columns are 1G..6G then
// a shared 7G/8G/9G column. Bands 9-12 transit in hours or days.
var rangeChangeRounds = map[int][7]int{
	8: {18, 12, 10, 9, 8, 7, 6},
	7: {15, 11, 9, 8, 7, 6, 5},
	6: {5, 4, 3, 2, 2, 2, 2},
	5: {2, 1, 1, 1, 1, 1, 1},
	4: {1, 1, 1, 1, 1, 1, 1},
	3: {1, 1, 1, 1, 1, 1, 1},
	2: {1, 1, 1, 1, 1, 1, 1},
	1: {1, 1, 1, 1, 1, 1, 1},
}

// RangeChangeRounds returns the Combat Rounds to change between the given Space
// Range band and the adjacent inner band at a constant thrust (Book 2 p.200), and
// whether that transit is within tactical (round-measured) range. Bands 9-12
// transit in hours or days (tactical false); band 0 is the target.
func RangeChangeRounds(band, gs int) (rounds int, tactical bool) {
	row, ok := rangeChangeRounds[band]
	if !ok {
		return 0, false
	}
	return row[gsColumn(gs)], true
}

// gsColumn maps a thrust in Gs to its column: 1G..6G are their own columns and
// 7G and above share the last column (Book 2 p.200).
func gsColumn(gs int) int {
	return clamp(gs, 1, 7) - 1
}
