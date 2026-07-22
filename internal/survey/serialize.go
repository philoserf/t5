package survey

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/philoserf/t5/internal/ehex"
	"github.com/philoserf/t5/internal/route"
	"github.com/philoserf/t5/internal/sectorgen"
	"github.com/philoserf/t5/internal/systemgen"
	"github.com/philoserf/t5/internal/tradecode"
	"github.com/philoserf/t5/internal/worldgen"
)

const (
	routeMetadataPrefix = "# Route: "
	ownerMetadataPrefix = "# Owner: "
)

// SEC is this package's lossless sector-document representation: the existing
// Second Survey record lines plus comment metadata for relationships that a
// world line cannot carry. Book 3 specifies O:nnnn for a Cy colony's owner but
// does not specify a text encoding for trade routes; "# Route:" is therefore a
// documented survey convention, not a claimed T5SS standard.
type SEC struct {
	Records    []RecordLine
	Routes     []route.Link
	Ownerships []Ownership
}

// SEC renders a generated survey as a lossless .sec document. Its world lines
// are exactly Record.SecondSurvey output; adding relationship metadata therefore
// does not change the existing Second Survey stream.
func (s Survey) SEC() string {
	lines := make([]string, len(s.Records))
	for i, rec := range s.Records {
		lines[i] = rec.SecondSurvey()
	}

	return renderSEC(lines, s.Routes, s.Ownerships)
}

// String renders a parsed SEC document in the same format as Survey.SEC.
func (s SEC) String() string {
	lines := make([]string, len(s.Records))
	for i, rec := range s.Records {
		lines[i] = rec.String()
	}

	return renderSEC(lines, s.Routes, s.Ownerships)
}

// renderSEC assembles a .sec document from already-rendered world lines: the
// one difference between Survey.SEC (Record.SecondSurvey) and SEC.String
// (RecordLine.String), sharing everything else — the metadata block and the
// trailing-newline trim.
func renderSEC(lines []string, routes []route.Link, ownerships []Ownership) string {
	var b strings.Builder

	for _, line := range lines {
		b.WriteString(line)
		b.WriteByte('\n')
	}

	writeSECMetadata(&b, routes, ownerships)

	return strings.TrimRight(b.String(), "\n")
}

func writeSECMetadata(b *strings.Builder, routes []route.Link, ownerships []Ownership) {
	if len(routes) == 0 && len(ownerships) == 0 {
		return
	}

	if b.Len() > 0 {
		b.WriteByte('\n')
	}

	for _, link := range routes {
		fmt.Fprintf(b, "%s%s %s J%d\n", routeMetadataPrefix, link.From, link.To, link.Jump)
	}

	for _, own := range ownerships {
		fmt.Fprintf(b, "%s%s O:%s\n", ownerMetadataPrefix, own.Colony, own.Owner)
	}
}

// ParseSEC decodes the document emitted by Survey.SEC. Blank lines and unrelated
// comments are ignored; malformed Route/Owner comments using this format fail.
func ParseSEC(data string) (SEC, error) {
	var doc SEC

	for lineNo, line := range strings.Split(strings.ReplaceAll(data, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") &&
			!strings.HasPrefix(line, routeMetadataPrefix) && !strings.HasPrefix(line, ownerMetadataPrefix) {
			continue
		}

		switch {
		case strings.HasPrefix(line, routeMetadataPrefix):
			link, err := parseRouteMetadata(strings.TrimPrefix(line, routeMetadataPrefix))
			if err != nil {
				return SEC{}, fmt.Errorf("survey: line %d: %w", lineNo+1, err)
			}

			doc.Routes = append(doc.Routes, link)
		case strings.HasPrefix(line, ownerMetadataPrefix):
			own, err := parseOwnerMetadata(strings.TrimPrefix(line, ownerMetadataPrefix))
			if err != nil {
				return SEC{}, fmt.Errorf("survey: line %d: %w", lineNo+1, err)
			}

			doc.Ownerships = append(doc.Ownerships, own)
		default:
			rec, err := ParseRecord(line)
			if err != nil {
				return SEC{}, fmt.Errorf("survey: line %d: %w", lineNo+1, err)
			}

			doc.Records = append(doc.Records, rec)
		}
	}

	if err := validateSECRelationships(doc); err != nil {
		return SEC{}, err
	}

	return doc, nil
}

func validateSECRelationships(doc SEC) error {
	records := make(map[sectorgen.Hex]RecordLine, len(doc.Records))
	for _, rec := range doc.Records {
		records[rec.Hex] = rec
	}

	for _, link := range doc.Routes {
		if _, ok := records[link.From]; !ok {
			return fmt.Errorf("survey: route start %s has no world record", link.From)
		}

		if _, ok := records[link.To]; !ok {
			return fmt.Errorf("survey: route end %s has no world record", link.To)
		}
	}

	for _, own := range doc.Ownerships {
		colony, colonyOK := records[own.Colony]

		_, ownerOK := records[own.Owner]
		if !colonyOK || !ownerOK {
			return fmt.Errorf("survey: ownership %s O:%s references a missing world", own.Colony, own.Owner)
		}

		if !slices.Contains(colony.Mainworld.TradeCodes, tradecode.Cy) {
			return fmt.Errorf("survey: owned world %s is not classified Cy", own.Colony)
		}
	}

	return nil
}

func parseRouteMetadata(s string) (route.Link, error) {
	fields := strings.Fields(s)
	if len(fields) != 3 || !strings.HasPrefix(fields[2], "J") {
		return route.Link{}, fmt.Errorf("bad route metadata %q", s)
	}

	from, fromOK := sectorgen.ParseHex(fields[0])

	to, toOK := sectorgen.ParseHex(fields[1])

	jump, jumpErr := strconv.Atoi(strings.TrimPrefix(fields[2], "J"))
	if jumpErr != nil || !fromOK || !toOK || jump < 1 || from.Distance(to) != jump {
		return route.Link{}, fmt.Errorf("bad route metadata %q", s)
	}

	return route.Link{From: from, To: to, Jump: jump}, nil
}

func parseOwnerMetadata(s string) (Ownership, error) {
	fields := strings.Fields(s)
	if len(fields) != 2 || !strings.HasPrefix(fields[1], "O:") {
		return Ownership{}, fmt.Errorf("bad owner metadata %q", s)
	}

	colony, colonyOK := sectorgen.ParseHex(fields[0])

	owner, ownerOK := sectorgen.ParseHex(strings.TrimPrefix(fields[1], "O:"))
	if !colonyOK || !ownerOK || colony == owner || colony.Distance(owner) > 6 {
		return Ownership{}, fmt.Errorf("bad owner metadata %q", s)
	}

	return Ownership{Colony: colony, Owner: owner}, nil
}

// Serialization for a full Second Survey record (#327). A record
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
