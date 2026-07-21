package chargen

// Rogue career data, transcribed from Book 1 p. 84.
//
// Each term the Rogue masterminds a Scheme (SchemeCareer, runRogueTerm): a
// Scheme rolled from the Rogue Schemes table pays out on a successful Reward,
// a failed Risk imprisons the Rogue and earns Infamy, and skill eligibility
// swings with the outcome (Successful Scheme 6 / Failed Scheme 3 / In Prison 2).
// The Scheme's "Mod +Terms" and "12 is always failure" rules live in the engine
// (ContinueRule.TermsMod and AutoFailOn12/Career.held); both footnotes sit under
// the whole To Begin / Risk & Reward / Continue block on p. 84, so both cover
// all four rolls.
// Deferred pieces (the exact prison sentence, the ±1 Flux modification, and
// selecting a previous career in place of the roll) are noted at runRogueTerm.
// The Rogue is a fixed-CC career with the p. 84 skill grid and mustering-out
// table (whose Benefit column, unusually, also gets the +Terms DM).

// allChars is every characteristic — the Rogue selects one as the Controlling
// Characteristic used throughout the career (Book 1 p. 84).
var allChars = []Characteristic{Strength, Dexterity, Endurance, Intelligence, Education, Social}

// soldierSkls is a representative "Soldier Skill" choice list (the full cascade
// is deferred), mirroring how starshipSkls stands in for "Starship Skill".
var soldierSkls = []string{"Fighter", "Gunner", "Heavy Weapons", "Tactics"}

// RogueCareer is the Rogue (Book 1 p. 84): a fixed-CC career. To Begin, Risk &
// Reward, and Continue all use the one selected Controlling Characteristic, so
// the career carries no Qualification — under FixedCC, beginCareer rolls against
// the chosen CC and never reads one.
var RogueCareer = Career{
	ID:               Rogue,
	Name:             "Rogue",
	CCMode:           FixedCC,
	ControllingChars: allChars,
	Continue:         ContinueRule{UseCC: true, TermsMod: true}, // "Mod +Terms" (Book 1 p. 84)
	AutoFailOn12:     true,                                      // "But, 12 is always automatic failure" (p. 84)
	EligPerTerm:      2,
	Term: RogueTerm,
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
		{choose(sciences...), major, minor, choose(oneArt...), choose(theTrades...), sk("Gambler")},
		// Col 2 — World Travel.
		{
			sk("Driver"),
			sk("Flyer"),
			sk("Hostile Environ"),
			sk("High-G"),
			sk("Vacc Suit"),
			sk("Navigation"),
		},
		// Col 3 — Space Travel ("Astrogator" on p. 84 is the Astrogation skill).
		{
			choose(starshipSkls...),
			sk("Pilot"),
			sk("Engineer"),
			sk("Zero-G"),
			sk("Vacc Suit"),
			sk("Astrogation"),
		},
		// Col 4 — Business.
		{sk("Trader"), sk("Broker"), sk("Computer"), sk("JOT"), sk("Teacher"), sk("Fighter")},
		// Col 5 — Vocation.
		{
			sk("Advocate"),
			sk("Counsellor"),
			cascade("Language", languages...),
			sk("Leader"),
			sk("Streetwise"),
			sk("Comms"),
		},
		// Col 6 — Avocation.
		{
			choose(oneArt...),
			choose(sciences...),
			sk("Athlete"),
			choose(soldierSkls...),
			choose(starshipSkls...),
			choose(theTrades...),
		},
	},
	// Muster-out (Book 1 p. 84), indexed 1-12 by (1D + DM). Money-column passages
	// are recorded at their standard cash value.
	MusterOut: MusterTable{ //nolint:dupl // per-career muster-out table; parallel by design
		1:  {Money: cash(40_000), Benefit: charAward(Strength)},
		2:  {Money: cash(250_000), Benefit: charAward(Education)},    // StarPass
		3:  {Money: cash(250_000), Benefit: named("Wafer Jack")},     // StarPass
		4:  {Money: cash(10_000), Benefit: charAward(Dexterity)},     // High Passage
		5:  {Money: cash(10_000), Benefit: charAward(Endurance)},     // High Passage
		6:  {Money: cash(250_000), Benefit: named("Life Insurance")}, // StarPass
		7:  {Money: cash(25_000), Benefit: named("Ship Share")},
		8:  {Money: cash(30_000), Benefit: knighthood()},
		9:  {Money: cash(35_000), Benefit: named("Ship Share")},
		10: {Money: cash(40_000), Benefit: named("Ship Share")},
		11: {Money: cash(50_000), Benefit: named("Ship Share")},
		12: {Money: cash(90_000), Benefit: knighthood()},
	},
}
