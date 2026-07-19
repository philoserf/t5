package shipgen

import (
	"strings"
	"testing"
)

// TestIdentifyingDefenses is the golden for the defense engine: rows of the
// book's own IDENTIFYING DEFENSES catalog (Book 2 p.176), derived from the p.174
// design tables.
//
// This catalog is worth more than p.167's was, because it does not stay on the
// standard rung: it varies the stage and the range across almost every row, so it
// exercises the cost multipliers, the tech-level shifts, and the tonnage divisors
// that the weapons catalog never touched. Every row below reproduces exactly —
// which is what pinned down three facts the tables alone do not tell you: that the
// book divides by three by multiplying by 0.33, that a defense takes no Mod from
// its tech stage, and that the Quad Turret costs MCr2.5.
func TestIdentifyingDefenses(t *testing.T) {
	cases := []struct {
		spec DefenseSpec
		want string
	}{
		// The plain rows: every defense in its Bolt-In at the standard range.
		{
			DefenseSpec{Model: NuclearDamper, Mount: BoltIn, Stage: Standard, Range: VDistant},
			"Standard Vdistant Bolt-In Nuclear Damper-12 Mod=+3. 3 tons. MCr4. R=07. (Electronic).",
		},
		{
			DefenseSpec{Model: MesonScreen, Mount: BoltIn, Stage: Standard, Range: VDistant},
			"Standard Vdistant Bolt-In Meson Screen-13 Mod=+3. 3 tons. MCr6. R=07. (Electronic).",
		},
		{
			DefenseSpec{Model: MagScrambler, Mount: BoltIn, Stage: Standard, Range: VDistant},
			"Standard Vdistant Bolt-In Mag Scrambler-14 Mod=+3. 3 tons. MCr4. R=07. (Magnetic).",
		},
		{
			DefenseSpec{Model: GravScrambler, Mount: BoltIn, Stage: Standard, Range: VDistant},
			"Standard Vdistant Bolt-In Grav Scrambler-17 Mod=+3. 3 tons. MCr5. R=07. (Gravitic).",
		},
		{
			DefenseSpec{Model: ElecScrambler, Mount: BoltIn, Stage: Standard, Range: VDistant},
			"Standard Vdistant Bolt-In Elec Scrambler-12 Mod=+3. 3 tons. MCr5. R=07. (Electronic).",
		},
		{
			DefenseSpec{Model: ProtonScreen, Mount: BoltIn, Stage: Standard, Range: VDistant},
			"Standard Vdistant Bolt-In Proton Screen-19 Mod=+3. 3 tons. MCr4. R=07. (Electronic).",
		},
		{
			DefenseSpec{Model: BlackGlobe, Mount: BoltIn, Stage: Standard, Range: VDistant},
			"Standard Vdistant Bolt-In Black Globe-16 Mod=+3. 3 tons. MCr13. R=07. (Electronic).",
		},
		{
			DefenseSpec{Model: StasisGlobe, Mount: BoltIn, Stage: Standard, Range: VDistant},
			"Standard Vdistant Bolt-In Stasis Globe-24 Mod=+3. 3 tons. MCr13. R=07. (Electronic).",
		},

		// The varied rows. Each one pins something down.
		//
		// x0.33, not /3: a 3-ton Bolt-In at Vlong is 0.99 tons, and MCr3 + 0.99.
		{
			DefenseSpec{Model: MesonScreen, Mount: BoltIn, Stage: Standard, Range: Vlong},
			"Standard Vlong Bolt-In Meson Screen-11 Mod=+3. 0.99 tons. MCr3.99. R=05. (Electronic).",
		},
		{
			DefenseSpec{Model: BlackGlobe, Mount: BoltIn, Stage: Basic, Range: Vlong},
			"Basic Vlong Bolt-In Black Globe-14 Mod=+3. 0.99 tons. MCr5.99. R=05. (Electronic).",
		},
		// ...and a 200-ton Main at Vlong is 66 tons, not 66.67.
		{
			DefenseSpec{Model: ProtonScreen, Mount: Main, Stage: Generic, Range: Vlong},
			"Generic Vlong Main Proton Screen-18 Mod=+1. 66 tons. MCr7.1. R=05. (Electronic).",
		},
		// The Quad Turret is MCr2.5: MCr10 globe at Early (x2) = 20, plus 2.5.
		{
			DefenseSpec{Model: WhiteGlobe, Mount: QuadTurret, Stage: Early, Range: VDistant},
			"Early Vdistant Quad Turret White Globe-19 Mod=+4. 1 ton. MCr22.5. R=07. (Electronic).",
		},
		{
			DefenseSpec{Model: ProtonScreen, Mount: QuadTurret, Stage: Generic, Range: VDistant},
			"Generic Vdistant Quad Turret Proton Screen-20 Mod=+4. 1 ton. MCr3. R=07. (Electronic).",
		},
		{
			DefenseSpec{Model: SilverGlobe, Mount: QuadTurret, Stage: Advanced, Range: Distant},
			"Advanced Distant Quad Turret Silver Globe-24 Mod=+4. 0.5 tons. MCr21.25. R=06. (Electronic).",
		},
		// The stage's Mod is not applied: Improved would add +1, Ultimate +4.
		{
			DefenseSpec{Model: MesonScreen, Mount: TripleTurret, Stage: Improved, Range: Vlong},
			"Improved Vlong Triple Turret Meson Screen-12 Mod=+3. 0.33 tons. MCr3.33. R=05. (Electronic).",
		},
		{
			DefenseSpec{Model: BlackGlobe, Mount: Bay, Stage: Ultimate, Range: Distant},
			"Ultimate Distant Bay Black Globe-19 Mod=+1. 25 tons. MCr32.5. R=06. (Electronic).",
		},
		// The rest of the stage cost ladder: Prototype x5, Modified /2, Basic /2.
		{
			DefenseSpec{Model: WhiteGlobe, Mount: TripleTurret, Stage: Prototype, Range: Distant},
			"Prototype Distant Triple Turret White Globe-17 Mod=+3. 0.5 tons. MCr50.5. R=06. (Electronic).",
		},
		{
			DefenseSpec{Model: GravScrambler, Mount: LargeBay, Stage: Prototype, Range: VDistant},
			"Prototype Vdistant Large Bay Grav Scrambler-15 Mod=+1. 100 tons. MCr20. R=07. (Gravitic).",
		},
		{
			DefenseSpec{Model: WhiteGlobe, Mount: DualBarbette, Stage: Modified, Range: Distant},
			"Modified Distant Dual Barbette White Globe-21 Mod=+2. 2.5 tons. MCr7. R=06. (Electronic).",
		},
		{
			DefenseSpec{Model: ElecScrambler, Mount: TripleTurret, Stage: Basic, Range: Vlong},
			"Basic Vlong Triple Turret Elec Scrambler-10 Mod=+3. 0.33 tons. MCr1.33. R=05. (Electronic).",
		},
		{
			DefenseSpec{Model: MagScrambler, Mount: BoltIn, Stage: Experimental, Range: Distant},
			"Experimental Distant Bolt-In Mag Scrambler-10 Mod=+3. 1.5 tons. MCr11.5. R=06. (Magnetic).",
		},
	}
	for _, c := range cases {
		d := DesignDefense(c.spec)
		if len(d.Problems) > 0 {
			t.Errorf("%s: unexpected problems %v", d.Name(), d.Problems)
		}

		if got := d.LongName(); got != c.want {
			t.Errorf("LongName mismatch\n got: %s\nwant: %s", got, c.want)
		}
	}
}

