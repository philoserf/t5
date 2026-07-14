package shipcombat

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/shipgen"
)

// TestMountModsDiffer is the defect this package used to have, now impossible to
// reproduce. There was one Mount type with one Mod table (T1 +1 ... T4 +4), and
// its doc said the Mod "adds to a weapon's attack and to Defensive Fire". The book
// has two tables, and they do not merely differ — they run opposite ways.
//
// A weapon in a Single Turret attacks at Mod -2; a defense in one defends at +1.
// The old table gave the weapon +1, so every Space Weapon Task built with it was
// three too generous. No golden caught it, because the goldens all passed mods as
// literal ints and never routed a mount through them. The Mod now comes from the
// designed component, which knows which table it was built from.
func TestMountModsDiffer(t *testing.T) {
	std := shipgen.Standard
	for _, c := range []struct {
		mount          shipgen.Mount
		attack, defend int
	}{
		{shipgen.SingleTurret, -2, 1},
		{shipgen.DualTurret, -1, 2},
		{shipgen.TripleTurret, 0, 3},
		{shipgen.QuadTurret, 1, 4},
	} {
		// The Mining Laser has no Mod of its own, so its attack Mod is the mount's.
		w := shipgen.DesignWeapon(shipgen.WeaponSpec{Model: shipgen.MiningLaser, Mount: c.mount, Stage: std, Range: shipgen.VDistant})
		if w.Mod != c.attack {
			t.Errorf("%v attacking = %+d, want %+d", c.mount, w.Mod, c.attack)
		}
		d := shipgen.DesignDefense(shipgen.DefenseSpec{Model: shipgen.NuclearDamper, Mount: c.mount, Stage: std, Range: shipgen.VDistant})
		if d.Mod != c.defend {
			t.Errorf("%v defending = %+d, want %+d", c.mount, d.Mod, c.defend)
		}
		if w.Mod == d.Mod {
			t.Errorf("%v attacks and defends at the same Mod (%+d) — the two tables differ", c.mount, w.Mod)
		}
	}
}

// TestAttackWithDesignedWeapon: a designed ship can finally fight. The weapon
// carries its own tech level and Mod into the task, where a caller used to have to
// invent both.
func TestAttackWithDesignedWeapon(t *testing.T) {
	// A Beam Laser in a Single Turret: TL 10, Mod +0 (the turret's -2, the laser's
	// own +2). Target = 10 + C+S+K 12 + 0 = 22 — the Murphy's number below.
	w := shipgen.DesignWeapon(shipgen.WeaponSpec{Model: shipgen.BeamLaser, Mount: shipgen.SingleTurret, Range: shipgen.VDistant})
	if got := SpaceWeaponTarget(w.TL, 12, w.Mod); got != 22 {
		t.Errorf("designed Beam Laser target = %d, want 22", got)
	}
	if !Attack(dice.NewScripted(3, 3, 3, 3, 3), w, 5, 12).Success {
		t.Errorf("5D=15 should hit target 22")
	}
	if Attack(dice.NewScripted(6, 5, 5, 5, 4), w, 5, 12).Success {
		t.Errorf("5D=25 should miss target 22")
	}

	// The same laser in a Quad Turret aims better (+1) and hits harder (4D vs 1D).
	quad := shipgen.DesignWeapon(shipgen.WeaponSpec{Model: shipgen.BeamLaser, Mount: shipgen.QuadTurret, Range: shipgen.VDistant})
	if quad.Mod != 3 || quad.Hits != 4 {
		t.Errorf("quad Beam Laser = Mod %+d, %dD; want +3, 4D", quad.Mod, quad.Hits)
	}
}

// TestDefendWithDesignedDefense: the Vanguard's Meson Screen, designed rather than
// hand-numbered. A Meson Screen (TL 13) in its Bolt-In (+3) against a Meson Gun-12
// is target 13 - 12 + 3 = 4, the book's own number (Book 2 p.196).
func TestDefendWithDesignedDefense(t *testing.T) {
	d := shipgen.DesignDefense(shipgen.DefenseSpec{Model: shipgen.MesonScreen, Mount: shipgen.BoltIn, Range: shipgen.VDistant})
	if d.TL != 13 || d.Mod != 3 {
		t.Fatalf("Meson Screen = TL %d Mod %+d, want TL 13 Mod +3", d.TL, d.Mod)
	}
	if got := DefensiveFireTarget(d.TL, 12, d.Mod); got != 4 {
		t.Errorf("Vanguard Meson Screen target = %d, want 4", got)
	}
	if !Defend(dice.NewScripted(4), d, 12).Success {
		t.Errorf("1D=4 should deflect at target 4")
	}
	if Defend(dice.NewScripted(5), d, 12).Success {
		t.Errorf("1D=5 should not deflect at target 4")
	}
}

