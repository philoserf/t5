package route

import (
	"testing"

	"github.com/philoserf/t5/internal/sectorgen"
)

func TestBuildDirectLinks(t *testing.T) {
	worlds := []World{
		{sectorgen.Hex{Col: 1, Row: 1}, 4},   // A: Important
		{sectorgen.Hex{Col: 1, Row: 3}, 5},   // B: Important, 2 pc from A
		{sectorgen.Hex{Col: 1, Row: 2}, 2},   // unimportant, ignored
		{sectorgen.Hex{Col: 10, Row: 10}, 6}, // C: Important but far (> J-4) from A/B
	}
	links := Build(worlds, DefaultJump)
	if len(links) != 1 {
		t.Fatalf("got %d links, want 1: %+v", len(links), links)
	}
	got := links[0]
	want := Link{From: sectorgen.Hex{Col: 1, Row: 1}, To: sectorgen.Hex{Col: 1, Row: 3}, Jump: 2}
	if got != want {
		t.Errorf("link = %+v, want %+v", got, want)
	}
}

func TestBuildThresholdAndRange(t *testing.T) {
	// A world at Ix 3 is not Important; no link even when in range.
	worlds := []World{
		{sectorgen.Hex{Col: 1, Row: 1}, 4},
		{sectorgen.Hex{Col: 1, Row: 3}, 3}, // in range but below the threshold
	}
	if links := Build(worlds, 0); len(links) != 0 {
		t.Errorf("sub-threshold neighbor linked: %+v", links)
	}

	// Two Important worlds exactly J-4 apart link; one parsec farther does not.
	near := []World{{sectorgen.Hex{Col: 1, Row: 1}, 4}, {sectorgen.Hex{Col: 5, Row: 3}, 4}}
	if d := near[0].Hex.Distance(near[1].Hex); d != 4 {
		t.Fatalf("test fixture distance = %d, want 4", d)
	}
	if links := Build(near, 0); len(links) != 1 {
		t.Errorf("J-4 pair not linked: %+v", links)
	}
	far := []World{{sectorgen.Hex{Col: 1, Row: 1}, 4}, {sectorgen.Hex{Col: 6, Row: 3}, 4}}
	if links := Build(far, 0); len(links) != 0 {
		t.Errorf("beyond-J-4 pair linked: %+v", links)
	}
}

func TestBuildIntermediateHops(t *testing.T) {
	// Two Important worlds 6 pc apart (beyond J-4) bridged by one lesser world
	// at the midpoint: no direct link, but a 2-hop route through the midpoint.
	a := sectorgen.Hex{Col: 1, Row: 1}
	mid := sectorgen.Hex{Col: 1, Row: 4}
	b := sectorgen.Hex{Col: 1, Row: 7}
	if d := a.Distance(b); d <= DefaultJump {
		t.Fatalf("fixture A-B distance = %d, want > %d", d, DefaultJump)
	}
	worlds := []World{{a, 4}, {mid, 0}, {b, 5}}
	links := Build(worlds, 0)
	if len(links) != 2 {
		t.Fatalf("got %d links, want 2 (A-mid, mid-B): %+v", len(links), links)
	}
	want := []Link{{From: a, To: mid, Jump: 3}, {From: mid, To: b, Jump: 3}}
	for i, w := range want {
		if links[i] != w {
			t.Errorf("link %d = %+v, want %+v", i, links[i], w)
		}
	}

	// With no stepping stone, the far Important pair stays unlinked.
	if links := Build([]World{{a, 4}, {b, 5}}, 0); len(links) != 0 {
		t.Errorf("unreachable Important pair linked: %+v", links)
	}
}

func TestExpectedTraffic(t *testing.T) {
	cases := map[int]int{7: 1000, 5: 1000, 4: 100, 3: 30, 2: 20, 1: 10, 0: 2, -1: 1, -2: 0, -5: 0}
	for ix, want := range cases {
		if got := ExpectedTraffic(ix); got != want {
			t.Errorf("ExpectedTraffic(%d) = %d, want %d", ix, got, want)
		}
	}
}

func TestBuildStableOrder(t *testing.T) {
	// Three mutually-close Important worlds -> three links in CCRR order,
	// regardless of input order.
	worlds := []World{
		{sectorgen.Hex{Col: 3, Row: 1}, 4},
		{sectorgen.Hex{Col: 1, Row: 1}, 4},
		{sectorgen.Hex{Col: 2, Row: 1}, 4},
	}
	links := Build(worlds, 0)
	if len(links) != 3 {
		t.Fatalf("got %d links, want 3: %+v", len(links), links)
	}
	if links[0].From != (sectorgen.Hex{Col: 1, Row: 1}) || links[2].From.Col < links[1].From.Col {
		t.Errorf("links not in CCRR order: %+v", links)
	}
}
