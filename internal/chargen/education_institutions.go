package chargen

import (
	"fmt"
	"slices"

	"github.com/philoserf/t5/internal/dice"
)

// Institution identifies a column in the Available Skills matrix or one of the
// educational programs on Book 1 p.60.
type Institution byte

const (
	// CollegeInstitution is the College/University/Masters/Professors matrix column.
	CollegeInstitution Institution = 'C'
	// LawSchool is the Law School matrix column and professional program.
	LawSchool Institution = 'L'
	// MedicalSchool is the Medical School matrix column and professional program.
	MedicalSchool Institution = 'M'
	// MilitaryAcademy is the Army/Marine academy matrix column and program.
	MilitaryAcademy Institution = 'A'
	// NavalAcademy is the Naval academy matrix column and program.
	NavalAcademy Institution = 'N'
	// MarineAcademy is the Marine academy matrix column. The printed table calls
	// both Medical School and Marine Academy "M"; the API keeps them distinct.
	MarineAcademy Institution = 'R'
	// School is the general, trade, and service-school matrix column.
	School Institution = 'S'
	// HonorsProgram is the simultaneous optional Honors check.
	HonorsProgram Institution = 'H'
	// OTC is Army Officer Training Corps.
	OTC Institution = 'O'
	// NOTC is Naval Officer Training Corps.
	NOTC Institution = 'T'
	// FlightSchool is the post-Honors flight program.
	FlightSchool Institution = 'F'
	// CommandCollege is assigned senior-officer education.
	CommandCollege Institution = 'D'
	// ArmySchool is an assigned Army ANM school.
	ArmySchool Institution = 'a'
	// NavySchool is an assigned Navy ANM school.
	NavySchool Institution = 'n'
	// MarineSchool is an assigned Marine ANM school.
	MarineSchool Institution = 'm'
)

// EducationRecord records completion of an institution and its formal status.
// Commission is "Army", "Navy", or "Marine" when the course confers Officer1;
// Flight is true when Flight School assigns the Flight branch.
type EducationRecord struct {
	Institution Institution
	Graduated   bool
	Honors      bool
	Commission  string
	Flight      bool
}

// availableSkillCodes is the complete p.60 seven-column matrix, transcribed by skill. A
// matrix entry means that institution may teach that skill; bold/knowledge-only
// typography does not change availability and remains a property of skill data.
// Internal matrix codes deliberately use X for medical and R for marine because
// the printed legend labels both columns M. This avoids silently merging them.
var availableSkillCodes = map[string]string{
	"Admin": "S", "Advocate": "L", "Animals": "", "Athlete": "C", "Broker": "C",
	"Bureaucrat": "C", "Comms": "S", "Computer": "S", "Counsellor": "C", "Designer": "C",
	"Diplomat": "R", "Driver": "", "Explosives": "S", "Fleet Tactics": "N", "Flyer": "",
	"Forensics": "X", "Gambler": "", "High-G": "S", "Hostile Env": "S", "JOT": "",
	"Language": "CNRS", "Leader": "NR", "Liaison": "NR", "Naval Arch": "N", "Seafarer": "",
	"Stealth": "", "Strategy": "NR", "Streetwise": "", "Survey": "S", "Survival": "S",
	"Tactics": "NRS", "Teacher": "C", "Trader": "S", "Vacc Suit": "S", "Zero-G": "S",
	"Astrogator": "CN", "Engineer": "", "Gunner": "N", "Medic": "XNRS", "Pilot": "",
	"Sensors": "NS", "Steward": "NS", "Fighter": "R", "Fwd Obs": "RS", "Hvy Wpns": "RS",
	"Navigation": "RS", "Recon": "RS", "Sapper": "RS", "Actor": "CS", "Artist": "CS",
	"Author": "CS", "Chef": "CS", "Dancer": "CS", "Musician": "CS", "Biologics": "CS",
	"Craftsman": "CS", "Electronics": "CS", "Fluidics": "CS", "Gravitics": "CS",
	"Magnetics": "CS", "Mechanic": "CS", "Photonics": "CS", "Polymers": "CS", "Program": "CS",
	"ACV": "AS", "Automotive": "CAS", "Grav Driver": "ANRS", "Legged Driver": "AS",
	"Mole Driver": "AS", "Tracked Driver": "ARS", "Wheeled Driver": "ANRS",
	"Aeronautics": "CANS", "Flapper": "AS", "Grav Flyer": "ANRS", "LTA": "AS",
	"Rotor": "AS", "Winged": "ANS", "Battle Dress": "ANR", "Beams": "AR",
	"Blades": "A", "Exotics": "AR", "Slug Throw": "ANRS", "Sprays": "AR", "Unarmed": "ARS",
	"Bay Wpns": "N", "Ortillery": "N", "Screens": "AN", "Spines": "N", "Turrets": "NR",
	"Small Craft": "ARS", "Pilot-ACS": "N", "Pilot-BCS": "N", "J-Drives": "NS",
	"Life Support": "ANS", "M-Drive": "NS", "P-Systems": "ANRS", "Rider": "AS",
	"Teamster": "AS", "Trainer": "ANRS", "Archeology": "C", "Biology": "C",
	"Chemistry": "C", "History": "C", "Linguistics": "CS", "Philosophy": "C",
	"Physics": "C", "Planetology": "C", "Psionicology": "C", "Psychohistory": "C",
	"Psychology": "C", "Robotics": "CANRS", "Sophontology": "C", "Aquanautics": "CS",
	"Grav Seafarer": "ANRS", "Boat": "RS", "Ship": "RS", "Sub": "RS",
	"Artillery": "AR", "Launcher": "AR", "Ordnance": "AR", "WMD": "ANR",
}