// TestWeaponsAsDefenses: a weapon allocated to Defensive Fire keeps its own tech
// level, price, and principle, but takes the defenses' mount table — and loses its
// attack Mod, the Beam Laser's +2 included, because it is no longer making an
// attack (Book 2 pp.176, 186).
//
// The principle is the weapon's own, and a weapon may rest on more than one: the
// book prints the Fusion and Plasma Guns as "(Elec/Grav)", not "(Electronic)".
func TestWeaponsAsDefenses(t *testing.T) {
	cases := []struct {
		spec WeaponSpec
		want string
	}{
		{
			WeaponSpec{FusionGun, SingleBarbette, Standard, VDistant},
			"Standard Vdistant Barbette Fusion Gun-12 Mod=+1. 3 tons. MCr4.5. R=07. (Elec/Grav).",
		},
		{
			WeaponSpec{PulseLaser, SingleTurret, Standard, VDistant},
			"Standard Vdistant Single Turret Pulse Laser-9 Mod=+1. 1 ton. MCr0.5. R=07. (Electronic).",
		},
		{
			WeaponSpec{PlasmaGun, SingleBarbette, Standard, VDistant},
			"Standard Vdistant Barbette Plasma Gun-11 Mod=+1. 3 tons. MCr4. R=07. (Elec/Grav).",
		},
		{
			WeaponSpec{SandCaster, SingleTurret, Standard, VDistant},
			"Standard Vdistant Single Turret SandCaster-9 Mod=+1. 1 ton. MCr0.3. R=07. (Electronic).",
		},
		// The Beam Laser attacks at Mod +2 over its mount; defending, it gets the
		// Single Turret's +1 like anything else.
		{
			WeaponSpec{BeamLaser, SingleTurret, Standard, VDistant},
			"Standard Vdistant Single Turret Beam Laser-10 Mod=+1. 1 ton. MCr0.7. R=07. (Electronic).",
		},
	}
	for _, c := range cases {
		d := DesignWeaponAsDefense(c.spec)
		if len(d.Problems) > 0 {
			t.Errorf("%s: unexpected problems %v", d.Name(), d.Problems)
		}

		if got := d.LongName(); got != c.want {
			t.Errorf("LongName mismatch\n got: %s\nwant: %s", got, c.want)
		}
	}

	// Attacking, the same laser is Mod +0 (turret -2, laser +2) — the two modes
	// genuinely differ.
	if attack := DesignWeapon(
		WeaponSpec{BeamLaser, SingleTurret, Standard, VDistant},
	); attack.Mod != 0 {
		t.Errorf("attacking Beam Laser Mod = %+d, want +0", attack.Mod)
	}
	// A weapon still needs a mount big enough to hold it.
	if d := DesignWeaponAsDefense(
		WeaponSpec{FusionGun, SingleTurret, Standard, VDistant},
	); len(
		d.Problems,
	) == 0 {
		t.Errorf("a Fusion Gun needs a Barbette, defending or not")
	}
}

