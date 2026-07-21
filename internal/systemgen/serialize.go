package systemgen

import (
	"fmt"
	"strings"

	"github.com/philoserf/t5/internal/ehex"
)

// Serialization for the stellar tail of a Second Survey record (#327). Stellar
// renders each of the up-to-eight star slots through Star.String,
// which has three forms — "BD" (a brown dwarf, one token), "K D" (a white dwarf,
// type then D), and "K8 V" (type+decimal then luminosity). ParseStellar reverses
// the flat list; the slot a star came from (Primary, a companion, …) is not in the
// string and is not recovered, but the ordered list re-renders identically, which
// is all a byte-identical record round-trip needs.

// ParseStellar decodes the space-separated star list of a record's stellar field
// into the flat, ordered stars Stellar produced. It is strict: each star must be
// one of the three canonical Star.String forms, and re-rendering the result must
// reproduce the input exactly, so the reader is the writer's inverse.
func ParseStellar(s string) ([]Star, error) {
	tokens := strings.Fields(s)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("systemgen: %q has no stars", s)
	}

	var stars []Star

	for i := 0; i < len(tokens); {
		star, used, err := parseStar(tokens[i:])
		if err != nil {
			return nil, fmt.Errorf("systemgen: %q: %w", s, err)
		}

		stars = append(stars, star)
		i += used
	}

	// The reader is the exact inverse of Stellar: re-rendering through the real
	// writer must reproduce the input, which rejects any non-canonical spelling a
	// loose parse might accept.
	if re := StellarString(stars); re != s {
		return nil, fmt.Errorf("systemgen: %q is not canonical stellar data (want %q)", s, re)
	}

	return stars, nil
}

// parseStar reads one star off the front of tokens, returning it and how many
// tokens it consumed (one for a brown dwarf, two otherwise).
func parseStar(tokens []string) (Star, int, error) {
	switch {
	case tokens[0] == "BD":
		return Star{Type: "BD"}, 1, nil

	case len(tokens) >= 2 && tokens[1] == "D":
		// A white dwarf: "Type D". Star.String drops the decimal here.
		return Star{Type: tokens[0], Size: "D"}, 2, nil

	case len(tokens) >= 2:
		// "TypeDecimal Size", e.g. "K8" + "V".
		t := tokens[0]
		if len(t) != 2 {
			return Star{}, 0, fmt.Errorf("%q is not a Type+Decimal", t)
		}

		dec, err := ehex.ParseDigit(t[1])
		if err != nil {
			return Star{}, 0, fmt.Errorf("star decimal %q: %w", t[1:], err)
		}

		return Star{Type: t[:1], Decimal: dec, Size: tokens[1]}, 2, nil

	default:
		return Star{}, 0, fmt.Errorf("%q is an incomplete star", tokens[0])
	}
}
