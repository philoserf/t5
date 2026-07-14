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
// used when the target is at Range S=5 or more. The round's guidance supplies the
// second asset: a hardwired brain is worth a flat 5, an operator-guided one is
// worth as much as the gunner flying it, and a self-aware one is worth whatever
// mind it rolled at launch (C = 6+1D, S = 1D, plus Flux — pass it as brainCSK).
func AttackWithMissile(r *dice.Roller, m shipgen.Missile, rangeBands, gunnerCSK, brainCSK int, mods ...int) dice.CheckResult {
	return ResolveMissile(r, rangeBands, m.TL, guidanceAsset(m.Spec.Guidance, gunnerCSK, brainCSK), mods...)
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
func Defend(r *dice.Roller, d shipgen.Defense, attackTL int) dice.CheckResult {
	return ResolveDefensiveFire(r, d.TL, attackTL, d.Mod)
}

// ArmorLayers is a designed ship's armor as Penetrate wants it: one Armor Value
// per layer (Book 2 p.86). Damage grinds through the layers in turn, so the shape
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
// and how much damage each can take. The hull ordinal keys the same p.86 Table H
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
