// Package kinematics provides Traveller's closed-form distance, signal-delay,
// and constant-acceleration calculations (Book 1 pp.31-35).
package kinematics

import (
	"math"
	"time"
)

// Traveller's rounded working constants from the Distance and Range charts.
// These deliberately reproduce the game tables rather than use more precise
// real-world measurements.
const (
	KilometersPerAU               = 150_000_000.0
	LightSpeedKilometersPerSecond = 300_000.0

	// StandardGravityMetersPerSecondSquared is the book's rounded working
	// value for one G — deliberately NOT the defined physical constant
	// 9.80665 m/s². The p.35 chart cells only reproduce with 10.
	StandardGravityMetersPerSecondSquared = 10.0
)

const metersPerKilometer = 1_000.0

// KilometersFromAU converts an astronomical distance to kilometers. A
// non-positive distance returns zero.
func KilometersFromAU(au float64) float64 {
	if au <= 0 {
		return 0
	}

	return au * KilometersPerAU
}

// LightDelay returns the one-way light-travel time across distanceKM. A
// non-positive distance returns zero.
func LightDelay(distanceKM float64) time.Duration {
	if distanceKM <= 0 {
		return 0
	}

	return seconds(distanceKM / LightSpeedKilometersPerSecond)
}

// ImpactTime returns the time to accelerate from rest across distanceKM and
// arrive at speed, using a constant acceleration measured in standard Gs
// (Book 1 p.35, Chart 9b). Non-positive inputs return zero.
func ImpactTime(distanceKM, accelerationG float64) time.Duration {
	if distanceKM <= 0 || accelerationG <= 0 {
		return 0
	}

	distanceM := distanceKM * metersPerKilometer
	acceleration := accelerationG * StandardGravityMetersPerSecondSquared

	return seconds(math.Sqrt(2 * distanceM / acceleration))
}

// StartStopTime returns the time to accelerate from rest to the midpoint of a
// distance, turn around, and decelerate to rest at the endpoint (Book 1 p.35,
// Chart 9a). Non-positive inputs return zero.
func StartStopTime(distanceKM, accelerationG float64) time.Duration {
	if distanceKM <= 0 || accelerationG <= 0 {
		return 0
	}

	distanceM := distanceKM * metersPerKilometer
	acceleration := accelerationG * StandardGravityMetersPerSecondSquared

	return seconds(2 * math.Sqrt(distanceM/acceleration))
}

// seconds converts a float second count to a time.Duration. The conversion is
// only meaningful within Duration's range (about ±292 years); NaN or
// out-of-range values produce implementation-defined garbage. In-book
// distances are far inside that range, so no clamping is done.
func seconds(value float64) time.Duration {
	return time.Duration(value * float64(time.Second))
}
