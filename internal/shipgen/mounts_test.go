package shipgen

import (
	"strings"
	"testing"
)

// armedScout is the package's golden ship — the Murphy-class Scout (murphySpec,
// design_test.go) — with weapons on its mount points. A 100-ton hull has exactly
// one, which is what makes it the sharpest test of the mounting rules.
func armedScout(weapons ...WeaponSpec) Ship {
	spec := murphySpec()
	spec.Weapons = weapons
	return Design(spec)
}

// TestArmedShipBudget: a weapon's mount tonnage comes out of the payload and its
// cost goes into the ship's, so arming a ship is a real trade against cargo.
func TestArmedShipBudget(t *testing.T) {
	bare := armedScout()
	// A Beam Laser in a Single Turret at the standard world range: 1 ton, MCr0.7.
	armed := armedScout(WeaponSpec{BeamLaser, SingleTurret, Standard, VDistant})

	if len(armed.Problems) > 0 {
		t.Fatalf("a scout with one turret should be a clean design, got %v", armed.Problems)
	}
	if got, want := bare.Tonnage.Payload-armed.Tonnage.Payload, 1; got != want {
		t.Errorf("the turret cost %dt of payload, want %dt", got, want)
	}
	if got, want := armed.Cost-bare.Cost, 700_000; got != want {
		t.Errorf("the turret cost Cr%d, want Cr%d", got, want)
	}
	if len(armed.Weapons) != 1 || armed.Weapons[0].TL != 10 {
		t.Fatalf("expected one TL-10 Beam Laser, got %+v", armed.Weapons)
	}
	// The ship card shows what the ship is carrying.
	if !strings.Contains(armed.String(), "Weapons: Standard Vdistant Single Turret Beam Laser-10") {
		t.Errorf("ship card is missing its weapon:\n%s", armed)
	}
}

// TestHardpointLimit is the rule that finally makes Hull.Hardpoints load-bearing:
// one mount per 100 tons. A 100-ton scout has exactly one, so a second weapon has
// nowhere to go.
func TestHardpointLimit(t *testing.T) {
	turret := WeaponSpec{BeamLaser, SingleTurret, Standard, VDistant}
	if one := armedScout(turret); len(one.Problems) > 0 {
		t.Fatalf("one turret fits a 100t hull: %v", one.Problems)
	}
	two := armedScout(turret, turret)
	if len(two.Problems) == 0 {
		t.Fatalf("two turrets should not fit a 100t hull's single hardpoint")
	}
	if !strings.Contains(two.Problems[0], "mount blocks") {
		t.Errorf("problem should name the mount shortfall, got %q", two.Problems[0])
	}

	// A bigger hull has more. A 400-ton hull carries four.
	four := Design(ShipSpec{
		TL: 12, HullLetter: 4, Config: Streamlined, Structure: FramePlate, ArmorLayers: 1,
		Maneuver: &DriveSpec{Letter: 4}, Power: &DriveSpec{Letter: 4},
		Weapons: []WeaponSpec{turret, turret, turret, turret},
	})
	for _, p := range four.Problems {
		if strings.Contains(p, "mount blocks") {
			t.Errorf("a 400t hull has four hardpoints, but: %s", p)
		}
	}
}

// TestFirmPoints: a 100-ton block may be spent as three FirmPoints instead of one
// HardPoint, but only sub-ton mounts fit them — so a scout can carry three tiny
// boarding-range turrets where it could carry only one full one.
func TestFirmPoints(t *testing.T) {
	// A turret built for Boarding range is a quarter-ton mount.
	tiny := WeaponSpec{MissileLauncher, SingleTurret, Standard, Boarding}
	if w := DesignWeapon(tiny); !w.Tons.SubTon() {
		t.Fatalf("a Boarding-range turret should be sub-ton, got %s", w.Tons)
	}
	if three := armedScout(tiny, tiny, tiny); len(three.Problems) > 0 {
		t.Errorf("three sub-ton mounts fit one block's FirmPoints: %v", three.Problems)
	}
	// A fourth needs a second block, which a 100-ton hull does not have.
	if four := armedScout(tiny, tiny, tiny, tiny); len(four.Problems) == 0 {
		t.Errorf("a fourth sub-ton mount needs a second block a 100t hull lacks")
	}
	// And a full-ton mount cannot ride a FirmPoint, so one full weapon plus any
	// sub-ton one already needs two blocks.
	full := WeaponSpec{BeamLaser, SingleTurret, Standard, VDistant}
	if mixed := armedScout(full, tiny); len(mixed.Problems) == 0 {
		t.Errorf("a full mount takes the block's only HardPoint, leaving no FirmPoint")
	}
}

