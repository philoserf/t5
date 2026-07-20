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
// subsector is 8 columns of 10 rows, so four subsectors span a row band.
//
// The two horizontal constants measure different things and are named for their
// units: subsectorWidth is hexes per subsector, subsectorsPerBand is subsectors
// per row band. Swapping the two would still divide the grid and still compile.
const (
	Columns           = 32
	Rows              = 40
	subsectorWidth    = 8                        // hex columns in one subsector
	subsectorHeight   = 10                       // hex rows in one subsector
	subsectorsPerBand = Columns / subsectorWidth // 4 subsectors across a row band
)

// A Hex is a location on a sector map: column 1-32, row 1-40 (Book 3 p.12). The
// upper-left hex is 0101, the lower-right 3240.
type Hex struct {
	Col, Row int
}

// Valid reports whether the hex is on the sector map: column 1-Columns, row
// 1-Rows. ParseHex and GenerateSector only ever build valid hexes, so this
// matters for a Hex a caller composed itself.
func (h Hex) Valid() bool {
	return h.Col >= 1 && h.Col <= Columns && h.Row >= 1 && h.Row <= Rows
}

// String renders the hex as its four-digit CCRR location, e.g. "0803". An off-map
// hex renders "????" rather than a location: this is the display path, so it stays
// total like ehex.Format and Density.String, and a four-character sentinel keeps
// the width of the positional record it prints into — while "3301" or "10005"
// would read as a real (or unparseable) location.
func (h Hex) String() string {
	if !h.Valid() {
		return "????"
	}

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

	// Ask Valid rather than restating its bounds — this is what makes its doc's
	// claim ("ParseHex and GenerateSector only ever build valid hexes") true by
	// construction rather than by two copies of the map's dimensions agreeing.
	h := Hex{Col: col, Row: row}
	if !h.Valid() {
		return Hex{}, false
	}

	return h, true
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

// ParseSubsector reads a subsector letter A-P, case-insensitively, reporting
// whether it is well-formed. Like ParseHex, it rejects anything longer than the
// name it expects, so a mistyped argument is caught rather than truncated.
func ParseSubsector(s string) (byte, bool) {
	if len(s) != 1 {
		return 0, false
	}

	letter := strings.ToUpper(s)[0]
	if letter < 'A' || letter > 'P' {
		return 0, false
	}

	return letter, true
}

// Subsector returns the hex's subsector letter A-P (Book 3 p.12): the sixteen
// 8x10 subsectors are lettered left-to-right, top-to-bottom in a 4x4 grid.
//
// It panics on an off-map hex rather than returning a letter, the strict half of
// the same split task.Difficulty.Dice draws: this letter is a filing decision
// (survey groups its records by it), and the arithmetic is happy to hand back a
// plausible wrong answer — Hex{33,1} bleeds into the next row band as 'E' and
// Hex{0,0} truncates to 'A', indistinguishable from the real upper-left
// subsector. A misfiled world would then be silent. Use Valid to pre-check a hex
// a caller composed itself; ParseHex and GenerateSector only build valid ones.
func (h Hex) Subsector() byte {
	if !h.Valid() {
		panic(fmt.Sprintf("sectorgen: hex %d,%d is off the sector map (1..%d, 1..%d)",
			h.Col, h.Row, Columns, Rows))
	}

	colBand := (h.Col - 1) / subsectorWidth
	rowBand := (h.Row - 1) / subsectorHeight

	return byte('A' + rowBand*subsectorsPerBand + colBand) //nolint:gosec // G115: value is a bounded 0-33 eHex digit
}

// Density is a region's stellar density (Book 3 p.13 Extended System Contents),
// which sets how likely a hex holds a star system.
type Density int

// System-presence densities (Book 3 p. 12).
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
// (Book 3 p.13). An out-of-range density yields no system rather than a lookup
// past the end of densityInfo.
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

// DeriveHex returns the substream one hex is generated from: the child of the
// sector's Roller keyed on the hex's coordinates. It is the single point where
// sectorgen and survey agree on a hex's stream, so a hex's system depends only
// on (sector seed, col, row) — not on any hex rolled before it. That is what
// makes a hex regenerable in isolation and a rule fix affecting one hex's draws
// unable to shift any other. Pass the sector Roller, not an already-derived one.
func DeriveHex(r *dice.Roller, h Hex) *dice.Roller {
	//nolint:gosec // G115: Col/Row are small positive coordinates; Derive treats the bits as an opaque discriminator.
	return r.Derive(uint64(h.Col), uint64(h.Row))
}

// RollHex rolls one hex of a sector from its own substream: whether a system is
// present at the given density and, if so, its coarse contents. It takes the
// sector's Roller and derives the hex's child itself (see DeriveHex), so the
// hex's coarse map depends only on (seed, col, row) and there is no way to pass
// the wrong Roller.
func RollHex(r *dice.Roller, d Density, h Hex) (StellarHex, bool) {
	child := DeriveHex(r, h)
	if !SystemPresent(child, d) {
		return StellarHex{}, false
	}

	return rollContents(child, h), true
}

// GenerateSector rolls all 1280 hexes of a sector at the given density and
// returns the populated ones in column-major CCRR order (Book 3 pp.12-13). Each
// hex is rolled from its own substream (see DeriveHex), so it requires a seeded
// Roller (New/NewWithSeed) and returns nil for a density outside the table.
//
// A hex's coarse contents depend only on the seed and its coordinates, so a
// sub-region rolls the same systems at the same coordinates as the whole sector
// does — but its printed record still depends on the region-wide passes (see
// survey.Survey.Subsector), so a subsector is selected from a surveyed sector
// rather than surveyed on its own.
func GenerateSector(r *dice.Roller, d Density) []StellarHex {
	if d < ExtraGalactic || d > Core {
		return nil
	}

	var systems []StellarHex

	for col := 1; col <= Columns; col++ {
		for row := 1; row <= Rows; row++ {
			h := Hex{col, row}
			if sh, ok := RollHex(r, d, h); ok {
				systems = append(systems, sh)
			}
		}
	}

	return systems
}
