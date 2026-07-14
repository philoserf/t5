package shipgen

import (
	"fmt"
	"strconv"
	"strings"
)

// Weapon design (Book 2 pp.83, 154-174). Like the rest of shipgen this is
// deterministic, not rolled: the designer picks a weapon model, a mount, a tech
// stage, and a range, and DesignWeapon computes what that installation weighs,
// costs, hits for, and hits with.
//
// The three inputs pull in different directions, which is the whole design game.
// The MOUNT supplies the tonnage, the damage dice, and the attack Mod — a bigger
// mount hits more often and harder, but a Main weapon is 200 tons. The RANGE
// scales the mount: reaching further multiplies its tonnage and cost and raises
// the tech level, while building for a knife fight divides them. The STAGE
// multiplies the weapon's cost for a tech level shift and an attack Mod.
//
// A weapon itself has no tonnage at all (p.156: "Base tonnage for a Weapon is
// zero tons") — a Beam Laser is the thing in the turret, and the turret is what
// takes up room.

// A WeaponSpec is a weapon installation as the designer chooses it: which weapon,
// in which mount, at which tech stage, built for which range. The zero value is
// a Standard-stage weapon, which is what most ships carry.
type WeaponSpec struct {
	Model WeaponID
	Mount Mount
	Stage Stage
	Range Range
}

// A Tonnage is a weight in hundredths of a ton. A mount built for a short range
// divides its tonnage (a turret at Boarding range is a quarter of a ton), and
// FirmPoints exist precisely to carry mounts "less than one ton" (Book 2 p.83),
// so whole tons would round a real installation away to nothing.
type Tonnage int

// Tons builds a Tonnage from whole tons.
func Tons(t int) Tonnage { return Tonnage(t * 100) }

// Ceil is the tonnage in whole tons, rounded up — a mount on a HardPoint "is at
// least 1 ton (round up)" (Book 2 p.83).
func (t Tonnage) Ceil() int { return (int(t) + 99) / 100 }

// SubTon reports whether the tonnage is under one ton, which is what a FirmPoint
// takes (and what a HardPoint would waste).
func (t Tonnage) SubTon() bool { return t < 100 }

// RoundUp is the tonnage rounded up to a whole ton, still a Tonnage — what a mount
// on a HardPoint is charged.
func (t Tonnage) RoundUp() Tonnage { return Tons(t.Ceil()) }

// String renders the tonnage as the book's tables do — "2", "0.25", "0.33".
func (t Tonnage) String() string {
	return strconv.FormatFloat(float64(t)/100, 'f', -1, 64)
}

// Phrase renders the tonnage with its unit, as the weapon tables write it:
// "1 ton", "2 tons", "0.25 tons".
func (t Tonnage) Phrase() string {
	if t == 100 {
		return "1 ton"
	}
	return t.String() + " tons"
}

// DefaultWeapon is a weapon as the book lists it: in its own minimum mount, at
// the standard rung of its range ladder, at Standard stage. It is the sensible
// starting point a designer then varies.
func DefaultWeapon(model WeaponID) WeaponSpec {
	if !validWeapon(model) {
		return WeaponSpec{Model: model}
	}
	w := weaponData[model]
	rng := AttackRange // the standard space range
	if w.scale == WorldScale {
		rng = VDistant // ...and the standard world one
	}
	return WeaponSpec{Model: model, Mount: w.minMount, Range: rng}
}

// A Weapon is a designed weapon installation: the spec plus everything derived
// from it. TL is the installed weapon's tech level (its base, shifted by the
// range and the stage), Mod is the total modifier on the Space Weapon Task, and
// Hits is the damage in dice. Problems reports an installation the book does not
// allow — Design folds these into the ship's own Problems rather than failing.
type Weapon struct {
	Spec WeaponSpec

	TL       int
	Mod      int
	Hits     int // damage dice
	Tons     Tonnage
	Cost     int // Cr
	Band     int // the S= or R= range band
	Scale    Scale
	Problems []string
}

