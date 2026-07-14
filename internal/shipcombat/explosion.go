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

// missileMassiveExplosion is the Sz+1D proximity table (Book 2 p.197), indexed by
// the Size + 1D roll: 6 or less a direct hit, 12 or more a clean miss.
var missileMassiveExplosion = map[int]MassiveExplosion{
	7:  {Proximity: "Hit", Blast: 90, BFE: 20, Rad: 10, Burn: 30},
	8:  {Proximity: "Hit", Blast: 40, BFE: 15, Rad: 10, Burn: 20},
	9:  {Proximity: "Very Near Miss", Blast: 30, BFE: 10, Rad: 10, Burn: 10},
	10: {Proximity: "Near Miss", Blast: 10, BFE: 5, Rad: 5, Burn: 5},
	11: {Proximity: "Far Miss", Blast: 5, BFE: 1, Rad: 1, Burn: 1},
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
		return missileMassiveExplosion[szPlus1D]
	}
}
