package chargen

// Agent career data, transcribed from Book 1 p.83. The Agent is a rankless
// career (no rank ladders) with a standard Risk & Reward and a Continue roll
// that eases with experience ("Str, Mod +Terms").
//
// Each term the Agent runs an Undercover Assignment (Term UndercoverTerm,
// awardUndercover): two years undercover in a rolled career, selecting one
// skill from that career's tables, then two completing the Mission. Skill
// eligibility is Per Term 2 + Undercover 1 + Successful Mission 4 (a held Risk).
// A successful Reward is a Commendation (RewardCommendation), and the muster
// Benefit column's DM is +Commendations (DMCommends). "Any Knowledge" in the
// Vocation column is a representative academic-subject choice; the specific
// C-column skill of the Undercover table is overridden by the Agent's selection.

// AgentCareer is the Agent (Book 1 p.83).
var AgentCareer = Career{
	ID:               Agent,
	Name:             "Agent",
	Qualify:          Qualification{Chars: []Characteristic{Endurance}}, // C3
	CCMode:           RotatingCC,
	ControllingChars: []Characteristic{Strength, Dexterity, Endurance, Intelligence}, // C1 C2 C3 C4
	Continue: ContinueRule{
		UseChar:  true,
		Char:     Strength,
		TermsMod: true,
	}, // Str, Mod +Terms
	EligPerTerm: 2, // Per Term 2 (Undercover 1 and Successful Mission 4 are added in awardUndercover)
	Term:        UndercoverTerm,
	RewardKind:  RewardCommendation,
	BenefitDM:   DMCommends,
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
			sk("Zero-G"),
			sk("Vacc Suit"),
			sk("Pilot"),
			choose(starshipSkls...),
			sk("Gunner"),
			sk("Sensors"),
		},
		// Col 3 — Mission.
		{
			sk("Survey"),
			sk("Survival"),
			sk("Hostile Environ"),
			sk("Animals"),
			sk("Bureaucrat"),
			sk("Navigation"),
		},
		// Col 4 — Conflict.
		{
			sk("Fighter"),
			choose(soldierSkls...),
			sk("Flyer"),
			sk("Stealth"),
			sk("Gunner"),
			sk("Streetwise"),
		},
		// Col 5 — Vocation ("Any Knowledge" is a representative academic choice).
		{
			choose(academicMajors...),
			sk("Admin"),
			cascade("Language", languages...),
			choose(starshipSkls...),
			sk("Forensics"),
			sk("Comms"),
		},
		// Col 6 — Avocation.
		{
			choose(oneArt...),
			choose(sciences...),
			sk("Athlete"),
			sk("Medic"),
			sk("Seafarer"),
			choose(theTrades...),
		},
	},
	// Muster-out (Book 1 p.83), indexed 1-12 by (1D + DM). Money passages are
	// recorded at cash value.
	MusterOut: MusterTable{ //nolint:dupl // per-career muster-out table; parallel by design
		1:  {Money: cash(1_000), Benefit: named("Ship Share")},          // Low Passage
		2:  {Money: cash(1_000), Benefit: named("Forbidden Knowledge")}, // Low Passage
		3:  {Money: cash(8_000), Benefit: named("Wafer Jack")},          // Mid Passage
		4:  {Money: cash(10_000), Benefit: charAward(Education)},        // High Passage
		5:  {Money: cash(15_000), Benefit: charAward(Strength)},
		6:  {Money: cash(250_000), Benefit: charAward(Dexterity)}, // StarPass
		7:  {Money: cash(25_000), Benefit: charAward(Endurance)},
		8:  {Money: cash(30_000), Benefit: named("Ship Share")},
		9:  {Money: cash(35_000), Benefit: named("Life Insurance")},
		10: {Money: cash(40_000), Benefit: named("TAS Fellowship")},
		11: {Money: cash(80_000), Benefit: fameAwardN(2)},
		12: {Money: cash(90_000), Benefit: knighthood()},
	},
}
