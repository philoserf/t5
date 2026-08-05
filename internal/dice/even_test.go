package dice

import "testing"

func TestEvenDist1to9(t *testing.T) {
	// (row, col) -> value, sampled from the Even 1-9 table (p.259).
	cases := []struct {
		row, col, want int
	}{
		{1, 1, 1},
		{1, 6, 3},
		{2, 2, 5},
		{3, 6, 9},
		{4, 1, 1},
		{5, 5, 5},
		{6, 4, 7},
		{6, 6, 9},
	}
	for _, c := range cases {
		got := scripted(c.row, c.col).EvenDist1to9()
		if got != c.want {
			t.Errorf("EvenDist1to9(row=%d,col=%d) = %d, want %d", c.row, c.col, got, c.want)
		}
	}
}

func TestEvenDist1to9Range(t *testing.T) {
	r := NewWithSeed(1)
	for range 5000 {
		if v := r.EvenDist1to9(); v < 1 || v > 9 {
			t.Fatalf("EvenDist1to9 out of range: %d", v)
		}
	}
}

func TestEvenDist0to9(t *testing.T) {
	cases := []struct {
		row, col, want int
	}{
		{1, 1, 0},
		{1, 4, 1},
		{2, 3, 2},
		{2, 6, 3},
		{3, 1, 4},
		{4, 4, 7},
		{5, 6, 9},
	}
	for _, c := range cases {
		got := scripted(c.row, c.col).EvenDist0to9()
		if got != c.want {
			t.Errorf("EvenDist0to9(row=%d,col=%d) = %d, want %d", c.row, c.col, got, c.want)
		}
	}
}

func TestEvenDist0to9RerollsRowSix(t *testing.T) {
	// First (row=6) pair is rerolled; the (row=3,col=1)=4 pair is used.
	got := scripted(6, 2, 3, 1).EvenDist0to9()
	if got != 4 {
		t.Fatalf("EvenDist0to9 with reroll = %d, want 4", got)
	}
}

func TestEvenDist1to10(t *testing.T) {
	// A 0-9 result of 0 (row=1,col=1) becomes 10.
	if got := scripted(1, 1).EvenDist1to10(); got != 10 {
		t.Fatalf("EvenDist1to10 zero case = %d, want 10", got)
	}
	// A non-zero result passes through (row=5,col=6 -> 9).
	if got := scripted(5, 6).EvenDist1to10(); got != 9 {
		t.Fatalf("EvenDist1to10 = %d, want 9", got)
	}

	r := NewWithSeed(7)
	for range 5000 {
		if v := r.EvenDist1to10(); v < 1 || v > 10 {
			t.Fatalf("EvenDist1to10 out of range: %d", v)
		}
	}
}

func TestIndex(t *testing.T) {
	r := NewWithSeed(42)
	// Every result is in range, and over many draws every value appears roughly
	// evenly (the rejection sampling removes the base-6 modulo bias).
	const draws = 12000

	for _, n := range []int{2, 8, 17} {
		counts := make([]int, n)
		for range draws {
			v := r.Index(n)
			if v < 0 || v >= n {
				t.Fatalf("Index(%d) = %d, out of [0,%d)", n, v, n)
			}

			counts[v]++
		}

		expected := draws / n
		for v, c := range counts {
			if c < expected*3/4 || c > expected*5/4 {
				t.Errorf(
					"Index(%d) value %d appeared %d times, want ~%d (uneven)",
					n,
					v,
					c,
					expected,
				)
			}
		}
	}
}

func TestIndexPanicsOnNonPositive(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("Index(0) should panic")
		}
	}()

	NewWithSeed(1).Index(0)
}
