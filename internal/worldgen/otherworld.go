package worldgen

import (
	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/uwp"
)

// OtherWorldType is a secondary (non-mainworld) world type (Book 3 p.29). Each
// type generates a partial UWP: some characteristics roll normally (the
// StSAHPGL-T mnemonic marks which), others are fixed or overridden by a formula.
type OtherWorldType int

// Types of non-mainworld in a system.
const (
	Hospitable OtherWorldType = iota // StSAHPGL-T (a full, habitable world)
	InnerWorld                       // StSAHPGL-T; Pop = 2D-4, Hyd = 2D-4
	BigWorld                         // StSAHPGL-T; Siz = 2D+7
	StormWorld                       // StSAHPGL-T; Siz = 2D, Atm = 2D+4, Hyd = 2D-4, Pop = 2D-6
	RadWorld                         // StSAH000-0; Siz = 2D
	Inferno                          // YSB0000-0; Siz = 6+1D
	Worldlet                         // StSAHPGL-T; Siz = 1D-3
	Iceworld                         // StSAHPGL-T; Pop = 2D-6
	Planetoids                       // St000PGL-T
)

// String names the world type for display.
func (t OtherWorldType) String() string {
	switch t {
	case InnerWorld:
		return "Inner World"
	case BigWorld:
		return "Big World"
	case StormWorld:
		return "Storm World"
	case RadWorld:
		return "Rad World"
	case Inferno:
		return "Inferno"
	case Worldlet:
		return "Worldlet"
	case Iceworld:
		return "Iceworld"
	case Planetoids:
		return "Planetoids"
	default:
		return "Hospitable"
	}
}

// spaceport maps a non-mainworld's Population minus a 1D roll to a spaceport
// class (Book 3 p.29 table 1B): 4+ Good (F), 3 Poor (G), 1-2 Basic (H), else
// None (Y). Non-mainworlds carry spaceports, never the A-E/X starports.
func spaceport(popLess1D int) byte {
	switch {
	case popLess1D >= 4:
		return 'F'
	case popLess1D == 3:
		return 'G'
	case popLess1D >= 1:
		return 'H'
	default:
		return 'Y'
	}
}

// NoSizeCap is the maxSize argument for a world under no parent-size limit —
// any world that is not a satellite (Book 3 p.21 caps only satellites).
const NoSizeCap = -1

// GenerateOtherWorld rolls a secondary world's UWP for the given type (Book 3
// p.29). Every world's population is capped at mwPop-1 (the mainworld stays the
// most populous), and DM-modified characteristics are read as a fresh 2D plus
// the shown modifier, floored at zero and clamped to the characteristic maximum.
func GenerateOtherWorld(r *dice.Roller, t OtherWorldType, mwPop int) uwp.Profile {
	return GenerateSatelliteWorld(r, t, mwPop, NoSizeCap)
}

