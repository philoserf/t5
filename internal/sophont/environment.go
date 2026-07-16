package sophont

// Sophont environment, transcribed from Book 3 p.227 (Sophont Creation step 05),
// verified against a rendered image of the page. The evolutionary environment
// feeds characteristic generation: the Native Terrain roll is the Environ DM
// (added to physical die counts), the locomotion decides the characteristic-name
// DM, and a carnivore's sub-niche shifts C3.

import (
	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/uwp"
)

// A Locomotion is a sophont's natural mode of movement (chart 05B).
type Locomotion int

const (
	Walker Locomotion = iota
	Amphibian
	Triphibian
	Aquatic
	Diver
	Flyer
	Flyphib
	Swimmer
)

func (l Locomotion) String() string {
	return [...]string{"Walker", "Amphibian", "Triphibian", "Aquatic", "Diver", "Flyer", "Flyphib", "Swimmer"}[l]
}

// An Environment is a species' evolutionary origin: its native terrain (whose
// Flux roll is the Environ DM), its locomotion, and its ecological niche (a basic
// class plus a sub-niche).
type Environment struct {
	Terrain    string
	EnvironDM  int
	Locomotion Locomotion
	Class      string // basic ecological class: Producer/Herbivore/Omnivore/Carnivore/Scavenger
	Niche      string // the sub-niche within that class
}

// locomotionMod is the characteristic-name DM (chart 06A): a Flyer is more
// likely to have the analog names (Agi/Sta/Ins), a Swimmer or Diver the other
// analogs (Gra/Vig/Tra).
func (e Environment) locomotionMod() int {
	switch e.Locomotion {
	case Flyer:
		return -2
	case Swimmer, Diver:
		return 2
	default:
		return 0
	}
}

// c3NicheMod is the C3 die-count DM (chart 06B): a carnivore Chaser is +2, a
// Pouncer -2.
func (e Environment) c3NicheMod() int {
	switch e.Niche {
	case "Chaser":
		return 2
	case "Pouncer":
		return -2
	default:
		return 0
	}
}

// terrainNames is the Native Terrain table (chart 05A), indexed by Flux+5. The
// Flux value is the Environ DM.
//
// Chart 05A also renames the -1/0/+1 terrains (Baked Lands / Twilight Zone /
// Frozen Lands) when the homeworld is a Twilight-Zone or tide-Locked world. That
// substitution is not modeled: it needs the homeworld's orbit/climate, which the
// reused worldgen mainworld does not carry (the Tz/Lk codes are deferred to
// system placement, and a standalone sophont homeworld has no star context).
var terrainNames = []string{
	"Mountain", "Desert", "Exotic", "Rough Wood", "Rough",
	"Clear", "Forest", "Wetlands", "Wetland Woods", "Ocean", "Ocean Depths",
}

// locomotionGrid is the Native Terrain & Locomotion table (chart 05B): terrain
// rows (Flux -5..+5) by a 1D column (0..5).
var locomotionGrid = [][]Locomotion{
	{Walker, Walker, Walker, Walker, Walker, Flyer},          // -5 Mountain
	{Walker, Walker, Walker, Walker, Walker, Flyer},          // -4 Desert
	{Amphibian, Walker, Walker, Walker, Flyphib, Flyer},      // -3 Exotic
	{Amphibian, Walker, Walker, Walker, Walker, Flyer},       // -2 Rough Wood
	{Amphibian, Walker, Walker, Walker, Walker, Flyer},       // -1 Rough
	{Walker, Walker, Walker, Walker, Walker, Walker},         // 0 Clear
	{Walker, Walker, Walker, Walker, Walker, Walker},         // +1 Forest
	{Amphibian, Aquatic, Walker, Walker, Triphibian, Flyer},  // +2 Wetlands
	{Amphibian, Walker, Walker, Walker, Triphibian, Flyphib}, // +3 Wetland Woods
	{Flyphib, Swimmer, Swimmer, Swimmer, Aquatic, Diver},     // +4 Ocean
	{Aquatic, Diver, Diver, Diver, Diver, Diver},             // +5 Ocean Depths
}

