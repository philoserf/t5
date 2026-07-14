package trade

import (
	"slices"

	"github.com/philoserf/t5/internal/dice"
)

// A TradeGood is a specific cargo produced by the Random Trade Goods chart (Book
// 2 pp.218-219): its name, its type category, an optional Trade Good Detail
// prefix, and — when it came from an Imbalance redirect — the trade class whose
// oversupply produced it.
type TradeGood struct {
	Name      string
	Type      string
	Detail    string
	Imbalance string
}

// String renders the good with its detail prefix, e.g. "Processed Antibiotics".
func (g TradeGood) String() string {
	if g.Detail != "" {
		return g.Detail + " " + g.Name
	}
	return g.Name
}

// RandomTradeGoods identifies a specific cargo available at a world (Book 2
// pp.218-219): it picks a trade-classification column, rolls the type block and
// the specific good, follows Imbalance redirects to another column, and applies
// a Trade Good Detail prefix from the world's other trade classes.
func RandomTradeGoods(r *dice.Roller, worldTCs []string) TradeGood {
	col, sourceTC := selectGoodsColumn(r, worldTCs)
	g := rollGoodsColumn(r, col, 0)
	g.Detail = tradeGoodsDetail(worldTCs, sourceTC)
	return g
}

// selectGoodsColumn chooses the column for a world (Book 2 p.218): from the
// world's column-eligible trade classes (defaulting to Non-Agricultural when
// none qualify, and picking randomly among several), mapping Ag to Ag-1 or Ag-2
// at random. It returns the column key and the trade class that chose it.
func selectGoodsColumn(r *dice.Roller, worldTCs []string) (column, sourceTC string) {
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
	// Ag alone splits across two columns on a die; every other class maps directly
	// (columnFor sends a bare Ag to Ag-1, which is also this roll's low half).
	if tc == "Ag" && r.Die() > 3 {
		return "Ag-2", tc
	}
	return columnFor(tc), tc
}

// columnFor maps a trade class to its Random Trade Goods column key. Ga and Fa
// share the Ag columns; the capital codes share one column. An Ag redirect
// (with no random sub-roll available) defaults to Ag-1.
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

// rollGoodsColumn rolls a column's type block and specific good (Book 2 p.218):
// an Imbalances block redirects to another trade class's column (bounded to
// avoid a redirect loop), tagging the result with the imbalance source.
func rollGoodsColumn(r *dice.Roller, column string, depth int) TradeGood {
	blocks := tradeGoodsColumns[column]
	block := blocks[r.Die()-1]
	entry := block.Goods[r.Die()-1]
	if block.Type == "Imbalances" && depth < 4 {
		g := rollGoodsColumn(r, columnFor(entry), depth+1)
		g.Imbalance = entry
		return g
	}
	return TradeGood{Name: entry, Type: block.Type}
}

// tradeGoodsDetail picks a Trade Good Detail prefix (Book 2 p.219): the first of
// the world's trade classes (other than the one that chose the column) carrying
// a detail label, with Hi's "Processed" omitted on Industrial worlds and Va's
// "Exotic" omitted on Asteroid worlds. Returns "" when none applies.
func tradeGoodsDetail(worldTCs []string, sourceTC string) string {
	industrial := slices.Contains(worldTCs, "In")
	asteroid := slices.Contains(worldTCs, "As")
	for _, tc := range worldTCs {
		if tc == sourceTC {
			continue
		}
		label, ok := tradeGoodsDetailLabel[tc]
		if !ok {
			continue
		}
		if tc == "Hi" && industrial {
			continue
		}
		if tc == "Va" && asteroid {
			continue
		}
		return label
	}
	return ""
}
