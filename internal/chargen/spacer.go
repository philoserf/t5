package chargen

// Spacer career data, transcribed from Book 1 p. 81. The Spacer (Navy) is the
// third armed-forces career and, like the Marine, pure data on the rank engine.
// Its enlisted track is the naval Rating ladder, so its "Rating Promotion" is
// the engine's EnlistedPromote. It differs from the Soldier/Marine in data: it
// begins vs Int, Risk & Reward rotates Str/Dex/Int, Commission and Rating
// Promotion are vs Dex, Officer Promotion vs Soc, Continue vs Str, plus its own
// naval rank titles, skill grid, and muster table.
//
// Same deferrals as the Soldier/Marine: the Branch and Operations R&R modifiers,
// the +1 skill for a Commission or Promotion, branch automatic skills, the
// "Retire x2" muster pay (a named benefit) and its +Officer-Rank Benefit DM, and
// Command College.

// spacerBranchOps is the Spacer's Naval Branch and Operations tables (Book 1
// p. 81). Naval Operations use no per-branch DM, so every branch's OpsDM is 0.
var spacerBranchOps = BranchOps{
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
	// Naval Operations (1D): Battle, Strike, Siege, Patrol, Mission, ANM School,
	// Shore Duty, Shore Duty.
	OpsMods: [10]int{1: 2, 2: 2, 3: 0, 4: 1, 5: 3, 6: 0, 7: 0, 8: 0},
}

// SpacerCareer is the Spacer (Book 1 p. 81).
var SpacerCareer = Career{
	ID:               Spacer,
	Name:             "Spacer",
	Qualify:          Qualification{Chars: []Characteristic{Intelligence}},
	CCMode:           RotateCC,
	ControllingChars: []Characteristic{Strength, Dexterity, Intelligence}, // C1 C2 C4
	Continue:         ContinueRule{UseChar: true, Char: Strength},         // C1
	Advance:          RollLow,
	EligPerTerm:      4,
	BenefitDM:        DMOfficerRank,
	RewardKind:       RewardMedal,
	BranchOps:        &spacerBranchOps,
	Commission:       PromotionRule{Char: Dexterity},                        // C2
	EnlistedPromote:  PromotionRule{Char: Dexterity, MedalsAndWounds: true}, // C2* (Rating Promotion)
	OfficerPromote:   PromotionRule{Char: Social, MedalsAndWounds: true},    // Soc*
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
		{bump(Strength), bump(Dexterity), bump(Endurance), bump(Intelligence), bump(Education), bump(Social)},
		// Col 1 — Shore Duty (Major/Minor lost without the education stage).
		{major, major, minor, minor, choose(theTrades...), choose(theTrades...)},
		// Col 2 — Battle.
		{sk("Fighter"), sk("Fleet Tactics"), sk("Pilot"), choose(starshipSkls...), sk("Gunner"), sk("Sensors")},
		// Col 3 — Patrol/Strike.
		{sk("Astrogation"), sk("Fleet Tactics"), sk("Computer"), choose(starshipSkls...), sk("Gunner"), sk("Sensors")},
		// Col 4 — Siege.
		{sk("Computer"), sk("Strategy"), sk("Counsellor"), sk("Gunner"), sk("Gunner"), sk("Sensors")},
		// Col 5 — Mission.
		{sk("Diplomat"), sk("Admin"), cascade("Language", languages...), choose(starshipSkls...), sk("Liaison"), sk("Comms")},
		// Col 6 — Technical.
		{choose(oneArt...), choose(sciences...), sk("Athlete"), sk("Medic"), sk("Zero-G"), choose(theTrades...)},
	},
	// Muster-out (Book 1 p. 81), indexed 1-12 by (1D + DM). Money passages are
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
