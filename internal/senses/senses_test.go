package senses

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/ehex"
)

func TestNoticeAtRangeReginaExample(t *testing.T) {
	// Book 1 p.190 worked example: Human V-16 spots a Size-6 cargo mover at
	// Range 6 moving Very Fast (+2). Target = 16 + (6-6) + 2 = 18, rolled on 6D.
	res := NoticeAtRange(dice.NewScripted(2, 2, 2, 2, 2, 2), Vision, 6, 6, 2) // 6D = 12
	if res.Target != 18 {
		t.Errorf("target = %d, want 18", res.Target)
	}

	if res.Roll != 12 || !res.Success {
		t.Errorf("roll 12 vs 18 should succeed: %+v", res)
	}
	// A roll of 24 (6x4) exceeds 18 and fails.
	if res := NoticeAtRange(dice.NewScripted(4, 4, 4, 4, 4, 4), Vision, 6, 6, 2); res.Success {
		t.Errorf("roll 24 vs 18 should fail: %+v", res)
	}

	// Second example: Size-6 mover at Range 5, motionless. Target = 16 + 1 = 17
	// on 5D.
	res2 := NoticeAtRange(dice.NewScripted(3, 3, 3, 3, 3), Vision, 6, 5) // 5D = 15
	if res2.Target != 17 || res2.Roll != 15 || !res2.Success {
		t.Errorf("Range-5 example: target/roll = %d/%d, want 17/15 success", res2.Target, res2.Roll)
	}
}

func TestRangeDiceFloor(t *testing.T) {
	// Range 0 (contact) and the R/T sub-bands roll a single die, not zero.
	res := NoticeAtRange(dice.NewScripted(6), Vision, 3, 0)
	if len(res.Faces) != 1 {
		t.Errorf("Range 0 rolled %d dice, want 1", len(res.Faces))
	}
}

func TestNoticeInContact(t *testing.T) {
	// Touch (Constant 6): 2D under 6 + benchmark.
	res := NoticeInContact(dice.NewScripted(2, 2), Touch, 3) // 2D=4 vs 6+3=9
	if res.Target != 9 || len(res.Faces) != 2 || !res.Success {
		t.Errorf("touch = %+v, want target 9, 2 dice, success", res)
	}
}

func TestRangeBand(t *testing.T) {
	cases := map[float64]int{5: 1, 50: 2, 200: 3, 1000: 5, 50_000: 7, 0: 0}
	for meters, want := range cases {
		if got := RangeBand(meters); got != want {
			t.Errorf("RangeBand(%gm) = %d, want %d", meters, got, want)
		}
	}
}

func TestAbsentSenseAlwaysFails(t *testing.T) {
	// Humans lack Awareness/Perception (Constant 0); their Actions can't succeed
	// even at short range against a large object.
	if res := NoticeAtRange(dice.NewScripted(1), Awareness, 5, 1); res.Success {
		t.Errorf("human Awareness should not succeed: %+v", res)
	}

	if res := NoticeInContact(dice.NewScripted(1, 1), Perception, 5); res.Success {
		t.Errorf("human Perception should not succeed: %+v", res)
	}
}

func TestSenseString(t *testing.T) {
	if Vision.String() != "V-16-RGB" || Touch.String() != "T-06-2" {
		t.Errorf("Sense.String = %q/%q, want V-16-RGB/T-06-2", Vision.String(), Touch.String())
	}
}

func TestHumanSenseIDs(t *testing.T) {
	want := []string{"V-16-RGB", "H-16-9382", "S-10-2", "T-06-2", "A-00-0", "P-00-00"}

	got := []Sense{Vision, Hearing, Smell, Touch, Awareness, Perception}
	for i, s := range got {
		if s.String() != want[i] {
			t.Errorf("sense %d = %q, want %q", i, s.String(), want[i])
		}
	}

	if !Vision.Available() || Awareness.Available() || Perception.Available() {
		t.Error("human sense availability does not match Constants")
	}
}

func TestSenseCodecRoundTrip(t *testing.T) {
	texts := []string{
		"V-16-RGB", "V-20-VHD", "H-16-9382", "H-20-A4B2",
		"S-10-2", "T-06-2", "A-16-1", "P-24-24",
	}
	for _, text := range texts {
		s, err := Parse(text)
		if err != nil {
			t.Fatalf("Parse(%q): %v", text, err)
		}

		encoded, err := s.MarshalText()
		if err != nil || string(encoded) != text {
			t.Errorf("MarshalText(Parse(%q)) = %q, %v", text, encoded, err)
		}

		var decoded Sense
		if err := decoded.UnmarshalText(encoded); err != nil || decoded != s {
			t.Errorf("UnmarshalText(%q) = %+v, %v; want %+v", text, decoded, err, s)
		}
	}
}

