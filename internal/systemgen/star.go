package systemgen

import (
	"fmt"

	"github.com/philoserf/t5/internal/dice"
)

// A Star is a stellar classification: a spectral Type (O, B, A, F, G, K, M, or
// BD for a brown/black dwarf), a spectral Decimal 0-9, and a luminosity Size
// (Ia, Ib, II, III, IV, V, VI, or D for a white dwarf). Decimal is -1 when it
// does not apply — a BD, or a Size-D star, which ignores its decimal.
type Star struct {
	Type    string
	Decimal int
	Size    string
}

// String renders the classification, e.g. "F7 V", "M0 III", "G D", or "BD".
func (s Star) String() string {
	switch {
	case s.Type == "BD":
		return "BD"
	case s.Size == "D":
		return s.Type + " D"
	default:
		return fmt.Sprintf("%s%d %s", s.Type, s.Decimal, s.Size)
	}
}

// spectralRow is one row of the Spectral Type and Size table (Book 3 p. 28):
// the spectral type for that Flux, and the luminosity size indexed by the
// star's spectral column (O, B, A, F, G, K, M).
type spectralRow struct {
	sp    string
	sizes [7]string
}

// column indexes sizes by spectral type.
var column = map[string]int{"O": 0, "B": 1, "A": 2, "F": 3, "G": 4, "K": 5, "M": 6}

// spectralTable is indexed by Flux+6, covering Flux -6..+8. Transcribed from
// the Spectral Type and Size table on Book 3 p. 28 and verified against a
// rendered image of that page (Flux -1 is type G, not F).
//
// Note: the book's Regina worked example states "Primary Spectral Type = Flux
// = -1 = F", which disagrees with this table (a known erratum in the example —
// its stated Flux is inconsistent with the table it references). The printed
// table is the generation rule and is authoritative here; do not "correct" the
// letters to match the example.
var spectralTable = []spectralRow{
	{"OB", [7]string{"Ia", "Ia", "Ia", "II", "II", "II", "II"}},
	{"A", [7]string{"Ia", "Ia", "Ia", "II", "II", "II", "II"}},
	{"A", [7]string{"Ib", "Ib", "Ib", "III", "III", "III", "II"}},
	{"F", [7]string{"II", "II", "II", "IV", "IV", "IV", "II"}},
	{"F", [7]string{"III", "III", "III", "V", "V", "V", "III"}},
	{"G", [7]string{"III", "III", "IV", "V", "V", "V", "V"}},
	{"G", [7]string{"III", "III", "V", "V", "V", "V", "V"}},
	{"K", [7]string{"V", "III", "V", "V", "V", "V", "V"}},
	{"K", [7]string{"V", "V", "V", "V", "V", "V", "V"}},
	{"M", [7]string{"V", "V", "V", "V", "V", "V", "V"}},
	{"M", [7]string{"IV", "IV", "V", "VI", "VI", "VI", "VI"}},
	{"M", [7]string{"D", "D", "D", "D", "D", "D", "D"}},
	{"BD", [7]string{"IV", "IV", "V", "VI", "VI", "VI", "VI"}},
	{"BD", [7]string{"IV", "IV", "V", "VI", "VI", "VI", "VI"}},
	{"BD", [7]string{"IV", "IV", "V", "VI", "VI", "VI", "VI"}},
}

// row returns the table row for a Flux value, clamped to the table's range.
func row(flux int) spectralRow {
	return spectralTable[min(max(flux, -6), 8)+6]
}

// classify builds a Star from already-rolled inputs: the Flux that selects the
// spectral type, the Flux that selects the size, the spectral decimal, and
// (for an "OB" type) whether to pick B rather than O. Keeping it pure lets the
// table and its special cases be tested directly.
func classify(typeFlux, sizeFlux, decimal int, pickB bool) Star {
	sp := row(typeFlux).sp
	if sp == "BD" {
		return Star{Type: "BD", Decimal: -1}
	}
	if sp == "OB" {
		if pickB {
			sp = "B"
		} else {
			sp = "O"
		}
	}

	size := row(sizeFlux).sizes[column[sp]]
	if size == "D" {
		return Star{Type: sp, Decimal: -1, Size: "D"}
	}

	// Forbidden type/size combinations fall back to main-sequence V (p. 28).
	switch {
	case size == "IV" && (sp == "M" || (sp == "K" && decimal >= 5)):
		size = "V"
	case size == "VI" && (sp == "A" || (sp == "F" && decimal <= 4)):
		size = "V"
	}
	return Star{Type: sp, Decimal: decimal, Size: size}
}

// rollStar rolls a complete star. A primary rolls Flux for its type and size;
// any other star derives from the primary's Flux values (type = primaryType +
// 1D-1, size = primarySize + 1D+2). It returns the star and the Flux values a
// companion would build on. A BD short-circuits: its decimal, size, and OB
// selection are not rolled.
func rollStar(r *dice.Roller, primary bool, primaryType, primarySize int) (star Star, typeFlux, sizeFlux int) {
	if primary {
		typeFlux = r.Flux()
	} else {
		typeFlux = primaryType + r.Die() - 1
	}
	sp := row(typeFlux).sp
	if sp == "BD" {
		return Star{Type: "BD", Decimal: -1}, typeFlux, 0
	}

	decimal := r.EvenDist0to9()
	if primary {
		sizeFlux = r.Flux()
	} else {
		sizeFlux = primarySize + r.Die() + 2
	}
	pickB := false
	if sp == "OB" {
		pickB = r.Die() >= 4
	}
	return classify(typeFlux, sizeFlux, decimal, pickB), typeFlux, sizeFlux
}
