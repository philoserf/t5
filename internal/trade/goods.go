package trade

import (
	"slices"

	"github.com/philoserf/t5/internal/dice"
)

// A Good is a specific cargo produced by the Random Trade Goods chart (Book
// 2 pp.218-219): its name, its type category, an optional Trade Good Detail
// prefix, and — when it came from an Imbalance redirect — the trade class whose
// oversupply produced it.
type Good struct {
	Name      string
	Type      string
	Detail    string
	Imbalance string
}

// String renders the good with its detail prefix, e.g. "Processed Antibiotics".
func (g Good) String() string {
	if g.Detail != "" {
		return g.Detail + " " + g.Name
	}

	return g.Name
}

// RandomTradeGoods identifies a specific cargo available at a world (Book 2
// pp.218-219): it picks a trade-classification column, rolls the type block and
// the specific good, follows Imbalance redirects to another column, and applies
// a Trade Good Detail prefix from the world's other trade classes.
func RandomTradeGoods(r *dice.Roller, worldTCs []string) Good {
	col, sourceTC := selectGoodsColumn(r, worldTCs)
	// The detail's footnotes key on where the goods came from, which after an
	// Imbalance redirect is no longer the column the world started on.
	g, origin := rollGoodsColumn(r, col)
	g.Detail = tradeGoodsDetail(worldTCs, sourceTC, origin)

	return g
}

// selectGoodsColumn chooses the column for a world (Book 2 p.218): from the
// world's column-eligible trade classes (defaulting to Non-Agricultural when
// none qualify, and picking randomly among several), mapping Ag to Ag-1 or Ag-2
// at random. It returns the column key and the trade class that chose it.
func selectGoodsColumn(r *dice.Roller, worldTCs []string) (string, string) {
	eligible := make([]string, 0, len(worldTCs))
	for _, tc := range worldTCs {
		if goodsColumnEligible[tc] {
			eligible = append(eligible, tc)
		}
	}

	if len(eligible) == 0 {
		return "Na", "Na"
	}
	// Index(1) consumes no dice, so a lone eligible class needs no special case.
	tc := eligible[r.Index(len(eligible))]

	return resolveColumn(r, tc), tc
}

// imbalancesBlock is the type name of the block whose entries are trade classes
// to redirect to rather than goods (Book 2 p.218).
const imbalancesBlock = "Imbalances"

// maxImbalanceHops bounds an Imbalance redirect chain. The chart's columns
// redirect to one another (As -> Ag -> Fl -> Na -> ...), so a chain can cycle;
// the book relies on it terminating naturally, and this is only a safety net.
const maxImbalanceHops = 4

// columnFor maps a trade class to its Random Trade Goods column key. Ga and Fa
// share the Ag columns and the capital codes share one column. A bare Ag needs a
// die to choose between Ag-1 and Ag-2 — use resolveColumn for that.
func columnFor(tc string) string {
	switch tc {
	case "Ag", "Ga":
		return "Ag-1"
	case "Fa":
		return "Ag-2"
	case "Cp", "Cs", "Cx":
		return "CpCsCx"
	default:
		return tc
	}
}

// resolveColumn maps a trade class to its column, rolling between Ag-1 and Ag-2
// for a bare Ag (Book 2 p.218: "If World TC=Ag is selected, pick randomly between
// Ag-1 and Ag-2"). This applies wherever an Ag column is chosen — the world's own
// class and an Imbalance redirect alike.
func resolveColumn(r *dice.Roller, tc string) string {
	if tc == "Ag" && r.Die() > 3 {
		return "Ag-2"
	}

	return columnFor(tc)
}

// rollGoodsColumn rolls a column's type block and specific good (Book 2 p.218).
// An Imbalances block names another trade class whose column the roll continues
// on, and the good carries that class as its Imbalance — the oversupply it came
// from, which is what earns the +Cr1,000 selling bonus (see ImbalanceBonus). A
// redirect chain can cycle, so past maxImbalanceHops a redirect is replaced by
// the column's first block of real goods: a trade class must never escape as a
// cargo name. It returns the good together with the column it finally came from,
// which is the origin the Trade Good Detail footnotes key on (see tradeGoodsDetail)
// and is not the starting column once a redirect has been followed.
func rollGoodsColumn(r *dice.Roller, column string) (Good, string) {
	imbalance := ""

	for hop := 0; ; hop++ {
		blocks, ok := tradeGoodsColumns[column]
		if !ok {
			// The key is always program data — a chart column or an Imbalances
			// redirect target — so an unknown one is a transcription bug, not
			// runtime input. Fail loudly rather than serve an empty cargo.
			panic("trade: unknown Random Trade Goods column " + column)
		}

		block := blocks[r.Die()-1]
		// Re-rolling here would be no safer than the chain it guards: a degenerate
		// die source that keeps naming the Imbalances block would spin forever. Take
		// the first real block instead, so termination does not depend on the dice.
		if hop >= maxImbalanceHops && block.Type == imbalancesBlock {
			block = firstGoodsBlock(blocks, column)
		}

		entry := block.Goods[r.Die()-1]
		if block.Type != imbalancesBlock {
			return Good{Name: entry, Type: block.Type, Imbalance: imbalance}, column
		}
		// The chain continues on the named class's column; the last redirect is
		// the oversupply that actually produced the goods.
		imbalance = entry
		column = resolveColumn(r, entry)
	}
}

// firstGoodsBlock returns a column's first block that holds real goods rather
// than Imbalance redirects. Every column in the chart has one (TestGoodsDataWellFormed
// asserts it), so a column without one is a transcription bug.
func firstGoodsBlock(blocks [6]goodsBlock, column string) goodsBlock {
	for _, b := range blocks {
		if b.Type != imbalancesBlock {
			return b
		}
	}

	panic("trade: Random Trade Goods column " + column + " has no block of real goods")
}

// tradeGoodsDetail picks a Trade Good Detail prefix (Book 2 p.219) from the
// world's trade classes other than the one that chose the column. The book says
// only "select one", so we take the first in the detail chart's own order — a
// fixed choice, so the same world always describes its cargo the same way
// regardless of how the caller ordered its codes. The chart's two footnotes omit
// a label when it would be redundant with the goods' own origin: Hi's "Processed"
// for goods out of the Industrial column, and Va's "Exotic" for goods out of the
// Asteroid column. Returns "" when no class carries a label.
func tradeGoodsDetail(worldTCs []string, sourceTC, column string) string {
	for _, tc := range tradeGoodsDetailOrder {
		if tc == sourceTC || !slices.Contains(worldTCs, tc) {
			continue
		}

		if tc == "Hi" && column == "In" {
			continue
		}

		if tc == "Va" && column == "As" {
			continue
		}

		if label, ok := tradeGoodsDetailLabel[tc]; ok {
			return label
		}
	}

	return ""
}
