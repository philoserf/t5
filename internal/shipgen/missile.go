package shipgen

import (
	"fmt"
	"slices"
	"strings"
)

// Missiles (Book 2 pp.157, 170). A missile is not a weapon: it is the ordnance a
// weapon throws, and the book identifies it separately —
//
//	Stage Type-TL Size Effect Guidance
//
// Its Stage and tech level come from the launcher that fires it. What the
// designer chooses is the Size (1-7, the object-size ladder from a bullet to a
// 100-ton ship), the Effect (the warhead), and the Guidance (the brain).
//
// The three constrain each other, which is the whole design problem. A launcher
// only throws certain sizes; each size only carries certain warheads; and the
// cleverer guidance systems need both a big enough missile to hold the brain and
// a high enough tech level to build one. An Operator-Guided missile needs Size 4
// and TL 9; a Self-Aware one needs Size 5 and TL 11.

// A MissileType is a warhead (Book 2 p.170), lettered S D X E N A K Y Z.
type MissileType int

// Missile warhead types.
const (
	Slug       MissileType = iota // S — a solid metal projectile
	Deadfall                      // D — gravity-driven, against a world surface
	Explosive                     // X — a high explosive charge
	EMP                           // E — an ElectroMagnetic Pulse, to kill electronics
	Nuke                          // N — fission or fusion
	AntiMatter                    // A
	Kinetic                       // K — damage through sheer velocity
	Decoy                         // Y — appears as any of SDXENAZ
	SensorPkg                     // Z — carries a sensor rather than a warhead
)

// The warhead names. Their letters are the const block's own comments above — the
// book indexes missiles by size, not by warhead letter, so nothing looks one up.
var missileTypeNames = [...]string{
	Slug:       "Slug",
	Deadfall:   "Deadfall",
	Explosive:  "Explosive",
	EMP:        "EMP",
	Nuke:       "Nuke",
	AntiMatter: "AM",
	Kinetic:    "Kinetic",
	Decoy:      "Decoy",
	SensorPkg:  "Sensor",
}

func (m MissileType) String() string {
	if m < 0 || int(m) >= len(missileTypeNames) {
		return "?"
	}

	return missileTypeNames[m]
}

// A Guidance is a missile's brain (Book 2 pp.157, 170) — how it finds its target
// once it is away. It is also an asset on the MissileLauncher Attack Task: an unguided
// missile contributes nothing, a hardwired one a flat 5, and an operator-guided
// one the gunner's own Characteristic + Skill + Knowledge.
type Guidance int

// Missile guidance packages.
const (
	UnGuided       Guidance = iota // UG — launched, and then on its own
	HardWired                      // HW — pre-programmed circuits
	OperatorGuided                 // OG — the gunner flies it in
	SelfAware                      // SA — an on-board brain
	DownLoaded                     // DL — the gunner's own personality, copied in
)

// guidanceData is Book 2 p.170: each system's code, the smallest missile that can
// house it, and the lowest tech level that can build one. UnGuided needs neither —
// it is the absence of a guidance system.
var guidanceData = [...]struct {
	code    string
	minSize int
	minTL   int
}{
	UnGuided:       {"UG", 0, 0},
	HardWired:      {"HW", 0, 5},
	OperatorGuided: {"OG", 4, 9},
	SelfAware:      {"SA", 5, 11},
	DownLoaded:     {"DL", 5, 13},
}

func (g Guidance) String() string {
	if g < 0 || int(g) >= len(guidanceData) {
		return "?"
	}

	return guidanceData[g].code
}

// Available reports whether a missile of the given size, thrown by a launcher of
// the given tech level, can carry this guidance system.
func (g Guidance) Available(size, launcherTL int) bool {
	if g < 0 || int(g) >= len(guidanceData) {
		return false
	}

	d := guidanceData[g]

	return size >= d.minSize && launcherTL >= d.minTL
}