// GenerateSatelliteWorld rolls a secondary world's UWP (Book 3 p.29) whose Size
// is limited to maxSize — the satellite-size rule, "a satellite is always
// smaller than its parent; if its size is generated as larger than the parent,
// adjust to fit" (Book 3 p.21). Pass NoSizeCap for an unlimited world.
//
// The cap is applied to Size as it is rolled, *before* the characteristics that
// derive from it, because Atmosphere is Flux+Size and Hydrographics is
// Flux+Atmosphere (World Creation chart, p.24) — adjusting Size after the fact
// would leave a profile describing the larger world that was rolled, breaking
// the chart's own structural rules ("If Siz=0, Atm=0", "If Siz <2, Hyd =0") and
// misclassifying the world's trade codes. Capping in place consumes exactly the
// same dice in the same order as the uncapped roll, so it re-derives the
// dependent characteristics rather than re-rolling them.
func GenerateSatelliteWorld(r *dice.Roller, t OtherWorldType, mwPop, maxSize int) uwp.Profile {
	maxPop := max(mwPop-1, 0)
	capPop := func(pop int) int { return clamp(pop, 0, maxPop) }

	capSize := func(size int) int {
		if maxSize != NoSizeCap && size > maxSize {
			return maxSize
		}

		return size
	}

	switch t {
	case Inferno:
		// YSB0000-0: no spaceport, an exotic (B) atmosphere, a large hot world.
		// A capped Inferno keeps its defining atmosphere unless it is cut to
		// Size 0, which the chart forces airless.
		size := capSize(6 + r.Die())
		atm := 11

		if size == 0 {
			atm = 0
		}

		return uwp.Profile{Starport: 'Y', Size: size, Atmosphere: atm}

	case RadWorld:
		// StSAH000-0: a bombarded world, uninhabited (Pop/Gov/Law/TL zero).
		size := capSize(r.Dice(2))
		atm := atmosphere(r.Flux(), size)
		hyd := hydrographics(r.Flux(), atm, size)

		return uwp.Profile{
			Starport:      spaceport(-r.Die()),
			Size:          size,
			Atmosphere:    atm,
			Hydrographics: hyd,
		}

	case Planetoids:
		// St000PGL-T: a belt-like body — no size, atmosphere, or hydrographics.
		return fullOtherWorld(r, 0, 0, 0, capPop(rollPopulation(r)))

	case InnerWorld:
		size := capSize(rollSize(r))
		atm := atmosphere(r.Flux(), size)
		hyd := sizedHydrographics(r.Dice(2)-4, size)

		return fullOtherWorld(r, size, atm, hyd, capPop(r.Dice(2)-4))

	case BigWorld:
		size := capSize(r.Dice(2) + 7)
		atm := atmosphere(r.Flux(), size)
		hyd := hydrographics(r.Flux(), atm, size)

		return fullOtherWorld(r, size, atm, hyd, capPop(rollPopulation(r)))

	case StormWorld:
		size := capSize(r.Dice(2))
		atm := sizedAtmosphere(r.Dice(2)+4, size)
		hyd := sizedHydrographics(r.Dice(2)-4, size)

		return fullOtherWorld(r, size, atm, hyd, capPop(r.Dice(2)-6))

	case Worldlet:
		size := capSize(max(r.Die()-3, 0))
		atm := atmosphere(r.Flux(), size)
		hyd := hydrographics(r.Flux(), atm, size)

		return fullOtherWorld(r, size, atm, hyd, capPop(rollPopulation(r)))

	case Iceworld:
		size := capSize(rollSize(r))
		atm := atmosphere(r.Flux(), size)
		hyd := hydrographics(r.Flux(), atm, size)

		return fullOtherWorld(r, size, atm, hyd, capPop(r.Dice(2)-6))

	default: // Hospitable
		size := capSize(rollSize(r))
		atm := atmosphere(r.Flux(), size)
		hyd := hydrographics(r.Flux(), atm, size)

		return fullOtherWorld(r, size, atm, hyd, capPop(rollPopulation(r)))
	}
}

// sizedAtmosphere clamps a type's own Atmosphere formula (StormWorld's 2D+4,
// which does not read Size) to the chart's structural rule: a Size-0 world is
// airless (World Creation chart, p.24, "If Siz=0, Atm=0").
func sizedAtmosphere(atm, size int) int {
	if size == 0 {
		return 0
	}

	return clamp(atm, 0, maxAtmosphere)
}

// sizedHydrographics clamps a type's own Hydrographics formula (the InnerWorld
// and StormWorld 2D-4, neither of which reads Size) to the chart's structural
// rule: a world smaller than Size 2 holds no water (p.24, "If Siz <2, Hyd =0").
func sizedHydrographics(hyd, size int) int {
	if size < 2 {
		return 0
	}

	return clamp(hyd, 0, maxHydrographics)
}

// fullOtherWorld finishes a world whose size/atmosphere/hydrographics/population
// are already rolled: Government, Law, the spaceport (from Population), and Tech
// Level, in that order (Book 3 p.29 StSAHPGL-T).
func fullOtherWorld(r *dice.Roller, size, atm, hyd, pop int) uwp.Profile {
	gov := government(r.Flux(), pop)
	lawLevel := law(r.Flux(), gov)
	sp := spaceport(pop - r.Die())

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
