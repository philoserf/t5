package shipcombat

import "testing"

// TestHitCompartment golden-locks the Book 2 p.197 Antares-vs-Joshua example:
// targeting Compartment -6, Flux +1/+5/0 across three rounds -> -5/-1/-6.
func TestHitCompartment(t *testing.T) {
	span := 10
	if got := HitCompartment(1, -6, span); got != -5 {
		t.Errorf("round 1 = %d, want -5", got)
	}
	if got := HitCompartment(5, -6, span); got != -1 {
		t.Errorf("round 2 = %d, want -1", got)
	}
	if got := HitCompartment(0, -6, span); got != -6 {
		t.Errorf("round 3 = %d, want -6", got)
	}
	// Center-of-Mass with no targeting is Compartment 0 (Book 2 p.199 Murphy).
	if got := HitCompartment(0, 0, span); got != 0 {
		t.Errorf("center of mass = %d, want 0", got)
	}
	// A Flux beyond the span clamps to the outermost compartment.
	if got := HitCompartment(6, 6, span); got != 10 {
		t.Errorf("clamp = %d, want 10", got)
	}
}

func TestPenetrate(t *testing.T) {
	// Book 2 p.199: 6 damage against a weakened AV=5 layer penetrates with 1 point
	// remaining; against a full AV=10 layer it does not.
	if pen, interior := Penetrate(6, []int{5}); !pen || interior != 1 {
		t.Errorf("Penetrate(6,[5]) = %v,%d, want true,1", pen, interior)
	}
	if pen, _ := Penetrate(6, []int{10}); pen {
		t.Errorf("Penetrate(6,[10]) should not penetrate")
	}
	// Two AV=6 layers: 15 damage clears both with 3 to the interior.
	if pen, interior := Penetrate(15, []int{6, 6}); !pen || interior != 3 {
		t.Errorf("Penetrate(15,[6,6]) = %v,%d, want true,3", pen, interior)
	}
	// Damage exactly equal to a single layer's AV is absorbed (nothing remains).
	if pen, _ := Penetrate(6, []int{6}); pen {
		t.Errorf("Penetrate(6,[6]) should be absorbed with no interior damage")
	}
}

func TestDamageLocation(t *testing.T) {
	cases := map[int]string{2: "Bridge", 7: "Drives", 8: "Power Plant", 12: "Computer"}
	for twoD, want := range cases {
		if got := DamageLocation(twoD); got != want {
			t.Errorf("DamageLocation(%d) = %q, want %q", twoD, got, want)
		}
	}
}

// TestSeverity golden-locks the Book 2 p.198 Vigilant recovery example: with 3
// SubCompartments InOp (Mod +3), 1D of 1/6/4 -> Formidable 4D / Destroyed / 7D.
func TestSeverity(t *testing.T) {
	if s := Severity(1, 3); s != 4 || SeverityLabel(s) != "Formidable" || Destroyed(s) {
		t.Errorf("Severity(1,+3) = %d (%s), want 4 Formidable", s, SeverityLabel(s))
	}
	if s := Severity(6, 3); s != 9 || !Destroyed(s) {
		t.Errorf("Severity(6,+3) = %d, want 9 Destroyed", s)
	}
	if s := Severity(4, 3); s != 7 || SeverityLabel(s) != "Impossible" {
		t.Errorf("Severity(4,+3) = %d (%s), want 7 Impossible", s, SeverityLabel(s))
	}
	if s := Severity(1, 0); s != 1 || SeverityLabel(s) != "Easy" {
		t.Errorf("Severity(1,0) = %d (%s), want 1 Easy", s, SeverityLabel(s))
	}
}
