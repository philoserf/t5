package systemgen

import (
	"fmt"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/ehex"
)

// GGClass is a gas giant's placement class (Book 3 p.29): a Small or Large Gas
// Giant — split at size S (Jupiter) — or an Ice Giant, which every second Small
// Gas Giant in a system becomes. The class selects the P2 placement column.
type GGClass int

const (
	SmallGasGiant GGClass = iota // sizes L..R (Neptune/Saturn-class)
	LargeGasGiant                // sizes S..Y (Jupiter-class and up)
	IceGiant                     // a converted Small Gas Giant, same size
)

// String renders the class as its placement-chart abbreviation.
func (c GGClass) String() string {
	switch c {
	case LargeGasGiant:
		return "LGG"
	case IceGiant:
		return "IG"
	default:
		return "SGG"
	}
}

// A GasGiant is one gas giant in a system: its size, physical detail, and
// placement class (Book 3 p.29). Placement into a concrete orbit is handled by
// the system's rotate scheduler and is not set here.
type GasGiant struct {
	Size     int     // eHex size code 20..32 (L..Y)
	Diameter int     // equatorial diameter, miles
	SkimG    float64 // a ship's G must exceed this to skim fuel (p.29)
	Class    GGClass
}

// String renders the gas giant as its size letter and class, e.g. "T LGG".
func (g GasGiant) String() string {
	return fmt.Sprintf("%s %s", ehex.Format(g.Size), g.Class)
}

const (
	ggMinSize   = 20 // eHex L: the smallest gas giant (Book 3 p.29 GG table row 1)
	ggLargeSize = 26 // eHex S: Jupiter-class and up are Large Gas Giants
)

// ggDetail is each gas-giant size's diameter (miles) and skim gravity, indexed
// by size code minus ggMinSize (Book 3 p.29, sizes L..Y / 20..32).
var ggDetail = [...]struct {
	diameter int
	skimG    float64
}{
	{20000, 0.2}, {30000, 0.3}, {40000, 0.4}, {50000, 0.5}, {60000, 0.6},
	{70000, 0.7}, {80000, 0.8}, {90000, 0.9}, {125000, 1.2}, {180000, 1.8},
	{220000, 2.2}, {250000, 2.5}, {250000, 3.0},
}

// gasGiantDetail builds a gas giant from its size code (20..32): a pure lookup
// into the Book 3 p.29 detail table. Sizes below Jupiter (S / 26) are Small Gas
// Giants, the rest Large.
func gasGiantDetail(size int) GasGiant {
	d := ggDetail[size-ggMinSize]
	class := SmallGasGiant
	if size >= ggLargeSize {
		class = LargeGasGiant
	}
	return GasGiant{Size: size, Diameter: d.diameter, SkimG: d.skimG, Class: class}
}

// rollGasGiants rolls n gas giants for a system (Book 3 p.29): each a 2D size
// (yielding sizes M..X), then every second Small Gas Giant becomes an Ice Giant
// of the same size. They are returned in rolled order.
func rollGasGiants(r *dice.Roller, n int) []GasGiant {
	giants := make([]GasGiant, n)
	smalls := 0
	for i := range giants {
		g := gasGiantDetail(r.Dice(2) + ggMinSize - 1) // 2D -> size code 21..31
		if g.Class == SmallGasGiant {
			if smalls++; smalls%2 == 0 {
				g.Class = IceGiant
			}
		}
		giants[i] = g
	}
	return giants
}