// AvailableSkills returns the p.60 skills available at an institution, in stable
// alphabetical order. Masters and Professors use CollegeInstitution; Army,
// Naval, Marine, Trade, apprentice, and training schools use School together
// with the appropriate academy column where the chart marks both.
func AvailableSkills(institution Institution) []string {
	column := matrixColumn(institution)
	result := make([]string, 0)

	for name, codes := range availableSkillCodes {
		if slices.Contains([]byte(codes), column) {
			result = append(result, name)
		}
	}

	slices.Sort(result)

	return result
}

func matrixColumn(institution Institution) byte {
	switch institution {
	case MedicalSchool:
		return 'X'
	case MarineAcademy:
		return 'R'
	default:
		return byte(institution)
	}
}

// AttendInstitution runs one explicitly selected higher, professional, or
// military institution. Selection is policy, not dice: callers decide which
// institution and (for academies) which service by choosing A or N. It returns
// false when prerequisites/admission/completion fail. Existing educate retains
// its established undergraduate/graduate dice order.
func AttendInstitution(r *dice.Roller, p Policy, c *Character, institution Institution) bool {
	switch institution {
	case MilitaryAcademy:
		return attendServiceAcademy(r, p, c, "Army")
	case NavalAcademy:
		return attendServiceAcademy(r, p, c, "Navy")
	case MarineAcademy:
		return attendServiceAcademy(r, p, c, "Marine")
	case MedicalSchool:
		return attendProfessional(r, p, c, institution, 4, "Medic", 4, "Doctor")
	case LawSchool:
		return attendProfessional(r, p, c, institution, 2, "Advocate", 2, "Attorney")
	case HonorsProgram:
		return attendHonors(r, p, c)
	case OTC:
		return attendOfficerTraining(r, p, c, institution, "Army", soldierEducationSkills)
	case NOTC:
		return attendOfficerTraining(r, p, c, institution, "Navy", shipEducationSkills)
	case FlightSchool:
		return attendFlightSchool(r, p, c)
	case CommandCollege:
		return attendCommandCollege(r, p, c)
	case ArmySchool:
		return attendANMSchool(r, p, c, MilitaryAcademy)
	case NavySchool:
		return attendANMSchool(r, p, c, NavalAcademy)
	case MarineSchool:
		return attendANMSchool(r, p, c, MarineAcademy)
	default:
		panic(fmt.Sprintf("chargen: institution %q has no standalone attendance process", institution))
	}
}

func attendANMSchool(r *dice.Roller, p Policy, c *Character, serviceColumn Institution) bool {
	c.Age++
	if !c.Check(r, bestChar(*c, Dexterity, Endurance), 2, 0).Success {
		return false
	}

	options := make([]string, 0)

	for name, codes := range availableSkillCodes {
		if slices.Contains([]byte(codes), matrixColumn(School)) &&
			slices.Contains([]byte(codes), matrixColumn(serviceColumn)) {
			options = append(options, name)
		}
	}

	slices.Sort(options)

	c.Skills.Raise(p.ChooseSkill(*c, options), 2)
	c.EducationHistory = append(
		c.EducationHistory,
		EducationRecord{Institution: serviceSchool(serviceColumn), Graduated: true},
	)

	return true
}

func serviceSchool(serviceColumn Institution) Institution {
	switch serviceColumn {
	case MilitaryAcademy:
		return ArmySchool
	case NavalAcademy:
		return NavySchool
	default:
		return MarineSchool
	}
}

