// Package route builds Traveller5 trade routes over a surveyed region (Book 3
// pp. 21, 27). Trade routes are deterministic — no dice: Important worlds
// (Ix >= 4) are linked to other Important worlds within jump range. The result
// is a pure graph over each world's hex location and Importance value.
package route

import (
	"sort"

	"github.com/philoserf/t5/internal/sectorgen"
)

// Important is the Importance-Extension threshold at or above which a world is
// an "Important World" that anchors trade routes (Book 3 p.27: "+4 or greater").
const Important = 4

// DefaultJump is the maximum jump length of a trade-route link (Book 3 p.27:
// "linked by established Trade Routes of J-4 or less").
const DefaultJump = 4

// A World is the trade-route-relevant summary of a surveyed system: its hex and
// its Importance value.
type World struct {
	Hex        sectorgen.Hex
	Importance int
}

// A Link is one trade-route segment between two worlds, with its jump length in
// parsecs.
type Link struct {
	From, To sectorgen.Hex
	Jump     int
}

// Build returns the direct trade-route links among the given worlds (Book 3
// p.27): a link joins every pair of Important worlds (Ix >= Important) that lie
// within maxJump parsecs of each other. A non-positive maxJump uses DefaultJump.
// Links are returned in a stable order (by From, then To) independent of the
// input order.
func Build(worlds []World, maxJump int) []Link {
	if maxJump <= 0 {
		maxJump = DefaultJump
	}
	var important []World
	for _, w := range worlds {
		if w.Importance >= Important {
			important = append(important, w)
		}
	}
	var links []Link
	for i := range important {
		for j := i + 1; j < len(important); j++ {
			if d := important[i].Hex.Distance(important[j].Hex); d >= 1 && d <= maxJump {
				links = append(links, Link{From: important[i].Hex, To: important[j].Hex, Jump: d})
			}
		}
	}
	sort.Slice(links, func(i, j int) bool {
		if links[i].From != links[j].From {
			return before(links[i].From, links[j].From)
		}
		return before(links[i].To, links[j].To)
	})
	return links
}

// before orders hexes in column-major (CCRR) order.
func before(a, b sectorgen.Hex) bool {
	if a.Col != b.Col {
		return a.Col < b.Col
	}
	return a.Row < b.Row
}
