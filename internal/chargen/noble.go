package chargen

// Noble career data, transcribed from Book 1 p. 85. The Noble plays Return &
// Intrigue (see runIntrigueTerm) instead of Risk & Reward: a contest for power
// that risks Exile and offers Elevation (a rise in Social Standing and Noble
// rank, each awarding a Land Grant). A Noble's rank is their Social Standing —
// see NobleTitle.
//
// Slice scope: begin is modelled as automatic (the "Soc B+" prerequisite is
// deferred, so any character can be a Noble here). Deferred: the Return/Intrigue
// modifiers (-Successful Intrigues / +Times Exiled), the Elevation Flux, the
// Base Fame (1.5 x Soc), the Land-Grant hex/profit detail, the muster Power
// (Proxy) column, and the StarPass courtesy upgrade (StarPass is recorded at its
// standard cash value).

// NobleTitle returns the Noble title for a Social Standing (Book 1 p. 85), or ""
// below noble rank. Titles above Viscount need Soc past the human maximum of 15.
func NobleTitle(soc int) string {
	switch soc {
	case 10:
		return "Gentleman"
	case 11:
		return "Knight"
	case 12:
		return "Baronet"
	case 13:
		return "Baron"
	case 14:
		return "Marquis"
	case 15:
		return "Viscount"
	default:
		return ""
	}
}

// NobleCareer is the Noble (Book 1 p. 85).
var NobleCareer = Career{
	ID:               Noble,
	Name:             "Noble",
	AutoBegin:        true,
	ReturnIntrigue:   true,
	ControllingChars: []Characteristic{Dexterity, Endurance, Intelligence, Education}, // C2 C3 C4 C5
	Continue:         ContinueRule{Fixed: 7},                                          // Continue 7
	Advance:          RollHigh,                                                        // Elevation is a roll-high check
	EligPerTerm:      4,
	BenefitDM:        DMTerms,
	Skills: SkillGrid{
		// Col 0 — Personal.
		{bump(Strength), bump(Dexterity), bump(Endurance), bump(Intelligence), bump(Education), bump(Social)},
		// Col 1 — Academic (Major/Minor lost without the education stage).
		{major, major, minor, minor, choose(theTrades...), choose(theTrades...)},
		// Col 2 — Travel.
		{sk("Driver"), sk("Flyer"), sk("Pilot"), choose(starshipSkls...), sk("High-G"), sk("Zero-G")},
		// Col 3 — General.
		{sk("Advocate"), sk("Counsellor"), sk("Bureaucrat"), sk("Liaison"), sk("Leader"), sk("Leader")},
		// Col 4 — Political.
		{sk("Liaison"), sk("Strategy"), sk("Tactics"), sk("Diplomat"), sk("Advocate"), sk("Leader")},
		// Col 5 — Vocation ("Capital" is the world-knowledge of the noble's fief).
		{sk("Capital"), sk("Admin"), cascade("Language", languages...), choose(starshipSkls...), sk("Bureaucrat"), sk("Comms")},
		// Col 6 — Technical.
		{choose(oneArt...), choose(sciences...), sk("Athlete"), choose(soldierSkls...), choose(starshipSkls...), choose(theTrades...)},
	},
	// Muster-out (Book 1 p. 85), indexed 1-12 by (1D + DM). StarPass is recorded
	// at cash value; the Power (Proxy) column is deferred.
	MusterOut: MusterTable{
		1:  {Money: cash(250_000), Benefit: named("Forbidden Knowledge")}, // StarPass
		2:  {Money: cash(250_000), Benefit: charAward(Strength)},          // StarPass
		3:  {Money: cash(250_000), Benefit: named("Wafer Jack")},          // StarPass
		4:  {Money: cash(250_000), Benefit: charAward(Education)},         // StarPass
		5:  {Money: cash(100_000), Benefit: named("Directorship")},
		6:  {Money: cash(200_000), Benefit: charAward(Dexterity)},
		7:  {Money: cash(300_000), Benefit: charAward(Endurance)},
		8:  {Money: cash(400_000), Benefit: charAward(Intelligence)},
		9:  {Money: cash(500_000), Benefit: named("Ship Share")},
		10: {Money: cash(600_000), Benefit: named("Life Insurance")},
		11: {Money: cash(700_000), Benefit: named("TAS Life Membership")},
		12: {Money: cash(800_000), Benefit: named("Directorship")},
	},
}
