package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/tradecode"
	"github.com/philoserf/t5/internal/worldgen"
)

func TestApplyHomeworldSkills(t *testing.T) {
	grant := func(codes ...tradecode.Code) skillLevels {
		var c Character
		ApplyHomeworldSkills(&c, worldgen.World{TradeCodes: codes}, DefaultPolicy{})

		return skillLevels{&c}
	}

	// A fixed single-skill code.
	if got := grant("Ag").level("Animals"); got != 1 {
		t.Errorf("Ag -> Animals = %d, want 1", got)
	}
	// Industrial and Rich are player choices; DefaultPolicy takes the first option.
	if got := grant("In").level(theTrades[0]); got != 1 {
		t.Errorf("In -> %s = %d, want 1", theTrades[0], got)
	}

	if got := grant("Ri").level(oneArt[0]); got != 1 {
		t.Errorf("Ri -> %s = %d, want 1", oneArt[0], got)
	}
	// A no-skill code (Barren) grants nothing, and an unknown code is ignored.
	if got := grant("Ba", "ZZ").any(); got {
		t.Error("Ba/unknown codes granted a skill, want none")
	}
	// Two codes mapping to the same skill stack (one grant per classification).
	if got := grant("Fl", "He").level("Hostile Environ"); got != 2 {
		t.Errorf("Fl+He -> Hostile Env = %d, want 2", got)
	}
}

// skillLevels is a tiny read helper over a character's skills.
type skillLevels struct{ c *Character }

func (s skillLevels) level(name string) int { return s.c.Skills.Level(name) }
func (s skillLevels) any() bool             { return s.c.Skills.String() != "" }

// TestHighGSpelledOneWay guards the one skill name the book spells two ways.
// Skill names are map keys, so an Ocean-World native's "Hi-G" and a Rogue's
// "High-G" would never stack — the homeworld level would sit in a bucket no
// career could raise. Every site uses the p.132 master-Skills-table spelling.
func TestHighGSpelledOneWay(t *testing.T) {
	c := Character{}
	ApplyHomeworldSkills(&c, worldgen.World{TradeCodes: []tradecode.Code{"Oc"}}, DefaultPolicy{})

	if c.Skills.Level("High-G") != 1 {
		t.Errorf("Ocean World grants %v, want High-G-1", c.Skills.List())
	}
	// The Rogue, Noble, and Functionary grids award the same skill; a level from
	// the homeworld and a level from a career must land on one entry.
	for _, career := range []Career{RogueCareer, NobleCareer, FunctionaryCareer} {
		found := false

		for _, col := range career.Skills {
			for _, cell := range col {
				if cell.Skill == "High-G" {
					found = true
				}

				if cell.Skill == "Hi-G" {
					t.Errorf("%s grid spells the skill %q; use High-G", career.Name, cell.Skill)
				}
			}
		}

		if !found {
			t.Errorf("%s grid has no High-G cell", career.Name)
		}
	}
}

// TestHomeworldSkillCoversEveryCode asserts that every Chart D code is accounted
// for by exactly one of homeworldSkill (grants a skill), the In/Ri player-choice
// codes (handled in ApplyHomeworldSkills), or homeworldNoSkill (grants none by
// rule). A newly added tradecode.Code can then never silently fall through the
// switch's default with no skill and no record of that being intentional.
func TestHomeworldSkillCoversEveryCode(t *testing.T) {
	playerChoice := map[tradecode.Code]bool{tradecode.In: true, tradecode.Ri: true}

	// In/Ri are handled by ApplyHomeworldSkills's switch, which shadows the map's
	// default lookup — so a homeworldSkill entry for either would be dead code the
	// partition check below could not see. Assert they are absent from the map.
	for c := range playerChoice {
		if _, ok := homeworldSkill[c]; ok {
			t.Errorf("player-choice code %q is also in homeworldSkill (a dead, shadowed entry)", c)
		}
	}

	for _, c := range tradecode.Order {
		_, hasSkill := homeworldSkill[c]
		grants := hasSkill || playerChoice[c]
		noSkill := homeworldNoSkill[c]

		switch {
		case grants && noSkill:
			t.Errorf("code %q is in both the skill and no-skill sets", c)
		case !grants && !noSkill:
			t.Errorf("code %q is in neither set: add it to homeworldSkill or homeworldNoSkill", c)
		}
	}
}
