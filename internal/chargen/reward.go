package chargen

import (
	"fmt"

	"github.com/philoserf/t5/internal/dice"
)

// Mid-career awards for the term engine: a successful Reward roll's token
// (Book 1 p.65), the Imperial Medals table (p.70), and the skill-grid the term
// draws its per-term skills from. Split from career.go (#331); all one package.

// A RewardKind is the token a successful Reward roll earns for a career.
type RewardKind int

// Reward tokens a career grants on a successful Reward roll.
const (
	RewardNone         RewardKind = iota // the reward benefit is deferred (Rogue via its Scheme, …)
	RewardMedal                          // armed forces: a Medal
	RewardPublication                    // the Scholar: a Publication
	RewardShipShares                     // the Merchant: escalating Ship Shares (Nth reward = N shares)
	RewardDiscovery                      // the Scout: a Discovery — a Land Grant and Fame +1 (Book 1 p.79)
	RewardCommendation                   // the Agent: an official Commendation (Book 1 p.83)
)

// A Medal is one award from the Imperial Medals table (Book 1 p.70). Mod is the
// bonus it adds to a Soldier/Spacer/Marine promotion target.
type Medal struct {
	Code string
	Name string
	Mod  int
}

// medalsTable is the p.70 Medals table, indexed by the line the book names: the
// *unmodified* successful Reward roll, +1 if the character is an Officer. 2D
// gives 2..12, so the officer bump makes 13 the top line; index 0 and 1 are unused.
//
// The Eneri Dinsha worked example (p.72) locks both ends of the lookup: a raw
// Reward roll of 3 with the Officer +1 reaches "Medals table line 4" for an XS,
// and a raw 9 with the same +1 reaches "line 10" for an MCUF.
var medalsTable = [14]Medal{
	2:  {"XS", "Exemplary Service", 1},
	3:  {"XS", "Exemplary Service", 1},
	4:  {"XS", "Exemplary Service", 1},
	5:  {"XS", "Exemplary Service", 1},
	6:  {"XS", "Exemplary Service", 1},
	7:  {"XS", "Exemplary Service", 1},
	8:  {"XS", "Exemplary Service", 1},
	9:  {"MCUF", "Meritorious Conduct Under Fire", 2},
	10: {"MCUF", "Meritorious Conduct Under Fire", 2},
	11: {"MCG", "Medal for Conspicuous Gallantry", 3},
	12: {"SEH", "Starburst for Extreme Heroism", 4},
	13: {"*SEH*", "SEH With Diamonds", 5},
}

// exemplaryService is the XS badge a held Risk earns ("Risk Success: Receive XS
// Exemplary Service Badge", Book 1 pp.82/86 and the Spacer's p.81). It is the
// table's own bottom entry, not a separate award.
var exemplaryService = medalsTable[2]

// awardMedal records one medal. The award itself is kept, not a tally of it: a
// character sheet reading "MCUF, SEH" cannot be recovered from a count and a mod
// sum, and #192's MCG/SEH muster-out rolls need to know which medal was earned.
// The count and the mod sum are both derived (MedalCount, MedalMods), so they
// cannot drift from each other or from the awards.
func awardMedal(c *Character, m Medal) {
	c.Medals = append(c.Medals, m)
}

// MedalCount is how many medals the character has been awarded.
func (c Character) MedalCount() int { return len(c.Medals) }

// MedalMods is the summed p.70 table mod of the character's medals — the bonus
// they contribute to a Soldier/Spacer/Marine promotion target. Two medals are not
// worth +2 unless both are XS: one MCUF alone is +2.
func (c Character) MedalMods() int {
	total := 0
	for _, m := range c.Medals {
		total += m.Mod
	}

	return total
}

// medalFor reads the award off the p.70 table for a successful Reward roll.
// The roll is the raw 2D, before any Risk & Reward mods: "Rew= Successful
// unmodified Reward Roll. If Officer, increase +1".
func medalFor(rawRoll int, officer bool) Medal {
	line := rawRoll
	if officer {
		line++
	}

	// 2D is 2..12 and the Officer bump tops at 13, so every reachable line is a real
	// row. Rather than clamp an unreachable value into one, fail on it: lines 0 and 1
	// are inside the array and would hand back a nameless zero-value Medal, which
	// MedalCount would count and no sheet could render.
	if line < 2 || line >= len(medalsTable) {
		panic(fmt.Sprintf("chargen: medal line %d out of range 2-%d", line, len(medalsTable)-1))
	}

	return medalsTable[line]
}

