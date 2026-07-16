package shipcombat

import "testing"

func TestAgility(t *testing.T) {
	if got := Agility(6, 2); got != 4 {
		t.Errorf("Agility(6,2) = %d, want 4", got)
	}
	if got := Agility(4, 4); got != 0 {
		t.Errorf("Agility(4,4) = %d, want 0", got)
	}
}

func TestRammingHits(t *testing.T) {
	// Book 2 p.197 Joshua has 21 compartments -> 441 ramming Hits.
	if got := RammingHits(21); got != 441 {
		t.Errorf("RammingHits(21) = %d, want 441", got)
	}
}

// TestRangeChangeRounds golden-locks the Book 2 p.200 Space Combat Range Bands
// grid: transit rounds by band and thrust.
func TestRangeChangeRounds(t *testing.T) {
	cases := []struct {
		band, gs, want int
	}{
		{8, 1, 18},
		{8, 6, 7},
		{8, 9, 6}, // 7-9G share the last column
		{7, 1, 15},
		{6, 1, 5},
		{6, 4, 2},
		{5, 1, 2},
		{5, 2, 1},
		{4, 3, 1},
		{1, 1, 1},
	}
	for _, c := range cases {
		got, tactical := RangeChangeRounds(c.band, c.gs)
		if !tactical || got != c.want {
			t.Errorf(
				"RangeChangeRounds(band %d, %dG) = %d,%v, want %d,true",
				c.band,
				c.gs,
				got,
				tactical,
				c.want,
			)
		}
	}
	// Deep-space and long bands transit in hours/days, not tactical rounds.
	if _, tactical := RangeChangeRounds(11, 1); tactical {
		t.Errorf("band 11 should not be a tactical (round-measured) change")
	}
	if _, tactical := RangeChangeRounds(0, 1); tactical {
		t.Errorf("band 0 is the target, no change below it")
	}
}
