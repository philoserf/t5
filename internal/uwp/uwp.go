// Package uwp models the Universal World Profile — the compact StSAHPGL-T
// summary of a Traveller mainworld (Starport, Size, Atmosphere, Hydrographics,
// Population, Government, Law, and Tech Level), e.g. Regina's "A788899-C".
package uwp

import (
	"strings"

	"github.com/philoserf/t5/internal/ehex"
)

// A Profile is a mainworld's Universal World Profile. Every field except
// Starport is an eHex value; Starport is a literal quality letter (A-E, X).
type Profile struct {
	Starport      byte
	Size          int
	Atmosphere    int
	Hydrographics int
	Population    int
	Government    int
	Law           int
	TechLevel     int
}

// String renders the profile in standard UWP notation: the starport letter,
// the six eHex characteristic digits, a hyphen, and the eHex Tech Level —
// for example "A788899-C".
func (p Profile) String() string {
	var b strings.Builder
	b.Grow(9)
	b.WriteByte(p.Starport)
	b.WriteByte(ehex.Digit(p.Size))
	b.WriteByte(ehex.Digit(p.Atmosphere))
	b.WriteByte(ehex.Digit(p.Hydrographics))
	b.WriteByte(ehex.Digit(p.Population))
	b.WriteByte(ehex.Digit(p.Government))
	b.WriteByte(ehex.Digit(p.Law))
	b.WriteByte('-')
	b.WriteByte(ehex.Digit(p.TechLevel))
	return b.String()
}
