package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// testCareer is a minimal Scout-shaped career for exercising the engine.
var testCareer = Career{
	ID:               Scout,
	Name:             "Scout",
	Qualify:          Qualification{Chars: []Characteristic{Intelligence}},
	CCMode:           RotateCC,
	ControllingChars: []Characteristic{Strength, Dexterity, Endurance, Intelligence},
	Continue:         ContinueRule{UseChar: true, Char: Intelligence},
}

// stopAfter is a policy that continues until a fixed number of terms is served.
type stopAfter struct{ terms int }

func (stopAfter) ChooseCC(_ Character, available []Characteristic) Characteristic {
	return available[0]
}
func (stopAfter) RiskMod(Character, int) int                        { return 0 }
func (stopAfter) ChooseSkillColumn(Character, SkillGrid) int        { return 0 }
func (stopAfter) ChooseSkill(_ Character, options []string) string  { return options[0] }
func (s stopAfter) Continue(_ Character, rec CareerRecord) bool     { return rec.Terms < s.terms }
func (stopAfter) MusterColumn(Character, CareerRecord) MusterColumn { return BenefitColumn }
func (stopAfter) RandomizeMusterDM() bool                           { return false }
func (stopAfter) PursueEducation(Character) bool                    { return false }
func (stopAfter) ChooseTradeSchool(Character) bool                  { return false }
func (stopAfter) PursueGraduateSchool(Character) bool               { return false }
func (stopAfter) TakeWaiver(Character, int) bool                    { return true }
func (stopAfter) NextCareer(Character) (Career, bool)               { return Career{}, false }
func (stopAfter) ChooseExplorerDuty(Character) bool                 { return true }

func TestContinueTarget(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 10, 8, 6}}
	if got := (ContinueRule{Fixed: 7, Mod: 2}).target(c); got != 9 {
		t.Errorf("fixed target = %d, want 9", got)
	}
	if got := (ContinueRule{UseChar: true, Char: Intelligence, Mod: -1}).target(c); got != 9 {
		t.Errorf("char target = %d, want 9 (Int 10 - 1)", got)
	}
}

func TestSelectCCRotation(t *testing.T) {
	c := Character{}
	run := careerRun{ccPool: append([]Characteristic(nil), testCareer.ControllingChars...)}
	p := stopAfter{} // ChooseCC returns available[0]
	// The pool drains as each is picked, then refills once exhausted.
	seen := map[Characteristic]int{}
	for range 8 {
		seen[selectCC(p, c, &run, testCareer)]++
	}
	for _, ch := range testCareer.ControllingChars {
		if seen[ch] != 2 {
			t.Errorf("over two cycles %s picked %d times, want 2", ch, seen[ch])
		}
	}
}

func TestSelectCCFixed(t *testing.T) {
	fixed := Career{CCMode: FixedCC, ControllingChars: []Characteristic{Endurance}}
	run := careerRun{}
	for range 3 {
		if got := selectCC(stopAfter{}, Character{}, &run, fixed); got != Endurance {
			t.Fatalf("fixed CC = %v, want Endurance", got)
		}
	}
}

func TestRunCareerRejectsEmptyCC(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("RunCareer did not panic on a career with no controlling characteristics")
		}
	}()
	c := Character{Age: 18}
	RunCareer(dice.NewScripted(3, 4), stopAfter{1}, &c, Career{Name: "Broken"})
}

func TestRunCareerTermCount(t *testing.T) {
	// Policy stops after 2 terms; Int 10 makes both Continue rolls (7) succeed.
	c := Character{scores: [count]int{7, 7, 7, 10, 8, 6}, Age: 18}
	RunCareer(dice.NewScripted(3, 4), stopAfter{2}, &c, testCareer)
	if len(c.Careers) != 1 || c.Careers[0].Terms != 2 || c.Age != 26 {
		t.Fatalf("run = %+v age %d, want 2 terms, age 26", c.Careers, c.Age)
	}
	if c.Careers[0].Outcome != MusteredOut {
		t.Errorf("outcome = %v, want MusteredOut", c.Careers[0].Outcome)
	}
}

func TestRunCareerMandatoryContinue(t *testing.T) {
	// Policy wants to stop after 1 term, but a natural 2 forces a second term.
	c := Character{scores: [count]int{7, 7, 7, 10, 8, 6}, Age: 18}
	// Term 1 Continue rolls 2 (mandatory); term 2 Continue rolls 7 (honours the stop).
	RunCareer(dice.NewScripted(1, 1, 3, 4), stopAfter{1}, &c, testCareer)
	if c.Careers[0].Terms != 2 {
		t.Fatalf("mandatory continue: %d terms, want 2", c.Careers[0].Terms)
	}
}

