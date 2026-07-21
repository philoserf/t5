package uwp

import (
	"math/rand/v2"
	"testing"

	"github.com/philoserf/t5/internal/ehex"
)

// TestProfileRoundTrip is the #327 property: every profile in the valid domain
// survives Parse(p.String()) unchanged, and MarshalText produces the same bytes.
// It sweeps every port letter against every characteristic value 0..Max at every
// one of the seven positions (others held at a mid baseline), so each field is
// exercised across its whole range at its own slot — a column-swap or off-by-one
// in the reader shows up as a specific field diverging.
func TestProfileRoundTrip(t *testing.T) {
	for _, sp := range []byte(portLetters) {
		for field := range 7 {
			for v := 0; v <= ehex.Max; v++ {
				// A mid-range baseline in every field, then the field under test set to v.
				p := Profile{Starport: sp}

				for i, f := range fieldPointers(&p) {
					*f = 7
					if i == field {
						*f = v
					}
				}

				text, err := p.MarshalText()
				if err != nil {
					t.Fatalf("MarshalText(%+v) errored on a valid profile: %v", p, err)
				}

				if string(text) != p.String() {
					t.Fatalf("MarshalText = %q, want String %q for a valid profile", text, p.String())
				}

				got, err := Parse(p.String())
				if err != nil {
					t.Fatalf("Parse(%q) errored: %v", p.String(), err)
				}

				if got != p {
					t.Fatalf("round-trip: Parse(%q) = %+v, want %+v (starport %c, field %d = %d)",
						p.String(), got, p, sp, field, v)
				}
			}
		}
	}
}

// TestProfileRoundTripRandomCombos covers cross-field combinations the one-field
// sweep does not: full random valid profiles. Seeded, so it is reproducible.
func TestProfileRoundTripRandomCombos(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	ports := []byte(portLetters)

	for range 5000 {
		p := Profile{
			Starport:      ports[rng.IntN(len(ports))],
			Size:          rng.IntN(ehex.Max + 1),
			Atmosphere:    rng.IntN(ehex.Max + 1),
			Hydrographics: rng.IntN(ehex.Max + 1),
			Population:    rng.IntN(ehex.Max + 1),
			Government:    rng.IntN(ehex.Max + 1),
			Law:           rng.IntN(ehex.Max + 1),
			TechLevel:     rng.IntN(ehex.Max + 1),
		}

		got, err := Parse(p.String())
		if err != nil || got != p {
			t.Fatalf("random round-trip: Parse(%q) = %+v,%v, want %+v", p.String(), got, err, p)
		}
	}
}

// TestMarshalTextRejectsOutOfDomain: MarshalText is non-lossy, so it errors rather
// than emit a '?' (which String tolerates and UnmarshalText could not read back).
func TestMarshalTextRejectsOutOfDomain(t *testing.T) {
	for _, p := range []Profile{
		{Starport: 'Q', Size: 7},            // starport not a port letter
		{Starport: 'A', Size: ehex.Max + 1}, // Size above the eHex range
		{Starport: 'A', TechLevel: -1},      // negative characteristic
		{Starport: 0},                       // the zero value
	} {
		if _, err := p.MarshalText(); err == nil {
			t.Errorf("MarshalText(%+v) succeeded, want an out-of-domain error", p)
		}

		if p.Valid() {
			t.Errorf("Valid(%+v) = true, want false", p)
		}
	}
}

// TestParseRejectsMalformed: the reader is strict, so a record that is the wrong
// length, misplaces the hyphen, or carries a bad digit or starport fails loudly.
func TestParseRejectsMalformed(t *testing.T) {
	for _, s := range []string{
		"",            // empty
		"A788899-C ",  // trailing space (wrong length)
		"A788899C",    // too short — no room for the hyphen
		"A788899+C",   // right length, wrong separator
		"a788899-C",   // lowercase starport — canonical is upper
		"A78a899-C",   // lowercase eHex digit — canonical is upper (ParseDigit tolerates it)
		"A788899-c",   // lowercase Tech Level digit
		"Q788899-C",   // starport not a port letter
		"A78I899-C",   // I is not an eHex digit
		"A788899-C-9", // too long
	} {
		if _, err := Parse(s); err == nil {
			t.Errorf("Parse(%q) succeeded, want a malformed-input error", s)
		}
	}
}

// TestReginaRoundTrip pins the canonical example both ways.
func TestReginaRoundTrip(t *testing.T) {
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
	if regina.String() != "A788899-C" {
		t.Fatalf("Regina String = %q, want A788899-C", regina.String())
	}

	if got, err := Parse("A788899-C"); err != nil || got != regina {
		t.Fatalf("Parse(A788899-C) = %+v,%v, want %+v", got, err, regina)
	}
}

// fieldPointers returns pointers to the seven characteristics in StSAHPGL-T order,
// so a sweep can set each field in turn.
func fieldPointers(p *Profile) [7]*int {
	return [7]*int{
		&p.Size, &p.Atmosphere, &p.Hydrographics,
		&p.Population, &p.Government, &p.Law, &p.TechLevel,
	}
}
