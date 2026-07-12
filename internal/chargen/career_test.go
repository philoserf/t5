package chargen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

// testCareer is a minimal Scout-shaped career for exercising the engine.
var testCareer = Career{
	ID:               Scout,
	Name:             "Scout",
	Qualify:          Qualification{Char: Intelligence},
	CCMode:           RotateCC,
	ControllingChars: []Characteristic{Strength, Dexterity, Endurance, Intelligence},
	Continue:         ContinueRule{UseChar: true, Char: Intelligence},
}

// stopAfter is a policy that continues until a fixed number of terms is served.
type stopAfter struct{ terms int }

func (stopAfter) ChooseCC(_ Character, available []Characteristic) Characteristic {
	return available[0]
}
func (s stopAfter) Continue(_ Character, rec CareerRecord) bool { return rec.Terms < s.terms }

func TestContinueTarget(t *testing.T) {
	c := Character{scores: [count]int{7, 7, 7, 10, 8, 6}}
	if got := (ContinueRule{Fixed: 7, Mod: 2}).target(c); got != 9 {
		t.Errorf("fixed target = %d, want 9", got)
	}
	if got := (ContinueRule{UseChar: true, Char: Intelligence, Mod: -1}).target(c); got != 9 {
		t.Errorf("char target = %d, want 9 (Int 10 - 1)", got)
	}
}

func TestSelectCCRotation(t *testing.T) {
	c := Character{}
	run := careerRun{ccPool: append([]Characteristic(nil), testCareer.ControllingChars...)}
	p := stopAfter{} // ChooseCC returns available[0]
	// The pool drains as each is picked, then refills once exhausted.
	seen := map[Characteristic]int{}
	for i := 0; i < 8; i++ {
		seen[selectCC(p, c, &run, testCareer)]++
	}
	for _, ch := range testCareer.ControllingChars {
		if seen[ch] != 2 {
			t.Errorf("over two cycles %s picked %d times, want 2", ch, seen[ch])
		}
	}
}

func TestSelectCCFixed(t *testing.T) {
	fixed := Career{CCMode: FixedCC, ControllingChars: []Characteristic{Endurance}}
	run := careerRun{}
	for i := 0; i < 3; i++ {
		if got := selectCC(stopAfter{}, Character{}, &run, fixed); got != Endurance {
			t.Fatalf("fixed CC = %v, want Endurance", got)
		}
	}
}

func TestRunCareerTermCount(t *testing.T) {
	// Policy stops after 2 terms; Int 10 makes both Continue rolls (7) succeed.
	c := Character{scores: [count]int{7, 7, 7, 10, 8, 6}, Age: 18}
	RunCareer(dice.NewScripted(3, 4), stopAfter{2}, &c, testCareer)
	if len(c.Careers) != 1 || c.Careers[0].Terms != 2 || c.Age != 26 {
		t.Fatalf("run = %+v age %d, want 2 terms, age 26", c.Careers, c.Age)
	}
	if c.Careers[0].Outcome != MusteredOut {
		t.Errorf("outcome = %v, want MusteredOut", c.Careers[0].Outcome)
	}
}

func TestRunCareerMandatoryContinue(t *testing.T) {
	// Policy wants to stop after 1 term, but a natural 2 forces a second term.
	c := Character{scores: [count]int{7, 7, 7, 10, 8, 6}, Age: 18}
	// Term 1 Continue rolls 2 (mandatory); term 2 Continue rolls 7 (honours the stop).
	RunCareer(dice.NewScripted(1, 1, 3, 4), stopAfter{1}, &c, testCareer)
	if c.Careers[0].Terms != 2 {
		t.Fatalf("mandatory continue: %d terms, want 2", c.Careers[0].Terms)
	}
}

func TestRunCareerStopsAtAging(t *testing.T) {
	// DefaultPolicy serves until aging begins (age 34) — four terms from 18.
	c := Character{scores: [count]int{7, 7, 7, 10, 8, 6}, Age: 18}
	// Three Continue rolls of 7, then at term four an aging check (high rolls,
	// no aging) and a final Continue roll.
	rolls := []int{3, 4, 3, 4, 3, 4, 6, 6, 6, 6, 6, 6, 3, 4}
	RunCareer(dice.NewScripted(rolls...), DefaultPolicy{}, &c, testCareer)
	if c.Careers[0].Terms != 4 || c.Age != 34 {
		t.Fatalf("default policy: %d terms age %d, want 4 terms age 34", c.Careers[0].Terms, c.Age)
	}
}

func TestGenerateCareeredQualify(t *testing.T) {
	// UPP rolls (12 dice), then a qualify roll of 2 (<= Int) succeeds, so the
	// character runs the career; the policy stops immediately.
	upp := []int{4, 4, 4, 4, 4, 4, 4, 4, 5, 5, 3, 3} // six 2D rolls; Int = 8
	seq := append(append([]int{}, upp...), 1, 1 /*qualify 2*/, 3, 4 /*continue*/)
	c := GenerateCareered(dice.NewScripted(seq...), stopAfter{1}, testCareer)
	if len(c.Careers) != 1 {
		t.Fatalf("qualified but no career recorded: %+v", c.Careers)
	}
	// A failed qualify (roll 12 > Int) yields no career.
	seq2 := append(append([]int{}, upp...), 6, 6 /*qualify 12*/)
	c2 := GenerateCareered(dice.NewScripted(seq2...), stopAfter{1}, testCareer)
	if len(c2.Careers) != 0 {
		t.Fatalf("failed qualify but career recorded: %+v", c2.Careers)
	}
}
