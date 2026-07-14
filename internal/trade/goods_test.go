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
		tcs      []string
		sourceTC string
		want     string
	}{
		{[]string{"Fl", "He", "Hi"}, "Fl", "Strange"}, // He is first labeled non-source
		{[]string{"Ri"}, "Ri", ""},                    // the only class chose the column
		{[]string{"Hi", "In"}, "Xx", ""},              // Hi's Processed omitted on Industrial
		{[]string{"Hi"}, "Xx", "Processed"},           // non-Industrial keeps Processed
		{[]string{"Va", "As"}, "Xx", "Strange"},       // Va's Exotic omitted on Asteroid, As -> Strange
	}
	for _, c := range cases {
		if got := tradeGoodsDetail(c.tcs, c.sourceTC); got != c.want {
			t.Errorf("tradeGoodsDetail(%v, %q) = %q, want %q", c.tcs, c.sourceTC, got, c.want)
		}
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
	cases := map[string]string{"Ga": "Ag-1", "Fa": "Ag-2", "Cp": "CpCsCx", "Cx": "CpCsCx", "Fl": "Fl", "Ri": "Ri"}
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
