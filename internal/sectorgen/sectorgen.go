// Package sectorgen implements Traveller5's interstellar mapping (Book 3
// pp. 12-15): the sector/subsector hex geometry, CCRR coordinates, and the
// System Contents tables that populate a sector with star systems, gas giants,
// and asteroid-belt mainworlds at a chosen stellar density.
package sectorgen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/philoserf/t5/internal/dice"
)

// Sector dimensions (Book 3 p.12): 32 columns of 40 rows = 1280 hexes; each
// subsector is 8 columns of 10 rows, so subsectorCols subsectors span a row band.
const (
	Columns       = 32
	Rows          = 40
	subsectorCol  = 8
	subsectorRow  = 10
	subsectorCols = Columns / subsectorCol // 4 subsectors across a row band
)

// A Hex is a location on a sector map: column 1-32, row 1-40 (Book 3 p.12). The
// upper-left hex is 0101, the lower-right 3240.
type Hex struct {
	Col, Row int
}

// String renders the hex as its four-digit CCRR location, e.g. "0803".
func (h Hex) String() string {
	return fmt.Sprintf("%02d%02d", h.Col, h.Row)
}

// ParseHex reads a four-digit CCRR location (e.g. "0803") into a Hex, reporting
// whether it is well-formed and inside the sector. All four characters must be
// digits: strconv would accept a sign, so "+436" would otherwise silently parse
// as hex 0436.
func ParseHex(s string) (Hex, bool) {
	if len(s) != 4 {
		return Hex{}, false
	}
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return Hex{}, false
		}
	}
	col, _ := strconv.Atoi(s[:2])
	row, _ := strconv.Atoi(s[2:])
	if col < 1 || col > Columns || row < 1 || row > Rows {
		return Hex{}, false
	}
	return Hex{Col: col, Row: row}, true
}

// Distance returns the number of parsecs (jump distance) between two hexes on
// the Traveller map (Book 3 p.12). Traveller hexes are flat-topped and arranged
// in columns with even columns shifted half a hex down ("even-q" offset); the
// distance is the cube-coordinate distance after converting from that offset.
func (h Hex) Distance(o Hex) int {
	ax, az := h.Col, h.Row-(h.Col+(h.Col&1))/2
	bx, bz := o.Col, o.Row-(o.Col+(o.Col&1))/2
	ay, by := -ax-az, -bx-bz
	return (max(ax-bx, bx-ax) + max(ay-by, by-ay) + max(az-bz, bz-az)) / 2
}

// Subsector returns the hex's subsector letter A-P (Book 3 p.12): the sixteen
// 8x10 subsectors are lettered left-to-right, top-to-bottom in a 4x4 grid.
func (h Hex) Subsector() byte {
	colBand := (h.Col - 1) / subsectorCol
	rowBand := (h.Row - 1) / subsectorRow
	return byte('A' + rowBand*subsectorCols + colBand)
}

// Density is a region's stellar density (Book 3 p.13 Extended System Contents),
// which sets how likely a hex holds a star system.
type Density int

const (
	ExtraGalactic Density = iota // 3D <= 3 (~1%)
	Rift                         // 2D <= 2 (~3%)
	Sparse                       // 1D <= 1 (17%)
	Scattered                    // 1D <= 2 (33%)
	Standard                     // 1D <= 3 (50%)
	Dense                        // 1D <= 4 (66%)
	Cluster                      // 1D <= 5 (83%)
	Core                         // 2D <= 11
)

