package shipgen

import (
	"strings"
	"testing"
)

// TestTypicalWeapons is the golden test for the whole weapon design engine: every
// row of the book's own TYPICAL WEAPONS catalog (Book 2 p.167), derived from the
// p.83 design tables. Each row exercises the full formula — the mount's tonnage,
// dice, and Mod; the range's tech-level shift and tonnage/cost multipliers; and
// the weapon's own base cost and specials.
//
// The p.167 catalog is the fixture rather than the similar one on p.155, which
// does not derive from the design tables it summarizes: p.155 prints "Beam
// Laser-10 Mod=-2" for an Orbit-range Single Turret where the tables give TL 11
// (base 10, +1 for Orbit) and Mod 0 (the turret's -2 plus the Beam Laser's own
// +2), and it prints 1.65 tons for a Bay that the tables put at 16.67. Its
// tonnages and costs agree with p.167; its tech levels and Mods do not. We follow
// the tables, which p.167 corroborates on every row.
func TestTypicalWeapons(t *testing.T) {
	cases := []struct {
		spec WeaponSpec
		want string
	}{
		{
			WeaponSpec{MissileLauncher, SingleTurret, Standard, AttackRange},
			"Standard Attack Range Single Turret Missile-7 Mod=-2. 1 ton. MCr2.2. Hits= 1D. S=07.",
		},
		{
			WeaponSpec{MiningLaser, SingleTurret, Standard, Orbit},
			"Standard Orbit Single Turret Mining Laser-9 Mod=-2. 2 tons. MCr1.1. Hits= 1D. R=08.",
		},
		{
			WeaponSpec{CommCaster, SingleTurret, Standard, AttackRange},
			"Standard Attack Range Single Turret CommCaster-8 Mod=-2. 1 ton. MCr5.2. Hits= 1D. S=07.",
		},
		{
			WeaponSpec{PulseLaser, SingleTurret, Standard, Orbit},
			"Standard Orbit Single Turret Pulse Laser-10 Mod=-2. 2 tons. MCr0.9. Hits= 2D. R=08.",
		},
		{
			WeaponSpec{SlugThrower, SingleTurret, Standard, Orbit},
			"Standard Orbit Single Turret Slug Thrower-10 Mod=-2. 2 tons. MCr0.8. Hits= 1D. R=08.",
		},
		{
			WeaponSpec{SandCaster, SingleTurret, Standard, Orbit},
			"Standard Orbit Single Turret SandCaster-10 Mod=-2. 2 tons. MCr0.7. Hits= 1D. R=08.",
		},
		// The Beam Laser's own +2 cancels the Single Turret's -2.
		{
			WeaponSpec{BeamLaser, SingleTurret, Standard, Orbit},
			"Standard Orbit Single Turret Beam Laser-11 Mod=+0. 2 tons. MCr1.1. Hits= 1D. R=08.",
		},
		{
			WeaponSpec{SalvoRack, Bay, Standard, Orbit},
			"Standard Orbit Bay Salvo Rack-11 Mod=+5. 100 tons. MCr25. Hits= 20D. R=08.",
		},
		{
			WeaponSpec{KKMissile, Bay, Standard, AttackRange},
			"Standard Attack Range Bay KK Missile-10 Mod=+5. 50 tons. MCr8. Hits= 20D. S=07.",
		},
		{
			WeaponSpec{DataCaster, SingleTurret, Standard, Orbit},
			"Standard Orbit Single Turret DataCaster-11 Mod=-2. 2 tons. MCr1.6. Hits= 1D. R=08.",
		},
		// The Hybrid is dual-scale, and p.167 builds it on the world ladder.
		{
			WeaponSpec{HybridSLM, TripleTurret, Standard, Orbit},
			"Standard Orbit Triple Turret Hybrid SLM-11 Mod=+0. 2 tons. MCr4. Hits= 3D. R=08.",
		},
		{
			WeaponSpec{PlasmaGun, SingleBarbette, Standard, Orbit},
			"Standard Orbit Barbette Plasma Gun-12 Mod=+2. 6 tons. MCr10. Hits= 5D. R=08.",
		},
		{
			WeaponSpec{ParticleAccel, SingleBarbette, Standard, AttackRange},
			"Standard Attack Range Barbette PA-11 Mod=+2. 3 tons. MCr5.5. Hits= 5D. S=07.",
		},
		{
			WeaponSpec{FusionGun, SingleBarbette, Standard, Orbit},
			"Standard Orbit Barbette Fusion Gun-13 Mod=+2. 6 tons. MCr10.5. Hits= 5D. R=08.",
		},
		{
			WeaponSpec{RailGun, Bay, Standard, AttackRange},
			"Standard Attack Range Bay Rail Gun-12 Mod=+5. 50 tons. MCr17. Hits= 20D. S=07.",
		},
		{
			WeaponSpec{Ortillery, Bay, Standard, Orbit},
			"Standard Orbit Bay Ortillery-13 Mod=+5. 100 tons. MCr30. Hits= 20D. R=08.",
		},
		{
			WeaponSpec{MesonGun, Main, Standard, AttackRange},
			"Standard Attack Range Main Meson Gun-13 Mod=+10. 200 tons. MCr25. Hits= 100D. S=07.",
		},
		{
			WeaponSpec{JumpDamper, SingleBarbette, Standard, Orbit},
			"Standard Orbit Barbette Jump Damper-15 Mod=+2. 6 tons. MCr24. Hits= 5D. R=08.",
		},
		{
			WeaponSpec{TractorPressor, SingleBarbette, Standard, Orbit},
			"Standard Orbit Barbette Tractor Pressor-17 Mod=+2. 6 tons. MCr14. Hits= 5D. R=08.",
		},
		{
			WeaponSpec{Inducer, SingleTurret, Standard, Orbit},
			"Standard Orbit Single Turret Inducer-18 Mod=-2. 2 tons. MCr1.6. Hits= 1D. R=08.",
		},
		{
			WeaponSpec{Stasis, SingleTurret, Standard, AttackRange},
			"Standard Attack Range Single Turret Stasis-21 Mod=-2. 1 ton. MCr5.2. Hits= 1D. S=07.",
		},
	}
	for _, c := range cases {
		w := DesignWeapon(c.spec)
		if len(w.Problems) > 0 {
			t.Errorf("%s: unexpected problems %v", w.Name(), w.Problems)
		}
		if got := w.LongName(); got != c.want {
			t.Errorf("LongName mismatch\n got: %s\nwant: %s", got, c.want)
		}
	}
}