// missileRow is one row of the MISSILES table (Book 2 p.170): a launcher throwing
// a missile of a given size, with the guidance systems and warheads that
// combination allows. A launcher may appear more than once — the MissileLauncher launcher
// throws Size 4, 5, or 6 — and each size it throws is its own row, because the
// bigger round carries different brains and different warheads.
var missileRows = [...]struct {
	launcher WeaponID
	size     int
	name     string // the round's own name: "Small Missile", "Small Craft"
	guidance []Guidance
	types    []MissileType
	perTon   int // how many fit in a ton (0 when the round is measured in tons)
	tonsEach int // tons per round (0 when many fit in a ton)
}{
	{SlugThrower, 2, "Slug", []Guidance{UnGuided}, []MissileType{Slug, Explosive}, 0, 0},
	{
		SalvoRack, 3, "Vsmall Missile",
		[]Guidance{UnGuided, HardWired},
		[]MissileType{Slug, Explosive, EMP, Decoy, SensorPkg},
		3000, 0,
	},
	{
		MissileLauncher, 4, "Small Missile",
		[]Guidance{UnGuided, HardWired, OperatorGuided},
		[]MissileType{Explosive, EMP, Decoy, SensorPkg},
		400, 0,
	},
	{
		AMMissile, 4, "Small Missile",
		[]Guidance{SelfAware, DownLoaded},
		[]MissileType{Deadfall, AntiMatter, Decoy, SensorPkg},
		400, 0,
	},
	{
		MissileLauncher, 5, "Missile",
		[]Guidance{HardWired, OperatorGuided, SelfAware, DownLoaded},
		[]MissileType{Explosive, EMP, Nuke, Decoy, SensorPkg},
		50, 0,
	},
	{
		Ortillery, 5, "Missile",
		[]Guidance{UnGuided, HardWired, OperatorGuided},
		[]MissileType{Deadfall, Decoy, SensorPkg},
		50, 0,
	},
	{
		MissileLauncher, 6, "Small Craft",
		[]Guidance{HardWired, OperatorGuided, SelfAware, DownLoaded},
		[]MissileType{Explosive, Decoy, SensorPkg},
		0, 5,
	},
	{
		RailGun, 6, "Small Craft",
		[]Guidance{UnGuided, HardWired, OperatorGuided, SelfAware, DownLoaded},
		[]MissileType{Deadfall, Decoy, SensorPkg},
		0, 5,
	},
	{
		KKMissile, 7, "Ship",
		[]Guidance{SelfAware, DownLoaded},
		[]MissileType{Kinetic, Decoy, SensorPkg},
		0, 100,
	},
}

// missileEffects is the MISSILE TYPES AND EFFECTS grid (Book 2 p.170), keyed by
// size and warhead. The effect is what the missile does when it arrives, and it
// is rendered as the book writes it: "Pen= 4", "EMP= 4", "ME /20".
//
// Note the book's own missile catalogs (pp.167, 170) print the EMP effect one
// lower than this grid does — "Size-4 EMP=3" where the grid two inches away says
// EMP= 4, and "Size-5 EMP=4" where it says 5 — while their Explosive values match
// the grid exactly. The grid is self-consistent (both Pen and EMP simply equal
// the Size), so we follow it, as we follow the design tables everywhere else.
func missileEffect(size int, t MissileType) (string, bool) { //nolint:cyclop // warhead table, one branch per row
	switch t {
	case Slug:
		if size >= 1 && size <= 3 {
			return fmt.Sprintf("Pen= %d", size-1), true
		}
	case Explosive:
		switch {
		case size >= 2 && size <= 5:
			return fmt.Sprintf("Pen= %d", size), true
		case size == 6:
			return "ME / 5", true
		}
	case EMP:
		if size >= 3 && size <= 5 {
			return fmt.Sprintf("EMP= %d", size), true
		}
	case Deadfall:
		switch size {
		case 4:
			return "ME /20", true
		case 5:
			return "ME /10", true
		case 6:
			return "ME / 5", true
		}
	case Nuke:
		if size == 5 {
			return "ME", true
		}
	case AntiMatter:
		if size == 4 {
			return "ME x2", true
		}
	case Kinetic:
		// Kinetic damage is the missile's speed, not a fixed number: Pen = Size x
		// Speed from the launch point.
		if size >= 6 && size <= 7 {
			return fmt.Sprintf("Pen= %dxSp", size), true
		}
	case Decoy:
		return "Decoy", true // appears as any of SDXENAZ
	case SensorPkg:
		return "Sensor", true // carries a sensor, and uses the sensor rules
	}

	return "", false
}

// A MissileSpec is a round as the designer orders it: a size the launcher throws,
// a warhead, and a brain.
type MissileSpec struct {
	Launcher WeaponID
	Size     int
	Type     MissileType
	Guidance Guidance
}

// A Missile is a designed round. Stage and TL come from the launcher that throws
// it (Book 2 p.155), not from the round itself.
type Missile struct {
	Spec MissileSpec

	Stage    Stage
	TL       int
	Effect   string
	PerTon   int // rounds per ton of magazine (0 if the round is measured in tons)
	TonsEach int // tons per round (0 if many fit in a ton)
	Problems []string
}