// DesignWeapon computes a weapon installation from its spec (Book 2 p.83). Like
// Design, it is total: it never errors, and an installation the book disallows —
// a mount smaller than the weapon's minimum, or a range on the wrong scale — is
// reported in Problems rather than refused, so a caller always gets a weapon to
// look at.
func DesignWeapon(spec WeaponSpec) Weapon {
	if !validWeapon(spec.Model) || !validMount(spec.Mount) || !validRange(spec.Range) {
		return Weapon{Spec: spec, Problems: []string{"unknown weapon, mount, or range"}}
	}
	w := weaponData[spec.Model]
	m := mountData[spec.Mount]
	rng := rangeData[spec.Range]

	var problems []string
	// The Bolt-In is a defense's mount: a weapon has to see out of the hull.
	if !m.weaponOK {
		problems = append(problems, fmt.Sprintf("a weapon cannot be installed in a %s", m.name))
	}
	// Each weapon has a minimum mount (p.155): a Meson Gun does not fit in a
	// turret. Anything at or above the minimum may be selected.
	if m.weaponOK && spec.Mount < w.minMount {
		problems = append(problems, fmt.Sprintf("%s needs at least a %s, not a %s",
			w.name, mountData[w.minMount].name, m.name))
	}
	// A weapon reaches on one ladder or the other, so a World range cannot be
	// asked of a Space weapon (p.83 Tables D and E). The two dual-scale weapons
	// take either.
	if !w.scale.reaches(rng.scale) {
		problems = append(problems, fmt.Sprintf("%s is a %s weapon and cannot be built for %s",
			w.name, scaleName(w.scale), rng.name))
	}

	tl, tons, cost, band := install(w.tl, w.cost, m.tons, m.cost, spec.Range, spec.Stage)
	return Weapon{
		Spec: spec,
		TL:   tl,
		// The mount supplies the Mod and the dice; the weapon and the stage adjust.
		Mod:  m.mod + w.mod + weaponStageMod[stageIndex(spec.Stage)],
		Hits: m.hits + w.hitsDice,
		Tons: tons,
		Cost: cost,
		Band: band,
		// The range decides which ladder this installation reaches on, which is
		// what a dual-scale weapon's choice of range settles.
		Scale:    rng.scale,
		Problems: problems,
	}
}

// Name is the weapon's model name with its installed tech level, e.g.
// "Beam Laser-11" — how the book names a weapon throughout.
func (w Weapon) Name() string {
	if !validWeapon(w.Spec.Model) {
		return "?"
	}
	return fmt.Sprintf("%s-%d", weaponData[w.Spec.Model].name, w.TL)
}

// RangeCode renders the weapon's range band as the book writes it, e.g. "R=08"
// for a world-range weapon or "S=07" for a space-range one.
func (w Weapon) RangeCode() string {
	scale := 'S'
	if w.Scale == WorldScale {
		scale = 'R'
	}
	return fmt.Sprintf("%c=%02d", scale, w.Band)
}

// LongName renders the weapon's full identity, the way the book's own weapon
// tables do (Book 2 p.155 "Identifying Weapons"):
//
//	Standard Orbit Single Turret Beam Laser-11 Mod=0. 2 tons. MCr1.1. Hits= 1D. R=08.
//
// This is to a weapon what the UWP is to a world: everything needed to define
// its usage, in one line.
func (w Weapon) LongName() string {
	if !validWeapon(w.Spec.Model) || !validMount(w.Spec.Mount) || !validRange(w.Spec.Range) {
		return "?"
	}
	// Stage may be omitted when it is Standard (p.155), but the book's own tables
	// print it, so we always do.
	return fmt.Sprintf("%s %s %s %s Mod=%+d. %s. %s. Hits= %dD. %s.",
		w.Spec.Stage, rangeData[w.Spec.Range].name, mountName(w.Spec.Mount), w.Name(),
		w.Mod, w.Tons.Phrase(), weaponMCr(w.Cost), w.Hits, w.RangeCode())
}

// WeaponByName finds a weapon by its name or its single-letter model code, both
// case-insensitively and ignoring spaces ("beam laser", "beamlaser", "L").
func WeaponByName(s string) (WeaponID, bool) {
	if len(s) == 1 {
		for id, w := range weaponData {
			if upper(s[0]) == w.letter {
				return WeaponID(id), true
			}
		}
	}
	norm := squash(s)
	for id, w := range weaponData {
		if squash(w.name) == norm {
			return WeaponID(id), true
		}
	}
	return 0, false
}

