package chargen

// Merchant career data, transcribed from Book 1 p. 80. The Merchant runs the
// standard Risk & Reward (a failed Risk still injures), but each Reward success
// is escalating Ship Shares (the Nth success is worth N shares — RewardShipShares)
// toward owning a starship. It has a dual rank track: a Rating ladder (Temp …
// Drive Helper), a Commission (vs Int) to the Officer ladder (Fourth Officer …
// Senior Captain), Rating Promotion vs Dex, and Officer Promotion.
//
// Slice scope: begin is modelled as the automatic Temp start (the Spacehand-vs-Dex
// and 4th-Officer-vs-Int begin options are deferred). Officer Promotion is
// modelled vs Int rather than the book's "Terms x2" (deferred, along with the
// "+3 if Int 8+" modifier and the Temp/Rating both-rolls-in-one-term rule). The
// muster "Ship Share" benefits stay named strings (they do not add to ShipShares),
// and the +Officer-Rank muster Money DM is treated as +Terms.

// MerchantCareer is the Merchant (Book 1 p. 80).
var MerchantCareer = Career{
	ID:               Merchant,
	Name:             "Merchant",
	AutoBegin:        true,                                                           // Temp start
	ControllingChars: []Characteristic{Strength, Dexterity, Endurance, Intelligence}, // C1 C2 C3 C4
	Continue:         ContinueRule{UseChar: true, Char: Strength},                    // Str
	Advance:          RollLow,
	EligPerTerm:      4,
	BenefitDM:        DMOfficerRank,
	RewardKind:       RewardShipShares,
	Commission:       PromotionRule{Char: Intelligence}, // Int
	EnlistedPromote:  PromotionRule{Char: Dexterity},    // Rating Promotion, Dex
	OfficerPromote:   PromotionRule{Char: Intelligence}, // approximates the book's Terms x2
	// Rating ladder (Temp is rank 1); auto-skills on some rungs (Book 1 p. 80).
	EnlistedRanks: []Rank{
		{Title: "Temp"},
		{Title: "Spacehand"},
		{Title: "Steward Apprentice", Skill: "Steward"},
		{Title: "Drive Helper", Skill: "Engineer"},
	},
	OfficerRanks: []Rank{
		{Title: "Fourth Officer", Skill: "Steward"},
		{Title: "Third Officer", Skill: "Engineer"},
		{Title: "Second Officer", Skill: "Astrogation"},
		{Title: "First Officer", Skill: "Pilot"},
		{Title: "Captain"},
		{Title: "Senior Captain"},
	},
	Skills: SkillGrid{
		// Col 0 — Personal.
		{bump(Strength), bump(Dexterity), bump(Endurance), bump(Intelligence), bump(Education), bump(Social)},
		// Col 1 — Academic (Major/Minor lost without the education stage).
		{major, major, minor, minor, choose(theTrades...), choose(theTrades...)},
		// Col 2 — Space Travel.
		{sk("Astrogation"), sk("Pilot"), sk("Medic"), sk("Sensors"), sk("Steward"), sk("Gunner")},
		// Col 3 — Trade.
		{sk("Broker"), sk("Trader"), sk("Diplomat"), sk("Admin"), sk("Steward"), sk("Trader")},
		// Col 4 — Business.
		{sk("Computer"), sk("Trader"), sk("Driver"), sk("Advocate"), sk("Steward"), sk("Comms")},
		// Col 5 — Vocation.
		{sk("Broker"), sk("Admin"), cascade("Language", languages...), choose(starshipSkls...), sk("JOT"), sk("Vacc Suit")},
		// Col 6 — Technical.
		{choose(oneArt...), choose(sciences...), sk("Computer"), sk("Comms"), sk("Medic"), choose(theTrades...)},
	},
	// Muster-out (Book 1 p. 80), indexed 1-12 by (1D + DM). Passages are recorded
	// at cash value.
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
