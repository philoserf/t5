package survey

import (
	"strings"
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/route"
	"github.com/philoserf/t5/internal/sectorgen"
)

func TestSECRoundTripWithRelationshipMetadata(t *testing.T) {
	const first = "0101 Maesavo E410100-7 Lo Co {-3}(500-2)[1139] B - - 233 17 Im K8 V BD K4 VI"

	const second = "0201 Regina A788899-C Ph Pa Ri Cy {+4}(D7E+4)[9C6D] BcCeF NS - 503 6 Im F7 V"

	a, err := ParseRecord(first)
	if err != nil {
		t.Fatal(err)
	}

	b, err := ParseRecord(second)
	if err != nil {
		t.Fatal(err)
	}

	doc := SEC{
		Records: []RecordLine{a, b},
		Routes: []route.Link{{
			From: sectorgen.Hex{Col: 1, Row: 1}, To: sectorgen.Hex{Col: 2, Row: 1}, Jump: 1,
		}},
		Ownerships: []Ownership{{
			Colony: sectorgen.Hex{Col: 2, Row: 1}, Owner: sectorgen.Hex{Col: 1, Row: 1},
		}},
	}

	want := first + "\n" + second + "\n\n" +
		"# Route: 0101 0201 J1\n" +
		"# Owner: 0201 O:0101"
	if got := doc.String(); got != want {
		t.Fatalf("SEC.String() =\n%s\nwant\n%s", got, want)
	}

	parsed, err := ParseSEC(want)
	if err != nil {
		t.Fatalf("ParseSEC: %v", err)
	}

	if got := parsed.String(); got != want {
		t.Fatalf("SEC round trip =\n%s\nwant\n%s", got, want)
	}
}

func TestSurveySECPreservesSecondSurveyLines(t *testing.T) {
	sv := Sector(dice.NewWithSeed(42), sectorgen.Sparse)
	sec := sv.SEC()

	wantPrefix := sv.Records[0].SecondSurvey() + "\n"
	if !strings.HasPrefix(sec, wantPrefix) {
		t.Fatalf("SEC changed first Second Survey line\n got %q\nwant %q", sec[:len(wantPrefix)], wantPrefix)
	}

	parsed, err := ParseSEC(sec)
	if err != nil {
		t.Fatalf("ParseSEC(generated): %v", err)
	}

	if parsed.String() != sec {
		t.Fatal("generated SEC did not round-trip byte-identically")
	}
}

func TestParseSECRejectsMalformedMetadata(t *testing.T) {
	for _, line := range []string{
		"# Route: 0101 0201 J2", // jump disagrees with endpoints
		"# Route: 0101 ZZZZ J1",
		"# Owner: 0101 O:0101", // cannot own itself
		"# Owner: 0101 O:3201", // farther than six hexes
	} {
		if _, err := ParseSEC(line); err == nil {
			t.Errorf("ParseSEC(%q) succeeded", line)
		}
	}
}

// TestSectorSaveLoadByteIdentical is the #327 acceptance: a generated sector,
// saved as its Second Survey record lines, survives a load/save cycle
// byte-for-byte. Every one of a full sector's ~640 records is rendered, parsed
// back through ParseRecord, and re-rendered; the whole blob must be unchanged.
// This exercises every field combination the generator actually produces —
// capital codes, no-trade-code worlds, way stations, multi-star systems — far
// past what a hand-built fixture reaches.
func TestSectorSaveLoadByteIdentical(t *testing.T) {
	for _, d := range []sectorgen.Density{sectorgen.Sparse, sectorgen.Standard, sectorgen.Dense} {
		sv := Sector(dice.NewWithSeed(42), d)
		if len(sv.Records) == 0 {
			t.Fatalf("density %v produced no records", d)
		}

		var save strings.Builder
		for _, rec := range sv.Records {
			save.WriteString(rec.SecondSurvey())
			save.WriteByte('\n')
		}

		saved := save.String()

		var resave strings.Builder

		for i, line := range strings.Split(strings.TrimRight(saved, "\n"), "\n") {
			rec, err := ParseRecord(line)
			if err != nil {
				t.Fatalf("density %v record %d: ParseRecord(%q): %v", d, i, line, err)
			}

			resave.WriteString(rec.String())
			resave.WriteByte('\n')
		}

		if resave.String() != saved {
			// Find the first differing line for a readable failure.
			a := strings.Split(saved, "\n")
			b := strings.Split(resave.String(), "\n")

			for i := range a {
				if i >= len(b) || a[i] != b[i] {
					t.Fatalf("density %v: save/load diverged at record %d\n got %q\nwant %q", d, i, b[i], a[i])
				}
			}
		}
	}
}

// TestParseRecordRejectsMalformed: the record reader is strict.
func TestParseRecordRejectsMalformed(t *testing.T) {
	good := "0101 Maesavo E410100-7 Lo Co {-3}(500-2)[1139] B - - 233 17 Im K8 V BD K4 VI"
	if _, err := ParseRecord(good); err != nil {
		t.Fatalf("ParseRecord(good) errored: %v", err)
	}

	for _, s := range []string{
		"", // empty
		"ZZZZ Maesavo E410100-7 Lo Co {-3}(500-2)[1139] B - - 233 17 Im K8 V",  // bad hex
		"0101 Maesavo E410100-7 Lo Co {-3}(500-2)[1139] B - - 2X 17 Im K8 V",   // bad PBG (2 digits)
		"0101 Maesavo E410100-7 Lo Co {-3}(500-2)[1139] B - - 233 X Im K8 V",   // bad world count
		"0101 Maesavo E410100-7 Lo Co {-3}(500-2)[1139] B - - 233 17 Im",       // no stellar
		"0101 Maesavo E410100-7 Lo Co {-3}(500-2)[1139] B - - 233 17 Im K88 V", // malformed star (3-char type token)
	} {
		if _, err := ParseRecord(s); err == nil {
			t.Errorf("ParseRecord(%q) succeeded, want a malformed-record error", s)
		}
	}
}
