// Package senses implements Traveller5's six senses (Book 1 pp.186-199),
// including their fixed-format sense identifiers and sense Actions.
package senses

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/ehex"
	"github.com/philoserf/t5/internal/rangeband"
	"github.com/philoserf/t5/internal/task"
)

const visionSpectrum = "VHDUSPBGRCANIFXZ"

// A Sense is a Traveller sense identifier. Detail has a sense-specific fixed
// format: three color-band letters for Vision; four eHex digits for Hearing's
// frequency, span, voice, and voice range; one eHex digit for Smell sharpness,
// Touch sensitivity, or Awareness acuity; and two eHex digits for Perception
// tone and Poice. Constant zero means the sense is absent.
type Sense struct {
	ID       byte
	Constant int
	Detail   string
}

// HearingDetail is the decoded Freq, Span, Voice, and Voice Range portion of a
// Hearing identifier.
type HearingDetail struct {
	Frequency  int
	Span       int
	Voice      int
	VoiceRange int
}

// PerceptionDetail is the decoded Tone and Poice portion of a Perception ID.
type PerceptionDetail struct {
	Tone  int
	Poice int
}

// Human sense IDs (Book 1 pp.188, 190-198). Humans lack Awareness and
// Perception, represented by Constant zero while retaining canonical details.
var (
	Vision     = Sense{'V', 16, "RGB"}
	Hearing    = Sense{'H', 16, "9382"}
	Smell      = Sense{'S', 10, "2"}
	Touch      = Sense{'T', 6, "2"}
	Awareness  = Sense{'A', 0, "0"}
	Perception = Sense{'P', 0, "00"}
)

// Available reports whether the sophont has this sense.
func (s Sense) Available() bool { return s.Constant > 0 }

// Valid reports whether the ID, decimal Constant, and per-sense Detail can be
// represented by the canonical fixed-format codec.
func (s Sense) Valid() bool {
	if s.Constant < 0 || s.Constant > 99 {
		return false
	}

	switch s.ID {
	case 'V':
		return validVisionBands(s.Detail)
	case 'H':
		return validEHex(s.Detail, 4)
	case 'S', 'T', 'A':
		return validEHex(s.Detail, 1)
	case 'P':
		return validEHex(s.Detail, 2)
	default:
		return false
	}
}

func validVisionBands(detail string) bool {
	if len(detail) != 3 {
		return false
	}

	if strings.Contains(visionSpectrum, detail) {
		return true
	}

	reversed := string([]byte{detail[2], detail[1], detail[0]})

	return strings.Contains(visionSpectrum, reversed)
}

// String renders the canonical fixed-format sense ID. Invalid caller-built
// values render "?"; MarshalText provides the checked persistence path.
func (s Sense) String() string {
	if !s.Valid() {
		return "?"
	}

	return fmt.Sprintf("%c-%02d-%s", s.ID, s.Constant, s.Detail)
}

// MarshalText implements encoding.TextMarshaler.
func (s Sense) MarshalText() ([]byte, error) {
	if !s.Valid() {
		return nil, fmt.Errorf("senses: %+v is not a valid sense ID", s)
	}

	return []byte(s.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler. It accepts only canonical,
// upper-case fixed-format IDs and leaves the receiver unchanged on error.
func (s *Sense) UnmarshalText(text []byte) error {
	q, err := Parse(string(text))
	if err != nil {
		return err
	}

	*s = q

	return nil
}

// Parse decodes a canonical fixed-format sense ID.
func Parse(text string) (Sense, error) {
	if len(text) < 6 || text[1] != '-' || text[4] != '-' {
		return Sense{}, fmt.Errorf("senses: %q is not in ID-00-detail form", text)
	}

	constant, err := strconv.Atoi(text[2:4])
	if err != nil {
		return Sense{}, fmt.Errorf("senses: %q has a non-decimal Constant", text)
	}

	s := Sense{ID: text[0], Constant: constant, Detail: text[5:]}
	if !s.Valid() {
		return Sense{}, fmt.Errorf("senses: %q has invalid detail for sense %q", text, s.ID)
	}

	if s.String() != text {
		return Sense{}, fmt.Errorf("senses: %q is not canonical", text)
	}

	return s, nil
}

// VisionBands returns the three adjacent wavelength codes for Vision.
func (s Sense) VisionBands() ([3]byte, bool) {
	if s.ID != 'V' || !s.Valid() {
		return [3]byte{}, false
	}

	return [3]byte{s.Detail[0], s.Detail[1], s.Detail[2]}, true
}

// HearingParameters decodes Hearing's frequency, span, voice, and voice range.
func (s Sense) HearingParameters() (HearingDetail, bool) {
	if s.ID != 'H' || !s.Valid() {
		return HearingDetail{}, false
	}

	d, _ := ehex.ParseDigits(s.Detail)

	return HearingDetail{d[0], d[1], d[2], d[3]}, true
}

// Level decodes the single detail used by Smell, Touch, or Awareness.
func (s Sense) Level() (int, bool) {
	if (s.ID != 'S' && s.ID != 'T' && s.ID != 'A') || !s.Valid() {
		return 0, false
	}

	v, _ := ehex.ParseDigit(s.Detail[0])

	return v, true
}

// PerceptionParameters decodes Perception's tone and Poice values.
func (s Sense) PerceptionParameters() (PerceptionDetail, bool) {
	if s.ID != 'P' || !s.Valid() {
		return PerceptionDetail{}, false
	}

	d, _ := ehex.ParseDigits(s.Detail)

	return PerceptionDetail{d[0], d[1]}, true
}

func validEHex(detail string, width int) bool {
	if len(detail) != width {
		return false
	}

	for i := range detail {
		if _, err := ehex.ParseDigit(detail[i]); err != nil || detail[i] >= 'a' && detail[i] <= 'z' {
			return false
		}
	}

	return true
}

// NoticeAtRange resolves an at-range sense Action (Vision/Hearing/Awareness/
// Perception; Book 1 p.190). It rolls R=Range dice — Range 0 and the R/T
// sub-bands count as 1 — and succeeds on a roll at or under Constant plus the
// Benchmark (objectSize - Range) plus any situational mods (Master Mods;
// higher is better). A sense the sophont lacks automatically fails.
func NoticeAtRange(r *dice.Roller, s Sense, objectSize, rng int, mods ...int) dice.CheckResult {
	if !s.Available() {
		return dice.CheckResult{}
	}

	return task.ResolveDice(r, rng, s.Constant+(objectSize-rng), mods...)
}

// NoticeInContact resolves an in-contact sense Action (Touch, and Smell by
// scent intensity; Book 1 p.187): 2D at or under Constant + benchmark + mods.
// A sense the sophont lacks automatically fails.
func NoticeInContact(r *dice.Roller, s Sense, benchmark int, mods ...int) dice.CheckResult {
	if !s.Available() {
		return dice.CheckResult{}
	}

	return task.ResolveDice(r, task.Average.Dice(), s.Constant+benchmark, mods...)
}

// RangeBand returns the sense Range band (0-9) for a distance in meters, using
// the shared world-range ladder (Book 1 pp.24, 188).
func RangeBand(meters float64) int {
	n, ok := rangeband.WorldForDistance(meters).Number()
	if !ok {
		return 1 // the R and T reading/talking sub-bands both count as Range 1
	}

	return n
}