// TestBoltInNeedsNoMountPoint is the Bolt-In's whole point: it "does not require
// a HardPoint or FirmPoint" (Book 2 p.176), so a ship can carry screens without
// giving up its gun. A 100-ton scout has exactly one mount point — and can still
// have both.
func TestBoltInNeedsNoMountPoint(t *testing.T) {
	spec := murphySpec()
	spec.Weapons = []WeaponSpec{{BeamLaser, SingleTurret, Standard, VDistant}}
	spec.Defenses = []DefenseSpec{{NuclearDamper, BoltIn, Standard, VDistant}}
	s := Design(spec)
	if len(s.Problems) > 0 {
		t.Fatalf("a turret and a bolt-in screen both fit a 100t hull: %v", s.Problems)
	}
	// It is free of the mount points, not of the tonnage: the screen still weighs
	// its three tons, and the turret one.
	bare := Design(murphySpec())
	if got, want := bare.Tonnage.Payload-s.Tonnage.Payload, 4; got != want {
		t.Errorf("turret + screen cost %dt of payload, want %dt", got, want)
	}
	if !strings.Contains(s.String(), "Defenses: Standard Vdistant Bolt-In Nuclear Damper-12") {
		t.Errorf("ship card is missing its screen:\n%s", s)
	}

	// A defense in a turret, though, is a mount like any other — and now it does
	// compete with the gun for the hull's one point.
	spec.Defenses = []DefenseSpec{{NuclearDamper, SingleTurret, Standard, VDistant}}
	if turreted := Design(spec); len(turreted.Problems) == 0 {
		t.Errorf("a turreted screen competes for the hardpoint the gun is using")
	}
}

// TestWeaponAboveShipTL: a yard cannot build above its own tech level. The range
// and stage shift a weapon's TL, so a long-range model can outrun the ship.
func TestWeaponAboveShipTL(t *testing.T) {
	// A Beam Laser (base TL 10) built for Geo range is TL 13 — beyond the
	// Murphy's TL 12.
	over := armedScout(WeaponSpec{BeamLaser, SingleTurret, Standard, Geo})
	if len(over.Problems) == 0 {
		t.Fatalf("a TL-13 weapon on a TL-12 ship should be a problem")
	}
	found := false
	for _, p := range over.Problems {
		if strings.Contains(p, "above the ship's TL-12") {
			found = true
		}
	}
	if !found {
		t.Errorf("problems should name the TL overrun, got %v", over.Problems)
	}
	// The same laser at the standard range is TL 10 and builds fine.
	if ok := armedScout(WeaponSpec{BeamLaser, SingleTurret, Standard, VDistant}); len(ok.Problems) > 0 {
		t.Errorf("a TL-10 weapon on a TL-12 ship is fine, got %v", ok.Problems)
	}
}

// TestWeaponProblemsReachTheShip: a weapon that cannot be built the way it was
// asked for says so on the ship, not just on the weapon.
func TestWeaponProblemsReachTheShip(t *testing.T) {
	// A Meson Gun needs a Main mount (200t) — it does not fit in a turret, and a
	// 100-ton hull could not carry the mount anyway.
	s := armedScout(WeaponSpec{MesonGun, SingleTurret, Standard, AttackRange})
	if len(s.Problems) == 0 {
		t.Fatalf("a Meson Gun in a scout's turret should be reported")
	}
	if !strings.Contains(strings.Join(s.Problems, "; "), "Main") {
		t.Errorf("problems should name the minimum mount, got %v", s.Problems)
	}
}

// TestWeaponTonnage: full mounts are charged whole tons (a HardPoint mount "is at
// least 1 ton, round up"), while sub-ton FirmPoint mounts keep their fractions
// and round up only once, together.
func TestWeaponTonnage(t *testing.T) {
	tiny := DesignWeapon(WeaponSpec{MissileLauncher, SingleTurret, Standard, Boarding}) // 0.25t
	full := DesignWeapon(WeaponSpec{BeamLaser, SingleTurret, Standard, VDistant})
	if got := armamentTonnage([]Weapon{full}, nil); got != 1 {
		t.Errorf("one 1t mount = %dt, want 1", got)
	}
	// Four quarter-ton mounts are one ton together, not four.
	if got := armamentTonnage([]Weapon{tiny, tiny, tiny, tiny}, nil); got != 1 {
		t.Errorf("four 0.25t mounts = %dt, want 1", got)
	}
	// A Bay built for Vlong is 16.67 tons, charged as 17.
	bay := DesignWeapon(WeaponSpec{SalvoRack, Bay, Standard, Vlong})
	if got := armamentTonnage([]Weapon{bay}, nil); got != 17 {
		t.Errorf("a 16.67t bay = %dt, want 17 (a mount rounds up)", got)
	}
	if got := armamentTonnage(nil, nil); got != 0 {
		t.Errorf("no weapons = %dt, want 0", got)
	}
}
