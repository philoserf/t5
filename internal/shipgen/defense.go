package shipgen

import (
	"fmt"
	"strings"
)

// Defense design (Book 2 pp.174, 175-187). Defenses are built exactly as weapons
// are — a device, a mount, a tech stage, and a range — and p.174 prints the same
// six tables for them that p.83 prints for weapons. So this reuses the weapon
// engine's Mount, Stage, and Range, and differs only where the book does:
//
//   - Defenses have their own mount table (defenseMountData). A turret's Mod is
//     positive here, not negative: a defense with more tubes has a better chance
//     of stopping what is coming, where a weapon in a small mount has a worse
//     chance of hitting. The Bolt-In is the defense's own mount — three tons
//     placed anywhere in the ship, needing neither hardpoint nor a view outside.
//   - The stage's Mod does not apply. p.174 prints the Mod column, but its own
//     worked defenses ignore it: an Improved Triple Turret is Mod +3, not +4; an
//     Ultimate Bay is +1, not +5. A defense's Mod is its mount's alone.
//   - A defense reaches at most Vdistant (see rangeData's defenseOK).
//   - A defense has no damage dice. It stops attacks; it does not make them.
//
// A weapon may also be allocated to the Defensive Fire mode (p.186), fighting off
// incoming missiles instead of shooting at ships — see DesignWeaponAsDefense.

// A DefenseID names a defense device (Book 2 p.174 Table A).
type DefenseID int

// The ten defense devices: the Dampers and Screens, the Scramblers, and the
// Globes. (The Jammer and Stealth Mask are anti-sensor Surface mounts, deferred
// with the rest of the sensor system.)
const (
	NuclearDamper DefenseID = iota // N
	MesonScreen                    // G
	ProtonScreen                   // R

	MagScrambler  // Q
	GravScrambler // J
	ElecScrambler // E

	BlackGlobe  // T
	WhiteGlobe  // U
	SilverGlobe // V
	StasisGlobe // W
)

// A Principle is the physics a device works on (Book 2 pp.83, 174). It decides
// what a defense can act against — a Magnetic scrambler does nothing to a device
// that works on gravitics — and a device may rest on more than one: the book lists
// a Fusion Gun as "Elec/Grav", and prints it that way when the gun is serving as a
// defense. So this is a set, not a single value.
type Principle int

const (
	Electronic Principle = 1 << iota
	Magnetic
	Gravitic
)

func (p Principle) String() string {
	var names []string
	for _, e := range []struct {
		bit  Principle
		name string
	}{{Electronic, "Elec"}, {Magnetic, "Mag"}, {Gravitic, "Grav"}} {
		if p&e.bit != 0 {
			names = append(names, e.name)
		}
	}
	switch len(names) {
	case 0:
		return "?"
	case 1:
		// Alone, each is spelled out in full: "Electronic", not "Elec".
		return map[Principle]string{Electronic: "Electronic", Magnetic: "Magnetic", Gravitic: "Gravitic"}[p]
	default:
		return strings.Join(names, "/") // "Elec/Grav"
	}
}

// defenseData is Book 2 p.174 Table A: each defense's letter, name, base tech
// level, range scale, base cost, and principle. Every one is a world-range device
// standing at R=7, and every one's natural mount is the Bolt-In.
var defenseData = [...]struct {
	letter    byte
	name      string
	tl        int
	cost      int // Cr
	principle Principle
}{
	NuclearDamper: {'N', "Nuclear Damper", 12, 1_000_000, Electronic},
	MesonScreen:   {'G', "Meson Screen", 13, 3_000_000, Electronic},
	ProtonScreen:  {'R', "Proton Screen", 19, 1_000_000, Electronic},

	MagScrambler:  {'Q', "Mag Scrambler", 14, 1_000_000, Magnetic},
	GravScrambler: {'J', "Grav Scrambler", 17, 2_000_000, Gravitic},
	ElecScrambler: {'E', "Elec Scrambler", 12, 2_000_000, Electronic},

	BlackGlobe:  {'T', "Black Globe", 16, 10_000_000, Electronic},
	WhiteGlobe:  {'U', "White Globe", 20, 10_000_000, Electronic},
	SilverGlobe: {'V', "Silver Globe", 22, 10_000_000, Electronic},
	StasisGlobe: {'W', "Stasis Globe", 24, 10_000_000, Electronic},
}

