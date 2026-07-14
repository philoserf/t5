// Package survey composes sectorgen and systemgen into a detailed sector survey:
// every hex that holds a star system (Book 3 p.13 System Contents) is generated
// as a full systemgen.System and rendered as a canonical Second Survey line, and
// the region-wide facts a single system cannot know — capitals, trade routes, and
// way stations — are settled across it. A sector is surveyed whole; callers select
// the part they want (Survey.Subsector, Survey.At) rather than surveying it.
package survey

import (
	"fmt"
	"sort"
	"strings"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/route"
	"github.com/philoserf/t5/internal/sectorgen"
	"github.com/philoserf/t5/internal/systemgen"
)

// wayStationSpacing is the trade-route length, in parsecs, per Scout Way Station
// (Book 3 p.28: "1 per 50 parsecs on trade route").
const wayStationSpacing = 50

// A Record is one surveyed hex: its location, a generated world name, and the
// full star system generated for it.
type Record struct {
	Hex    sectorgen.Hex
	Name   string
	System systemgen.System
}

// SecondSurvey renders the record's canonical one-line system record (hex, name,
// UWP, extensions, bases, PBG, worlds, allegiance, stellar data). Allegiance
// defaults to Imperial.
func (rec Record) SecondSurvey() string {
	return rec.System.SecondSurvey(rec.Hex.String(), rec.Name, "Im")
}

// A Survey is a surveyed region: its per-hex system records and the trade routes
// linking their Important worlds (Book 3 pp. 21, 27).
type Survey struct {
	Records []Record
	Routes  []route.Link
}

// Sector surveys an entire sector (all 1280 hexes) at the given density: every
// present hex gets a full generated system and name, each subsector's highest-
// Importance Starport-A world is marked its capital (Cs) and the region's is the
// sector capital (Cx), then trade routes are laid among the Important worlds and
// Scout Way Stations sited along them. The coarse gas-giant and asteroid map
// symbols constrain the full generation, so the long-range preview and the
// detailed system agree.
//
// A sector is always surveyed whole, and callers select the part they want (see
// Subsector and At). A region cannot be surveyed piecemeal: a world's own printed
// record depends on the rest of the sector — whether it is a capital is decided
// against every other world, and its bases and Importance shift when a trade route
// running through it earns a way station. (The systems also share one dice stream,
// so a hex's world depends on every hex rolled before it, but that is an artifact
// of the roller; the region-wide passes are the reason that would survive any
// redesign.)
func Sector(r *dice.Roller, d sectorgen.Density) Survey {
	hexes := sectorgen.GenerateSector(r, d)
	records := make([]Record, len(hexes))
	for i, h := range hexes {
		records[i] = Record{
			Hex:    h.Hex,
			Name:   worldName(r),
			System: systemgen.GenerateForMap(r, h.GasGiant, h.AsteroidMainworld),
		}
	}
	// Capitals from base Importance, then routes, then Way Stations (which bump
	// Importance) — a single pass; a Way Station's +1 does not re-trigger route
	// or capital selection.
	markSectorCapitals(records)
	links := route.Build(worldsOf(records), route.DefaultJump)
	placeWayStations(records, links)
	return Survey{Records: records, Routes: links}
}

// At finds the record for a hex, reporting whether that hex holds a star system.
// Records stay in CCRR order (String, route building, and way-station placement
// all depend on it), so this is a scan rather than a map lookup — it runs once
// per drill-in, against a sector that took far longer to generate.
func (s Survey) At(hex sectorgen.Hex) (Record, bool) {
	for _, rec := range s.Records {
		if rec.Hex == hex {
			return rec, true
		}
	}
	return Record{}, false
}

// Subsector selects the records of one subsector from the survey, by its letter
// A-P (lower case accepted). It is a view of an already-surveyed sector, not a
// survey of its own: a subsector generated in isolation would hold different
// worlds (see Sector), so selecting is the only way to see part of a region and
// still agree with the rest of it.
//
// A subsector with no star system selects nothing, and so does a letter outside
// A-P — callers taking a letter from a user should reject it first with
// sectorgen.ParseSubsector rather than read an empty result as an empty region.
func (s Survey) Subsector(letter byte) []Record {
	if letter >= 'a' && letter <= 'z' {
		letter -= 'a' - 'A'
	}
	var out []Record
	for _, rec := range s.Records {
		if rec.Hex.Subsector() == letter {
			out = append(out, rec)
		}
	}
	return out
}

