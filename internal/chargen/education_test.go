package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

// eduPolicy is a scripted education policy: it hands out a fixed list of Major /
// Minor picks in order and always takes waivers.
type eduPolicy struct {
	DefaultPolicy
	picks []string
	i     int
}

func (p *eduPolicy) ChooseSkill(_ Character, _ []string) string {
	s := p.picks[p.i]
	p.i++
	return s
}

// TestGoldenCollege reproduces the Book 1 p. 60 worked example: Eneri Dinsha
// (9AB56A) goes to the College of Regina — rejected then admitted on a Waiver,
// passes years 1 and 4, waives the failures of years 2 and 3, and graduates
// with a BA (Edu 6 -> 8), a Psychology Major-2 and a Robotics Minor-1: 9AB58A.
func TestGoldenCollege(t *testing.T) {
	c := Character{scores: [count]int{9, 10, 11, 5, 6, 10}} // 9AB56A; Int 5, Edu 6, Soc 10
	p := &eduPolicy{picks: []string{"Psychology", "Robotics"}}

	seq := []int{
		5, 6, // apply Check Edu(6): 11 > 6, rejected
		4, 5, // Waiver Check Soc(10) mod 0: 9 <= 10, admitted
		1, 1, // year 1 Check Edu(6): 2, pass -> Psychology-1
		5, 6, // year 2: 11, fail
		4, 4, // Waiver Soc(10) mod -1 = 9: 8, continue
		5, 6, // year 3: 11, fail
		3, 3, // Waiver Soc(10) mod -2 = 8: 6, continue
		1, 1, // year 4: 2, pass -> Psychology-2, Robotics-1 (2nd pass)
	}
	attendAcademic(dice.NewScripted(seq...), p, &c, college)

	if got := c.UPP(); got != "9AB58A" {
		t.Errorf("UPP = %q, want %q (Edu 6 -> 8 on graduation)", got, "9AB58A")
	}
	if c.Major != "Psychology" || c.Skills.Level("Psychology") != 2 {
		t.Errorf("Major = %q level %d, want Psychology 2", c.Major, c.Skills.Level("Psychology"))
	}
	if c.Minor != "Robotics" || c.Skills.Level("Robotics") != 1 {
		t.Errorf("Minor = %q level %d, want Robotics 1", c.Minor, c.Skills.Level("Robotics"))
	}
	if len(c.Degrees) != 1 || c.Degrees[0] != "BA" {
		t.Errorf("Degrees = %v, want [BA]", c.Degrees)
	}
}

func TestCollegePrerequisite(t *testing.T) {
	// Edu 4 is below the College prerequisite of 5: no attendance, no dice.
	c := Character{scores: [count]int{7, 7, 7, 7, 4, 7}}
	attendAcademic(dice.NewScripted(1, 1), DefaultPolicy{}, &c, college)
	if len(c.Degrees) != 0 || c.Major != "" || c.Score(Education) != 4 {
		t.Errorf("under-prereq character was educated: %+v", c)
	}
}

func TestED5RaisesLowEdu(t *testing.T) {
	// Edu 3, Check Int(8) passes -> Edu raised to 5.
	c := Character{scores: [count]int{7, 7, 7, 8, 3, 7}}
	attemptED5(dice.NewScripted(3, 4), &c) // roll 7 <= Int 8, pass
	if c.Score(Education) != 5 {
		t.Errorf("ED5 pass: Edu = %d, want 5", c.Score(Education))
	}
	// A failed Int Check leaves Edu unchanged.
	c2 := Character{scores: [count]int{7, 7, 7, 4, 3, 7}}
	attemptED5(dice.NewScripted(6, 6), &c2) // roll 12 > Int 4, fail
	if c2.Score(Education) != 3 {
		t.Errorf("ED5 fail: Edu = %d, want 3", c2.Score(Education))
	}
	// Edu already above the ED5 ceiling is a no-op: the Check is never made, so
	// Edu stays 6 rather than being pulled down to the ED5 target of 5 (a passing
	// roll of 7 is scripted, which would set Edu 5 if the Check were rolled).
	c3 := Character{scores: [count]int{7, 7, 7, 8, 6, 7}}
	attemptED5(dice.NewScripted(3, 4), &c3)
	if c3.Score(Education) != 6 {
		t.Errorf("ED5 no-op: Edu = %d, want 6", c3.Score(Education))
	}
}

