package dice

import "testing"

func TestManyDice10(t *testing.T) {
	// Ten faces 1,2,3,1,2,3,1,2,3,1 reused cyclically through 100D = 190 (Book 1 p.260).
	if got := scripted(1, 2, 3, 1, 2, 3, 1, 2, 3, 1).ManyDice10(100); got != 190 {
		t.Errorf("ManyDice10(100) = %d, want 190", got)
	}

	if got := scripted(1, 2, 3).ManyDice10(0); got != 0 {
		t.Errorf("ManyDice10(0) = %d, want 0", got)
	}
}

func TestManyDice2D(t *testing.T) {
	// 2D=5 (2+3); subsample 1,2,4,5,5 reused through 100D = 340 (Book 1 p.260).
	if got := scripted(2, 3, 1, 2, 4, 5, 5).ManyDice2D(100); got != 340 {
		t.Errorf("ManyDice2D(100) = %d, want 340", got)
	}
	// 2D=2 (1+1); subsample 1,3 over 75D = 38·1 + 37·3 = 149.
	if got := scripted(1, 1, 1, 3).ManyDice2D(75); got != 149 {
		t.Errorf("ManyDice2D(75) = %d, want 149", got)
	}
}

func TestAverage35(t *testing.T) {
	if got := Average35(50); got != 175 {
		t.Errorf("Average35(50) = %v, want 175", got)
	}

	if got := Average35(51); got != 178.5 {
		t.Errorf("Average35(51) = %v, want 178.5", got)
	}

	if got := Average35(0); got != 0 {
		t.Errorf("Average35(0) = %v, want 0", got)
	}
}

func TestManyDice35Flux(t *testing.T) {
	// value = (7 + Flux) / 2, times n. Flux = d6 - d6.
	cases := []struct {
		name string
		hi   int // first die
		lo   int // second die (Flux = hi - lo)
		want float64
	}{
		{"flux +5", 6, 1, 600},
		{"flux -5", 1, 6, 100},
		{"flux 0", 4, 4, 350},
		{"flux +2", 5, 3, 450},
	}
	for _, c := range cases {
		if got := scripted(c.hi, c.lo).ManyDice35Flux(100); got != c.want {
			t.Errorf("%s: ManyDice35Flux(100) = %v, want %v", c.name, got, c.want)
		}
	}

	if got := scripted(4, 4).ManyDice35Flux(0); got != 0 {
		t.Errorf("ManyDice35Flux(0) = %v, want 0", got)
	}
}
