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
// Importance Starport-A world is marked the subsector capital (Cs). The coarse
// survey flags (gas giant / asteroid) are the long-range preview; the full
// system generation supersedes them.
func Subsector(r *dice.Roller, d sectorgen.Density, letter byte) []Record {
	hexes := sectorgen.GenerateSubsector(r, d, letter)
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
// capital (Cs; Book 3 Chart D p.26 requires Starport A). If no world has Starport
// A, none is marked.
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
	if best >= 0 {
		records[best].System.Mainworld.SetCapital("Cs")
	}
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
