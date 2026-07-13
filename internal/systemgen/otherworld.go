package systemgen

import (
	"github.com/philoserf/t5/internal/uwp"
	"github.com/philoserf/t5/internal/worldgen"
)

// An OtherWorld is a detailed secondary world occupying an orbit: its type, its
// generated UWP, and its trade codes (Book 3 p.29).
type OtherWorld struct {
	Type       worldgen.OtherWorldType
	Profile    uwp.Profile
	TradeCodes []string
}

// orbitZone classifies an orbit relative to a star's habitable zone (Book 3
// p.21): inner (HZ-2 or less), hospitable (HZ-1 to HZ+1), or outer (HZ+2 or
// greater). A star with no habitable zone treats every orbit as outer.
type orbitZone int

const (
	innerZone orbitZone = iota
	hospitableZone
	outerZone
)

func zoneOf(orbit, hz int, hasHZ bool) orbitZone {
	switch {
	case !hasHZ || orbit >= hz+2:
		return outerZone
	case orbit <= hz-2:
		return innerZone
	default:
		return hospitableZone
	}
}

// otherWorldType picks a secondary world's type from a 1D roll and where its
// orbit sits relative to the habitable zone (Book 3 p.29). Outer-zone worlds use
// the cold outer table; inner- and hospitable-zone worlds share the inner table.
func otherWorldType(orbit, hz int, hasHZ bool, oneD int) worldgen.OtherWorldType {
	if zoneOf(orbit, hz, hasHZ) == outerZone {
		switch oneD {
		case 1:
			return worldgen.Worldlet
		case 3:
			return worldgen.BigWorld
		case 5:
			return worldgen.RadWorld
		default: // 2, 4, 6
			return worldgen.Iceworld
		}
	}
	switch oneD {
	case 1:
		return worldgen.Inferno
	case 2:
		return worldgen.InnerWorld
	case 3:
		return worldgen.BigWorld
	case 4:
		return worldgen.StormWorld
	case 5:
		return worldgen.RadWorld
	default: // 6
		return worldgen.Hospitable
	}
}
