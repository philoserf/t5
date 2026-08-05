// Package shipgen designs Traveller5 Adventure Class Ships (Book 2, pp. 30-95).
// Ship design is deterministic: the designer chooses tonnage, tech level,
// configuration, structure, armor, and drive sizes, and Design computes the
// derived performance, fuel, and cost. The compact result is the Quick Ship
// Profile (QSP) — the ship equivalent of a world's UWP.
//
// This package covers the core design engine (hull, drives, the Drive Potential
// formula, fuel, armor, and the QSP). Weapons, sensors, defenses, and crew/
// accommodations are deferred.
package shipgen

import (
	"fmt"
	"strings"

	"github.com/philoserf/t5/internal/ehex"
)

// HullConfig is a hull configuration (Book 2 p.71), which sets streamlining
// (friction), agility, stability, max acceleration, and whether the hull can
// land. Its QSP letter is one of C B P U S A L.
type HullConfig int

// Hull configurations.
const (
	Cluster HullConfig = iota
	Braced
	Planetoid
	Unstreamlined
	Streamlined
	Airframe
	Lifting
)

var configData = [...]struct {
	letter byte
	name   string
}{
	Cluster:       {'C', "Cluster"},
	Braced:        {'B', "Braced"},
	Planetoid:     {'P', "Planetoid"},
	Unstreamlined: {'U', "Unstreamlined"},
	Streamlined:   {'S', "Streamlined"},
	Airframe:      {'A', "Airframe"},
	Lifting:       {'L', "Lifting Body"},
}

// Letter returns the config's QSP letter, or '?' if the value is out of range.
func (c HullConfig) Letter() byte {
	if c < Cluster || c > Lifting {
		return '?'
	}

	return configData[c].letter
}

func (c HullConfig) String() string {
	if c < Cluster || c > Lifting {
		return "?"
	}

	return configData[c].name
}

// HullMaterial is a hull's structural material (Book 2 p.52); it sets the base
// armor value and a cost multiplier. FramePlate is the default (zero value).
type HullMaterial int

// Hull structures.
const (
	FramePlate HullMaterial = iota // Frame-and-Plate (default)
	Shell                          // base AV = TL/2
	Polymer
	FeNi
	Organic // x0.5 cost
	Charged // x2 cost
)

// structureData is the single table backing String, StructureNames, and
// StructureByName, so the display name and the CLI's short lookup key cannot
// drift apart the way structureNames and generate.go's separate
// structureByName map once could — that map's own "frame-plate" entry could
// never be reached, since StructureByName squashes hyphens out of its input
// before the lookup.
var structureData = [...]struct {
	display string // rendered form, e.g. in a ship card
	cliName string // -structure flag's short form, e.g. "plate"
	alias   string // another accepted -structure spelling, squashed at lookup ("" if none)
}{
	FramePlate: {"Frame-and-Plate", "plate", "frameplate"},
	Shell:      {"Shell", "shell", ""},
	Polymer:    {"Polymer", "polymer", ""},
	FeNi:       {"FeNi", "feni", ""},
	Organic:    {"Organic", "organic", ""},
	Charged:    {"Charged", "charged", ""},
}

func (s HullMaterial) String() string {
	if s < FramePlate || s > Charged {
		return "?"
	}

	return structureData[s].display
}

// StructureNames returns the six structures' CLI short names in order
// (plate, shell, polymer, feni, organic, charged), for building flag help and
// error text that cannot drift from what StructureByName actually accepts.
func StructureNames() []string {
	names := make([]string, len(structureData))
	for i, d := range structureData {
		names[i] = d.cliName
	}

	return names
}

// DriveKind is a drive type. Design covers the three core drives; Potential
// means thrust in Gs (Maneuver), jump number in parsecs (Jump), or the EP tier
// (Power Plant).
type DriveKind int

// Drive kinds.
const (
	Maneuver DriveKind = iota
	Jump
	Power
)

var driveKindNames = [...]string{"Maneuver", "Jump", "Power Plant"}

func (k DriveKind) String() string {
	if k < Maneuver || k > Power {
		return "?"
	}

	return driveKindNames[k]
}