// TestDefenseRangeLimit: a defense stands at the standard rung of its ladder and
// may be built for less reach, never for more — "Defense Range can be decreased
// but not increased" (Book 2 p.177), where "The Standard Defense Range is R=7".
// On the world ladder that rules out Orbit, Far, and Geo.
func TestDefenseRangeLimit(t *testing.T) {
	for _, r := range []Range{Orbit, Far, Geo} {
		d := DesignDefense(DefenseSpec{Model: BlackGlobe, Mount: BoltIn, Stage: Standard, Range: r})
		if len(d.Problems) == 0 {
			t.Errorf("a defense built for %s should be reported", rangeData[r].name)
		}
	}

	for _, r := range []Range{Vlong, Distant, VDistant} {
		if d := DesignDefense(
			DefenseSpec{Model: BlackGlobe, Mount: BoltIn, Stage: Standard, Range: r},
		); len(
			d.Problems,
		) > 0 {
			t.Errorf("%s is a legal defense range: %v", rangeData[r].name, d.Problems)
		}
	}
}

// TestDefenseSpaceRangeLimit: the same "decreased but not increased" rule caps the
// space ladder at its own standard rung, S=7 Attack Range (Book 2 p.177; the space
// table is p.83 Table D, printed again for defenses on p.174). The screens never
// reach this check — they are all world-range devices — but the dual-scale Hybrid
// S-L-M, which p.174 marks "S=7* R=7*" and lists among the Weapons As Defenses,
// does. Before this rule was enforced it designed clean at S=12, a reach the code's
// own comment said no defense has.
func TestDefenseSpaceRangeLimit(t *testing.T) {
	for _, r := range []Range{Boarding, FighterRange, ShortRange, AttackRange} {
		if d := DesignWeaponAsDefense(
			WeaponSpec{HybridSLM, SingleTurret, Standard, r},
		); len(d.Problems) > 0 {
			t.Errorf("%s is at or below the standard rung and is legal: %v",
				rangeData[r].name, d.Problems)
		}
	}

	for _, r := range []Range{LongRange, DeepSpace} {
		if d := DesignWeaponAsDefense(
			WeaponSpec{HybridSLM, SingleTurret, Standard, r},
		); len(d.Problems) == 0 {
			t.Errorf("a defense built for %s reaches past S=7 and should be reported, got a clean %s",
				rangeData[r].name, d.LongName())
		}
	}
	// Attacking, the same weapon at the same range is perfectly legal — the cap is
	// the defense's, not the Hybrid's.
	if w := DesignWeapon(
		WeaponSpec{HybridSLM, SingleTurret, Standard, DeepSpace},
	); len(w.Problems) > 0 {
		t.Errorf("a Hybrid SLM attacking at Deep Space is legal: %v", w.Problems)
	}
}

