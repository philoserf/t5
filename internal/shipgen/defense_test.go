package shipgen

import "testing"

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
		{DefenseSpec{NuclearDamper, BoltIn, Standard, VDistant},
			"Standard Vdistant Bolt-In Nuclear Damper-12 Mod=+3. 3 tons. MCr4. R=07. (Electronic)."},
		{DefenseSpec{MesonScreen, BoltIn, Standard, VDistant},
			"Standard Vdistant Bolt-In Meson Screen-13 Mod=+3. 3 tons. MCr6. R=07. (Electronic)."},
		{DefenseSpec{MagScrambler, BoltIn, Standard, VDistant},
			"Standard Vdistant Bolt-In Mag Scrambler-14 Mod=+3. 3 tons. MCr4. R=07. (Magnetic)."},
		{DefenseSpec{GravScrambler, BoltIn, Standard, VDistant},
			"Standard Vdistant Bolt-In Grav Scrambler-17 Mod=+3. 3 tons. MCr5. R=07. (Gravitic)."},
		{DefenseSpec{ElecScrambler, BoltIn, Standard, VDistant},
			"Standard Vdistant Bolt-In Elec Scrambler-12 Mod=+3. 3 tons. MCr5. R=07. (Electronic)."},
		{DefenseSpec{ProtonScreen, BoltIn, Standard, VDistant},
			"Standard Vdistant Bolt-In Proton Screen-19 Mod=+3. 3 tons. MCr4. R=07. (Electronic)."},
		{DefenseSpec{BlackGlobe, BoltIn, Standard, VDistant},
			"Standard Vdistant Bolt-In Black Globe-16 Mod=+3. 3 tons. MCr13. R=07. (Electronic)."},
		{DefenseSpec{StasisGlobe, BoltIn, Standard, VDistant},
			"Standard Vdistant Bolt-In Stasis Globe-24 Mod=+3. 3 tons. MCr13. R=07. (Electronic)."},

		// The varied rows. Each one pins something down.
		//
		// x0.33, not /3: a 3-ton Bolt-In at Vlong is 0.99 tons, and MCr3 + 0.99.
		{DefenseSpec{MesonScreen, BoltIn, Standard, Vlong},
			"Standard Vlong Bolt-In Meson Screen-11 Mod=+3. 0.99 tons. MCr3.99. R=05. (Electronic)."},
		{DefenseSpec{BlackGlobe, BoltIn, Basic, Vlong},
			"Basic Vlong Bolt-In Black Globe-14 Mod=+3. 0.99 tons. MCr5.99. R=05. (Electronic)."},
		// ...and a 200-ton Main at Vlong is 66 tons, not 66.67.
		{DefenseSpec{ProtonScreen, Main, Generic, Vlong},
			"Generic Vlong Main Proton Screen-18 Mod=+1. 66 tons. MCr7.1. R=05. (Electronic)."},
		// The Quad Turret is MCr2.5: MCr10 globe at Early (x2) = 20, plus 2.5.
		{DefenseSpec{WhiteGlobe, QuadTurret, Early, VDistant},
			"Early Vdistant Quad Turret White Globe-19 Mod=+4. 1 ton. MCr22.5. R=07. (Electronic)."},
		{DefenseSpec{ProtonScreen, QuadTurret, Generic, VDistant},
			"Generic Vdistant Quad Turret Proton Screen-20 Mod=+4. 1 ton. MCr3. R=07. (Electronic)."},
		{DefenseSpec{SilverGlobe, QuadTurret, Advanced, Distant},
			"Advanced Distant Quad Turret Silver Globe-24 Mod=+4. 0.5 tons. MCr21.25. R=06. (Electronic)."},
		// The stage's Mod is not applied: Improved would add +1, Ultimate +4.
		{DefenseSpec{MesonScreen, TripleTurret, Improved, Vlong},
			"Improved Vlong Triple Turret Meson Screen-12 Mod=+3. 0.33 tons. MCr3.33. R=05. (Electronic)."},
		{DefenseSpec{BlackGlobe, Bay, Ultimate, Distant},
			"Ultimate Distant Bay Black Globe-19 Mod=+1. 25 tons. MCr32.5. R=06. (Electronic)."},
		// The rest of the stage cost ladder: Prototype x5, Modified /2, Basic /2.
		{DefenseSpec{WhiteGlobe, TripleTurret, Prototype, Distant},
			"Prototype Distant Triple Turret White Globe-17 Mod=+3. 0.5 tons. MCr50.5. R=06. (Electronic)."},
		{DefenseSpec{GravScrambler, LargeBay, Prototype, VDistant},
			"Prototype Vdistant Large Bay Grav Scrambler-15 Mod=+1. 100 tons. MCr20. R=07. (Gravitic)."},
		{DefenseSpec{WhiteGlobe, DualBarbette, Modified, Distant},
			"Modified Distant Dual Barbette White Globe-21 Mod=+2. 2.5 tons. MCr7. R=06. (Electronic)."},
		{DefenseSpec{ElecScrambler, TripleTurret, Basic, Vlong},
			"Basic Vlong Triple Turret Elec Scrambler-10 Mod=+3. 0.33 tons. MCr1.33. R=05. (Electronic)."},
		{DefenseSpec{MagScrambler, BoltIn, Experimental, Distant},
			"Experimental Distant Bolt-In Mag Scrambler-10 Mod=+3. 1.5 tons. MCr11.5. R=06. (Magnetic)."},
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
// level and price but takes the defenses' mount table — and loses its attack Mod,
// the Beam Laser's +2 included, because it is no longer making an attack
// (Book 2 pp.176, 186).
func TestWeaponsAsDefenses(t *testing.T) {
	cases := []struct {
		spec WeaponSpec
		want string
	}{
		{WeaponSpec{FusionGun, SingleBarbette, Standard, VDistant},
			"Standard Vdistant Barbette Fusion Gun-12 Mod=+1. 3 tons. MCr4.5. R=07. (Electronic)."},
		{WeaponSpec{PulseLaser, SingleTurret, Standard, VDistant},
			"Standard Vdistant Single Turret Pulse Laser-9 Mod=+1. 1 ton. MCr0.5. R=07. (Electronic)."},
		{WeaponSpec{PlasmaGun, SingleBarbette, Standard, VDistant},
			"Standard Vdistant Barbette Plasma Gun-11 Mod=+1. 3 tons. MCr4. R=07. (Electronic)."},
		{WeaponSpec{SandCaster, SingleTurret, Standard, VDistant},
			"Standard Vdistant Single Turret SandCaster-9 Mod=+1. 1 ton. MCr0.3. R=07. (Electronic)."},
		// The Beam Laser attacks at Mod +2 over its mount; defending, it gets the
		// Single Turret's +1 like anything else.
		{WeaponSpec{BeamLaser, SingleTurret, Standard, VDistant},
			"Standard Vdistant Single Turret Beam Laser-10 Mod=+1. 1 ton. MCr0.7. R=07. (Electronic)."},
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
	if attack := DesignWeapon(WeaponSpec{BeamLaser, SingleTurret, Standard, VDistant}); attack.Mod != 0 {
		t.Errorf("attacking Beam Laser Mod = %+d, want +0", attack.Mod)
	}
	// A weapon still needs a mount big enough to hold it.
	if d := DesignWeaponAsDefense(WeaponSpec{FusionGun, SingleTurret, Standard, VDistant}); len(d.Problems) == 0 {
		t.Errorf("a Fusion Gun needs a Barbette, defending or not")
	}
}

// TestDefenseRangeLimit: a defense reaches at most Vdistant. Book 2 p.174 greys
// out Orbit, Far, and Geo on the defenses' copy of the World range table, and the
// Defense Ranges table (p.179) stops at R=7 — a defense can be built for less
// reach, never for more.
func TestDefenseRangeLimit(t *testing.T) {
	for _, r := range []Range{Orbit, Far, Geo} {
		d := DesignDefense(DefenseSpec{BlackGlobe, BoltIn, Standard, r})
		if len(d.Problems) == 0 {
			t.Errorf("a defense built for %s should be reported", rangeData[r].name)
		}
	}
	for _, r := range []Range{Vlong, Distant, VDistant} {
		if d := DesignDefense(DefenseSpec{BlackGlobe, BoltIn, Standard, r}); len(d.Problems) > 0 {
			t.Errorf("%s is a legal defense range: %v", rangeData[r].name, d.Problems)
		}
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
	if d := DesignDefense(DefenseSpec{Model: DefenseID(99)}); len(d.Problems) == 0 || d.LongName() != "?" {
		t.Errorf("an unknown defense should be reported, got %+v", d)
	}
	if d := DesignWeaponAsDefense(WeaponSpec{Model: WeaponID(99)}); len(d.Problems) == 0 {
		t.Errorf("an unknown weapon-as-defense should be reported")
	}
}
