package worldgen

import (
	"strings"
	"testing"
)

// TestParseAllegiance: allegiance is referee-supplied (Book 3 p.28). An empty code
// defaults to Imperial; a two-letter code (whether or not the core rules list it,
// since "Other abbreviations are possible") is well-formed; a longer string is not,
// but still yields the default to render.
func TestParseAllegiance(t *testing.T) {
	if a, ok := ParseAllegiance(""); !ok || a != Imperial {
		t.Errorf("empty allegiance = %q,%v, want Im,true", a, ok)
	}
	// Two-letter Chart F codes, and four-letter Second Survey codes, are all
	// well-formed and preserved (ImDd is the Domain of Deneb — dropping it to "Im"
	// would emit a different polity).
	for _, code := range []string{"Im", "Zh", "Dr", "ImDd", "NaHu"} {
		if a, ok := ParseAllegiance(code); !ok || string(a) != code {
			t.Errorf("ParseAllegiance(%q) = %q,%v, want %q,true", code, a, ok, code)
		}
	}
	// A malformed code is rejected, but still yields the default so the caller
	// renders a valid record. Wrong length, and — the shape the length check alone
	// let through — right length but not letters.
	bad := []string{
		"X", "Imp", "Imperial", // wrong length
		"Im d", "i ", "\tI", // whitespace: would add a field token to the record
		"42", "!!", "Im-", "I.", // digits and punctuation
		"\x00\x00", "Im\x00", // control bytes, the #240 NUL-in-the-stream class
	}
	for _, code := range bad {
		if a, ok := ParseAllegiance(code); ok || a != Imperial {
			t.Errorf("ParseAllegiance(%q) = %q,%v, want Im,false", code, a, ok)
		}
	}

	if !Imperial.Valid() || !Allegiance("ImDd").Valid() || Allegiance("Imp").Valid() {
		t.Errorf("Valid accepts a 2- or 4-letter code only")
	}
}

// TestAllegianceString: Allegiance is an exported string type, so a caller can
// convert into it directly and skip ParseAllegiance. String is the display half of
// the pair (as uwp's formatStarport is for a caller-built Profile): it renders a
// valid code unchanged and substitutes a visible sentinel for anything else, so a
// malformed value cannot reach the record stream by that route either.
func TestAllegianceString(t *testing.T) {
	for _, code := range []string{"Im", "Zh", "ImDd", "NaHu"} {
		if got := Allegiance(code).String(); got != code {
			t.Errorf("Allegiance(%q).String() = %q, want %q", code, got, code)
		}
	}

	for _, code := range []string{"", "X", "Imp", "Im d", "42", "!!", "\x00\x00"} {
		if got := Allegiance(code).String(); got != InvalidAllegiance {
			t.Errorf("Allegiance(%q).String() = %q, want %q", code, got, InvalidAllegiance)
		}
	}
}

// TestAllegianceRecordFieldIntegrity: the Second Survey record is positional and
// whitespace-delimited, so the allegiance field must always be exactly one token of
// printable, non-space characters. This is the property the length-only check broke:
// "Im d" is four characters, passed Valid, and split into two fields — shifting the
// stellar column left for any consumer parsing by position.
func TestAllegianceRecordFieldIntegrity(t *testing.T) {
	codes := []string{
		"", "Im", "ImDd", "NaHu", // well-formed
		"Im d", "i ", "42", "!!", "\x00\x00", "Imperial", "\tI", // malformed
	}

	for _, code := range codes {
		a, _ := ParseAllegiance(code)
		field := a.String()

		if n := len(strings.Fields(field)); n != 1 {
			t.Errorf("ParseAllegiance(%q) rendered %q: %d whitespace fields, want exactly 1",
				code, field, n)
		}

		for i := range len(field) {
			if c := field[i]; c <= ' ' || c > '~' {
				t.Errorf("ParseAllegiance(%q) rendered %q: byte %d = %q is not printable ASCII",
					code, field, i, c)
			}
		}
	}
}
