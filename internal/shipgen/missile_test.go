package shipgen

import (
	"strings"
	"testing"
)

// launcher designs a Missile launcher at a given tech level, by picking the range
// that shifts its base TL-7 up to it. The round's tech level is the launcher's.
func launcherAtTL(model WeaponID, stage Stage, rng Range) Weapon {
	return DesignWeapon(WeaponSpec{model, weaponData[model].minMount, stage, rng})
}

// TestIdentifyingMissiles is the golden for the missile engine: the book's own
// IDENTIFYING MISSILES catalog (Book 2 p.170), an Advanced Missile-10 launcher
// throwing every round it can.
//
// The catalog's *guidance* clauses derive exactly, and they are the interesting
// part: a Size-4 round at TL 10 can be Operator-Guided or HardWired, while a
// Size-5 round could also hold a Self-Aware or DownLoaded brain — except that
// those need TL 11 and TL 13, so at TL 10 the book prints them as "not available".
// That whole sentence falls out of two minimum columns.
//
// The catalog's EMP *effects*, though, print one lower than the effects grid on
// the same page (see missileEffect). We follow the grid, so the EMP rows below say
// EMP= 4 and EMP= 5 where the catalog says 3 and 4.
func TestIdentifyingMissiles(t *testing.T) {
	// Advanced shifts TL +3, so a base-TL-7 Missile launcher is TL 10.
	l := launcherAtTL(MissileLauncher, Advanced, AttackRange)
	if l.TL != 10 {
		t.Fatalf("launcher TL = %d, want 10", l.TL)
	}

	cases := []struct {
		size int
		typ  MissileType
		want string
	}{
		// Size 4: OG and HW are both within reach at TL 10.
		{4, EMP, "Advanced Missile-10 Size-4 EMP= 4. Guidance: OG HW"},
		{4, Explosive, "Advanced Missile-10 Size-4 Pen= 4. Guidance: OG HW"},
		{4, Decoy, "Advanced Missile-10 Size-4 Decoy. Guidance: OG HW"},
		{4, SensorPkg, "Advanced Missile-10 Size-4 Sensor. Guidance: OG HW"},
		// Size 5: the round is big enough for a Self-Aware or DownLoaded brain,
		// but TL 10 cannot build one — exactly as the book says.
		{5, EMP, "Advanced Missile-10 Size-5 EMP= 5. Guidance: OG HW (SA DL not available)"},
		{5, Nuke, "Advanced Missile-10 Size-5 ME. Guidance: OG HW (SA DL not available)"},
		{5, Explosive, "Advanced Missile-10 Size-5 Pen= 5. Guidance: OG HW (SA DL not available)"},
		{5, Decoy, "Advanced Missile-10 Size-5 Decoy. Guidance: OG HW (SA DL not available)"},
	}
	for _, c := range cases {
		m := DesignMissile(l, MissileSpec{Size: c.size, Type: c.typ, Guidance: OperatorGuided})
		if len(m.Problems) > 0 {
			t.Errorf("Size-%d %s: unexpected problems %v", c.size, c.typ, m.Problems)
		}

		if got := m.LongName(); got != c.want {
			t.Errorf("LongName mismatch\n got: %s\nwant: %s", got, c.want)
		}
	}
}

