package tradecode

import "testing"

// TestOrderIsComplete pins the canonical set at 46 codes with no duplicates, so a
// dropped or doubled code fails loudly rather than silently changing the registry.
func TestOrderIsComplete(t *testing.T) {
	if len(Order) != 46 {
		t.Errorf("Order has %d codes, want 46 (Book 3 Chart D)", len(Order))
	}

	seen := make(map[Code]bool, len(Order))
	for _, c := range Order {
		if len(c) != 2 {
			t.Errorf("code %q is not two letters", c)
		}

		if seen[c] {
			t.Errorf("code %q appears twice in Order", c)
		}

		seen[c] = true
	}
}

// TestRankAndValid: every Order member is Valid and ranks at its index; an unknown
// code is not Valid and ranks last (total, never panicking).
func TestRankAndValid(t *testing.T) {
	for i, c := range Order {
		if !Valid(c) {
			t.Errorf("%q is in Order but not Valid", c)
		}

		if Rank(c) != i {
			t.Errorf("Rank(%q) = %d, want %d", c, Rank(c), i)
		}
	}

	unknown := Code("Zz")
	if Valid(unknown) {
		t.Error("Valid(Zz) = true, want false")
	}

	if Rank(unknown) != len(Order) {
		t.Errorf("Rank(unknown) = %d, want %d (sorts last)", Rank(unknown), len(Order))
	}
}
