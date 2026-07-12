package chargen

// Rogue career data, transcribed from Book 1 p. 84.
//
// Slice scope: the Rogue's defining Scheme mechanic (each term the Rogue
// masterminds a Scheme rolled from the Scheme table for a payoff, is imprisoned
// on a failed Risk, and gains Infamy) is deferred, along with its consequences —
// the Scheme-driven skill eligibility (Failed Scheme 1 / Successful Scheme 4 / In
// Prison 2, over the base Per-Term 2), the +Terms modifier on Risk & Reward and
// Continue, and the "12 is always failure" rule. The Rogue here runs as a
// fixed-CC career with the standard Risk & Reward, the p. 84 skill grid, base
// Per-Term 2 eligibility, and the p. 84 mustering-out table (whose Benefit
// column, unusually, also gets the +Terms DM).

// allChars is every characteristic — the Rogue selects one as the Controlling
// Characteristic used throughout the career (Book 1 p. 84).
var allChars = []Characteristic{Strength, Dexterity, Endurance, Intelligence, Education, Social}

// soldierSkls is a representative "Soldier Skill" choice list (the full cascade
// is deferred), mirroring how starshipSkls stands in for "Starship Skill".
var soldierSkls = []string{"Fighter", "Gunner", "Heavy Weapons", "Tactics"}

// RogueCareer is the Rogue (Book 1 p. 84): a fixed-CC career. To Begin, Risk &
// Reward, and Continue all use the one selected Controlling Characteristic.
var RogueCareer = Career{
	ID:               Rogue,
	Name:             "Rogue",
	Qualify:          Qualification{Chars: allChars},
	CCMode:           FixedCC,
	ControllingChars: allChars,
	Continue:         ContinueRule{UseCC: true},
	Advance:          RollLow,
	EligPerTerm:      2,
	BenefitDM:        DMTerms,
	Skills: SkillGrid{
		// Col 0 — Personal.
		{bump(Strength), bump(Dexterity), bump(Endurance), bump(Intelligence), bump(Education), bump(Social)},
		// Col 1 — Academic (Major/Minor lost without the education stage).
		{choose(sciences...), major, minor, choose(oneArt...), choose(theTrades...), sk("Gambler")},
		// Col 2 — World Travel.
		{sk("Driver"), sk("Flyer"), sk("Hostile Environ"), sk("High-G"), sk("Vacc Suit"), sk("Navigation")},
		// Col 3 — Space Travel ("Astrogator" on p. 84 is the Astrogation skill).
		{choose(starshipSkls...), sk("Pilot"), sk("Engineer"), sk("Zero-G"), sk("Vacc Suit"), sk("Astrogation")},
		// Col 4 — Business.
		{sk("Trader"), sk("Broker"), sk("Computer"), sk("JOT"), sk("Teacher"), sk("Fighter")},
		// Col 5 — Vocation.
		{sk("Advocate"), sk("Counsellor"), cascade("Language", languages...), sk("Leader"), sk("Streetwise"), sk("Comms")},
		// Col 6 — Avocation.
		{choose(oneArt...), choose(sciences...), sk("Athlete"), choose(soldierSkls...), choose(starshipSkls...), choose(theTrades...)},
	},
	// Muster-out (Book 1 p. 84), indexed 1-12 by (1D + DM). Money-column passages
	// are recorded at their standard cash value.
	MusterOut: MusterTable{
		1:  {Money: cash(40_000), Benefit: charAward(Strength)},
		2:  {Money: cash(250_000), Benefit: charAward(Education)},    // StarPass
		3:  {Money: cash(250_000), Benefit: named("Wafer Jack")},     // StarPass
		4:  {Money: cash(10_000), Benefit: charAward(Dexterity)},     // High Passage
		5:  {Money: cash(10_000), Benefit: charAward(Endurance)},     // High Passage
		6:  {Money: cash(250_000), Benefit: named("Life Insurance")}, // StarPass
		7:  {Money: cash(25_000), Benefit: named("Ship Share")},
		8:  {Money: cash(30_000), Benefit: named("Knighthood")},
		9:  {Money: cash(35_000), Benefit: named("Ship Share")},
		10: {Money: cash(40_000), Benefit: named("Ship Share")},
		11: {Money: cash(50_000), Benefit: named("Ship Share")},
		12: {Money: cash(90_000), Benefit: named("Knighthood")},
	},
}
