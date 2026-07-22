package vehiclegen

// Endurance is the operating-time class selected on Chart H.
type Endurance uint8

const (
	// Minutes is less than an hour of operation.
	Minutes Endurance = iota
	// Hours is less than a day of operation.
	Hours
	// Days is less than a week of operation.
	Days
	// Weeks is less than a month of operation.
	Weeks
	// Months is less than a year of operation.
	Months
	// Year is one or more years of operation.
	Year
)

// Range is the planetographic range class produced from Speed and Endurance.
type Range uint8

const (
	// Local is travel within a Terrain Hex.
	Local Range = iota
	// Regional is travel within a World Hex.
	Regional
	// Continental is travel within a World Triangle.
	Continental
	// World is travel anywhere on the world.
	World
)

// EnduranceRange converts operating-time class and Vehicle Speed to the range
// class printed on pp.147/153. The chart covers Speeds 1-13.
func EnduranceRange(endurance Endurance, speed int) (Range, bool) { //nolint:cyclop // direct chart transcription
	if speed < 1 || speed > 13 || endurance > Year {
		return 0, false
	}

	switch endurance {
	case Minutes:
		return Local, true
	case Hours:
		switch {
		case speed <= 4:
			return Local, true
		case speed <= 6:
			return Regional, true
		case speed <= 9:
			return Continental, true
		default:
			return World, true
		}
	case Days:
		if speed <= 3 {
			return Regional, true
		}

		if speed <= 6 {
			return Continental, true
		}

		return World, true
	case Weeks, Months:
		if speed == 1 {
			return Continental, true
		}

		return World, true
	case Year:
		return World, true
	default:
		return 0, false
	}
}

// Bulk is the VehicleMaker relative-size class.
type Bulk int8

const (
	// VLight is the smallest bulk class.
	VLight Bulk = -2
	// Light is smaller than standard.
	Light Bulk = -1
	// Medium is standard bulk.
	Medium Bulk = 0
	// Heavy carries larger loads than standard.
	Heavy Bulk = 1
	// VHeavy is the largest bulk class.
	VHeavy Bulk = 2
)

// Occupants returns normal human-sized occupant capacity (Book 3 pp.139,147).
// Standard and larger vehicles provide one place per N tons, where N rises with
// endurance; fractional capacity is discarded. Vlight/Lite special cases are
// exactly those printed by the chart.
func Occupants(bulk Bulk, tons float64, endurance Endurance) int {
	if tons <= 0 || bulk < VLight || bulk > VHeavy || endurance > Year {
		return 0
	}

	if bulk == VLight {
		if endurance == Hours || endurance == Minutes {
			return 1
		}

		return 0
	}

	if bulk == Light {
		switch endurance {
		case Minutes, Hours:
			return 2
		case Days:
			return 1
		default:
			return 0
		}
	}

	divisor := [...]int{1, 1, 2, 3, 4, 5}[endurance]

	return int(tons) / divisor
}
