// Package trade implements the Traveller5 Trade and Commerce pricing engine
// (Book 2 pp. 209-221): the Cargo ID, the source-world Cost of goods, the
// market-world Price, and the Actual Value Table that turns an expected Price
// into a realized selling price. Speculative cargo is bought at a Source World
// for its Cost and sold at a Market World for a fraction or multiple of its
// Price. Costs are plain int Cr.
//
// The package also covers the Random Trade Goods chart that names a cargo
// (goods.go), the passenger, freight, and mail economics (shipping.go,
// contracts.go), and the broker who negotiates a sale (brokers.go).
package trade

import (
	"slices"
	"strconv"
	"strings"

	"github.com/philoserf/t5/internal/ehex"
	"github.com/philoserf/t5/internal/tradecode"
)

// baseCost and basePrice anchor the Cost and Price computations (Book 2 p.221).
const (
	baseCost  = 3000
	basePrice = 5000
)

// valueClasses are the trade classifications that affect Cost and Price (Book 2
// p.221), in the chart order used to render a Cargo ID. Every other trade code
// may influence availability but never value.
var valueClasses = []tradecode.Code{
	tradecode.Ag,
	tradecode.As,
	tradecode.Ba,
	tradecode.De,
	tradecode.Fl,
	tradecode.Hi,
	tradecode.Ic,
	tradecode.In,
	tradecode.Lo,
	tradecode.Na,
	tradecode.Ni,
	tradecode.Po,
	tradecode.Ri,
	tradecode.Va,
}

// costMod is each value class's per-ton cost modifier at the source world (Book 2
// p.221 A); Ic and Na are members of the value set but move the cost by zero.
var costMod = map[tradecode.Code]int{
	tradecode.Ag: -1000, tradecode.As: -1000, tradecode.Ba: 1000, tradecode.De: 1000, tradecode.Fl: 1000,
	tradecode.Hi: -1000, tradecode.Ic: 0, tradecode.In: -1000, tradecode.Lo: 1000, tradecode.Na: 0,
	tradecode.Ni: 1000, tradecode.Po: -1000, tradecode.Ri: 1000, tradecode.Va: 1000,
}

// priceMatch maps a source trade class to the market trade classes that shift a
// cargo's Price and the per-match amount (Book 2 p.221 B): Poor lowers the
// price, every other class raises it.
var priceMatch = map[tradecode.Code]struct {
	markets []tradecode.Code
	per     int
}{
	tradecode.Ag: {
		[]tradecode.Code{
			tradecode.Ag,
			tradecode.As,
			tradecode.De,
			tradecode.Hi,
			tradecode.In,
			tradecode.Ri,
			tradecode.Va,
		},
		1000,
	},
	tradecode.As: {[]tradecode.Code{tradecode.As, tradecode.In, tradecode.Ri, tradecode.Va}, 1000},
	tradecode.Ba: {[]tradecode.Code{tradecode.In}, 1000},
	tradecode.De: {[]tradecode.Code{tradecode.De}, 1000},
	tradecode.Fl: {[]tradecode.Code{tradecode.Fl, tradecode.In}, 1000},
	tradecode.Hi: {[]tradecode.Code{tradecode.Hi}, 1000},
	tradecode.In: {
		[]tradecode.Code{
			tradecode.Ag,
			tradecode.As,
			tradecode.De,
			tradecode.Fl,
			tradecode.Hi,
			tradecode.In,
			tradecode.Ri,
			tradecode.Va,
		},
		1000,
	},
	tradecode.Na: {[]tradecode.Code{tradecode.As, tradecode.De, tradecode.Va}, 1000},
	tradecode.Po: {[]tradecode.Code{tradecode.Ag, tradecode.Hi, tradecode.In, tradecode.Ri}, -1000},
	tradecode.Ri: {
		[]tradecode.Code{tradecode.Ag, tradecode.De, tradecode.Hi, tradecode.In, tradecode.Ri},
		1000,
	},
	tradecode.Va: {[]tradecode.Code{tradecode.As, tradecode.In, tradecode.Va}, 1000},
}

// actualValue is the percentage of Price a sale realizes at each effective Flux
// (Flux + Broker DM, clamped to -5..+8), indexed by effective-Flux + 5 (Book 2
// p.221 Actual Value Table). This is not a raw-Flux table — the Broker DM folds
// in before the lookup, widening it past Flux's +5 — so dice.FluxIndex, which
// maps a bare Flux onto a 13-entry table, does not apply here.
var actualValue = [14]int{40, 50, 70, 80, 90, 100, 110, 120, 130, 150, 170, 200, 300, 400}

// ValueClasses filters a world's trade codes to the value-relevant set, in the
// chart order used by a Cargo ID (Book 2 p.221). It also de-duplicates, since it
// walks the chart order rather than the caller's codes.
func ValueClasses(tcs []tradecode.Code) []tradecode.Code {
	out := make([]tradecode.Code, 0, len(tcs))
	for _, code := range valueClasses {
		if slices.Contains(tcs, code) {
			out = append(out, code)
		}
	}

	return out
}