// TestDefenseRangeCodeScale: a defense built on the space ladder prints S=, not R=.
// The two ladders are distinct (Book 2 p.83 Tables D and E; p.174 prints both for
// defenses), and p.175 says the band "duplicates the Defense name Range information"
// — so an S=07 Attack Range installation labelled R=07 would claim a 50 km reach for
// a weapon covering the whole Attack Range band. Screens never reach this, being
// world-range devices; the dual-scale Hybrid S-L-M does.
func TestDefenseRangeCodeScale(t *testing.T) {
	space := DesignWeaponAsDefense(WeaponSpec{HybridSLM, SingleTurret, Standard, AttackRange})
	if len(space.Problems) > 0 {
		t.Fatalf("a Hybrid SLM defending at Attack Range is legal: %v", space.Problems)
	}

	if space.Scale != SpaceScale || space.RangeCode() != "S=07" {
		t.Errorf("space-range defense band = %q, want S=07", space.RangeCode())
	}

	if !strings.HasSuffix(space.LongName(), "S=07. (Electronic).") {
		t.Errorf("space-range defense LongName = %q, want it to end S=07. (Electronic).",
			space.LongName())
	}
	// The same weapon on the world ladder still prints R=, as the whole p.176
	// catalog does.
	world := DesignWeaponAsDefense(WeaponSpec{HybridSLM, SingleTurret, Standard, VDistant})
	if world.Scale != WorldScale || world.RangeCode() != "R=07" {
		t.Errorf("world-range defense band = %q, want R=07", world.RangeCode())
	}
}

// TestNoWeaponIsBoltIn: the Bolt-In is the defenses' mount. A weapon needs to see
// out of the hull, so it cannot be bolted anywhere inside it.
func TestNoWeaponIsBoltIn(t *testing.T) {
	if w := DesignWeapon(WeaponSpec{BeamLaser, BoltIn, Standard, VDistant}); len(w.Problems) == 0 {
		t.Errorf("a weapon in a Bolt-In mount should be reported")
	}
}

// TestDefenseProblems: DesignDefense is total, like the rest of the package.
func TestDefenseProblems(t *testing.T) {
	if d := DesignDefense(
		DefenseSpec{Model: DefenseID(99)},
	); len(d.Problems) == 0 ||
		d.LongName() != "?" {
		t.Errorf("an unknown defense should be reported, got %+v", d)
	}

	if d := DesignWeaponAsDefense(WeaponSpec{Model: WeaponID(99)}); len(d.Problems) == 0 {
		t.Errorf("an unknown weapon-as-defense should be reported")
	}
}

