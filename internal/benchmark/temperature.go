package benchmark

// Temperature is one Hot and Cold benchmark. Hits are inflicted per one-minute
// round (Book 1 p.37).
type Temperature struct {
	Level      int
	Kelvin     int
	Celsius    int
	Hits       int
	Descriptor string
}

var temperatures = [...]Temperature{
	{-12, 0, -273, 144, "Absolute Zero"},
	{-11, 25, -250, 121, "Hydrogen Ice; LHyd"},
	{-10, 50, -225, 100, "Oxygen Ice"},
	{-9, 75, -200, 81, "Nitrogen Ice"},
	{-8, 100, -175, 64, ""},
	{-7, 125, -150, 49, ""},
	{-6, 150, -125, 36, ""},
	{-5, 175, -100, 25, ""},
	{-4, 200, -75, 16, "Radon Ice"},
	{-3, 225, -50, 9, ""},
	{-2, 250, -25, 4, ""},
	{-1, 275, 0, 1, "Cold"},
	{0, 300, 25, 0, "Human Temperate Environment"},
	{1, 325, 50, 1, "Hot"},
	{2, 350, 75, 4, ""},
	{3, 375, 100, 9, "Water boils"},
	{4, 400, 125, 16, "Sulfur melts"},
	{5, 425, 150, 25, ""},
	{6, 450, 175, 36, ""},
	{7, 475, 200, 49, ""},
	{8, 500, 225, 64, "Tin melts"},
	{9, 525, 250, 81, "Fire"},
	{10, 550, 275, 100, ""},
	{11, 575, 300, 121, ""},
	{12, 600, 325, 144, ""},
	{13, 700, 425, 350, "Lead melts"},
	{14, 800, 525, 400, ""},
	{15, 900, 625, 450, "Aluminum melts"},
	{16, 1_000, 725, 500, ""},
	{17, 2_000, 1_725, 1_000, "Titanium melts"},
	{18, 3_000, 2_725, 1_500, "Spectral M Star surface"},
	{19, 4_000, 3_725, 2_000, "Spectral K Star surface"},
}

// TemperatureAt returns a Hot and Cold benchmark from -12 through +19.
func TemperatureAt(level int) (Temperature, bool) {
	if level < -12 || level > 19 {
		return Temperature{}, false
	}

	return temperatures[level+12], true
}

// HeatHits returns the hits per minute above 600 Kelvin. At or below 600 K,
// callers should use TemperatureAt; negative temperatures return zero.
func HeatHits(kelvin float64) int {
	if kelvin <= 600 {
		return 0
	}

	return int(kelvin / 2)
}

// Insulation is one protection rating and its protected Celsius range.
type Insulation struct {
	Rating int
	MinC   int
	MaxC   int
	ShipAV int
}

var insulations = map[int]Insulation{
	144: {144, -275, 325, 15}, 121: {121, -250, 300, 12}, 100: {100, -225, 275, 10},
	81: {81, -200, 250, 8}, 64: {64, -175, 225, 7}, 49: {49, -150, 200, 5},
	36: {36, -125, 175, 4}, 25: {25, -100, 150, 3}, 16: {16, -75, 125, 2},
	9: {9, -50, 100, 1}, 4: {4, -25, 75, 1}, 1: {1, 0, 50, 1},
}

// InsulationAt returns a printed insulation rating.
func InsulationAt(rating int) (Insulation, bool) {
	value, ok := insulations[rating]

	return value, ok
}
