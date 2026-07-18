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
// The generation is a pipeline rather than nine independent recipes, so the two
// rules that hold for every world type are structural, not something a case body
// opts into:
//
//   - Size is rolled once (typeSize) and capped once, *above* the switch. Every
//     type rolls Size first, so hoisting it consumes exactly the dice the
//     per-type roll did, and the cap necessarily lands before the characteristics
//     that derive from Size — Atmosphere is Flux+Size and Hydrographics is
//     Flux+Atmosphere (World Creation chart, p.24). Capping afterwards would
//     leave a profile describing the larger world that was rolled.
//   - Every profile returns through chartProfile, which applies the chart's
//     Size-driven rules ("If Siz=0, Atm=0", "If Siz <2, Hyd =0"). A new world
//     type cannot produce an out-of-chart profile by forgetting a helper.
func GenerateSatelliteWorld(r *dice.Roller, t OtherWorldType, mwPop, maxSize int) uwp.Profile {
	maxPop := max(mwPop-1, 0)
	capPop := func(pop int) int { return clamp(pop, 0, maxPop) }

	size := capSize(typeSize(r, t), maxSize)

	return chartProfile(rollOtherWorld(r, t, size, capPop))
}

// typeSize rolls a world type's Size formula (Book 3 p.29). It is the single
// point at which Size enters generation, which is what lets the satellite cap
// and the chart's Size-driven rules both be applied before any characteristic
// derives from Size. Planetoids yield 0 *without rolling*: a belt has no Size
// formula, and spending a die here would shift every later roll.
func typeSize(r *dice.Roller, t OtherWorldType) int {
	switch t {
	case Planetoids:
		return 0
	case Inferno:
		return 6 + r.Die()
	case BigWorld:
		return r.Dice(2) + 7
	case RadWorld, StormWorld:
		return r.Dice(2)
	case Worldlet:
		return max(r.Die()-3, 0)
	default: // Hospitable, InnerWorld, Iceworld
		return rollSize(r)
	}
}

// capSize applies the satellite-size rule (Book 3 p.21) to a rolled Size.
func capSize(size, maxSize int) int {
	if maxSize == NoSizeCap {
		return size
	}

	return min(size, maxSize)
}

// rollOtherWorld rolls the characteristics a type does not take from Size, given
// the already-rolled and already-capped size. Each case is the book's formula
// transcription and nothing more: the chart's structural conditions are applied
// for it, by fullOtherWorld and chartProfile.
func rollOtherWorld(r *dice.Roller, t OtherWorldType, size int, capPop func(int) int) uwp.Profile {
	switch t {
	case Inferno:
		// YSB0000-0: no spaceport, an exotic (B) atmosphere, a large hot world.
		return uwp.Profile{Starport: 'Y', Size: size, Atmosphere: 11}

	case RadWorld:
		// StSAH000-0: a bombarded world, uninhabited (Pop/Gov/Law/TL zero).
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
		return fullOtherWorld(r, size, 0, 0, capPop(rollPopulation(r)))

	case InnerWorld:
		atm := atmosphere(r.Flux(), size)
		hyd := r.Dice(2) - 4

		return fullOtherWorld(r, size, atm, hyd, capPop(r.Dice(2)-4))

	case StormWorld:
		atm := r.Dice(2) + 4
		hyd := r.Dice(2) - 4

		return fullOtherWorld(r, size, atm, hyd, capPop(r.Dice(2)-6))

	case BigWorld:
		atm := atmosphere(r.Flux(), size)
		hyd := hydrographics(r.Flux(), atm, size)

		return fullOtherWorld(r, size, atm, hyd, capPop(rollPopulation(r)))

	case Worldlet:
		atm := atmosphere(r.Flux(), size)
		hyd := hydrographics(r.Flux(), atm, size)

		return fullOtherWorld(r, size, atm, hyd, capPop(rollPopulation(r)))

	case Iceworld:
		atm := atmosphere(r.Flux(), size)
		hyd := hydrographics(r.Flux(), atm, size)

		return fullOtherWorld(r, size, atm, hyd, capPop(r.Dice(2)-6))

	default: // Hospitable
		atm := atmosphere(r.Flux(), size)
		hyd := hydrographics(r.Flux(), atm, size)

		return fullOtherWorld(r, size, atm, hyd, capPop(rollPopulation(r)))
	}
}

// chartProfile applies the World Creation chart's structural rules to an
// assembled profile (Book 3 p.24): "If Atm<0 or Siz=0, Atm=0" and "If Siz <2,
// Hyd =0". Every non-mainworld profile returns through here, so the rules hold
// for a type carrying its own Atm/Hyd formula — and for a future type whose
// author never learns the rules exist.
//
// Both helpers are idempotent, so a value that already consulted Size (anything
// from atmosphere/hydrographics, or from fullOtherWorld's own normalization)
// passes through untouched.
func chartProfile(p uwp.Profile) uwp.Profile {
	p.Atmosphere = sizedAtmosphere(p.Atmosphere, p.Size)
	p.Hydrographics = sizedHydrographics(p.Hydrographics, p.Size)

	return p
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
//
// Atmosphere and Hydrographics are normalized on entry rather than at assembly
// because Tech Level reads both: the chart's structural rules must be settled
// before anything derives from them.
func fullOtherWorld(r *dice.Roller, size, atm, hyd, pop int) uwp.Profile {
	atm = sizedAtmosphere(atm, size)
	hyd = sizedHydrographics(hyd, size)

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