// TestRefusedDefenseKeepsItsIdentity: a weapon that cannot serve as a defense is
// still that weapon, and the ship card has to say which one it refused. The device
// name survives the refusal; only a device the package cannot name at all is "?".
// It is named WITHOUT a tech level and without the rest of the line, because install
// never ran — the zeros would be fabrications, not facts.
func TestRefusedDefenseKeepsItsIdentity(t *testing.T) {
	d := DesignWeaponAsDefense(WeaponSpec{MesonGun, Main, Standard, VDistant})
	if len(d.Problems) == 0 {
		t.Fatalf("a Meson Gun is not point defence; it should be refused")
	}

	if d.Name() != "Meson Gun" {
		t.Errorf("refused defense Name = %q, want %q", d.Name(), "Meson Gun")
	}

	if d.LongName() != "Meson Gun" {
		t.Errorf("refused defense LongName = %q, want %q", d.LongName(), "Meson Gun")
	}
	// A device the package cannot name at all still reads "?".
	if u := DesignWeaponAsDefense(WeaponSpec{Model: WeaponID(99)}); u.Name() != "?" {
		t.Errorf("unknown weapon-as-defense Name = %q, want ?", u.Name())
	}

	if u := DesignDefense(DefenseSpec{Model: DefenseID(99)}); u.Name() != "?" {
		t.Errorf("unknown defense Name = %q, want ?", u.Name())
	}
}

// TestDefenseScale: a defense reaches on the world ladder. Without a scale check a
// screen could be built for a Space range and would print an R= band from the wrong
// ladder — "R=12" for a Black Globe at Deep Space, a reach the book says no defense
// has.
func TestDefenseScale(t *testing.T) {
	for _, r := range []Range{Boarding, FighterRange, ShortRange, AttackRange, LongRange, DeepSpace} {
		d := DesignDefense(DefenseSpec{Model: BlackGlobe, Mount: BoltIn, Range: r})
		if len(d.Problems) == 0 {
			t.Errorf("a screen built for the space range %s should be reported", rangeData[r].name)
		}
	}
}

// TestDefenseSpecZeroValue: the zero value is not a usable default — Mount's zero is
// a Single Turret and Range's is Boarding, a space range. DefaultDefense is the way
// to get the Bolt-In at Vdistant the book lists every defense in.
func TestDefenseSpecZeroValue(t *testing.T) {
	if d := DesignDefense(DefenseSpec{}); len(d.Problems) == 0 {
		t.Errorf(
			"the zero DefenseSpec is a Boarding-range turret and should be reported, got %s",
			d.LongName(),
		)
	}

	if d := DesignDefense(DefaultDefense(NuclearDamper)); len(d.Problems) > 0 {
		t.Errorf("DefaultDefense should be clean, got %v", d.Problems)
	}
}

// TestOnlyNineWeaponsDefend: Book 2 p.174 lists exactly nine weapons under "Weapons
// As Defenses". A Meson Gun is not point defence.
func TestOnlyNineWeaponsDefend(t *testing.T) {
	defenders := 0

	for id := range WeaponID(len(weaponData)) {
		d := DesignWeaponAsDefense(DefaultWeapon(id))
		if len(d.Problems) == 0 {
			defenders++
		}
	}

	if defenders != 9 {
		t.Errorf("%d weapons can serve as defenses, want the book's 9", defenders)
	}

	for _, id := range []WeaponID{MesonGun, KKMissile, Stasis} {
		if d := DesignWeaponAsDefense(DefaultWeapon(id)); len(d.Problems) == 0 {
			t.Errorf("a %s should not serve as a defense", weaponData[id].name)
		}
	}
}

// TestNoWeaponBoltsInAsADefense: a weapon has to see out of the hull to shoot,
// defending or not. Comparing mounts with "<" alone never caught this — BoltIn sorts
// ABOVE every real mount, so the minimum-mount check passed every weapon, and a
// Meson Gun could be bolted inside the hull as point defence, consuming no mount
// point at all.
func TestNoWeaponBoltsInAsADefense(t *testing.T) {
	for _, id := range []WeaponID{BeamLaser, MesonGun, PlasmaGun} {
		spec := DefaultWeapon(id)

		spec.Mount = BoltIn
		if d := DesignWeaponAsDefense(spec); len(d.Problems) == 0 {
			t.Errorf("a %s cannot be bolted inside the hull, got a clean %s",
				weaponData[id].name, d.LongName())
		}
	}
}
