package chargen

// Entertainer career data, transcribed from Book 1 p. 77. The Entertainer's
// success is measured by Fame, not Risk & Reward: each term (see runFameTerm) a
// Flux roll shifts Fame, and a rise grants Talent +1 and two extra skills. There
// is no career injury. Initial Fame and Talent are one 2D roll (rolled at career
// start); Continue is vs the character's Fame — the more famous, the easier to
// keep performing.
//
// Slice scope: the six specialties (Artist/Actor/Author/Dancer/Musician/Chef)
// differ only in their Begin characteristics; this models the Actor (Begin vs
// Dex or End) and defers the others. Also deferred: the optional second/third
// Flux rolls per term, the Fame Descriptor table and Stage Name, Talent as a
// performance skill substitute, the Comeback, the muster row 13 (TAS Life, only
// reachable at very high Fame), and the muster Money-column +Fame/3 DM (treated
// as the standard +Terms).

// EntertainerCareer is the Entertainer (Book 1 p. 77), modelled as the Actor
// specialty.
var EntertainerCareer = Career{
	ID:               Entertainer,
	Name:             "Entertainer",
	Qualify:          Qualification{Chars: []Characteristic{Dexterity, Endurance}}, // Begin Actor: C2 or C3
	FameCareer:       true,
	Continue:         ContinueRule{UseFame: true}, // Continue vs Fame
	Advance:          RollLow,
	EligPerTerm:      4,
	MusterBenefitDMT: true, // Benefits column DM is +Terms
	Skills: SkillGrid{
		// Col 0 — Personal.
		{bump(Strength), bump(Dexterity), bump(Endurance), bump(Intelligence), bump(Education), bump(Social)},
		// Col 1 — Academic (Major/Minor lost without the education stage).
		{major, major, minor, minor, choose(theTrades...), choose(theTrades...)},
		// Col 2 — Travel.
		{sk("Zero-G"), sk("Vacc Suit"), sk("Pilot"), sk("Astrogation"), sk("Sensors"), choose(starshipSkls...)},
		// Col 3 — General.
		{sk("Survey"), sk("Survival"), sk("Hostile Environ"), sk("Animals"), sk("Bureaucrat"), sk("Navigation")},
		// Col 4 — Business.
		{sk("Broker"), sk("Trader"), sk("Advocate"), sk("Liaison"), sk("Diplomat"), sk("Bureaucrat")},
		// Col 5 — Vocation.
		{sk("Broker"), choose(oneArt...), cascade("Language", languages...), sk("Admin"), choose(oneArt...), sk("Bureaucrat")},
		// Col 6 — Avocation.
		{choose(oneArt...), choose(oneArt...), choose(theTrades...), sk("Athlete"), sk("Medic"), choose(theTrades...)},
	},
	// Muster-out (Book 1 p. 77), indexed 1-12 by (1D + DM). Money passages are
	// recorded at cash value. Row 13 (TAS Life) is deferred (reachable only at
	// very high Fame); row 12 stands.
	MusterOut: MusterTable{
		1:  {Money: cash(1_000), Benefit: charAward(Education)},  // Low Passage
		2:  {Money: cash(1_000), Benefit: charAward(Education)},  // Low Passage
		3:  {Money: cash(8_000), Benefit: named("Wafer Jack")},   // Mid Passage
		4:  {Money: cash(10_000), Benefit: charAward(Education)}, // High Passage
		5:  {Money: cash(15_000), Benefit: charAward(Strength)},
		6:  {Money: cash(250_000), Benefit: charAward(Dexterity)}, // StarPass
		7:  {Money: cash(25_000), Benefit: charAward(Endurance)},
		8:  {Money: cash(30_000), Benefit: charAward(Intelligence)},
		9:  {Money: cash(35_000), Benefit: named("Fame +1")},
		10: {Money: cash(40_000), Benefit: named("Ship Share")},
		11: {Money: cash(50_000), Benefit: named("TAS Fellowship")},
		12: {Money: cash(400_000), Benefit: named("Knighthood")},
	},
}
