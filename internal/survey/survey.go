// Package survey composes sectorgen and systemgen into a detailed sector survey:
// every hex that holds a star system (Book 3 p.13 System Contents) is generated
// as a full systemgen.System and rendered as a canonical Second Survey line, and
// the subsector capital is identified (the stub systemgen leaves to region
// context).
package survey

import (
	"strings"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/sectorgen"
	"github.com/philoserf/t5/internal/systemgen"
	"github.com/philoserf/t5/internal/worldgen"
)

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
// Importance Starport-A world is marked the subsector capital (Cs).
func Subsector(r *dice.Roller, d sectorgen.Density, letter byte) []Record {
	return detail(r, sectorgen.GenerateSubsector(r, d, letter))
}

// Sector surveys a whole sector the same way Subsector does (one capital marked
// across all 1280 hexes).
func Sector(r *dice.Roller, d sectorgen.Density) []Record {
	return detail(r, sectorgen.GenerateSector(r, d))
}

// detail generates a full system and name for each present hex, then marks the
// capital. The coarse survey flags (gas giant / asteroid) are the long-range
// preview; the full generation supersedes them.
func detail(r *dice.Roller, hexes []sectorgen.StellarHex) []Record {
	records := make([]Record, len(hexes))
	for i, h := range hexes {
		records[i] = Record{
			Hex:    h.Hex,
			Name:   worldName(r),
			System: systemgen.Generate(r),
		}
	}
	markCapital(records)
	return records
}

// markCapital tags the highest-Importance Starport-A world as the subsector
// capital (Cs) and recomputes its nobility as a capital (Book 3 Chart D p.26:
// capitals require Starport A). If no world has Starport A, none is marked.
func markCapital(records []Record) {
	best := -1
	for i := range records {
		if records[i].System.Mainworld.Profile.Starport != 'A' {
			continue
		}
		if best < 0 || records[i].System.Mainworld.Importance > records[best].System.Mainworld.Importance {
			best = i
		}
	}
	if best < 0 {
		return
	}
	w := &records[best].System.Mainworld
	w.TradeCodes = append(w.TradeCodes, "Cs")
	w.Nobility = worldgen.Nobility(w.TradeCodes, w.Importance, true)
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
	for range 2 + index(r, 2) {
		b.WriteString(nameConsonants[index(r, len(nameConsonants))])
		b.WriteString(nameVowels[index(r, len(nameVowels))])
	}
	name := b.String()
	return strings.ToUpper(name[:1]) + name[1:]
}

// index returns a roughly uniform value in [0, n) by assembling enough base-6
// die rolls to cover n.
func index(r *dice.Roller, n int) int {
	x := 0
	for pow := 1; pow < n; pow *= 6 {
		x = x*6 + r.Die() - 1
	}
	return x % n
}
