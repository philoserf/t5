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
	AttendCollege(dice.NewScripted(seq...), p, &c)

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
	AttendCollege(dice.NewScripted(1, 1), DefaultPolicy{}, &c)
	if len(c.Degrees) != 0 || c.Major != "" || c.Score(Education) != 4 {
		t.Errorf("under-prereq character was educated: %+v", c)
	}
}

func TestCollegeRejectedNoWaiver(t *testing.T) {
	// Admission fails and the policy declines a waiver: nothing gained.
	c := Character{scores: [count]int{7, 7, 7, 6, 6, 6}}
	AttendCollege(dice.NewScripted(6, 6 /*apply 12, fail*/), noWaiver{}, &c)
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
	AttendCollege(dice.NewScripted(seq...), noWaiver{}, &c)
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
