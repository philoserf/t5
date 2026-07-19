package chargen

// Marine career data, transcribed from Book 1 p. 86. The Marine is a second
// armed-forces career, pure data on the rank engine: it differs from the Soldier
// only in which characteristics govern its rolls (Risk & Reward rotates Str/Int,
// Enlisted Promotion is vs Str, Officer Promotion vs Int, Continue vs Str), its
// rank titles, skill grid, and muster table.
//
// Slice scope matches the Soldier's: the
// +1 skill for a Commission or Promotion, branch automatic skills, the "Retire
// x2" muster pay (recorded as a named benefit) and its +Officer-Rank Benefit DM,
// and Command College are all deferred.

// marineBranchOps is the Marine's Branch and Operations tables (Book 1 p. 86).
var marineBranchOps = BranchOps{
	Branches: [9]Branch{
		1: {"Infantry", 1, 2},
		2: {"Infantry", 1, 2},
		3: {"Artillery", 1, 5},
		4: {"Cavalry", 1, 3},
		5: {"Protected", 2, 1},
		6: {"Commando", 2, 0},
		7: {"Technical", 0, 6},
		8: {"Medical", 0, 4},
	},
	// Operations (1D + Branch OpsDM): Combat, Peace Keeper, Mission, ANM
	// School, Combat, Peace Keeper, Mission, Garrison.
	OpsMods: [10]int{1: 2, 2: 2, 3: 1, 4: 2, 5: 0, 6: 3, 7: 1, 8: 2, 9: 0},
}

// MarineCareer is the Marine (Book 1 p. 86).
var MarineCareer = Career{
	ID:               Marine,
	Name:             "Marine",
	Qualify:          Qualification{Chars: []Characteristic{Strength}},
	CCMode:           RotateCC,
	ControllingChars: []Characteristic{Strength, Intelligence},    // C1 C4
	Continue:         ContinueRule{UseChar: true, Char: Strength}, // C1
	EligPerTerm:      4,
	BenefitDM:        DMOfficerRank,
	RewardKind:       RewardMedal,
	BranchOps:        &marineBranchOps,
	Commission:       PromotionRule{Char: Endurance},                           // C3
	EnlistedPromote:  PromotionRule{Char: Strength, MedalsAndWounds: true},     // C1*
	OfficerPromote:   PromotionRule{Char: Intelligence, MedalsAndWounds: true}, // C4*
	EnlistedRanks: []Rank{
		{Title: "Private", Skill: "Fighter"},
		{Title: "Lance Corporal"},
		{Title: "Sergeant", Skill: "Heavy Weapons"},
		{Title: "Staff Sergeant", Skill: "Tactics"},
		{Title: "Master Sergeant", Skill: "Leader"},
		{Title: "Sergeant Major"},
	},
	OfficerRanks: []Rank{
		{Title: "2nd Lieutenant", Skill: "Leader"},
		{Title: "1st Lieutenant"},
		{Title: "Captain"},
		{Title: "Force Commander", Skill: "Tactics"},
		{Title: "Lt Colonel"},
		{Title: "Colonel", Skill: "Leader"},
		{Title: "Brigadier"},
	},
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
		// Col 1 — Garrison (Major/Minor lost without the education stage).
		{choose(theTrades...), major, minor, sk("Gambler"), sk("Athlete"), choose(theTrades...)},
		// Col 2 — Combat.
		{
			sk("Fighter"),
			sk("Fighter"),
			choose(soldierSkls...),
			choose(soldierSkls...),
			sk("Leader"),
			sk("Tactics"),
		},
		// Col 3 — Peacekeeper.
		{
			sk("Vacc Suit"),
			sk("Fighter"),
			sk("Hostile Environ"),
			sk("Stealth"),
			sk("Leader"),
			sk("Tactics"),
		},
		// Col 4 — Occupation.
		{
			sk("Fighter"),
			sk("Fighter"),
			sk("Flyer"),
			sk("Stealth"),
			sk("Leader"),
			sk("Heavy Weapons"),
		},
		// Col 5 — Mission.
		{
			choose(soldierSkls...),
			sk("Survival"),
			cascade("Language", languages...),
			sk("Gunner"),
			sk("Leader"),
			sk("Fighter"),
		},
		// Col 6 — Technical.
		{
			choose(oneArt...),
			choose(sciences...),
			sk("Explosives"),
			sk("Medic"),
			sk("Seafarer"),
			choose(theTrades...),
		},
	},
	// Muster-out (Book 1 p. 86), indexed 1-12 by (1D + DM). Money passages are
	// recorded at cash value; "Retire x2" is a named retirement benefit. Rows 11
	// and 12 repeat row 10 (the printed table ends at 10; higher rolls clamp).
	MusterOut: MusterTable{
		1:  {Money: cash(1_000), Benefit: named("Forbidden Knowledge")}, // Low Passage
		2:  {Money: cash(8_000), Benefit: charAward(Strength)},          // Mid Passage
		3:  {Money: cash(10_000), Benefit: named("Wafer Jack")},         // High Passage
		4:  {Money: cash(250_000), Benefit: charAward(Education)},       // StarPass
		5:  {Money: cash(30_000), Benefit: charAward(Intelligence)},
		6:  {Money: cash(40_000), Benefit: charAward(Dexterity)},
		7:  {Money: cash(50_000), Benefit: named("Life Insurance")},
		8:  {Money: retirementX2(), Benefit: named("Ship Share")},
		9:  {Money: retirementX2(), Benefit: named("Directorate")},
		10: {Money: cash(60_000), Benefit: knighthood()},
		11: {Money: cash(60_000), Benefit: knighthood()},
		12: {Money: cash(60_000), Benefit: knighthood()},
	},
}
