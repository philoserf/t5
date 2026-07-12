package chargen

// A Policy supplies the player choices a career throws up, so that generation
// stays deterministic and testable: the roller provides randomness, the policy
// provides decisions. More choice points (skill column, muster column) are added
// as the slices that need them land.
type Policy interface {
	// ChooseCC picks the term's Controlling Characteristic from those available.
	ChooseCC(c Character, available []Characteristic) Characteristic
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

// Continue keeps serving until aging begins (age 34), a simple heuristic that
// yields a typical few-term character.
func (DefaultPolicy) Continue(c Character, _ CareerRecord) bool {
	return c.Age < physicalAgingAge
}
