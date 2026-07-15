// Package rangeband implements Traveller5's range bands (Book 1 pp. 24-29): the
// standardized distance ladder used by senses, combat, and travel. There are two
// scales over one underlying set of distances — World-surface ranges R= (Contact
// and 0-9) and Space ranges S= (0-13) — related by S = R - 5.
package rangeband

import (
	"math"
	"strconv"
)

// A Band is one range band: its code on its scale, a descriptor, an optional
// space-combat band letter, and a representative distance in meters.
type Band struct {
	Code       string  // R= "0","R","T","1".."9" or S= "0","B","1".."13"
	Descriptor string  // e.g. "Medium", "Far Orbit", "Outer System"
	Combat     string  // space-combat band letter (B/F1/F2/SR/AR/LR/DS), if any
	Meters     float64 // representative distance from the zero point
}

// Number returns the band's numeric index on its scale and whether it has one.
// The lettered Contact sub-bands (R, T) and the Space Boarding band (B) have no
// number and return ok false.
func (b Band) Number() (int, bool) {
	n, err := strconv.Atoi(b.Code)
	if err != nil {
		return 0, false
	}
	return n, true
}

// worldBands is the World-surface range ladder R= (Book 1 p.24). Contact,
// Reading, and Talking are the lettered bands inside R=0.
var worldBands = []Band{
	{"0", "Contact", "", 0},
	{"R", "Reading", "", 0.5},
	{"T", "Talking", "", 1.5},
	{"1", "Vshort", "", 5},
	{"2", "Short", "", 50},
	{"3", "Medium", "", 150},
	{"4", "Long", "", 500},
	{"5", "Vlong", "", 1_000},
	{"6", "Distant", "", 5_000},
	{"7", "Vdistant", "", 50_000},
	{"8", "Orbit", "", 500_000},
	{"9", "Far Orbit", "", 5_000_000},
}

// spaceBands is the Space range ladder S= (Book 1 p.29). Boarding (B) is the
// lettered band between S=1 and S=0. Descriptors and combat-band letters are
// transcribed from the p.29 Space Ranges chart.
var spaceBands = []Band{
	{"0", "Contact", "", 0},
	{"B", "Boarding", "B", 1_000},
	{"1", "Close Fighter", "F1", 5_000},
	{"2", "Fighter", "F2", 50_000},
	{"3", "Orbit", "", 500_000},
	{"4", "Far Orbit", "SR", 5_000_000},
	{"5", "Short Range", "", 50_000_000},
	{"6", "Attack Range", "AR", 250_000_000},
	{"7", "Missile", "AR", 500_000_000},
	{"8", "Long Range", "LR", 2_500_000_000},
	{"9", "Long Range", "LR", 5_000_000_000},
	{"10", "Siege", "", 50_000_000_000},
	{"11", "Deep Space", "DS", 150_000_000_000},
	{"12", "Deep Space", "DS", 500_000_000_000},
	{"13", "Outer System", "", 1_500_000_000_000},
}

// WorldBands and SpaceBands return copies of the two ladders, in order.
func WorldBands() []Band { return append([]Band(nil), worldBands...) }
func SpaceBands() []Band { return append([]Band(nil), spaceBands...) }

// WorldBand returns the R= band with the given code.
func WorldBand(code string) (Band, bool) { return find(worldBands, code) }

// SpaceBand returns the S= band with the given code.
func SpaceBand(code string) (Band, bool) { return find(spaceBands, code) }

func find(bands []Band, code string) (Band, bool) {
	for _, b := range bands {
		if b.Code == code {
			return b, true
		}
	}
	return Band{}, false
}

// WorldForDistance returns the World-surface band nearest a distance in meters,
// choosing the band whose representative distance is closest on a log scale
// (each band covers roughly half-way to its neighbours; Book 1 p.24).
func WorldForDistance(meters float64) Band { return nearest(worldBands, meters) }

// SpaceForDistance returns the Space band nearest a distance in meters.
func SpaceForDistance(meters float64) Band { return nearest(spaceBands, meters) }

func nearest(bands []Band, meters float64) Band {
	if meters <= 0 {
		return bands[0] // Contact / zero point
	}
	lm := math.Log10(meters)
	best, bestGap := bands[0], math.Inf(1)
	for _, b := range bands {
		if b.Meters <= 0 {
			continue // the zero-point band has no finite distance
		}
		if gap := math.Abs(lm - math.Log10(b.Meters)); gap < bestGap {
			best, bestGap = b, gap
		}
	}
	return best
}

// WorldToSpace converts an R= code to its S= code: S = R - 5. R=5 is Boarding
// (B); R=4 or less collapses to S=0 (Book 1 p.29).
func WorldToSpace(worldCode string) (string, bool) {
	if _, ok := WorldBand(worldCode); !ok {
		return "", false
	}
	switch worldCode {
	case "5":
		return "B", true
	case "6", "7", "8", "9":
		n, _ := strconv.Atoi(worldCode)
		return strconv.Itoa(n - 5), true
	default: // 0, R, T, 1-4 all collapse to S=0
		return "0", true
	}
}

// SpaceToWorld converts an S= code to its R= code: R = S + 5. Boarding (B) maps
// to R=5, S=0 to R=0 (Book 1 p.29). Codes S>4 give R>9, beyond the World scale's
// named bands but valid on the shared ladder.
func SpaceToWorld(spaceCode string) (string, bool) {
	if _, ok := SpaceBand(spaceCode); !ok {
		return "", false
	}
	switch spaceCode {
	case "0":
		return "0", true
	case "B":
		return "5", true
	default:
		n, _ := strconv.Atoi(spaceCode)
		return strconv.Itoa(n + 5), true
	}
}

// WorldSubBand returns the fractional World range band for a distance, log-
// interpolating between the bracketing numeric bands 1..9 (Book 1 p.25: a sub-
// band such as 6.4 locates a range between two whole bands). Distances at or
// below R=1 return 1; at or above R=9, 9.
func WorldSubBand(meters float64) float64 {
	numeric := worldBands[3:] // R=1 .. R=9
	if meters <= numeric[0].Meters {
		return 1
	}
	for i := 1; i < len(numeric); i++ {
		if meters <= numeric[i].Meters {
			lo, hi := numeric[i-1], numeric[i]
			frac := (math.Log10(meters) - math.Log10(lo.Meters)) /
				(math.Log10(hi.Meters) - math.Log10(lo.Meters))
			return float64(i) + frac // band i is R=(i+1); index i-1 is R=i
		}
	}
	return 9
}
