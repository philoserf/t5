// Package vehiclegen implements the Traveller5 VehicleMaker design charts
// (Book 3 pp.132-158). Designs are assembled by applying chart-row column
// modifiers; the supporting speed, power, endurance, range, and occupant tables
// are exposed as pure functions.
package vehiclegen

var speedKPH = [...]int{
	0, 5, 10, 20, 30, 50, 100, 300, 500, 700, 1_000, 2_000, 3_000, 5_000,
	10_000, 20_000, 30_000,
}

// SpeedKPH returns the representative kilometres per hour for Vehicle Speed
// 0-16 (Book 3 p.147). Speeds 17-19 have collision values but no printed kph.
func SpeedKPH(speed int) (int, bool) {
	if speed < 0 || speed >= len(speedKPH) {
		return 0, false
	}

	return speedKPH[speed], true
}

// CollisionDicePerTon is Blow damage dice per ton at relative Speed. The chart
// is the square of Speed and extends through Speed 19.
func CollisionDicePerTon(speed int) (int, bool) {
	if speed < 0 || speed > 19 {
		return 0, false
	}

	return speed * speed, true
}

// Beastpower returns BP = tons * Speed^3 (Book 3 p.146).
func Beastpower(tons float64, speed int) float64 {
	if tons <= 0 || speed < 0 {
		return 0
	}

	return tons * float64(speed*speed*speed)
}
