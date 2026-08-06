package chargen

// Spacer career data, transcribed from Book 1 p.81. The Spacer (Navy) is the
// third armed-forces career and, like the Marine, pure data on the rank engine.
// Its enlisted track is the naval Rating ladder, so its "Rating Promotion" is
// the engine's EnlistedPromote. It differs from the Soldier/Marine in data: it
// begins vs Int, Risk & Reward rotates Str/Dex/Int, Commission and Rating
// Promotion are vs Dex, Officer Promotion vs Soc, Continue vs Str, plus its own
// naval rank titles, skill grid, and muster table.
//
// The Branch and Operations modifiers to Risk & Reward are wired via
// spacerBranchOps and the engine's branchOpsMod.
//
// The page's "Enlisted may select a new Branch upon Promotion" is the Spacer's
// shorthand for the general rule at p.66, which the engine follows instead: a
// non-officer may change Branch at the end of every term, and a Commission
// re-reads the Branch from the Officer column, where the Spacer's Crew becomes
// Line. That is what makes the Officer column below reachable — see BranchOps.
//
// Its remaining deferrals match the Soldier's: branch automatic skills, the
// restriction on which skills a Commission or Promotion may buy, and Command
// College.

// spacerBranchOps is the Spacer's Naval Branch and Operations tables (Book 1
// p.81). Naval Operations use no per-branch DM, so every branch's OpsDM is 0.
//
// NAVAL BRANCH is the one Branch table in the book that prints two columns —
// "1D Officer Mod Enlisted Mod" — and they disagree on four rolls: 1 and 2 are
// Line for an officer but Crew for the enlisted; 3 is Line 1 for an officer but
// Engineer 0 for the enlisted; 6 is Flight 2 for an officer but Gunnery 1 for
// the enlisted. Rolls 4, 5, 7, and 8 agree. Every Spacer enters the career
// enlisted, so the Enlisted column is the one that applies at the only point the
// engine selects a Branch; both are transcribed so the lookup can read the
// column matching the character's status (BranchOps.branchFor).
var spacerBranchOps = BranchOps{
	// The Officer column.
	Branches: [9]Branch{
		1: {"Line", 1, 0},
		2: {"Line", 1, 0},
		3: {"Line", 1, 0},
		4: {"Engineer", 0, 0},
		5: {"Gunnery", 1, 0},
		6: {"Flight", 2, 0},
		7: {"Technical", 0, 0},
		8: {"Medical", 0, 0},
	},
	// The Enlisted column.
	EnlistedBranches: &[9]Branch{
		1: {"Crew", 1, 0},
		2: {"Crew", 1, 0},
		3: {"Engineer", 0, 0},
		4: {"Engineer", 0, 0},
		5: {"Gunnery", 1, 0},
		6: {"Gunnery", 1, 0},
		7: {"Technical", 0, 0},
		8: {"Medical", 0, 0},
	},
	// Naval Operations (1D): Battle, Strike, Siege, Patrol, Mission, ANM School,
	// Shore Duty, Shore Duty.
	OpsMods: [10]int{1: 2, 2: 2, 3: 0, 4: 1, 5: 3, 6: 0, 7: 0, 8: 0},
}

// SpacerCareer is the Spacer (Book 1 p.81).
var SpacerCareer = Career{
	ID:               Spacer,
	Name:             "Spacer",
	Qualify:          Qualification{Chars: []Characteristic{Intelligence}},
	CCMode:           RotatingCC,
	ControllingChars: []Characteristic{Strength, Dexterity, Intelligence}, // C1 C2 C4
	Continue:         ContinueRule{UseChar: true, Char: Strength},         // C1
	EligPerTerm:      4,
	BenefitDM:        DMOfficerRank,
	RewardKind:       RewardMedal,
	BranchOps:        &spacerBranchOps,
	Commission:       PromotionRule{Char: Dexterity}, // C2
	EnlistedPromote: PromotionRule{
		Char:      Dexterity,
		MedalMods: true,
	}, // C2* (Rating Promotion)
	OfficerPromote: PromotionRule{Char: Social, MedalMods: true}, // Soc*
	EnlistedRanks: []Rank{
		{Title: "Spacehand", Skill: "Fighter"},
		{Title: "Able Spacer"},
		{Title: "Petty Officer Second"},
		{Title: "Petty Officer First", Skill: "Gunner"},
		{Title: "Chief Petty Officer", Skill: "Sensors"},
		{Title: "Master Chief Petty Officer"},
	},
	OfficerRanks: []Rank{
		{Title: "Ensign", Skill: "Astrogation"},
		{Title: "Sublieutenant"},
		{Title: "Lieutenant", Skill: "Engineer"},
		{Title: "Lt Commander", Skill: "Pilot"},
		{Title: "Commander"},
		{Title: "Captain", Skill: "Leader"},
		{Title: "Admiral"},
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
		// Col 1 — Shore Duty (Major/Minor lost without the education stage).
		{major, major, minor, minor, choose(theTrades...), choose(theTrades...)},
		// Col 2 — Battle.
		{
			sk("Fighter"),
			sk("Fleet Tactics"),
			sk("Pilot"),
			choose(starshipSkls...),
			sk("Gunner"),
			sk("Sensors"),
		},
		// Col 3 — Patrol/Strike.
		{
			sk("Astrogation"),
			sk("Fleet Tactics"),
			sk("Computer"),
			choose(starshipSkls...),
			sk("Gunner"),
			sk("Sensors"),
		},
		// Col 4 — Siege.
		{
			sk("Computer"),
			sk("Strategy"),
			sk("Counsellor"),
			sk("Gunner"),
			sk("Gunner"),
			sk("Sensors"),
		},
		// Col 5 — Mission.
		{
			sk("Diplomat"),
			sk("Admin"),
			cascade("Language", languages...),
			choose(starshipSkls...),
			sk("Liaison"),
			sk("Comms"),
		},
		// Col 6 — Technical.
		{
			choose(oneArt...),
			choose(sciences...),
			sk("Athlete"),
			sk("Medic"),
			sk("Zero-G"),
			choose(theTrades...),
		},
	},
	// Muster-out (Book 1 p.81), indexed 1-12 by (1D + DM). Money passages are
	// recorded at cash value; "Retire x2" is a named retirement benefit. Row 12
	// repeats row 11 (the printed table ends at 11; higher rolls clamp).
	MusterOut: MusterTable{
		1:  {Money: cash(1_000), Benefit: named("Forbidden Knowledge")}, // Low Passage
		2:  {Money: cash(8_000), Benefit: charAward(Strength)},          // Mid Passage
		3:  {Money: cash(10_000), Benefit: named("Wafer Jack")},         // High Passage
		4:  {Money: cash(250_000), Benefit: charAward(Education)},       // StarPass
		5:  {Money: cash(30_000), Benefit: charAward(Strength)},
		6:  {Money: cash(40_000), Benefit: charAward(Dexterity)},
		7:  {Money: cash(50_000), Benefit: charAward(Endurance)},
		8:  {Money: retirementX2(), Benefit: charAward(Intelligence)},
		9:  {Money: retirementX2(), Benefit: named("Ship Share")},
		10: {Money: cash(60_000), Benefit: named("Life Insurance")},
		11: {Money: cash(70_000), Benefit: knighthood()},
		12: {Money: cash(70_000), Benefit: knighthood()},
	},
}
