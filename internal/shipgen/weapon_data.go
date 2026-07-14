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

	SlugThrower     // B
	SalvoRack       // V
	RailGun         // R
	MissileLauncher // M
	KKMissile       // N
	AMMissile       // X

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

	SlugThrower:     {'B', "Slug Thrower", 9, SingleTurret, WorldScale, 200_000, 0, 0},
	SalvoRack:       {'V', "Salvo Rack", 10, Bay, WorldScale, 10_000_000, 0, 0},
	RailGun:         {'R', "Rail Gun", 12, Bay, SpaceScale, 12_000_000, 0, 0},
	MissileLauncher: {'M', "Missile", 7, SingleTurret, SpaceScale, 2_000_000, 0, 0},
	KKMissile:       {'N', "KK Missile", 10, Bay, SpaceScale, 3_000_000, 0, 0},
	AMMissile:       {'X', "AM Missile", 20, SingleBarbette, SpaceScale, 5_000_000, 0, 0},

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

// The mounts, in ascending size — the order matters, because a weapon may go in
// "any Mount equal to or greater than the Minimum" (Book 2 p.155), which is a
// comparison on this ordering.
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

	// BoltIn is the defenses' own mount (Book 2 p.176): three tons placed
	// anywhere inside the ship, needing neither a hardpoint nor a view outside.
	// No weapon uses it — it comes last so it sits outside the size ordering
	// above, which it does not belong to.
	BoltIn // Bo
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
// weapon's tech level and scales the mount's tonnage and cost.
//
// The multipliers are hundredths, not exact fractions, because that is what the
// book computes with: it divides by three by multiplying by 0.33. Its own worked
// defenses (p.176) put a 200-ton Main mount at Vlong at "66 tons" (200 x 0.33)
// and a 3-ton Bolt-In at "0.99 tons. MCr3.99" — numbers exact division would give
// as 66.67 and 1.00. Following the book's arithmetic reproduces its tables to the
// credit; being more precise than the book would not.
//
// defenseOK marks the ranges a defense may be built for. A defense reaches at
// most Vdistant: Book 2 p.174 greys out Orbit, Far, and Geo on its copy of the
// World table, and the Defense Ranges table (p.179) stops at R=7. A defense can
// be built for less reach, never for more.
var rangeData = [...]struct {
	name      string
	scale     Scale
	band      int // the S= or R= number the record prints
	tlMod     int
	tons      int // hundredths: 33 = x0.33, 200 = x2
	cost      int // hundredths
	defenseOK bool
}{
	Boarding:     {"Boarding", SpaceScale, 0, -3, 25, 25, true},
	FighterRange: {"Fighter Range", SpaceScale, 2, -2, 33, 33, true},
	ShortRange:   {"Short Range", SpaceScale, 5, -1, 50, 50, true},
	AttackRange:  {"Attack Range", SpaceScale, 7, 0, 100, 100, true},
	LongRange:    {"Long Range", SpaceScale, 9, 1, 200, 300, true},
	DeepSpace:    {"Deep Space", SpaceScale, 12, 2, 300, 500, true},

	Vlong:    {"Vlong", WorldScale, 5, -2, 33, 33, true},
	Distant:  {"Distant", WorldScale, 6, -1, 50, 50, true},
	VDistant: {"Vdistant", WorldScale, 7, 0, 100, 100, true},
	Orbit:    {"Orbit", WorldScale, 8, 1, 200, 300, false},
	Far:      {"Far", WorldScale, 9, 2, 300, 500, false},
	Geo:      {"Geo", WorldScale, 10, 3, 400, 600, false},
}

// weaponStageData is Book 2 p.83 Table B (printed again for defenses on p.174) —
// the stage ladder as weapons use it. It shares the eleven stages and their tech
// level shifts with the drives' X table (stageData, drive.go), but it is not the
// same table:
//
//   - The cost multipliers agree on every stage but one. Modified costs a drive
//     x1 (and saves it half its tonnage); it costs a weapon /2. The book prints
//     both, and each table governs its own kind of component — the defense
//     catalog's Modified White Globe (p.176, MCr7) only works at /2, and the
//     drives have no worked Modified example to contradict their own table.
//   - The Mod column is weapons-only; drives have no attack. It is NOT the tech
//     level shift under another name: Generic raises tech level by +1 but grants
//     no Mod.
//
// mod applies to weapons alone. A defense takes its Mod from its mount and none
// from its stage — see defense.go.
var weaponStageData = [...]struct {
	costNum, costDen int
	mod              int
}{
	Standard:     {1, 1, 0},
	Experimental: {10, 1, -3},
	Prototype:    {5, 1, -2},
	Early:        {2, 1, -1},
	Basic:        {1, 2, 0},
	Alternate:    {1, 1, 0},
	Improved:     {1, 1, 1},
	Generic:      {1, 2, 0},
	Modified:     {1, 2, 2},
	Advanced:     {2, 1, 3},
	Ultimate:     {3, 1, 4},
}
