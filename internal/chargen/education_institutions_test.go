package chargen

import (
	"slices"
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestAvailableSkillsMatrix(t *testing.T) {
	cases := []struct {
		institution Institution
		present     []string
		absent      []string
	}{
		{
			CollegeInstitution,
			[]string{"Athlete", "Psychology", "Robotics", "Teacher"},
			[]string{"Advocate", "Medic", "Tactics"},
		},
		{LawSchool, []string{"Advocate"}, []string{"Medic", "Psychology"}},
		{MedicalSchool, []string{"Forensics", "Medic"}, []string{"Advocate", "Astrogator", "Fighter", "Robotics"}},
		{MilitaryAcademy, []string{"Battle Dress", "Robotics", "WMD"}, []string{"Astrogator", "Fleet Tactics"}},
		{NavalAcademy, []string{"Astrogator", "Fleet Tactics", "Pilot-ACS"}, []string{"Advocate", "Psychology"}},
		{MarineAcademy, []string{"Fighter", "Medic", "Robotics"}, []string{"Advocate", "Astrogator", "Forensics"}},
		{School, []string{"Admin", "Biologics", "Wheeled Driver"}, []string{"Advocate", "Fleet Tactics"}},
	}

	for _, tc := range cases {
		got := AvailableSkills(tc.institution)
		if !slices.IsSorted(got) {
			t.Errorf("AvailableSkills(%q) is not stable/sorted", tc.institution)
		}

		for _, skill := range tc.present {
			if !slices.Contains(got, skill) {
				t.Errorf("AvailableSkills(%q) missing %q", tc.institution, skill)
			}
		}

		for _, skill := range tc.absent {
			if slices.Contains(got, skill) {
				t.Errorf("AvailableSkills(%q) unexpectedly contains %q", tc.institution, skill)
			}
		}
	}

	first := AvailableSkills(CollegeInstitution)
	first[0] = "mutated"

	if AvailableSkills(CollegeInstitution)[0] == "mutated" {
		t.Error("AvailableSkills returned shared mutable storage")
	}
}

func TestServiceAcademies(t *testing.T) {
	for _, tc := range []struct {
		institution Institution
		service     string
	}{
		{MilitaryAcademy, "Army"},
		{NavalAcademy, "Navy"},
		{MarineAcademy, "Marine"},
	} {
		c := Character{scores: [count]int{7, 7, 7, 8, 6, 7}, Age: 18}

		r := dice.NewScripted(3, 3, 3, 3, 3, 3, 3, 3, 3, 3) // admission, then four years
		if !AttendInstitution(r, DefaultPolicy{}, &c, tc.institution) {
			t.Fatalf("%q did not graduate", tc.institution)
		}

		if c.Age != 22 || c.Score(Education) != 8 || !c.hasDegree("BA") {
			t.Errorf("academy result age=%d Edu=%d degrees=%v", c.Age, c.Score(Education), c.Degrees)
		}

		last := c.EducationHistory[len(c.EducationHistory)-1]
		if last.Commission != tc.service || !last.Graduated {
			t.Errorf("academy status = %+v, want %s Officer1", last, tc.service)
		}
	}
}

func TestProfessionalSchools(t *testing.T) {
	cases := []struct {
		institution Institution
		years       int
		skill       string
		level       int
		degree      string
	}{
		{MedicalSchool, 4, "Medic", 4, "Doctor"},
		{LawSchool, 2, "Advocate", 2, "Attorney"},
	}
	for _, tc := range cases {
		c := Character{scores: [count]int{7, 7, 7, 9, 9, 7}, Age: 22, Degrees: []string{"BA", "Honors BA"}}

		rolls := make([]int, (tc.years+1)*2)
		for i := range rolls {
			rolls[i] = 3
		}

		if !AttendInstitution(dice.NewScripted(rolls...), DefaultPolicy{}, &c, tc.institution) {
			t.Fatalf("%q did not graduate", tc.institution)
		}

		if c.Age != 22+tc.years || c.Score(Education) != 10 || c.Skills.Level(tc.skill) != tc.level ||
			!c.hasDegree(tc.degree) {
			t.Errorf(
				"professional result age=%d Edu=%d %s=%d degrees=%v",
				c.Age,
				c.Score(Education),
				tc.skill,
				c.Skills.Level(tc.skill),
				c.Degrees,
			)
		}
	}
}

func TestHonorsOfficerTrainingAndFlight(t *testing.T) {
	c := Character{scores: [count]int{7, 8, 7, 8, 8, 7}, Age: 22, Major: "Psychology", Degrees: []string{"BA"}}
	if !AttendInstitution(dice.NewScripted(3, 3), DefaultPolicy{}, &c, HonorsProgram) {
		t.Fatal("Honors failed")
	}

	if c.Skills.Level("Psychology") != 1 || !c.hasDegree("Honors BA") || c.Age != 22 {
		t.Errorf("Honors result = age %d skills %v degrees %v", c.Age, c.Skills, c.Degrees)
	}

	if !AttendInstitution(dice.NewScripted(3, 3), DefaultPolicy{}, &c, NOTC) {
		t.Fatal("NOTC failed")
	}

	if c.EducationHistory[len(c.EducationHistory)-1].Commission != "Navy" {
		t.Errorf("NOTC status = %+v", c.EducationHistory)
	}

	if !AttendInstitution(dice.NewScripted(3, 3), DefaultPolicy{}, &c, FlightSchool) {
		t.Fatal("Flight School failed")
	}

	if c.Age != 23 || c.Skills.Level("Pilot") != 3 || !c.EducationHistory[len(c.EducationHistory)-1].Flight {
		t.Errorf("Flight result age=%d Pilot=%d history=%+v", c.Age, c.Skills.Level("Pilot"), c.EducationHistory)
	}
}

func TestPrerequisiteWaiverIsPolicy(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 8, 8, 8}, Age: 18}
	if AttendInstitution(dice.NewScripted(1), tradeNoWaiver{}, &c, MedicalSchool) {
		t.Fatal("Medical School admitted character without Honors BA or waiver")
	}

	if c.Age != 18 {
		t.Errorf("declined prerequisite waiver cost time: age=%d", c.Age)
	}

	// DefaultPolicy attempts the prerequisite waiver first, then admission and
	// four yearly checks: this explicit order is stable for scripted generation.
	r := dice.NewScripted(3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3)
	if !AttendInstitution(r, DefaultPolicy{}, &c, MedicalSchool) {
		t.Fatal("Medical School prerequisite waiver did not permit attendance")
	}
}

func TestAssignedMilitarySchools(t *testing.T) {
	c := Character{scores: [count]int{7, 8, 8, 9, 9, 7}, Age: 30}
	if !AttendInstitution(dice.NewScripted(3, 3), DefaultPolicy{}, &c, NavySchool) {
		t.Fatal("Naval ANM School failed")
	}

	if c.Age != 31 || c.EducationHistory[len(c.EducationHistory)-1].Institution != NavySchool {
		t.Errorf("ANM status age=%d history=%+v", c.Age, c.EducationHistory)
	}

	if !AttendInstitution(dice.NewScripted(3, 3, 3, 3), DefaultPolicy{}, &c, CommandCollege) {
		t.Fatal("Command College failed")
	}

	if c.Age != 32 || c.EducationHistory[len(c.EducationHistory)-1].Institution != CommandCollege {
		t.Errorf("Command College status age=%d history=%+v", c.Age, c.EducationHistory)
	}
}
