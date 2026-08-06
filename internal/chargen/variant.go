package chargen

import "github.com/philoserf/t5/internal/dice"

// The per-shape term variants (Book 1 careers): the careers that replace the
// Standard Risk & Reward term with something of their own — Citizen Life, Fame,
// Masterpiece, Office Politics, Return & Intrigue, a Rogue Scheme, an Undercover
// Assignment. runTerm (career.go) dispatches to these by TermVariant. Split from
// career.go (#331); all one package.

// citizenJobs is a representative list of Citizen Job/Hobby skills; the book's
// full Citizen Skills-and-Knowledges table (Book 1 p.78) is deferred.
var citizenJobs = []string{
	"Admin", "Broker", "Trader", "Computer", "Steward", "Liaison", "Counsellor", "Driver",
}

// runCitizenTerm resolves a Citizen's term (Book 1 p.78). Citizen Life replaces
// Risk & Reward: a benign 2D roll at or under the Controlling Characteristic —
// success grants a Job or Hobby skill, failure does nothing (no injury). The
// character always survives the career step; only aging can end it.
func runCitizenTerm(
	r *dice.Roller,
	p Policy,
	c *Character,
	run *careerRun,
	career Career,
	cc Characteristic,
) TermOutcome {
	if c.Check(r, cc, 2, 0).Success {
		awardCitizenLife(p, c, run)
	}

	awardSkills(r, p, c, career)

	return Ongoing
}

// runFameTerm resolves an Entertainer's term (Book 1 p.77). At the start of the
// term, events shift Fame by a Flux roll; if Fame increases the character gains
// Talent +1 and two extra skill rolls. There is no injury; only aging can end
// the career. (The optional second and third Flux rolls are deferred.)
func runFameTerm(
	r *dice.Roller,
	p Policy,
	c *Character,
	run *careerRun,
	career Career,
) TermOutcome {
	elig := career.EligPerTerm
	// Term 1 keeps the initial 2D Fame/Talent; the "+F" Flux begins at Term 2
	// (Book 1 p.77). run.terms is 0 on the first served term.
	if run.terms > 0 {
		before := c.Fame

		c.Fame = max(c.Fame+r.Flux(), 0)
		if c.Fame > before {
			c.Talent++
			elig += 2 // "If Fame Increases: 2 [skills] and Talent+1"
		}
	}

	awardSkillsN(r, p, c, career, elig)

	return Ongoing
}

// masterpieceMinimum is the fewest Master Points that allow a Masterpiece
// attempt (Book 1 p.75).
const masterpieceMinimum = 40

// masterpieceValue returns a Masterpiece's sale value (Book 1 p.75):
// Cr150,000 plus Cr10,000 per Master Point over 40, doubled for a Perfect
// Masterpiece (55+ Master Points).
func masterpieceValue(points int) int {
	v := 150_000 + (points-masterpieceMinimum)*10_000
	if points >= 55 {
		v *= 2
	}

	return v
}

// courierElig is the skill eligibility of a Scout's Courier duty (Book 1 p.79);
// Explorer duty grants the career's full EligPerTerm.
const courierElig = 4

// runCraftsmanTerm resolves a Craftsman's term (Book 1 p.75). The Craftsman
// attempts a Masterpiece instead of Risk & Reward: Master Points total the
// Controlling Characteristic, the Craftsman skill, and up to five other skills
// at level 6+. With at least 40 points a 9D roll at or under the total creates a
// Masterpiece; otherwise the attempt fails. Either way the Craftsman skill rises
// +1 (learning); success grants three extra skill rolls, failure one. There is
// no injury.
func runCraftsmanTerm(
	r *dice.Roller,
	p Policy,
	c *Character,
	career Career,
	cc Characteristic,
) TermOutcome {
	points := c.Score(
		cc,
	) + c.Skills.Level(
		"Craftsman",
	) + c.Skills.TopLevels(
		5,
		6,
		"Craftsman",
		"Language",
	)
	elig := career.EligPerTerm

	if points >= masterpieceMinimum && r.Dice(9) <= points {
		c.Masterpieces++
		c.MasterpieceValue += masterpieceValue(points)
		elig += 3
	} else {
		elig++
	}

	c.Skills.Raise("Craftsman", 1) // learning from the work, success or failure
	awardSkillsN(r, p, c, career, elig)

	return Ongoing
}

// runPoliticsTerm resolves a Functionary's term (Book 1 p.87). Office Politics
// replaces Risk & Reward with two unmodified rolls against the Controlling
// Characteristic: a failed Risk ends the career (job loss — the character
// musters out normally, uninjured), and a successful Reward is a promotion. It
// returns MusteredOut on a lost job, otherwise Ongoing.
func runPoliticsTerm(
	r *dice.Roller,
	p Policy,
	c *Character,
	run *careerRun,
	career Career,
	cc Characteristic,
) TermOutcome {
	riskKept := c.Check(r, cc, 2, 0).Success

	elig := career.EligPerTerm
	if c.Check(r, cc, 2, 0).Success &&
		run.rank < len(career.EnlistedRanks) {
		promoteRank(c, run, career.EnlistedRanks)

		elig++ // one extra skill on promotion (Book 1 p.82)
	}

	awardSkillsN(r, p, c, career, elig)

	if !riskKept {
		return MusteredOut
	}

	return Ongoing
}

