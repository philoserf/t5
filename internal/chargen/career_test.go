package chargen

import (
	"slices"
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

// testBegin resolves a Begin under DefaultPolicy on a fresh run, the way
// serveCareer does.
func testBegin(r *dice.Roller, c *Character, career Career) bool {
	run := newCareerRun(career)

	return beginCareer(r, DefaultPolicy{}, c, &run, career)
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
func (stopAfter) RerollBranch(Character, CareerRecord) bool         { return false }
func (stopAfter) RerollBranchOnCommission(Character) bool           { return false }

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
	// Each term rolls Risk, Reward and Continue, all 2D: two terms, six rolls,
	// twelve dice, every one of them a 7.
	RunCareer(dice.NewScripted(slices.Repeat([]int{3, 4}, 6)...), stopAfter{2}, &c, testCareer)

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
	// Every term rolls Risk, Reward and Continue in that order, so the Continue
	// rolls are the third and sixth 2D of the script: term 1 Continues on a
	// natural 2 (mandatory), term 2 on a 7 (which honours the stop).
	RunCareer(dice.NewScripted(
		3, 4, // term 1 Risk: 7
		3, 4, // term 1 Reward: 7
		1, 1, // term 1 Continue: natural 2, mandatory
		3, 4, // term 2 Risk: 7
		3, 4, // term 2 Reward: 7
		3, 4, // term 2 Continue: 7, the policy's stop is honoured
	), stopAfter{1}, &c, testCareer)

	if c.Careers[0].Terms != 2 {
		t.Fatalf("mandatory continue: %d terms, want 2", c.Careers[0].Terms)
	}
}

func TestRunCareerStopsAtAging(t *testing.T) {
	// Every roll is 7: Risk, Reward, and Continue all succeed, and the age-34
	// aging check (2D=7 < stage 5 is false) inflicts nothing. DefaultPolicy then
	// serves until aging begins — four terms from 18 to 34.
	c := Character{scores: [count]int{7, 7, 7, 10, 8, 6}, Age: 18}
	// Four terms of Risk/Reward/Continue (24 dice) plus the age-34 aging check's
	// three physical 2D rolls (6 dice): thirty dice, every roll a 7.
	RunCareer(dice.NewScripted(slices.Repeat([]int{3, 4}, 15)...), DefaultPolicy{}, &c, testCareer)

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
//
//nolint:unparam // test helper kept general over the controlling characteristic
func runOneTerm(r *dice.Roller, c *Character, cc Characteristic) TermOutcome {
	career := Career{Name: "T", ControllingChars: []Characteristic{cc}}
	run := careerRun{ccPool: []Characteristic{cc}}

	return runTerm(r, stopAfter{}, c, &run, career)
}

func TestRunTermRiskSuccess(t *testing.T) {
	// Strength 7, Risk target 7: a roll of 7 survives and the CC is untouched.
	c := Character{scores: [count]int{7, 7, 7, 8, 8, 8}}
	// Risk 7 (held), then the Reward roll every term makes (Book 1 p.65).
	if got := runOneTerm(dice.NewScripted(3, 4, 3, 4), &c, Strength); got != Ongoing {
		t.Fatalf("outcome = %v, want Ongoing", got)
	}

	if c.scores[Strength] != 7 || c.WoundBadges != 0 {
		t.Fatalf("survived term changed state: Str %d badges %d", c.scores[Strength], c.WoundBadges)
	}
}

func TestRunTermWounded(t *testing.T) {
	// Strength 7, Risk fails (roll 8 > 7), Flux -1 -> reduced to 6: a wound.
	c := Character{scores: [count]int{7, 7, 7, 8, 8, 8}}

	// Reward is rolled even though Risk failed (Book 1 p.65), so its 2D follows
	// the injury Flux in the script; career "T" has RewardNone, so it grants
	// nothing here beyond consuming its roll.
	got := runOneTerm(
		dice.NewScripted(4, 4 /*risk 8, fail*/, 2, 3 /*flux -1*/, 3, 4 /*reward*/),
		&c,
		Strength,
	)
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

	got := runOneTerm(
		dice.NewScripted(4, 4 /*risk 8*/, 1, 5 /*flux -4*/, 3, 4 /*reward*/),
		&c,
		Strength,
	)
	if got != Disabled || c.scores[Strength] != 3 {
		t.Fatalf("disabled: outcome %v Str %d, want Disabled/3", got, c.scores[Strength])
	}
}

func TestRunTermFatal(t *testing.T) {
	// Strength 3, Risk fails (roll 4 > 3), Flux -3 -> 0: fatal.
	c := Character{scores: [count]int{3, 7, 7, 8, 8, 8}}

	got := runOneTerm(
		dice.NewScripted(2, 2 /*risk 4*/, 1, 4 /*flux -3*/, 3, 4 /*reward*/),
		&c,
		Strength,
	)
	if got != Died || !c.Dead {
		t.Fatalf("fatal: outcome %v dead %v, want Died/true", got, c.Dead)
	}
}

// TestRunTermRewardRollsWhenRiskFails locks Book 1 p.65's sequence: "The
// Character rolls for Risk ... and determines the outcome. He then rolls again
// for Reward ... and determines the consequences." Reward is rolled every term,
// not only on a held Risk.
//
// It is shaped after the Eneri Dinsha worked example's first term (p.66), which
// is the book's own demonstration: Eneri fails Risk against Endurance-11, takes
// a Wound Badge, and *still* rolls Reward — "he rolls 3 and succeeds again. He
// will receive a Medal." Eneri's Branch/Operations and Caution mods are stripped
// to zero here so the test isolates the rule from the armed-forces mod stack.
//
// The Reward target is deliberately checked against the ORIGINAL characteristic:
// Eneri's Reward is rolled against Endurance-11 even though the injury has just
// dropped him to Endurance-10.
func TestRunTermRewardRollsWhenRiskFails(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 11, 8, 8, 8}} // Endurance 11, as Eneri
	career := Career{
		Name:             "T",
		ControllingChars: []Characteristic{Endurance},
		RewardKind:       RewardMedal,
	}
	run := careerRun{ccPool: []Characteristic{Endurance}}

	// Risk 2D=12 > End 11 fails; Flux -1 wounds him to End 10; Reward 2D=3 is
	// still rolled and succeeds against the original End 11, earning a Medal.
	got := runTerm(
		dice.NewScripted(6, 6 /*risk 12, fail*/, 2, 3 /*flux -1*/, 1, 2 /*reward 3*/),
		stopAfter{},
		&c,
		&run,
		career,
	)

	if got != Ongoing {
		t.Fatalf("outcome = %v, want Ongoing (a wound does not end the term)", got)
	}

	if c.WoundBadges != 1 || c.scores[Endurance] != 10 {
		t.Errorf("injury: badges %d End %d, want 1/10", c.WoundBadges, c.scores[Endurance])
	}

	if c.Medals != 1 {
		t.Errorf("Medals = %d, want 1: Reward is rolled even when Risk fails (Book 1 p.65)",
			c.Medals)
	}
}

