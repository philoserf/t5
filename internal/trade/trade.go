// Package trade implements the Traveller5 Trade and Commerce pricing engine
// (Book 2 pp. 209-221): the Cargo ID, the source-world Cost of goods, the
// market-world Price, and the Actual Value Table that turns an expected Price
// into a realized selling price. Speculative cargo is bought at a Source World
// for its Cost and sold at a Market World for a fraction or multiple of its
// Price. Costs are plain int Cr.
//
// This package covers the value engine; the Random Trade Goods detail tables
// (specific good names) and passenger/freight availability are separate.
package trade

import (
	"fmt"
	"slices"
	"strings"

	"github.com/philoserf/t5/internal/ehex"
)

// baseCost and basePrice anchor the Cost and Price computations (Book 2 p.221).
const (
	baseCost  = 3000
	basePrice = 5000
)

// valueClasses are the trade classifications that affect Cost and Price (Book 2
// p.221), in the chart order used to render a Cargo ID. Every other trade code
// may influence availability but never value.
var valueClasses = []string{"Ag", "As", "Ba", "De", "Fl", "Hi", "Ic", "In", "Lo", "Na", "Ni", "Po", "Ri", "Va"}

// costMod is each value class's per-ton cost modifier at the source world (Book 2
// p.221 A); Ic and Na are members of the value set but move the cost by zero.
var costMod = map[string]int{
	"Ag": -1000, "As": -1000, "Ba": 1000, "De": 1000, "Fl": 1000,
	"Hi": -1000, "Ic": 0, "In": -1000, "Lo": 1000, "Na": 0,
	"Ni": 1000, "Po": -1000, "Ri": 1000, "Va": 1000,
}

// priceMatch maps a source trade class to the market trade classes that shift a
// cargo's Price and the per-match amount (Book 2 p.221 B): Poor lowers the
// price, every other class raises it.
var priceMatch = map[string]struct {
	markets []string
	per     int
}{
	"Ag": {[]string{"Ag", "As", "De", "Hi", "In", "Ri", "Va"}, 1000},
	"As": {[]string{"As", "In", "Ri", "Va"}, 1000},
	"Ba": {[]string{"In"}, 1000},
	"De": {[]string{"De"}, 1000},
	"Fl": {[]string{"Fl", "In"}, 1000},
	"Hi": {[]string{"Hi"}, 1000},
	"In": {[]string{"Ag", "As", "De", "Fl", "Hi", "In", "Ri", "Va"}, 1000},
	"Na": {[]string{"As", "De", "Va"}, 1000},
	"Po": {[]string{"Ag", "Hi", "In", "Ri"}, -1000},
	"Ri": {[]string{"Ag", "De", "Hi", "In", "Ri"}, 1000},
	"Va": {[]string{"As", "In", "Va"}, 1000},
}

// actualValue maps an effective Flux (Flux + Broker DM, clamped to -5..+8) to the
// percentage of Price a sale realizes (Book 2 p.221 Actual Value Table).
var actualValue = map[int]int{
	-5: 40, -4: 50, -3: 70, -2: 80, -1: 90, 0: 100,
	1: 110, 2: 120, 3: 130, 4: 150, 5: 170, 6: 200, 7: 300, 8: 400,
}

// ValueClasses filters a world's trade codes to the value-relevant set, in the
// chart order used by a Cargo ID (Book 2 p.221).
func ValueClasses(tcs []string) []string {
	var out []string
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
func Cost(sourceTL int, sourceTCs []string) int {
	c := baseCost + sourceTL*100
	for _, code := range ValueClasses(sourceTCs) {
		c += costMod[code]
	}
	return max(c, 0)
}

// Price is the per-ton expected selling price of a source world's goods at a
// market world (Book 2 p.221 B): a Cr5,000 base plus Cr1,000 for each source
// class matched by a market class (Poor subtracts), all scaled by a Tech Level
// effect of 10% per level the source exceeds the market. Floored at zero.
func Price(sourceTL, marketTL int, sourceTCs, marketTCs []string) int {
	p := basePrice
	for _, code := range ValueClasses(sourceTCs) {
		m, ok := priceMatch[code]
		if !ok {
			continue
		}
		for _, mc := range m.markets {
			if slices.Contains(marketTCs, mc) {
				p += m.per
			}
		}
	}
	p += (sourceTL - marketTL) * p / 10
	return max(p, 0)
}

// BrokerDM is the Actual Value bonus a broker of the given skill provides: half
// skill rounded up, capped at +4 (Book 2 p.221).
func BrokerDM(brokerSkill int) int {
	return min((brokerSkill+1)/2, 4)
}

// ActualValuePercent is the percentage of Price a sale realizes at the given
// Flux with a broker of the given skill (Book 2 p.221 Actual Value Table). The
// effective Flux is clamped to the table's -5..+8 range.
func ActualValuePercent(flux, brokerSkill int) int {
	e := clamp(flux+BrokerDM(brokerSkill), -5, 8)
	return actualValue[e]
}

// SellingPrice is the realized per-ton sale price: the expected Price times the
// Actual Value percentage (Book 2 p.221).
func SellingPrice(price, flux, brokerSkill int) int {
	return price * ActualValuePercent(flux, brokerSkill) / 100
}

// CargoID renders a cargo's identity (Book 2 p.221): the source Tech Level as an
// eHex digit, its value trade classes, the computed per-ton Cost, and a trailing
// allegiance code when the source is not Imperial. E.g. "8-De Hi In Na Po Cr1,800".
func CargoID(sourceTL int, sourceTCs []string, allegiance string) string {
	id := fmt.Sprintf("%s-%s Cr%s",
		ehex.Format(sourceTL), strings.Join(ValueClasses(sourceTCs), " "), commas(Cost(sourceTL, sourceTCs)))
	if allegiance != "" && allegiance != "Im" {
		id += " " + allegiance
	}
	return id
}

// clamp constrains v to [lo, hi].
func clamp(v, lo, hi int) int {
	return min(max(v, lo), hi)
}

// commas renders an integer with thousands separators, e.g. 1800 -> "1,800".
func commas(n int) string {
	s := fmt.Sprintf("%d", n)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	var b strings.Builder
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
