package rangeband

import (
	"math"
	"testing"
)

func TestWorldBand(t *testing.T) {
	b, ok := WorldBand("3")
	if !ok || b.Descriptor != "Medium" || b.Meters != 150 {
		t.Errorf("WorldBand(3) = %+v,%v, want Medium/150", b, ok)
	}
	if _, ok := WorldBand("Z"); ok {
		t.Errorf("WorldBand(Z) should not exist")
	}
}

func TestSpaceBand(t *testing.T) {
	if b, _ := SpaceBand("B"); b.Descriptor != "Boarding" || b.Combat != "B" || b.Meters != 1000 {
		t.Errorf("SpaceBand(B) = %+v, want Boarding/B/1000", b)
	}
	if b, _ := SpaceBand("13"); b.Descriptor != "Outer System" || b.Meters != 1.5e12 {
		t.Errorf("SpaceBand(13) = %+v, want Outer System/1.5e12", b)
	}
}

func TestForDistance(t *testing.T) {
	// Exact representative distances land on their band.
	if b := WorldForDistance(150); b.Code != "3" {
		t.Errorf("WorldForDistance(150m) = %q, want 3", b.Code)
	}
	if b := WorldForDistance(5_000); b.Code != "6" {
		t.Errorf("WorldForDistance(5km) = %q, want 6", b.Code)
	}
	if b := WorldForDistance(0); b.Code != "0" {
		t.Errorf("WorldForDistance(0) = %q, want 0 (Contact)", b.Code)
	}
	if b := SpaceForDistance(5_000_000); b.Code != "4" {
		t.Errorf("SpaceForDistance(5000km) = %q, want 4 (Far Orbit)", b.Code)
	}
	if b := SpaceForDistance(1.5e12); b.Code != "13" {
		t.Errorf("SpaceForDistance(1.5bn km) = %q, want 13", b.Code)
	}
}

func TestBandNumber(t *testing.T) {
	b3, _ := WorldBand("3")
	if n, ok := b3.Number(); n != 3 || !ok {
		t.Errorf("WorldBand(3).Number() = %d,%v, want 3,true", n, ok)
	}
	s13, _ := SpaceBand("13")
	if n, ok := s13.Number(); n != 13 || !ok {
		t.Errorf("SpaceBand(13).Number() = %d,%v, want 13,true", n, ok)
	}
	// The lettered Contact/Boarding sub-bands have no number.
	for _, b := range []Band{{Code: "R"}, {Code: "T"}, {Code: "B"}} {
		if _, ok := b.Number(); ok {
			t.Errorf("Band(%q).Number() should report false", b.Code)
		}
	}
}

func TestConversion(t *testing.T) {
	world := map[string]string{"3": "0", "4": "0", "5": "B", "6": "1", "9": "4"}
	for r, wantS := range world {
		if s, ok := WorldToSpace(r); !ok || s != wantS {
			t.Errorf("WorldToSpace(%q) = %q,%v, want %q", r, s, ok, wantS)
		}
	}
	space := map[string]string{"0": "0", "B": "5", "4": "9", "13": "18"}
	for s, wantR := range space {
		if r, ok := SpaceToWorld(s); !ok || r != wantR {
			t.Errorf("SpaceToWorld(%q) = %q,%v, want %q", s, r, ok, wantR)
		}
	}
}

func TestWorldSubBand(t *testing.T) {
	// A representative distance lands exactly on its band.
	if got := WorldSubBand(150); got != 3 {
		t.Errorf("WorldSubBand(150m) = %g, want 3", got)
	}
	// Below R=1 clamps to 1, at/above R=9 clamps to 9.
	if got := WorldSubBand(1); got != 1 {
		t.Errorf("WorldSubBand(1m) = %g, want 1", got)
	}
	if got := WorldSubBand(1e9); got != 9 {
		t.Errorf("WorldSubBand(1e9) = %g, want 9", got)
	}
	// 8 km sits just into band 6 (the book's gas-giant 6.x sub-bands).
	if got := WorldSubBand(8_000); math.Abs(got-6.2) > 0.05 {
		t.Errorf("WorldSubBand(8km) = %g, want ~6.2", got)
	}
}