// runIntrigueTerm resolves a Noble's term (Book 1 p.85). Return & Intrigue
// replaces Risk & Reward, with no injury. A Noble in Exile rolls Return (success
// ends the Exile); one not in Exile rolls Intrigue (failure begins an Exile;
// success offers an Elevation — a roll-high check of 2D at or over Social
// Standing that, on success, raises Soc by 1 and awards a Land Grant).
func runIntrigueTerm(
	r *dice.Roller,
	p Policy,
	c *Character,
	run *careerRun,
	career Career,
	cc Characteristic,
) TermOutcome {
	elevated := false

	switch {
	case run.exiled:
		if c.Check(r, cc, 2, 0).Success {
			run.exiled = false // Return from Exile
		}
	case c.Check(r, cc, 2, 0).Success:
		if r.Dice(2) >= c.Score(Social) && c.scores[Social] < maxCharacteristic {
			c.scores[Social]++ // Elevation to the next Noble rank
			c.LandGrants++
			elevated = true
		}
	default:
		run.exiled = true // Exile
	}

	elig := career.EligPerTerm
	if elevated {
		elig += 2 // "When Elevated 2" (Book 1 p.85)
	}

	awardSkillsN(r, p, c, career, elig)

	return Ongoing
}

// A schemeValue is a Rogue Scheme's payoff (Book 1 p.84): a credit value or,
// for the Scout and Merchant schemes, one Ship Share.
type schemeValue struct {
	credits int
	share   bool
}

// rogueSchemes is the Rogue Schemes table (Book 1 p.84), indexed by Flux + 6
// (Flux runs -6..+6). A raw Flux roll reaches only -5..+5; the ±1 after-roll
// modification that extends it to the -6/+6 rows is deferred, as is a Rogue's
// option to select any previous career in place of the roll.
var rogueSchemes = [13]schemeValue{
	{credits: 200_000}, // -6 Craftsman
	{credits: 100_000}, // -5 Scholar
	{credits: 300_000}, // -4 Entertainer
	{credits: 50_000},  // -3 Citizen
	{share: true},      // -2 Scout
	{share: true},      // -1 Merchant
	{credits: 100_000}, //  0 Spacer
	{credits: 50_000},  // +1 Soldier
	{credits: 100_000}, // +2 Agent
	{credits: 100_000}, // +3 Rogue
	{credits: 500_000}, // +4 Noble
	{credits: 50_000},  // +5 Marine
	{credits: 100_000}, // +6 Functionary
}

// Rogue skill eligibility (Book 1 p.84 B block): base Per Term 2, plus 4 for a
// Successful Scheme (Risk held) or 1 for a Failed Scheme (Risk lost); a term
// served in prison instead grants 2 In-Prison skills and nothing else.
const (
	rogueSuccessElig = 6 // Per Term 2 + Successful Scheme 4
	rogueFailElig    = 3 // Per Term 2 + Failed Scheme 1
	roguePrisonElig  = 2 // In Prison 2 (Personal or Academic column only, no Term or Scheme skills)
)

// runRogueTerm resolves a Rogue's term (Book 1 p.84). The Rogue masterminds a
// Scheme — a plan to amass wealth at others' expense — rolled from the Rogue
// Schemes table for a Value V. Risk and Reward are two rolls against the
// Controlling Characteristic (Mod +Terms; a Risk roll of 12 always fails): a
// successful Reward pays V x (1 + CC - R + Mods), halved when the Risk failed,
// while a failed Risk sends the Rogue to prison next term and earns Fame +1
// (actually Infamy). A prison term draws only In-Prison skills (Personal or
// Academic) and no Scheme payoff.
//
// Deferred: the exact prison sentence (0-4 years at the start of the next term
// is simplified to "any positive sentence costs the following term"), the ±1
// Flux modification, and selecting a previous career in place of the roll. The
// Scheme carries no injury — only aging can end the career.
func runRogueTerm(
	r *dice.Roller,
	p Policy,
	c *Character,
	run *careerRun,
	career Career,
	cc Characteristic,
) TermOutcome {
	if run.inPrison {
		run.inPrison = false

		awardPrisonSkills(r, p, c, career, roguePrisonElig)

		return Ongoing
	}

	scheme := rogueSchemes[dice.FluxIndex(r.Flux())]
	ccVal := c.Score(cc)
	riskMod := p.RiskMod(*c, ccVal) // caution (+), bravery (-), or 0

	// "Mod +Terms": experience eases both rolls; Caution/Bravery flips sign for
	// the Reward (Book 1 p.84, "opposite sign Mods").
	risk := r.Resolve(dice.Check{Dice: 2, Target: ccVal + riskMod + run.terms})
	riskOK := career.held(risk)
	rewardMods := -riskMod + run.terms

	reward := r.Resolve(dice.Check{Dice: 2, Target: ccVal + rewardMods})
	if career.held(reward) {
		payScheme(c, scheme, ccVal+rewardMods, reward.Roll, riskOK)
	}

	elig := rogueSuccessElig

	if !riskOK {
		c.Fame++ // Infamy
		elig = rogueFailElig

		negMods := 0
		if riskMod < 0 {
			negMods = -riskMod
		}

		if min(max(negMods+r.Flux(), 0), 4) > 0 { // a positive sentence means prison next term
			run.inPrison = true
		}
	}

	awardSkillsN(r, p, c, career, elig)

	return Ongoing
}

