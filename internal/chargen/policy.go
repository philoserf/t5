package chargen

// A Policy supplies the player choices a career throws up, so that generation
// stays deterministic and testable: the roller provides randomness, the policy
// provides decisions. More choice points (skill column, muster column) are added
// as the slices that need them land.
type Policy interface { //nolint:interfacebloat // intentionally aggregates every player choice as one seam
	// ChooseCC picks the term's Controlling Characteristic from those available.
	// The engine only ever passes a non-empty slice (RunCareer rejects a career
	// with no controlling characteristics).
	ChooseCC(c Character, available []Characteristic) Characteristic
	// RiskMod returns the term's Risk & Reward modifier for a Controlling
	// Characteristic of value cc: positive for Caution (safer, worse reward),
	// negative for Bravery (riskier, better reward), 0 for no modifier.
	RiskMod(c Character, cc int) int
	// ChooseSkillColumn picks a column (0-6) of a career's skill grid for one
	// skill award.
	ChooseSkillColumn(c Character, grid SkillGrid) int
	// ChooseSkill picks one skill from a choice cell's options.
	ChooseSkill(c Character, options []string) string
	// Continue reports whether the character wishes to serve another term.
	Continue(c Character, rec CareerRecord) bool
	// MusterColumn picks the Money or Benefit column for one muster-out roll.
	MusterColumn(c Character, rec CareerRecord) MusterColumn
	// RandomizeMusterDM reports whether the optional muster-out DM (Book 1 p.68:
	// "any value from 0 to the total") is randomized per roll. False applies the
	// full DM — the simplest deterministic choice, used by the tests.
	RandomizeMusterDM() bool
	// PursueEducation reports whether the character attends an educational
	// institution before their career (Book 1 stage C).
	PursueEducation(c Character) bool
	// ChooseTradeSchool reports whether an educated character attends a vocational
	// Trade School (Major +2) rather than the academic College/University path
	// (Book 1 p. 60).
	ChooseTradeSchool(c Character) bool
	// PursueGraduateSchool reports whether a degree-holding character goes on to a
	// post-graduate program (the Masters) after their Bachelor's (Book 1 p. 60).
	PursueGraduateSchool(c Character) bool
	// TakeWaiver reports whether the character attempts an Educational Waiver
	// after an adverse roll, given the number of waivers already attempted.
	TakeWaiver(c Character, priorWaivers int) bool
	// NextCareer chooses another career to attempt after the current one, or
	// returns false to finish character generation (Book 1: characters may serve
	// more than one career).
	NextCareer(c Character) (Career, bool)
	// ChooseExplorerDuty reports whether a Scout takes Explorer duty this term
	// (Risk & Reward, more skills) rather than the safer Courier duty (Book 1 p. 79).
	ChooseExplorerDuty(c Character) bool
}

// DefaultPolicy makes reasonable automatic choices, so a character can be
// generated with no interaction (like the world and system generators).
type DefaultPolicy struct{}

// ChooseCC picks the highest-scoring available characteristic — the best odds on
// the Risk & Reward roll.
func (DefaultPolicy) ChooseCC(c Character, available []Characteristic) Characteristic {
	return bestChar(c, available...)
}

// RiskMod takes no modifier — the neutral choice.
func (DefaultPolicy) RiskMod(Character, int) int { return 0 }

// ChooseSkillColumn spreads a character's training: among the columns (skipping
// the Personal column 0) whose cells all award this character something, it
// picks the one currently least developed, so eligibility rotates across
// specialties rather than piling onto a single column and driving one skill
// toward the cap. The productivity check is character-aware: a College graduate
// finds the Academic column (Major/Minor) productive, while an uneducated Scout —
// for whom those cells are lost — does not.
func (DefaultPolicy) ChooseSkillColumn(c Character, grid SkillGrid) int {
	best, bestLevel := -1, 0
	for col := 1; col < len(grid); col++ {
		if productive, level := columnScore(
			c,
			grid[col],
		); productive &&
			(best < 0 || level < bestLevel) {
			best, bestLevel = col, level
		}
	}

	if best < 0 {
		return 1
	}

	return best
}

