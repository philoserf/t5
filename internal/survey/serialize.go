package survey

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/philoserf/t5/internal/ehex"
	"github.com/philoserf/t5/internal/sectorgen"
	"github.com/philoserf/t5/internal/systemgen"
	"github.com/philoserf/t5/internal/worldgen"
)

// Serialization for a full Second Survey record (REDESIGN.md §2, #327). A record
// renders as
//
//	hex  name  <world>  PBG  worlds  allegiance  <stellar>
//
// e.g. "0101 Maesavo E410100-7 Lo Co {-3}(500-2)[1139] B - - 233 17 Im K8 V BD K4 VI".
// A RecordLine is the parsed form: the fields a record line carries, and nothing
// else. It is deliberately not a systemgen.System — the line is a lossy projection
// of one (no orbit map; the stellar field flattens the star slots) — so it is its
// own flat serialization type that round-trips its own String. ParseRecord and
// String are inverses over a valid record, and a whole sector saved as its record
// lines survives a load/save cycle byte-identically.

// A RecordLine holds one parsed Second Survey record.
type RecordLine struct {
	Hex        sectorgen.Hex
	Name       string
	Mainworld  worldgen.World
	PopDigit   int // the P of PBG (the mainworld's population multiplier digit)
	Belts      int // the B of PBG
	Giants     int // the G of PBG
	Worlds     int
	Allegiance worldgen.Allegiance
	Stars      []systemgen.Star
}

// String renders the record exactly as systemgen.System.SecondSurvey does, so a
// RecordLine parsed from a record re-renders it byte-for-byte.
func (r RecordLine) String() string {
	stars := make([]string, len(r.Stars))
	for i, s := range r.Stars {
		stars[i] = s.String()
	}

	pbg := systemgen.PBGString(r.PopDigit, r.Belts, r.Giants)

	return fmt.Sprintf("%s %s %s %s %d %s %s",
		r.Hex, r.Name, r.Mainworld.SecondSurvey(), pbg, r.Worlds, r.Allegiance, strings.Join(stars, " "))
}

// ParseRecord decodes a Second Survey record line. It is strict throughout — a
// malformed hex, world, PBG, world count, allegiance, or star fails loudly.
func ParseRecord(line string) (RecordLine, error) {
	fields := strings.Fields(line)

	// hex, name, then the world portion (ending at the zone, three fields past the
	// {Ix}(Ex)[Cx] token), then PBG, worlds, allegiance, and the stellar tail.
	ext := -1

	for i, f := range fields {
		if strings.HasPrefix(f, "{") {
			ext = i

			break
		}
	}

	// hex name UWP … {ext} nob bases zone PBG worlds alleg star…: ext ≥ 4, and at
	// least four fields (PBG, worlds, alleg, one star) follow the zone.
	if ext < 4 || ext+7 > len(fields) {
		return RecordLine{}, fmt.Errorf("survey: %q is not a Second Survey record", line)
	}

	hex, ok := sectorgen.ParseHex(fields[0])
	if !ok {
		return RecordLine{}, fmt.Errorf("survey: %q: bad hex %q", line, fields[0])
	}

	world, err := worldgen.ParseWorld(strings.Join(fields[2:ext+4], " "))
	if err != nil {
		return RecordLine{}, fmt.Errorf("survey: %q: %w", line, err)
	}

	pop, belts, giants, err := parsePBG(fields[ext+4])
	if err != nil {
		return RecordLine{}, fmt.Errorf("survey: %q: %w", line, err)
	}

	worlds, err := strconv.Atoi(fields[ext+5])
	if err != nil {
		return RecordLine{}, fmt.Errorf("survey: %q: bad world count %q: %w", line, fields[ext+5], err)
	}

	allegiance, ok := worldgen.ParseAllegiance(fields[ext+6])
	if !ok {
		return RecordLine{}, fmt.Errorf("survey: %q: bad allegiance %q", line, fields[ext+6])
	}

	stars, err := systemgen.ParseStellar(strings.Join(fields[ext+7:], " "))
	if err != nil {
		return RecordLine{}, fmt.Errorf("survey: %q: %w", line, err)
	}

	return RecordLine{
		Hex: hex, Name: fields[1], Mainworld: world,
		PopDigit: pop, Belts: belts, Giants: giants,
		Worlds: worlds, Allegiance: allegiance, Stars: stars,
	}, nil
}

// parsePBG decodes the three eHex digits of the PBG field (population multiplier,
// belts, gas giants).
func parsePBG(s string) (int, int, int, error) {
	if len(s) != 3 {
		return 0, 0, 0, fmt.Errorf("PBG %q is not three digits", s)
	}

	d, err := ehex.ParseDigits(s)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("PBG %q: %w", s, err)
	}

	return d[0], d[1], d[2], nil
}
