package senses

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
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
	want := []string{"V-16-RGB", "H-16-9392", "S-10-2", "T-06-2", "A-00-0", "P-00-00"}

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
		"V-16-RGB", "V-20-VHD", "H-16-9392", "H-20-A4B2",
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
	if !ok || hearing != (HearingDetail{Frequency: 9, Span: 3, Voice: 9, VoiceRange: 2}) {
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