// Hull and drive size letters are the eHex letters (A-Z omitting I and O) used
// as an ordinal 1..24 — Hull A = 100t (ordinal 1), Hull Z = 2400t (ordinal 24),
// per the QSP Entries table (Book 2 p.93). This is distinct from the eHex
// numeric value of the letter (where A = 10).

const maxLetter = 24 // Z, the largest hull/drive size

// LetterOrdinal maps a hull/drive letter (A-Z omitting I/O, case-insensitive) to
// its ordinal 1..24, or 0 if the byte is not one of the eHex letters — the
// reverse of a hull/drive size code.
func LetterOrdinal(c byte) int {
	if c >= 'a' && c <= 'z' {
		c -= 'a' - 'A'
	}

	if i := strings.IndexByte(ehex.Alphabet, c); i >= 10 {
		return i - 9
	}

	return 0
}

// ordinalLetter returns the letter for an ordinal 1..24, or '?' out of range.
func ordinalLetter(n int) byte {
	if n < 1 || n > maxLetter {
		return '?'
	}

	return ehex.Digit(9 + n)
}

// HullTons is the nominal tonnage of a hull size ordinal: ordinal * 100 tons.
func HullTons(ordinal int) int { return ordinal * 100 }

// A DriveSpec is a requested drive: its size letter (as an ordinal 1..24, or an
// even 26..48 for an extended "letter2" size) and its TL stage (Standard is the
// zero value / baseline).
type DriveSpec struct {
	Letter int
	Stage  Stage
}

// MinTL and MaxTL bound the Tech Level a ship can be designed at (Book 2 p.51,
// "Tech Level Limits"): "Within the Imperium, the maximum shipyard Tech Level is
// 15. Within this ship design system, the maximum available Tech Level is 21."
// 21 is therefore the ceiling of the design system itself, not of the setting —
// a TL-15 Imperial shipyard is the in-universe limit, and the extra rungs exist
// for shipyards outside it. The floor is 0, the lowest eHex Tech Level; a
// negative TL is not a Tech Level at all, and rendering one produces a ship card
// whose TL field is not an eHex value ("TL--5").
//
// Design does not enforce these — it is total, and reports an infeasible design
// in Ship.Problems rather than refusing it. They are here for the callers that
// take a Tech Level from outside (a CLI flag, a config file) and must reject a
// value that is not a Tech Level before it reaches a record.
const (
	MinTL = 0
	MaxTL = 21
)

// A ShipSpec is the design input: the choices a naval architect makes. Tons of 0
// uses the hull letter's nominal tonnage; a non-zero Tons exercises the
// over/undertonnage rules. Maneuver/Jump/Power are nil when the ship lacks that
// drive.
type ShipSpec struct {
	Name        string // display name, e.g. "Murphy-class Scout"
	Mission     string // QSP mission code, e.g. "S"
	TL          int
	HullLetter  int // ordinal 1..24
	Tons        int // 0 = nominal (HullLetter*100)
	Config      HullConfig
	Structure   HullMaterial
	ArmorLayers int

	Maneuver *DriveSpec
	Jump     *DriveSpec
	Power    *DriveSpec

	Weapons  []WeaponSpec
	Defenses []DefenseSpec

	FuelScoop, FuelPurifier bool
}

// A Hull is a designed hull with its derived attributes.
type Hull struct {
	Letter, Tons       int
	Config             HullConfig
	Structure          HullMaterial
	Friction           int // as a divisor/multiplier code (Book 2 p.71)
	MaxG               int
	Agility, Stability int
	LandCapable        bool
	// Hardpoints is the hull's weapon mount capacity: one per 100 tons (Book 2
	// p.156). Each may instead be allocated as three FirmPoints, which carry
	// sub-ton mounts only — see mounts.go, which spends them.
	Hardpoints int // Tons / 100
	BaseArmor  int // = TL (Shell: TL/2)
	Cost       int // Cr
}

// A Drive is a designed drive with its performance and cost.
type Drive struct {
	Kind       DriveKind
	Letter     int
	Potential  int // thrust-G / jump-N / EP tier
	Stage      Stage
	Tons, Cost int // Cost in Cr
	// Fuel is the tons of fuel this drive demands, set in the fuel phase (it
	// needs the hull tonnage, which the drive does not know). A Maneuver drive
	// is always 0: its consumption is part of power-plant operations (Book 2
	// p.79).
	Fuel int
}

