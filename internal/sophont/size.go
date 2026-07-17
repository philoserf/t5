package sophont

// Sophont size, transcribed from Book 3 pp.236-237 (Sophont Creation steps 14-15),
// verified against a rendered image of the page. Size is a species' average
// volume in liters, approximately equal to its mass in kg, computed from the
// three physical characteristics' die counts. Human (all 2D) is the benchmark,
// Size 72.

import "math"

// physicalHalfDice returns twice the weighted physical-dice total (chart 14A):
// C1 Str full weight, C2 Dex full / Grace or Agility half, C3 End full / Stamina
// double / Vigor half. Working in half-dice keeps the weighting exact without
// floating point (every multiplier below is even, so Size stays an integer).
func physicalHalfDice(chars [6]CharSpec) int {
	half := 2 * chars[0].Dice // C1 Str, full weight
	half += c2HalfWeight(chars[1])
	half += c3HalfWeight(chars[2])

	return half
}

func c2HalfWeight(c CharSpec) int {
	if c.Name == Gra || c.Name == Agi {
		return c.Dice // half of the full-weight 2*Dice
	}

	return 2 * c.Dice // Dex, full weight
}

func c3HalfWeight(c CharSpec) int {
	switch c.Name {
	case Sta:
		return 4 * c.Dice // double weight
	case Vig:
		return c.Dice // half weight
	default:
		return 2 * c.Dice // End, full weight
	}
}

// sizeMultiplier is 12 for an ordinary sophont (an average 3.5 per die × ~3.4 kg
// of flesh per characteristic point). A Bulk sophont — Str 4D or greater — is
// disproportionately large and scales super-linearly at StrDice*12 (chart 14C:
// 4D=48, 5D=60 … 8D=96).
func sizeMultiplier(strDice int) int {
	if strDice <= 3 {
		return 12
	}

	return strDice * 12
}

// Size returns a species' average Size — its volume in liters, roughly equal to
// its mass in kg (Book 3 p.236). Human (all 2D) = 72, the benchmark.
//
// The Bulk table (14C) writes the total as a raw "(C1+C2+C3) Dice", but the
// weighting from 14A (half-Grace, double-Stamina …) is the single stated rule,
// so it applies here too; for a no-analog species the two agree. Note the page's
// own Virushi prose (10D × 120 = 1200) contradicts its Bulk table (5D → × 60,
// i.e. 600); the table governs and no Virushi figure is treated as canonical.
func Size(chars [6]CharSpec) int {
	return physicalHalfDice(chars) * sizeMultiplier(chars[0].Dice) / 2
}

// Height returns a sophont's height (or length, if horizontal) in meters from
// its Size — or Bulk, both in kg — and Body Form Profile (Book 3 pp.236-237):
// the cube root of the volume in m3 scaled by BFP squared. A Size-72, BFP-9
// Human is ~1.8 m. This reproduces the book's precomputed Height/Bulk grids
// exactly, so the grids are not transcribed. The BFP itself comes from the body
// structure step, which is deferred; call this with a chosen BFP (9 = typical
// Human) once that step exists.
func Height(sizeKg int, bfp float64) float64 {
	return math.Cbrt(float64(sizeKg) / 1000 * bfp * bfp)
}
