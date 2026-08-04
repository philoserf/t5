package vehiclegen

// Dimensions is one readable row of the p.148 Vehicle Design Box. Squares are
// 1.5 m cubes; the table intentionally has two alternative one-ton and nine-ton
// layouts.
type Dimensions struct {
	Tons            float64
	LengthM, WidthM float64
	HeightM         float64
	LengthSquares   int
	WidthSquares    int
	HeightSquares   int
}

// DesignBoxRows returns a copy of every dimensions row printed on p.148.
func DesignBoxRows() []Dimensions { return append([]Dimensions(nil), designBoxRows...) }

var designBoxRows = []Dimensions{ //nolint:gochecknoglobals // immutable chart
	{.25, 1.5, 1.5, 1.5, 1, 1, 1},
	{.50, 3, 1.5, 1.5, 2, 1, 1},
	{.75, 4.5, 1.5, 1.5, 3, 1, 1},
	{1, 3, 3, 1.5, 2, 2, 1},
	{1, 6, 1.5, 1.5, 4, 1, 1},
	{2, 6, 3, 1.5, 4, 2, 1},
	{3, 9, 3, 1.5, 6, 2, 1},
	{4, 12, 3, 1.5, 8, 2, 1},
	{5, 7.5, 3, 3, 5, 2, 2},
	{6, 9, 3, 3, 6, 2, 2},
	{7, 10.5, 3, 3, 7, 2, 2},
	{8, 12, 3, 3, 8, 2, 2},
	{9, 9, 4.5, 3, 6, 3, 2},
	{9, 13.5, 3, 3, 9, 2, 2},
	{10, 15, 3, 3, 10, 2, 2},
	{11, 10.5, 4.5, 3, 7, 3, 2},
	{12, 12, 4.5, 3, 8, 3, 2},
	{13, 13.5, 4.5, 3, 9, 3, 2},
	{15, 15, 4.5, 3, 10, 3, 2},
	{16, 16.5, 4.5, 3, 11, 3, 2},
}

// Access is a vehicle/terrain chart cell (Book 3 p.143).
type Access uint8

const (
	// Disallowed is possible only with the chart's harder operation task.
	Disallowed Access = iota
	// Accessible is routine allowed terrain.
	Accessible
	// Prohibited cannot be entered.
	Prohibited
	// GridOnly requires a Grid-equipped vehicle.
	GridOnly
	// Conditional depends on details the chart does not settle.
	Conditional
	// Temporary is only temporarily accessible.
	Temporary
)

// SurfaceAccess returns the p.143 surface-terrain cell. Vehicle is one of
// People, Cars, OffRoad, ACV, Wheel, Track, Legged, Lifter, or GDrive.
func SurfaceAccess(terrain, vehicle string) (Access, bool) {
	row, ok := surfaceAccess[terrain]
	if !ok {
		return 0, false
	}

	a, ok := row[vehicle]

	return a, ok
}

func accessRow(values ...Access) map[string]Access {
	vehicles := [...]string{"People", "Cars", "OffRoad", "ACV", "Wheel", "Track", "Legged", "Lifter", "GDrive"}

	row := make(map[string]Access, len(vehicles))
	for i, vehicle := range vehicles {
		row[vehicle] = values[i]
	}

	return row
}

