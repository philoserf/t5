package worldgen

import (
	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/uwp"
)

// OtherWorldType is a secondary (non-mainworld) world type (Book 3 p.29). Each
// type generates a partial UWP: some characteristics roll normally (the
// StSAHPGL-T mnemonic marks which), others are fixed or overridden by a formula.
type OtherWorldType int

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

// GenerateOtherWorld rolls a secondary world's UWP for the given type (Book 3
// p.29). Every world's population is capped at mwPop-1 (the mainworld stays the
// most populous), and DM-modified characteristics are read as a fresh 2D plus
// the shown modifier, floored at zero and clamped to the characteristic maximum.
func GenerateOtherWorld(r *dice.Roller, t OtherWorldType, mwPop int) uwp.Profile {
	maxPop := max(mwPop-1, 0)
	capPop := func(pop int) int { return clamp(pop, 0, maxPop) }

	switch t {
	case Inferno:
		// YSB0000-0: no spaceport, an exotic (B) atmosphere, a large hot world.
		return uwp.Profile{Starport: 'Y', Size: 6 + r.Die(), Atmosphere: 11}

	case RadWorld:
		// StSAH000-0: a bombarded world, uninhabited (Pop/Gov/Law/TL zero).
		size := r.Dice(2)
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
		size := rollSize(r)
		atm := atmosphere(r.Flux(), size)
		hyd := clamp(r.Dice(2)-4, 0, maxHydrographics)
		return fullOtherWorld(r, size, atm, hyd, capPop(r.Dice(2)-4))

	case BigWorld:
		size := r.Dice(2) + 7
		atm := atmosphere(r.Flux(), size)
		hyd := hydrographics(r.Flux(), atm, size)
		return fullOtherWorld(r, size, atm, hyd, capPop(rollPopulation(r)))

	case StormWorld:
		size := r.Dice(2)
		atm := clamp(r.Dice(2)+4, 0, maxAtmosphere)
		hyd := clamp(r.Dice(2)-4, 0, maxHydrographics)
		return fullOtherWorld(r, size, atm, hyd, capPop(r.Dice(2)-6))

	case Worldlet:
		size := max(r.Die()-3, 0)
		atm := atmosphere(r.Flux(), size)
		hyd := hydrographics(r.Flux(), atm, size)
		return fullOtherWorld(r, size, atm, hyd, capPop(rollPopulation(r)))

	case Iceworld:
		size := rollSize(r)
		atm := atmosphere(r.Flux(), size)
		hyd := hydrographics(r.Flux(), atm, size)
		return fullOtherWorld(r, size, atm, hyd, capPop(r.Dice(2)-6))

	default: // Hospitable
		size := rollSize(r)
		atm := atmosphere(r.Flux(), size)
		hyd := hydrographics(r.Flux(), atm, size)
		return fullOtherWorld(r, size, atm, hyd, capPop(rollPopulation(r)))
	}
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