// nicheClasses is the basic ecological class (the "Niche" column of chart 05C),
// indexed by Flux+6 (Flux -6..+6).
var nicheClasses = []string{
	"Producer", "Producer", "Herbivore", "Herbivore", "Omnivore", "Omnivore", "Omnivore",
	"Omnivore", "Omnivore", "Carnivore", "Carnivore", "Scavenger", "Scavenger",
}

// nicheCols holds the sub-niche columns of chart 05C, each indexed by Flux+6.
var nicheCols = map[string][]string{
	"Herbivore": {
		"Grazer",
		"Grazer",
		"Grazer",
		"Intermittent",
		"Intermittent",
		"Intermittent",
		"Intermittent",
		"Grazer",
		"Grazer",
		"Grazer",
		"Grazer",
		"Grazer",
		"Filter",
	},
	"Omnivore": {
		"Hunter",
		"Hunter",
		"Hunter",
		"Hunter",
		"Hunter",
		"Gatherer",
		"H/G",
		"Gatherer",
		"Gatherer",
		"Gatherer",
		"Gatherer",
		"Gatherer",
		"Eater",
	},
	"Carnivore": {
		"Pouncer",
		"Pouncer",
		"Pouncer",
		"Pouncer",
		"Pouncer",
		"Pouncer",
		"Chaser",
		"Chaser",
		"Chaser",
		"Chaser",
		"Trapper",
		"Siren",
		"Killer",
	},
	"Scavenger": {
		"Carrion-Eater",
		"Carrion-Eater",
		"Carrion-Eater",
		"Hijacker",
		"Hijacker",
		"Hijacker",
		"Intimidator",
		"Intimidator",
		"Intimidator",
		"Intimidator",
		"Intimidator",
		"Reducer",
		"Reducer",
	},
	"Producer": {
		"Collector",
		"Collector",
		"Collector",
		"Collector",
		"Collector",
		"Collector",
		"Basker",
		"Basker",
		"Basker",
		"Basker",
		"Basker",
		"Basker",
		"Basker",
	},
}

// rollEnvironment rolls a species' evolutionary environment from its homeworld
// profile, whose Atmosphere/Size/Hydrographics modify the locomotion roll.
func rollEnvironment(r *dice.Roller, home uwp.Profile) Environment {
	dm := clamp(r.Flux(), -5, 5)
	loco := rollLocomotion(r, dm, home)
	class, niche := rollNiche(r, dm)
	return Environment{
		Terrain:    terrainNames[dm+5],
		EnvironDM:  dm,
		Locomotion: loco,
		Class:      class,
		Niche:      niche,
	}
}

// rollLocomotion rolls 1D on the terrain row, shifting the column by the
// cumulative homeworld DMs (chart 05B): Atm 8+ -2, Size 5- -1, Hyd 6+ +1, Hyd
// 9+ a further +1.
func rollLocomotion(r *dice.Roller, dm int, home uwp.Profile) Locomotion {
	col := r.Die() - 1
	if home.Atmosphere >= 8 {
		col -= 2
	}
	if home.Size <= 5 {
		col--
	}
	if home.Hydrographics >= 6 {
		col++
	}
	if home.Hydrographics >= 9 {
		col++
	}
	return locomotionGrid[dm+5][clamp(col, 0, 5)]
}

// rollNiche rolls the basic class, then a second Flux (plus the Environ DM) for
// the sub-niche within that class column (chart 05C: the Environ DM modifies the
// sub-niche columns but not the basic class).
func rollNiche(r *dice.Roller, dm int) (class, niche string) {
	class = nicheClasses[dice.FluxIndex(r.Flux())]
	niche = nicheCols[class][dice.FluxIndex(r.Flux()+dm)]
	return class, niche
}