var surfaceAccess = map[string]map[string]Access{ //nolint:gochecknoglobals // chart
	"Air Corridor": accessRow(
		Prohibited,
		Prohibited,
		Prohibited,
		Prohibited,
		Prohibited,
		Prohibited,
		Prohibited,
		GridOnly,
		GridOnly,
	),
	"Grid": accessRow(
		Prohibited,
		GridOnly,
		GridOnly,
		Disallowed,
		Disallowed,
		Disallowed,
		Prohibited,
		GridOnly,
		Disallowed,
	),
	"Highway": accessRow(
		Accessible,
		Accessible,
		Accessible,
		Disallowed,
		Accessible,
		Disallowed,
		Prohibited,
		Accessible,
		Disallowed,
	),
	"Road": accessRow(
		Accessible,
		Accessible,
		Accessible,
		Disallowed,
		Accessible,
		Accessible,
		Accessible,
		Accessible,
		Disallowed,
	),
	"Trail": accessRow(
		Accessible,
		Disallowed,
		Disallowed,
		Disallowed,
		Disallowed,
		Disallowed,
		Accessible,
		Disallowed,
		Disallowed,
	),
	"Clear": accessRow(
		Accessible,
		Disallowed,
		Accessible,
		Accessible,
		Accessible,
		Accessible,
		Accessible,
		Disallowed,
		Disallowed,
	),
	"Clear Wooded": accessRow(
		Accessible,
		Disallowed,
		Disallowed,
		Disallowed,
		Disallowed,
		Accessible,
		Accessible,
		Disallowed,
		Disallowed,
	),
	"Wetland": accessRow(
		Accessible,
		Disallowed,
		Disallowed,
		Accessible,
		Disallowed,
		Disallowed,
		Accessible,
		Disallowed,
		Disallowed,
	),
	"Wetland Wooded": accessRow(
		Accessible,
		Disallowed,
		Disallowed,
		Disallowed,
		Disallowed,
		Disallowed,
		Accessible,
		Disallowed,
		Disallowed,
	),
	"Rough": accessRow(
		Accessible,
		Disallowed,
		Disallowed,
		Disallowed,
		Accessible,
		Accessible,
		Accessible,
		Disallowed,
		Disallowed,
	),
	"Rough Wooded": accessRow(
		Accessible,
		Disallowed,
		Disallowed,
		Disallowed,
		Disallowed,
		Accessible,
		Accessible,
		Disallowed,
		Disallowed,
	),
	"Mountain": accessRow(
		Accessible,
		Disallowed,
		Disallowed,
		Disallowed,
		Disallowed,
		Accessible,
		Accessible,
		Disallowed,
		Disallowed,
	),
	"River": accessRow(
		Disallowed,
		Disallowed,
		Disallowed,
		Accessible,
		Accessible,
		Accessible,
		Disallowed,
		Disallowed,
		Disallowed,
	),
	"Lake": accessRow(
		Prohibited,
		Prohibited,
		Prohibited,
		Accessible,
		Accessible,
		Disallowed,
		Disallowed,
		Disallowed,
		Disallowed,
	),
	"Ocean": accessRow(
		Prohibited,
		Prohibited,
		Prohibited,
		Accessible,
		Prohibited,
		Prohibited,
		Prohibited,
		Disallowed,
		Disallowed,
	),
}

// FlyerAccess returns the p.143 flyer-terrain cell. Flyer is People, Wing,
// Flap, Rotor, LTA, Lifter, GDrive, or MDrive. Blank cells in the source are
// Disallowed; "dependent on other details" is Conditional.
func FlyerAccess(terrain, flyer string) (Access, bool) {
	row, ok := flyerAccess[terrain]
	if !ok {
		return 0, false
	}

	a, ok := row[flyer]

	return a, ok
}

func flyerRow(values ...Access) map[string]Access {
	vehicles := [...]string{"People", "Wing", "Flap", "Rotor", "LTA", "Lifter", "GDrive", "MDrive"}

	row := make(map[string]Access, len(vehicles))
	for i, vehicle := range vehicles {
		row[vehicle] = values[i]
	}

	return row
}

