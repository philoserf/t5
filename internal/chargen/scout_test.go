package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
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
func (goldenPolicy) RandomizeMusterDM() bool                           { return false }
func (goldenPolicy) PursueEducation(Character) bool                    { return false }
func (goldenPolicy) ChooseTradeSchool(Character) bool                  { return false }
func (goldenPolicy) PursueGraduateSchool(Character) bool               { return false }
func (goldenPolicy) TakeWaiver(Character, int) bool                    { return true }
func (goldenPolicy) NextCareer(Character) (Career, bool)               { return Career{}, false }
func (goldenPolicy) ChooseExplorerDuty(Character) bool                 { return true }

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
		// Muster out: 2 rolls (2 terms), Benefit column. Each Explorer Reward
		// success was a Discovery (Fame +1), so Fame is 2 and DM = +Fame/2 = 1.
		4, // 4 + 1 = row 5 -> Str +1 (7 -> 8)
		1, // 1 + 1 = row 2 -> Forbidden Knowledge
	}

	// A fixed homeworld: Ag -> Animals-1, Va -> Vacc Suit-1 (dice-free grants that
	// do not collide with the Survey/Navigation earned in the career).
	homeworld := worldgen.World{TradeCodes: []string{"Ag", "Va"}}

	c := GenerateCareered(dice.NewScripted(seq...), goldenPolicy{}, homeworld, ScoutCareer)

	if got := c.UPP(); got != "877777" {
		t.Errorf("UPP = %q, want %q (Str 7 +1 muster benefit)", got, "877777")
	}

	if c.Age != 26 {
		t.Errorf("Age = %d, want 26 (18 + 2 terms)", c.Age)
	}

	if c.Skills.Level("Animals") != 1 || c.Skills.Level("Vacc Suit") != 1 {
		t.Errorf("homeworld skills: Animals=%d Vacc Suit=%d, want 1/1",
			c.Skills.Level("Animals"), c.Skills.Level("Vacc Suit"))
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

	if len(c.Benefits) != 1 || c.Benefits[0] != "Forbidden Knowledge" {
		t.Errorf("Benefits = %v, want [Forbidden Knowledge]", c.Benefits)
	}
	// Each term's Explorer Reward succeeded, so each was a Discovery: two
	// Discoveries, two Land Grants, and Fame 2.
	if c.Discoveries != 2 || c.LandGrants != 2 || c.Fame != 2 {
		t.Errorf("Discoveries=%d LandGrants=%d Fame=%d, want 2/2/2",
			c.Discoveries, c.LandGrants, c.Fame)
	}

	if c.WoundBadges != 0 || c.Dead {
		t.Errorf("unexpected injury: badges %d dead %v", c.WoundBadges, c.Dead)
	}
}

// TestScoutBeginNoRetry: a Scout's Begin is a single roll — no Begin retry (Book 1
// p.65: Begin retry is a per-career property, and the Scout's box grants none; its
// "Retry R&R C5" is a term-level R&R retry, not a Begin retry, and is deferred).
// This once wrongly retried a failed Begin against Education.
func TestScoutBeginNoRetry(t *testing.T) {
	c := Character{scores: [count]int{5, 5, 5, 7, 12, 7}} // Str/Dex/End 5, high Edu 12
	// Begin 2D=8 > qualify 5 fails, and there is no retry — high Education does not
	// rescue it, because that was the conflated R&R-retry stat, not a Begin retry.
	if testBegin(dice.NewScripted(6, 2), &c, ScoutCareer) {
		t.Error("a failed Scout Begin has no retry (8 > 5), even with high Education")
	}
	// A passing Begin still enters on the single roll.
	if !testBegin(dice.NewScripted(2, 1), &c, ScoutCareer) { // 2D=3 <= 5
		t.Error("a Scout Begin of 3 <= 5 should enter the career")
	}
}

// TestFailedBeginCostsAYear locks Book 1 p.65: "Each failed attempt (both Begin
// or Retry) takes one year." Only a rolled refusal costs the year — an automatic
// entry never attempts, so it never fails.
func TestFailedBeginCostsAYear(t *testing.T) {
	c := Character{scores: [count]int{5, 5, 5, 7, 7, 7}, Age: startingAge}

	if testBegin(dice.NewScripted(6, 2), &c, ScoutCareer) { // 8 > best(Str/Dex/End) = 5
		t.Fatal("Begin 8 > 5 should be refused")
	}

	if c.Age != startingAge+1 {
		t.Errorf("after a refused Begin Age = %d, want %d", c.Age, startingAge+1)
	}

	if !testBegin(dice.NewScripted(2, 1), &c, ScoutCareer) { // 3 <= 5, admitted
		t.Fatal("Begin 3 <= 5 should be accepted")
	}

	if c.Age != startingAge+1 {
		t.Errorf("a successful Begin aged the character: Age = %d, want %d", c.Age, startingAge+1)
	}
	// The Citizen's Begin is automatic — no attempt, so no year. The scripted face
	// is never drawn (drawing it would fail 6,... against no target at all).
	if !testBegin(dice.NewScripted(6), &c, CitizenCareer) {
		t.Fatal("the Citizen career auto-begins")
	}

	if c.Age != startingAge+1 {
		t.Errorf("an automatic Begin aged the character: Age = %d, want %d", c.Age, startingAge+1)
	}
}

// courierPolicy is goldenPolicy but takes Courier duty (no Risk & Reward).
type courierPolicy struct{ goldenPolicy }

func (courierPolicy) ChooseExplorerDuty(Character) bool { return false }

// TestScoutCourierDuty confirms Courier duty skips Risk & Reward (no injury even
// for a frail Scout) and grants the 4 Courier skills instead of the 8 Explorer.
func TestScoutCourierDuty(t *testing.T) {
	c := Character{scores: [count]int{2, 2, 2, 7, 7, 7}} // frail — Explorer would risk injury
	run := careerRun{ccPool: []Characteristic{Strength}}
	// goldenPolicy picks column 3 (Exploration); die 1 -> Survey.
	got := runTerm(dice.NewScripted(1, 1, 1, 1), courierPolicy{}, &c, &run, ScoutCareer)
	if got != Ongoing || c.WoundBadges != 0 {
		t.Errorf(
			"Courier duty: outcome %v wounds %d, want Ongoing/0 (no Risk & Reward)",
			got,
			c.WoundBadges,
		)
	}

	if c.Skills.Level("Survey") != 4 {
		t.Errorf("Survey = %d, want 4 (Courier grants 4 skills, not 8)", c.Skills.Level("Survey"))
	}
}