// MountByCode finds a mount by its code ("T1", "Bay", "Bo") or its name ("single
// turret", "bolt-in"), case-insensitively. The Bolt-In is a defense mount only —
// DesignWeapon rejects a weapon asked to sit in one.
func MountByCode(s string) (Mount, bool) {
	norm := squash(s)
	for m, d := range mountData {
		if squash(d.code) == norm || squash(d.name) == norm {
			return Mount(m), true
		}
	}
	return 0, false
}

// RangeByName finds a range by name ("orbit", "attack range", "vdistant"),
// case-insensitively.
func RangeByName(s string) (Range, bool) {
	norm := squash(s)
	for r, d := range rangeData {
		if squash(d.name) == norm {
			return Range(r), true
		}
	}
	return 0, false
}

// WeaponNames lists every weapon's name, for a "known weapons" message.
func WeaponNames() []string {
	names := make([]string, len(weaponData))
	for i, w := range weaponData {
		names[i] = w.name
	}
	return names
}

// MountCodes lists every mount's code, for a "known mounts" message.
func MountCodes() []string {
	codes := make([]string, len(mountData))
	for i, m := range mountData {
		codes[i] = m.code
	}
	return codes
}

// RangeNames lists every range's name, for a "known ranges" message.
func RangeNames() []string {
	names := make([]string, len(rangeData))
	for i, r := range rangeData {
		names[i] = r.name
	}
	return names
}

// squash lowercases a name and drops its spaces, so "Beam Laser" and "beamlaser"
// are the same lookup key.
func squash(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", ""))
}

func upper(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c - ('a' - 'A')
	}
	return c
}

func validWeapon(id WeaponID) bool { return id >= 0 && int(id) < len(weaponData) }
func validMount(m Mount) bool      { return m >= 0 && int(m) < len(mountData) }
func validRange(r Range) bool      { return r >= 0 && int(r) < len(rangeData) }

// mountName is the mount's name, for every renderer that prints one.
func mountName(m Mount) string {
	if !validMount(m) {
		return "?"
	}
	return mountData[m].name
}

// install applies the scaling rule that weapons and defenses share (Book 2 p.156):
// the stage shifts the device's tech level and prices the device, while the range
// shifts the tech level again and scales the MOUNT's tonnage and cost. "Range
// Effects apply to the Mount but not the Weapon"; the device brings no tonnage of
// its own.
//
// It lives in one place because it is one rule. The two builders differ in what
// they do with the Mod, the damage dice, and what they refuse to build — not in
// this arithmetic.
func install(deviceTL, deviceCost, mountTons, mountCost int, rng Range, stage Stage) (tl int, tons Tonnage, cost, band int) {
	r := rangeData[rng]
	st := stageCostData[stageIndex(stage)]
	return deviceTL + r.tlMod + stageTL(stage),
		Tonnage(mountTons * r.tons),
		deviceCost*st.num/st.den + mountCost*r.cost/100,
		r.band
}

// stageTL is the tech level a stage shifts a component by. Weapons, defenses, and
// drives share it — it is the one column of the two stage tables that agrees
// everywhere (see weaponStageData).
func stageTL(s Stage) int { return stageData[stageIndex(s)].tlDelta }

// stageIndex bounds a Stage to the tables, so an out-of-range one reads as
// Standard rather than panicking.
func stageIndex(s Stage) Stage {
	if s < 0 || int(s) >= len(stageData) {
		return Standard
	}
	return s
}

func scaleName(s Scale) string {
	if s == WorldScale {
		return "world-range"
	}
	return "space-range"
}

// weaponMCr renders a Cr amount as the weapon tables write it: "MCr1.1", "MCr25"
// — no space, and no trailing zeros.
func weaponMCr(cr int) string {
	s := strconv.FormatFloat(float64(cr)/1_000_000, 'f', -1, 64)
	return "MCr" + s
}
