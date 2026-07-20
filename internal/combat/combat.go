// Package combat implements Traveller5 personal combat (Book 1 pp. 200-227): the
// Move-Attack-Damage round resolved as roll-low Tasks over range bands. It
// provides the three attack Numbers, the three attack tasks (Ranged, Melee,
// Impact), and damage resolution against armor and characteristics.
package combat

import (
	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/task"
)

// ShootingNumber is the pre-calculated Ranged-attack asset: Dexterity plus the
// weapon skill and its knowledge (Book 1 p.201).
func ShootingNumber(dex, weaponSkill, knowledge int) int { return dex + weaponSkill + knowledge }

// MeleeNumber is the pre-calculated Melee-attack asset: Strength plus the weapon
// skill and its knowledge (Book 1 p.201).
func MeleeNumber(str, weaponSkill, knowledge int) int { return str + weaponSkill + knowledge }

// ImpactNumber is a combatant's Impact asset: its current Speed (Book 1 p.201).
func ImpactNumber(speed int) int { return speed }

// Stance adjusts a target's apparent Size (Book 1 p.203).
type Stance int

// Combat stances.
const (
	Standing  Stance = iota // no change
	Crouching               // Size -1
	Prone                   // Size -2
	Evading                 // Size -1
)

// stanceSizeMods is each stance's adjustment to a target's apparent Size.
var stanceSizeMods = [...]int{Standing: 0, Crouching: -1, Prone: -2, Evading: -1}

func (s Stance) sizeMod() int {
	if s < Standing || s > Evading {
		return 0
	}

	return stanceSizeMods[s]
}

// TargetSize is a target's apparent size for an attack: object Size minus Range,
// reduced by the target's stance (Book 1 p.203). A value below zero means the
// target cannot ordinarily be seen or attacked.
func TargetSize(size, rng int, stance Stance) int {
	return size - rng + stance.sizeMod()
}

// Ranged resolves a Ranged Attack (Book 1 p.204): it rolls R=Range dice (a Range
// below 1 counts as 1D; a dice count greater than the weapon skill adds +1D for
// the "This Is Hard!" rule) at or under the Shooting Number + target size + mods.
// Success is a hit. The targetSize argument is the target's apparent size —
// TargetSize(size, rng, stance), i.e. Size minus Range adjusted by the stance mod
// (Crouch -1, Prone -2, Evading -1). A value below zero means the target cannot
// be seen and cannot be attacked, and the attack automatically fails (Book 1
// p.203).
//
// TIH! is stated in dice, not range: "If a task requires more dice than the
// character has applicable skill levels, then increase the difficulty one level"
// (p.128), and it is switched off when "Skill plus JOT is equal to or greater
// than the number of dice being rolled on a task". So the comparison is against
// the floored dice count, not the raw Range. At Range 0 the two differ: the
// floor is there because "there is always some slight chance, even at such very
// close range, that an unskilled attacker may miss" (p.204), and comparing the
// raw Range would make 0 > 0 false and delete exactly that case.
func Ranged(
	r *dice.Roller,
	shootingNumber, weaponSkill, targetSize, rng int,
	mods ...int,
) dice.CheckResult {
	if targetSize < 0 {
		return dice.CheckResult{} // apparent size below zero: the target cannot be seen or attacked
	}

	nd := max(rng, 1)
	if nd > weaponSkill {
		nd++ // This Is Hard!: the roll needs more dice than the attacker has skill
	}

	return task.ResolveDice(r, nd, shootingNumber+targetSize, mods...)
}

// Melee resolves a Melee Attack (Book 1 p.203): 2D at or under the Attacker
// Melee Number minus the Defender Melee Number, plus mods (spent Dexterity
// points, Evasion -1). Success is a hit; the roles then reverse for the reply.
func Melee(r *dice.Roller, attackerMN, defenderMN int, mods ...int) dice.CheckResult {
	return task.ResolveDice(r, task.Average.Dice(), attackerMN-defenderMN, mods...)
}

// Impact resolves an Impact Attack — a collision, ram, or fall (Book 1 p.202).
// It is the defender's avoidance roll: 2D at or under the defender's C2 (Dex,
// Agility, or Grace) minus the attacker's Speed. On Success the target dodges;
// when the check fails the target is hit and takes ImpactDamageDice(speed) dice.
// (The book's prose says "success is a hit", but its own worked example rolls
// low to dodge — the example governs.)
func Impact(r *dice.Roller, defenderC2, attackerSpeed int) dice.CheckResult {
	return task.ResolveDice(r, task.Average.Dice(), defenderC2-attackerSpeed)
}

// ImpactDamageDice is the number of damage dice an Impact inflicts: Speed
// squared (Book 1 p.202).
func ImpactDamageDice(speed int) int { return speed * speed }

// Absorb returns the hits that get through armor: those beyond the armor rating,
// which absorbs hits equal to its value (Book 1 p.202).
func Absorb(hits, armor int) int { return max(hits-armor, 0) }

// Wound applies penetrating hits to a character's three physical characteristics
// (Strength, Dexterity, Endurance), cascading from each into the next as it
// reaches zero (Book 1 p.202). It returns the reduced characteristics.
func Wound(physicals [3]int, hits int) [3]int {
	for i := range physicals {
		if hits <= 0 {
			break
		}

		if hits >= physicals[i] {
			hits -= physicals[i]
			physicals[i] = 0
		} else {
			physicals[i] -= hits
			hits = 0
		}
	}

	return physicals
}