// TestMissileTypeIsItsLauncher: the Type in a missile's LongName is the launcher's
// own weapon identifier, not the literal word "Missile" (Book 2 p.155: "Type is the
// Weapon Identifier (which is Missile), suffixed by the Tech Level of the Launcher"
// — the p.170 example reads "Missile" because that IS the M launcher's name).
//
// The book prints the other launchers the same way, which is the golden here:
//
//	Std KK Missile-11 Size-7 Kinetic Pen=7xSp. SA
//
// (our guidance clause is the derived "Guidance: SA (DL not available)" form the
// rest of the catalog uses — DownLoaded needs TL 13.)
func TestMissileTypeIsItsLauncher(t *testing.T) {
	// A KK Missile launcher is base TL 10; Long Range shifts it to 11, the book's.
	kk := launcherAtTL(KKMissile, Standard, LongRange)
	if kk.TL != 11 {
		t.Fatalf("KK launcher TL = %d, want 11", kk.TL)
	}

	m := DesignMissile(kk, MissileSpec{Size: 7, Type: Kinetic, Guidance: SelfAware})
	if len(m.Problems) > 0 {
		t.Errorf("a Size-7 Self-Aware kinetic round at TL 11 is legal: %v", m.Problems)
	}

	want := "Standard KK Missile-11 Size-7 Pen= 7xSp. Guidance: SA (DL not available)"
	if got := m.LongName(); got != want {
		t.Errorf("LongName mismatch\n got: %s\nwant: %s", got, want)
	}
	// And a launcher the book prints no catalog for still names itself, rather than
	// claiming to throw a Missile.
	sl := launcherAtTL(SlugThrower, Standard, VDistant)
	if got := DesignMissile(
		sl, MissileSpec{Size: 2, Type: Slug, Guidance: UnGuided},
	).LongName(); got != "Standard Slug Thrower-9 Size-2 Pen= 1." {
		t.Errorf("Slug Thrower round LongName = %q", got)
	}
}

// TestMissileGuidanceReach: the cleverer brains need both a big enough round to
// house them and a high enough tech level to build one (Book 2 p.170). Raise the
// launcher's tech level and the same round becomes Self-Aware.
func TestMissileGuidanceReach(t *testing.T) {
	// A TL-13 launcher (Ultimate shifts +4, plus Long Range +1, from base 7... use
	// a range/stage pair that lands on 13).
	hi := launcherAtTL(MissileLauncher, Ultimate, LongRange) // 7 + 1 + 4 = 12
	if hi.TL != 12 {
		t.Fatalf("launcher TL = %d, want 12", hi.TL)
	}
	// At TL 12 a Size-5 round can be Self-Aware (needs 11) but not DownLoaded
	// (needs 13).
	m := DesignMissile(hi, MissileSpec{Size: 5, Type: Explosive, Guidance: SelfAware})
	if len(m.Problems) > 0 {
		t.Errorf("a Size-5 Self-Aware round at TL 12 is legal: %v", m.Problems)
	}

	if !strings.Contains(m.LongName(), "Guidance: OG HW SA (DL not available)") {
		t.Errorf("TL-12 Size-5 guidance = %q", m.LongName())
	}
	// DownLoaded needs TL 13.
	if dl := DesignMissile(
		hi,
		MissileSpec{Size: 5, Type: Explosive, Guidance: DownLoaded},
	); len(
		dl.Problems,
	) == 0 {
		t.Errorf("DownLoaded guidance needs TL 13; a TL-12 launcher cannot build one")
	}
	// And Self-Aware needs Size 5 — a Size-4 round has nowhere to put the brain,
	// which is a size limit, not a tech one.
	if small := DesignMissile(
		hi,
		MissileSpec{Size: 4, Type: Explosive, Guidance: SelfAware},
	); len(
		small.Problems,
	) == 0 {
		t.Errorf("a Size-4 round is too small for a Self-Aware brain")
	}
}

