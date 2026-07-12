package chargen

// A Policy supplies the player choices a career throws up, so that generation
// stays deterministic and testable: the roller provides randomness, the policy
// provides decisions. More choice points (skill column, muster column) are added
// as the slices that need them land.
type Policy interface {
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
}

// DefaultPolicy makes reasonable automatic choices, so a character can be
// generated with no interaction (like the world and system generators).
type DefaultPolicy struct{}

// ChooseCC picks the highest-scoring available characteristic — the best odds on
// the Risk & Reward roll.
func (DefaultPolicy) ChooseCC(c Character, available []Characteristic) Characteristic {
	best := available[0]
	for _, ch := range available[1:] {
		if c.Score(ch) > c.Score(best) {
			best = ch
		}
	}
	return best
}

// RiskMod takes no modifier — the neutral choice.
func (DefaultPolicy) RiskMod(Character, int) int { return 0 }

// ChooseSkillColumn favours the first specialty column (skipping the Personal
// column 0) whose cells all award something, so a default character gains a
// varied spread of skills rather than characteristic bumps or empty cells (for
// the Scout, this skips the mostly-empty Academic column for Courier).
func (DefaultPolicy) ChooseSkillColumn(_ Character, grid SkillGrid) int {
	for col := 1; col < len(grid); col++ {
		if !hasEmptyCell(grid[col]) {
			return col
		}
	}
	return 1
}

// hasEmptyCell reports whether a skill-grid column has any non-awarding cell.
func hasEmptyCell(column [6]Cell) bool {
	for _, cell := range column {
		if cell.Kind == NoAward {
			return true
		}
	}
	return false
}

// ChooseSkill takes the first option.
func (DefaultPolicy) ChooseSkill(_ Character, options []string) string { return options[0] }

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
