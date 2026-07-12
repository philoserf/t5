package chargen

// Scholar career data, transcribed from Book 1 p. 76. The Scholar's life is
// Research and Publication: it runs the standard Risk & Reward (Research is the
// Risk, so a failure still injures; Publication is the Reward, and a success is
// a Publication — the Scholar's reward token). Publications raise the Promotion
// and Continue targets. The Scholar has a single rank ladder (Lecturer …
// Distinguished Professor) with no officer track or commission.
//
// Slice scope: begin is modelled as a roll vs Edu (the "automatic if Edu 8+"
// and the Amateur/Non-Traditional low-Edu paths are simplified — a Scholar here
// starts at rank 1). Deferred: Tenure (so promotion is not gated past Scholar 3),
// the Scholar's own Major/Minor selection when the character has no degree (so
// the Academic Major/Minor cells are lost for the uneducated, as elsewhere), the
// Award-Winning double Publication, the Research-Success Major+2 and Promoted+1
// skill eligibility, and the muster Money-column +Scholar-Rank DM (treated as
// the standard +Terms).

// ScholarCareer is the Scholar (Book 1 p. 76).
var ScholarCareer = Career{
	ID:               Scholar,
	Name:             "Scholar",
	Qualify:          Qualification{Chars: []Characteristic{Education}}, // Begin vs Edu
	CCMode:           RotateCC,
	ControllingChars: []Characteristic{Strength, Dexterity, Endurance, Intelligence}, // C1 C2 C3 C4
	Continue:         ContinueRule{UseChar: true, Char: Education, PubsMod: true},    // Edu, Mod +Pubs
	Advance:          RollLow,
	EligPerTerm:      4,
	BenefitDM:        DMTerms,
	RewardKind:       RewardPublication,
	EnlistedPromote:  PromotionRule{Char: Intelligence, PubsMod: true}, // Int, Mod +Pubs
	// A single rank ladder (Book 1 p. 76): rank 1 is Lecturer. No auto-skills.
	EnlistedRanks: []Rank{
		{Title: "Lecturer"},
		{Title: "Instructor"},
		{Title: "Assistant Professor"},
		{Title: "Associate Professor"},
		{Title: "Professor"},
		{Title: "Distinguished Professor"},
	},
	Skills: SkillGrid{
		// Col 0 — Personal.
		{bump(Strength), bump(Dexterity), bump(Endurance), bump(Intelligence), bump(Education), bump(Social)},
		// Col 1 — Academic.
		{major, major, minor, minor, choose(theTrades...), choose(theTrades...)},
		// Col 2 — Travel.
		{sk("Seafarer"), sk("Navigation"), sk("Hostile Environ"), sk("Flyer"), sk("Driver"), sk("Vacc Suit")},
		// Col 3 — General.
		{sk("Survey"), sk("Astrogation"), sk("Hostile Environ"), sk("Survival"), sk("Animals"), sk("Bureaucrat")},
		// Col 4 — Conflict.
		{sk("Fighter"), sk("Fighter"), sk("Stealth"), sk("Flyer"), sk("Gunner"), sk("Sensors")},
		// Col 5 — Vocation.
		{sk("Admin"), cascade("Language", languages...), choose(sciences...), sk("Comms"), choose(starshipSkls...), sk("Bureaucrat")},
		// Col 6 — Avocation.
		{sk("Seafarer"), choose(oneArt...), choose(sciences...), sk("Athlete"), sk("Medic"), choose(theTrades...)},
	},
	// Muster-out (Book 1 p. 76), indexed 1-12 by (1D + DM). Money passages are
	// recorded at cash value. Row 12 repeats row 11 (the printed table ends at 11).
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
		12: {Money: cash(50_000), Benefit: named("TAS Fellowship")},
	},
}
