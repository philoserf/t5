// Package tradecode is the Traveller5 Trade Classification vocabulary (Book 3
// Chart D, p.26): the two-letter codes a world's UWP and context earn it (As, Ag,
// Hi, In, Ri, …). It is the one type and the one canonical set, imported by every
// package that produces or consumes a trade code — worldgen produces them, chargen
// grants homeworld skills from them, trade prices cargo by them, survey and
// systemgen tag them — so the vocabulary cannot drift or a dead code slip in.
//
// A Code is a defined string type over the 46 Chart D codes below. Writing a code
// that is not one of them (chargen once keyed a homeworld skill on "Ds", a code
// worldgen never emits, #329) is a compile error, not a silently dead map entry.
// This is the trade CLASSIFICATION vocabulary only; trade GOODS (internal/trade's
// goods table) are a separate set and are not Codes.
package tradecode

import "strings"

// A Code is one Traveller5 trade classification (Book 3 Chart D, p.26). The valid
// values are the exported constants below; construct codes from those, not from
// string literals, so the compiler rejects an unknown code.
type Code string

// The Chart D codes, in the book's section order (Book 3 p.26). Order is Order's
// order, which is the order OrderTradeCodes renders them in.
const (
	// Planetary.

	As Code = "As" // Asteroid Belt
	De Code = "De" // Desert
	Fl Code = "Fl" // Fluid oceans (non-water)
	Ga Code = "Ga" // Garden
	He Code = "He" // Hellworld
	Ic Code = "Ic" // Ice-capped
	Oc Code = "Oc" // Ocean
	Va Code = "Va" // Vacuum
	Wa Code = "Wa" // Water
	Sa Code = "Sa" // Satellite (far moon)
	Lk Code = "Lk" // Locked (close moon)

	// Population.

	Di Code = "Di" // Dieback
	Ba Code = "Ba" // Barren
	Lo Code = "Lo" // Low population
	Ni Code = "Ni" // Non-industrial
	Ph Code = "Ph" // Pre-high population
	Hi Code = "Hi" // High population

	// Economic.

	Pa Code = "Pa" // Pre-agricultural
	Ag Code = "Ag" // Agricultural
	Na Code = "Na" // Non-agricultural
	Px Code = "Px" // Prison/Exile camp
	Pi Code = "Pi" // Pre-industrial
	In Code = "In" // Industrial
	Po Code = "Po" // Poor
	Pr Code = "Pr" // Pre-rich
	Ri Code = "Ri" // Rich

	// Climate.

	Fr Code = "Fr" // Frozen
	Ho Code = "Ho" // Hot
	Co Code = "Co" // Cold
	Tr Code = "Tr" // Tropic
	Tu Code = "Tu" // Tundra
	Tz Code = "Tz" // Twilight zone

	// Secondary.

	Fa Code = "Fa" // Farming
	Mi Code = "Mi" // Mining
	Mr Code = "Mr" // Military rule
	Pe Code = "Pe" // Penal colony
	Re Code = "Re" // Reserve

	// Political.

	Cp Code = "Cp" // Subsector capital
	Cs Code = "Cs" // Sector capital
	Cx Code = "Cx" // Imperial capital
	Cy Code = "Cy" // Colony

	// Special.

	Fo Code = "Fo" // Forbidden
	Pz Code = "Pz" // Puzzle
	Da Code = "Da" // Dangerous
	Ab Code = "Ab" // Data repository
	An Code = "An" // Ancient site
)

// Order is the canonical set of all 46 Chart D codes, in Chart D section order
// (Book 3 p.26). It is the single source of truth for which codes exist and in
// what order they render; Rank and Valid are built from it.
var Order = []Code{
	As, De, Fl, Ga, He, Ic, Oc, Va, Wa, Sa, Lk, // Planetary
	Di, Ba, Lo, Ni, Ph, Hi, // Population
	Pa, Ag, Na, Px, Pi, In, Po, Pr, Ri, // Economic
	Fr, Ho, Co, Tr, Tu, Tz, // Climate
	Fa, Mi, Mr, Pe, Re, // Secondary
	Cp, Cs, Cx, Cy, // Political
	Fo, Pz, Da, Ab, An, // Special
}

// rank maps each code to its index in Order, for O(1) Rank/Valid.
var rank = func() map[Code]int {
	m := make(map[Code]int, len(Order))
	for i, c := range Order {
		m[c] = i
	}

	return m
}()

// Valid reports whether c is one of the Chart D codes. A code from a string
// literal or from outside the vocabulary is not.
func Valid(c Code) bool {
	_, ok := rank[c]

	return ok
}

// Rank returns c's position in Chart D order (Book 3 p.26), for sorting a code
// list. An unknown code ranks after every known one, so a stray code sorts last
// rather than panicking — the same total behavior the old rankOf had.
func Rank(c Code) int {
	if i, ok := rank[c]; ok {
		return i
	}

	return len(Order)
}

// Join renders a code list as a sep-separated string. It is the one place codes
// cross back into text for a record, so every renderer joins them the same way
// (the codes are already strings, but []Code does not satisfy strings.Join).
func Join(codes []Code, sep string) string {
	parts := make([]string, len(codes))
	for i, c := range codes {
		parts[i] = string(c)
	}

	return strings.Join(parts, sep)
}