var (
	shipEducationSkills    = []string{"Astrogator", "Engineer", "Gunner", "Medic", "Pilot", "Sensors", "Steward"}
	soldierEducationSkills = []string{"Fighter", "Fwd Obs", "Hvy Wpns", "Navigation", "Recon", "Sapper"}
)

func attendServiceAcademy(r *dice.Roller, p Policy, c *Character, service string) bool {
	waivers := 0
	if c.Score(Education) < 6 && !waiverGranted(r, p, c, &waivers) {
		return false
	}

	prog := academicProgram{
		name: service + " Service Academy", years: 4, awardsMajor: true,
		awardsMinor: true, gradEdu: 8, degree: "BA", available: AvailableSkills(academyInstitution(service)),
	}
	before := len(c.Degrees)
	attendAcademicWithWaivers(r, p, c, prog, waivers)

	graduated := len(c.Degrees) > before
	if graduated {
		c.EducationHistory = append(
			c.EducationHistory,
			EducationRecord{Institution: academyInstitution(service), Graduated: true, Commission: service},
		)
	}

	return graduated
}

func academyInstitution(service string) Institution {
	switch service {
	case "Navy":
		return NavalAcademy
	case "Marine":
		return MarineAcademy
	default:
		return MilitaryAcademy
	}
}

func attendProfessional(
	r *dice.Roller,
	p Policy,
	c *Character,
	institution Institution,
	years int,
	skill string,
	level int,
	degree string,
) bool {
	waivers := 0
	if !c.hasDegree("Honors BA") && !waiverGranted(r, p, c, &waivers) {
		return false
	}

	ch := bestChar(*c, Intelligence, Education)
	if !admitted(r, p, c, ch, &waivers) {
		return false
	}

	for range years {
		c.Age++
		if !c.Check(r, ch, 2, 0).Success && !waiverGranted(r, p, c, &waivers) {
			return false
		}
	}

	c.Skills.Raise(skill, level)

	if c.scores[Education] < 10 {
		c.scores[Education] = 10
	} else {
		c.scores[Education] = min(c.scores[Education]+1, maxCharacteristic)
	}

	c.Degrees = append(c.Degrees, degree)
	c.EducationHistory = append(c.EducationHistory, EducationRecord{Institution: institution, Graduated: true})

	return true
}

func attendHonors(r *dice.Roller, p Policy, c *Character) bool {
	if !c.hasDegree("BA") {
		return false
	}

	waivers := 0

	ch := bestChar(*c, Intelligence, Education)
	if !c.Check(r, ch, 2, 0).Success && !waiverGranted(r, p, c, &waivers) {
		return false
	}

	if c.Major != "" {
		c.Skills.Raise(c.Major, 1)
	}

	c.Degrees = append(c.Degrees, "Honors BA")
	c.EducationHistory = append(
		c.EducationHistory,
		EducationRecord{Institution: HonorsProgram, Graduated: true, Honors: true},
	)

	return true
}

func attendOfficerTraining(
	r *dice.Roller,
	p Policy,
	c *Character,
	institution Institution,
	service string,
	skills []string,
) bool {
	ch := bestChar(*c, Intelligence, Education)
	if !c.Check(r, ch, 2, 0).Success {
		return false
	}

	chosen := p.ChooseSkill(*c, skills)
	c.Skills.Raise(chosen, 1)
	c.EducationHistory = append(
		c.EducationHistory,
		EducationRecord{Institution: institution, Graduated: true, Commission: service},
	)

	return true
}

func attendFlightSchool(r *dice.Roller, p Policy, c *Character) bool {
	waivers := 0
	if !c.hasDegree("Honors BA") && !waiverGranted(r, p, c, &waivers) {
		return false
	}

	c.Age++
	if !c.Check(r, Dexterity, 2, 0).Success {
		return false
	}

	c.Skills.Raise("Pilot", 3)
	c.EducationHistory = append(
		c.EducationHistory,
		EducationRecord{Institution: FlightSchool, Graduated: true, Flight: true},
	)

	return true
}

func attendCommandCollege(r *dice.Roller, p Policy, c *Character) bool {
	waivers := 0
	if !admitted(r, p, c, bestChar(*c, Intelligence, Education), &waivers) {
		return false
	}

	c.Age++
	if !c.Check(r, bestChar(*c, Intelligence, Education), 2, 0).Success {
		return false
	}

	options := append(AvailableSkills(MilitaryAcademy), AvailableSkills(NavalAcademy)...)
	for range 2 {
		c.Skills.Raise(p.ChooseSkill(*c, options), 1)
	}

	c.EducationHistory = append(c.EducationHistory, EducationRecord{Institution: CommandCollege, Graduated: true})

	return true
}
