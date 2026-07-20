package shipcombat

import (
	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/shipgen"
)

// The bridge between a designed ship and a fought one.
//
// The tasks in this package take primitives — a tech level, a Mod, an armor value,
// a compartment count — because that is what the book's tables are. But a caller
// had to invent every one of them: nothing produced a weapon's TL or its Mod, so a
// designed ship could not actually fight. shipgen now designs weapons, defenses,
// and rounds, and these functions feed them straight in.
//
// The two packages already agreed on more than they knew. HullLocations is keyed
// by the same 1..24 hull ordinal that shipgen.Hull.Letter carries, so a designed
// hull already knew its own compartments — nobody had ever connected them.

// Attack rolls a designed weapon's Space Weapon Task against a target at the given
// range (Book 2 p.195). The weapon supplies its tech level and its Mod — the one
// its mount, its tech stage, and the weapon itself add up to — and the gunner
// supplies C+S+K: their Characteristic, plus Gunner skill, plus the knowledge of
// this kind of weapon. Extra situational mods are added on top.
//
// A gunner whose C+S+K falls short of the weapon's tech level may use the weapon
// console's TL instead (p.156); pass whichever is better.
func Attack(r *dice.Roller, w shipgen.Weapon, rangeBands, csk int, mods ...int) dice.CheckResult {
	return ResolveSpaceWeapon(r, rangeBands, w.TL, csk, append([]int{w.Mod}, mods...)...)
}

// AttackWithMissile rolls a designed round's Missile Attack Task (Book 2 p.197),
// used when the target is at Range S=6 or more; at S=5 or less a missile uses the
// Space Weapon Task instead, so call Attack. The round's guidance supplies the
// second asset: a hardwired brain is worth a flat 5, an operator-guided one is
// worth as much as the gunner flying it, and a self-aware one is worth whatever
// mind it rolled at launch (C = 6+1D, S = 1D, plus Flux — pass it as brainCSK).
//
// The book states the boundary four times and disagrees with itself once: p.195
// ("Missiles at S=5 or Less … use the Space Weapon Task" / "Missiles at S=6 or
// More … use the Space Missile Task"), p.196, and p.197's own TO HIT (MISSILE)
// all put the split at 6, while p.197's opening line says "if Attack Range is
// S=5 or more". Resolved against that outlier by three-to-one.
func AttackWithMissile(
	r *dice.Roller,
	m shipgen.Missile,
	rangeBands, gunnerCSK, brainCSK int,
	mods ...int,
) dice.CheckResult {
	return ResolveMissile(
		r,
		rangeBands,
		m.TL,
		guidanceAsset(m.Spec.Guidance, gunnerCSK, brainCSK),
		mods...,
	)
}

// guidanceAsset is what a missile's brain is worth on the Missile Attack Task
// (Book 2 p.157). An unguided round contributes nothing, a hardwired one a flat 5,
// and a guided one is only as good as whoever — or whatever — is flying it: the
// gunner's own C+S+K, or, for a self-aware round, the mind it rolled at launch
// (C = 6+1D, S = 1D, plus Flux).
//
// The Guidance type is a design property and lives in shipgen; what it is worth in
// a fight is a target-number rule, and belongs here beside the other two.
func guidanceAsset(g shipgen.Guidance, gunnerCSK, brainCSK int) int {
	switch g {
	case shipgen.HardWired:
		return 5
	case shipgen.OperatorGuided, shipgen.DownLoaded:
		return gunnerCSK
	case shipgen.SelfAware:
		return brainCSK
	default: // UnGuided
		return 0
	}
}

// A MassiveExplosionMultiplier scales a missile option's effects when it hits via
// the Space Weapon Task at close range (Book 2 p.196 Massive Explosion, Weapons
// Task): a multiple of the target's Armor Value defeated, and of the Rad, Flash,
// and EMP effects. EMP applies only to nuclear options.
type MassiveExplosionMultiplier struct {
	AV, Rad, Flash, EMP int
}

