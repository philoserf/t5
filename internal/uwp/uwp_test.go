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
