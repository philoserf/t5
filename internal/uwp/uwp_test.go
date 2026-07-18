package uwp

import "testing"

func TestString(t *testing.T) {
	// Regina, the canonical worked example from Book 3.
	regina := Profile{
		Starport:      'A',
		Size:          7,
		Atmosphere:    8,
		Hydrographics: 8,
		Population:    8,
		Government:    9,
		Law:           9,
		TechLevel:     12,
	}
	if got := regina.String(); got != "A788899-C" {
		t.Fatalf("String() = %q, want %q", got, "A788899-C")
	}
}

func TestStringOutOfRangeDoesNotPanic(t *testing.T) {
	// A caller may build a Profile with an out-of-range field; String must
	// render "?" rather than panic.
	p := Profile{
		Starport:      'A',
		Size:          40,
		Atmosphere:    8,
		Hydrographics: 8,
		Population:    8,
		Government:    9,
		Law:           9,
		TechLevel:     12,
	}
	if got := p.String(); got != "A?88899-C" {
		t.Fatalf("String() = %q, want A?88899-C", got)
	}

	// The zero value is constructible; its unset Starport must render "?" and
	// not put a NUL byte into the record stream.
	if got := (Profile{}).String(); got != "?000000-0" {
		t.Fatalf("Profile{}.String() = %q, want %q", got, "?000000-0")
	}

	// Any other byte outside the port domain renders "?" too.
	for _, c := range []byte{'a', '~', 'I', 'Z', '0'} {
		if got := (Profile{Starport: c}).String(); got != "?000000-0" {
			t.Errorf("Profile{Starport: %q}.String() = %q, want %q", c, got, "?000000-0")
		}
	}
}

func TestStringPortLetters(t *testing.T) {
	// Starport carries a mainworld starport (A-E, X) or a secondary world's
	// spaceport (F, G, H, Y — worldgen/otherworld.go); all render as themselves.
	for _, c := range []byte{'A', 'B', 'C', 'D', 'E', 'X', 'F', 'G', 'H', 'Y'} {
		if got := (Profile{Starport: c}).String(); got[0] != c {
			t.Errorf("Profile{Starport: %q}.String() = %q, want leading %q", c, got, c)
		}
	}
}

func TestStringEHexFields(t *testing.T) {
	// A high-end world exercising eHex digits across every field.
	p := Profile{
		Starport:      'X',
		Size:          10,
		Atmosphere:    15,
		Hydrographics: 10,
		Population:    12,
		Government:    15,
		Law:           18,
		TechLevel:     10,
	}
	if got := p.String(); got != "XAFACFJ-A" {
		t.Fatalf("String() = %q, want %q", got, "XAFACFJ-A")
	}
}
