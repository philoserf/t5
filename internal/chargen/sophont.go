package chargen

// The sophont bridge: create an individual character from a Sophont Creation
// species template (internal/sophont, Book 3 pp.228-230). Where the human
// Generate rolls each characteristic as a flat 2D, a sophont rolls each per its
// species' die count, then takes on an assigned gender's (and caste's)
// characteristic differences.
//
// Deferred: a sophont serving careers (GenerateCareered still rolls a human UPP),
// and aging keyed to the species' own life stages rather than the human ladder —
// both need deeper chargen integration and have no consumer yet.

import (
	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/sophont"
)

// GenerateSophont creates an individual of a non-human species at the starting
// age: it rolls the six characteristics per the species' die counts (with the
// Rolling-Higher scaling), assigns a gender and — if the species has castes — a
// caste, applies their characteristic differences, and adopts the species'
// homeworld.
//
// A Human-equivalent species (all-2D, single gender, no caste) reproduces the
// human Generate: the six characteristic rolls come first, in the same order, so
// a shared dice stream yields the same UPP. A sophont characteristic may exceed
// the eHex range, so render it through cmd/sophont's profile rather than UPP.
func GenerateSophont(r *dice.Roller, species sophont.Species) Character {
	c := Character{Age: startingAge, Homeworld: species.Homeworld}
	for i := range c.scores {
		c.scores[i] = sophont.RollValue(r, species.Chars[i].Dice)
	}

	c.Gender = rollTrait(r, &c, species.Gender.Table, species.Gender.Differences)
	if species.Caste != nil {
		c.Caste = rollTrait(r, &c, species.Caste.Table, species.Caste.Differences)
	}

	return c
}

// rollTrait assigns a gender or caste: it rolls 2D on the species' determination
// table and applies the resulting trait's characteristic differences. The base
// gender and the Common caste carry no difference (they are absent from the map);
// a Skilled caste has an empty table and assigns nothing.
func rollTrait(
	r *dice.Roller,
	c *Character,
	table [13]string,
	diffs map[string]sophont.Difference,
) string {
	name := table[r.Dice(2)]
	if diff, ok := diffs[name]; ok {
		applyDifference(r, c, diff)
	}

	return name
}

// applyDifference imposes a caste- or gender-based Difference on C1-C5: extra
// Strength dice, then the flat modifiers, floored at 0. There is no upper cap:
// maxCharacteristic is the human in-play limit, but a sophont characteristic can
// legitimately run far higher (an 8D Strength reaches 48). C6 is never adjusted.
func applyDifference(r *dice.Roller, c *Character, d sophont.Difference) {
	if d.C1Dice > 0 {
		c.scores[Strength] += r.Dice(d.C1Dice)
	}

	for i := range 5 {
		c.scores[i] = max(c.scores[i]+d.Mods[i], 0)
	}
}