// TestWorkedWeaponExamples is a second, independent golden: the weapon writeups
// on Book 2 p.158, which spell out their own installations at Vdistant (the
// standard world range) rather than p.167's Orbit. Two catalogs, on different
// pages at different ranges, agreeing with one set of tables.
func TestWorkedWeaponExamples(t *testing.T) {
	cases := []struct {
		spec WeaponSpec
		want string
	}{
		{
			WeaponSpec{MiningLaser, SingleTurret, Standard, VDistant},
			"Standard Vdistant Single Turret Mining Laser-8 Mod=-2. 1 ton. MCr0.7. Hits= 1D. R=07.",
		},
		{
			WeaponSpec{PulseLaser, SingleTurret, Standard, VDistant},
			"Standard Vdistant Single Turret Pulse Laser-9 Mod=-2. 1 ton. MCr0.5. Hits= 2D. R=07.",
		},
		{
			WeaponSpec{BeamLaser, SingleTurret, Standard, VDistant},
			"Standard Vdistant Single Turret Beam Laser-10 Mod=+0. 1 ton. MCr0.7. Hits= 1D. R=07.",
		},
		{
			WeaponSpec{PlasmaGun, SingleBarbette, Standard, VDistant},
			"Standard Vdistant Barbette Plasma Gun-11 Mod=+2. 3 tons. MCr4. Hits= 5D. R=07.",
		},
		{
			WeaponSpec{FusionGun, SingleBarbette, Standard, VDistant},
			"Standard Vdistant Barbette Fusion Gun-12 Mod=+2. 3 tons. MCr4.5. Hits= 5D. R=07.",
		},
	}
	for _, c := range cases {
		if got := DesignWeapon(c.spec).LongName(); got != c.want {
			t.Errorf("LongName mismatch\n got: %s\nwant: %s", got, c.want)
		}
	}
}

