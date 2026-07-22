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
		{MedicalSchool, []string{"Forensics", "Medic"}, []string{"Advocate", "Astrogation", "Fighter", "Robotics"}},
		{MilitaryAcademy, []string{"Battle Dress", "Robotics", "WMD"}, []string{"Astrogation", "Fleet Tactics"}},
		{NavalAcademy, []string{"Astrogation", "Fleet Tactics", "Pilot-ACS"}, []string{"Advocate", "Psychology"}},
		{MarineAcademy, []string{"Fighter", "Medic", "Robotics"}, []string{"Advocate", "Astrogation", "Forensics"}},
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

func TestServiceAcademyEduWaiver(t *testing.T) {
	// Below Edu 6, admission requires a successful prerequisite Waiver first.
	c := Character{scores: [count]int{7, 7, 7, 8, 5, 7}, Age: 18}

	r := dice.NewScripted(3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3) // waiver, admission, then four years
	if !AttendInstitution(r, DefaultPolicy{}, &c, MilitaryAcademy) {
		t.Fatal("Military Academy did not graduate an Edu 5 character on a granted waiver")
	}

	if c.Score(Education) != 8 || !c.hasDegree("BA") {
		t.Errorf("waived academy result Edu=%d degrees=%v", c.Score(Education), c.Degrees)
	}

	// A declined waiver rejects admission outright, before any dice are rolled.
	declined := Character{scores: [count]int{7, 7, 7, 8, 5, 7}, Age: 18}
	if AttendInstitution(dice.NewScripted(1), tradeNoWaiver{}, &declined, MilitaryAcademy) {
		t.Fatal("Military Academy admitted an Edu 5 character without a waiver")
	}

	if declined.Age != 18 || len(declined.EducationHistory) != 0 {
		t.Errorf("declined waiver changed character: age=%d history=%+v", declined.Age, declined.EducationHistory)
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

func TestOTC(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 8, 8, 7}, Age: 22}
	if !AttendInstitution(dice.NewScripted(3, 3), DefaultPolicy{}, &c, OTC) {
		t.Fatal("OTC failed")
	}

	last := c.EducationHistory[len(c.EducationHistory)-1]
	if last.Institution != OTC || last.Commission != "Army" || !last.Graduated {
		t.Errorf("OTC status = %+v", last)
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
	for _, institution := range []Institution{ArmySchool, NavySchool, MarineSchool} {
		c := Character{scores: [count]int{7, 8, 8, 9, 9, 7}, Age: 30}
		if !AttendInstitution(dice.NewScripted(3, 3), DefaultPolicy{}, &c, institution) {
			t.Fatalf("%q failed", institution)
		}

		if c.Age != 31 || c.EducationHistory[len(c.EducationHistory)-1].Institution != institution {
			t.Errorf("%q status age=%d history=%+v", institution, c.Age, c.EducationHistory)
		}
	}

	c := Character{scores: [count]int{7, 8, 8, 9, 9, 7}, Age: 30}
	if !AttendInstitution(dice.NewScripted(3, 3, 3, 3), DefaultPolicy{}, &c, CommandCollege) {
		t.Fatal("Command College failed")
	}

	if c.Age != 31 || c.EducationHistory[len(c.EducationHistory)-1].Institution != CommandCollege {
		t.Errorf("Command College status age=%d history=%+v", c.Age, c.EducationHistory)
	}
}

func TestCommandCollegeUsesCharactersService(t *testing.T) {
	c := Character{
		scores:           [count]int{7, 8, 8, 9, 9, 7},
		Age:              30,
		EducationHistory: []EducationRecord{{Institution: NavalAcademy, Commission: "Navy", Graduated: true}},
	}

	if !AttendInstitution(dice.NewScripted(3, 3, 3, 3), DefaultPolicy{}, &c, CommandCollege) {
		t.Fatal("Command College failed")
	}

	// DefaultPolicy.ChooseSkill takes the least-held option each time, so a
	// Navy officer's two picks are deterministically the Naval Academy skill
	// list's first two entries — never a Military-Academy-only skill.
	want := AvailableSkills(NavalAcademy)
	if c.Skills.Level(want[0]) != 1 || c.Skills.Level(want[1]) != 1 {
		t.Errorf(
			"Command College did not grant the Naval Academy's own skills: %s=%d %s=%d",
			want[0], c.Skills.Level(want[0]), want[1], c.Skills.Level(want[1]),
		)
	}
}

func TestCommandAcademyUsesMostRecentCommission(t *testing.T) {
	cases := []struct {
		name    string
		history []EducationRecord
		want    Institution
	}{
		{"no history", nil, MilitaryAcademy},
		{"army commission", []EducationRecord{{Commission: "Army"}}, MilitaryAcademy},
		{"navy commission", []EducationRecord{{Commission: "Navy"}}, NavalAcademy},
		{"marine commission", []EducationRecord{{Commission: "Marine"}}, MarineAcademy},
		{
			"most recent wins",
			[]EducationRecord{{Commission: "Army"}, {Commission: "Navy"}},
			NavalAcademy,
		},
	}

	for _, tc := range cases {
		c := &Character{EducationHistory: tc.history}
		if got := commandAcademy(c); got != tc.want {
			t.Errorf("%s: commandAcademy() = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestOneShotInstitutionsDoNotStack(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 8, 8, 7}, Age: 20, Degrees: []string{"BA"}}

	if !AttendInstitution(dice.NewScripted(3, 3), DefaultPolicy{}, &c, HonorsProgram) {
		t.Fatal("Honors did not grant on first attendance")
	}

	degreesBefore := len(c.Degrees)
	if AttendInstitution(dice.NewScripted(1), DefaultPolicy{}, &c, HonorsProgram) {
		t.Fatal("Honors granted a second time for the same character")
	}

	if len(c.Degrees) != degreesBefore {
		t.Errorf("repeat Honors attendance changed Degrees: %v", c.Degrees)
	}
}

func TestHonorsFailureHasNoEffect(t *testing.T) {
	// Book 1 p.59: "Failure has no effect" — unlike every prerequisite, a
	// failed Honors check must not attempt a Waiver.
	c := Character{scores: [count]int{7, 7, 7, 1, 1, 7}, Age: 20, Degrees: []string{"BA"}}

	if AttendInstitution(dice.NewScripted(6, 6), DefaultPolicy{}, &c, HonorsProgram) {
		t.Fatal("Honors granted on a failed check")
	}

	if len(c.Degrees) != 1 || len(c.EducationHistory) != 0 {
		t.Errorf("failed Honors check changed character: degrees=%v history=%+v", c.Degrees, c.EducationHistory)
	}
}
