// Package rangeband implements Traveller5's range bands (Book 1 pp.24-29): the
// standardized distance ladder used by senses, combat, and travel. There are two
// scales over one underlying set of distances — World-surface ranges R= (Contact
// and 0-9) and Space ranges S= (0-13) — related by S = R - 5.
package rangeband

import (
	"math"
	"strconv"
)

// A Band is one range band: its code on its scale, a descriptor, an optional
// space-combat band letter, a representative distance in meters, and the upper
// bound of the span it covers.
type Band struct {
	Code       string  // R= "0","R","T","1".."9" or S= "0","B","1".."13"
	Descriptor string  // e.g. "Medium", "Far Orbit", "Outer System"
	Combat     string  // space-combat band letter (B/F1/F2/SR/AR/LR/DS), if any
	Meters     float64 // representative distance from the zero point
	Upper      float64 // inclusive top of the band's span; 0 when the book prints none
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
// Reading, and Talking are the lettered bands inside R=0. Upper is the p.24
// Range Band Width column, whose spans are *not* the log midpoints between
// representative distances — Vshort's 5 m band runs 3 m to 25 m, so 20 m is
// still Vshort even though it is nearer 50 m on a log scale. The book prints
// shared edges (Vshort "3 m to 25 m", Short "25 m to 100 m"), resolved here as
// upper-inclusive. The lettered bands nest inside Contact's own 3 m span, so
// they are ordered finest-first: a distance takes the narrowest band covering it.
var worldBands = []Band{
	{"0", "Contact", "", 0, 0.25},
	{"R", "Reading", "", 0.5, 1},
	{"T", "Talking", "", 1.5, 3},
	{"1", "Vshort", "", 5, 25},
	{"2", "Short", "", 50, 100},
	{"3", "Medium", "", 150, 300},
	{"4", "Long", "", 500, 750},
	{"5", "Vlong", "", 1_000, 3_000},
	{"6", "Distant", "", 5_000, 25_000},
	{"7", "Vdistant", "", 50_000, 250_000},
	{"8", "Orbit", "", 500_000, 2_500_000},
	{"9", "Far Orbit", "", 5_000_000, 25_000_000},
}

// spaceBands is the Space range ladder S= (Book 1 p.29). Boarding (B) is the
// lettered band between S=1 and S=0. Descriptors and combat-band letters are
// transcribed from the p.29 Space Ranges chart. The chart's Band column stacks
// each two-word combat-band name across the pair of rows it covers ("Short" on
// S=5 above "Range" on S=4), so every combat letter spans two rows: SR on 4-5,
// AR on 6-7, LR on 8-9, DS on 11-12. S=0, 3, 10, and 13 sit in no combat band.
// The chart prints no Range Band Width column, so Upper is left zero here.
var spaceBands = []Band{
	{"0", "Contact", "", 0, 0},
	{"B", "Boarding", "B", 1_000, 0},
	{"1", "Close Fighter", "F1", 5_000, 0},
	{"2", "Fighter", "F2", 50_000, 0},
	{"3", "Orbit", "", 500_000, 0},
	{"4", "Far Orbit", "SR", 5_000_000, 0},
	{"5", "Short Range", "SR", 50_000_000, 0},
	{"6", "Attack Range", "AR", 250_000_000, 0},
	{"7", "Missile", "AR", 500_000_000, 0},
	{"8", "Long Range", "LR", 2_500_000_000, 0},
	{"9", "Long Range", "LR", 5_000_000_000, 0},
	{"10", "Siege", "", 50_000_000_000, 0},
	{"11", "Deep Space", "DS", 150_000_000_000, 0},
	{"12", "Deep Space", "DS", 500_000_000_000, 0},
	{"13", "Outer System", "", 1_500_000_000_000, 0},
}

// WorldBands returns a copy of the World-surface range ladder, in order.
func WorldBands() []Band { return append([]Band(nil), worldBands...) }

// SpaceBands returns a copy of the Space range ladder, in order.
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

// WorldForDistance returns the World-surface band containing a distance in
// meters, by the Range Band Width column printed on Book 1 p.24 — the first band
// whose Upper bound the distance does not exceed, bands running finest-first so
// the lettered Reading/Talking sub-bands win inside Contact's span. Beyond the
// top of the ladder (25,000 km) the distance is Far Orbit. The p.24 widths are
// not the log midpoints between representative distances, so this deliberately
// does not use nearest: 20 m is Vshort (3-25 m), not Short.
func WorldForDistance(meters float64) Band {
	if meters <= 0 {
		return worldBands[0] // Contact / zero point
	}

	for _, b := range worldBands {
		if meters <= b.Upper {
			return b
		}
	}

	return worldBands[len(worldBands)-1] // past Far Orbit's 25,000 km
}

// SpaceForDistance returns the Space band nearest a distance in meters on a log
// scale. The p.29 Space Ranges chart prints no Range Band Width column — only a
// representative distance per row — so unlike the World ladder there are no
// published boundaries to select by.
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
// (B); R=4 or less collapses to S=0 (Book 1 p.29). It is the inverse of
// SpaceToWorld over the whole shared ladder, so it accepts the extended numeric
// codes R=10..18 that lie past the World scale's named bands but that the p.29
// chart's own R= column prints (S=13 is R=18) and SpaceToWorld produces.
func WorldToSpace(worldCode string) (string, bool) {
	if worldCode == "5" {
		return "B", true
	}

	if b, named := WorldBand(worldCode); named {
		n, numeric := b.Number()
		if !numeric || n < 5 { // 0, R, T, 1-4 all collapse to S=0
			return "0", true
		}

		return strconv.Itoa(n - 5), true
	}
	// The extended tail of the shared ladder: R=10..18 name no World band of
	// their own, but the chart pairs each with an S= row.
	n, err := strconv.Atoi(worldCode)
	if err != nil || n < 10 || n > 18 {
		return "", false
	}

	return strconv.Itoa(n - 5), true
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
