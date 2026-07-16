// Package shipcombat implements the Traveller5 Space Combat resolution engine
// (Book 2 pp. 193-204): the Space Weapon, Missile, and Defensive Fire tasks, hit
// location by Flux, penetration against layered armor, damage application, and
// the missile Massive Explosion. Each attack follows the five-step sequence —
// Declared, Defenses, To Hit, Penetration, Damage.
//
// The tasks are roll-low and resolve through internal/task; this package supplies
// the T5-specific dice counts and target numbers plus the penetration/damage
// tables. They still take primitives — a tech level, a Mod, an armor value — since
// that is what the book's tables are, but nobody has to invent them any more:
// shipgen designs the weapons, defenses, and rounds, and ship.go feeds them in.
// Attack, Defend, and AttackWithMissile take designed components directly.
package shipcombat

import (
	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/task"
)

// A mount's Mod is not defined here. It was, and it was wrong: one Mount type
// carried one Mod table (T1 +1 ... T4 +4), used for both attacking and defending.
// The book has two. A weapon in a Single Turret attacks at Mod -2 (Book 2 p.83);
// a defense in one defends at +1 (p.174) — a bigger mount aims worse but defends
// better. The Mod now comes from the designed component, which knows which table
// it was built from: shipgen.Weapon.Mod and shipgen.Defense.Mod.

// SpaceWeaponDice is the Space Weapon Task's dice count (Book 2 p.195): one die
// per range band (minimum one), plus one more when the weapon's TL is below the
// range.
func SpaceWeaponDice(rangeBands, weaponTL int) int {
	n := max(rangeBands, 1)
	if weaponTL < rangeBands {
		n++
	}
	return n
}

// SpaceWeaponTarget is the Space Weapon Task's target number (Book 2 p.195):
// weapon TL + the gunner's C+S+K (or the weapon console's TL, whichever is used)
// + the weapon's mods.
func SpaceWeaponTarget(weaponTL, csk, mods int) int {
	return weaponTL + csk + mods
}

// ResolveSpaceWeapon rolls the Space Weapon Task (Book 2 p.195): (range-band)D
// roll-low against weaponTL + C+S+K + mods. Used by most weapons and by missiles
// at Range 5 or less.
func ResolveSpaceWeapon(
	r *dice.Roller,
	rangeBands, weaponTL, csk int,
	mods ...int,
) dice.CheckResult {
	return task.ResolveDice(
		r,
		SpaceWeaponDice(rangeBands, weaponTL),
		SpaceWeaponTarget(weaponTL, csk, 0),
		mods...)
}

// A missile's Guidance is a property of the round, chosen when it is designed, so
// the type lives with the design (shipgen.Guidance). What it is WORTH on the
// Missile Attack Task is combat's business, and lives here: guidanceAsset.

// MissileDice is the Missile Attack Task's dice count (Book 2 p.197): a constant
// 5D, plus one more when the missile TL is below the range.
func MissileDice(missileTL, rangeBands int) int {
	if missileTL < rangeBands {
		return 6
	}
	return 5
}

// MissileTarget is the Missile Attack Task's target number (Book 2 p.197):
// missile TL + the guidance value + mods.
func MissileTarget(missileTL, guidanceValue, mods int) int {
	return missileTL + guidanceValue + mods
}

// ResolveMissile rolls the Missile Attack Task (Book 2 p.197): 5D (6D if the
// missile TL is below the range) roll-low against missileTL + guidance + mods.
// Used for missiles at Range 6 or more.
func ResolveMissile(
	r *dice.Roller,
	rangeBands, missileTL, guidanceValue int,
	mods ...int,
) dice.CheckResult {
	return task.ResolveDice(
		r,
		MissileDice(missileTL, rangeBands),
		MissileTarget(missileTL, guidanceValue, 0),
		mods...)
}

// DefensiveFireTarget is the Defensive Fire Task's target number (Book 2 p.196):
// the defending weapon's TL minus the attacking weapon's TL plus the defending
// mount's Mod. A non-positive target cannot succeed on 1D — the defense has no
// effect.
func DefensiveFireTarget(defenseTL, attackTL, mountMod int) int {
	return defenseTL - attackTL + mountMod
}

// ResolveDefensiveFire rolls the Defensive Fire Task (Book 2 p.196): 1D roll-low
// against defenseTL - attackTL + mountMod. Success deflects the attack.
func ResolveDefensiveFire(r *dice.Roller, defenseTL, attackTL, mountMod int) dice.CheckResult {
	return task.ResolveDice(r, 1, DefensiveFireTarget(defenseTL, attackTL, mountMod))
}
