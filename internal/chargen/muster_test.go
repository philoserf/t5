package chargen

import "testing"

// TestApplyKnighthood locks the Book 1 p.68 rule: a Knighthood raises any Soc to
// 11 (or +1 if already 11+); in the armed forces it is officer-only, and a
// non-officer receives Soc +1 instead. No golden reaches a Knighthood muster row.
func TestApplyKnighthood(t *testing.T) {
	// Non-armed-forces: raise any Soc to 11.
	c := Character{scores: [count]int{7, 7, 7, 7, 7, 7}}
	applyKnighthood(&c, ScoutCareer, CareerRecord{})
	if got := c.Score(Social); got != 11 {
		t.Errorf("Scout knight Soc 7 -> %d, want 11", got)
	}
	// Already 11+: +1.
	c2 := Character{scores: [count]int{7, 7, 7, 7, 7, 12}}
	applyKnighthood(&c2, ScoutCareer, CareerRecord{})
	if got := c2.Score(Social); got != 13 {
		t.Errorf("Scout knight Soc 12 -> %d, want 13", got)
	}
	// Armed forces (SoldierCareer has Branch/Ops): an officer is knighted to 11.
	off := Character{scores: [count]int{7, 7, 7, 7, 7, 7}}
	applyKnighthood(&off, SoldierCareer, CareerRecord{Officer: true})
	if got := off.Score(Social); got != 11 {
		t.Errorf("Soldier officer knight Soc 7 -> %d, want 11", got)
	}
	// Armed forces non-officer: Soc +1 only.
	enl := Character{scores: [count]int{7, 7, 7, 7, 7, 7}}
	applyKnighthood(&enl, SoldierCareer, CareerRecord{Officer: false})
	if got := enl.Score(Social); got != 8 {
		t.Errorf("Soldier non-officer knight Soc 7 -> %d, want 8 (+1 only)", got)
	}
}

// TestMusterRollCount locks Book 1 p.67: one roll per term, doubled when
// disabled, plus one per Commendation and one at Fame 19+.
func TestMusterRollCount(t *testing.T) {
	base := CareerRecord{Terms: 3}
	cases := []struct {
		name string
		c    Character
		rec  CareerRecord
		want int
	}{
		{"terms only", Character{}, base, 3},
		{"plus commendations", Character{Commendations: 2}, base, 5},
		{"plus fame 19", Character{Fame: 19}, base, 4},
		{"below fame 19", Character{Fame: 18}, base, 3},
		{"disabled doubles terms only", Character{Commendations: 1}, CareerRecord{Terms: 2, Outcome: Disabled}, 5},
	}
	for _, tc := range cases {
		if got := musterRollCount(tc.c, tc.rec); got != tc.want {
			t.Errorf("%s: musterRollCount = %d, want %d", tc.name, got, tc.want)
		}
	}
}
