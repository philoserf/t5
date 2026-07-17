package shipcombat

// Massive Explosion (Book 2 p.197). A missile with an explosive option (AM, Nuke)
// resolved through the Missile Attack Task applies a Massive Explosion whose
// effects fall off with proximity, itself set by a Size + 1D roll.

// A MassiveExplosion is the effect of a missile detonation at a given proximity
// (Book 2 p.197). Blast, BFE (Bang/Flash/EMP), Rad, and Burn are damage dice
// counts (D); Vaporized is a direct hit that destroys the target outright. Bang
// only applies in atmosphere, and EMP only for nuclear options.
type MassiveExplosion struct {
	Proximity             string
	Vaporized             bool
	Blast, BFE, Rad, Burn int
}

// missileMassiveExplosion is the Sz+1D proximity table for rolls 7-11 (Book 2
// p.197), indexed by the Size + 1D roll - 7; 6 or less is a direct hit and 12 or
// more a clean miss, handled in MissileMassiveExplosion.
var missileMassiveExplosion = [...]MassiveExplosion{
	{Proximity: "Hit", Blast: 90, BFE: 20, Rad: 10, Burn: 30},            // 7
	{Proximity: "Hit", Blast: 40, BFE: 15, Rad: 10, Burn: 20},            // 8
	{Proximity: "Very Near Miss", Blast: 30, BFE: 10, Rad: 10, Burn: 10}, // 9
	{Proximity: "Near Miss", Blast: 10, BFE: 5, Rad: 5, Burn: 5},         // 10
	{Proximity: "Far Miss", Blast: 5, BFE: 1, Rad: 1, Burn: 1},           // 11
}

// MissileMassiveExplosion returns a detonation's effects at a Size + 1D roll
// (Book 2 p.197). A roll of 6 or less is a Direct Hit (Vaporized, 100D); 12 or
// more is a clean Miss with no effect.
func MissileMassiveExplosion(szPlus1D int) MassiveExplosion {
	switch {
	case szPlus1D <= 6:
		return MassiveExplosion{Proximity: "Direct Hit", Vaporized: true, Blast: 100}
	case szPlus1D >= 12:
		return MassiveExplosion{Proximity: "Miss"}
	default:
		return missileMassiveExplosion[szPlus1D-7]
	}
}

// A MassiveExplosionMultiplier scales a missile option's effects when it hits via
// the Space Weapon Task at close range (Book 2 p.196 Massive Explosion, Weapons
// Task): a multiple of the target's Armor Value defeated, and of the Rad, Flash,
// and EMP effects. EMP applies only to nuclear options.
type MassiveExplosionMultiplier struct {
	AV, Rad, Flash, EMP int
}

// weaponsMassiveExplosion is the Weapons-Task Massive Explosion table by missile
// option (Book 2 p.196).
var weaponsMassiveExplosion = map[string]MassiveExplosionMultiplier{
	"AM":        {AV: 10, Flash: 1},
	"Nuke":      {AV: 10, Rad: 1, Flash: 1, EMP: 1},
	"EMP":       {AV: 1, EMP: 1},
	"KK":        {AV: 1},
	"DeadFall4": {AV: 1, Flash: 1},
	"DeadFall5": {AV: 5, Flash: 2},
	"DeadFall6": {AV: 10, Flash: 3},
}

// WeaponsMassiveExplosion returns the effect multipliers for a missile option
// detonating via the Space Weapon Task (Book 2 p.196), and whether the option is
// tabulated.
func WeaponsMassiveExplosion(option string) (MassiveExplosionMultiplier, bool) {
	m, ok := weaponsMassiveExplosion[option]

	return m, ok
}