// payScheme applies a successful Rogue Scheme Reward (Book 1 p.84): a Ship
// Share Scheme grants one share; a credit Scheme pays V x (1 + CC - R + Mods),
// which — since the Reward succeeded (R <= rewardTarget = CC + Mods) — is at
// least V, and is halved when the Risk failed.
func payScheme(c *Character, s schemeValue, rewardTarget, rewardRoll int, riskOK bool) {
	if s.share {
		c.ShipShares++

		return
	}

	payoff := s.credits * (1 + rewardTarget - rewardRoll)
	if !riskOK {
		payoff /= 2 // "Payoff (if any) is halved"
	}

	c.Credits += payoff
}

// awardPrisonSkills grants a Rogue's In-Prison skill rolls (Book 1 p.84), drawn
// only from the Personal and Academic columns (grid columns 0-1). The engine, not
// the Policy, picks between the two — Policy.ChooseSkillColumn is deliberately not
// consulted, since a policy that named any other column would have to be overruled
// anyway. The policy still resolves choice cells, via applyCell.
func awardPrisonSkills(r *dice.Roller, p Policy, c *Character, career Career, n int) {
	// In-Prison skills come from Personal (col 0) or Academic (col 1) only (Book 1
	// p.84). Prefer Academic when the character has a Major/Minor there worth
	// raising; otherwise the always-productive Personal column (characteristic
	// bumps), so every prison roll lands on a real award.
	const personal, academic = 0, 1

	col := personal
	if productive, _ := columnScore(*c, career.Skills[academic]); productive {
		col = academic
	}

	for range n {
		applyCell(p, c, career.Skills[col][r.Die()-1])
	}
}

// undercoverAssignments maps an Agent's Undercover Assignment roll — die A
// (1-3) then die B (1-6) — to the career whose skill tables the Agent draws
// from that term (Book 1 p.83). "Army" is the Soldier, "Navy" the Spacer; the
// Enlisted/Officer split and the specific-skill C column are flavor the Agent
// overrides by selecting the skill.
var undercoverAssignments = [4][7]CareerID{
	1: {1: Soldier, 2: Soldier, 3: Marine, 4: Marine, 5: Spacer, 6: Spacer},
	2: {1: Scholar, 2: Scholar, 3: Entertainer, 4: Entertainer, 5: Citizen, 6: Citizen},
	3: {1: Merchant, 2: Merchant, 3: Scout, 4: Scout, 5: Noble, 6: Functionary},
}

// undercoverAssignment rolls an Agent's Undercover Assignment career (Book 1 p.
// 83): die A rerolled until it is 1-3, then die B (1-6).
func undercoverAssignment(r *dice.Roller) CareerID {
	a := r.Die()
	for a > 3 {
		a = r.Die()
	}

	return undercoverAssignments[a][r.Die()]
}

// awardUndercover grants an Agent's term skills (Book 1 p.83): the two Per-Term
// skills from the Agent's own tables, one Undercover skill selected from the
// tables of a rolled Undercover Assignment career, and — on a Successful Mission
// (a held Risk) — four more. The "select (not roll)" Undercover skill is modelled
// as the first skill of a policy-chosen column of the borrowed grid.
func awardUndercover(r *dice.Roller, p Policy, c *Character, career Career, missionOK bool) {
	borrowed := CareerByID(undercoverAssignment(r)).Skills
	col := min(max(p.ChooseSkillColumn(*c, borrowed), 0), len(borrowed)-1)
	applyCell(p, c, borrowed[col][0]) // Undercover 1: the selected skill

	elig := career.EligPerTerm
	if missionOK {
		elig += 4 // Successful Mission
	}

	awardSkillsN(r, p, c, career, elig)
}

// awardCitizenLife applies one Citizen Life success on the Job/Hobby schedule
// (Book 1 p.78): the 1st success sets the Job at level 4, the 2nd sets the
// Hobby at level 2, and later successes alternate Job/Hobby at +1 (odd = Job,
// even = Hobby). Job and Hobby, once chosen, do not change.
func awardCitizenLife(p Policy, c *Character, run *careerRun) {
	run.citizenWins++
	switch {
	case run.citizenWins == 1:
		run.job = p.ChooseSkill(*c, citizenJobs)
		c.Skills.Raise(run.job, 4)
	case run.citizenWins == 2:
		run.hobby = p.ChooseSkill(*c, without(citizenJobs, run.job))
		c.Skills.Raise(run.hobby, 2)
	case run.citizenWins%2 == 1:
		c.Skills.Raise(run.job, 1)
	default:
		c.Skills.Raise(run.hobby, 1)
	}
}
