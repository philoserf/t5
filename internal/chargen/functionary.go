package chargen

// Functionary career data, transcribed from Book 1 p. 87. The Functionary plays
// Office Politics (see runPoliticsTerm) instead of Risk & Reward: two unmodified
// rolls against the Controlling Characteristic — a failed Risk ends the career
// (job loss, no injury) and a successful Reward is a promotion up the F0-F8 rank
// ladder (Clerk … Secretary).
//
// Slice scope: the "Total Terms x3" begin (Functionary is never a first career)
// is deferred, so the career AutoBegins; the F7 UnderSecretary's 1D ordinal, the
// prior-career-specific F6 titles (College President, etc.), the +1 skill per
// promotion, and the +Officer-Rank muster Money DM (treated as +Terms) are also
// deferred. The "Pension x2" muster pay doubles the Functionary's Cr15,000/year
// pension (an entitlement; see entitlement.go).

// FunctionaryCareer is the Functionary (Book 1 p. 87).
var FunctionaryCareer = Career{
	ID:               Functionary,
	Name:             "Functionary",
	AutoBegin:        true,
	OfficePolitics:   true,
	ControllingChars: []Characteristic{Dexterity, Endurance, Intelligence, Education}, // C2 C3 C4 C5
	Continue:         ContinueRule{Fixed: 15},                                         // Office Politics decides continuation; the policy chooses
	Advance:          RollLow,
	EligPerTerm:      4,
	BenefitDM:        DMRank,
	// The F0-F8 rank ladder (Book 1 p. 87); rank 1 is Clerk (F0).
	EnlistedRanks: []Rank{
		{Title: "Clerk", Skill: "Bureaucrat"},
		{Title: "Supervisor"},
		{Title: "Senior Supervisor", Skill: "Admin"},
		{Title: "Manager", Skill: "Bureaucrat"},
		{Title: "Senior Manager"},
		{Title: "Assistant Director"},
		{Title: "Director"},
		{Title: "UnderSecretary"},
		{Title: "Secretary"},
	},
	Skills: SkillGrid{
		// Col 0 — Personal.
		{bump(Strength), bump(Dexterity), bump(Endurance), bump(Intelligence), bump(Education), bump(Social)},
		// Col 1 — Academic (Major/Minor lost without the education stage).
		{major, major, minor, minor, choose(theTrades...), choose(theTrades...)},
		// Col 2 — World Travel.
		{sk("High-G"), sk("Vacc Suit"), sk("Driver"), sk("Flyer"), sk("Navigation"), sk("Seafarer")},
		// Col 3 — General ("Any Skill" is a representative Citizen-job choice).
		{choose(theTrades...), choose(oneArt...), choose(sciences...), choose(citizenJobs...), sk("Bureaucrat"), sk("Leader")},
		// Col 4 — Business.
		{sk("Advocate"), sk("Broker"), sk("Trader"), sk("Teacher"), choose(theTrades...), sk("Driver")},
		// Col 5 — Vocation.
		{sk("Advocate"), sk("Comms"), cascade("Language", languages...), sk("Admin"), sk("Bureaucrat"), sk("Comms")},
		// Col 6 — Avocation.
		{choose(oneArt...), choose(sciences...), sk("Athlete"), sk("Designer"), sk("Seafarer"), choose(theTrades...)},
	},
	// Muster-out, indexed 1-12 by (1D + DM). The career page (p. 87) prints 11
	// rows; row 12 (Pension x2 | Knighthood) comes from the consolidated
	// Muster-Out Tables. "Pension x2" is a named benefit.
	MusterOut: MusterTable{
		1:  {Money: cash(5_000), Benefit: named("Forbidden Knowledge")},
		2:  {Money: cash(10_000), Benefit: charAward(Strength)},
		3:  {Money: cash(15_000), Benefit: named("Wafer Jack")},
		4:  {Money: cash(20_000), Benefit: charAward(Strength)},
		5:  {Money: cash(10_000), Benefit: charAward(Dexterity)},  // High Passage
		6:  {Money: cash(250_000), Benefit: charAward(Endurance)}, // StarPass
		7:  {Money: cash(25_000), Benefit: charAward(Intelligence)},
		8:  {Money: cash(30_000), Benefit: named("Life Insurance")},
		9:  {Money: cash(35_000), Benefit: named("TAS Fellowship")},
		10: {Money: pensionX2(), Benefit: knighthood()},
		11: {Money: pensionX2(), Benefit: named("Directorship")},
		12: {Money: pensionX2(), Benefit: knighthood()},
	},
}
