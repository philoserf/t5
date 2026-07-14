package shipgen

// The weapon design tables, transcribed from the rendered image of Book 2 p.83
// ("Starship Weapons2"), which carries the whole weapon-design system on one
// page: Table A (the weapons), Table C (the mounts), and Tables D and E (the
// range effects). Table B (TL Stage Effects) is the stage ladder shipgen already
// has for drives — see stageData in drive.go, which this reuses.
//
// The dense grids do not survive pdftotext, so these were read cell by cell off
// the page image.

// A WeaponID names a weapon model. The book identifies each by a single letter
// (Book 2 p.83 Table A), which is what the ship's weapon record prints.
type WeaponID int

// The 23 weapon models of Book 2 p.83, in the page's own grouping: Beams,
// Missiles, Exotics, and Specials.
const (
	MiningLaser WeaponID = iota // J
	PulseLaser                  // K
	BeamLaser                   // L
	PlasmaGun                   // P
	FusionGun                   // F

	SlugThrower // B
	SalvoRack   // V
	RailGun     // R
	Missile     // M
	KKMissile   // N
	AMMissile   // X

	JumpDamper     // T
	TractorPressor // U
	Inducer        // H
	Disruptor      // W
	Stasis         // E

	DataCaster    // D
	SandCaster    // S
	Ortillery     // Q
	CommCaster    // C
	HybridSLM     // Y
	ParticleAccel // A
	MesonGun      // G
)

// A Scale is the range ladder a weapon reaches on: the great distances of space
// (S=, Book 1 pp.24-29) or the shorter world ranges (R=). It selects which range
// table applies — Space Range Effects (p.83 Table D) or World Range Effects
// (Table E).
type Scale int

const (
	SpaceScale Scale = iota // S=
	WorldScale              // R=
	// A few weapons reach on both ladders — the book marks them "S=7* R=7*"
	// (p.83 Table A) — and may be built for a range on either.
	EitherScale
)

// reaches reports whether a weapon on this scale can be built for a range on the
// given one.
func (s Scale) reaches(r Scale) bool { return s == EitherScale || s == r }

// weaponData is Book 2 p.83 Table A: each weapon's letter, name, base tech
// level, minimum mount, range scale, and base cost. A weapon has no tonnage of
// its own — "Base tonnage for a Weapon is zero tons" (p.156); the tonnage is
// entirely the mount's.
//
// mod and hitsDice are the two per-weapon exceptions the page's prose adds to
// the tables (p.158): the Beam Laser is easier to aim (Mod +2 "in addition to
// any others") and the Pulse Laser hits harder (+1D "on any Pulse Laser
// installation"). Every other weapon takes its Mod and Hits from the mount.
var weaponData = [...]struct {
	letter   byte
	name     string
	tl       int
	minMount Mount
	scale    Scale
	cost     int // Cr
	mod      int // weapon's own attack Mod, beyond the mount's
	hitsDice int // extra damage dice, beyond the mount's
}{
	MiningLaser: {'J', "Mining Laser", 8, SingleTurret, WorldScale, 500_000, 0, 0},
	PulseLaser:  {'K', "Pulse Laser", 9, SingleTurret, WorldScale, 300_000, 0, 1},
	BeamLaser:   {'L', "Beam Laser", 10, SingleTurret, WorldScale, 500_000, 2, 0},
	PlasmaGun:   {'P', "Plasma Gun", 11, SingleBarbette, WorldScale, 1_000_000, 0, 0},
	FusionGun:   {'F', "Fusion Gun", 12, SingleBarbette, WorldScale, 1_500_000, 0, 0},

	SlugThrower: {'B', "Slug Thrower", 9, SingleTurret, WorldScale, 200_000, 0, 0},
	SalvoRack:   {'V', "Salvo Rack", 10, Bay, WorldScale, 10_000_000, 0, 0},
	RailGun:     {'R', "Rail Gun", 12, Bay, SpaceScale, 12_000_000, 0, 0},
	Missile:     {'M', "Missile", 7, SingleTurret, SpaceScale, 2_000_000, 0, 0},
	KKMissile:   {'N', "KK Missile", 10, Bay, SpaceScale, 3_000_000, 0, 0},
	AMMissile:   {'X', "AM Missile", 20, SingleBarbette, SpaceScale, 5_000_000, 0, 0},

	JumpDamper:     {'T', "Jump Damper", 14, SingleBarbette, WorldScale, 15_000_000, 0, 0},
	TractorPressor: {'U', "Tractor Pressor", 16, SingleBarbette, WorldScale, 5_000_000, 0, 0},
	Inducer:        {'H', "Inducer", 17, SingleTurret, WorldScale, 1_000_000, 0, 0},
	Disruptor:      {'W', "Disruptor", 18, SingleBarbette, SpaceScale, 15_000_000, 0, 0},
	Stasis:         {'E', "Stasis", 21, SingleTurret, SpaceScale, 5_000_000, 0, 0},

	DataCaster: {'D', "DataCaster", 10, SingleTurret, WorldScale, 1_000_000, 0, 0},
	SandCaster: {'S', "SandCaster", 9, SingleTurret, WorldScale, 100_000, 0, 0},
	Ortillery:  {'Q', "Ortillery", 12, Bay, WorldScale, 15_000_000, 0, 0},
	CommCaster: {'C', "CommCaster", 8, SingleTurret, SpaceScale, 5_000_000, 0, 0},
	// The Hybrid and the PA are the two dual-scale weapons ("S=7* R=7*").
	HybridSLM:     {'Y', "Hybrid SLM", 10, SingleTurret, EitherScale, 1_000_000, 0, 0},
	ParticleAccel: {'A', "PA", 11, SingleBarbette, EitherScale, 2_500_000, 0, 0},
	MesonGun:      {'G', "Meson Gun", 13, Main, SpaceScale, 5_000_000, 0, 0},
}

