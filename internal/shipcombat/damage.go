package shipcombat

// Hit location, penetration, and damage (Book 2 pp. 195-198).

// Damage-application capacities (Book 2 p.198): a SubCompartment absorbs up to 10
// damage points before overflowing to an adjacent one; 60 points knock out a
// full Compartment.
const (
	SubCompartmentCapacity = 10
	CompartmentCapacity    = 60
)

// HitCompartment is the compartment a hit lands on (Book 2 p.195): Flux plus the
// targeting mod (the targeted compartment number, or 0 for Center-of-Mass),
// clamped to the ship's compartment span [-span, +span].
func HitCompartment(flux, targetMod, span int) int {
	return clamp(flux+targetMod, -span, span)
}

// Penetrate resolves damage against a hit compartment's armor layers (Book 2
// p.197): each layer whose AV the damage meets or exceeds is reduced to zero for
// the rest of the battle and the surplus carries to the next layer; a layer the
// damage cannot meet stops the attack. Damage surviving every layer penetrates,
// and at least one point must remain to reach the interior. Returns whether it
// penetrated and the interior damage.
func Penetrate(damage int, layerAV []int) (penetrated bool, interior int) {
	for _, av := range layerAV {
		if damage < av {
			return false, 0
		}
		damage -= av
	}
	return damage >= 1, damage
}

// damageLocation is the Ship column of the L1 Damage Location table (Book 2
// p.198), indexed by a 2D roll.
var damageLocation = map[int]string{
	2: "Bridge", 3: "Hold", 4: "Sensors", 5: "Protections", 6: "Life Support",
	7: "Drives", 8: "Power Plant", 9: "Hull", 10: "Weaponry", 11: "Astrogation", 12: "Computer",
}

// DamageLocation names the ship system a hit strikes on a 2D roll (Book 2 p.198
// L1 Damage Location, Ship column); the roll is clamped to 2-12.
func DamageLocation(twoD int) string {
	return damageLocation[clamp(twoD, 2, 12)]
}

// severityLabel names each Damage/Diagnosis Severity level (Book 2 p.198),
// indexed by the clamped severity 1-9.
var severityLabel = [...]string{
	1: "Easy", 2: "Average", 3: "Difficult", 4: "Formidable",
	5: "Staggering", 6: "Hopeless", 7: "Impossible", 8: "Beyond", 9: "Destroyed",
}

// Severity resolves a Damage or Diagnosis Severity roll (Book 2 p.198): 1D plus a
// mod (typically the number of SubCompartments InOp), clamped to 1-9. Levels 1-8
// are the difficulty dice to Diagnose or Repair (Easy 1D … Beyond 8D); 9 is
// Destroyed — beyond repair.
func Severity(oneD, mod int) int {
	return clamp(oneD+mod, 1, 9)
}

// SeverityLabel names a severity level 1-9 (Book 2 p.198).
func SeverityLabel(severity int) string {
	return severityLabel[clamp(severity, 1, 9)]
}

// Destroyed reports whether a severity level is beyond repair (Book 2 p.198).
func Destroyed(severity int) bool { return severity >= 9 }

// clamp constrains v to [lo, hi].
func clamp(v, lo, hi int) int {
	return min(max(v, lo), hi)
}
