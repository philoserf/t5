package chargen

import (
	"testing"
)

// TestRetirementPay covers the armed-forces Retirement Pay (Book 1 p.69): 4+
// combined active-duty terms, Cr2,000/term enlisted or Cr3,000/term officer,
// collectible at the character's current age.
func TestRetirementPay(t *testing.T) {
	cases := []struct {
		name    string
		age     int
		careers []CareerRecord
		want    []Entitlement
	}{
		{
			"enlisted Soldier, 4 terms", 34,
			[]CareerRecord{{Career: Soldier, Terms: 4}},
			[]Entitlement{{"Enlisted Retirement", 8000, 34}},
		},
		{
			"officer Marine, 5 terms", 38,
			[]CareerRecord{{Career: Marine, Terms: 5, Officer: true}},
			[]Entitlement{{"Officer Retirement", 15000, 38}},
		},
		{
			"combined military terms", 34,
			[]CareerRecord{{Career: Soldier, Terms: 2}, {Career: Spacer, Terms: 2}},
			[]Entitlement{{"Enlisted Retirement", 8000, 34}},
		},
		{
			"only 3 terms — no retirement pay", 30,
			[]CareerRecord{{Career: Soldier, Terms: 3}},
			nil,
		},
		{
			"Scout is not armed forces — no retirement pay", 34,
			[]CareerRecord{{Career: Scout, Terms: 5}},
			nil,
		},
		{
			// The rate follows the final military muster status, applied to all terms.
			"mixed service, officer last", 38,
			[]CareerRecord{{Career: Soldier, Terms: 3}, {Career: Marine, Terms: 2, Officer: true}},
			[]Entitlement{{"Officer Retirement", 15000, 38}},
		},
		{
			"mixed service, enlisted last", 38,
			[]CareerRecord{{Career: Marine, Terms: 2, Officer: true}, {Career: Soldier, Terms: 3}},
			[]Entitlement{{"Enlisted Retirement", 10000, 38}},
		},
	}
	for _, tc := range cases {
		c := &Character{Age: tc.age, Careers: tc.careers}
		computeEntitlements(c)

		if !equalEntitlements(c.Entitlements, tc.want) {
			t.Errorf("%s: entitlements = %+v, want %+v", tc.name, c.Entitlements, tc.want)
		}
	}
}

// TestCivilPension covers the Citizen/Functionary pension (Book 1 p.69): from age
// 66, Cr5,000/year for a Citizen, Cr15,000/year for a Functionary (which replaces
// the Citizen pension).
func TestCivilPension(t *testing.T) {
	cases := []struct {
		name    string
		careers []CareerRecord
		want    []Entitlement
	}{
		{
			"Citizen",
			[]CareerRecord{{Career: Citizen, Terms: 1}},
			[]Entitlement{{"Citizen's Pension", 5000, 66}},
		},
		{
			"Functionary",
			[]CareerRecord{{Career: Functionary, Terms: 1}},
			[]Entitlement{{"Functionary's Pension", 15000, 66}},
		},
		{
			"Functionary replaces Citizen",
			[]CareerRecord{{Career: Citizen, Terms: 1}, {Career: Functionary, Terms: 1}},
			[]Entitlement{{"Functionary's Pension", 15000, 66}},
		},
	}
	for _, tc := range cases {
		c := &Character{Careers: tc.careers}
		computeEntitlements(c)

		if !equalEntitlements(c.Entitlements, tc.want) {
			t.Errorf("%s: entitlements = %+v, want %+v", tc.name, c.Entitlements, tc.want)
		}
	}
}

// TestPensionDoubling checks the "Pension x2" stacking (Book 1 p.68): each
// doubling adds one multiple of the base (one x2 = double, two = triple).
func TestPensionDoubling(t *testing.T) {
	c := &Character{Careers: []CareerRecord{{Career: Functionary, Terms: 1}}, pensionDoublings: 2}
	computeEntitlements(c)

	if len(c.Entitlements) != 1 || c.Entitlements[0].PerYear != 45000 {
		t.Errorf("doubled pension = %+v, want Cr45,000/year (15,000 x 3)", c.Entitlements)
	}
}

// TestPensionX2Benefit checks that a "Pension x2" muster award raises the pension
// doubling counter.
func TestPensionX2Benefit(t *testing.T) {
	c := &Character{}
	applyBenefit(c, pensionX2(), FunctionaryCareer, CareerRecord{})

	if c.pensionDoublings != 1 {
		t.Errorf("pensionDoublings = %d, want 1", c.pensionDoublings)
	}
}

