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
// A characteristic outside eHex range renders as "?" rather than panicking:
// Profile is an exported model whose fields a caller may set directly, and a
// String method must not crash.
func (p Profile) String() string {
	var b strings.Builder
	b.Grow(9)
	b.WriteByte(p.Starport)
	b.WriteString(ehex.Format(p.Size))
	b.WriteString(ehex.Format(p.Atmosphere))
	b.WriteString(ehex.Format(p.Hydrographics))
	b.WriteString(ehex.Format(p.Population))
	b.WriteString(ehex.Format(p.Government))
	b.WriteString(ehex.Format(p.Law))
	b.WriteByte('-')
	b.WriteString(ehex.Format(p.TechLevel))
	return b.String()
}
