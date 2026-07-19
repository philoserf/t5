package chargen

import "github.com/philoserf/t5/internal/worldgen"

// Homeworld skills (Book 1, "Birthworlds and Homeworlds", p. 56). A character
// receives one skill for each Trade Classification of their homeworld — e.g. a
// character from an Agricultural (Ag) world receives Animals-1. Homeworld skills
// are granted flat at level 1 (the book's "-1" notation), not through the
// cascade progression.

// oneArt and theTrades are the canonical "choose one" lists (Book 1 p. 56),
// shared by the homeworld Rich/Industrial grants and the career skill grids.
var (
	oneArt    = []string{"Actor", "Artist", "Author", "Chef", "Dancer", "Musician"}
	theTrades = []string{
		"Biologics", "Craftsman", "Electronics", "Fluidics", "Gravitics",
		"Magnetics", "Mechanic", "Photonics", "Polymers", "Programmer",
	}
)

// homeworldSkill maps a Trade Classification to the flat skill(s) it grants
// (Book 1 p. 56). Codes that grant no skill are absent; the two "choose one"
// codes, In (One Trade) and Ri (One Art), are handled in ApplyHomeworldSkills.
// The full table is transcribed (not only the UWP-determinable codes worldgen
// currently emits) so a selected or charted homeworld resolves too.
var homeworldSkill = map[string][]string{
	"Ag": {"Animals"},             // Agricultural
	"As": {"Zero-G"},              // Asteroid
	"Co": {"Hostile Environ"},     // Cold
	"Cp": {"Admin"},               // Subsector Capital
	"Cs": {"Bureaucrat"},          // Sector Capital
	"Cx": {"Language"},            // Capital
	"Da": {"Fighter"},             // Dangerous
	"De": {"Survival"},            // Desert
	"Ds": {"Vacc Suit", "Zero-G"}, // Deep Space
	"Fa": {"Animals"},             // Farming
	"Fl": {"Hostile Environ"},     // Fluid
	"Fr": {"Hostile Environ"},     // Frozen
	"Ga": {"Trader"},              // Garden World
	"He": {"Hostile Environ"},     // Hellworld
	"Hi": {"Streetwise"},          // High Population
	"Ho": {"Hostile Environ"},     // Hot
	"Ic": {"Vacc Suit"},           // Ice-Capped
	"Lo": {"Flyer"},               // Low Population
	"Mi": {"Survey"},              // Mining
	"Na": {"Survey"},              // Non-agricultural
	"Ni": {"Driver"},              // Non-industrial
	// "Oc" reads "Hi-G" on p.56, but the skill is spelled "High-G" on the p.132
	// master Skills table, in the index, and on every career grid that awards it
	// (Rogue p.84, Noble p.85, Functionary p.87) — the p.154 definition's own
	// headword lists "High-Gravity" first among its alternates. Skill names are
	// keys here: two spellings never stack, so an Ocean-World native's level would
	// sit in a bucket no career could ever raise. Normalized to the majority form.
	"Oc": {"High-G"},    // Ocean World
	"Pa": {"Trader"},    // Pre-Agricultural
	"Pi": {"JOT"},       // Pre-Industrial
	"Po": {"Steward"},   // Poor
	"Pr": {"Craftsman"}, // Pre-Rich
	"Tr": {"Survival"},  // Tropic
	"Tu": {"Survival"},  // Tundra
	"Tz": {"Driver"},    // Twilight Zone
	"Va": {"Vacc Suit"}, // Vacuum
	"Wa": {"Seafarer"},  // Water World
}

// ApplyHomeworldSkills grants a character their homeworld skills: one per Trade
// Classification of the world. Industrial (In) and Rich (Ri) worlds grant a
// player-chosen One Trade or One Art; the rest are fixed. No dice are rolled —
// choices come from the policy — so a homeworld composes into generation without
// disturbing any dice sequence.
func ApplyHomeworldSkills(c *Character, world worldgen.World, p Policy) {
	for _, code := range world.TradeCodes {
		switch code {
		case "In":
			c.Skills.Raise(p.ChooseSkill(*c, theTrades), 1)
		case "Ri":
			c.Skills.Raise(p.ChooseSkill(*c, oneArt), 1)
		default:
			for _, s := range homeworldSkill[code] {
				c.Skills.Raise(s, 1)
			}
		}
	}
}
