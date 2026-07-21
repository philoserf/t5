package chargen

// Citizen career data, transcribed from Book 1 p. 78. The Citizen is entered
// automatically (To Begin = Auto) and replaces Risk & Reward with the benign
// Citizen Life roll (see runCitizenTerm): a success grants a Job or Hobby skill
// on a fixed schedule, a failure nothing — there is no career injury. It still
// gains the standard per-term grid skills, and Continues on a fixed target of
// 10.
//
// Slice scope: the Job/Hobby skill is chosen from a representative list
// (citizenJobs) rather than rolled on the book's full Citizen Skills-and-
// Knowledges table (the big A/B/C table on p. 78), and the Edu-select option for
// a Job/Hobby in the character's Major/Minor is deferred.

// CitizenCareer is the Citizen (Book 1 p. 78).
var CitizenCareer = Career{
	ID:               Citizen,
	Name:             "Citizen",
	AutoBegin:        true,
	Term:             CitizenTerm,
	ControllingChars: []Characteristic{Strength, Dexterity, Endurance, Intelligence}, // C1 C2 C3 C4
	Continue:         ContinueRule{Fixed: 10},                                        // 10-
	EligPerTerm:      4,
	BenefitDM:        DMTerms,
	Skills: SkillGrid{
		// Col 0 — Personal.
		{
			bump(Strength),
			bump(Dexterity),
			bump(Endurance),
			bump(Intelligence),
			bump(Education),
			bump(Social),
		},
		// Col 1 — Academic (Major/Minor lost without the education stage).
		{major, major, minor, minor, choose(theTrades...), choose(theTrades...)},
		// Col 2 — Travel.
		{
			sk("Seafarer"),
			sk("Navigation"),
			sk("Hostile Environ"),
			sk("Flyer"),
			sk("Driver"),
			sk("Vacc Suit"),
		},
		// Col 3 — General.
		{sk("Admin"), sk("Broker"), sk("Computer"), sk("Animals"), sk("Bureaucrat"), sk("Trader")},
		// Col 4 — Business.
		{
			sk("Advocate"),
			sk("Broker"),
			sk("Trader"),
			sk("Liaison"),
			sk("Counsellor"),
			sk("Teacher"),
		},
		// Col 5 — Vocation.
		{
			choose(oneArt...),
			choose(sciences...),
			choose(theTrades...),
			sk("Driver"),
			sk("Bureaucrat"),
			sk("Computer"),
		},
		// Col 6 — Avocation.
		{
			choose(oneArt...),
			choose(sciences...),
			sk("JOT"),
			sk("Athlete"),
			sk("Medic"),
			choose(theTrades...),
		},
	},
	// Muster-out (Book 1 p. 78), indexed 1-12 by (1D + DM). Money passages are
	// recorded at cash value. Row 12 repeats row 11 (the printed table ends at 11).
	MusterOut: MusterTable{
		1:  {Money: cash(1_000), Benefit: charAward(Strength)},   // Low Passage
		2:  {Money: cash(1_000), Benefit: charAward(Strength)},   // Low Passage
		3:  {Money: cash(8_000), Benefit: named("Wafer Jack")},   // Mid Passage
		4:  {Money: cash(10_000), Benefit: charAward(Education)}, // High Passage
		5:  {Money: cash(15_000), Benefit: charAward(Strength)},
		6:  {Money: cash(250_000), Benefit: charAward(Dexterity)}, // StarPass
		7:  {Money: cash(25_000), Benefit: charAward(Endurance)},
		8:  {Money: cash(30_000), Benefit: charAward(Intelligence)},
		9:  {Money: cash(35_000), Benefit: charAward(Social)},
		10: {Money: cash(40_000), Benefit: named("TAS Fellowship")},
		11: {Money: cash(50_000), Benefit: named("Ship Share")},
		12: {Money: cash(50_000), Benefit: named("Ship Share")},
	},
}