// DesignMissile builds a round for a designed launcher (Book 2 p.170). Like the
// rest of shipgen it is total: a round the launcher cannot throw, a warhead the
// size cannot carry, or a brain too clever for its tech level is reported in
// Problems rather than refused.
func DesignMissile(launcher Weapon, spec MissileSpec) Missile {
	spec.Launcher = launcher.Spec.Model
	m := Missile{Spec: spec, Stage: launcher.Spec.Stage, TL: launcher.TL}

	i, ok := findMissileRow(spec.Launcher, spec.Size)
	if !ok {
		m.Problems = append(m.Problems, fmt.Sprintf("%s does not throw a Size-%d round",
			launcher.Name(), spec.Size))

		return m
	}

	row := missileRows[i]
	m.PerTon, m.TonsEach = row.perTon, row.tonsEach

	effect, ok := missileEffect(spec.Size, spec.Type)
	if !ok {
		m.Problems = append(m.Problems, fmt.Sprintf("a Size-%d round carries no %s warhead",
			spec.Size, spec.Type))
	}

	m.Effect = effect
	if !slices.Contains(row.types, spec.Type) {
		m.Problems = append(m.Problems, fmt.Sprintf("%s does not throw a %s round",
			launcher.Name(), spec.Type))
	}

	if !slices.Contains(row.guidance, spec.Guidance) {
		m.Problems = append(m.Problems, fmt.Sprintf("a Size-%d %s round cannot be %s",
			spec.Size, row.name, spec.Guidance))
	} else if !spec.Guidance.Available(spec.Size, launcher.TL) {
		d := guidanceData[spec.Guidance]
		m.Problems = append(
			m.Problems,
			fmt.Sprintf("%s guidance needs Size %d at TL %d (this is Size %d at TL %d)",
				spec.Guidance, d.minSize, d.minTL, spec.Size, launcher.TL),
		)
	}

	return m
}

// LongName renders the round's identity as the book's own missile tables do
// (Book 2 p.170 "Identifying Missiles"):
//
//	Advanced Missile-10 Size-4 EMP= 4. Guidance: OG HW
//
// The guidance clause lists the brains this round can actually carry, and names
// the ones it cannot — a Size-5 round at TL 10 is too early for a Self-Aware or
// DownLoaded mind, and the book says so in parentheses.
func (m Missile) LongName() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s Missile-%d Size-%d %s.", m.Stage, m.TL, m.Spec.Size, m.Effect)
	// Which brains this round can carry, and which it cannot — derived, not stored:
	// everything the split needs is already in the Missile and the launcher's row.
	if can, cannot := m.guidanceOptions(); len(can) > 0 || len(cannot) > 0 {
		if len(can) > 0 {
			fmt.Fprintf(&b, " Guidance: %s", guidanceList(can))
		}

		if len(cannot) > 0 {
			fmt.Fprintf(&b, " (%s not available)", guidanceList(cannot))
		}
	}

	return b.String()
}

// guidanceOptions splits the brains this round's launcher offers into those its
// size and tech level allow, and those they do not.
func (m Missile) guidanceOptions() ([]Guidance, []Guidance) {
	i, ok := findMissileRow(m.Spec.Launcher, m.Spec.Size)
	if !ok {
		return nil, nil
	}

	var can, cannot []Guidance

	for _, g := range missileRows[i].guidance {
		if g == UnGuided {
			continue // not a guidance system; it is the lack of one
		}

		if g.Available(m.Spec.Size, m.TL) {
			can = append(can, g)
		} else {
			cannot = append(cannot, g)
		}
	}

	return can, cannot
}

// guidanceOrder is the book's print order — the cleverest brain first, since that
// is the one a designer is choosing between.
var guidanceOrder = [...]Guidance{OperatorGuided, HardWired, SelfAware, DownLoaded}

// guidanceList renders guidance codes in the book's order.
func guidanceList(gs []Guidance) string {
	codes := make([]string, 0, len(gs))
	for _, g := range guidanceOrder {
		if slices.Contains(gs, g) {
			codes = append(codes, g.String())
		}
	}

	return strings.Join(codes, " ")
}

func findMissileRow(launcher WeaponID, size int) (int, bool) {
	for i, r := range missileRows {
		if r.launcher == launcher && r.size == size {
			return i, true
		}
	}

	return 0, false
}
