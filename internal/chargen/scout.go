package chargen

// Scout career data, transcribed from Book 1 p. 79 (the Scout career page).
//
// Slice scope (the rest of the career page is deferred): a Scout each term
// chooses a Duty — Courier (avoid Risk & Reward, 4 skill rolls) or Explorer
// (run Risk & Reward, 8 skill rolls). This slice fixes the iconic Explorer
// duty, so every term runs Risk & Reward and grants 8 skill rolls. Also
// deferred: the R&R reward (Discovery, Land Grant, Fame) and the Major/Minor
// academic awards, which require the education stage this generator does not yet
// model — those grid cells are transcribed as NoAward ("lost" per the page's
// footnote when the character has no Major/Minor). The Personal Soc bump carries
// a second footnote ("lost if Soc is a Caste"); Caste is not modelled, so it is
// applied unconditionally. The muster Benefit column's DM is +Fame/2 and Fame is
// not yet tracked, so a benefit roll is 1D only (rows 1-6) — the high-Fame rows
// (Ship Share at 8 … Knighthood at 12) are unreachable via Benefit until Fame
// exists; the Money column reaches them through its +Terms DM.

// Cell constructors keep the grid readable.
func bump(ch Characteristic) Cell   { return Cell{Kind: AwardBump, Char: ch} }
func sk(name string) Cell           { return Cell{Kind: AwardSkill, Skill: name} }
func choose(options ...string) Cell { return Cell{Kind: AwardChoice, Options: options} }

// cascade is a choice among knowledges under a cascade parent skill (e.g. a
// Language knowledge), granted via the K-K-S progression rather than as a flat
// skill.
func cascade(parent string, knowledges ...string) Cell {
	return Cell{Kind: AwardChoice, Skill: parent, Options: knowledges}
}

func cash(credits int) Benefit            { return Benefit{Kind: Cash, Value: credits} }
func charAward(ch Characteristic) Benefit { return Benefit{Kind: CharBump, Value: 1, Char: ch} }
func named(name string) Benefit           { return Benefit{Kind: Named, Name: name} }

// lost is an academic Major/Minor award the character cannot claim without the
// (not-yet-modelled) education stage — the page's footnote says it is lost.
var lost = Cell{Kind: NoAward}

// The choice cells' option lists are a representative subset, not the full T5
// cascade. "One X" and "Starship Skill" choose among distinct skills (flat
// raise); a chosen Starship Skill that is itself a cascade (Pilot, Engineer, …)
// is left flat here — its sub-knowledge is deferred. languages are knowledges
// under the Language cascade skill (see cascade()).
var (
	trades       = []string{"Steward", "Trader", "Craftsman"}
	arts         = []string{"Artist", "Author", "Performer"}
	sciences     = []string{"Biologist", "Chemist", "Physicist"}
	starshipSkls = []string{"Pilot", "Astrogation", "Engineer", "Sensors", "Gunner"}
	languages    = []string{"Galanglic", "Vilani", "Zdetl"}
)

// ScoutCareer is the Scout career (Book 1 p. 79). Qualify vs the best of
// Str/Dex/End; Risk & Reward rotates Str/Dex/End; Continue vs Int.
var ScoutCareer = Career{
	ID:               Scout,
	Name:             "Scout",
	Qualify:          Qualification{Chars: []Characteristic{Strength, Dexterity, Endurance}},
	CCMode:           RotateCC,
	ControllingChars: []Characteristic{Strength, Dexterity, Endurance},
	Continue:         ContinueRule{UseChar: true, Char: Intelligence},
	Advance:          RollLow,
	EligPerTerm:      8, // Explorer duty (Courier duty's 4 is deferred)
	Skills: SkillGrid{
		// Col 0 — Personal.
		{bump(Strength), bump(Dexterity), bump(Endurance), bump(Intelligence), bump(Education), bump(Social)},
		// Col 1 — Academic (Major/Minor lost without the education stage).
		{lost, lost, lost, lost, choose(trades...), choose(trades...)},
		// Col 2 — Courier.
		{sk("Comms"), cascade("Language", languages...), sk("Computer"), sk("JOT"), sk("Gunner"), choose(starshipSkls...)},
		// Col 3 — Exploration.
		{sk("Survey"), sk("Survival"), sk("Hostile Environ"), sk("Animals"), sk("Vacc Suit"), sk("Navigation")},
		// Col 4 — Business.
		{sk("Diplomat"), sk("Sensors"), sk("Trader"), sk("Teacher"), sk("Fighter"), sk("Streetwise")},
		// Col 5 — Vocation.
		{sk("Survey"), sk("Flyer"), cascade("Language", languages...), choose(starshipSkls...), sk("Engineer"), sk("Comms")},
		// Col 6 — Avocation.
		{choose(arts...), choose(sciences...), sk("Seafarer"), sk("Athlete"), sk("Medic"), choose(trades...)},
	},
	// Muster-out (Book 1 p. 79), indexed 1-12 by (1D + DM). Money-column passages
	// are recorded at their standard cash value.
	MusterOut: MusterTable{
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
		11: {Money: cash(70_000), Benefit: named("Fame +2")},
		12: {Money: cash(80_000), Benefit: named("Knighthood")},
	},
}