// columnScore reports, in one pass, whether every cell in a column awards this
// character something and the sum of the current levels of the skills it would
// raise — a proxy for how developed the column already is, used to spread
// training toward the least-developed productive column.
func columnScore(c Character, column [6]Cell) (bool, int) {
	productive := true
	level := 0

	for _, cell := range column {
		if !cellAwards(c, cell) {
			productive = false
		}

		level += cellLevel(c, cell)
	}

	return productive, level
}

// cellLevel is the character's current level in the skill a cell would raise (a
// choice cell counts the least-held option, matching ChooseSkill; a Major/Minor
// cell the declared subject).
func cellLevel(c Character, cell Cell) int {
	switch cell.Kind {
	case AwardSkill:
		return c.Skills.Level(cell.Skill)
	case AwardChoice:
		lowest := 0
		for i, o := range cell.Options {
			if level := c.Skills.Level(o); i == 0 || level < lowest {
				lowest = level
			}
		}

		return lowest
	case AwardMajor:
		return c.Skills.Level(c.Major)
	case AwardMinor:
		return c.Skills.Level(c.Minor)
	default:
		return 0
	}
}

// cellAwards reports whether a cell grants this character something: an empty
// cell never does, and a Major/Minor cell only does once college has declared
// one.
func cellAwards(c Character, cell Cell) bool {
	switch cell.Kind {
	case NoAward:
		return false
	case AwardMajor:
		return c.Major != ""
	case AwardMinor:
		return c.Minor != ""
	default:
		return true
	}
}

// ChooseSkill takes the offered skill the character has least of, so repeated
// choices — "One Trade", "One Art", a Starship Skill — spread across the options
// instead of piling onto the first one.
func (DefaultPolicy) ChooseSkill(c Character, options []string) string {
	best := options[0]
	for _, o := range options[1:] {
		if c.Skills.Level(o) < c.Skills.Level(best) {
			best = o
		}
	}

	return best
}

// Continue keeps serving until aging begins (age 34), a simple heuristic that
// yields a typical few-term character.
func (DefaultPolicy) Continue(c Character, _ CareerRecord) bool {
	return c.Age < physicalAgingAge
}

// MusterColumn takes Benefits — the non-cash awards (characteristic bumps, Ship
// Shares, memberships) generally outweigh a single money roll.
func (DefaultPolicy) MusterColumn(Character, CareerRecord) MusterColumn {
	return BenefitColumn
}

// RandomizeMusterDM randomizes the optional muster-out DM (Book 1 p.68) rather
// than always taking the full value, since a generated character has no reason
// to steer the roll toward a preferred row.
func (DefaultPolicy) RandomizeMusterDM() bool { return true }

// PursueEducation sends a character into the education stage (College or, at Edu
// 7+, University) when they meet the College prerequisite — education raises Edu
// and grants a Major and Minor worth having. A low-Edu character is left to enter
// a career directly rather than sit ED5 first.
func (DefaultPolicy) PursueEducation(c Character) bool {
	return c.Score(Education) >= college.preReqEdu
}

// TakeWaiver always attempts a waiver — staying enrolled beats washing out.
func (DefaultPolicy) TakeWaiver(Character, int) bool { return true }

// ChooseTradeSchool takes the academic path — a College/University degree raises
// Edu and grants a Major and Minor, generally worth more than a Trade School's
// single vocational Major.
func (DefaultPolicy) ChooseTradeSchool(Character) bool { return false }

// PursueGraduateSchool stops at the Bachelor's — a character enters their career
// rather than spending more years in a post-graduate program.
func (DefaultPolicy) PursueGraduateSchool(Character) bool { return false }

// NextCareer stops after one career; a policy that wants a multi-career
// character (e.g. the CLI's -careers sequence) overrides this.
func (DefaultPolicy) NextCareer(Character) (Career, bool) { return Career{}, false }

// ChooseExplorerDuty takes the iconic Explorer duty — the risk buys more skills.
func (DefaultPolicy) ChooseExplorerDuty(Character) bool { return true }
