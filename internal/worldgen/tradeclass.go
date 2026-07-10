package worldgen

import (
	"strings"

	"github.com/philoserf/t5/internal/ehex"
	"github.com/philoserf/t5/internal/uwp"
)

// A tcRule qualifies a world for a trade classification when each named
// characteristic's eHex digit appears in the rule's allowed set. An empty set
// means the characteristic is unconstrained. Only Size, Atmosphere,
// Hydrographics, and Population are needed for the UWP-determinable
// classifications below.
type tcRule struct {
	code               string
	siz, atm, hyd, pop string
}

// tcRules holds the trade classifications that can be decided from a mainworld
// UWP alone (Book 3, "Trade Classifications" table, p. 25), in table order.
//
// Deliberately excluded are classifications that need information beyond the
// UWP — climate / habitable-zone (Frozen, Hot, Cold, Tropic, Tundra, Twilight
// Zone, Farming), mainworld or orbit status (Satellite, Locked, Mining), the
// starport-qualified Barren/Dieback pair, and referee-assigned political and
// special codes. Emitting those from a bare UWP would be a guess, so they are
// left to a fuller system-aware pass.
var tcRules = []tcRule{
	// Planetary.
	{"As", "0", "0", "0", ""},
	{"De", "", "23456789", "0", ""},
	{"Fl", "", "ABC", "123456789A", ""},
	{"Ga", "678", "568", "567", ""},
	{"He", "3456789ABC", "2479ABC", "012", ""},
	{"Ic", "", "01", "123456789A", ""},
	{"Oc", "ABCDEF", "3456789DEF", "A", ""},
	{"Va", "", "0", "", ""},
	{"Wa", "3456789", "3456789DEF", "A", ""},
	// Population.
	{"Lo", "", "", "", "123"},
	{"Ni", "", "", "", "456"},
	{"Ph", "", "", "", "8"},
	{"Hi", "", "", "", "9ABCDEF"},
	{"Pa", "", "456789", "45678", "48"},
	{"Ag", "", "456789", "45678", "567"},
	{"Na", "", "0123", "0123", "6789ABCDEF"},
	// Economic.
	{"Pi", "", "012479", "", "78"},
	{"In", "", "012479ABC", "", "9ABCDEF"},
	{"Po", "", "2345", "0123", ""},
	{"Pr", "", "68", "", "59"},
	{"Ri", "", "68", "", "678"},
}

// TradeClassifications returns the two-letter trade classification codes a
// mainworld qualifies for, in table order. It reports only classifications
// determinable from the UWP; see tcRules for what is intentionally left out.
func TradeClassifications(p uwp.Profile) []string {
	var out []string
	for _, r := range tcRules {
		if allows(r.siz, p.Size) && allows(r.atm, p.Atmosphere) &&
			allows(r.hyd, p.Hydrographics) && allows(r.pop, p.Population) {
			out = append(out, r.code)
		}
	}
	return out
}

// allows reports whether value v satisfies an allowed eHex-digit set. An empty
// set allows anything.
func allows(set string, v int) bool {
	return set == "" || strings.IndexByte(set, ehex.Digit(v)) >= 0
}