// WeaponsMassiveExplosion returns the effect multipliers a designed round applies
// when it detonates through the Space Weapon Task (Book 2 p.196), and whether the
// round is one of the tabulated options. Most warheads are not: an Explosive or a
// Slug does its Penetration and no more.
//
// The book's table is these seven rows:
//
//	                        AV=  Rad  Flash  EMP
//	A AM Missile            10x         1x
//	N Missile Nuke Option   10x   1x    1x     1x
//	E Missile EMP Option     1x               1x
//	K KK Missile             1x
//	D DeadFall Size-4        1x         1x
//	D DeadFall Size-5        5x         2x
//	D DeadFall Size-6       10x         3x
//
// (The printed sheet letters the first row X; A is the Anti-Matter warhead's own
// letter on p.170, and every other row's letter is its warhead's, so X is a
// typo for A.) DeadFall is the one option the table splits by size, and it splits
// it exactly where p.170 does — the warhead exists at Size 4, 5, and 6, dividing
// the explosion by 20, 10, and 5 — so the heavier round defeats more armor and
// flashes brighter. That size is MissileSpec.Size; it was never a property of the
// warhead, which is why keying this table on a missile *type* alone could not
// express it.
//
// This takes a whole designed Missile rather than a (warhead, size) pair because
// the pair can be mismatched and the round cannot: DesignMissile has already
// checked that this warhead exists at this size, from this launcher.
//
// So a round that failed design has no tabulated explosion. Only the DeadFall arm
// reads Size, so without this check an undesignable Size-7 anti-matter round —
// which a caller can write as a Missile literal, as this package's own tests
// do — still reported a full 10x armour-defeating detonation while an equally
// undesignable Size-7 DeadFall correctly reported none. The two arms of one table
// disagreed about whether an untabulated round is tabulated.
func WeaponsMassiveExplosion(m shipgen.Missile) (MassiveExplosionMultiplier, bool) {
	if len(m.Problems) > 0 {
		return MassiveExplosionMultiplier{}, false
	}

	switch m.Spec.Type {
	case shipgen.AntiMatter:
		return MassiveExplosionMultiplier{AV: 10, Flash: 1}, true
	case shipgen.Nuke:
		return MassiveExplosionMultiplier{AV: 10, Rad: 1, Flash: 1, EMP: 1}, true
	case shipgen.EMP:
		return MassiveExplosionMultiplier{AV: 1, EMP: 1}, true
	case shipgen.Kinetic:
		return MassiveExplosionMultiplier{AV: 1}, true
	case shipgen.Deadfall:
		return deadfallMassiveExplosion(m.Spec.Size)
	default: // Slug, Explosive, Decoy, SensorPkg — no Massive Explosion
		return MassiveExplosionMultiplier{}, false
	}
}

// deadfallMassiveExplosion is the DeadFall size ladder of the p.196 table. A size
// the warhead does not come in is not tabulated.
func deadfallMassiveExplosion(size int) (MassiveExplosionMultiplier, bool) {
	switch size {
	case 4:
		return MassiveExplosionMultiplier{AV: 1, Flash: 1}, true
	case 5:
		return MassiveExplosionMultiplier{AV: 5, Flash: 2}, true
	case 6:
		return MassiveExplosionMultiplier{AV: 10, Flash: 3}, true
	default:
		return MassiveExplosionMultiplier{}, false
	}
}

// Defend rolls a designed defense against an incoming attack (Book 2 p.196): 1D
// under the defense's tech level, less the attacker's, plus the defense's mount
// Mod. Success deflects the attack — it does not hit.
//
// This is where the old Mount.Mod would have been wrong in the other direction: a
// defense in a Quad Turret defends at +4, where a weapon in one attacks at only
// +1. The Defense carries the Mod from the table it was actually built from.
//
// A weapon allocated to the Anti-Missile Defensive Fire mode (p.186) is a Defense
// by the time it reaches here — build it with shipgen.DesignWeaponAsDefense — so it
// defends through this same call.
// A defense the designer refused is not aboard: shipgen already charges the ship
// nothing for it, spends no tonnage on it, and renders it as a bare name. It must
// not then intercept anything, or a refused screen becomes a free one — the ship
// gets the protection precisely because it never paid for it.
func Defend(r *dice.Roller, d shipgen.Defense, attackTL int) dice.CheckResult {
	if !d.Installed() {
		// No hardware, no interception. Resolve against an impossible target so the
		// result reads as a clean miss rather than a special case for every caller.
		return ResolveDefensiveFire(r, 0, attackTL, noDefenseMod)
	}

	return ResolveDefensiveFire(r, d.TL, attackTL, d.Mod)
}

// noDefenseMod is the modifier for defensive fire that has nothing to fire with:
// large enough that no roll succeeds, so an absent defense never intercepts.
const noDefenseMod = -100

// ArmorLayers is a designed ship's armor as Penetrate wants it: one Armor Value
// per layer (Book 2 p.75). Damage grinds through the layers in turn, so the shape
// matters — four layers of AV-6 are not one layer of AV-24.
//
// Armor.AV is already the per-layer value: "identical layers are not summed on the
// record" (shipgen/fuelarmor.go). Dividing it by the layer count — which this once
// did — armored every multi-layer ship at a fraction of its real protection, and
// the more layers it bought the thinner each one became.
func ArmorLayers(s shipgen.Ship) []int {
	if s.Armor.Layers <= 0 {
		return nil
	}

	layers := make([]int, s.Armor.Layers)
	for i := range layers {
		layers[i] = s.Armor.AV
	}

	return layers
}

// Card is a designed ship's combat record: the compartments a hit is located in,
// and how much damage each can take. The hull ordinal keys the same p.95 Table H
// this package already carried — a designed hull has always known this about
// itself, and nothing asked.
func Card(s shipgen.Ship) (HullLocation, bool) {
	return HullLocations(s.Hull.Letter)
}

// ShipAgility is how much of a designed ship's thrust is left for dodging (Book 2
// p.200): its maneuver drive's rating, less the Gs it is spending on going
// somewhere.
//
// The thrust is capped by the hull's configuration, not just by the drive: a
// Cluster hull is rated for 1G however large a drive is bolted to it (Book 2 p.71).
// Design reports a drive that outruns its hull as a Problem; here we fly the ship
// the hull can actually fly.
func ShipAgility(s shipgen.Ship, usedGs int) int {
	if s.Maneuver == nil {
		return 0
	}

	return Agility(min(s.Maneuver.Potential, s.Hull.MaxG), usedGs)
}