// TestWeaponRangeEffects: the range is the design lever. Reaching further
// multiplies the mount's tonnage and cost and raises the weapon's tech level;
// building for a knife fight divides them, down to a mount that fits a FirmPoint.
func TestWeaponRangeEffects(t *testing.T) {
	// A Beam Laser (base TL 10) in a Single Turret (1 ton, MCr0.2), across the
	// world-range ladder (Book 2 p.83 Table E).
	cases := []struct {
		rng      Range
		wantTL   int
		wantTons Tonnage
		wantCost int
	}{
		// The book divides by three by multiplying by 0.33 (see rangeData).
		{Vlong, 8, 33, 500_000 + 66_000},       // -2 TL, tons x0.33, cost x0.33
		{Distant, 9, 50, 500_000 + 100_000},    // -1 TL, tons /2, cost /2
		{VDistant, 10, 100, 500_000 + 200_000}, // standard: no change
		{Orbit, 11, 200, 500_000 + 600_000},    // +1 TL, tons x2, cost x3
		{Far, 12, 300, 500_000 + 1_000_000},    // +2 TL, tons x3, cost x5
		{Geo, 13, 400, 500_000 + 1_200_000},    // +3 TL, tons x4, cost x6
	}
	for _, c := range cases {
		w := DesignWeapon(WeaponSpec{BeamLaser, SingleTurret, Standard, c.rng})
		if w.TL != c.wantTL || w.Tons != c.wantTons || w.Cost != c.wantCost {
			t.Errorf("%s Beam Laser: TL %d, %s tons, Cr%d; want TL %d, %s tons, Cr%d",
				rangeData[c.rng].name, w.TL, w.Tons, w.Cost, c.wantTL, c.wantTons, c.wantCost)
		}
	}

	// A turret built for Boarding is a quarter-ton mount — it goes on a FirmPoint,
	// and whole-ton arithmetic would have rounded it away to nothing.
	board := DesignWeapon(WeaponSpec{MissileLauncher, SingleTurret, Standard, Boarding})
	if board.Tons != 25 || !board.Tons.SubTon() || board.Tons.Ceil() != 1 {
		t.Errorf("Boarding turret = %s tons (subton %v, ceil %d), want 0.25/true/1",
			board.Tons, board.Tons.SubTon(), board.Tons.Ceil())
	}
}

