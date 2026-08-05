package benchmark

// SizeCode is a benchmark size digit. R (Reading) and T (Talking) are the two
// sub-unit codes between 0 and 1 (Book 1 pp.42-43).
type SizeCode string

// Benchmark size codes.
const (
	SizeZero SizeCode = "0"
	SizeR    SizeCode = "R"
	SizeT    SizeCode = "T"
	Size1    SizeCode = "1"
	Size2    SizeCode = "2"
	Size3    SizeCode = "3"
	Size4    SizeCode = "4"
	Size5    SizeCode = "5"
	Size6    SizeCode = "6"
	Size7    SizeCode = "7"
	Size8    SizeCode = "8"
	Size9    SizeCode = "9"
)

// SizeEntry identifies a broad benchmark size.
type SizeEntry struct {
	Code       SizeCode
	Descriptor string
}

var sizes = map[SizeCode]SizeEntry{
	SizeZero: {SizeZero, ""}, SizeR: {SizeR, "Reading"}, SizeT: {SizeT, "Talking"},
	Size1: {Size1, "Coin"}, Size2: {Size2, "Eye"}, Size3: {Size3, "Head"},
	Size4: {Size4, "Rifle"}, Size5: {Size5, "Person"}, Size6: {Size6, "Vehicle"},
	Size7: {Size7, "Ship"}, Size8: {Size8, "Big Ship"}, Size9: {Size9, "Moonlet"},
}

// Size returns a broad benchmark size.
func Size(code SizeCode) (SizeEntry, bool) {
	value, ok := sizes[code]

	return value, ok
}

// DecimalSize returns a size's length in millimeters for decimal digit 0-9.
// The table defines columns 0, R, T, and 1-7; Size 0 has decimal entries only
// for digits 1-9.
func DecimalSize(code SizeCode, decimal int) (float64, bool) {
	if decimal < 0 || decimal > 9 {
		return 0, false
	}

	row, ok := decimalSizes[code]
	if !ok || (code == SizeZero && decimal == 0) {
		return 0, false
	}

	return row[decimal], true
}

var decimalSizes = map[SizeCode][10]float64{
	SizeZero: {0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9},
	SizeR:    {1, 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 1.8, 1.9},
	SizeT:    {2, 2.5, 3, 3.5, 4, 4.5, 5, 5.5, 6, 6.5},
	Size1:    {7, 15, 20, 30, 35, 40, 50, 55, 60, 70},
	Size2:    {75, 90, 100, 110, 120, 140, 150, 160, 180, 190},
	Size3:    {200, 250, 300, 350, 400, 450, 500, 600, 650, 700},
	Size4:    {750, 800, 900, 1_000, 1_050, 1_100, 1_200, 1_300, 1_350, 1_400},
	Size5:    {1_500, 1_600, 1_700, 1_800, 1_900, 2_000, 5_000, 5_500, 6_000, 6_500},
	Size6:    {7_500, 15_000, 20_000, 30_000, 35_000, 40_000, 50_000, 55_000, 60_000, 70_000},
	Size7:    {75_000, 150_000, 200_000, 300_000, 350_000, 400_000, 500_000, 550_000, 600_000, 700_000},
}

// RandomSize returns a varied size in millimeters for a Flux result from -5
// through +5. The table defines columns 0, R, T, and 1-7; Size 0 has no
// non-positive Flux entries.
func RandomSize(code SizeCode, flux int) (float64, bool) {
	if flux < -5 || flux > 5 {
		return 0, false
	}

	row, ok := randomSizes[code]
	if !ok || (code == SizeZero && flux <= 0) {
		return 0, false
	}

	return row[flux+5], true
}

var randomSizes = map[SizeCode][11]float64{
	SizeZero: {0, 0, 0, 0, 0, 0, 0.2, 0.4, 0.6, 0.8, 1},
	SizeR:    {0.5, 0.6, 0.7, 0.8, 0.9, 1, 1.2, 1.4, 1.6, 1.8, 2},
	SizeT:    {1.5, 1.6, 1.7, 1.8, 1.9, 2, 3, 4, 5, 6, 7},
	Size1:    {4.5, 5, 5.5, 6, 6.5, 7, 20, 35, 50, 60, 75},
	Size2:    {40, 50, 55, 60, 70, 75, 100, 120, 150, 180, 200},
	Size3:    {140, 150, 160, 180, 190, 200, 300, 400, 500, 600, 750},
	Size4:    {450, 500, 550, 600, 650, 750, 900, 1_050, 1_200, 1_350, 1_500},
	Size5:    {1_100, 1_200, 1_300, 1_350, 1_400, 1_500, 1_700, 1_900, 5_000, 6_000, 7_500},
	Size6:    {2_000, 5_000, 5_500, 6_000, 6_500, 7_500, 20_000, 35_000, 50_000, 60_000, 75_000},
	Size7:    {40_000, 50_000, 55_000, 60_000, 70_000, 75_000, 200_000, 350_000, 500_000, 600_000, 750_000},
}