// Armor is a hull's armor: layer count, armor value, and tonnage. Armor imposes
// no separate monetary cost (Book 2 p.75).
type Armor struct {
	Layers, AV, Tons int
}

// Fuel is a ship's fuel tankage.
type Fuel struct {
	Tons, Cost int
}

// A Budget accounts hull tonnage: Used by components, and the Payload residual
// available for cargo/staterooms (negative means the design is over budget).
type Budget struct {
	Hull, Used, Payload int
}

// A Ship is a fully designed vessel: its spec plus every derived component.
// Problems is empty for a clean design and otherwise records feasibility
// findings (over budget, a TL-capped drive, an underpowered plant).
type Ship struct {
	Spec ShipSpec

	Hull     Hull
	Maneuver *Drive
	Jump     *Drive
	Power    *Drive
	Armor    Armor
	Fuel     Fuel
	Weapons  []Weapon
	Defenses []Defense

	Cost     int // total, Cr
	Tonnage  Budget
	Problems []Problem
}

// QSP renders the Quick Ship Profile (Book 2 p.93): Mission-HullConfig G Jump,
// e.g. "S-AL22" — Scout, Hull A, Lifting Body, 2G, Jump-2. A non-jump
// interstellar drive would append its prefix (deferred).
func (s Ship) QSP() string {
	g, jump := potential(s.Maneuver), potential(s.Jump)

	return fmt.Sprintf("%s-%c%c%s%s",
		s.Spec.Mission, ordinalLetter(s.Hull.Letter), s.Hull.Config.Letter(),
		ehex.Format(g), ehex.Format(jump))
}

func potential(d *Drive) int {
	if d == nil {
		return 0
	}

	return d.Potential
}

// String renders a full ship card.
func (s Ship) String() string {
	var b strings.Builder

	name := s.Spec.Name
	if name == "" {
		name = "Ship"
	}

	fmt.Fprintf(&b, "%s  %s\n", name, s.QSP())
	fmt.Fprintf(&b, "Hull:    %c %dt %s · TL-%d · %s\n",
		ordinalLetter(s.Hull.Letter), s.Hull.Tons, s.Hull.Config, s.Spec.TL, s.Hull.Structure)

	drives := make([]string, 0, 3)
	if s.Maneuver != nil {
		drives = append(
			drives,
			fmt.Sprintf("Maneuver-%s %dG", driveLabel(s.Maneuver.Letter), s.Maneuver.Potential),
		)
	}

	if s.Jump != nil {
		drives = append(
			drives,
			fmt.Sprintf("Jump-%s J-%d", driveLabel(s.Jump.Letter), s.Jump.Potential),
		)
	}

	if s.Power != nil {
		drives = append(drives, "Power-"+driveLabel(s.Power.Letter))
	}

	if len(drives) > 0 {
		fmt.Fprintf(&b, "Drives:  %s\n", strings.Join(drives, " · "))
	}

	if s.Armor.Layers > 0 {
		fmt.Fprintf(&b, "Armor:   %d layers AV-%d\n", s.Armor.Layers, s.Armor.AV)
	}

	for i, w := range s.Weapons {
		label := "Weapons:"
		if i > 0 {
			label = "        " // the rest of the battery lines up under the first
		}

		fmt.Fprintf(&b, "%s %s\n", label, w.LongName())
	}

	for i, d := range s.Defenses {
		label := "Defenses:"
		if i > 0 {
			label = "         "
		}

		fmt.Fprintf(&b, "%s %s\n", label, d.LongName())
	}

	fmt.Fprintf(
		&b,
		"Fuel:    %dt · Cost %s · Payload %dt",
		s.Fuel.Tons,
		mcr(s.Cost),
		s.Tonnage.Payload,
	)

	for _, p := range s.Problems {
		fmt.Fprintf(&b, "\n! %s", p)
	}

	return b.String()
}

// mcr renders a Cr amount as MegaCredits with one decimal, e.g. "MCr 70.3".
func mcr(cr int) string {
	return fmt.Sprintf("MCr %.1f", float64(cr)/1_000_000)
}