// assertRoundTrip checks Parse(s.String()) == s and that MarshalText /
// UnmarshalText reproduce s exactly.
func assertRoundTrip(t *testing.T, s Sense) {
	t.Helper()

	text := s.String()

	parsed, err := Parse(text)
	if err != nil {
		t.Fatalf("Parse(%q): %v", text, err)
	}

	if parsed != s {
		t.Fatalf("Parse(%q) = %+v, want %+v", text, parsed, s)
	}

	encoded, err := s.MarshalText()
	if err != nil || string(encoded) != text {
		t.Fatalf("MarshalText(%+v) = %q, %v; want %q", s, encoded, err, text)
	}

	var decoded Sense
	if err := decoded.UnmarshalText(encoded); err != nil || decoded != s {
		t.Fatalf("UnmarshalText(%q) = %+v, %v; want %+v", encoded, decoded, err, s)
	}
}

func TestSenseCodecRoundTripExhaustiveVision(t *testing.T) {
	// Every valid Vision detail is an adjacent 3-letter window of the 16-letter
	// spectrum, read in either direction: 14 forward + 14 reversed.
	var details []string

	for i := 0; i+3 <= len(visionSpectrum); i++ {
		window := visionSpectrum[i : i+3]
		details = append(details, window, string([]byte{window[2], window[1], window[0]}))
	}

	if len(details) != 28 {
		t.Fatalf("enumerated %d Vision details, want 28", len(details))
	}

	for _, detail := range details {
		for _, constant := range []int{0, 1, 9, 16, 99} {
			assertRoundTrip(t, Sense{ID: 'V', Constant: constant, Detail: detail})
		}
	}
}

func TestSenseCodecRoundTripExhaustiveEHex(t *testing.T) {
	// For each eHex-detail sense, sweep every valid eHex digit through every
	// detail position (the other positions held at '0').
	widths := map[byte]int{'H': 4, 'S': 1, 'T': 1, 'A': 1, 'P': 2}
	for id, width := range widths {
		for pos := range width {
			for i := range len(ehex.Alphabet) {
				detail := []byte("0000"[:width])
				detail[pos] = ehex.Alphabet[i]

				for _, constant := range []int{0, 1, 9, 16, 99} {
					assertRoundTrip(t, Sense{ID: id, Constant: constant, Detail: string(detail)})
				}
			}
		}
	}
}

func TestInvalidSenseFailsClosed(t *testing.T) {
	// String must render "?" and MarshalText must error for every invalid
	// Sense — the checked persistence path must not fail open.
	invalid := []Sense{
		{ID: 'V', Constant: 16, Detail: "RBG"},  // non-adjacent bands
		{ID: 'V', Constant: 16, Detail: "rgb"},  // lowercase bands
		{ID: 'V', Constant: 100, Detail: "RGB"}, // Constant above 99
		{ID: 'V', Constant: -1, Detail: "RGB"},  // negative Constant
		{ID: 'H', Constant: 16, Detail: "939"},  // wrong detail width
		{ID: 'S', Constant: 10, Detail: "I"},    // I is not an eHex digit
		{ID: 'P', Constant: 24, Detail: "2o"},   // lowercase eHex
		{ID: 'X', Constant: 16, Detail: "0"},    // unknown sense ID
		{},                                      // zero value
	}
	for _, s := range invalid {
		if s.Valid() {
			t.Errorf("Valid(%+v) = true, want false", s)
		}

		if got := s.String(); got != "?" {
			t.Errorf("String(%+v) = %q, want %q", s, got, "?")
		}

		if encoded, err := s.MarshalText(); err == nil {
			t.Errorf("MarshalText(%+v) = %q, nil; want error", s, encoded)
		}
	}
}

func TestSenseCodecRejectsMalformed(t *testing.T) {
	bad := []string{
		"V-16", "V-016-RGB", "V-16-rgb", "V-16-RBG", // bands must be canonical and adjacent
		"H-16-939", "H-16-93I2", "S-10-22", "T-6-2", "A-00-I", "P-24-2o",
		"X-16-0", "V-AA-RGB", "V-100-RGB",
	}
	for _, text := range bad {
		if _, err := Parse(text); err == nil {
			t.Errorf("Parse(%q) succeeded, want error", text)
		}
	}

	original := Vision
	if err := original.UnmarshalText([]byte("V-16-rgb")); err == nil || original != Vision {
		t.Errorf("failed UnmarshalText changed receiver: %+v, %v", original, err)
	}
}

func TestSenseDetails(t *testing.T) {
	bands, ok := Vision.VisionBands()
	if !ok || bands != [3]byte{'R', 'G', 'B'} {
		t.Errorf("VisionBands = %q, %v", bands, ok)
	}

	hearing, ok := Hearing.HearingParameters()
	if !ok || hearing != (HearingDetail{Frequency: 9, Span: 3, Voice: 8, VoiceRange: 2}) {
		t.Errorf("HearingParameters = %+v, %v", hearing, ok)
	}

	if sharpness, ok := Smell.Level(); !ok || sharpness != 2 {
		t.Errorf("Smell.Level = %d, %v", sharpness, ok)
	}

	p, err := Parse("P-24-24")
	if err != nil {
		t.Fatal(err)
	}

	if detail, ok := p.PerceptionParameters(); !ok || detail != (PerceptionDetail{Tone: 2, Poice: 4}) {
		t.Errorf("PerceptionParameters = %+v, %v", detail, ok)
	}

	if _, ok := Vision.Level(); ok {
		t.Error("Vision unexpectedly exposed a single-detail level")
	}
}
