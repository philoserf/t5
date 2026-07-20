package trade

import (
	"slices"
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/tradecode"
)

// TestRandomTradeGoodsZivije golden-locks the Book 2 p.210 Zivije example: an Fl
// world, first 1D=3 (Pharma), second 1D=5 (Antibiotics).
func TestRandomTradeGoodsZivije(t *testing.T) {
	g := RandomTradeGoods(dice.NewScripted(3, 5), []tradecode.Code{"Fl"})
	if g.Name != "Antibiotics" || g.Type != "Pharma" || g.Imbalance != "" {
		t.Errorf("goods = %+v, want Antibiotics/Pharma", g)
	}
}

// TestRandomTradeGoodsImbalance golden-locks the Book 2 p.210 Knorbes example:
// an Ag world (via Ga -> Ag-1), first 1D=6 (Imbalances) second 1D=5 (Na), then
// recursing on Na: first 1D=4 (Rares) second 1D=5 (Pelts).
func TestRandomTradeGoodsImbalance(t *testing.T) {
	g := RandomTradeGoods(dice.NewScripted(6, 5, 4, 5), []tradecode.Code{"Ga"})
	if g.Name != "Pelts" || g.Type != "Rares" || g.Imbalance != "Na" {
		t.Errorf("goods = %+v, want Pelts/Rares via Imbalance Na", g)
	}
}

func TestTradeGoodString(t *testing.T) {
	if got := (Good{Name: "Antibiotics", Detail: "Processed"}).String(); got != "Processed Antibiotics" {
		t.Errorf("String() = %q, want %q", got, "Processed Antibiotics")
	}

	if got := (Good{Name: "Pelts"}).String(); got != "Pelts" {
		t.Errorf("String() = %q, want %q", got, "Pelts")
	}
}

func TestTradeGoodsDetail(t *testing.T) {
	cases := []struct {
		tcs      []tradecode.Code
		sourceTC tradecode.Code
		col      string
		want     string
	}{
		{[]tradecode.Code{"Fl", "He", "Hi"}, "Fl", "Fl", "Strange"}, // He precedes Hi in chart order
		{[]tradecode.Code{"Ri"}, "Ri", "Ri", ""},                    // the only class chose the column
		{[]tradecode.Code{"Hi"}, "Xx", "Fl", "Processed"},           // goods not from the In column
		{
			[]tradecode.Code{"Hi"},
			"Xx",
			"In",
			"",
		}, // Hi's Processed omitted for In-column goods
		{[]tradecode.Code{"Va"}, "Xx", "Ic", "Exotic"}, // goods not from the As column
		{
			[]tradecode.Code{"Va"},
			"Xx",
			"As",
			"",
		}, // Va's Exotic omitted for As-column goods
		// After an Imbalance redirect the starting class and the origin column
		// differ, and it is the origin that must not label its own goods: an
		// {As, Ri} world whose As column redirects to Ri yielded "Quality Ri-good".
		// Ba is the next class in chart order, so a redirect suppresses one label
		// without suppressing the Detail entirely.
		{[]tradecode.Code{"As", "Ri"}, "As", "Ri", ""},
		{[]tradecode.Code{"As", "Ba", "Ri"}, "As", "Ri", "Gathered"},
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
	a := tradeGoodsDetail([]tradecode.Code{"Ri", "Ic", "As"}, "In", "In")

	b := tradeGoodsDetail([]tradecode.Code{"As", "Ic", "Ri"}, "In", "In")
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

	g := RandomTradeGoods(dice.NewScripted(sixes...), []tradecode.Code{"Ga"})
	if g.Type == imbalancesBlock {
		t.Fatalf("good escaped as an Imbalances entry: %+v", g)
	}

	if goodsColumnEligible[g.Name] {
		t.Errorf("good's Name %q is a trade class, not a good", g.Name)
	}
}

// TestRollGoodsColumnUnknownColumnPanics guards the column lookup: the key is
// always program data (a chart column or an Imbalances redirect target), so an
// unknown one is a transcription bug and must fail loudly rather than yield an
// empty cargo.
func TestRollGoodsColumnUnknownColumnPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("rollGoodsColumn on an unknown column did not panic")
		}
	}()

	g, col := rollGoodsColumn(dice.NewScripted(1, 1), "Zz")
	t.Errorf("rollGoodsColumn returned %+v (column %q) for an unknown column", g, col)
}

