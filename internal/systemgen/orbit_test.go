package systemgen

import "testing"

func TestHZOrbit(t *testing.T) {
	cases := []struct {
		star Star
		want int
		ok   bool
	}{
		{Star{Type: "F", Decimal: 7, Size: "V"}, 4, true}, // Regina primary (golden)
		{Star{Type: "O", Size: "Ia"}, 15, true},
		{Star{Type: "M", Size: "V"}, 0, true},
		{Star{Type: "K", Size: "III"}, 8, true},
		{Star{Type: "G", Size: "D"}, 0, true},   // white dwarf
		{Star{Type: "O", Size: "VI"}, 0, false}, // "-" cell
		{Star{Type: "M", Size: "IV"}, 0, false}, // "-" cell
		{Star{Type: "BD", Decimal: -1}, 0, false},
		{Star{Type: "Z", Size: "V"}, 0, false}, // unknown type
	}
	for _, c := range cases {
		got, ok := HZOrbit(c.star)
		if got != c.want || ok != c.ok {
			t.Errorf("HZOrbit(%v) = (%d, %v), want (%d, %v)", c.star, got, ok, c.want, c.ok)
		}
	}
}

func TestOrbitAU(t *testing.T) {
	cases := map[int]float64{0: 0.2, 3: 1.0, 5: 2.8, 6: 5.2, 20: 78700, -1: 0, 21: 0}
	for orbit, want := range cases {
		if got := OrbitAU(orbit); got != want {
			t.Errorf("OrbitAU(%d) = %v, want %v", orbit, got, want)
		}
	}
}