// A Mount is the structure a weapon is installed in (Book 2 p.83 Table C). The
// mount, not the weapon, supplies the tonnage, the damage dice, and the attack
// Mod: a weapon in a bigger mount hits more often and hits harder.
type Mount int

const (
	SingleTurret   Mount = iota // T1
	DualTurret                  // T2
	TripleTurret                // T3
	QuadTurret                  // T4
	SingleBarbette              // B1
	DualBarbette                // B2
	Bay                         // Bay
	LargeBay                    // LBay
	Main                        // M
)

// mountData is Book 2 p.83 Table C. Extendable and Deployable are omitted: they
// are tonnage-and-cost surcharges on a turret ("in addition to Turret cost"),
// not mounts a weapon is installed in on their own.
var mountData = [...]struct {
	code string
	name string
	tons int
	mod  int
	hits int // damage dice
	cost int // Cr
}{
	SingleTurret:   {"T1", "Single Turret", 1, -2, 1, 200_000},
	DualTurret:     {"T2", "Dual Turret", 1, -1, 2, 500_000},
	TripleTurret:   {"T3", "Triple Turret", 1, 0, 3, 1_000_000},
	QuadTurret:     {"T4", "Quad Turret", 1, 1, 4, 1_500_000},
	SingleBarbette: {"B1", "Barbette", 3, 2, 5, 3_000_000},
	DualBarbette:   {"B2", "Dual Barbette", 5, 3, 10, 4_000_000},
	Bay:            {"Bay", "Bay", 50, 5, 20, 5_000_000},
	LargeBay:       {"LBay", "Large Bay", 100, 8, 30, 10_000_000},
	Main:           {"Main", "Main", 200, 10, 100, 20_000_000},
}

// A Range is the reach a weapon and mount are built for. Choosing a shorter
// range than the standard buys a cheaper, lighter, lower-TL installation;
// reaching further costs tonnage, money, and tech level. The effects apply to
// the mount, not the weapon (Book 2 p.156).
type Range int

// The six Space ranges (p.83 Table D) and the six World ranges (Table E). A
// weapon is built for a range on its own scale; AttackRange and VDistant are the
// standard (unmodified) rung of each ladder.
const (
	Boarding     Range = iota // S=0
	FighterRange              // S=2
	ShortRange                // S=5
	AttackRange               // S=7  — standard
	LongRange                 // S=9
	DeepSpace                 // S=12

	Vlong    // R=5
	Distant  // R=6
	VDistant // R=7  — standard
	Orbit    // R=8
	Far      // R=9
	Geo      // R=10
)

// rangeData is Book 2 p.83 Tables D (Space) and E (World). Each range shifts the
// weapon's tech level and scales the mount's tonnage and cost. The multipliers
// are stored as fractions (num/den) to stay exact in integer credits.
var rangeData = [...]struct {
	name             string
	scale            Scale
	band             int // the S= or R= number the weapon record prints
	tlMod            int
	tonsNum, tonsDen int
	costNum, costDen int
}{
	Boarding:     {"Boarding", SpaceScale, 0, -3, 1, 4, 1, 4},
	FighterRange: {"Fighter Range", SpaceScale, 2, -2, 1, 3, 1, 3},
	ShortRange:   {"Short Range", SpaceScale, 5, -1, 1, 2, 1, 2},
	AttackRange:  {"Attack Range", SpaceScale, 7, 0, 1, 1, 1, 1},
	LongRange:    {"Long Range", SpaceScale, 9, 1, 2, 1, 3, 1},
	DeepSpace:    {"Deep Space", SpaceScale, 12, 2, 3, 1, 5, 1},

	Vlong:    {"Vlong", WorldScale, 5, -2, 1, 3, 1, 3},
	Distant:  {"Distant", WorldScale, 6, -1, 1, 2, 1, 2},
	VDistant: {"Vdistant", WorldScale, 7, 0, 1, 1, 1, 1},
	Orbit:    {"Orbit", WorldScale, 8, 1, 2, 1, 3, 1},
	Far:      {"Far", WorldScale, 9, 2, 3, 1, 5, 1},
	Geo:      {"Geo", WorldScale, 10, 3, 4, 1, 6, 1},
}

// weaponStageMod is the Mod column of Book 2 p.83 Table B, the one column of the
// stage ladder that drives do not use. It is NOT the same as the stage's tlDelta
// (drive.go): Generic raises tech level by +1 but grants no Mod. The p.156
// summary lists Generic as +1, contradicting the p.83 design table it summarizes;
// we follow p.83, the table the designer actually builds from.
var weaponStageMod = [...]int{
	Standard:     0,
	Experimental: -3,
	Prototype:    -2,
	Early:        -1,
	Basic:        0,
	Alternate:    0,
	Improved:     1,
	Generic:      0,
	Modified:     2,
	Advanced:     3,
	Ultimate:     4,
}