// TestWeaponStageEffects: a tech stage shifts the weapon's TL, multiplies its
// cost, and modifies the attack — but never touches the mount's tonnage
// (Book 2 p.83 Table B, "Applies to Weapon (not Mount)").
func TestWeaponStageEffects(t *testing.T) {
	std := DesignWeapon(WeaponSpec{BeamLaser, SingleTurret, Standard, VDistant})
	for _, c := range []struct {
		stage   Stage
		wantTL  int
		wantMod int
	}{
		{Experimental, 7, -3 + 0}, // TL -3, Mod -3, and the Beam Laser's +2 vs turret -2
		{Prototype, 8, -2},
		{Early, 9, -1},
		{Standard, 10, 0},
		{Improved, 11, 1},
		{Modified, 12, 2},
		{Advanced, 13, 3},
		{Ultimate, 14, 4},
	} {
		w := DesignWeapon(WeaponSpec{BeamLaser, SingleTurret, c.stage, VDistant})
		if w.TL != c.wantTL {
			t.Errorf("%s Beam Laser TL = %d, want %d", c.stage, w.TL, c.wantTL)
		}
		// The turret's -2 and the laser's +2 cancel, so Mod is the stage's alone.
		if w.Mod != c.wantMod {
			t.Errorf("%s Beam Laser Mod = %+d, want %+d", c.stage, w.Mod, c.wantMod)
		}
		// The stage never changes the mount's tonnage.
		if w.Tons != std.Tons {
			t.Errorf(
				"%s changed tonnage to %s (want %s — stages apply to the weapon, not the mount)",
				c.stage,
				w.Tons,
				std.Tons,
			)
		}
	}

	// Generic raises tech level by +1 but grants no Mod — the one place the
	// weapon stage table and the drive stage ladder disagree (p.83 Table B).
	gen := DesignWeapon(WeaponSpec{BeamLaser, SingleTurret, Generic, VDistant})
	if gen.TL != 11 || gen.Mod != 0 {
		t.Errorf("Generic Beam Laser = TL %d Mod %+d, want TL 11 Mod +0", gen.TL, gen.Mod)
	}
}

// TestWeaponProblems: DesignWeapon is total. An installation the book disallows
// is reported, not refused — the designer still gets a weapon to look at.
func TestWeaponProblems(t *testing.T) {
	// A Meson Gun needs a Main mount; it does not fit in a turret.
	turretMeson := DesignWeapon(WeaponSpec{MesonGun, SingleTurret, Standard, AttackRange})
	if len(turretMeson.Problems) == 0 {
		t.Errorf("a Meson Gun in a Single Turret should be a problem")
	}
	if turretMeson.TL == 0 {
		t.Errorf("a problem installation should still be designed, got a zero weapon")
	}
	// A bigger mount than the minimum is fine (p.155: "Any Mount equal to or
	// greater than the Minimum can be selected").
	if w := DesignWeapon(WeaponSpec{MiningLaser, Bay, Standard, VDistant}); len(w.Problems) > 0 {
		t.Errorf("a Mining Laser in a Bay is legal, got %v", w.Problems)
	}
	// A world-range weapon cannot be built for a space range.
	if w := DesignWeapon(
		WeaponSpec{MiningLaser, SingleTurret, Standard, DeepSpace},
	); len(
		w.Problems,
	) == 0 {
		t.Errorf("a world-range Mining Laser built for Deep Space should be a problem")
	}
	// An unknown model does not panic.
	if w := DesignWeapon(
		WeaponSpec{Model: WeaponID(99)},
	); len(w.Problems) == 0 ||
		w.LongName() != "?" {
		t.Errorf("an unknown weapon should be reported, got %+v", w)
	}
}

// TestWeaponLetters: every weapon has the single-letter model code the book
// identifies it by, no two share one, and each builds cleanly in its own minimum
// mount at its standard range — the form the book lists it in.
func TestWeaponLetters(t *testing.T) {
	seen := map[byte]string{}
	for id := range WeaponID(len(weaponData)) {
		w := DesignWeapon(DefaultWeapon(id))
		letter := weaponData[id].letter
		if prior, dup := seen[letter]; dup {
			t.Errorf("weapons %s and %s share the letter %c", prior, w.Name(), letter)
		}
		seen[letter] = w.Name()
		if len(w.Problems) > 0 {
			t.Errorf("%s in its own minimum mount at its standard range should be legal, got %v",
				w.Name(), w.Problems)
		}
		if !strings.Contains(w.LongName(), string(rune(letter))) && w.Name() == "" {
			t.Errorf("%s has no name", w.Name())
		}
	}
	if len(seen) != 23 {
		t.Errorf("got %d distinct weapon letters, want 23 (Book 2 p.83 Table A)", len(seen))
	}
}