// TestImbalanceHopCapTerminates guards the hop cap itself. The Na and Fl columns
// redirect to each other (Na block 6 entry 6 -> Fl; Fl block 6 entry 4 -> Na), so
// this script ping-pongs to the cap and then offers nothing but 6s — the block
// that is Imbalances in both columns. A re-rolling cap never escapes that.
func TestImbalanceHopCapTerminates(t *testing.T) {
	// Na -> Fl -> Na -> Fl -> (cap), then only sixes.
	prefix := []int{6, 6, 6, 4, 6, 6, 6, 4}
	// budget is generous room past the prefix for the capped hop's own rolls; a
	// draw beyond it means the cap is looping. The source then yields 1s so the
	// test reports a count rather than hanging.
	const budget = 4

	drawn := 0
	r := dice.NewSource(func() int {
		drawn++

		switch {
		case drawn <= len(prefix):
			return prefix[drawn-1]
		case drawn <= len(prefix)+budget:
			return 6
		default:
			return 1
		}
	})

	g, _ := rollGoodsColumn(r, "Na")

	if drawn > len(prefix)+budget {
		t.Fatalf("rollGoodsColumn drew %d dice (budget %d): the hop cap is re-rolling without bound",
			drawn, len(prefix)+budget)
	}

	if g.Type == imbalancesBlock || goodsColumnEligible[g.Name] {
		t.Errorf("good escaped as an Imbalances entry: %+v", g)
	}
}

// TestRandomTradeGoodsDetailAfterRedirect locks the Trade Good Detail prefix to
// the column the goods actually came from after an Imbalance redirect, not the
// column the world started on. The book prints no worked example of the two
// together, so these are hand-traced from the pp.218-219 charts.
func TestRandomTradeGoodsDetailAfterRedirect(t *testing.T) {
	// Redirected INTO As: a Ga/Va world starts on Ag-1, block 6 (Imbalances)
	// entry 1 (As), then rolls Valuta/Platinum on the As column. Va's "Exotic"
	// is redundant with an asteroid origin (p.219 footnote), so it is omitted —
	// keying on the starting Ag-1 column would wrongly keep it.
	into := RandomTradeGoods(dice.NewScripted(1, 6, 1, 3, 1), []tradecode.Code{"Ga", "Va"})
	if into.Name != "Platinum" || into.Imbalance != "As" || into.Detail != "" {
		t.Errorf("redirected into As: %+v, want Platinum via As with no detail", into)
	}

	// Redirected AWAY from As: an As/Va world starts on As, block 6 (Imbalances)
	// entry 2 (De), then rolls Rares/Nectars on the De column. The goods are no
	// longer of asteroid origin, so Va's "Exotic" applies — keying on the
	// starting As column would wrongly drop it.
	away := RandomTradeGoods(dice.NewScripted(1, 6, 2, 5, 3), []tradecode.Code{"As", "Va"})
	if away.Name != "Nectars" || away.Imbalance != "De" || away.Detail != "Exotic" {
		t.Errorf("redirected away from As: %+v, want Exotic Nectars via De", away)
	}
}

func TestSelectGoodsColumnDefault(t *testing.T) {
	// A world with no column-eligible trade class falls back to Non-Agricultural.
	col, src := selectGoodsColumn(dice.NewScripted(1), []tradecode.Code{"Hi", "He"})
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

			if b.Type != imbalancesBlock {
				continue
			}
			// An Imbalances entry becomes the next column, so a typo here — a real
			// trade class with no column of its own, say "Cs" — would send the roll
			// to a key that does not exist. Cross-check every redirect target,
			// including both halves of a bare Ag.
			for j, tc := range b.Goods {
				targets := []string{columnFor(tc)}
				if tc == "Ag" {
					targets = append(targets, "Ag-2")
				}

				for _, target := range targets {
					if _, ok := tradeGoodsColumns[target]; !ok {
						t.Errorf("%s Imbalances entry %d (%q) redirects to unknown column %q",
							col, j+1, tc, target)
					}
				}
			}
		}
	}

	// Every column must offer real goods somewhere, or the hop cap has nothing to
	// fall back on (see rollGoodsColumn).
	for col, blocks := range tradeGoodsColumns {
		if !slices.ContainsFunc(blocks[:], func(b goodsBlock) bool { return b.Type != imbalancesBlock }) {
			t.Errorf("column %s is entirely Imbalances blocks", col)
		}
	}
}