func TestRunCareerStopsAtAging(t *testing.T) {
	// Every roll is 7: Risk, Reward, and Continue all succeed, and the age-34
	// aging check (2D=7 < stage 5 is false) inflicts nothing. DefaultPolicy then
	// serves until aging begins — four terms from 18 to 34.
	c := Character{scores: [count]int{7, 7, 7, 10, 8, 6}, Age: 18}
	RunCareer(dice.NewScripted(3, 4), DefaultPolicy{}, &c, testCareer)
	if c.Careers[0].Terms != 4 || c.Age != 34 {
		t.Fatalf("default policy: %d terms age %d, want 4 terms age 34", c.Careers[0].Terms, c.Age)
	}
}

func TestClassifyInjury(t *testing.T) {
	cases := []struct {
		original, negMods, flux int
		wantInjury              Injury
		wantValue               int
	}{
		{7, 0, 0, Unharmed, 7},   // no reduction
		{7, 0, 2, Unharmed, 7},   // Flux healed above original -> restored
		{7, 0, -1, Wounded, 6},   // reduced by 1
		{7, 0, -3, Wounded, 4},   // reduced by 3 (still a wound)
		{7, 0, -4, Disabling, 3}, // reduced by 4
		{7, 3, -1, Disabling, 3}, // bravery -3 plus Flux -1 = reduced by 4
		{3, 0, -3, Fatal, 0},     // reduced to exactly 0
		{4, 0, -5, Fatal, -1},    // reduced below 0 (worst Flux)
	}
	for _, c := range cases {
		gotInjury, gotVal := classifyInjury(c.original, c.negMods, c.flux)
		if gotInjury != c.wantInjury || gotVal != c.wantValue {
			t.Errorf("classifyInjury(%d,%d,%d) = (%v,%d), want (%v,%d)",
				c.original, c.negMods, c.flux, gotInjury, gotVal, c.wantInjury, c.wantValue)
		}
	}
}

// runOneTerm is a helper: build a single-CC run and resolve one term.
func runOneTerm(r *dice.Roller, c *Character, cc Characteristic) TermOutcome {
	career := Career{Name: "T", ControllingChars: []Characteristic{cc}}
	run := careerRun{ccPool: []Characteristic{cc}}
	return runTerm(r, stopAfter{}, c, &run, career)
}

func TestRunTermRiskSuccess(t *testing.T) {
	// Strength 7, Risk target 7: a roll of 7 survives and the CC is untouched.
	c := Character{scores: [count]int{7, 7, 7, 8, 8, 8}}
	if got := runOneTerm(dice.NewScripted(3, 4), &c, Strength); got != Ongoing {
		t.Fatalf("outcome = %v, want Ongoing", got)
	}
	if c.scores[Strength] != 7 || c.WoundBadges != 0 {
		t.Fatalf("survived term changed state: Str %d badges %d", c.scores[Strength], c.WoundBadges)
	}
}

func TestRunTermWounded(t *testing.T) {
	// Strength 7, Risk fails (roll 8 > 7), Flux -1 -> reduced to 6: a wound.
	c := Character{scores: [count]int{7, 7, 7, 8, 8, 8}}
	got := runOneTerm(dice.NewScripted(4, 4 /*risk 8, fail*/, 2, 3 /*flux -1*/), &c, Strength)
	if got != Ongoing || c.scores[Strength] != 6 || c.WoundBadges != 1 {
		t.Fatalf(
			"wound: outcome %v Str %d badges %d, want Ongoing/6/1",
			got,
			c.scores[Strength],
			c.WoundBadges,
		)
	}
}

func TestRunTermDisabled(t *testing.T) {
	// Risk fails, Flux -4 reduces Strength 7 -> 3 (reduced by 4): Disabled.
	c := Character{scores: [count]int{7, 7, 7, 8, 8, 8}}
	got := runOneTerm(dice.NewScripted(4, 4 /*risk 8*/, 1, 5 /*flux -4*/), &c, Strength)
	if got != Disabled || c.scores[Strength] != 3 {
		t.Fatalf("disabled: outcome %v Str %d, want Disabled/3", got, c.scores[Strength])
	}
}

func TestRunTermFatal(t *testing.T) {
	// Strength 3, Risk fails (roll 4 > 3), Flux -3 -> 0: fatal.
	c := Character{scores: [count]int{3, 7, 7, 8, 8, 8}}
	got := runOneTerm(dice.NewScripted(2, 2 /*risk 4*/, 1, 4 /*flux -3*/), &c, Strength)
	if got != Died || !c.Dead {
		t.Fatalf("fatal: outcome %v dead %v, want Died/true", got, c.Dead)
	}
}

