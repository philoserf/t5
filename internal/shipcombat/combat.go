// Package shipcombat implements the Traveller5 Space Combat resolution engine
// (Book 2 pp. 193-204): the Space Weapon, Missile, and Defensive Fire tasks, hit
// location by Flux, penetration against layered armor, damage application, and
// the missile Massive Explosion. Each attack follows the five-step sequence —
// Declared, Defenses, To Hit, Penetration, Damage.
//
// The tasks are roll-low and resolve through internal/task; this package supplies
// the T5-specific dice counts and target numbers plus the penetration/damage
// tables. It works on primitives (TLs, mods, armor values, compartment numbers)
// rather than a full ShipCard, which the ship-design weapon/compartment model
// (deferred in shipgen) will later supply.
package shipcombat

import (
	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/task"
)

// A Mount holds a weapon; its Mod adds to a weapon's attack and to Defensive Fire
// (Book 2 pp. 195-196), roughly the number of tubes.
type Mount int

const (
	SingleTurret   Mount = iota // T1
	DualTurret                  // T2
	TripleTurret                // T3
	QuadTurret                  // T4
	SingleBarbette              // B1
	DualBarbette                // B2
)

var mountMod = [...]int{
	SingleTurret: 1, DualTurret: 2, TripleTurret: 3, QuadTurret: 4,
	SingleBarbette: 1, DualBarbette: 2,
}

// Mod is the mount's attack/defense modifier (Book 2 p.196): T1/B1 +1, T2/B2 +2,
// T3 +3, T4 +4.
func (m Mount) Mod() int {
	if int(m) < 0 || int(m) >= len(mountMod) {
		return 0
	}
	return mountMod[m]
}

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
func ResolveSpaceWeapon(r *dice.Roller, rangeBands, weaponTL, csk, mods int) dice.CheckResult {
	return task.ResolveDice(r, SpaceWeaponDice(rangeBands, weaponTL), SpaceWeaponTarget(weaponTL, csk, mods))
}

// Guidance is a missile's guidance option (Book 2 p.197); Value turns it into the
// value it adds to the Missile Attack Task target.
type Guidance int

const (
	UnGuided       Guidance = iota // 0
	HardWired                      // 5
	OperatorGuided                 // = gunner C+S+K
	SelfAware                      // rolled: C=6+1D, S=1D, +Flux (caller-computed)
	DownLoad                       // = gunner C+S+K
)

// Value is a guidance option's fixed contribution to the Missile Attack target
// (Book 2 p.197): UnGuided 0, HardWired 5, Operator-Guided and DownLoad the
// gunner's C+S+K. SelfAware has no fixed value — the missile rolls its own
// C=6+1D, S=1D, +Flux — so Value returns 0 for it and the caller passes the
// rolled value to ResolveMissile directly.
func (g Guidance) Value(gunnerCSK int) int {
	switch g {
	case HardWired:
		return 5
	case OperatorGuided, DownLoad:
		return gunnerCSK
	default:
		return 0
	}
}

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
func ResolveMissile(r *dice.Roller, rangeBands, missileTL, guidanceValue, mods int) dice.CheckResult {
	return task.ResolveDice(r, MissileDice(missileTL, rangeBands), MissileTarget(missileTL, guidanceValue, mods))
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
