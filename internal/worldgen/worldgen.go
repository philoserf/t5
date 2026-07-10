// Package worldgen creates a Traveller5 mainworld Universal World Profile from
// the dice, following the world-creation rules in Book 3 (the Chart C
// checklist and the World Creation card, pp. 16-25).
//
// Each characteristic is a small formula over the dice and the characteristics
// already rolled. The pure formula functions (atmosphere, hydrographics, …)
// take their rolls as arguments so they can be tested at their edges; Generate
// rolls in checklist order and composes them.
package worldgen

import (
	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/uwp"
)

// eHex maxima for the characteristics that clamp, per the World Creation card.
const (
	maxAtmosphere    = 15 // F
	maxHydrographics = 10 // A
	maxGovernment    = 15 // F
	maxLaw           = 18 // J
)

// Generate rolls a complete mainworld UWP in checklist order: Starport, Size,
// Atmosphere, Hydrographics, Population, Government, Law, then Tech Level.
func Generate(r *dice.Roller) uwp.Profile {
	sp := starport(r.Dice(2))
	size := rollSize(r)
	atm := atmosphere(r.Flux(), size)
	hyd := hydrographics(r.Flux(), atm, size)
	pop := rollPopulation(r)
	gov := government(r.Flux(), pop)
	lawLevel := law(r.Flux(), gov)
	return uwp.Profile{
		Starport:      sp,
		Size:          size,
		Atmosphere:    atm,
		Hydrographics: hyd,
		Population:    pop,
		Government:    gov,
		Law:           lawLevel,
		TechLevel:     techLevel(r.Die(), sp, size, atm, hyd, pop, gov),
	}
}

// starport maps a 2D roll to a starport quality letter (Book 3 table 1).
func starport(twoD int) byte {
	switch {
	case twoD <= 4:
		return 'A'
	case twoD <= 6:
		return 'B'
	case twoD <= 8:
		return 'C'
	case twoD == 9:
		return 'D'
	case twoD <= 11:
		return 'E'
	default:
		return 'X'
	}
}

// rollSize is 2D-2; a result of 10 is rerolled as 1D+9 for the largest worlds.
func rollSize(r *dice.Roller) int {
	if s := r.Dice(2) - 2; s != 10 {
		return s
	}
	return r.Die() + 9
}

// atmosphere is Flux+Size, forced to 0 for a size-0 world and clamped to F.
func atmosphere(flux, size int) int {
	if size == 0 {
		return 0
	}
	return clamp(flux+size, 0, maxAtmosphere)
}

// hydrographics is Flux+Atmosphere, zero for worlds smaller than size 2, with a
// -4 modifier for thin or very dense atmospheres, clamped to A.
func hydrographics(flux, atm, size int) int {
	if size < 2 {
		return 0
	}
	h := flux + atm
	if atm < 2 || atm > 9 {
		h -= 4
	}
	return clamp(h, 0, maxHydrographics)
}

// rollPopulation is 2D-2; a result of 10 is rerolled as 2D+3.
func rollPopulation(r *dice.Roller) int {
	if p := r.Dice(2) - 2; p != 10 {
		return p
	}
	return r.Dice(2) + 3
}

// government is Flux+Population, clamped to F.
func government(flux, pop int) int {
	return clamp(flux+pop, 0, maxGovernment)
}

// law is Flux+Government, clamped to J.
func law(flux, gov int) int {
	return clamp(flux+gov, 0, maxLaw)
}

// techLevel is 1D plus modifiers drawn from the starport and every
// characteristic, per the World Creation card, and never falls below 0.
func techLevel(oneD int, sp byte, size, atm, hyd, pop, gov int) int {
	tl := oneD
	switch sp {
	case 'A':
		tl += 6
	case 'B':
		tl += 4
	case 'C':
		tl += 2
	case 'F':
		tl++
	case 'X':
		tl -= 4
	}
	switch size {
	case 0, 1:
		tl += 2
	case 2, 3, 4:
		tl++
	}
	if atm <= 3 || atm >= 10 {
		tl++
	}
	switch hyd {
	case 9:
		tl++
	case 10:
		tl += 2
	}
	switch {
	case pop >= 1 && pop <= 5:
		tl++
	case pop == 9:
		tl += 2
	case pop >= 10:
		tl += 4
	}
	switch gov {
	case 0, 5:
		tl++
	case 13:
		tl -= 2
	}
	return max(tl, 0)
}

func clamp(v, lo, hi int) int {
	return min(max(v, lo), hi)
}
