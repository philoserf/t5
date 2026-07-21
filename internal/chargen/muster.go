package chargen

import (
	"slices"

	"github.com/philoserf/t5/internal/dice"
)

// Muster-out machinery for the term engine (Book 1 pp. 67-70): the benefit
// tables, the per-term award rolls, and their Money/Benefit column DMs. Split
// from career.go (#331); all one package.

// A MusterDM selects the die modifier a career's Benefit muster-out column adds
// to each 1D roll (Book 1 pp. 67-70, each career page's muster DM line).
type MusterDM int

// Sources of the muster-out Benefit-column die modifier.
const (
	DMNone        MusterDM = iota // no modifier (the zero value)
	DMTerms                       // + terms served
	DMOfficerRank                 // + rank, only while on the officer track (armed forces, Merchant)
	DMRank                        // + rank on any track, ladder numbered from 1 — no career uses it yet
	DMRankF0                      // + rank on a ladder numbered from 0 (the Functionary; see below)
	DMFameHalf                    // + Fame/2 (the Scout)
	DMCommends                    // + Commendations (the Agent)
)

// benefitDM returns the value of a Benefit-column muster DM for a character.
//
// A rank DM is the rank *number the career prints*, and careers number their
// ladders differently. Most start at 1 — the Scholar, say, begins at Scholar1
// (Book 1 p.65, "Scholars begin with formal rank (Scholar = Scholar1)") — which
// is the shape DMRank encodes. The Functionary's ladder runs F0 Clerk … F8
// Secretary (p.87), so its first rung is 0 — a Clerk's "+Officer Rank" muster DM
// is +0, not +1. rec.Rank is always a 1-based ladder index, so DMRankF0
// subtracts the difference.
//
// DMRank itself is unused: no career sets it. On p.71 a career's two DMs are
// printed Money-then-Benefits, and the rank ones land on Money — the Scholar's
// pair is "+ Scholar Rank +Terms", i.e. Money +Scholar Rank and Benefits +Terms,
// which is why ScholarCareer.BenefitDM is DMTerms. The Benefits-column rank DMs
// are either officer-track-only (DMOfficerRank) or 0-based (the Functionary's
// DMRankF0), so DMRank waits for a career that prints a plain 1-based one.
func benefitDM(dm MusterDM, c Character, rec CareerRecord) int {
	switch dm {
	case DMTerms:
		return rec.Terms
	case DMOfficerRank:
		if rec.Officer {
			return rec.Rank
		}

		return 0
	case DMRank:
		return rec.Rank
	case DMRankF0:
		return max(rec.Rank-1, 0)
	case DMFameHalf:
		return c.Fame / 2
	case DMCommends:
		return rec.Commendations
	default:
		return 0
	}
}

// A MusterColumn selects which column of a muster-out row a benefit comes from.
type MusterColumn int

// Muster-out table columns.
const (
	MoneyColumn MusterColumn = iota
	BenefitColumn
)

// A BenefitKind identifies a mustering-out award.
type BenefitKind int

// Muster-out benefit kinds.
const (
	Cash         BenefitKind = iota // Value credits
	CharBump                        // +Value to characteristic Char
	FameBump                        // +Value to Fame (the Entertainer's "Fame +1")
	Named                           // a named benefit (Ship Share, TAS Fellowship, …)
	Knighted                        // a Knighthood: raises Soc (Book 1 p.68); Name is still shown
	PensionX2                       // doubles the character's Pension (Book 1 pp.68-69)
	RetirementX2                    // doubles the character's Retirement Pay (Book 1 pp.68-69)
)

// A Benefit is one mustering-out award.
type Benefit struct {
	Kind  BenefitKind
	Value int
	Char  Characteristic
	Name  string
}

// A MusterRow is one row of a muster-out table: a Money award and a Benefit
// award; the character takes one per roll.
type MusterRow struct {
	Money   Benefit
	Benefit Benefit
}

// A MusterTable is a career's mustering-out table, indexed 1-12 by the muster
// roll (1D plus a column DM — see MusterOut).
type MusterTable [13]MusterRow // index 1-12 used

// MusterOut resolves a character's mustering-out benefits (Book 1 pp. 67-70).
// The character rolls once per term served (doubled when disabled); each roll is
// 1D plus the column DM. The Money column adds +Terms; the Benefit column adds
// the career's BenefitDM. The policy chooses the column.
func MusterOut(r *dice.Roller, p Policy, c *Character, rec CareerRecord, career Career) {
	rolls := musterRollCount(*c, rec)
	for range rolls {
		col := p.MusterColumn(*c, rec)

		fullDM := rec.Terms // Money column DM
		if col == BenefitColumn {
			fullDM = benefitDM(career.BenefitDM, *c, rec)
		}

		award := rollMusterAward(r, p, career, col, fullDM)
		// A result duplicating an unusable named benefit (a second Wafer Jack) may
		// be rerolled until different (Book 1 p.69). The cap keeps a degenerate
		// table from looping; if it is hit with the award still a duplicate, the
		// roll grants nothing rather than a useless repeat.
		for tries := 0; isDuplicateBenefit(*c, award) && tries < musterRerollCap; tries++ {
			award = rollMusterAward(r, p, career, col, fullDM)
		}

		if !isDuplicateBenefit(*c, award) {
			applyBenefit(c, award, career, rec)
		}
	}

	// The career's automatic awards land last, after the rolled ones. Order is
	// not a rule — the page prints them beside the table, not in it — but taking
	// them last keeps an automatic named award out of the rolled awards'
	// duplicate check, so adding one cannot shift a single die.
	if career.AutoBenefits != nil {
		for _, award := range career.AutoBenefits(rec) {
			applyBenefit(c, award, career, rec)
		}
	}
}