// defenseMountData is Book 2 p.174 Table C — the defenses' own mount table, whose
// Mods run the other way from the weapons' (a bigger mount defends better rather
// than aiming worse).
//
// The Quad Turret costs MCr2.5, not the MCr1.5 the table prints. The page's own
// worked defenses say so three times over and unanimously: the Early Vdistant
// QuadTurret White Globe is MCr22.5 (a MCr10 globe at x2, plus the mount), the
// Generic Vdistant QuadTurret Proton Screen is MCr3, and the Advanced Distant
// QuadTurret Silver Globe is MCr21.25 — each of which needs a MCr2.5 mount and
// none of which works with MCr1.5. Every other cell of the table reproduces those
// rows exactly, so the printed 1.5 is the outlier, and we follow the rows.
var defenseMountData = [...]struct {
	tons int
	mod  int
	cost int // Cr
}{
	BoltIn:         {3, 3, 3_000_000},
	SingleTurret:   {1, 1, 200_000},
	DualTurret:     {1, 2, 500_000},
	TripleTurret:   {1, 3, 1_000_000},
	QuadTurret:     {1, 4, 2_500_000},
	SingleBarbette: {3, 1, 3_000_000},
	DualBarbette:   {5, 2, 4_000_000},
	Bay:            {50, 1, 5_000_000},
	LargeBay:       {100, 1, 10_000_000},
	Main:           {200, 1, 20_000_000},
}

// A DefenseSpec is a defense installation as the designer chooses it. The zero
// value is a Standard-stage Nuclear Damper in a Bolt-In at Vdistant.
//
// Weapon names a weapon serving in the Defensive Fire mode instead of a screen
// (Book 2 p.186) — a laser shooting down incoming rounds rather than shooting at
// ships. It is a pointer because it is the exception: when it is set, Model is
// ignored. Spelling the exception out in the spec is what lets the spec alone
// determine the result, as it does everywhere else in this package. It used to be
// carried in a shadow field on the Defense, which left every weapon-as-defense
// claiming Model == NuclearDamper.
type DefenseSpec struct {
	Model  DefenseID
	Weapon *WeaponID

	Mount Mount
	Stage Stage
	Range Range
}

// A Defense is a designed defense installation. There is no Hits: a defense stops
// attacks, it does not make them.
type Defense struct {
	Spec DefenseSpec

	Device    string // the screen's name, or the weapon's when one is standing in
	TL        int
	Mod       int
	Tons      Tonnage
	Cost      int // Cr
	Band      int
	Principle Principle
	Problems  []string
}

// A device is the thing being installed as a defense: a screen of its own, or a
// weapon standing in for one.
type device struct {
	name      string
	tl, cost  int
	principle Principle
}

// DesignDefense computes a defense installation from its spec (Book 2 pp.174,
// 186). Like the rest of shipgen it is total: a defense the book does not allow is
// reported in Problems, not refused.
//
// A spec naming a Weapon builds that weapon in the Defensive Fire mode instead —
// it keeps its own tech level and price but takes the defenses' mount table, and
// loses whatever attack Mod it had (the Beam Laser's +2 among them), because it is
// no longer making an attack.
func DesignDefense(spec DefenseSpec) Defense {
	dev, problems := defenseDevice(spec)
	if problems != nil {
		return Defense{Spec: spec, Problems: problems}
	}
	if !validDefenseMount(spec.Mount) || !validRange(spec.Range) {
		return Defense{Spec: spec, Device: dev.name, Problems: []string{"unknown mount or range"}}
	}
	m := defenseMountData[spec.Mount]
	rng := rangeData[spec.Range]

	if !rng.defenseOK {
		problems = append(problems, fmt.Sprintf("a defense cannot be built for %s (Vdistant is the furthest)", rng.name))
	}
	// A weapon standing in as a defense still needs a mount big enough to hold it.
	if spec.Weapon != nil {
		if w := weaponData[*spec.Weapon]; spec.Mount < w.minMount {
			problems = append(problems, fmt.Sprintf("%s needs at least a %s",
				w.name, mountName(w.minMount)))
		}
	}

	tl, tons, cost, band := install(dev.tl, dev.cost, m.tons, m.cost, spec.Range, spec.Stage)
	return Defense{
		Spec:   spec,
		Device: dev.name,
		TL:     tl,
		// The mount's Mod alone. A defense takes none from its tech stage, which is
		// why the stage's Mod column is not even in scope here (weaponStageMod).
		Mod:       m.mod,
		Tons:      tons,
		Cost:      cost,
		Band:      band,
		Principle: dev.principle,
		Problems:  problems,
	}
}