// TestMissileLauncherLimits: a launcher throws the sizes and warheads its own row
// allows, and nothing else.
func TestMissileLauncherLimits(t *testing.T) {
	missile := launcherAtTL(MissileLauncher, Standard, AttackRange) // TL 7
	// The Missile launcher throws Sizes 4, 5, and 6 — not 3 or 7.
	for _, size := range []int{4, 5, 6} {
		if m := DesignMissile(
			missile,
			MissileSpec{Size: size, Type: Explosive, Guidance: HardWired},
		); len(
			m.Problems,
		) > 0 {
			t.Errorf("a Missile launcher throws Size-%d: %v", size, m.Problems)
		}
	}

	for _, size := range []int{3, 7} {
		if m := DesignMissile(
			missile,
			MissileSpec{Size: size, Type: Explosive, Guidance: HardWired},
		); len(
			m.Problems,
		) == 0 {
			t.Errorf("a Missile launcher should not throw Size-%d", size)
		}
	}
	// A Nuke is a Size-5 warhead. A Size-4 round cannot carry one.
	if m := DesignMissile(
		missile,
		MissileSpec{Size: 4, Type: Nuke, Guidance: HardWired},
	); len(
		m.Problems,
	) == 0 {
		t.Errorf("a Size-4 round carries no Nuke")
	}
	// The KK Missile throws only its 100-ton Size-7 kinetic round.
	kk := launcherAtTL(KKMissile, Standard, AttackRange)

	m := DesignMissile(kk, MissileSpec{Size: 7, Type: Kinetic, Guidance: SelfAware})
	if m.TonsEach != 100 {
		t.Errorf("a KK round is %d tons each, want 100", m.TonsEach)
	}
	// Kinetic damage is the round's speed, not a fixed number.
	if !strings.Contains(m.Effect, "7xSp") {
		t.Errorf("Size-7 kinetic effect = %q, want Pen= 7xSp", m.Effect)
	}
}

// TestMissileMagazine: the smaller rounds are counted by the ton, the larger ones
// weigh tons apiece (Book 2 p.170) — which is what a magazine is budgeted from.
func TestMissileMagazine(t *testing.T) {
	rack := launcherAtTL(SalvoRack, Standard, VDistant)
	if m := DesignMissile(
		rack,
		MissileSpec{Size: 3, Type: Explosive, Guidance: HardWired},
	); m.PerTon != 3000 {
		t.Errorf("Vsmall Missiles are %d/ton, want 3000", m.PerTon)
	}

	missile := launcherAtTL(MissileLauncher, Standard, AttackRange)
	if m := DesignMissile(
		missile,
		MissileSpec{Size: 4, Type: Explosive, Guidance: HardWired},
	); m.PerTon != 400 {
		t.Errorf("Small Missiles are %d/ton, want 400", m.PerTon)
	}

	if m := DesignMissile(
		missile,
		MissileSpec{Size: 5, Type: Explosive, Guidance: HardWired},
	); m.PerTon != 50 {
		t.Errorf("Missiles are %d/ton, want 50", m.PerTon)
	}

	if m := DesignMissile(
		missile,
		MissileSpec{Size: 6, Type: Explosive, Guidance: HardWired},
	); m.TonsEach != 5 {
		t.Errorf("a Small Craft round is %d tons, want 5", m.TonsEach)
	}
}

// TestGuidanceReachTable locks the design constraints each brain carries (Book 2
// p.170): the smallest round that can house it, and the lowest tech level that can
// build one. What a brain is WORTH in a fight is shipcombat's rule, and is tested
// there.
func TestGuidanceReachTable(t *testing.T) {
	cases := []struct {
		g        Guidance
		size, tl int
		want     bool
	}{
		{UnGuided, 1, 0, true},   // needs nothing; it is the lack of a brain
		{HardWired, 1, 5, true},  // TL 5
		{HardWired, 1, 4, false}, // ...and not before
		{OperatorGuided, 4, 9, true},
		{OperatorGuided, 3, 9, false}, // too small to house it
		{OperatorGuided, 4, 8, false}, // too early to build it
		{SelfAware, 5, 11, true},
		{SelfAware, 4, 11, false},
		{DownLoaded, 5, 13, true},
		{DownLoaded, 5, 12, false},
	}
	for _, c := range cases {
		if got := c.g.Available(c.size, c.tl); got != c.want {
			t.Errorf("%s at Size %d TL %d = %v, want %v", c.g, c.size, c.tl, got, c.want)
		}
	}
}
