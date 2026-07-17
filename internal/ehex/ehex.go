// Package ehex implements Traveller's extended hexadecimal ("eHex") digits.
//
// A single eHex digit encodes a value from 0 to 33 as one character: 0-9 for
// 0-9, then the letters A-Z with I and O omitted (they read too easily as 1
// and 0). Universal profiles such as the UWP are strings of these digits, so
// a Size of 10 prints as "A" and a Tech Level of 12 as "C".
package ehex

import (
	"fmt"
	"strings"
)

// Alphabet is the ordered eHex digit set; a value's index in this string is
// its digit. I and O are omitted, following standard Traveller usage.
const Alphabet = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// Max is the largest value a single eHex digit can represent.
const Max = len(Alphabet) - 1

// Digit returns the eHex character for v, which must be in 0..Max. It panics
// otherwise: generators clamp their outputs to valid ranges, so an out-of-range
// value is a programming error, not runtime input.
func Digit(v int) byte {
	if v < 0 || v > Max {
		panic(fmt.Sprintf("ehex: value %d out of range 0..%d", v, Max))
	}

	return Alphabet[v]
}

// Format returns the eHex digit for v as a string, or "?" when v is outside
// 0..Max. Unlike Digit it never panics, so it is safe on display paths — String
// methods, logging — where the value may come from outside the generators.
func Format(v int) string {
	if v < 0 || v > Max {
		return "?"
	}

	return string(Alphabet[v])
}

// ParseDigit returns the value of a single eHex character, case-insensitively.
// It reports an error for any character not in the alphabet.
func ParseDigit(c byte) (int, error) {
	if c >= 'a' && c <= 'z' {
		c -= 'a' - 'A'
	}

	if v := strings.IndexByte(Alphabet, c); v >= 0 {
		return v, nil
	}

	return 0, fmt.Errorf("ehex: %q is not an eHex digit", c)
}
