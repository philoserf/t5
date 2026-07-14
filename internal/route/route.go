// Package route builds Traveller5 trade routes over a surveyed region (Book 3
// pp. 21, 27). Trade routes are deterministic — no dice: Important worlds
// (Ix >= 4) are linked to other Important worlds within jump range, and where no
// direct link is possible they are joined through intermediate lesser worlds.
// The result is a pure graph over each world's hex location and Importance.
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

// A World is one surveyed system relevant to routing: its hex and its Importance
// value. Worlds below the Important threshold are never route endpoints but may
// serve as intermediate stepping stones.
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

// Build returns the trade-route links among the given worlds (Book 3 pp. 21,
// 27). A link joins every pair of Important worlds (Ix >= Important) within
// maxJump parsecs; an Important world with no such direct link is joined to the
// nearest reachable Important world by a shortest hop-path through intermediate
// worlds (each hop <= maxJump). A non-positive maxJump uses DefaultJump. Links
// are returned de-duplicated, in stable CCRR order.
func Build(worlds []World, maxJump int) []Link {
	if maxJump <= 0 {
		maxJump = DefaultJump
	}

	// Work in a stable CCRR order so adjacency (and thus BFS) is deterministic.
	sorted := append([]World(nil), worlds...)
	sort.Slice(sorted, func(i, j int) bool { return before(sorted[i].Hex, sorted[j].Hex) })

	// adj[i] lists the indices within maxJump of world i, in ascending (CCRR)
	// order since sorted is CCRR-ordered.
	adj := make([][]int, len(sorted))
	isImportant := make([]bool, len(sorted))
	var importantIdx []int
	for i, w := range sorted {
		if w.Importance >= Important {
			isImportant[i] = true
			importantIdx = append(importantIdx, i)
		}
	}
	for i := range sorted {
		for j := range sorted {
			if i == j {
				continue
			}
			if d := sorted[i].Hex.Distance(sorted[j].Hex); d >= 1 && d <= maxJump {
				adj[i] = append(adj[i], j)
			}
		}
	}

	links := map[Link]bool{}
	addPath := func(path []int) {
		for k := 1; k < len(path); k++ {
			links[orderedLink(sorted[path[k-1]].Hex, sorted[path[k]].Hex)] = true
		}
	}

	// Direct links among Important worlds within range.
	connected := make([]bool, len(sorted))
	for _, i := range importantIdx {
		for _, j := range adj[i] {
			if isImportant[j] {
				links[orderedLink(sorted[i].Hex, sorted[j].Hex)] = true
				connected[i] = true
			}
		}
	}

	// Intermediate connections: an Important world with no direct Important
	// neighbour is joined to the nearest reachable Important world by a shortest
	// hop-path through any worlds (Book 3 p.21). Unreachable ones stay unlinked.
	for _, start := range importantIdx {
		if connected[start] {
			continue
		}
		addPath(shortestPathToImportant(start, adj, isImportant))
	}

	out := make([]Link, 0, len(links))
	for l := range links {
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return before(out[i].From, out[j].From)
		}
		return before(out[i].To, out[j].To)
	})
	return out
}

// shortestPathToImportant BFS-searches from start for the nearest other
// Important world, returning the path of indices (start..target) or nil if none
// is reachable. Neighbours are visited in CCRR order, so ties break
// deterministically toward the lower-CCRR world.
func shortestPathToImportant(start int, adj [][]int, isImportant []bool) []int {
	prev := map[int]int{start: -1}
	for queue := []int{start}; len(queue) > 0; {
		cur := queue[0]
		queue = queue[1:]
		if cur != start && isImportant[cur] {
			var path []int
			for at := cur; at != -1; at = prev[at] {
				path = append(path, at)
			}
			for l, r := 0, len(path)-1; l < r; l, r = l+1, r-1 {
				path[l], path[r] = path[r], path[l]
			}
			return path
		}
		for _, n := range adj[cur] {
			if _, seen := prev[n]; !seen {
				prev[n] = cur
				queue = append(queue, n)
			}
		}
	}
	return nil
}

// orderedLink builds a Link with its endpoints in CCRR order so undirected links
// de-duplicate regardless of which end they were found from.
func orderedLink(a, b sectorgen.Hex) Link {
	if before(b, a) {
		a, b = b, a
	}
	return Link{From: a, To: b, Jump: a.Distance(b)}
}

// before orders hexes in column-major (CCRR) order.
func before(a, b sectorgen.Hex) bool {
	if a.Col != b.Col {
		return a.Col < b.Col
	}
	return a.Row < b.Row
}