func TestApplyCell(t *testing.T) {
	var c Character
	// A plain skill.
	applyCell(stopAfter{}, &c, Cell{Kind: AwardSkill, Skill: "Navigation"})
	if c.Skills.Level("Navigation") != 1 {
		t.Errorf("Navigation = %d, want 1", c.Skills.Level("Navigation"))
	}
	// A cascade skill grants a knowledge (K-K-S).
	applyCell(stopAfter{}, &c, Cell{Kind: AwardSkill, Skill: "Pilot", Knowledge: "Small Craft"})
	if c.Skills.KnowledgeLevel("Pilot", "Small Craft") != 1 || c.Skills.Level("Pilot") != 0 {
		t.Errorf(
			"cascade award wrong: K=%d S=%d",
			c.Skills.KnowledgeLevel("Pilot", "Small Craft"),
			c.Skills.Level("Pilot"),
		)
	}
	// A choice picks the first option (DefaultPolicy/stopAfter).
	applyCell(stopAfter{}, &c, Cell{Kind: AwardChoice, Options: []string{"Gambler", "Carousing"}})
	if c.Skills.Level("Gambler") != 1 {
		t.Errorf("choice award = %d, want Gambler 1", c.Skills.Level("Gambler"))
	}
	// A cascade choice grants the chosen option as a knowledge under the parent.
	applyCell(
		stopAfter{},
		&c,
		Cell{Kind: AwardChoice, Skill: "Language", Options: []string{"Galanglic", "Vilani"}},
	)
	if c.Skills.KnowledgeLevel("Language", "Galanglic") != 1 || c.Skills.Level("Language") != 0 {
		t.Errorf("cascade choice wrong: K=%d S=%d, want 1/0",
			c.Skills.KnowledgeLevel("Language", "Galanglic"), c.Skills.Level("Language"))
	}
	// A characteristic bump, capped at the human maximum.
	c.scores[Strength] = 14
	applyCell(stopAfter{}, &c, Cell{Kind: AwardBump, Char: Strength})
	applyCell(stopAfter{}, &c, Cell{Kind: AwardBump, Char: Strength})
	if c.scores[Strength] != maxCharacteristic {
		t.Errorf("bump cap: Str = %d, want %d", c.scores[Strength], maxCharacteristic)
	}
	// Major/Minor cells raise the declared subject.
	c.Major, c.Minor = "Psychology", "Robotics"
	applyCell(stopAfter{}, &c, Cell{Kind: AwardMajor})
	applyCell(stopAfter{}, &c, Cell{Kind: AwardMinor})
	if c.Skills.Level("Psychology") != 1 || c.Skills.Level("Robotics") != 1 {
		t.Errorf("major/minor award: Psychology=%d Robotics=%d, want 1/1",
			c.Skills.Level("Psychology"), c.Skills.Level("Robotics"))
	}
}

func TestApplyCellMajorLostWithoutEducation(t *testing.T) {
	// With no Major/Minor declared, the Academic cells are lost (no-op).
	var c Character
	applyCell(stopAfter{}, &c, Cell{Kind: AwardMajor})
	applyCell(stopAfter{}, &c, Cell{Kind: AwardMinor})
	if c.Skills.String() != "" {
		t.Errorf("uneducated major/minor cells granted skills: %q", c.Skills.String())
	}
}

func TestChooseSkillColumnDependsOnEducation(t *testing.T) {
	// The Scout Academic column (col 1) is productive only for a graduate; an
	// uneducated character falls through to Courier (col 2).
	uneducated := DefaultPolicy{}.ChooseSkillColumn(Character{}, ScoutCareer.Skills)
	if uneducated != 2 {
		t.Errorf("uneducated column = %d, want 2 (Courier)", uneducated)
	}
	graduate := DefaultPolicy{}.ChooseSkillColumn(
		Character{Major: "Psychology", Minor: "Robotics"},
		ScoutCareer.Skills,
	)
	if graduate != 1 {
		t.Errorf("graduate column = %d, want 1 (Academic)", graduate)
	}
}

// commsGrid is a grid whose column 0 always awards Comms.
func commsGrid() SkillGrid {
	var g SkillGrid
	for row := range g[0] {
		g[0][row] = Cell{Kind: AwardSkill, Skill: "Comms"}
	}
	return g
}

func TestAwardSkills(t *testing.T) {
	c := Character{}
	career := Career{EligPerTerm: 3, Skills: commsGrid()}
	// stopAfter picks column 0; each 1D lands on a Comms cell.
	awardSkills(dice.NewScripted(2, 4, 6), stopAfter{}, &c, career)
	if got := c.Skills.Level("Comms"); got != 3 {
		t.Fatalf("Comms = %d, want 3 (three eligibility rolls)", got)
	}
}