var flyerAccess = map[string]map[string]Access{ //nolint:gochecknoglobals // chart
	"Orbit":   flyerRow(Prohibited, Prohibited, Prohibited, Prohibited, Prohibited, Accessible, Accessible, Accessible),
	"Upper":   flyerRow(Prohibited, Prohibited, Prohibited, Prohibited, Prohibited, Accessible, Accessible, Accessible),
	"Mid":     flyerRow(Prohibited, Disallowed, Prohibited, Prohibited, Prohibited, Accessible, Accessible, Accessible),
	"Low":     flyerRow(Prohibited, Disallowed, Disallowed, Disallowed, Disallowed, Accessible, Accessible, Accessible),
	"NOP":     flyerRow(Prohibited, Accessible, Accessible, Accessible, Disallowed, Accessible, Accessible, Accessible),
	"Under5m": flyerRow(Accessible, Disallowed, Disallowed, Temporary, Temporary, Accessible, Temporary, Temporary),
	"Atm0":    flyerRow(Prohibited, Prohibited, Prohibited, Prohibited, Prohibited, Accessible, Accessible, Accessible),
	"Atm1":    flyerRow(Prohibited, Prohibited, Prohibited, Prohibited, Prohibited, Accessible, Accessible, Accessible),
	"Atm2":    flyerRow(Prohibited, Accessible, Prohibited, Prohibited, Prohibited, Accessible, Accessible, Accessible),
	"Atm3":    flyerRow(Prohibited, Accessible, Prohibited, Prohibited, Prohibited, Accessible, Accessible, Accessible),
	"Atm4":    flyerRow(Prohibited, Accessible, Prohibited, Prohibited, Prohibited, Accessible, Accessible, Accessible),
	"Atm5":    flyerRow(Prohibited, Accessible, Prohibited, Prohibited, Prohibited, Accessible, Accessible, Accessible),
	"Atm6":    flyerRow(Prohibited, Accessible, Prohibited, Accessible, Accessible, Accessible, Accessible, Accessible),
	"Atm7":    flyerRow(Prohibited, Accessible, Prohibited, Accessible, Accessible, Accessible, Accessible, Accessible),
	"Atm8":    flyerRow(Prohibited, Accessible, Accessible, Accessible, Accessible, Accessible, Accessible, Accessible),
	"Atm9":    flyerRow(Prohibited, Accessible, Accessible, Accessible, Accessible, Accessible, Accessible, Accessible),
	"AtmA":    flyerRow(Prohibited, Accessible, Accessible, Accessible, Accessible, Accessible, Accessible, Accessible),
	"AtmB":    flyerRow(Prohibited, Accessible, Disallowed, Accessible, Accessible, Accessible, Accessible, Accessible),
	"AtmC":    flyerRow(Prohibited, Accessible, Disallowed, Disallowed, Accessible, Accessible, Accessible, Accessible),
	"AtmD": flyerRow(
		Prohibited,
		Conditional,
		Conditional,
		Conditional,
		Conditional,
		Accessible,
		Accessible,
		Accessible,
	),
	"AtmE": flyerRow(
		Prohibited,
		Conditional,
		Conditional,
		Conditional,
		Conditional,
		Accessible,
		Accessible,
		Accessible,
	),
	"AtmF": flyerRow(
		Prohibited,
		Conditional,
		Conditional,
		Conditional,
		Conditional,
		Accessible,
		Accessible,
		Accessible,
	),
}

// SeafaringAccess returns the p.143 water-terrain cell for Boat, Ship,
// Hydrofoil, or Sub.
func SeafaringAccess(terrain, craft string) (Access, bool) {
	row, ok := seafaringAccess[terrain]
	if !ok {
		return 0, false
	}

	a, ok := row[craft]

	return a, ok
}

var seafaringAccess = map[string]map[string]Access{ //nolint:gochecknoglobals // chart
	"Ocean":    {"Boat": Disallowed, "Ship": Accessible, "Hydrofoil": Accessible, "Sub": Accessible},
	"Islands":  {"Boat": Accessible, "Ship": Disallowed, "Hydrofoil": Accessible, "Sub": Disallowed},
	"Shore":    {"Boat": Accessible, "Ship": Disallowed, "Hydrofoil": Accessible, "Sub": Disallowed},
	"River":    {"Boat": Accessible, "Ship": Disallowed, "Hydrofoil": Disallowed, "Sub": Disallowed},
	"Sea Port": {"Boat": Accessible, "Ship": Accessible, "Hydrofoil": Accessible, "Sub": Accessible},
}

// Altitude is one row of the p.144 Vehicles and Levels table. Range is retained
// as printed because the chart includes decimal sub-bands such as 7.4.
type Altitude struct {
	Meters                                            float64
	Range                                             string
	Level                                             string
	Winged, Rotor, Flapper, LTA, Lifter, Grav, Rocket Access
}

