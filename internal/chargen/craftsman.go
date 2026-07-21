package chargen

// Craftsman career data, transcribed from Book 1 p. 75. The Craftsman does not
// roll Risk & Reward; instead each term attempts a Masterpiece (see
// runCraftsmanTerm): Master Points total the Controlling Characteristic, the
// Craftsman skill, and up to five other skills at level 6+, and with 40+ points
// a 9D roll creates a work of art. The Craftsman skill rises every term, and
// Continue is vs twice the Craftsman skill — so the more skilled keep working.
//
// Slice scope: the "To Begin" prerequisite (two skills at 6+ and Craftsman-1 —
// the Craftsman is typically not a first career) is deferred, so the career
// AutoBegins; a fresh Craftsman rarely reaches 40 Master Points and mostly gains
// skills from failed attempts. Also deferred: the Masterpiece Value table and
// QREBS allocation, the Perfect Masterpiece, and Vintage value.

// CraftsmanCareer is the Craftsman (Book 1 p. 75).
var CraftsmanCareer = Career{
	ID:               Craftsman,
	Name:             "Craftsman",
	AutoBegin:        true,
	Term:             CraftsmanTerm,
	ControllingChars: []Characteristic{Strength, Dexterity, Endurance, Intelligence}, // C1 C2 C3 C4
	Continue: ContinueRule{
		UseSkill:  "Craftsman",
		SkillMult: 2,
	}, // Craftsman x2
	EligPerTerm: 4,
	BenefitDM:   DMTerms,
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
			sk("Seafarer"),
			sk("Navigation"),
			sk("Hostile Environ"),
			sk("Flyer"),
			sk("Driver"),
			sk("Vacc Suit"),
		},
		// Col 3 — General.
		{
			sk("Animals"),
			sk("Comms"),
			sk("Computer"),
			sk("Designer"),
			sk("Designer"),
			sk("Designer"),
		},
		// Col 4 — Business.
		{sk("Liaison"), sk("Comms"), sk("Bureaucrat"), sk("Diplomat"), sk("Leader"), sk("Trader")},
		// Col 5 — Vocation ("New Trade" is a representative trade choice).
		{
			sk("Naval Architect"),
			choose(oneArt...),
			choose(theTrades...),
			choose(theTrades...),
			choose(theTrades...),
			choose(theTrades...),
		},
		// Col 6 — Avocation.
		{
			sk("Animals"),
			choose(oneArt...),
			choose(sciences...),
			sk("Athlete"),
			sk("Medic"),
			choose(theTrades...),
		},
	},
	// Muster-out (Book 1 p. 75), indexed 1-12 by (1D + DM). Money passages are
	// recorded at cash value.
	MusterOut: MusterTable{
		1:  {Money: cash(1_000), Benefit: named("Forbidden Knowledge")}, // Low Passage
		2:  {Money: cash(1_000), Benefit: named("Forbidden Knowledge")}, // Low Passage
		3:  {Money: cash(8_000), Benefit: named("Wafer Jack")},          // Mid Passage
		4:  {Money: cash(10_000), Benefit: charAward(Education)},        // High Passage
		5:  {Money: cash(15_000), Benefit: charAward(Strength)},
		6:  {Money: cash(250_000), Benefit: charAward(Dexterity)}, // StarPass
		7:  {Money: cash(25_000), Benefit: charAward(Endurance)},
		8:  {Money: cash(30_000), Benefit: charAward(Intelligence)},
		9:  {Money: cash(35_000), Benefit: named("Ship Share")},
		10: {Money: cash(40_000), Benefit: named("TAS Fellowship")},
		11: {Money: cash(50_000), Benefit: named("Directorate")},
		12: {Money: cash(60_000), Benefit: named("TAS Life Membership")},
	},
}