// TestRunTermRewardRollsWhenRiskKills confirms the Reward roll is drawn from the
// dice stream even on a fatal term, so a death does not silently shift every
// subsequent roll of a multi-character generation.
func TestRunTermRewardRollsWhenRiskKills(t *testing.T) {
	c := Character{scores: [count]int{3, 7, 7, 8, 8, 8}}
	career := Career{
		Name:             "T",
		ControllingChars: []Characteristic{Strength},
		RewardKind:       RewardMedal,
	}
	run := careerRun{ccPool: []Characteristic{Strength}}
	r := dice.NewScripted(2, 2 /*risk 4, fail*/, 1, 4 /*flux -3, fatal*/, 1, 2 /*reward*/, 6)

	if got := runTerm(r, stopAfter{}, &c, &run, career); got != Died {
		t.Fatalf("outcome = %v, want Died", got)
	}
	// The Reward 2D was consumed, so the next draw is the trailing 6, not part
	// of the reward roll.
	if next := r.Die(); next != 6 {
		t.Errorf("next die = %d, want 6: the Reward roll should have consumed its 2D", next)
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
	// Risk fails (8 > 7), Flux -4 disables: no skills awarded. The Reward roll
	// still happens first (Book 1 p.65), so its 2D is scripted after the Flux.
	if got := runTerm(
		dice.NewScripted(4, 4, 1, 5, 3, 4),
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

	// A qualify roll of 2 (<= Int) succeeds, so the character runs the career:
	// one term of Risk, Reward and Continue, then the muster-out benefit die.
	seq := append(append([]int{}, upp...),
		1, 1, // qualify: 2
		3, 4, // term 1 Risk: 7
		3, 4, // term 1 Reward: 7
		3, 4, // term 1 Continue: 7, the policy's one-term stop is honoured
		4, // muster-out benefit
	)

	c := GenerateCareered(dice.NewScripted(seq...), stopAfter{1}, worldgen.World{}, testCareer)
	if len(c.Careers) != 1 || c.Careers[0].Career != testCareer.ID {
		t.Fatalf("qualified but no career recorded: %+v", c.Careers)
	}

	// A failed qualify (12) has no Begin retry (Book 1 p.65: no career in this
	// edition grants one), so the character falls straight back to the auto-begin
	// Citizen life rather than getting a second roll or ending up careerless.
	// After the failed qualify the character serves one auto-begin Citizen term:
	// the Citizen Life roll, its skill roll, the Continue roll, and muster-out.
	seqDraft := append(append([]int{}, upp...), 6, 6 /*qualify 12*/)
	seqDraft = append(seqDraft, slices.Repeat([]int{4}, 9)...)

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
	// Thirty dice: the twelve UPP dice and then two one-term Citizen careers.
	c := GenerateCareered(
		dice.NewScripted(slices.Repeat([]int{3, 4}, 15)...),
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

// rerollingPolicy always takes the p. 66 end-of-term Branch reroll, and rolls
// for a new Branch on Commission — the opposite of every default.
type rerollingPolicy struct{ stopAfter }

func (rerollingPolicy) RerollBranch(Character, CareerRecord) bool { return true }
func (rerollingPolicy) RerollBranchOnCommission(Character) bool   { return true }

// TestCommissionBranchColumnSwitch locks the one Branch mapping Book 1 p. 66
// names by hand: "for Spacers, Crew becomes Line". A commissioned Spacer keeps
// his Branch, but reads it from the Officer column — Crew has no officer
// counterpart and becomes Line, while a branch printed in both columns is
// unchanged. A career with a single Branch table keeps exactly what it had.
func TestCommissionBranchColumnSwitch(t *testing.T) {
	for _, tc := range []struct {
		roll int
		want string
	}{
		{1, "Line"},     // enlisted Crew — the case the book spells out
		{2, "Line"},     // enlisted Crew
		{3, "Engineer"}, // present in both columns: unchanged
		{5, "Gunnery"},
		{7, "Technical"},
		{8, "Medical"},
	} {
		if got := spacerBranchOps.commissionBranch(tc.roll).Name; got != tc.want {
			t.Errorf("Spacer branch roll %d commissioned = %q, want %q", tc.roll, got, tc.want)
		}
	}

	// The engine path: a commissioned Spacer who keeps his Branch (the default)
	// rolls no die and takes the Officer column's mod for the same roll.
	c := Character{scores: [count]int{7, 7, 7, 7, 7, 7}}
	run := careerRun{officer: true, branchRoll: 1} // enlisted Crew, just commissioned
	changeBranchOnCommission(dice.NewScripted(6), stopAfter{}, &c, &run, SpacerCareer)

	if want := spacerBranchOps.Branches[1]; run.branchMod != want.Mod || run.branchRoll != 1 {
		t.Errorf("commissioned Spacer = roll %d mod %d, want roll 1 mod %d (officer Line)",
			run.branchRoll, run.branchMod, want.Mod)
	}

	// The Soldier prints one Branch table, so a Commission changes nothing.
	for roll := 1; roll <= 8; roll++ {
		want := soldierBranchOps.Branches[roll].Name
		if got := soldierBranchOps.commissionBranch(roll).Name; got != want {
			t.Errorf("Soldier branch roll %d commissioned = %q, want %q (single column)", roll, got, want)
		}
	}
}

// TestRerollBranchIsDiceNeutralByDefault is the guarantee that made this rule
// safe to add: keeping the Branch — what every default policy does — rolls no
// die, so the armed-forces goldens do not move. A policy that rerolls draws one
// extra die per surviving term, and an officer never gets the offer.
func TestRerollBranchIsDiceNeutralByDefault(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 7, 7, 7}}
	run := careerRun{branchRoll: 1, branchMod: 9, branchOpsDM: 9}

	// Keeping: the script's lone 6 (enlisted Gunnery, mod 1) is never drawn, so
	// the Branch keeps the mod 9 no table holds.
	rerollBranch(dice.NewScripted(6), stopAfter{}, &c, &run, SpacerCareer, CareerRecord{})

	if run.branchMod != 9 {
		t.Errorf("keeping changed the Branch: mod = %d, want the untouched 9", run.branchMod)
	}

	// Rerolling: one die, read from the Enlisted column (roll 6 = Gunnery, mod 1).
	rerollBranch(dice.NewScripted(6), rerollingPolicy{}, &c, &run, SpacerCareer, CareerRecord{})

	if run.branchRoll != 6 || run.branchMod != 1 {
		t.Errorf("reroll = roll %d mod %d, want roll 6 mod 1 (enlisted Gunnery)", run.branchRoll, run.branchMod)
	}

	// An Officer may not change Branch (p. 66), whatever the policy wants.
	officer := careerRun{officer: true, branchRoll: 1, branchMod: 9}
	rerollBranch(dice.NewScripted(6), rerollingPolicy{}, &c, &officer, SpacerCareer, CareerRecord{})

	if officer.branchMod != 9 {
		t.Errorf("an officer changed Branch: mod = %d, want the untouched 9", officer.branchMod)
	}
}
