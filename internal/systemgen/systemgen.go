// Package systemgen creates a Traveller5 star system: its stars, the number of
// gas giants, planetoid belts, and worlds, and the mainworld profile. It builds
// on internal/worldgen for the mainworld and follows Book 3's Master System
// Generation Checklist and the System Generation card (pp. 16-17, 28).
//
// Orbital placement and per-world detailing are deferred; this generates the
// system's census and stellar makeup.
package systemgen

import (
	"fmt"
	"strings"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/uwp"
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

	GasGiants int
	Belts     int
	Worlds    int
	Mainworld uwp.Profile
}

// Generate rolls a complete system: the mainworld, the stars, and the counts.
func Generate(r *dice.Roller) System {
	s := System{Mainworld: worldgen.Generate(r)}

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

	s.GasGiants = gasGiants(r)
	s.Belts = belts(r)
	s.Worlds = 1 + s.GasGiants + s.Belts + r.Dice(2)
	return s
}

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
	for _, e := range []struct {
		label string
		star  *Star
	}{
		{"Primary Companion", s.PrimaryCompanion},
		{"Close", s.Close},
		{"Close Companion", s.CloseCompanion},
		{"Near", s.Near},
		{"Near Companion", s.NearCompanion},
		{"Far", s.Far},
		{"Far Companion", s.FarCompanion},
	} {
		if e.star != nil {
			fmt.Fprintf(&b, "%s: %s\n", e.label, *e.star)
		}
	}
	fmt.Fprintf(&b, "Gas Giants: %d  Belts: %d  Worlds: %d\n", s.GasGiants, s.Belts, s.Worlds)
	fmt.Fprintf(&b, "Mainworld: %s", s.Mainworld)
	return b.String()
}