// musterRerollCap bounds duplicate-benefit rerolls (Book 1 p.69's "until a
// different benefit is received") so a table of all-identical benefits cannot
// loop forever.
const musterRerollCap = 12

// rollMusterAward rolls one mustering-out award: 1D plus a column DM, read off
// the chosen column. The DM is optional and partial (Book 1 p.68), so a policy
// that randomizes it selects any value from 0 to the full DM; otherwise the full
// DM applies.
func rollMusterAward(
	r *dice.Roller,
	p Policy,
	career Career,
	col MusterColumn,
	fullDM int,
) Benefit {
	dm := fullDM
	if fullDM > 0 && p.RandomizeMusterDM() {
		dm = r.Index(fullDM + 1)
	}

	row := min(max(r.Die()+dm, 1), 12)
	if col == BenefitColumn {
		return career.MusterOut[row].Benefit
	}

	return career.MusterOut[row].Money
}

// rerollableDuplicates are the named benefits that are useless when held twice
// (Book 1 p.69's "unwanted or unusable" examples — Wafer Jack, TAS Member — plus
// Life Insurance). Others accumulate, so a repeat is kept, not rerolled: Ship
// Share, Forbidden Knowledge, passages, and StarPass all stack, and a second
// Knighthood still grants Soc +1 (p.68, applyKnighthood), so it too is applied
// rather than rerolled. Deciding which benefits stack is a judgment the book,
// which lists Knighthood as a reroll example yet gives it a repeat effect, leaves
// genuinely open.
var rerollableDuplicates = map[string]bool{
	"Wafer Jack":     true,
	"TAS Fellowship": true,
	"Life Insurance": true,
}

// isDuplicateBenefit reports whether an award repeats a single-instance named
// benefit the character already holds. Cash, characteristic bumps, Fame, and the
// pension doublings carry no name and never count as duplicates.
func isDuplicateBenefit(c Character, b Benefit) bool {
	return rerollableDuplicates[b.Name] && slices.Contains(c.Benefits, b.Name)
}

// musterRollCount is the number of mustering-out rolls (Book 1 p.67): one per
// term served, doubled when disabled, plus one per Commendation and one if Fame
// is 19+. The book's "per MCG/SEH medal" extra rolls are deferred — the engine
// tracks a flat Medal count, not the specific top-tier medals the rule keys on.
func musterRollCount(c Character, rec CareerRecord) int {
	rolls := rec.Terms
	if rec.Outcome == Disabled {
		rolls *= 2
	}

	rolls += rec.Commendations
	if c.Fame >= 19 {
		rolls++
	}

	return rolls
}

// applyBenefit applies one mustering-out award. career and rec supply the
// context a Knighthood needs (armed-forces careers knight only officers).
func applyBenefit(c *Character, b Benefit, career Career, rec CareerRecord) {
	switch b.Kind {
	case Cash:
		c.Credits += b.Value
	case CharBump:
		// A Characteristic Improvement that would raise a characteristic above 15 is
		// lost, not clamped (Book 1 p.68: "If a benefit elevates a characteristic
		// above 15, that benefit is lost"). Clamping quietly banked the partial gain.
		if c.scores[b.Char]+b.Value <= maxCharacteristic {
			c.scores[b.Char] += b.Value
		}
	case FameBump:
		c.Fame += b.Value
	case Named:
		c.Benefits = append(c.Benefits, b.Name)
	case Knighted:
		applyKnighthood(c, career, rec)
		c.Benefits = append(c.Benefits, b.Name) // the title still shows on the sheet
	case PensionX2:
		c.pensionDoublings++
	case RetirementX2:
		c.retirementDoublings++
	}
}

// applyKnighthood raises Social Standing per Book 1 p.68: a Knighthood raises
// any Soc to B (11), or +1 if already 11+. In the armed forces (the Branch/Ops
// careers Spacer/Soldier/Marine) Knighthood is officer-only; a non-officer
// receives Soc +1 instead.
func applyKnighthood(c *Character, career Career, rec CareerRecord) {
	if career.BranchOps != nil && !rec.Officer {
		c.scores[Social] = min(c.scores[Social]+1, maxCharacteristic)

		return
	}

	if c.scores[Social] < 11 {
		c.scores[Social] = 11
	} else {
		c.scores[Social] = min(c.scores[Social]+1, maxCharacteristic)
	}
}
