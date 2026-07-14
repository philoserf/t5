package shipcombat

import "testing"

// TestHullLocations golden-locks the Book 2 p.86 Table H compartment structure.
func TestHullLocations(t *testing.T) {
	cases := []struct {
		ord  int
		want HullLocation
	}{
		{1, HullLocation{5, 4, 20, 3}},      // A 100t: 5 compartments, span ±4
		{8, HullLocation{11, 5, 70, 12}},    // H 800t: span ±5
		{22, HullLocation{23, 11, 100, 17}}, // X 2200t: 23 compartments, span ±11
	}
	for _, c := range cases {
		got, ok := HullLocations(c.ord)
		if !ok || got != c.want {
			t.Errorf("HullLocations(%d) = %+v,%v, want %+v", c.ord, got, ok, c.want)
		}
	}
	// Hull Z (the largest, 2400t) is tabulated with 25 compartments and span ±12;
	// ordinals beyond it are not.
	if got, ok := HullLocations(24); !ok || got != (HullLocation{25, 12, 100, 17}) {
		t.Errorf("HullLocations(24) = %+v,%v, want Z {25,12,100,17}", got, ok)
	}
	if _, ok := HullLocations(25); ok {
		t.Errorf("ordinal 25 (beyond Z) should not be tabulated")
	}
	// A hit on a Hull-A ship centers on Compartment 0 and clamps to its ±4 span.
	a, _ := HullLocations(1)
	if got := HitCompartment(6, 2, a.Span); got != 4 {
		t.Errorf("Hull-A hit clamp = %d, want 4 (span ±%d)", got, a.Span)
	}
}

func TestSubCompartmentsKnockedOut(t *testing.T) {
	cases := map[int]int{0: 0, 4: 1, 10: 1, 15: 2, 60: 6, 61: 7}
	for points, want := range cases {
		if got := SubCompartmentsKnockedOut(points); got != want {
			t.Errorf("SubCompartmentsKnockedOut(%d) = %d, want %d", points, got, want)
		}
	}
	// Sixty points fill a full compartment's six SubCompartments.
	if SubCompartmentsKnockedOut(CompartmentCapacity) != SubCompartmentsPerCompartment {
		t.Errorf("60 points should knock out all %d SubCompartments", SubCompartmentsPerCompartment)
	}
}
