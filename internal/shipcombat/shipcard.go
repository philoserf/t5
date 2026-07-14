package shipcombat

// The ShipCard compartment model (Book 2 p.86 Table H Hull Location and
// Compartment Tonnage). A hull's hit locations are its Compartments, numbered
// within a Span (-Span..+Span) and each holding six SubCompartments; Flux picks
// the hit Compartment and 1D the SubCompartment.

// A HullLocation describes a hull's compartment structure (Book 2 p.86 Table H):
// the number of Compartments (hit locations), the Span they are numbered within,
// and the tonnage per Compartment and per SubCompartment.
type HullLocation struct {
	Compartments    int
	Span            int
	CompartmentTons int
	SubTons         int
}

// hullLocation is Table H indexed by hull size ordinal (A=1 … X=22); hulls larger
// than X (Y, Z) are not tabulated.
var hullLocation = map[int]HullLocation{
	1:  {5, 4, 20, 3},     // A 100t
	2:  {7, 4, 30, 5},     // B 200t
	3:  {7, 4, 40, 7},     // C 300t
	4:  {9, 4, 45, 8},     // D 400t
	5:  {9, 4, 55, 9},     // E 500t
	6:  {9, 4, 65, 11},    // F 600t
	7:  {9, 4, 70, 12},    // G 700t
	8:  {11, 5, 70, 12},   // H 800t
	9:  {11, 5, 80, 13},   // J 900t
	10: {11, 5, 100, 17},  // K 1000t
	11: {11, 5, 100, 17},  // L 1100t
	12: {13, 6, 100, 17},  // M 1200t
	13: {13, 6, 100, 17},  // N 1300t
	14: {15, 7, 100, 17},  // P 1400t
	15: {15, 7, 100, 17},  // Q 1500t
	16: {17, 8, 100, 17},  // R 1600t
	17: {17, 8, 100, 17},  // S 1700t
	18: {19, 9, 100, 17},  // T 1800t
	19: {19, 9, 100, 17},  // U 1900t
	20: {21, 10, 100, 17}, // V 2000t
	21: {21, 10, 100, 17}, // W 2100t
	22: {23, 11, 100, 17}, // X 2200t
}

// SubCompartmentsPerCompartment is the number of SubCompartments in each
// Compartment (a 1D refinement of a hit location, Book 2 p.86).
const SubCompartmentsPerCompartment = 6

// HullLocations returns a hull's compartment structure by hull size ordinal
// (A=1 … X=22) and whether the hull is tabulated (Book 2 p.86 Table H).
func HullLocations(hullOrdinal int) (HullLocation, bool) {
	h, ok := hullLocation[hullOrdinal]
	return h, ok
}
