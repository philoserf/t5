package chargen

import (
	"github.com/philoserf/t5/internal/tradecode"
	"github.com/philoserf/t5/internal/worldgen"
)

// Homeworld skills (Book 1, "Birthworlds and Homeworlds", p.56). A character
// receives one skill for each Trade Classification of their homeworld — e.g. a
// character from an Agricultural (Ag) world receives Animals-1. Homeworld skills
// are granted flat at level 1 (the book's "-1" notation), not through the
// cascade progression.

// oneArt and theTrades are the canonical "choose one" lists (Book 1 p.56),
// shared by the homeworld Rich/Industrial grants and the career skill grids.
var (
	oneArt    = []string{"Actor", "Artist", "Author", "Chef", "Dancer", "Musician"}
	theTrades = []string{
		"Biologics", "Craftsman", "Electronics", "Fluidics", "Gravitics",
		"Magnetics", "Mechanic", "Photonics", "Polymers", "Programmer",
	}
)

// homeworldSkill maps a Trade Classification to the flat skill(s) it grants
// (Book 1 p.56). Codes that grant no skill are absent; the two "choose one"
// codes, In (One Trade) and Ri (One Art), are handled in ApplyHomeworldSkills.
// The full table is transcribed (not only the UWP-determinable codes worldgen
// currently emits) so a selected or charted homeworld resolves too.
var homeworldSkill = map[tradecode.Code][]string{
	tradecode.Ag: {"Animals"},         // Agricultural
	tradecode.As: {"Zero-G"},          // Asteroid
	tradecode.Co: {"Hostile Environ"}, // Cold
	tradecode.Cp: {"Admin"},           // Subsector Capital
	tradecode.Cs: {"Bureaucrat"},      // Sector Capital
	tradecode.Cx: {"Language"},        // Capital
	tradecode.Da: {"Fighter"},         // Dangerous
	tradecode.De: {"Survival"},        // Desert
	tradecode.Fa: {"Animals"},         // Farming
	tradecode.Fl: {"Hostile Environ"}, // Fluid
	tradecode.Fr: {"Hostile Environ"}, // Frozen
	tradecode.Ga: {"Trader"},          // Garden World
	tradecode.He: {"Hostile Environ"}, // Hellworld
	tradecode.Hi: {"Streetwise"},      // High Population
	tradecode.Ho: {"Hostile Environ"}, // Hot
	tradecode.Ic: {"Vacc Suit"},       // Ice-Capped
	tradecode.Lo: {"Flyer"},           // Low Population
	tradecode.Mi: {"Survey"},          // Mining
	tradecode.Na: {"Survey"},          // Non-agricultural
	tradecode.Ni: {"Driver"},          // Non-industrial
	// Oc reads "Hi-G" on p.56, but the skill is spelled "High-G" on the p.132
	// master Skills table, in the index, and on every career grid that awards it
	// (Rogue p.84, Noble p.85, Functionary p.87) — the p.154 definition's own
	// headword lists "High-Gravity" first among its alternates. Skill names are
	// keys here: two spellings never stack, so an Ocean-World native's level would
	// sit in a bucket no career could ever raise. Normalized to the majority form.
	tradecode.Oc: {"High-G"},    // Ocean World
	tradecode.Pa: {"Trader"},    // Pre-Agricultural
	tradecode.Pi: {"JOT"},       // Pre-Industrial
	tradecode.Po: {"Steward"},   // Poor
	tradecode.Pr: {"Craftsman"}, // Pre-Rich
	tradecode.Tr: {"Survival"},  // Tropic
	tradecode.Tu: {"Survival"},  // Tundra
	tradecode.Tz: {"Driver"},    // Twilight Zone
	tradecode.Va: {"Vacc Suit"}, // Vacuum
	tradecode.Wa: {"Seafarer"},  // Water World
}

// homeworldNoSkill is every Chart D code that intentionally grants no homeworld
// skill (Book 1 p.56 lists none for them, and In/Ri are handled in
// ApplyHomeworldSkills rather than here). It exists so TestHomeworldSkillCoversEveryCode
// can assert that every tradecode.Code is accounted for by exactly one of the two
// sets — a newly added code can never silently fall through the switch's default.
var homeworldNoSkill = map[tradecode.Code]bool{
	tradecode.Ba: true, // Barren
	tradecode.Di: true, // Dieback
	tradecode.Ph: true, // Pre-High Population
	tradecode.Px: true, // Prison/Exile Camp
	tradecode.Pe: true, // Penal Colony
	tradecode.Re: true, // Reserve
	tradecode.Sa: true, // Satellite
	tradecode.Lk: true, // Locked
	tradecode.Mr: true, // Military Rule
	tradecode.Cy: true, // Colony
	tradecode.Fo: true, // Forbidden
	tradecode.Pz: true, // Puzzle
	tradecode.Ab: true, // Data repository
	tradecode.An: true, // Ancient site
}

// ApplyHomeworldSkills grants a character their homeworld skills: one per Trade
// Classification of the world. Industrial (In) and Rich (Ri) worlds grant a
// player-chosen One Trade or One Art; the rest are fixed. No dice are rolled —
// choices come from the policy — so a homeworld composes into generation without
// disturbing any dice sequence.
func ApplyHomeworldSkills(c *Character, world worldgen.World, p Policy) {
	for _, code := range world.TradeCodes {
		switch code {
		case tradecode.In:
			c.Skills.Raise(p.ChooseSkill(*c, theTrades), 1)
		case tradecode.Ri:
			c.Skills.Raise(p.ChooseSkill(*c, oneArt), 1)
		default:
			for _, s := range homeworldSkill[code] {
				c.Skills.Raise(s, 1)
			}
		}
	}
}
