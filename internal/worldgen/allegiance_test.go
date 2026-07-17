package worldgen

import "testing"

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
	// A code of any other length is malformed: rejected, but still yields the
	// default so the caller renders a valid record.
	for _, bad := range []string{"X", "Imp", "Imperial"} {
		if a, ok := ParseAllegiance(bad); ok || a != Imperial {
			t.Errorf("ParseAllegiance(%q) = %q,%v, want Im,false", bad, a, ok)
		}
	}

	if !Imperial.Valid() || !Allegiance("ImDd").Valid() || Allegiance("Imp").Valid() {
		t.Errorf("Valid accepts a 2- or 4-letter code only")
	}
}
