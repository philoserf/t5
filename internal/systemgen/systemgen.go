// Package systemgen creates a Traveller5 star system: its stars, the number of
// gas giants, planetoid belts, and worlds, and the mainworld profile. It builds
// on internal/worldgen for the mainworld and follows Book 3's Master System
// Generation Checklist and the System Generation card (pp. 16-17, 28).
//
// Secondary stars are placed in orbit bands; per-world detailing (worlds in
// orbits, habitable zone) is deferred.
package systemgen

import (
	"fmt"
	"strings"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/ehex"
	"github.com/philoserf/t5/internal/worldgen"
)

// starPresent is the Flux threshold at or above which an optional star exists
// (Book 3 p. 28: present on a Flux of +3, +4, or +5).
const starPresent = 3

// A System is a generated star system. The Primary always exists; the other
// stars are present only when rolled, so they are pointers that are nil when
// absent. Each of Primary, Close, Near, and Far may have a companion.
type System struct {
	Primary          Star
	PrimaryCompanion *Star
	Close            *Star
	CloseCompanion   *Star
	Near             *Star
	NearCompanion    *Star
	Far              *Star
	FarCompanion     *Star

	// Orbit numbers of the secondary stars around the Primary (Book 3 p. 28),
	// valid only when the corresponding star is present. A companion orbits
	// inside its own star's orbit and has no separate number; the Primary's
	// companion sits inside Orbit 0.
	CloseOrbit int
	NearOrbit  int
	FarOrbit   int

	GasGiants int
	Belts     int
	Worlds    int
	Mainworld worldgen.World
}

// Generate rolls a complete system: the gas-giant and belt counts, the
// mainworld built from them, then the stars, the world count, and the orbits.
func Generate(r *dice.Roller) System {
	var s System
	// Gas giants and belts are needed to detail the mainworld (the Economic
	// Extension and the PBG use them), so roll them before the mainworld.
	s.GasGiants = gasGiants(r)
	s.Belts = belts(r)
	// Capital status needs region context (sectorgen), so default to false.
	s.Mainworld = worldgen.GenerateWorld(r, s.GasGiants, s.Belts, false)

	var primaryType, primarySize int
	s.Primary, primaryType, primarySize = rollStar(r, true, 0, 0)

	// Every non-primary star — secondaries and companions alike — derives from
	// the primary's Flux values (Book 3 p. 28). present rolls one optional star,
	// returning nil unless its presence Flux clears the threshold.
	present := func() *Star {
		if r.Flux() < starPresent {
			return nil
		}
		star, _, _ := rollStar(r, false, primaryType, primarySize)
		return &star
	}

	s.PrimaryCompanion = present()
	s.Close = present()
	s.Near = present()
	s.Far = present()
	if s.Close != nil {
		s.CloseCompanion = present()
	}
	if s.Near != nil {
		s.NearCompanion = present()
	}
	if s.Far != nil {
		s.FarCompanion = present()
	}

	s.Worlds = 1 + s.GasGiants + s.Belts + r.Dice(2)

	// Place the secondary stars in their orbit bands. Rolled after the counts
	// so placement does not shift the stars and counts already set for this
	// system. (Across multiple systems from one roller these rolls do advance
	// the shared stream, as any added roll would; each system stays
	// reproducible for a given seed and version.)
	if s.Close != nil {
		s.CloseOrbit = closeOrbit(r)
	}
	if s.Near != nil {
		s.NearOrbit = nearOrbit(r)
	}
	if s.Far != nil {
		s.FarOrbit = farOrbit(r)
	}
	return s
}

// closeOrbit is 1D-1 (orbits 0-5); nearOrbit is 5+1D (orbits 6-11); farOrbit is
// 11+1D (orbits 12-17). Book 3 p. 28, "Place Stars in Orbits".
func closeOrbit(r *dice.Roller) int { return r.Die() - 1 }
func nearOrbit(r *dice.Roller) int  { return 5 + r.Die() }
func farOrbit(r *dice.Roller) int   { return 11 + r.Die() }

// gasGiants is 2D/2-2, dropping the fraction and flooring at zero (range 0-4).
func gasGiants(r *dice.Roller) int {
	return max(r.Dice(2)/2-2, 0)
}

// belts is 1D-3, floored at zero (range 0-3).
func belts(r *dice.Roller) int {
	return max(r.Die()-3, 0)
}

// String renders a one-block summary of the system.
func (s System) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Primary: %s\n", s.Primary)
	// orbit is -1 for stars without a numbered orbit (companions).
	for _, e := range []struct {
		label string
		star  *Star
		orbit int
	}{
		{"Primary Companion", s.PrimaryCompanion, -1},
		{"Close", s.Close, s.CloseOrbit},
		{"Close Companion", s.CloseCompanion, -1},
		{"Near", s.Near, s.NearOrbit},
		{"Near Companion", s.NearCompanion, -1},
		{"Far", s.Far, s.FarOrbit},
		{"Far Companion", s.FarCompanion, -1},
	} {
		if e.star == nil {
			continue
		}
		if e.orbit >= 0 {
			fmt.Fprintf(&b, "%s: %s (Orbit %d)\n", e.label, *e.star, e.orbit)
		} else {
			fmt.Fprintf(&b, "%s: %s\n", e.label, *e.star)
		}
	}
	// PBG already carries the belt and gas-giant counts (its 2nd and 3rd
	// digits), so they are not labelled separately.
	fmt.Fprintf(&b, "Worlds: %d  PBG: %s\n", s.Worlds, s.PBG())
	fmt.Fprintf(&b, "Mainworld: %s", s.Mainworld.SecondSurvey())
	return b.String()
}

// PBG renders the Population-Belts-Giants digits: the mainworld's population
// multiplier digit, the planetoid-belt count, and the gas-giant count, each an
// eHex digit (e.g. Regina's "703").
func (s System) PBG() string {
	return fmt.Sprintf("%s%s%s",
		ehex.Format(s.Mainworld.PopulationDigit), ehex.Format(s.Belts), ehex.Format(s.GasGiants))
}

// Stellar renders the system's stars as a compact space-joined list, primary
// first (e.g. "F8 V M6 VI F5 VI").
func (s System) Stellar() string {
	stars := []string{s.Primary.String()}
	for _, st := range []*Star{s.PrimaryCompanion, s.Close, s.CloseCompanion, s.Near, s.NearCompanion, s.Far, s.FarCompanion} {
		if st != nil {
			stars = append(stars, st.String())
		}
	}
	return strings.Join(stars, " ")
}

// SecondSurvey renders the canonical one-line system record:
//
//	hex  name  UWP TCs {Ix}(Ex)[Cx] N B Z  PBG  Worlds  allegiance  Stellar
//
// hex and name are supplied by the caller (they need sector context); allegiance
// defaults to "Im" (Imperial) when empty.
func (s System) SecondSurvey(hex, name, allegiance string) string {
	if allegiance == "" {
		allegiance = "Im"
	}
	return fmt.Sprintf("%s %s %s %s %d %s %s",
		hex, name, s.Mainworld.SecondSurvey(), s.PBG(), s.Worlds, allegiance, s.Stellar())
}
