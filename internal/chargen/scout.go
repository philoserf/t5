package chargen

// Scout career data, transcribed from Book 1 p.79 (the Scout career page).
//
// Each term a Scout chooses a Duty (Policy.ChooseExplorerDuty): Courier (avoid
// Risk & Reward, 4 skill rolls) or Explorer (run Risk & Reward, EligPerTerm = 8
// skill rolls). A successful Explorer Reward is a Discovery (RewardDiscovery):
// a Land Grant and Fame +1, which in turn feeds the muster Benefit column's
// +Fame/2 DM (BenefitDM DMFameHalf). The Academic column's Major/Minor awards
// are live too — the education stage now runs, so those cells are real
// AwardMajor/AwardMinor cells and it is applyCell, not the transcription, that
// honors the page's "lost if the character has no Major/Minor" footnote.
//
// Deferred: the Personal Soc bump's second footnote, "lost if Soc is a Caste".
// Caste is modelled now (chargen.Character.Caste, from a casted sophont species
// — see sophont.go), so the input exists; what is missing is the rule reading
// it, and the bump is still applied unconditionally.

// Cell constructors keep the grid readable.
func bump(ch Characteristic) Cell   { return Cell{Kind: AwardBump, Char: ch} }
func sk(name string) Cell           { return Cell{Kind: AwardSkill, Skill: name} }
func choose(options ...string) Cell { return Cell{Kind: AwardChoice, Options: options} }

// cascade is a choice among knowledges under a cascade parent skill (e.g. a
// Language knowledge), granted via the K-K-S progression rather than as a flat
// skill.
//
//nolint:unparam // general cascade helper; parent kept explicit for the readable grids
func cascade(parent string, knowledges ...string) Cell {
	return Cell{Kind: AwardChoice, Skill: parent, Options: knowledges}
}

func cash(credits int) Benefit            { return Benefit{Kind: Cash, Value: credits} }
func charAward(ch Characteristic) Benefit { return Benefit{Kind: CharBump, Value: 1, Char: ch} }
func fameAward() Benefit                  { return Benefit{Kind: FameBump, Value: 1} }
func named(name string) Benefit           { return Benefit{Kind: Named, Name: name} }
func knighthood() Benefit                 { return Benefit{Kind: Knighted, Name: "Knighthood"} }
func pensionX2() Benefit                  { return Benefit{Kind: PensionX2} }
func retirementX2() Benefit               { return Benefit{Kind: RetirementX2} }
func fameAwardN(n int) Benefit            { return Benefit{Kind: FameBump, Value: n} }

// major and minor award the character's College Major/Minor (Book 1 p.79
// Academic column). The page's footnote — "lost if the character has no
// Major/Minor" — is honored by applyCell: an uneducated character gains nothing.
var (
	major = Cell{Kind: AwardMajor}
	minor = Cell{Kind: AwardMinor}
)

// One Art and One Trade choices draw the canonical lists (oneArt, theTrades from
// homeworld.go). The remaining choice lists are a representative subset, not the
// full T5 cascade: "Starship Skill" chooses among distinct skills (flat raise);
// a chosen Starship Skill that is itself a cascade (Pilot, Engineer, …) is left
// flat, its sub-knowledge deferred. languages are knowledges under the Language
// cascade skill (see cascade()).
var (
	sciences     = []string{"Biologist", "Chemist", "Physicist"}
	starshipSkls = []string{"Pilot", "Astrogation", "Engineer", "Sensors", "Gunner"}
	languages    = []string{"Galanglic", "Vilani", "Zdetl"}
)

// ScoutCareer is the Scout career (Book 1 p.79). Qualify vs the best of
// Str/Dex/End; Risk & Reward rotates Str/Dex/End; Continue vs Int.
//
// Deferred: the Scout's box entry "Retry R&R C5" is a term-level retry of a failed
// Risk & Reward roll against Education (C5) — NOT a Begin retry. It was once
// modelled as a Begin retry against Education, which conflated the two; the page
// underspecifies which roll it retries and how, so rather than guess a misread into
// the term loop it stays out until the mechanic is pinned down. The Scout's other
// R&R escape, avoiding the rolls entirely via Courier Duty, is implemented
// (Term ScoutTerm).
var ScoutCareer = Career{
	ID:               Scout,
	Name:             "Scout",
	Qualify:          Qualification{Chars: []Characteristic{Strength, Dexterity, Endurance}},
	CCMode:           RotatingCC,
	ControllingChars: []Characteristic{Strength, Dexterity, Endurance},
	Continue:         ContinueRule{UseChar: true, Char: Intelligence},
	EligPerTerm:      8, // Explorer duty
	Term:             ScoutTerm,
	RewardKind:       RewardDiscovery, // a successful Explorer Reward is a Discovery
	BenefitDM:        DMFameHalf,
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
		// Col 1 — Academic: Major, Minor, One Trade, One Trade.
		{major, major, minor, minor, choose(theTrades...), choose(theTrades...)},
		// Col 2 — Courier.
		{
			sk("Comms"),
			cascade("Language", languages...),
			sk("Computer"),
			sk("JOT"),
			sk("Gunner"),
			choose(starshipSkls...),
		},
		// Col 3 — Exploration.
		{
			sk("Survey"),
			sk("Survival"),
			sk("Hostile Environ"),
			sk("Animals"),
			sk("Vacc Suit"),
			sk("Navigation"),
		},
		// Col 4 — Business.
		{
			sk("Diplomat"),
			sk("Sensors"),
			sk("Trader"),
			sk("Teacher"),
			sk("Fighter"),
			sk("Streetwise"),
		},
		// Col 5 — Vocation.
		{
			sk("Survey"),
			sk("Flyer"),
			cascade("Language", languages...),
			choose(starshipSkls...),
			sk("Engineer"),
			sk("Comms"),
		},
		// Col 6 — Avocation.
		{
			choose(oneArt...),
			choose(sciences...),
			sk("Seafarer"),
			sk("Athlete"),
			sk("Medic"),
			choose(theTrades...),
		},
	},
	// Muster-out (Book 1 p.79), indexed 1-12 by (1D + DM). Money-column passages
	// are recorded at their standard cash value.
	MusterOut: MusterTable{ //nolint:dupl // per-career muster-out table; parallel by design
		1:  {Money: cash(1_000), Benefit: named("Ship Share")},          // Low Passage
		2:  {Money: cash(8_000), Benefit: named("Forbidden Knowledge")}, // Mid Passage
		3:  {Money: cash(10_000), Benefit: named("Wafer Jack")},         // High Passage
		4:  {Money: cash(250_000), Benefit: charAward(Education)},       // StarPass
		5:  {Money: cash(30_000), Benefit: charAward(Strength)},
		6:  {Money: cash(40_000), Benefit: charAward(Dexterity)},
		7:  {Money: cash(50_000), Benefit: charAward(Endurance)},
		8:  {Money: cash(60_000), Benefit: named("Ship Share")},
		9:  {Money: cash(60_000), Benefit: named("Life Insurance")},
		10: {Money: cash(60_000), Benefit: named("TAS Fellowship")},
		11: {Money: cash(70_000), Benefit: fameAwardN(2)},
		12: {Money: cash(80_000), Benefit: knighthood()},
	},
}
