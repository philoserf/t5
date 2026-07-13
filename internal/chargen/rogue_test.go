package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// TestGoldenRogue traces a complete two-term Rogue whose Schemes both succeed,
// exercising the fixed-CC seam and the Scheme payoff end-to-end. The character
// is all 7s, so the policy's fixed CC (the first controlling characteristic,
// Strength) and the qualify target (best of all six) coincide at 7. Every roll
// is 3,4 (= 7) unless noted.
func TestGoldenRogue(t *testing.T) {
	seq := []int{
		// UPP: six 2D, each 7 -> "777777"; qualify target = best of all six = 7.
		3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4,
		3, 4, // qualify: 7 <= 7, enters the career
		// Term 1: fixed CC = Strength (7), 0 prior terms.
		3, 3, // Scheme Flux = 0 -> Spacer, Cr100,000
		3, 4, // Risk 7 <= 7 (CC 7) held -> Successful Scheme
		3, 4, // Reward 7 <= 7 (CC 7); R = 7 -> payoff 100,000 x (1+7-7) = Cr100,000
		// Successful Scheme = 6 skill rolls from Space Travel (col 3), row 2 = Pilot.
		2, 2, 2, 2, 2, 2,
		3, 4, // continue (UseCC -> vs Strength 7): 7, policy wants term 2
		// Term 2: fixed CC still Strength (7), 1 prior term (Mods +1).
		3, 3, // Scheme Flux = 0 -> Spacer, Cr100,000
		3, 4, // Risk 7 <= 8 (CC 7 +1 term) held
		3, 4, // Reward 7 <= 8 (CC 7 +1 term); R = 7 -> payoff 100,000 x (1+8-7) = Cr200,000
		2, 2, 2, 2, 2, 2, // Pilot x6 again -> 12
		3, 4, // continue: policy stops after term 2
		// Muster out: 2 rolls (2 terms), Benefit column. Rogue's Benefit column
		// also gets the +Terms DM (=2).
		3, // 3 + 2 = row 5 -> End +1 (7 -> 8)
		4, // 4 + 2 = row 6 -> Life Insurance
	}

	c := GenerateCareered(dice.NewScripted(seq...), goldenPolicy{}, worldgen.World{}, RogueCareer)

	if got := c.UPP(); got != "778777" {
		t.Errorf("UPP = %q, want %q (End 7 +1 muster benefit)", got, "778777")
	}
	if c.Age != 26 {
		t.Errorf("Age = %d, want 26 (18 + 2 terms)", c.Age)
	}
	if got := c.Skills.Level("Pilot"); got != 12 {
		t.Errorf("Pilot = %d, want 12 (6 Successful-Scheme rolls x 2 terms)", got)
	}
	if c.Credits != 300_000 {
		t.Errorf("Credits = %d, want 300000 (Cr100,000 + Cr200,000 Scheme payoffs)", c.Credits)
	}
	if c.Fame != 0 {
		t.Errorf("Fame = %d, want 0 (no failed Scheme, so no Infamy)", c.Fame)
	}
	if len(c.Careers) != 1 {
		t.Fatalf("careers = %+v, want one record", c.Careers)
	}
	rec := c.Careers[0]
	if rec.Career != Rogue || rec.Terms != 2 || rec.Outcome != MusteredOut {
		t.Errorf("record = %+v, want Rogue/2 terms/MusteredOut", rec)
	}
	if len(c.Benefits) != 1 || c.Benefits[0] != "Life Insurance" {
		t.Errorf("Benefits = %v, want [Life Insurance]", c.Benefits)
	}
}

// TestRogueSchemeInfamy drives a failed Risk directly: the Reward still lands
// (halved payoff), the Rogue earns Infamy (Fame +1) and a prison sentence, and
// the following prison term grants only In-Prison skills (Personal/Academic).
func TestRogueSchemeInfamy(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 7, 7, 7}}
	run := careerRun{fixed: Strength, fixedChosen: true} // fixed CC already chosen

	// Failure term: Scheme Flux 6-2 = +4 -> Noble, Cr500,000.
	seq := []int{
		6, 2, // Scheme Flux = +4 -> Noble, Cr500,000
		6, 6, // Risk 12 -> always a failure (Failed Scheme)
		3, 4, // Reward 7 <= 7 (CC 7); R = 7 -> payoff 500,000 x (1+7-7) = 500,000, halved = 250,000
		5, 1, // prison Flux = +4 -> positive sentence -> prison next term
		2, 2, 2, // Failed Scheme = 3 skill rolls, Space Travel (col 3) row 2 = Pilot
	}
	if out := runRogueTerm(dice.NewScripted(seq...), goldenPolicy{}, &c, &run, RogueCareer, Strength); out != Ongoing {
		t.Fatalf("failure term outcome = %v, want Ongoing (a Scheme carries no injury)", out)
	}
	if c.Credits != 250_000 {
		t.Errorf("Credits = %d, want 250000 (Cr500,000 payoff halved by the failed Risk)", c.Credits)
	}
	if c.Fame != 1 {
		t.Errorf("Fame = %d, want 1 (Infamy from the failed Scheme)", c.Fame)
	}
	if !run.inPrison {
		t.Errorf("inPrison = false, want true (a positive sentence imprisons next term)")
	}
	if got := c.Skills.Level("Pilot"); got != 3 {
		t.Errorf("Pilot = %d, want 3 (Failed Scheme eligibility)", got)
	}

	// Prison term: only 2 In-Prison skills. This Rogue is uneducated, so Academic
	// (col 1) is unproductive and prison draws from Personal (col 0): row 6 = Soc
	// bump, twice, Soc 7 -> 9. No Scheme payoff, and the prison flag clears.
	if out := runRogueTerm(dice.NewScripted(6, 6), goldenPolicy{}, &c, &run, RogueCareer, Strength); out != Ongoing {
		t.Fatalf("prison term outcome = %v, want Ongoing", out)
	}
	if run.inPrison {
		t.Errorf("inPrison = true, want false (the sentence is served)")
	}
	if got := c.Score(Social); got != 9 {
		t.Errorf("Soc = %d, want 9 (two In-Prison Personal bumps for an uneducated Rogue)", got)
	}
	if c.Credits != 250_000 {
		t.Errorf("Credits = %d, want 250000 unchanged (no Scheme in prison)", c.Credits)
	}
}

func TestRogueFixedCCChosenOnce(t *testing.T) {
	// Under FixedCC the policy selects the CC once (highest score) and it is
	// reused every term.
	c := Character{scores: [count]int{6, 6, 9, 6, 6, 6}} // Endurance highest
	run := careerRun{}
	first := selectCC(DefaultPolicy{}, c, &run, RogueCareer)
	if first != Endurance {
		t.Fatalf("first CC = %v, want Endurance (highest)", first)
	}
	// Even if scores later shift, the fixed choice holds.
	c.scores[Strength] = 15
	if again := selectCC(DefaultPolicy{}, c, &run, RogueCareer); again != Endurance {
		t.Errorf("second CC = %v, want the same fixed Endurance", again)
	}
}
