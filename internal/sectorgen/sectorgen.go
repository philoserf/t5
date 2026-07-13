// Package sectorgen implements Traveller5's interstellar mapping (Book 3
// pp. 12-15): the sector/subsector hex geometry, CCRR coordinates, and the
// System Contents tables that populate a sector with star systems, gas giants,
// and asteroid-belt mainworlds at a chosen stellar density.
package sectorgen

import (
	"fmt"

	"github.com/philoserf/t5/internal/dice"
)

// Sector dimensions (Book 3 p.12): 32 columns of 40 rows = 1280 hexes; each
// subsector is 8 columns of 10 rows.
const (
	Columns      = 32
	Rows         = 40
	subsectorCol = 8
	subsectorRow = 10
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

// Subsector returns the hex's subsector letter A-P (Book 3 p.12): the sixteen
// 8x10 subsectors are lettered left-to-right, top-to-bottom in a 4x4 grid.
func (h Hex) Subsector() byte {
	colBand := (h.Col - 1) / subsectorCol
	rowBand := (h.Row - 1) / subsectorRow
	return byte('A' + rowBand*4 + colBand)
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

var densityNames = [...]string{
	"Extra Galactic", "Rift", "Sparse", "Scattered",
	"Standard", "Dense", "Cluster", "Core",
}

func (d Density) String() string {
	if d < ExtraGalactic || d > Core {
		return "?"
	}
	return densityNames[d]
}

// presenceRoll is each density's system-presence check: roll `dice` D6 and a
// system is present at or under `threshold` (Book 3 p.13). Core is transcribed
// as the printed "11 or less on 2D" (its ~91% note is approximate).
var presenceRoll = [...]struct{ dice, threshold int }{
	ExtraGalactic: {3, 3},
	Rift:          {2, 2},
	Sparse:        {1, 1},
	Scattered:     {1, 2},
	Standard:      {1, 3},
	Dense:         {1, 4},
	Cluster:       {1, 5},
	Core:          {2, 11},
}

// SystemPresent rolls whether a hex holds a star system at the given density
// (Book 3 p.13).
func SystemPresent(r *dice.Roller, d Density) bool {
	roll := presenceRoll[d]
	return r.Dice(roll.dice) <= roll.threshold
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
	return generate(r, d, 1, Columns, 1, Rows)
}

// GenerateSubsector rolls the 80 hexes of one subsector — letter 'A'-'P' — at
// the given density (Book 3 p.12). An out-of-range letter yields no systems.
func GenerateSubsector(r *dice.Roller, d Density, letter byte) []StellarHex {
	if letter < 'A' || letter > 'P' {
		return nil
	}
	i := int(letter - 'A')
	colBand, rowBand := i%4, i/4
	c0 := colBand*subsectorCol + 1
	r0 := rowBand*subsectorRow + 1
	return generate(r, d, c0, c0+subsectorCol-1, r0, r0+subsectorRow-1)
}

// generate rolls every hex in the inclusive column/row range and returns the
// populated ones in column-major CCRR order.
func generate(r *dice.Roller, d Density, colLo, colHi, rowLo, rowHi int) []StellarHex {
	var systems []StellarHex
	for col := colLo; col <= colHi; col++ {
		for row := rowLo; row <= rowHi; row++ {
			if SystemPresent(r, d) {
				systems = append(systems, rollContents(r, Hex{col, row}))
			}
		}
	}
	return systems
}