// densityInfo is the single table of each density's display name and its
// system-presence check: roll `dice` D6 and a system is present at or under
// `threshold` (Book 3 p.13). Core uses 2D <= 10, not the row's literal "11 or
// less": that text conflicts with every other Core figure the book prints — the
// stated 91% density, the Per-Sector 1170/1280 (91.4%), and the Count-Off [12]
// (~92%) all triangulate to 2D <= 10 (33/36 = 91.7%), while 2D <= 11 is 97.2%.
// The "11" reads as a typo for "10"; we follow the corroborated 91%.
var densityInfo = [...]struct {
	name            string
	dice, threshold int
}{
	ExtraGalactic: {"Extra Galactic", 3, 3},
	Rift:          {"Rift", 2, 2},
	Sparse:        {"Sparse", 1, 1},
	Scattered:     {"Scattered", 1, 2},
	Standard:      {"Standard", 1, 3},
	Dense:         {"Dense", 1, 4},
	Cluster:       {"Cluster", 1, 5},
	Core:          {"Core", 2, 10},
}

func (d Density) String() string {
	if d < ExtraGalactic || d > Core {
		return "?"
	}
	return densityInfo[d].name
}

// DensityNames returns the eight density names in order (Extra Galactic … Core).
func DensityNames() []string {
	names := make([]string, len(densityInfo))
	for i, info := range densityInfo {
		names[i] = info.name
	}
	return names
}

// DensityByName returns the Density matching a name case- and space-
// insensitively (e.g. "core", "extragalactic"), and whether it was found.
func DensityByName(name string) (Density, bool) {
	norm := normalizeName(name)
	for i, info := range densityInfo {
		if normalizeName(info.name) == norm {
			return Density(i), true
		}
	}
	return 0, false
}

func normalizeName(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", ""))
}

// SystemPresent rolls whether a hex holds a star system at the given density
// (Book 3 p.13). An out-of-range density yields no system, matching the letter
// guard in GenerateSubsector.
func SystemPresent(r *dice.Roller, d Density) bool {
	if d < ExtraGalactic || d > Core {
		return false
	}
	info := densityInfo[d]
	return r.Dice(info.dice) <= info.threshold
}

// A StellarHex is a populated hex: its location and the coarse system contents a
// long-range survey reports (Book 3 p.12).
type StellarHex struct {
	Hex               Hex
	GasGiant          bool // a gas giant is present (wilderness refueling)
	AsteroidMainworld bool // the mainworld is an asteroid belt (Size 0)
}

// rollContents rolls a present system's coarse contents (Book 3 p.13): a gas
// giant on 2D <= 8, and an asteroid-belt mainworld on 2D <= 2 (a natural 2).
func rollContents(r *dice.Roller, h Hex) StellarHex {
	return StellarHex{
		Hex:               h,
		GasGiant:          r.Dice(2) <= 8,
		AsteroidMainworld: r.Dice(2) <= 2,
	}
}

// GenerateSector rolls all 1280 hexes of a sector at the given density and
// returns the populated ones in column-major CCRR order (Book 3 pp.12-13).
func GenerateSector(r *dice.Roller, d Density) []StellarHex {
	return generate(r, d, Hex{1, 1}, Hex{Columns, Rows})
}

// GenerateSubsector rolls the 80 hexes of one subsector — letter 'A'-'P' — at
// the given density (Book 3 p.12). An out-of-range letter yields no systems.
func GenerateSubsector(r *dice.Roller, d Density, letter byte) []StellarHex {
	if letter < 'A' || letter > 'P' {
		return nil
	}
	i := int(letter - 'A')
	col := (i%subsectorCols)*subsectorCol + 1
	row := (i/subsectorCols)*subsectorRow + 1
	return generate(r, d, Hex{col, row}, Hex{col + subsectorCol - 1, row + subsectorRow - 1})
}

// generate rolls every hex in the inclusive rectangle from lo to hi and returns
// the populated ones in column-major CCRR order.
func generate(r *dice.Roller, d Density, lo, hi Hex) []StellarHex {
	var systems []StellarHex
	for col := lo.Col; col <= hi.Col; col++ {
		for row := lo.Row; row <= hi.Row; row++ {
			if SystemPresent(r, d) {
				systems = append(systems, rollContents(r, Hex{col, row}))
			}
		}
	}
	return systems
}
