package chargen

import "fmt"

// Functionary career data, transcribed from Book 1 p.87. The Functionary plays
// Office Politics (see runPoliticsTerm) instead of Risk & Reward: two unmodified
// rolls against the Controlling Characteristic — a failed Risk ends the career
// (job loss, no injury) and a successful Reward is a promotion up the F0-F8 rank
// ladder (Clerk … Secretary).
//
// Slice scope: the "Total Terms x3" begin (Functionary is never a first career)
// is deferred, so the career AutoBegins; the F7 UnderSecretary's 1D ordinal, the
// prior-career-specific F6 titles (College President, etc.), and the
// +Officer-Rank muster Money DM (treated as +Terms) are also deferred. The
// "Pension x2" muster pay doubles the Functionary's Cr15,000/year pension (an
// entitlement; see entitlement.go).
//
// The page prints two automatic muster-out awards beside the table. The Gold
// Watch is granted (functionaryAutoBenefits). The second — "Automatic:
// Directorship if Rank F6+" — is deferred: a Directorship's substance is its
// annual Cr36,000 (p.68), and there is no machinery to pay it. entitlement.go
// computes only the professor's pension, armed-forces retirement pay, and the
// civil pension, all of which are age-gated career pensions rather than a
// benefit-granted annuity. Since the rolled row 11 already hands out an unvalued
// "Directorship", granting a second one automatically would add a name and no
// income; the annuity and the automatic grant should land together.

// functionaryAutoBenefits is the Functionary's automatic muster-out award
// (Book 1 p.87): "Automatic: Gold Watch (Value= 100 x Terms as Functionary)".
//
// The value rides in the benefit's name rather than in Credits or a new Benefit
// kind. A Gold Watch is an object with a resale value, not a payout — adding
// Cr100/term to a character's cash would spend a keepsake they still own — and
// the value is not mechanical: nothing in Book 1 reads it. So it wants a place
// on the sheet, not in the engine, and Named already goes there. A dedicated
// kind would buy a structured field no rule consults.
func functionaryAutoBenefits(rec CareerRecord) []Benefit {
	return []Benefit{named(fmt.Sprintf("Gold Watch (Cr%d)", goldWatchValuePerTerm*rec.Terms))}
}

// goldWatchValuePerTerm is the Gold Watch's Cr per term served (Book 1 p.87).
const goldWatchValuePerTerm = 100

// FunctionaryCareer is the Functionary (Book 1 p.87).
var FunctionaryCareer = Career{
	ID:        Functionary,
	Name:      "Functionary",
	AutoBegin: true,
	Term:      PoliticsTerm,
	ControllingChars: []Characteristic{
		Dexterity,
		Endurance,
		Intelligence,
		Education,
	}, // C2 C3 C4 C5
	Continue: ContinueRule{
		Fixed: 15,
	}, // Office Politics decides continuation; the policy chooses
	EligPerTerm: 4,
	BenefitDM:   DMRankF0, // the ladder prints F0..F8, so a Clerk's muster DM is +0 (p.87)
	// The F0-F8 rank ladder (Book 1 p.87); rank 1 is Clerk (F0).
	EnlistedRanks: []Rank{
		{Title: "Clerk", Skill: "Bureaucrat"},
		{Title: "Supervisor"},
		{Title: "Senior Supervisor", Skill: "Admin"},
		{Title: "Manager", Skill: "Bureaucrat"},
		{Title: "Senior Manager"},
		{Title: "Assistant Director"},
		{Title: "Director"},
		{Title: "UnderSecretary"},
		{Title: "Secretary"},
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
		// Col 1 — Academic (Major/Minor lost without the education stage).
		{major, major, minor, minor, choose(theTrades...), choose(theTrades...)},
		// Col 2 — World Travel.
		{
			sk("High-G"),
			sk("Vacc Suit"),
			sk("Driver"),
			sk("Flyer"),
			sk("Navigation"),
			sk("Seafarer"),
		},
		// Col 3 — General ("Any Skill" is a representative Citizen-job choice).
		{
			choose(theTrades...),
			choose(oneArt...),
			choose(sciences...),
			choose(citizenJobs...),
			sk("Bureaucrat"),
			sk("Leader"),
		},
		// Col 4 — Business.
		{
			sk("Advocate"),
			sk("Broker"),
			sk("Trader"),
			sk("Teacher"),
			choose(theTrades...),
			sk("Driver"),
		},
		// Col 5 — Vocation.
		{
			sk("Advocate"),
			sk("Comms"),
			cascade("Language", languages...),
			sk("Admin"),
			sk("Bureaucrat"),
			sk("Comms"),
		},
		// Col 6 — Avocation.
		{
			choose(oneArt...),
			choose(sciences...),
			sk("Athlete"),
			sk("Designer"),
			sk("Seafarer"),
			choose(theTrades...),
		},
	},
	// Muster-out, indexed 1-12 by (1D + DM). The career page (p.87) prints 11
	// rows; row 12 (Pension x2 | Knighthood) comes from the consolidated
	// Muster-Out Tables. "Pension x2" is a named benefit.
	MusterOut: MusterTable{
		1:  {Money: cash(5_000), Benefit: named("Forbidden Knowledge")},
		2:  {Money: cash(10_000), Benefit: charAward(Strength)},
		3:  {Money: cash(15_000), Benefit: named("Wafer Jack")},
		4:  {Money: cash(20_000), Benefit: charAward(Strength)},
		5:  {Money: cash(10_000), Benefit: charAward(Dexterity)},  // High Passage
		6:  {Money: cash(250_000), Benefit: charAward(Endurance)}, // StarPass
		7:  {Money: cash(25_000), Benefit: charAward(Intelligence)},
		8:  {Money: cash(30_000), Benefit: named("Life Insurance")},
		9:  {Money: cash(35_000), Benefit: named("TAS Fellowship")},
		10: {Money: pensionX2(), Benefit: knighthood()},
		11: {Money: pensionX2(), Benefit: named("Directorship")},
		12: {Money: pensionX2(), Benefit: knighthood()},
	},
	AutoBenefits: functionaryAutoBenefits,
}
