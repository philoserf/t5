package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// TestPayScheme pins payScheme's formula (Book 1 p.84: a credit Scheme pays
// V x (1 + CC - R), halved when the Risk failed) against literals computed
// independently of the function itself — the case a Copilot review on #370
// correctly flagged: the golden tests below build their expectations by
// CALLING payScheme, so a bug in payScheme's own arithmetic needs coverage
// that doesn't also call it.
func TestPayScheme(t *testing.T) {
	cases := []struct {
		name                  string
		value                 schemeValue
		target, roll          int
		riskOK                bool
		wantCredits, wantSubs int
	}{
		{"roll equals target", schemeValue{credits: 100_000}, 7, 7, true, 100_000, 0},
		{"roll under target pays more", schemeValue{credits: 100_000}, 8, 5, true, 400_000, 0},
		{"failed risk halves the payoff", schemeValue{credits: 500_000}, 7, 7, false, 250_000, 0},
		{"ship-share scheme grants a share, no credits", schemeValue{share: true}, 7, 7, true, 0, 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var c Character

			payScheme(&c, tc.value, tc.target, tc.roll, tc.riskOK)

			if c.Credits != tc.wantCredits {
				t.Errorf("Credits = %d, want %d", c.Credits, tc.wantCredits)
			}

			if c.ShipShares != tc.wantSubs {
				t.Errorf("ShipShares = %d, want %d", c.ShipShares, tc.wantSubs)
			}
		})
	}
}

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

	// Derived by calling payScheme itself rather than hand-tracing its formula
	// into a comment (#357): each term's Spacer Scheme (Cr100,000) pays out at
	// the CC target/roll that term actually rolled, held both times.
	var want Character

	payScheme(&want, schemeValue{credits: 100_000}, 7, 7, true) // term 1: CC 7, roll 7
	payScheme(&want, schemeValue{credits: 100_000}, 8, 7, true) // term 2: CC 7 +1 term, roll 7

	if c.Credits != want.Credits {
		t.Errorf("Credits = %d, want %d (two Spacer Scheme payoffs via payScheme)", c.Credits, want.Credits)
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

// TestRogueBeginsAgainstItsFixedCC locks Book 1 p.84: the Rogue's box reads "To
// Begin CC", and the CC he selects is "then used throughout his career (not just
// in the current Term)". Begin therefore rolls against that one characteristic —
// not against a best-of-all-six qualification set, which would let a stat the
// Rogue never uses buy him into the career.
func TestRogueBeginsAgainstItsFixedCC(t *testing.T) {
	// Str 4 is the policy's pick (lowest, via lowCC); Soc 12 is the best of six.
	c := Character{scores: [count]int{4, 4, 4, 4, 4, 12}, Age: startingAge}
	run := newCareerRun(RogueCareer)

	if beginCareer(dice.NewScripted(4, 4), lowCC{}, &c, &run, RogueCareer) {
		t.Error("Begin 8 > the chosen CC (Str 4) should be refused, even with Soc 12")
	}

	if run.fixed != Strength {
		t.Errorf("fixed CC = %v, want Strength (the policy's pick at Begin)", run.fixed)
	}
	// The same run carries that choice into the terms: no second selection.
	run2 := newCareerRun(RogueCareer)
	if !beginCareer(dice.NewScripted(1, 2), lowCC{}, &c, &run2, RogueCareer) { // 3 <= 4
		t.Fatal("Begin 3 <= Str 4 should enter the career")
	}

	if got := selectCC(lowCC{}, c, &run2, RogueCareer); got != Strength {
		t.Errorf("term CC = %v, want the Strength chosen at Begin", got)
	}
}

// lowCC picks the lowest-scoring available characteristic, so the CC and the
// best-of-six qualification target cannot be confused for each other.
type lowCC struct{ DefaultPolicy }

func (lowCC) ChooseCC(c Character, available []Characteristic) Characteristic {
	best := available[0]
	for _, ch := range available[1:] {
		if c.Score(ch) < c.Score(best) {
			best = ch
		}
	}

	return best
}

// TestRogueTwelveAlwaysFails locks the second footnote under the Rogue's box
// (Book 1 p.84): "But, 12 is always automatic failure." It sits at the same
// level as "*Mod +Terms." beneath all three printed targets — To Begin, Risk &
// Reward, Continue — so it governs every one of them. Each case gives the Rogue
// a CC of 12 so the arithmetic would otherwise pass on a natural 12.
func TestRogueTwelveAlwaysFails(t *testing.T) {
	twelve := Character{scores: [count]int{12, 12, 12, 12, 12, 12}, Age: startingAge}

	t.Run("begin", func(t *testing.T) {
		c := twelve
		run := newCareerRun(RogueCareer)

		if beginCareer(dice.NewScripted(6, 6), DefaultPolicy{}, &c, &run, RogueCareer) {
			t.Error("Begin rolled 12 against CC 12 and entered the career")
		}
	})

	t.Run("risk and reward", func(t *testing.T) {
		c := twelve
		run := newCareerRun(RogueCareer)
		run.fixed, run.fixedChosen = Strength, true

		seq := []int{
			3, 3, // Scheme Flux 0 -> Spacer, Cr100,000
			6, 6, // Risk 12 vs CC 12: auto-failure -> prison track, Infamy
			6, 6, // Reward 12 vs CC 12: auto-failure -> no payoff
			1, 1, // the prison-sentence Flux (0), so no prison term follows
			1, 1, 1, // 3 Failed-Scheme skill rolls
		}
		runRogueTerm(dice.NewScripted(seq...), goldenPolicy{}, &c, &run, RogueCareer, Strength)

		if c.Credits != 0 {
			t.Errorf("Credits = %d, want 0 (a Reward of 12 pays nothing)", c.Credits)
		}

		if c.Fame != 1 {
			t.Errorf("Fame = %d, want 1 (a Risk of 12 earns Infamy)", c.Fame)
		}
	})

	t.Run("continue", func(t *testing.T) {
		c := twelve
		run := newCareerRun(RogueCareer)
		run.fixed, run.fixedChosen = Strength, true
		rec := CareerRecord{Career: Rogue, Terms: 1} // "Mod +Terms" lifts the target to 13

		if continues(dice.NewScripted(6, 6), goldenPolicy{}, c, RogueCareer, rec, &run) {
			t.Error("Continue rolled 12 against a target of 13 and stayed in the career")
		}
	})
}

// TestRogueSchemeInfamy drives a failed Risk directly: the Reward still lands
// (halved payoff), the Rogue earns Infamy (Fame +1) and a prison sentence, and
// the following prison term grants only In-Prison skills (Personal/Academic).
func TestRogueSchemeInfamy(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 7, 7, 7}}
	run := careerRun{fixed: Strength, fixedChosen: true} // fixed CC already chosen

	// Derived by calling payScheme directly (#357) rather than hand-tracing its
	// halving into a comment: Noble Scheme (Cr500,000), CC target/roll both 7,
	// halved for the failed Risk.
	var want Character
	payScheme(&want, schemeValue{credits: 500_000}, 7, 7, false)

	// Failure term: Scheme Flux 6-2 = +4 -> Noble, Cr500,000.
	seq := []int{
		6, 2, // Scheme Flux = +4 -> Noble, Cr500,000
		6, 6, // Risk 12 -> always a failure (Failed Scheme)
		3, 4, // Reward 7 <= 7 (CC 7); R = 7 -> payoff 500,000 x (1+7-7) = 500,000, halved = 250,000
		5, 1, // prison Flux = +4 -> positive sentence -> prison next term
		2, 2, 2, // Failed Scheme = 3 skill rolls, Space Travel (col 3) row 2 = Pilot
	}
	if out := runRogueTerm(
		dice.NewScripted(seq...),
		goldenPolicy{},
		&c,
		&run,
		RogueCareer,
		Strength,
	); out != Ongoing {
		t.Fatalf("failure term outcome = %v, want Ongoing (a Scheme carries no injury)", out)
	}

	if c.Credits != want.Credits {
		t.Errorf(
			"Credits = %d, want %d (Cr500,000 Noble payoff halved by the failed Risk, via payScheme)",
			c.Credits, want.Credits,
		)
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
	if out := runRogueTerm(
		dice.NewScripted(6, 6),
		goldenPolicy{},
		&c,
		&run,
		RogueCareer,
		Strength,
	); out != Ongoing {
		t.Fatalf("prison term outcome = %v, want Ongoing", out)
	}

	if run.inPrison {
		t.Errorf("inPrison = true, want false (the sentence is served)")
	}

	if got := c.Score(Social); got != 9 {
		t.Errorf("Soc = %d, want 9 (two In-Prison Personal bumps for an uneducated Rogue)", got)
	}

	if c.Credits != want.Credits {
		t.Errorf("Credits = %d, want %d unchanged (no Scheme in prison)", c.Credits, want.Credits)
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