func TestRunTermAwardsSkillsOnSurvival(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 8, 8, 8}}
	career := Career{
		Name:             "S",
		ControllingChars: []Characteristic{Strength},
		EligPerTerm:      1,
		Skills:           commsGrid(),
	}
	run := careerRun{ccPool: []Characteristic{Strength}}
	// Risk 7 (survive), Reward 7, then one skill roll.
	if got := runTerm(
		dice.NewScripted(3, 4, 3, 4, 5),
		stopAfter{},
		&c,
		&run,
		career,
	); got != Ongoing {
		t.Fatalf("outcome = %v, want Ongoing", got)
	}
	if c.Skills.Level("Comms") != 1 {
		t.Errorf("survived term granted %d Comms, want 1", c.Skills.Level("Comms"))
	}
}

func TestRunTermNoSkillsWhenDisabled(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 8, 8, 8}}
	career := Career{
		Name:             "S",
		ControllingChars: []Characteristic{Strength},
		EligPerTerm:      1,
		Skills:           commsGrid(),
	}
	run := careerRun{ccPool: []Characteristic{Strength}}
	// Risk fails (8 > 7), Flux -4 disables: no skills awarded.
	if got := runTerm(
		dice.NewScripted(4, 4, 1, 5),
		stopAfter{},
		&c,
		&run,
		career,
	); got != Disabled {
		t.Fatalf("outcome = %v, want Disabled", got)
	}
	if c.Skills.Level("Comms") != 0 {
		t.Errorf("disabled term granted skills: Comms %d", c.Skills.Level("Comms"))
	}
}

func TestApplyCellChoiceEmptyPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("AwardChoice with no options did not panic")
		}
	}()
	var c Character
	applyCell(stopAfter{}, &c, Cell{Kind: AwardChoice})
}

// badColumn is a policy that returns an out-of-range skill column.
type badColumn struct{ stopAfter }

func (badColumn) ChooseSkillColumn(Character, SkillGrid) int { return 99 }

func TestAwardSkillsBadColumnPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("out-of-range skill column did not panic")
		}
	}()
	c := Character{}
	awardSkills(dice.NewScripted(3), badColumn{}, &c, Career{EligPerTerm: 1, Skills: commsGrid()})
}

func TestQualificationTargetRejectsEmpty(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Qualification.target did not panic on an empty characteristic set")
		}
	}()
	Qualification{}.target(Character{})
}

func TestGenerateCareeredQualify(t *testing.T) {
	// UPP rolls (12 dice) giving Int = 8; the qualify/retry rolls follow.
	upp := []int{4, 4, 4, 4, 4, 4, 4, 4, 5, 5, 3, 3}

	// A qualify roll of 2 (<= Int) succeeds, so the character runs the career.
	seq := append(append([]int{}, upp...), 1, 1 /*qualify 2*/, 3, 4 /*continue*/)
	c := GenerateCareered(dice.NewScripted(seq...), stopAfter{1}, worldgen.World{}, testCareer)
	if len(c.Careers) != 1 || c.Careers[0].Career != testCareer.ID {
		t.Fatalf("qualified but no career recorded: %+v", c.Careers)
	}

	// A failed qualify (12) has no Begin retry (Book 1 p.65: no career in this
	// edition grants one), so the character falls straight back to the auto-begin
	// Citizen life rather than getting a second roll or ending up careerless.
	seqDraft := append(append([]int{}, upp...), 6, 6 /*qualify 12*/)
	cDraft := GenerateCareered(
		dice.NewScripted(seqDraft...),
		stopAfter{1},
		worldgen.World{},
		testCareer,
	)
	if len(cDraft.Careers) != 1 || cDraft.Careers[0].Career != Citizen {
		t.Fatalf("a failed Begin should fall back to Citizen (no retry): %+v", cDraft.Careers)
	}
}

// twoCitizen serves an extra Citizen career once, then stops.
type twoCitizen struct {
	stopAfter
	served bool
}

func (p *twoCitizen) NextCareer(Character) (Career, bool) {
	if !p.served {
		p.served = true
		return CitizenCareer, true
	}
	return Career{}, false
}

func TestMultiCareer(t *testing.T) {
	// The Citizen auto-begins, so a character serves two one-term Citizen careers
	// in sequence regardless of the exact rolls.
	c := GenerateCareered(
		dice.NewScripted(3, 4),
		&twoCitizen{stopAfter: stopAfter{1}},
		worldgen.World{},
		CitizenCareer,
	)
	if len(c.Careers) != 2 {
		t.Fatalf("careers = %d, want 2 (a sequence of two)", len(c.Careers))
	}
	if c.Age != 26 {
		t.Errorf("age = %d, want 26 (18 + 2 careers x 1 term)", c.Age)
	}
	if c.Careers[0].Career != Citizen || c.Careers[1].Career != Citizen {
		t.Errorf("records = %+v, want two Citizen careers", c.Careers)
	}
}
