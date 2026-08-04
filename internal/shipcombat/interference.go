package shipcombat

import "github.com/philoserf/t5/internal/shipgen"

// Interference and clustered protection (Book 2 p.201).

// MaximumInterferenceRange is the greatest world range at which a ship, fighter,
// or small craft may interfere with an attack against another target. The book
// states this on the world-range ladder as R=7 (50 km), not as Space Range S=7.
const MaximumInterferenceRange = 7

// AttackKind is the class of attack relevant to interference. Missile and beam
// attacks name their special defensive fire; every other attack may be met by an
// available defense.
type AttackKind uint8

const (
	// MissileAttack may be met by Anti-Missile fire.
	MissileAttack AttackKind = iota
	// BeamAttack may be met by Anti-Beam fire.
	BeamAttack
	// OtherAttack may be met by an available applicable defense.
	OtherAttack
)

// InterferenceOptions reports the actions an eligible interfering craft may
// take. Substitution is the fallback available against every attack; the other
// field set depends on the attack kind.
type InterferenceOptions struct {
	AntiMissileFire  bool
	AntiBeamFire     bool
	AvailableDefense bool
	SubstituteTarget bool
}

// CanInterfere reports whether a ship, fighter, or small craft at separation R=
// rangeR from an attacked target is close enough to interfere (Book 2 p.201).
// The range is on the world-range ladder; negative range codes are invalid.
func CanInterfere(rangeR int) bool {
	return rangeR >= 0 && rangeR <= MaximumInterferenceRange
}

// InterferenceAgainst returns the interference actions available against an
// attack at the given R= separation from its target. Outside R=7 there are none.
func InterferenceAgainst(kind AttackKind, rangeR int) InterferenceOptions {
	if !CanInterfere(rangeR) {
		return InterferenceOptions{}
	}

	options := InterferenceOptions{SubstituteTarget: true}

	switch kind {
	case MissileAttack:
		options.AntiMissileFire = true
	case BeamAttack:
		options.AntiBeamFire = true
	case OtherAttack:
		options.AvailableDefense = true
	default:
		return InterferenceOptions{}
	}

	return options
}

// ProtectedByDefense reports whether a nearby craft is within an assisting
// defense's range. Both values must be codes on the same range ladder. Cluster
// membership itself is a tactical declaration; p.201 imposes no additional roll
// or numerical limit on forming a cluster.
func ProtectedByDefense(separation, defenseRange int) bool {
	return separation >= 0 && defenseRange >= 0 && separation <= defenseRange
}

// DecoyInterferes reports whether a designed decoy missile may substitute itself
// for an attacked target. It must be valid, within R=7, and at least Target Size
// minus one. Target Size may be any value on the object-size scale.
func DecoyInterferes(rangeR int, decoy shipgen.Missile, targetSize int) bool {
	return targetSize >= 0 && len(decoy.Problems) == 0 && decoy.Spec.Type == shipgen.Decoy &&
		CanInterfere(rangeR) && decoy.Spec.Size >= targetSize-1
}