// Altitudes returns a copy of the flyer altitude-capability table.
func Altitudes() []Altitude { return append([]Altitude(nil), altitudeRows...) }

var altitudeRows = []Altitude{ //nolint:gochecknoglobals // immutable chart
	{50_000_000, "10", "Geo", Prohibited, Prohibited, Prohibited, Prohibited, Prohibited, Prohibited, Accessible},
	{5_000_000, "9", "Far Orbit", Prohibited, Prohibited, Prohibited, Prohibited, Prohibited, Accessible, Accessible},
	{500_000, "8", "Orbit", Prohibited, Prohibited, Prohibited, Prohibited, Accessible, Accessible, Accessible},
	{200_000, "7.4", "Upper4", Prohibited, Prohibited, Prohibited, Prohibited, Accessible, Accessible, Accessible},
	{100_000, "7.2", "Upper2", Accessible, Prohibited, Prohibited, Prohibited, Accessible, Accessible, Accessible},
	{50_000, "7", "Upper", Accessible, Prohibited, Prohibited, Prohibited, Accessible, Accessible, Accessible},
	{5_000, "6", "Mid", Accessible, Prohibited, Prohibited, Accessible, Accessible, Accessible, Accessible},
	{1_000, "5", "Airspace5", Accessible, Accessible, Accessible, Accessible, Accessible, Accessible, Accessible},
	{500, "4", "Airspace4", Accessible, Accessible, Accessible, Accessible, Accessible, Accessible, Accessible},
	{150, "3", "Airspace3", Accessible, Accessible, Accessible, Accessible, Accessible, Accessible, Accessible},
	{50, "2", "NOP", Disallowed, Accessible, Accessible, Accessible, Accessible, Accessible, Accessible},
	{5, "1", "Near Surface", Disallowed, Accessible, Accessible, Accessible, Accessible, Accessible, Accessible},
	{1.5, "T", "", Disallowed, Accessible, Accessible, Accessible, Accessible, Accessible, Accessible},
	{.5, "R", "", Disallowed, Accessible, Disallowed, Accessible, Accessible, Accessible, Accessible},
	{0, "0", "Surface", Accessible, Accessible, Accessible, Accessible, Accessible, Accessible, Accessible},
}

// Depth is a p.145 ocean-depth row. Pressure is bars; zero means none is
// printed for the surface/wading/fording rows.
type Depth struct {
	Meters     float64
	Range      string
	Descriptor string
	Pressure   int
	Comment    string
}

// Depths returns a copy of the ocean depth/pressure table.
func Depths() []Depth { return append([]Depth(nil), depthRows...) }

var depthRows = []Depth{ //nolint:gochecknoglobals // immutable chart
	{150, "3", "", 0, "above surface"},
	{50, "2", "Tsunami", 4, "if Tsunami"},
	{5, "1", "Vbig Waves", 3, "if Vbig Waves"},
	{1.5, "T", "Big Waves", 2, "if Big Waves"},
	{.5, "R", "Waves", 1, "if Waves"},
	{0, "0", "Surface", 0, ""},
	{-.5, "-R", "Wading", 0, "Beast Swimmers"},
	{-1.5, "-T", "Fording", 0, "Beast Swimmers"},
	{-5, "-1", "Pond", 1, "Beast Swimmers"},
	{-50, "-2", "Thermocline", 5, "Pond Bottom"},
	{-150, "-3", "Shelf", 15, "Continental Shelf"},
	{-500, "-4", "Lake Bottom", 50, "Lake Bottom"},
	{-1_000, "-5", "Deep Lake", 100, "Deep Lake"},
	{-5_000, "-6", "Bottoms", 500, "Ocean Bottom"},
	{-50_000, "-7", "Depths", 5_000, "Max Non Ocean World"},
	{-500_000, "-8", "Abyss", 50_000, "Ocean World Abyss"},
	{-5_000_000, "-9", "", 500_000, "Never Encountered?"},
}