// Cost is the per-ton purchase cost of goods at a source world (Book 2 p.221 A):
// a Cr3,000 base, each value trade class's cost modifier, and Cr100 per source
// Tech Level, floored at zero.
func Cost(sourceTL int, sourceTCs []tradecode.Code) int {
	return costOf(sourceTL, ValueClasses(sourceTCs))
}

// costOf sums the base cost, the Cr100/TL effect, and the cost modifiers over
// already-filtered value classes, floored at zero (Book 2 p.221 A).
func costOf(sourceTL int, valueClasses []tradecode.Code) int {
	c := baseCost + sourceTL*100
	for _, code := range valueClasses {
		c += costMod[code]
	}

	return max(c, 0)
}

// Price is the per-ton expected selling price of a source world's goods at a
// market world (Book 2 p.221 B): a Cr5,000 base plus Cr1,000 for each source
// class matched by a market class (Poor subtracts), all scaled by a Tech Level
// effect of 10% per level the source exceeds the market. Floored at zero.
func Price(sourceTL, marketTL int, sourceTCs, marketTCs []tradecode.Code) int {
	p := basePrice
	// A class with no priceMatch row yields the zero value, whose nil markets make
	// the inner loop a no-op — no presence check needed.
	for _, code := range ValueClasses(sourceTCs) {
		m := priceMatch[code]
		for _, mc := range m.markets {
			if slices.Contains(marketTCs, mc) {
				p += m.per
			}
		}
	}

	p += (sourceTL - marketTL) * p / 10

	return max(p, 0)
}

// ActualValuePercent is the percentage of Price a sale realizes at the given
// Flux with a broker of the given skill (Book 2 p.221 Actual Value Table). The
// effective Flux is clamped to the table's -5..+8 range.
func ActualValuePercent(flux, brokerSkill int) int {
	e := clamp(flux+BrokerDM(brokerSkill), -5, 8)

	return actualValue[e+5]
}

// SellingPrice is the realized per-ton sale price: the expected Price times the
// Actual Value percentage (Book 2 p.221).
func SellingPrice(price, flux, brokerSkill int) int {
	return price * ActualValuePercent(flux, brokerSkill) / 100
}

// EstimateActualValue is the Actual Value percentage range a merchant with Trader
// skill can foresee once one of the two Flux dice is thrown early (Book 2 pp.210,
// 221): with the first die known, the second still varies 1-6, bounding the sale.
// The Broker DM applies as in ActualValuePercent. E.g. a first die of 6 with no
// broker bounds the outcome to 100%-170%.
func EstimateActualValue(firstDie, brokerSkill int) (int, int) {
	minPct := 1 << 30
	maxPct := 0

	for second := 1; second <= 6; second++ {
		pct := ActualValuePercent(firstDie-second, brokerSkill)
		minPct = min(minPct, pct)
		maxPct = max(maxPct, pct)
	}

	return minPct, maxPct
}

// CargoID renders a cargo's identity (Book 2 p.221): the source Tech Level as an
// eHex digit, its value trade classes, the computed per-ton Cost, and a trailing
// allegiance code when the source is not Imperial. E.g. "8-De Hi In Na Po Cr1,800".
//
// Only the value classes are listed, matching the p.221 chart and its worked
// examples (Efate's Hi In An renders "D-Hi In Cr2,300" — the non-value An is
// dropped). The looser p.210 prose examples keep some non-value codes while still
// dropping An; the book is inconsistent, and the mechanical chart wins here.
func CargoID(sourceTL int, sourceTCs []tradecode.Code, allegiance string) string {
	vc := ValueClasses(sourceTCs)

	id := ehex.Format(sourceTL)
	if len(vc) > 0 { // a world with no value class has no class field, not an empty one
		id += "-" + tradecode.Join(vc, " ")
	}

	id += " Cr" + commas(costOf(sourceTL, vc))
	if allegiance != "" && allegiance != "Im" {
		id += " " + allegiance
	}

	return id
}

// ImbalanceBonus is the premium a cargo of Imbalance goods earns when sold into a
// market that carries the trade class whose oversupply produced it (Book 2 p.211:
// "If these goods are sold on a market world with this Trade Classification,
// increase their Price +Cr1,000"). Add it to Price; it is zero for ordinary goods.
func ImbalanceBonus(g Good, marketTCs []tradecode.Code) int {
	if g.Imbalance != "" && slices.Contains(marketTCs, tradecode.Code(g.Imbalance)) {
		return 1_000
	}

	return 0
}

// clamp constrains v to [lo, hi].
func clamp(v, lo, hi int) int {
	return min(max(v, lo), hi)
}

// commas renders an integer with thousands separators, e.g. 1800 -> "1,800".
func commas(n int) string {
	s := strconv.Itoa(n)

	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}

	var b strings.Builder
	b.Grow(len(s) + len(s)/3)

	for i := range len(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}

		b.WriteByte(s[i])
	}

	if neg {
		return "-" + b.String()
	}

	return b.String()
}