func TestEducateSelectsUniversity(t *testing.T) {
	// Edu 7 qualifies for University (grad Edu 9); all rolls at 7 pass.
	c := Character{scores: [count]int{7, 7, 7, 7, 7, 7}}
	seq := []int{3, 4, 3, 4, 3, 4, 3, 4, 3, 4} // apply + four passes, each 7 <= Edu 7
	educate(dice.NewScripted(seq...), &eduPolicy{picks: []string{"Physics", "History"}}, &c)
	if c.Score(Education) != 9 || len(c.Degrees) != 1 {
		t.Errorf("University graduation: Edu = %d degrees %v, want 9 / [BA]", c.Score(Education), c.Degrees)
	}
}

func TestCollegeRejectedNoWaiver(t *testing.T) {
	// Admission fails and the policy declines a waiver: nothing gained.
	c := Character{scores: [count]int{7, 7, 7, 6, 6, 6}}
	attendAcademic(dice.NewScripted(6, 6 /*apply 12, fail*/), noWaiver{}, &c, college)
	if len(c.Degrees) != 0 || c.Major != "" {
		t.Errorf("rejected applicant was educated: %+v", c)
	}
}

func TestCollegeFailsOut(t *testing.T) {
	// Admitted, but fails a year and declines the waiver: no graduation, yet the
	// passes before the wash-out are kept.
	c := Character{scores: [count]int{7, 7, 7, 6, 6, 6}}
	seq := []int{
		3, 3, // apply Check Edu(6): 6 <= 6, admitted
		1, 1, // year 1: pass -> Major +1
		6, 6, // year 2: fail, no waiver -> wash out
	}
	attendAcademic(dice.NewScripted(seq...), noWaiver{}, &c, college)
	if len(c.Degrees) != 0 {
		t.Errorf("washed-out student graduated: degrees %v", c.Degrees)
	}
	if c.Major == "" || c.Skills.Level(c.Major) != 1 {
		t.Errorf("pre-washout pass not kept: Major %q level %d", c.Major, c.Skills.Level(c.Major))
	}
}

// noWaiver is DefaultPolicy but never takes a waiver.
type noWaiver struct{ DefaultPolicy }

func (noWaiver) TakeWaiver(Character, int) bool { return false }

// tradePolicy takes the vocational Trade School path.
type tradePolicy struct{ DefaultPolicy }

func (tradePolicy) ChooseTradeSchool(Character) bool { return true }

// TestTradeSchool runs a character through a one-year Trade School: admitted on
// the Int check, passes the year, and earns a vocational Major +2 (Biologics,
// theTrades[0], via DefaultPolicy.ChooseSkill) — with no Minor, Edu bump, or
// degree, and taking the academic path's place.
func TestTradeSchool(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 7, 6, 7}} // Int 7, Edu 6
	seq := []int{
		3, 4, // apply Check Int(7): 7 <= 7, admitted
		3, 4, // year Check Int(7): 7 <= 7, pass
	}
	educate(dice.NewScripted(seq...), tradePolicy{}, &c)
	if c.Major != "Biologics" {
		t.Errorf("Major = %q, want Biologics", c.Major)
	}
	if c.Skills.Level("Biologics") != 2 {
		t.Errorf("Biologics = %d, want 2 (Trade School Major +2)", c.Skills.Level("Biologics"))
	}
	if c.Score(Education) != 6 {
		t.Errorf("Edu = %d, want 6 unchanged (Trade School grants no Edu bump)", c.Score(Education))
	}
	if c.Minor != "" || len(c.Degrees) != 0 {
		t.Errorf("Trade School granted a Minor %q or degree %v, want none", c.Minor, c.Degrees)
	}
}

// gradPolicy takes the academic path and goes on to graduate school.
type gradPolicy struct{ DefaultPolicy }

func (gradPolicy) PursueGraduateSchool(Character) bool { return true }

