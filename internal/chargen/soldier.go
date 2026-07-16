package chargen

// Soldier career data, transcribed from Book 1 p. 82. The Soldier is the first
// armed-forces career, exercising the rank engine: it begins enlisted (Private),
// may earn a Commission to the officer track, and promotes up the enlisted or
// officer ladder — promotion rolls sweetened by Medals (from Reward successes)
// and Wound Badges.
//
// The Branch and Operations modifiers to Risk & Reward (p. 82) are wired via
// soldierBranchOps and the engine's branchOpsMod. Deferred: the +1 skill for a
// Commission or Promotion, the branch-specific automatic skills (Medic / Trade),
// the muster "Retire x2" pay (recorded as a named benefit), and the Command
// College at O4.

// soldierBranchOps is the Soldier's Branch and Operations tables (Book 1 p. 82).
var soldierBranchOps = BranchOps{
	Branches: [9]Branch{
		1: {"Infantry", 1, 1},
		2: {"Infantry", 1, 1},
		3: {"Artillery", 1, 5},
		4: {"Cavalry", 1, 3},
		5: {"Protected", 2, 0},
		6: {"Protected", 2, 0},
		7: {"Technical", 0, 6},
		8: {"Medical", 0, 4},
	},
	// Operations (1D + Branch OpsDM): Combat, Combat, Peace Keeper, Mission, ANM
	// School, Combat, Peace Keeper, Mission, Base.
	OpsMods: [10]int{1: 2, 2: 2, 3: 1, 4: 2, 5: 0, 6: 3, 7: 1, 8: 2, 9: 0},
}

// SoldierCareer is the Soldier (Book 1 p. 82): an armed-forces career with
// enlisted and officer rank ladders.
var SoldierCareer = Career{
	ID:               Soldier,
	Name:             "Soldier",
	Qualify:          Qualification{Chars: []Characteristic{Strength}},
	CCMode:           RotateCC,
	ControllingChars: []Characteristic{Strength, Endurance, Intelligence}, // C1 C3 C4
	Continue:         ContinueRule{UseChar: true, Char: Endurance},        // C3
	Advance:          RollLow,
	EligPerTerm:      4,
	BenefitDM:        DMOfficerRank,
	RewardKind:       RewardMedal,
	BranchOps:        &soldierBranchOps,
	Commission:       PromotionRule{Char: Endurance},                        // C3
	EnlistedPromote:  PromotionRule{Char: Endurance, MedalsAndWounds: true}, // C3*
	OfficerPromote:   PromotionRule{Char: Social, MedalsAndWounds: true},    // Soc*
	EnlistedRanks: []Rank{
		{Title: "Private", Skill: "Fighter"},
		{Title: "Corporal"},
		{Title: "Sergeant", Skill: "Heavy Weapons"},
		{Title: "Staff Sergeant", Skill: "Leader"},
		{Title: "Master Sergeant"},
		{Title: "Sergeant Major"},
	},
	OfficerRanks: []Rank{
		{Title: "2nd Lieutenant", Skill: "Leader"},
		{Title: "1st Lieutenant"},
		{Title: "Captain"},
		{Title: "Major", Skill: "Tactics"},
		{Title: "Lt Colonel"},
		{Title: "Colonel", Skill: "Leader"},
		{Title: "General"},
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
		// Col 1 — Base (Major/Minor lost without the education stage).
		{major, major, minor, minor, choose(theTrades...), choose(theTrades...)},
		// Col 2 — Combat.
		{sk("Fighter"), sk("Vacc Suit"), sk("Fighter"), sk("Stealth"), sk("Leader"), sk("Tactics")},
		// Col 3 — Peacekeeper.
		{
			sk("Admin"),
			sk("Fighter"),
			sk("Hostile Environ"),
			sk("Animals"),
			sk("Liaison"),
			sk("Navigation"),
		},
		// Col 4 — Occupation.
		{
			sk("Fighter"),
			sk("Vacc Suit"),
			sk("Driver"),
			sk("Stealth"),
			sk("Heavy Weapons"),
			sk("Sensors"),
		},
		// Col 5 — Mission.
		{
			choose(soldierSkls...),
			sk("Liaison"),
			cascade("Language", languages...),
			choose(soldierSkls...),
			sk("Computer"),
			sk("Tactics"),
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
	// Muster-out (Book 1 p. 82), indexed 1-12 by (1D + DM). Money passages are
	// recorded at cash value; "Retire x2" is a named retirement benefit. Rows 11
	// and 12 repeat row 10 (the printed table ends at 10; higher rolls clamp).
	MusterOut: MusterTable{
		1:  {Money: cash(1_000), Benefit: named("Forbidden Knowledge")}, // Low Passage
		2:  {Money: cash(8_000), Benefit: charAward(Strength)},          // Mid Passage
		3:  {Money: cash(10_000), Benefit: named("Wafer Jack")},         // High Passage
		4:  {Money: cash(250_000), Benefit: charAward(Education)},       // StarPass
		5:  {Money: cash(30_000), Benefit: charAward(Intelligence)},
		6:  {Money: cash(40_000), Benefit: charAward(Dexterity)},
		7:  {Money: cash(50_000), Benefit: charAward(Endurance)},
		8:  {Money: retirementX2(), Benefit: named("Life Insurance")},
		9:  {Money: retirementX2(), Benefit: named("TAS Fellowship")},
		10: {Money: cash(60_000), Benefit: knighthood()},
		11: {Money: cash(60_000), Benefit: knighthood()},
		12: {Money: cash(60_000), Benefit: knighthood()},
	},
}
