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

// otherWorldType picks a secondary world's type from a 1D roll and where its
// orbit sits relative to the habitable zone (Book 3 p.29). Worlds beyond HZ+1
// use the cold outer table; those inside (including the HZ band) use the inner
// table. A primary with no habitable zone (hasHZ false) uses the outer table.
func otherWorldType(orbit, hz int, hasHZ bool, oneD int) worldgen.OtherWorldType {
	if !hasHZ || orbit > hz+1 {
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