// TestGraduateLadder climbs the whole academic ladder for a graduate-minded
// character: College (Psychology Major-4, Robotics Minor-2, BA, Edu 6 -> 8), a
// Masters that raises only the Minor (Robotics -> 3, MA, Edu 9), and a terminal
// Professors program that awards no skill (Professor, Edu 12).
func TestGraduateLadder(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 8, 6, 7}} // Int 8, Edu 6 -> College
	seq := []int{
		3, 4, // College apply Check Int(8): 7 <= 8, admitted
		3, 4, 3, 4, 3, 4, 3, 4, // 4 College years pass -> Psychology-4, Robotics-2, BA, Edu 8
		3, 4, // Masters apply (best of Int 8/Edu 8): 7 <= 8, admitted
		3, 4, 3, 4, // 2 Masters years pass -> Robotics +1 (2nd pass) -> 3, MA, Edu 9
		3, 4, // Professors apply (best of Int 8/Edu 9): 7 <= 9, admitted
		3, 4, 3, 4, // 2 Professors years pass -> no skill, Professor, Edu 12
	}
	educate(dice.NewScripted(seq...), gradPolicy{}, &c)
	if c.Skills.Level("Psychology") != 4 || c.Skills.Level("Robotics") != 3 {
		t.Errorf("Psychology=%d Robotics=%d, want 4/3 (Masters raises the Minor; Professors neither)",
			c.Skills.Level("Psychology"), c.Skills.Level("Robotics"))
	}
	if c.Score(Education) != 12 {
		t.Errorf("Edu = %d, want 12 (Professor graduation)", c.Score(Education))
	}
	if !c.hasDegree("BA") || !c.hasDegree("MA") || !c.hasDegree("Professor") {
		t.Errorf("Degrees = %v, want BA, MA, and Professor", c.Degrees)
	}
}

// TestMastersNeedsDegree confirms a Masters is refused without a prior degree.
func TestMastersNeedsDegree(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 9, 9, 7}} // no degree
	attendAcademic(dice.NewScripted(1, 1, 1, 1, 1, 1), DefaultPolicy{}, &c, masters)
	if len(c.Degrees) != 0 || c.Score(Education) != 9 {
		t.Errorf("Masters admitted without a BA: degrees %v Edu %d", c.Degrees, c.Score(Education))
	}
}

// TestProfessorsDegree confirms the terminal Professors program awards no Major
// or Minor — only Edu 12 and the Professor title — and requires a prior MA.
func TestProfessorsDegree(t *testing.T) {
	c := Character{
		scores:  [count]int{7, 7, 7, 9, 9, 7},
		Major:   "Psychology",
		Minor:   "Robotics",
		Degrees: []string{"BA", "MA"},
	}
	c.Skills.Raise("Psychology", 4)
	c.Skills.Raise("Robotics", 3)
	seq := []int{
		3, 4, // apply Check (best of Int 9/Edu 9): 7 <= 9, admitted
		3, 4, 3, 4, // 2 years pass — no Major or Minor awarded
	}
	attendAcademic(dice.NewScripted(seq...), DefaultPolicy{}, &c, professors)
	if c.Skills.Level("Psychology") != 4 || c.Skills.Level("Robotics") != 3 {
		t.Errorf("Professors changed Major/Minor: Psychology=%d Robotics=%d, want 4/3 unchanged",
			c.Skills.Level("Psychology"), c.Skills.Level("Robotics"))
	}
	if c.Score(Education) != 12 {
		t.Errorf("Edu = %d, want 12 (Professor graduation)", c.Score(Education))
	}
	if !c.hasDegree("Professor") {
		t.Errorf("Degrees = %v, want a Professor degree", c.Degrees)
	}
}

// TestProfessorsNeedsMasters confirms Professors is refused without an MA.
func TestProfessorsNeedsMasters(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 9, 9, 7}, Degrees: []string{"BA"}}
	attendAcademic(dice.NewScripted(1, 1, 1, 1, 1, 1), DefaultPolicy{}, &c, professors)
	if c.hasDegree("Professor") || c.Score(Education) != 9 {
		t.Errorf("Professors admitted without an MA: degrees %v Edu %d", c.Degrees, c.Score(Education))
	}
}

// tradeNoWaiver takes Trade School but never waives a failed year.
type tradeNoWaiver struct{ DefaultPolicy }

func (tradeNoWaiver) ChooseTradeSchool(Character) bool { return true }
func (tradeNoWaiver) TakeWaiver(Character, int) bool   { return false }

// TestTradeSchoolFailsOut confirms a failed year with no waiver grants no Major.
func TestTradeSchoolFailsOut(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 7, 6, 7}}
	seq := []int{
		3, 4, // apply Check Int(7): admitted
		6, 6, // year Check Int(7): 12 > 7, fail; no waiver -> washes out
	}
	educate(dice.NewScripted(seq...), tradeNoWaiver{}, &c)
	if c.Major != "" || c.Skills.Level("Biologics") != 0 {
		t.Errorf("failed Trade School still granted Major %q (Biologics %d), want none",
			c.Major, c.Skills.Level("Biologics"))
	}
}
