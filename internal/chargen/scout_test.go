package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

// goldenPolicy makes fully deterministic choices so a scripted Scout can be
// traced by hand: rotate the CC in order, take no Risk modifier, always draw
// skills from the Exploration column (all plain skills), serve exactly two
// terms, and take the Benefit column when mustering out.
type goldenPolicy struct{}

func (goldenPolicy) ChooseCC(_ Character, available []Characteristic) Characteristic {
	return available[0]
}
func (goldenPolicy) RiskMod(Character, int) int                        { return 0 }
func (goldenPolicy) ChooseSkillColumn(Character, SkillGrid) int        { return 3 } // Exploration
func (goldenPolicy) ChooseSkill(_ Character, options []string) string  { return options[0] }
func (goldenPolicy) Continue(_ Character, rec CareerRecord) bool       { return rec.Terms < 2 }
func (goldenPolicy) MusterColumn(Character, CareerRecord) MusterColumn { return BenefitColumn }

// TestGoldenScout traces a complete two-term Scout end-to-end from a scripted
// roller and policy, locking the Scout transcription and the whole career
// lifecycle. Every roll below is 3,4 (= 7) unless noted.
func TestGoldenScout(t *testing.T) {
	seq := []int{
		// UPP: six 2D, each 3+4 = 7, so "777777"; qualify target = best(Str/Dex/End) = 7.
		3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4,
		3, 4, // qualify: 7 <= 7, enters the career
		// Term 1: CC = Str (7). Risk 7 <= 7 survive; Reward 7.
		3, 4, // risk
		3, 4, // reward
		// 8 skill rolls from Exploration: rows 1,1,1,1 (Survey) then 6,6,6,6 (Navigation).
		1, 1, 1, 1, 6, 6, 6, 6,
		3, 4, // continue: 7 <= Int 7, policy wants term 2
		// Term 2: CC = Dex (7). Risk 7 survive; Reward 7.
		3, 4, // risk
		3, 4, // reward
		1, 1, 1, 1, 6, 6, 6, 6, // Survey x4, Navigation x4 again
		3, 4, // continue: policy stops after term 2 (muster out)
		// Muster out: 2 rolls (2 terms), Benefit column, DM 0.
		5, // row 5 -> Str +1 (7 -> 8)
		1, // row 1 -> Ship Share
	}

	c := GenerateCareered(dice.NewScripted(seq...), goldenPolicy{}, ScoutCareer)

	if got := c.UPP(); got != "877777" {
		t.Errorf("UPP = %q, want %q (Str 7 +1 muster benefit)", got, "877777")
	}
	if c.Age != 26 {
		t.Errorf("Age = %d, want 26 (18 + 2 terms)", c.Age)
	}
	if got := c.Skills.Level("Survey"); got != 8 {
		t.Errorf("Survey = %d, want 8", got)
	}
	if got := c.Skills.Level("Navigation"); got != 8 {
		t.Errorf("Navigation = %d, want 8", got)
	}
	if len(c.Careers) != 1 {
		t.Fatalf("careers = %+v, want one record", c.Careers)
	}
	rec := c.Careers[0]
	if rec.Career != Scout || rec.Terms != 2 || rec.Outcome != MusteredOut {
		t.Errorf("record = %+v, want Scout/2 terms/MusteredOut", rec)
	}
	if c.Credits != 0 {
		t.Errorf("Credits = %d, want 0 (both muster rolls took benefits)", c.Credits)
	}
	if len(c.Benefits) != 1 || c.Benefits[0] != "Ship Share" {
		t.Errorf("Benefits = %v, want [Ship Share]", c.Benefits)
	}
	if c.WoundBadges != 0 || c.Dead {
		t.Errorf("unexpected injury: badges %d dead %v", c.WoundBadges, c.Dead)
	}
}
