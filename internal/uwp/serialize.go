package uwp

import (
	"fmt"

	"github.com/philoserf/t5/internal/ehex"
)

// Serialization for Profile (#327). String is the display path:
// total, and it renders an out-of-domain field as '?' so it never panics on a
// caller-built value. That makes it lossy, so a rendered profile could not be read
// back. The MarshalText/UnmarshalText pair (encoding.TextMarshaler /
// TextUnmarshaler) is the persistence path: non-lossy and reversible, so a stored
// profile round-trips. For a Valid profile the two produce the same nine bytes;
// they differ only off-domain, where String emits '?' and MarshalText errors
// rather than write bytes UnmarshalText could not read back.

// characteristicPositions are the string offsets of the seven eHex characteristics
// in the StSAHPGL-T form "Xdddddd-d": Size..Law at 1-6 and Tech Level at 8 (0 is
// the starport, 7 the hyphen). UnmarshalText reads the characteristics from here.
var characteristicPositions = [7]int{1, 2, 3, 4, 5, 6, 8}

// Valid reports whether every field of p is within its encodable domain: the
// Starport a port letter, each characteristic a 0..ehex.Max eHex value. String and
// MarshalText render a Valid profile without a lossy '?'. Generated profiles are
// always valid; an invalid one is a programming error, not persistable data.
func (p Profile) Valid() bool {
	if !isPortLetter(p.Starport) {
		return false
	}

	for _, v := range p.characteristics() {
		if v < 0 || v > ehex.Max {
			return false
		}
	}

	return true
}

// MarshalText encodes p as its canonical StSAHPGL-T text (encoding.TextMarshaler).
// It is non-lossy: an out-of-domain profile is an error rather than a '?'-bearing
// string that UnmarshalText could not read back. Parse (or UnmarshalText) is its
// exact inverse over the valid domain.
func (p Profile) MarshalText() ([]byte, error) {
	if !p.Valid() {
		return nil, fmt.Errorf("uwp: %+v has an out-of-domain field and cannot be encoded losslessly", p)
	}

	return []byte(p.String()), nil
}

// UnmarshalText parses canonical StSAHPGL-T text into p (encoding.TextUnmarshaler).
// It is strict — the text must be exactly the nine-character "Xdddddd-d" form that
// String emits, canonical and upper-case — so a malformed or non-canonical record
// fails loudly rather than yielding a plausible-but-wrong profile. In particular a
// lower-case eHex digit is rejected even though ehex.ParseDigit would tolerate it,
// so the codec is an exact bijection: Marshal(Unmarshal(text)) == text. p is left
// unchanged on error.
func (p *Profile) UnmarshalText(text []byte) error {
	s := string(text)
	if len(s) != 9 || s[7] != '-' {
		return fmt.Errorf("uwp: %q is not a StSAHPGL-T profile (want nine characters, Xdddddd-d)", s)
	}

	if !isPortLetter(s[0]) {
		return fmt.Errorf("uwp: %q has starport %q, not one of %q", s, s[0], portLetters)
	}

	var d [7]int

	for i, pos := range characteristicPositions {
		v, err := ehex.ParseDigit(s[pos])
		if err != nil {
			return fmt.Errorf("uwp: %q: %w", s, err)
		}

		d[i] = v
	}

	q := Profile{
		Starport: s[0], Size: d[0], Atmosphere: d[1], Hydrographics: d[2],
		Population: d[3], Government: d[4], Law: d[5], TechLevel: d[6],
	}
	// The parsed profile must render back to exactly the input, which rejects any
	// non-canonical spelling (a lower-case digit ParseDigit accepted) and makes the
	// reader the exact inverse of String.
	if q.String() != s {
		return fmt.Errorf("uwp: %q is not in canonical form (canonical is %q)", s, q.String())
	}

	*p = q

	return nil
}

// Parse decodes a canonical StSAHPGL-T profile string, the inverse of MarshalText
// and a convenience over UnmarshalText for the common by-value call.
func Parse(s string) (Profile, error) {
	var p Profile
	if err := p.UnmarshalText([]byte(s)); err != nil {
		return Profile{}, err
	}

	return p, nil
}

// characteristics returns the seven eHex characteristics in StSAHPGL-T order, so
// Valid and any future codec iterate them the same way String writes them.
func (p Profile) characteristics() [7]int {
	return [7]int{p.Size, p.Atmosphere, p.Hydrographics, p.Population, p.Government, p.Law, p.TechLevel}
}
