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

// Config is a hull configuration (Book 2 p.71), which sets streamlining
// (friction), agility, stability, max acceleration, and whether the hull can
// land. Its QSP letter is one of C B P U S A L.
type Config int

const (
	Cluster Config = iota
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
func (c Config) Letter() byte {
	if c < Cluster || c > Lifting {
		return '?'
	}
	return configData[c].letter
}

func (c Config) String() string {
	if c < Cluster || c > Lifting {
		return "?"
	}
	return configData[c].name
}

// Structure is a hull's structural material (Book 2 p.52); it sets the base
// armor value and a cost multiplier. FramePlate is the default (zero value).
type Structure int

const (
	FramePlate Structure = iota // Frame-and-Plate (default)
	Shell                       // base AV = TL/2
	Polymer
	FeNi
	Organic // x0.5 cost
	Charged // x2 cost
)

var structureNames = [...]string{
	"Frame-and-Plate",
	"Shell",
	"Polymer",
	"FeNi",
	"Organic",
	"Charged",
}

func (s Structure) String() string {
	if s < FramePlate || s > Charged {
		return "?"
	}
	return structureNames[s]
}

// DriveKind is a drive type. Design covers the three core drives; Potential
// means thrust in Gs (Maneuver), jump number in parsecs (Jump), or the EP tier
// (Power Plant).
type DriveKind int

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
	Config      Config
	Structure   Structure
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
	Config             Config
	Structure          Structure
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
	Fuel       int // tons of fuel this drive demands
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
	Problems []string
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
		drives = append(drives, fmt.Sprintf("Power-%s", driveLabel(s.Power.Letter)))
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
