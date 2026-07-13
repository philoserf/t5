package worldgen

import (
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
// UWP: climate / habitable-zone (Frozen, Hot, Cold, Tropic, Tundra, Twilight
// Zone, Farming) and non-mainworld / orbit status (Satellite, Locked, Mining,
// Penal), which await systemgen per-world placement (catalog #15), and the
// referee-assigned political and special codes.
var tcRules = []tcRule{
	// Planetary.
	{code: "As", siz: "0", atm: "0", hyd: "0"},
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

// TradeClassifications returns the two-letter trade classification codes a
// mainworld qualifies for, in table order. It reports only classifications
// determinable from the UWP; see tcRules for what is intentionally left out.
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
