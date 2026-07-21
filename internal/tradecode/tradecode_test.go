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

// TestRank: every Order member ranks at its index; an unknown code ranks last
// (total, never panicking) so OrderTradeCodes's sort stays defined.
func TestRank(t *testing.T) {
	for i, c := range Order {
		if Rank(c) != i {
			t.Errorf("Rank(%q) = %d, want %d", c, Rank(c), i)
		}

		if !Valid(c) {
			t.Errorf("%q is in Order but not Valid", c)
		}
	}

	if unknown := Code("Zz"); Rank(unknown) != len(Order) {
		t.Errorf("Rank(unknown) = %d, want %d (sorts last)", Rank(unknown), len(Order))
	}

	if Valid("Zz") {
		t.Error("Valid(Zz) = true, want false for a non-Chart-D code")
	}
}
