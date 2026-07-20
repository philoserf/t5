package worldgen

import (
	"slices"
	"strings"

	"github.com/philoserf/t5/internal/ehex"
	"github.com/philoserf/t5/internal/uwp"
)

// A tcRule qualifies a world for a trade classification when each named
// characteristic's eHex digit appears in the rule's allowed set (and its
// starport letter is in port). An empty set means the characteristic is
// unconstrained.
type tcRule struct {
	code                         string
	siz, atm, hyd, pop, gov, law string // eHex-digit sets; "" = unconstrained
	port                         string // allowed Starport letters; "" = any
}

// tcRules holds the trade classifications that can be decided from a mainworld
// UWP alone (Book 3, Chart D, p. 26), in table order.
//
// Deliberately excluded are classifications that need information beyond the
// UWP. The climate / habitable-zone codes (Tropic, Tundra, Frozen, Twilight
// Zone) are supplied by ClimateCodes once systemgen places the mainworld in an
// orbit; non-mainworld / orbit status (Satellite, Locked, Mining, Penal,
// Farming) awaits the rest of systemgen per-world placement (catalog #15); and
// the political and special codes are referee-assigned.
var tcRules = []tcRule{
	// Planetary. As (asteroid belt) is deliberately absent: it is not a UWP code
	// but a body-type fact, since a belt and a tiny Size-0 world share the same
	// profile. TradeClassificationsWithContext adds it from the caller's belt
	// context (#324/#328).
	{code: "De", atm: "23456789", hyd: "0"},
	{code: "Fl", atm: "ABC", hyd: "123456789A"},
	{code: "Ga", siz: "678", atm: "568", hyd: "567"},
	{code: "He", siz: "3456789ABC", atm: "2479ABC", hyd: "012"},
	{code: "Ic", atm: "01", hyd: "123456789A"},
	{code: "Oc", siz: "ABCDEF", atm: "3456789DEF", hyd: "A"},
	{code: "Va", atm: "0"},
	{code: "Wa", siz: "3456789", atm: "3456789DEF", hyd: "A"},
	// Population. Di and Ba share the Pop0/Gov0/Law0 core, split by starport so
	// exactly one fires (Book 3 p.26: Ba requires Starport E or X).
	{code: "Di", pop: "0", gov: "0", law: "0", port: "ABCD"},
	{code: "Ba", pop: "0", gov: "0", law: "0", port: "EX"},
	{code: "Lo", pop: "123"},
	{code: "Ni", pop: "456"},
	{code: "Ph", pop: "8"},
	{code: "Hi", pop: "9ABCDEF"},
	{code: "Pa", atm: "456789", hyd: "45678", pop: "48"},
	{code: "Ag", atm: "456789", hyd: "45678", pop: "567"},
	{code: "Na", atm: "0123", hyd: "0123", pop: "6789ABCDEF"},
	{code: "Px", atm: "23AB", hyd: "12345", pop: "3456", law: "6789"}, // Prison/Exile Camp
	// Economic.
	{code: "Pi", atm: "012479", pop: "78"},
	{code: "In", atm: "012479ABC", pop: "9ABCDEF"},
	{code: "Po", atm: "2345", hyd: "0123"},
	{code: "Pr", atm: "68", pop: "59"},
	{code: "Ri", atm: "68", pop: "678"},
	// Secondary.
	{code: "Re", pop: "01234", gov: "6", law: "045"}, // Reserve
}

// chartDOrder is every trade code in Book 3 Chart D (p.26) section order:
// Planetary, Population, Economic, Climate, Secondary, Political, Special. A code's
// position in this list is its canonical rank, which OrderTradeCodes sorts by — so
// a world's codes render in the book's order however they were accumulated (base
// codes at generation, zone codes appended after, climate and satellite codes at
// placement, a capital code stamped by the region survey later still).
var chartDOrder = []string{
	// Planetary.
	"As", "De", "Fl", "Ga", "He", "Ic", "Oc", "Va", "Wa", "Sa", "Lk",
	// Population.
	"Di", "Ba", "Lo", "Ni", "Ph", "Hi",
	// Economic.
	"Pa", "Ag", "Na", "Px", "Pi", "In", "Po", "Pr", "Ri",
	// Climate.
	"Fr", "Ho", "Co", "Tr", "Tu", "Tz",
	// Secondary.
	"Fa", "Mi", "Mr", "Pe", "Re",
	// Political.
	"Cp", "Cs", "Cx", "Cy",
	// Special.
	"Fo", "Pz", "Da", "Ab", "An",
}

// chartDRank maps a code to its Chart D position; a code not in the chart sorts
// after every one that is (kept in its original relative order).
var chartDRank = func() map[string]int {
	m := make(map[string]int, len(chartDOrder))
	for i, code := range chartDOrder {
		m[code] = i
	}

	return m
}()

// OrderTradeCodes returns the codes sorted into Chart D order (Book 3 p.26). It
// does not mutate the input, and it is what renderers call so a world's stored code
// order — which follows the accumulation sequence, not the book's — never reaches a
// record. Unknown codes keep their relative order, after the known ones.
func OrderTradeCodes(tcs []string) []string {
	out := slices.Clone(tcs)
	slices.SortStableFunc(out, func(a, b string) int {
		return rankOf(a) - rankOf(b)
	})

	return out
}

func rankOf(code string) int {
	if r, ok := chartDRank[code]; ok {
		return r
	}

	return len(chartDOrder)
}

// TradeClassifications returns the two-letter trade classification codes a
// mainworld qualifies for, in table order. It reports only classifications
// determinable from the UWP alone; see tcRules for what is intentionally left
// out — including As, which needs the caller's belt context and is added by
// TradeClassificationsWithContext, not here.
func TradeClassifications(p uwp.Profile) []string {
	var out []string

	for _, r := range tcRules {
		if allows(r.siz, p.Size) && allows(r.atm, p.Atmosphere) &&
			allows(r.hyd, p.Hydrographics) && allows(r.pop, p.Population) &&
			allows(r.gov, p.Government) && allows(r.law, p.Law) &&
			portAllows(r.port, p.Starport) {
			out = append(out, r.code)
		}
	}

	return out
}

// portAllows reports whether starport letter s is in the allowed set (a literal
// letter, not an eHex digit). An empty set allows any starport.
func portAllows(set string, s byte) bool {
	if set == "" {
		return true
	}

	return strings.IndexByte(set, s) >= 0
}

// allows reports whether value v satisfies an allowed eHex-digit set. An empty
// set allows anything; a value outside eHex range matches no specific set (and
// never panics, so the exported classifier is safe on hand-built profiles).
func allows(set string, v int) bool {
	if set == "" {
		return true
	}

	if v < 0 || v > ehex.Max {
		return false
	}

	return strings.IndexByte(set, ehex.Digit(v)) >= 0
}
