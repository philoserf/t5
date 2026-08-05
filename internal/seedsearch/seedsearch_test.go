package seedsearch

import "testing"

func TestFindReturnsFirstMatch(t *testing.T) {
	got := Find(t, 100, "a seed divisible by 7", func(seed uint64) bool {
		return seed%7 == 0
	})

	if got != 7 {
		t.Errorf("Find() = %d, want 7 (the first match, not just any match)", got)
	}
}

func TestFindSearchesFromOne(t *testing.T) {
	var seen []uint64

	Find(t, 5, "always", func(seed uint64) bool {
		seen = append(seen, seed)

		return seed == 3
	})

	want := []uint64{1, 2, 3}
	if len(seen) != len(want) {
		t.Fatalf("visited seeds %v, want %v", seen, want)
	}

	for i, s := range want {
		if seen[i] != s {
			t.Errorf("visited seeds %v, want %v", seen, want)

			break
		}
	}
}
