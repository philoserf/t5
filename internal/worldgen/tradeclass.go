package worldgen

import (
	"slices"
	"strings"

	"github.com/philoserf/t5/internal/ehex"
	"github.com/philoserf/t5/internal/tradecode"
	"github.com/philoserf/t5/internal/uwp"
)

// A tcRule qualifies a world for a trade classification when each named
// characteristic's eHex digit appears in the rule's allowed set (and its
// starport letter is in port). An empty set means the characteristic is
// unconstrained.
type tcRule struct {
	code                         tradecode.Code
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
	// Planetary. As (asteroid belt) is a Chart-D UWP code — Siz0/Atm0/Hyd0 (Book 3
	// p.26, and the p.16 rule that a Size-0 mainworld IS a belt). It is correct for
	// a mainworld and for a genuine Planetoids belt, but a Size-0 secondary Worldlet
	// renders the same St000... without being a belt, so TradeClassificationsWithContext
	// strips As from a non-belt non-mainworld (#324).
	{code: tradecode.As, siz: "0", atm: "0", hyd: "0"},
	{code: tradecode.De, atm: "23456789", hyd: "0"},
	{code: tradecode.Fl, atm: "ABC", hyd: "123456789A"},
	{code: tradecode.Ga, siz: "678", atm: "568", hyd: "567"},
	{code: tradecode.He, siz: "3456789ABC", atm: "2479ABC", hyd: "012"},
	{code: tradecode.Ic, atm: "01", hyd: "123456789A"},
	{code: tradecode.Oc, siz: "ABCDEF", atm: "3456789DEF", hyd: "A"},
	{code: tradecode.Va, atm: "0"},
	{code: tradecode.Wa, siz: "3456789", atm: "3456789DEF", hyd: "A"},
	// Population. Di and Ba share the Pop0/Gov0/Law0 core, split by starport so
	// exactly one fires (Book 3 p.26: Ba requires Starport E or X).
	{code: tradecode.Di, pop: "0", gov: "0", law: "0", port: "ABCD"},
	{code: tradecode.Ba, pop: "0", gov: "0", law: "0", port: "EX"},
	{code: tradecode.Lo, pop: "123"},
	{code: tradecode.Ni, pop: "456"},
	{code: tradecode.Ph, pop: "8"},
	{code: tradecode.Hi, pop: "9ABCDEF"},
	{code: tradecode.Pa, atm: "456789", hyd: "45678", pop: "48"},
	{code: tradecode.Ag, atm: "456789", hyd: "45678", pop: "567"},
	{code: tradecode.Na, atm: "0123", hyd: "0123", pop: "6789ABCDEF"},
	{code: tradecode.Px, atm: "23AB", hyd: "12345", pop: "3456", law: "6789"}, // Prison/Exile Camp
	// Economic.
	{code: tradecode.Pi, atm: "012479", pop: "78"},
	{code: tradecode.In, atm: "012479ABC", pop: "9ABCDEF"},
	{code: tradecode.Po, atm: "2345", hyd: "0123"},
	{code: tradecode.Pr, atm: "68", pop: "59"},
	{code: tradecode.Ri, atm: "68", pop: "678"},
	// Secondary.
	{code: tradecode.Re, pop: "01234", gov: "6", law: "045"}, // Reserve
}

// OrderTradeCodes returns the codes sorted into Chart D order (Book 3 p.26). It
// does not mutate the input, and it is what renderers call so a world's stored code
// order — which follows the accumulation sequence, not the book's — never reaches a
// record. Unknown codes keep their relative order, after the known ones. The order
// is tradecode.Order, ranked by tradecode.Rank — the single source of truth.
func OrderTradeCodes(tcs []tradecode.Code) []tradecode.Code {
	out := slices.Clone(tcs)
	slices.SortStableFunc(out, func(a, b tradecode.Code) int {
		return tradecode.Rank(a) - tradecode.Rank(b)
	})

	return out
}

// TradeClassifications returns the two-letter trade classification codes a
// mainworld qualifies for, in table order. It reports only classifications
// determinable from the UWP; see tcRules for what is intentionally left out. It
// emits As for any Size0/Atm0/Hyd0 profile (correct for a mainworld, Book 3 p.16);
// TradeClassificationsWithContext strips As from a secondary world that is not a
// belt, since a Size-0 Worldlet shares the profile without being one (#324).
func TradeClassifications(p uwp.Profile) []tradecode.Code {
	var out []tradecode.Code

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
