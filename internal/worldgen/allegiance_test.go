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
	// A Chart F code and a setting-specific one are both well-formed.
	for _, code := range []string{"Im", "Zh", "So", "Dr"} {
		if a, ok := ParseAllegiance(code); !ok || string(a) != code {
			t.Errorf("ParseAllegiance(%q) = %q,%v, want %q,true", code, a, ok, code)
		}
	}
	// A malformed code is rejected but still yields the default.
	if a, ok := ParseAllegiance("Imperial"); ok || a != Imperial {
		t.Errorf("a long code should be rejected with the default, got %q,%v", a, ok)
	}
	if !Imperial.Valid() || Allegiance("X").Valid() {
		t.Errorf("Valid checks for a two-character code")
	}
}
