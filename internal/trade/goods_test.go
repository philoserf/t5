package trade

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

// TestRandomTradeGoodsZivije golden-locks the Book 2 p.210 Zivije example: an Fl
// world, first 1D=3 (Pharma), second 1D=5 (Antibiotics).
func TestRandomTradeGoodsZivije(t *testing.T) {
	g := RandomTradeGoods(dice.NewScripted(3, 5), []string{"Fl"})
	if g.Name != "Antibiotics" || g.Type != "Pharma" || g.Imbalance != "" {
		t.Errorf("goods = %+v, want Antibiotics/Pharma", g)
	}
}

// TestRandomTradeGoodsImbalance golden-locks the Book 2 p.210 Knorbes example:
// an Ag world (via Ga -> Ag-1), first 1D=6 (Imbalances) second 1D=5 (Na), then
// recursing on Na: first 1D=4 (Rares) second 1D=5 (Pelts).
func TestRandomTradeGoodsImbalance(t *testing.T) {
	g := RandomTradeGoods(dice.NewScripted(6, 5, 4, 5), []string{"Ga"})
	if g.Name != "Pelts" || g.Type != "Rares" || g.Imbalance != "Na" {
		t.Errorf("goods = %+v, want Pelts/Rares via Imbalance Na", g)
	}
}

func TestTradeGoodString(t *testing.T) {
	if got := (TradeGood{Name: "Antibiotics", Detail: "Processed"}).String(); got != "Processed Antibiotics" {
		t.Errorf("String() = %q, want %q", got, "Processed Antibiotics")
	}
	if got := (TradeGood{Name: "Pelts"}).String(); got != "Pelts" {
		t.Errorf("String() = %q, want %q", got, "Pelts")
	}
}

func TestTradeGoodsDetail(t *testing.T) {
	cases := []struct {
		tcs           []string
		sourceTC, col string
		want          string
	}{
		{[]string{"Fl", "He", "Hi"}, "Fl", "Fl", "Strange"}, // He precedes Hi in chart order
		{[]string{"Ri"}, "Ri", "Ri", ""},                    // the only class chose the column
		{[]string{"Hi"}, "Xx", "Fl", "Processed"},           // goods not from the In column
		{
			[]string{"Hi"},
			"Xx",
			"In",
			"",
		}, // Hi's Processed omitted for In-column goods
		{[]string{"Va"}, "Xx", "Ic", "Exotic"}, // goods not from the As column
		{
			[]string{"Va"},
			"Xx",
			"As",
			"",
		}, // Va's Exotic omitted for As-column goods
	}
	for _, c := range cases {
		if got := tradeGoodsDetail(c.tcs, c.sourceTC, c.col); got != c.want {
			t.Errorf(
				"tradeGoodsDetail(%v, %q, col %q) = %q, want %q",
				c.tcs,
				c.sourceTC,
				c.col,
				got,
				c.want,
			)
		}
	}
}

// TestTradeGoodsDetailOrderIndependent locks the fix for a prefix that used to
// depend on how the caller happened to order a world's trade codes.
func TestTradeGoodsDetailOrderIndependent(t *testing.T) {
	a := tradeGoodsDetail([]string{"Ri", "Ic", "As"}, "In", "In")
	b := tradeGoodsDetail([]string{"As", "Ic", "Ri"}, "In", "In")
	if a != b {
		t.Errorf("same world, different code order gave %q vs %q", a, b)
	}
	if a != "Strange" { // As is first in the p.219 chart order
		t.Errorf("detail = %q, want Strange (first in chart order)", a)
	}
}

// TestImbalanceNeverLeaksTradeCode guards the redirect chain: however many
// Imbalance hops it takes, the good's Name must be a real good, never a trade
// class. Rolling block 6 (Imbalances) forever would previously leak "Ic" as a
// cargo name.
func TestImbalanceNeverLeaksTradeCode(t *testing.T) {
	// A scripted roller that always rolls 6: block 6 is Imbalances in every column
	// that has one, so this is the pathological all-redirects case.
	sixes := make([]int, 200)
	for i := range sixes {
		sixes[i] = 6
	}
	g := RandomTradeGoods(dice.NewScripted(sixes...), []string{"Ga"})
	if g.Type == imbalancesBlock {
		t.Fatalf("good escaped as an Imbalances entry: %+v", g)
	}
	if goodsColumnEligible[g.Name] {
		t.Errorf("good's Name %q is a trade class, not a good", g.Name)
	}
}

func TestSelectGoodsColumnDefault(t *testing.T) {
	// A world with no column-eligible trade class falls back to Non-Agricultural.
	col, src := selectGoodsColumn(dice.NewScripted(1), []string{"Hi", "He"})
	if col != "Na" || src != "Na" {
		t.Errorf("selectGoodsColumn(no eligible) = %q/%q, want Na/Na", col, src)
	}
}

func TestColumnFor(t *testing.T) {
	cases := map[string]string{
		"Ga": "Ag-1",
		"Fa": "Ag-2",
		"Cp": "CpCsCx",
		"Cx": "CpCsCx",
		"Fl": "Fl",
		"Ri": "Ri",
	}
	for tc, want := range cases {
		if got := columnFor(tc); got != want {
			t.Errorf("columnFor(%q) = %q, want %q", tc, got, want)
		}
	}
}

// TestGoodsDataWellFormed guards the transcription: every column has six blocks,
// each with a non-empty type and six non-empty entries.
func TestGoodsDataWellFormed(t *testing.T) {
	for col, blocks := range tradeGoodsColumns {
		for i, b := range blocks {
			if b.Type == "" {
				t.Errorf("%s block %d has no type", col, i+1)
			}
			for j, good := range b.Goods {
				if good == "" {
					t.Errorf("%s block %d good %d is empty", col, i+1, j+1)
				}
			}
		}
	}
}
