package systemgen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func scriptedRoller(vals ...int) *dice.Roller {
	i := 0
	return dice.NewSource(func() int {
		v := vals[i%len(vals)]
		i++
		return v
	})
}

func TestStarString(t *testing.T) {
	cases := []struct {
		star Star
		want string
	}{
		{Star{Type: "F", Decimal: 7, Size: "V"}, "F7 V"},
		{Star{Type: "M", Decimal: 0, Size: "III"}, "M0 III"},
		{Star{Type: "G", Decimal: -1, Size: "D"}, "G D"},
		{Star{Type: "BD", Decimal: -1}, "BD"},
	}
	for _, c := range cases {
		if got := c.star.String(); got != c.want {
			t.Errorf("Star%+v.String() = %q, want %q", c.star, got, c.want)
		}
	}
}

func TestClassifyTable(t *testing.T) {
	// typeFlux picks the spectral type; sizeFlux picks the row, the type picks
	// the column. Values read from the Book 3 p. 28 table.
	cases := []struct {
		name                     string
		typeFlux, sizeFlux, deci int
		pickB                    bool
		wantType, wantSize       string
		wantDecimal              int
	}{
		{"F dwarf", -2, 0, 7, false, "F", "V", 7},      // type row -2 = F; size row 0, F col = V
		{"G subgiant", -1, -3, 2, false, "G", "IV", 2}, // type row -1 = G; size row -3, G col = IV
		{"K giant", 1, -4, 0, false, "K", "III", 0},    // type row +1 = K; size row -4, K col = III
		{"M main", 3, 2, 5, false, "M", "V", 5},        // type row +3 = M; size row +2, M col = V
	}
	for _, c := range cases {
		got := classify(c.typeFlux, c.sizeFlux, c.deci, c.pickB)
		if got.Type != c.wantType || got.Size != c.wantSize || got.Decimal != c.wantDecimal {
			t.Errorf("%s: classify = %+v, want type %s size %s decimal %d",
				c.name, got, c.wantType, c.wantSize, c.wantDecimal)
		}
	}
}

func TestClassifyBD(t *testing.T) {
	// typeFlux +6 -> BD; decimal and size are ignored.
	got := classify(6, 0, 5, false)
	if got.Type != "BD" || got.Size != "" || got.Decimal != -1 {
		t.Fatalf("classify BD = %+v, want {BD -1 }", got)
	}
}

func TestClassifySizeDIgnoresDecimal(t *testing.T) {
	// typeFlux +2 -> K; sizeFlux +5 -> row of all D.
	got := classify(2, 5, 8, false)
	if got.Size != "D" || got.Decimal != -1 {
		t.Fatalf("classify size D = %+v, want size D decimal -1", got)
	}
}

func TestClassifyOBSelection(t *testing.T) {
	// typeFlux -6 -> OB. pickB chooses between O and B.
	o := classify(-6, -6, 0, false)
	b := classify(-6, -6, 0, true)
	if o.Type != "O" || b.Type != "B" {
		t.Fatalf("OB selection: O=%s B=%s, want O and B", o.Type, b.Type)
	}
}

func TestClassifyForbiddenSizeFallsBackToV(t *testing.T) {
	// Size IV is forbidden for K5-K9. type row +1 = K; size row -3 gives IV in
	// the K column; decimal 7 makes it K7, so it must fall back to V.
	if got := classify(1, -3, 7, false); got.Type != "K" || got.Size != "V" {
		t.Fatalf("K7 size-IV fallback = %+v, want K V", got)
	}
	// A K4 on the same cell keeps IV (only K5-K9 are forbidden).
	if got := classify(1, -3, 4, false); got.Size != "IV" {
		t.Fatalf("K4 size = %+v, want IV", got)
	}
	// Size VI is forbidden for F0-F4. type row -2 = F; size row +4 gives VI in
	// the F column; decimal 2 makes it F2, so it must fall back to V.
	if got := classify(-2, 4, 2, false); got.Type != "F" || got.Size != "V" {
		t.Fatalf("F2 size-VI fallback = %+v, want F V", got)
	}
}
