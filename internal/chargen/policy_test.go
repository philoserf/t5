package chargen

import "testing"

// TestChooseSkillColumnSpreads confirms the default policy trains the
// least-developed column, so eligibility rotates rather than piling onto one.
func TestChooseSkillColumnSpreads(t *testing.T) {
	var grid SkillGrid
	for i := range grid[1] {
		grid[1][i] = sk("A") // column 1 raises A
		grid[2][i] = sk("B") // column 2 raises B
	}
	c := Character{}
	c.Skills.Raise("A", 5) // column 1 is already developed

	if col := (DefaultPolicy{}).ChooseSkillColumn(c, grid); col != 2 {
		t.Errorf("ChooseSkillColumn = %d, want 2 (the less-developed column)", col)
	}
}

// TestChooseSkillLeastHeld confirms repeated choices spread across the options
// rather than piling onto the first.
func TestChooseSkillLeastHeld(t *testing.T) {
	c := Character{}
	c.Skills.Raise("Biologics", 3)

	if got := (DefaultPolicy{}).ChooseSkill(
		c,
		[]string{"Biologics", "Craftsman"},
	); got != "Craftsman" {
		t.Errorf("ChooseSkill = %q, want Craftsman (the least-held option)", got)
	}
	// With nothing held yet, the first option is fine.
	if got := (DefaultPolicy{}).ChooseSkill(
		Character{},
		[]string{"Biologics", "Craftsman"},
	); got != "Biologics" {
		t.Errorf("ChooseSkill = %q, want Biologics (all options unheld)", got)
	}
}
