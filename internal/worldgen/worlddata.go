package worldgen

import (
	"slices"
	"strings"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/tradecode"
	"github.com/philoserf/t5/internal/uwp"
)

// Additional per-world data from Book 3 Chart F (p. 28): Nobility, Bases,
// Travel Zone, and Native Status.

// nobleRank pairs a nobility code with the trade codes that grant it.
var nobleRanks = []struct {
	code  string
	codes []tradecode.Code
}{
	{"c", []tradecode.Code{tradecode.Pa, tradecode.Pr}}, // Baronet
	{"C", []tradecode.Code{tradecode.Ag, tradecode.Ri}}, // Baron
	{"D", []tradecode.Code{tradecode.Pi}},               // Marquis
	{"e", []tradecode.Code{tradecode.Ph}},               // Viscount
	{"E", []tradecode.Code{tradecode.In, tradecode.Hi}}, // Count
}

// Nobility returns the string of noble codes a world holds (Book 3 p. 28). A
// Knight (B) is always present; further ranks follow from trade codes, from a
// capital designation (Duke F), or from an Importance of 4+ on a non-capital
// (Duke f). Archduke (G) is assigned at the domain level and is not derived
// here. isCapital marks a subsector or sector capital.
func Nobility(tcs []tradecode.Code, ix int, isCapital bool) string {
	var b strings.Builder
	b.WriteString("B") // Knight, always present

	for _, rank := range nobleRanks {
		for _, code := range rank.codes {
			if slices.Contains(tcs, code) {
				b.WriteString(rank.code)

				break
			}
		}
	}

	switch {
	case isCapital:
		b.WriteString("F") // Duke of a capital
	case ix >= 4:
		b.WriteString("f") // Duke of an important, non-capital world
	}

	return b.String()
}

// RollBases rolls for the presence of Naval and Scout bases, each present when
// 2D rolls at or under a starport-dependent threshold (Book 3 p. 28). Starports
// C, E, and X support no Naval base; E and X support no Scout base. Naval Depot
// and Way Station need region context and are not rolled here.
func RollBases(r *dice.Roller, starport byte) (bool, bool) {
	var naval, scout bool

	switch starport {
	case 'A':
		naval = r.Dice(2) <= 6
	case 'B':
		naval = r.Dice(2) <= 5
	}

	switch starport {
	case 'A':
		scout = r.Dice(2) <= 4
	case 'B':
		scout = r.Dice(2) <= 5
	case 'C':
		scout = r.Dice(2) <= 6
	case 'D':
		scout = r.Dice(2) <= 7
	}

	return naval, scout
}

// ZoneName names a travel-zone code (Book 3 p. 28); anything but Amber or Red is
// Green, the common case.
func ZoneName(zone byte) string {
	switch zone {
	case 'A':
		return "Amber"
	case 'R':
		return "Red"
	default:
		return "Green"
	}
}

// TravelZone classifies a world as Green ('G'), Amber ('A'), or Red ('R') from
// its Government+Law level and starport (Book 3 p. 28): Gov+Law of 20+ is Amber,
// 22+ is Red, and a class-X starport is Red.
func TravelZone(p uwp.Profile) byte {
	switch {
	case p.Starport == 'X':
		return 'R'
	case p.Government+p.Law >= 22:
		return 'R'
	case p.Government+p.Law >= 20:
		return 'A'
	default:
		return 'G'
	}
}

// ZoneCodes returns the trade classifications a world's travel zone earns it
// (Book 3 p. 28, and the same rows on Chart D p. 26): an Amber world is Dangerous
// (Da) if its population is 6 or less and Puzzling (Pz) if 7 or more, and a Red
// world is Forbidden (Fo). A Green world earns none.
//
// These sit in TradeClassifications' deliberate exclusions no longer: the zone is
// a pure function of the UWP (see TravelZone), so the codes are as determinable as
// any other. The catalog once called them referee-only; the book gives the rule
// outright.
func ZoneCodes(p uwp.Profile) []tradecode.Code {
	switch TravelZone(p) {
	case 'A':
		if p.Population <= 6 {
			return []tradecode.Code{tradecode.Da}
		}

		return []tradecode.Code{tradecode.Pz}
	case 'R':
		return []tradecode.Code{tradecode.Fo}
	default:
		return nil
	}
}

// NativeStatus describes the origin of a world's inhabitants (Book 3 p. 28,
// NIL), classified from Population, Atmosphere, Tech Level, and Government. A
// Government of 1 or 6 overrides to Corporate or Colonists. Atmosphere is read
// in three bands: thin (0–1), exotic (A–C), and standard (2–9, D–F).
func NativeStatus(p uwp.Profile) string { //nolint:cyclop // transcribes the p.28 native-status table
	thinAtm := p.Atmosphere <= 1
	exoticAtm := p.Atmosphere >= 10 && p.Atmosphere <= 12
	developed := p.TechLevel >= 1

	// An uninhabited world is classified purely by atmosphere and tech
	// evidence; the Government override below applies only to inhabited worlds.
	if p.Population == 0 {
		switch {
		case thinAtm:
			if developed {
				return "Vanished Transplants"
			}

			return "None"
		case exoticAtm:
			if developed {
				return "Catastrophic EXN"
			}

			return "Extinct Exotic Natives"
		default:
			if developed {
				return "Catastrophic XN"
			}

			return "Extinct Natives"
		}
	}

	// Chart F's inhabited rows all read "TL 1+"; a populated world at TL 0 is not on
	// the chart at all. Its inhabitants are living, not extinct, so TL does not
	// change their label the way it does for the depopulated evidence-of-life rows
	// above — a Pop-1 TL-0 world is still Transients. TL is not re-checked here.
	switch p.Government {
	case 1:
		return "Corporate"
	case 6:
		return "Colonists"
	}

	switch {
	case p.Population <= 3:
		return "Transients"
	case p.Population <= 6:
		return "Settlers"
	case thinAtm:
		return "Transplants"
	case exoticAtm:
		return "Exotic Natives"
	default:
		return "Natives"
	}
}