// TestRetirementDoubling covers the "Retire x2" muster result (Book 1 p.82, the
// military tables' Money rows 8-9): an officer Marine's Cr3,000/term x 4 terms
// doubles to Cr24,000/year.
func TestRetirementDoubling(t *testing.T) {
	c := &Character{
		Age:                 34,
		Careers:             []CareerRecord{{Career: Marine, Terms: 4, Officer: true}},
		retirementDoublings: 1,
	}
	computeEntitlements(c)

	if len(c.Entitlements) != 1 || c.Entitlements[0].PerYear != 24000 {
		t.Errorf("doubled retirement = %+v, want Cr24,000/year (3,000 x 4 x 2)", c.Entitlements)
	}
}

// TestRetirementX2Benefit checks that a "Retire x2" muster award raises the
// retirement doubling counter.
func TestRetirementX2Benefit(t *testing.T) {
	c := &Character{}
	applyBenefit(c, retirementX2(), MarineCareer, CareerRecord{})

	if c.retirementDoublings != 1 {
		t.Errorf("retirementDoublings = %d, want 1", c.retirementDoublings)
	}
}

// TestIsDuplicateBenefit checks that only single-instance named benefits are
// treated as rerollable duplicates; stackable ones and cash are not.
func TestIsDuplicateBenefit(t *testing.T) {
	c := Character{Benefits: []string{"Wafer Jack", "Ship Share"}}
	if !isDuplicateBenefit(c, named("Wafer Jack")) {
		t.Error("a held Wafer Jack should be a rerollable duplicate")
	}

	if isDuplicateBenefit(c, named("Ship Share")) {
		t.Error("Ship Shares stack — not a duplicate")
	}

	if isDuplicateBenefit(c, named("TAS Fellowship")) {
		t.Error("a benefit not yet held is not a duplicate")
	}

	if isDuplicateBenefit(c, cash(1000)) {
		t.Error("cash never duplicates")
	}
}

// TestRandomizeMusterDMPolicy pins the two policies' DM behavior.
func TestRandomizeMusterDMPolicy(t *testing.T) {
	if !(DefaultPolicy{}).RandomizeMusterDM() {
		t.Error("DefaultPolicy should randomize the muster DM")
	}

	if (goldenPolicy{}).RandomizeMusterDM() {
		t.Error("goldenPolicy should apply the full DM (deterministic)")
	}
}

// TestProfessorsPension covers the tenured Professor's pension (Book 1 p.69): a
// flat Cr10,000/year from age 66 that stacks with other pensions.
func TestProfessorsPension(t *testing.T) {
	// Tenured Scholar -> Professor's Pension.
	c := &Character{Tenured: true, Careers: []CareerRecord{{Career: Scholar, Terms: 4}}}
	computeEntitlements(c)

	if !equalEntitlements(c.Entitlements, []Entitlement{{"Professor's Pension", 10000, 66}}) {
		t.Errorf("tenured Scholar entitlements = %+v, want one Professor's Pension", c.Entitlements)
	}

	// An untenured Scholar gets nothing.
	c = &Character{Careers: []CareerRecord{{Career: Scholar, Terms: 4}}}
	computeEntitlements(c)

	if len(c.Entitlements) != 0 {
		t.Errorf("untenured Scholar entitlements = %+v, want none", c.Entitlements)
	}

	// It stacks with a Functionary pension (Book 1 p.69 allows duplicate entitlements).
	c = &Character{
		Tenured: true,
		Careers: []CareerRecord{{Career: Functionary, Terms: 1}, {Career: Scholar, Terms: 4}},
	}
	computeEntitlements(c)

	want := []Entitlement{{"Functionary's Pension", 15000, 66}, {"Professor's Pension", 10000, 66}}
	if !equalEntitlements(c.Entitlements, want) {
		t.Errorf("Functionary+Professor entitlements = %+v, want both pensions", c.Entitlements)
	}
}

// TestDeadCollectsNothing confirms a character who died in service gets no
// entitlement, even with qualifying service.
func TestDeadCollectsNothing(t *testing.T) {
	c := &Character{Age: 34, Dead: true, Careers: []CareerRecord{{Career: Soldier, Terms: 4}}}
	computeEntitlements(c)

	if len(c.Entitlements) != 0 {
		t.Errorf("dead character entitlements = %+v, want none", c.Entitlements)
	}
}

func equalEntitlements(a, b []Entitlement) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
