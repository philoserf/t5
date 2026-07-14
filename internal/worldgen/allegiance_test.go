package worldgen

import "testing"

// TestParseAllegiance: allegiance is referee-supplied (Book 3 p.28). An empty code
// defaults to Imperial, the nine Chart F codes are recognized, and a setting may
// use a two-letter code the core rules do not list.
func TestParseAllegiance(t *testing.T) {
	if a, ok := ParseAllegiance(""); !ok || a != Imperial {
		t.Errorf("empty allegiance = %q,%v, want Im,true", a, ok)
	}
	// The Chart F codes and their names.
	cases := map[Allegiance]string{
		Imperial: "Imperial", Zhodani: "Zhodani", Solomani: "Solomani",
		Kkree: "K'kree", Aslan: "Aslan", NonAligned: "Non-Aligned",
	}
	for code, name := range cases {
		a, ok := ParseAllegiance(string(code))
		if !ok || a != code {
			t.Errorf("ParseAllegiance(%q) = %q,%v", code, a, ok)
		}
		if a.Name() != name {
			t.Errorf("%q.Name() = %q, want %q", code, a.Name(), name)
		}
	}
	// A setting-specific code is well-formed and prints as its bare code.
	if a, ok := ParseAllegiance("Dr"); !ok || a.Name() != "Dr" {
		t.Errorf("an unlisted two-letter code should be accepted: %q,%v", a, ok)
	}
	// A malformed code is rejected but still yields the default to render.
	if a, ok := ParseAllegiance("Imperial"); ok {
		t.Errorf("a long code should be rejected, got %q,%v", a, ok)
	}
	if !Imperial.Valid() || Allegiance("X").Valid() {
		t.Errorf("Valid checks for a two-character code")
	}
}
