// Package uwp models the Universal World Profile — the compact StSAHPGL-T
// summary of a Traveller world (Starport, Size, Atmosphere, Hydrographics,
// Population, Government, Law, and Tech Level), e.g. Regina's "A788899-C".
// Mainworlds and the secondary worlds sharing their system use the same form.
package uwp

import (
	"strings"

	"github.com/philoserf/t5/internal/ehex"
)

// A Profile is a world's Universal World Profile. Every field except Starport
// is an eHex value; Starport is a literal port letter, and its domain depends on
// the kind of world: a mainworld carries a starport quality A-E or X (Book 3
// table 1), a secondary world a spaceport class F, G, H or Y (Book 3 p.29 table
// 1B). Code switching on this field must cover both sets — worldgen's portTable
// does — because a Profile alone does not say which kind of world it describes.
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
// A characteristic outside eHex range — or a Starport outside portLetters —
// renders as "?" rather than panicking: Profile is an exported model whose
// fields a caller may set directly, and a String method must not crash.
func (p Profile) String() string {
	var b strings.Builder
	b.Grow(9)
	b.WriteByte(formatStarport(p.Starport))
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

// portLetters is the Starport field's domain: the mainworld starport qualities
// A-E and X, plus the spaceport letters F, G, H and Y that worldgen's secondary
// worlds store in the same field (Book 3 p.29 table 1B).
const portLetters = "ABCDEXFGHY"

// formatStarport returns c when it is a port letter, and '?' otherwise. Like
// ehex.Format it never panics, so it is safe on the display path where the byte
// may come from a caller-built Profile — including the zero value, whose unset
// Starport would otherwise emit a NUL into a piped record stream.
func formatStarport(c byte) byte {
	if strings.IndexByte(portLetters, c) < 0 {
		return '?'
	}

	return c
}