// TestShipBridge: the pieces a fight needs, taken from a designed ship rather than
// invented — its armor layers, its compartments, and its agility.
func TestShipBridge(t *testing.T) {
	ship := shipgen.Design(shipgen.ShipSpec{
		TL: 12, HullLetter: 1, Config: shipgen.Lifting, Structure: shipgen.Shell, ArmorLayers: 2,
		Maneuver: &shipgen.DriveSpec{Letter: 1}, Power: &shipgen.DriveSpec{Letter: 1},
	})
	// The Murphy's two Shell layers, AV-6 all told.
	layers := ArmorLayers(ship)
	if len(layers) != 2 {
		t.Fatalf("armor = %v, want 2 layers", layers)
	}
	// The hull ordinal already keyed Table H — the packages agreed all along.
	card, ok := Card(ship)
	if !ok || card.Compartments == 0 {
		t.Errorf("a Hull-A ship should have a compartment card, got %+v", card)
	}
	// Agility is the thrust left over after going somewhere.
	if got := ShipAgility(ship, 1); got != ship.Maneuver.Potential-1 {
		t.Errorf("agility spending 1G = %d, want %d", got, ship.Maneuver.Potential-1)
	}
}

// TestSpaceWeaponMurphy golden-locks the Book 2 p.199 Murphy example: a Beam
// Laser-10 fired via the ship's Console-10/12 with no mods -> target 22, and one
// range-band die per range.
func TestSpaceWeaponMurphy(t *testing.T) {
	if got := SpaceWeaponTarget(10, 12, 0); got != 22 {
		t.Errorf("Murphy Console-12 target = %d, want 22", got)
	}
	if got := SpaceWeaponDice(7, 10); got != 7 { // R=7, weapon TL 10 not below range
		t.Errorf("SpaceWeaponDice(R=7) = %d, want 7", got)
	}
	if got := SpaceWeaponDice(5, 10); got != 5 {
		t.Errorf("SpaceWeaponDice(R=5) = %d, want 5", got)
	}
	// Weapon TL below the range adds a die.
	if got := SpaceWeaponDice(12, 10); got != 13 {
		t.Errorf("SpaceWeaponDice(R=12, TL=10) = %d, want 13", got)
	}
	// Gryphon targeting the turret at Compartment+2 with Console-10, Mod +1 -> 21.
	if got := SpaceWeaponTarget(10, 10, 1); got != 21 {
		t.Errorf("Gryphon turret target = %d, want 21", got)
	}
	// A scripted low roll hits the 22 target; a high roll misses.
	if !ResolveSpaceWeapon(dice.NewScripted(3, 3, 3, 3, 3), 5, 10, 12, 0).Success {
		t.Errorf("5D=15 should hit target 22")
	}
	if ResolveSpaceWeapon(dice.NewScripted(6, 5, 5, 5, 4), 5, 10, 12, 0).Success {
		t.Errorf("5D=25 should miss target 22")
	}
}

// TestDefensiveFire golden-locks the Book 2 p.196 Vanguard example (Meson
// Screen-13 vs Meson Gun-12, Mount +3 -> 4) and the p.199 Gryphon SandCaster
// (Defense-10 vs Attack-12, Mount 0 -> -2, no effect).
func TestDefensiveFire(t *testing.T) {
	if got := DefensiveFireTarget(13, 12, 3); got != 4 {
		t.Errorf("Vanguard Meson Screen target = %d, want 4", got)
	}
	if got := DefensiveFireTarget(10, 12, 0); got != -2 {
		t.Errorf("Gryphon SandCaster target = %d, want -2", got)
	}
	// The SandCaster (-2 target) can never deflect on 1D.
	if ResolveDefensiveFire(dice.NewScripted(1), 10, 12, 0).Success {
		t.Errorf("a -2 defense target must always fail")
	}
	// The Meson Screen (target 4) deflects on a 1D of 4.
	if !ResolveDefensiveFire(dice.NewScripted(4), 13, 12, 3).Success {
		t.Errorf("1D=4 should deflect at target 4")
	}
}

func TestMissileTask(t *testing.T) {
	// Guidance is a property of the designed round (shipgen.Guidance); what it is
	// worth on this task is a combat rule, and lives here: a flat 5 for a hardwired
	// brain, the gunner's own C+S+K for one they fly, and the missile's rolled mind
	// for a self-aware one.
	const gunner, brain = 8, 9
	cases := map[shipgen.Guidance]int{
		shipgen.UnGuided:       0,
		shipgen.HardWired:      5,
		shipgen.OperatorGuided: gunner,
		shipgen.DownLoaded:     gunner, // the gunner's personality, copied in
		shipgen.SelfAware:      brain,  // the round's own mind, rolled at launch
	}
	for g, want := range cases {
		if got := guidanceAsset(g, gunner, brain); got != want {
			t.Errorf("%v asset = %d, want %d", g, got, want)
		}
	}
	if got := MissileTarget(10, 5, 1); got != 16 { // HardWired missile TL-10, mod +1
		t.Errorf("MissileTarget = %d, want 16", got)
	}
	// 5D at range, 6D when missile TL is below the range.
	if MissileDice(10, 5) != 5 || MissileDice(10, 12) != 6 {
		t.Errorf("MissileDice wrong")
	}
}
