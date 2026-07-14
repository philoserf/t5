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
	st := stageData[stageIndex(spec.Stage)]

	var problems []string
	// Each weapon has a minimum mount (p.155): a Meson Gun does not fit in a
	// turret. Anything at or above the minimum may be selected.
	if spec.Mount < w.minMount {
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

	return Weapon{
		Spec: spec,
		// The stage and the range both shift the weapon's tech level.
		TL: w.tl + rng.tlMod + st.tlDelta,
		// The mount supplies the Mod and the dice; the weapon and the stage adjust.
		Mod:  m.mod + w.mod + weaponStageMod[stageIndex(spec.Stage)],
		Hits: m.hits + w.hitsDice,
		// Range effects apply to the mount, not the weapon (p.156) — so tonnage
		// is the mount's, scaled, and the weapon adds none of its own.
		Tons: Tons(m.tons) * Tonnage(rng.tonsNum) / Tonnage(rng.tonsDen),
		// The stage prices the weapon; the range prices the mount.
		Cost: w.cost*st.costNum/st.costDen + m.cost*rng.costNum/rng.costDen,

		Band: rng.band,
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

// Letter is the weapon's single-letter model code (Book 2 p.83 Table A).
func (w Weapon) Letter() byte {
	if !validWeapon(w.Spec.Model) {
		return '?'
	}
	return weaponData[w.Spec.Model].letter
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
	var b strings.Builder
	// Stage may be omitted when it is Standard (p.155), but the book's own tables
	// print it, so we always do.
	fmt.Fprintf(&b, "%s %s %s %s", w.Spec.Stage, rangeData[w.Spec.Range].name,
		mountData[w.Spec.Mount].name, w.Name())
	fmt.Fprintf(&b, " Mod=%+d. %s. %s. Hits= %dD. %s.",
		w.Mod, w.Tons.Phrase(), weaponMCr(w.Cost), w.Hits, w.RangeCode())
	return b.String()
}

func validWeapon(id WeaponID) bool { return id >= 0 && int(id) < len(weaponData) }
func validMount(m Mount) bool      { return m >= 0 && int(m) < len(mountData) }
func validRange(r Range) bool      { return r >= 0 && int(r) < len(rangeData) }

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