// grantReward awards the career's reward token for a successful Reward roll
// (Book 1 p.65). ccVal is the term's original Controlling Characteristic value,
// which the Scholar's Award-Winning threshold is measured against.
//
// The Reward is rolled every term, whether the Risk was held or lost (p.65: "The
// Character rolls for Risk ... and determines the outcome. He then rolls again for
// Reward ... and determines the consequences"). The Eneri Dinsha worked example
// (p.72) fails Risk in both of his terms — taking a Wound Badge in the first — and
// still rolls Reward and takes a Medal each time. The target keeps the ORIGINAL
// Controlling Characteristic: his second-term Reward is "10 +2 +1 -2" against his
// pre-injury Dexterity-10.
//
// The roll happens either way, so the dice stream does not depend on the Risk
// outcome — but a character the Risk roll killed collects nothing: the book has him
// "determine the consequences", and a corpse has none. Without that guard a Merchant
// killed in his term still banks Ship Shares and a Scout still records a Discovery,
// both of which feed muster-out and the character sheet.
func grantReward(
	c *Character,
	run *careerRun,
	career Career,
	reward dice.CheckResult,
	ccVal int,
) {
	switch career.RewardKind {
	case RewardNone:
		// the reward is deferred (e.g. the Rogue's Scheme); nothing to grant
	case RewardMedal:
		awardMedal(c, medalFor(reward.Roll, run.officer))
	case RewardPublication:
		c.Publications++
		if reward.Roll <= ccVal-4 {
			c.Publications++ // Award-Winning (Book 1 p.76): a Publication 4 under the CC counts as two
		}
	case RewardShipShares:
		run.rewards++ // the Nth Reward success is worth N Ship Shares
		c.ShipShares += run.rewards
	case RewardDiscovery:
		c.Discoveries++ // a valuable new world or feature: a Land Grant and Fame +1
		c.LandGrants++
		c.Fame++
	case RewardCommendation:
		c.Commendations++
		run.commends++
	}
}

// awardSkills grants the term's skill eligibility (Book 1 p.65).
func awardSkills(r *dice.Roller, p Policy, c *Character, career Career) {
	awardSkillsN(r, p, c, career, career.EligPerTerm)
}

// awardSkillsN grants n skill rolls: for each the policy picks a column of the
// career's skill grid and 1D selects the row.
func awardSkillsN(r *dice.Roller, p Policy, c *Character, career Career, n int) {
	for range n {
		col := p.ChooseSkillColumn(*c, career.Skills)
		if col < 0 || col >= len(career.Skills) {
			panic(
				fmt.Sprintf(
					"chargen: skill column %d out of range 0-%d",
					col,
					len(career.Skills)-1,
				),
			)
		}

		applyCell(p, c, career.Skills[col][r.Die()-1])
	}
}

// applyCell applies one skill-grid cell: raise a skill (cascade skills grant a
// knowledge via the K-K-S progression), bump a characteristic (capped at the
// human maximum), or resolve a player choice among options.
// cascadeParent returns the cascade parent this cell awards a Knowledge under,
// or "" when the cell is not a cascade cell. Two shapes qualify: an AwardSkill
// carrying a Knowledge, and an AwardChoice whose Options are knowledges under a
// named parent (what the cascade() grid helper builds).
//
// It lives here, beside applyCell, because applyCell's routing and the career
// grids' cascade-cell test must agree about what a cascade cell IS. Encoding
// that twice — once in the engine, once in a test — means a third shape added
// here would silently stop being covered there.
func (cell Cell) cascadeParent() string {
	switch cell.Kind {
	case AwardSkill:
		if cell.Knowledge != "" {
			return cell.Skill
		}
	case AwardChoice:
		return cell.Skill // "" for a plain choice among flat skills
	case NoAward, AwardBump, AwardMajor, AwardMinor:
	}

	return ""
}

func applyCell(p Policy, c *Character, cell Cell) {
	switch cell.Kind {
	case NoAward:
		// an empty cell: nothing to apply
	case AwardSkill:
		if parent := cell.cascadeParent(); parent != "" {
			c.Skills.GrantCascade(parent, cell.Knowledge)
		} else {
			c.Skills.Raise(cell.Skill, 1)
		}
	case AwardBump:
		c.scores[cell.Char] = min(c.scores[cell.Char]+1, maxCharacteristic)
	case AwardChoice:
		if len(cell.Options) == 0 {
			panic("chargen: AwardChoice cell has no options")
		}

		chosen := p.ChooseSkill(*c, cell.Options)
		if parent := cell.cascadeParent(); parent != "" {
			// A cascade choice: the options are knowledges under the parent skill
			// (e.g. Language/Galanglic), granted via the K-K-S progression.
			c.Skills.GrantCascade(parent, chosen)
		} else {
			c.Skills.Raise(chosen, 1)
		}
	case AwardMajor:
		if c.Major != "" { // lost if the character has no Major (never went to college)
			c.Skills.Raise(c.Major, 1)
		}
	case AwardMinor:
		if c.Minor != "" {
			c.Skills.Raise(c.Minor, 1)
		}
	}
}

// maxCharacteristic is the human cap on a characteristic raised in play (eHex F).
const maxCharacteristic = 15

// A CellKind identifies what a skill-grid cell awards.
type CellKind int

// Skill-grid cell kinds.
const (
	NoAward     CellKind = iota // an empty cell
	AwardSkill                  // raise Skill (with Knowledge for a cascade skill)
	AwardBump                   // raise the characteristic Char
	AwardChoice                 // grant one skill the policy picks from Options
	AwardMajor                  // raise the character's College Major (lost if none)
	AwardMinor                  // raise the character's College Minor (lost if none)
)

// A Cell is one entry in a career's skill grid.
type Cell struct {
	Kind CellKind
	// Skill names the skill for AwardSkill; for AwardChoice it optionally names a
	// cascade parent, in which case Options are knowledges granted under it.
	Skill     string
	Knowledge string         // AwardSkill: the knowledge, for a cascade skill
	Char      Characteristic // AwardBump
	Options   []string       // AwardChoice: the skills (or knowledges) to pick among
}

// A SkillGrid is a career's skill table: seven columns of six rows. The column
// is chosen (see Policy.ChooseSkillColumn); the row is a 1D roll.
type SkillGrid [7][6]Cell
