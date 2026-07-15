package chargen

// Entitlements are the recurring annual income a character carries out of their
// careers (Book 1 p. 69). Retirement Pay (Soldier/Spacer/Marine on active duty
// for 4+ combined terms) is collectible immediately — retirement is the act of
// leaving the service to adventure. Pensions (Citizen/Functionary) begin at Life
// Stage 9 Retirement (age 66). A "Pension x2" or "Retirement x2" muster result
// doubles the base per the p. 68 stacking (the first x2 doubles it, the second
// triples it, and so on: multiplier = 1 + the number of doublings).
//
// Deferred: the Reserve Pension (the engine models no Reserves/OTC/NOTC) and the
// Professor's Pension (Scholar tenure is not modelled).

// retirementAge is Life Stage 9, when Citizen/Functionary pensions begin.
const retirementAge = 66

// An Entitlement is one stream of recurring annual income (Book 1 p. 69).
type Entitlement struct {
	Source  string // e.g. "Officer Retirement", "Functionary's Pension"
	PerYear int    // credits per year
	FromAge int    // the age at which collection begins
}

// computeEntitlements derives a character's recurring income from their career
// records and any "x2" muster results. A character who died in service collects
// nothing.
func computeEntitlements(c *Character) {
	c.Entitlements = nil
	if c.Dead {
		return
	}
	if e, ok := retirementPay(c); ok {
		c.Entitlements = append(c.Entitlements, e)
	}
	if e, ok := civilPension(c); ok {
		c.Entitlements = append(c.Entitlements, e)
	}
}

// retirementPay is the armed-forces Retirement Pay (Book 1 p. 69): a Soldier,
// Spacer, or Marine with 4+ combined active-duty terms collects, per term,
// Cr3,000 if they mustered as an officer, else Cr2,000. Collection begins at once.
// The rate follows the character's final armed-forces muster status (the book's
// "musters out as an Officer"), applied to their total combined terms.
func retirementPay(c *Character) (Entitlement, bool) {
	terms, officer := 0, false
	for _, rec := range c.Careers {
		if CareerByID(rec.Career).BranchOps != nil { // an armed-forces career
			terms += rec.Terms
			officer = rec.Officer // the most recent armed-forces career's status
		}
	}
	if terms < 4 {
		return Entitlement{}, false
	}
	rate, source := 2000, "Enlisted Retirement"
	if officer {
		rate, source = 3000, "Officer Retirement"
	}
	return Entitlement{
		Source:  source,
		PerYear: rate * terms * (1 + c.retirementDoublings),
		FromAge: c.Age,
	}, true
}

// civilPension is the Citizen's Pension (Book 1 p. 69): a Functionary of 1+ terms
// collects Cr15,000/year, or a Citizen of 1+ terms Cr5,000/year (the Functionary
// pension replaces the Citizen one). It begins at Retirement, age 66.
func civilPension(c *Character) (Entitlement, bool) {
	var functionary, citizen bool
	for _, rec := range c.Careers {
		switch rec.Career {
		case Functionary:
			functionary = true
		case Citizen:
			citizen = true
		}
	}
	base, source := 0, ""
	switch {
	case functionary:
		base, source = 15000, "Functionary's Pension"
	case citizen:
		base, source = 5000, "Citizen's Pension"
	default:
		return Entitlement{}, false
	}
	return Entitlement{
		Source:  source,
		PerYear: base * (1 + c.pensionDoublings),
		FromAge: retirementAge,
	}, true
}