// DesignWeaponAsDefense allocates a weapon to the Defensive Fire mode (Book 2
// p.186): a Plasma or Fusion Gun, or a Mining, Pulse, or Beam Laser, shooting down
// incoming missiles instead of attacking ships.
func DesignWeaponAsDefense(spec WeaponSpec) Defense {
	return DesignDefense(DefenseSpec{
		Weapon: &spec.Model, Mount: spec.Mount, Stage: spec.Stage, Range: spec.Range,
	})
}

// defenseDevice resolves what a spec is actually installing — a screen, or a
// weapon serving as one.
func defenseDevice(spec DefenseSpec) (device, []string) {
	if spec.Weapon != nil {
		if !validWeapon(*spec.Weapon) {
			return device{}, []string{"unknown weapon"}
		}
		w := weaponData[*spec.Weapon]
		return device{w.name, w.tl, w.cost, w.principle}, nil
	}
	if !validDefense(spec.Model) {
		return device{}, []string{"unknown defense"}
	}
	d := defenseData[spec.Model]
	return device{d.name, d.tl, d.cost, d.principle}, nil
}

// Name is the defense's name with its installed tech level, e.g. "Black Globe-16".
func (d Defense) Name() string {
	if d.Device == "" {
		return "?"
	}
	return fmt.Sprintf("%s-%d", d.Device, d.TL)
}

// LongName renders the defense's identity as the book's own tables do (Book 2
// p.176 "Identifying Defenses"):
//
//	Standard Vdistant Bolt-In Nuclear Damper-12 Mod=+3. 3 tons. MCr4. R=07. (Electronic).
func (d Defense) LongName() string {
	if d.Device == "" || !validDefenseMount(d.Spec.Mount) || !validRange(d.Spec.Range) {
		return "?"
	}
	return fmt.Sprintf("%s %s %s %s Mod=%+d. %s. %s. R=%02d. (%s).",
		d.Spec.Stage, rangeData[d.Spec.Range].name, mountName(d.Spec.Mount),
		d.Name(), d.Mod, d.Tons.Phrase(), weaponMCr(d.Cost), d.Band, d.Principle)
}

func validDefense(id DefenseID) bool { return id >= 0 && int(id) < len(defenseData) }

// validDefenseMount accepts every weapon mount plus the Bolt-In. Unlike a weapon,
// a defense has no minimum mount to meet: the Bolt-In is its natural home, and
// the others "have advantages and are sometimes used" (Book 2 p.176).
func validDefenseMount(m Mount) bool { return m >= 0 && int(m) < len(defenseMountData) }

// DefenseByName finds a defense by name or single-letter code, case-insensitively
// and ignoring spaces ("black globe", "blackglobe"). The letters are not unique
// across defenses (the Grav Scrambler and the Jammer are both J), so a name is the
// reliable key and a letter matches the first defense that carries it.
func DefenseByName(s string) (DefenseID, bool) {
	norm := squash(s)
	for id, d := range defenseData {
		if squash(d.name) == norm {
			return DefenseID(id), true
		}
	}
	if len(s) == 1 {
		for id, d := range defenseData {
			if upper(s[0]) == d.letter {
				return DefenseID(id), true
			}
		}
	}
	return 0, false
}

// DefenseNames lists every defense's name, for a "known defenses" message.
func DefenseNames() []string {
	names := make([]string, len(defenseData))
	for i, d := range defenseData {
		names[i] = d.name
	}
	return names
}

// DefaultDefense is a defense as the book lists it: in a Bolt-In at Vdistant,
// Standard stage.
func DefaultDefense(model DefenseID) DefenseSpec {
	return DefenseSpec{Model: model, Mount: BoltIn, Range: VDistant}
}
