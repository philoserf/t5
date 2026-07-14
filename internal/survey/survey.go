// Package survey composes sectorgen and systemgen into a detailed sector survey:
// every hex that holds a star system (Book 3 p.13 System Contents) is generated
// as a full systemgen.System and rendered as a canonical Second Survey line, and
// the subsector capital is identified (the stub systemgen leaves to region
// context).
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

// Subsector surveys one subsector (letter 'A'-'P') at the given density: every
// present hex gets a full generated system and a world name, and the highest-
// Importance Starport-A world is marked the subsector capital (Cs). The coarse
// gas-giant survey flag constrains the full generation (present → at least one
// gas giant, absent → none), so the long-range preview and the detailed system
// agree.
func Subsector(r *dice.Roller, d sectorgen.Density, letter byte) []Record {
	hexes := sectorgen.GenerateSubsector(r, d, letter)
	records := make([]Record, len(hexes))
	for i, h := range hexes {
		records[i] = Record{
			Hex:    h.Hex,
			Name:   worldName(r),
			System: systemgen.GenerateWithGasGiants(r, h.GasGiant),
		}
	}
	markCapital(records)
	return records
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
// Scout Way Stations sited along them.
func Sector(r *dice.Roller, d sectorgen.Density) Survey {
	hexes := sectorgen.GenerateSector(r, d)
	records := make([]Record, len(hexes))
	for i, h := range hexes {
		records[i] = Record{
			Hex:    h.Hex,
			Name:   worldName(r),
			System: systemgen.GenerateWithGasGiants(r, h.GasGiant),
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

// markCapital tags the highest-Importance Starport-A world as the subsector
// capital (Cs; Book 3 Chart D p.26 requires Starport A). If no world has Starport
// A, none is marked.
func markCapital(records []Record) {
	if b := bestCapital(records, indices(records)); b >= 0 {
		records[b].System.Mainworld.SetCapital("Cs")
	}
}

// markSectorCapitals marks the region's capital (Cx) and each subsector's
// capital (Cs), skipping the subsector that already holds the sector capital so
// no world carries both and no subsector gets two.
func markSectorCapitals(records []Record) {
	sectorCap := bestCapital(records, indices(records))
	if sectorCap >= 0 {
		records[sectorCap].System.Mainworld.SetCapital("Cx")
	}
	bySub := map[byte][]int{}
	for i := range records {
		l := records[i].Hex.Subsector()
		bySub[l] = append(bySub[l], i)
	}
	for _, idxs := range bySub {
		if b := bestCapital(records, idxs); b >= 0 && b != sectorCap {
			records[b].System.Mainworld.SetCapital("Cs")
		}
	}
}

// bestCapital returns the index (into records) of the highest-Importance
// Starport-A world among idxs, or -1 if none qualifies (Book 3 Chart D p.26).
func bestCapital(records []Record, idxs []int) int {
	best := -1
	for _, i := range idxs {
		if records[i].System.Mainworld.Profile.Starport != 'A' {
			continue
		}
		if best < 0 || records[i].System.Mainworld.Importance > records[best].System.Mainworld.Importance {
			best = i
		}
	}
	return best
}

// indices returns 0..len(records)-1.
func indices(records []Record) []int {
	idxs := make([]int, len(records))
	for i := range idxs {
		idxs[i] = i
	}
	return idxs
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