// String renders the survey: each hex's canonical Second Survey line annotated
// with its expected weekly ship traffic, followed by a trade-route listing.
func (s Survey) String() string {
	var b strings.Builder
	for _, rec := range s.Records {
		fmt.Fprintf(&b, "%s  [~%d/wk]\n", rec.SecondSurvey(), route.ExpectedTraffic(rec.System.Mainworld.Importance))
	}
	if len(s.Routes) > 0 {
		fmt.Fprintf(&b, "\nTrade Routes (%d):\n", len(s.Routes))
		for _, l := range s.Routes {
			fmt.Fprintf(&b, "  %s-%s J%d\n", l.From, l.To, l.Jump)
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// worldsOf projects the survey records to the route package's world summaries.
func worldsOf(records []Record) []route.World {
	worlds := make([]route.World, len(records))
	for i, rec := range records {
		worlds[i] = route.World{Hex: rec.Hex, Importance: rec.System.Mainworld.Importance}
	}
	return worlds
}

// markSectorCapitals marks the region's capital and each subsector's, skipping the
// subsector that already holds the sector capital so no world carries both and no
// subsector gets two.
//
// The codes are Book 3 Chart D (p.26): Cp is a Subsector Capital, Cs is a Sector
// Capital, and Cx is the Imperial Capital. They were swapped here — a sector
// capital was marked Cx and a subsector capital Cs — so every generated sector
// stamped the wrong codes into the Second Survey line, promoting each of its
// sixteen subsector capitals to a sector capital and its sector capital to the
// capital of the Imperium. (chargen/homeworld.go had them right all along, which
// is how the repo came to disagree with itself.)
//
// Cx is not generated: there is one Imperial Capital, and which sector holds it is
// not a fact about the sector being surveyed.
func markSectorCapitals(records []Record) {
	sectorCap := bestCapital(records, nil)
	if sectorCap >= 0 {
		records[sectorCap].System.Mainworld.SetCapital("Cs")
	}
	bySub := map[byte][]int{}
	for i := range records {
		l := records[i].Hex.Subsector()
		bySub[l] = append(bySub[l], i)
	}
	for _, idxs := range bySub {
		if b := bestCapital(records, idxs); b >= 0 && b != sectorCap {
			records[b].System.Mainworld.SetCapital("Cp")
		}
	}
}

// bestCapital returns the index (into records) of the highest-Importance
// Starport-A world among idxs, or -1 if none qualifies (Book 3 Chart D p.26).
// A nil idxs considers every record; an empty (non-nil) one considers none.
func bestCapital(records []Record, idxs []int) int {
	best := -1
	consider := func(i int) {
		if records[i].System.Mainworld.Profile.Starport != 'A' {
			return
		}
		if best < 0 || records[i].System.Mainworld.Importance > records[best].System.Mainworld.Importance {
			best = i
		}
	}
	if idxs == nil {
		for i := range records {
			consider(i)
		}
		return best
	}
	for _, i := range idxs {
		consider(i)
	}
	return best
}

// placeWayStations sites Scout Way Stations along the trade routes at a density
// of about one per wayStationSpacing parsecs of total route length (Book 3
// p.28). The book gives a frequency, not exact placement, so stations go to the
// busiest route hubs — the worlds touched by the most links, ties broken by CCRR
// order. Setting a station bumps the host's Importance (+1).
func placeWayStations(records []Record, links []route.Link) {
	total := 0
	degree := map[sectorgen.Hex]int{}
	for _, l := range links {
		total += l.Jump
		degree[l.From]++
		degree[l.To]++
	}
	n := total / wayStationSpacing
	if n == 0 {
		return
	}
	hubs := make([]sectorgen.Hex, 0, len(degree))
	for h := range degree {
		hubs = append(hubs, h)
	}
	sort.Slice(hubs, func(i, j int) bool {
		if degree[hubs[i]] != degree[hubs[j]] {
			return degree[hubs[i]] > degree[hubs[j]]
		}
		return beforeHex(hubs[i], hubs[j])
	})
	if n > len(hubs) {
		n = len(hubs)
	}
	byHex := map[sectorgen.Hex]int{}
	for i := range records {
		byHex[records[i].Hex] = i
	}
	for _, h := range hubs[:n] {
		records[byHex[h]].System.Mainworld.SetWayStation()
	}
}

// beforeHex orders hexes in column-major (CCRR) order.
func beforeHex(a, b sectorgen.Hex) bool {
	if a.Col != b.Col {
		return a.Col < b.Col
	}
	return a.Row < b.Row
}

// nameConsonants and nameVowels build simple pronounceable world names; T5 has
// no name generator, so this supplies a placeholder until one exists.
var (
	nameConsonants = []string{"b", "c", "d", "f", "g", "h", "j", "k", "l", "m", "n", "p", "r", "s", "t", "v", "z"}
	nameVowels     = []string{"a", "e", "i", "o", "u", "y", "ae", "ia"}
)

// worldName builds a two- or three-syllable pronounceable name.
func worldName(r *dice.Roller) string {
	var b strings.Builder
	for range 2 + r.Index(2) { // 2 or 3 syllables
		b.WriteString(nameConsonants[r.Index(len(nameConsonants))])
		b.WriteString(nameVowels[r.Index(len(nameVowels))])
	}
	name := b.String()
	return strings.ToUpper(name[:1]) + name[1:]
}
