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

// ChooseSkillColumn favours a specialty column (1) over the Personal column (0),
// so a default character actually gains skills rather than only characteristic
// bumps.
func (DefaultPolicy) ChooseSkillColumn(Character, SkillGrid) int { return 1 }

// ChooseSkill takes the first option.
func (DefaultPolicy) ChooseSkill(_ Character, options []string) string { return options[0] }

// Continue keeps serving until aging begins (age 34), a simple heuristic that
// yields a typical few-term character.
func (DefaultPolicy) Continue(c Character, _ CareerRecord) bool {
	return c.Age < physicalAgingAge
}
