package shipcombat

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestMountMod(t *testing.T) {
	cases := map[Mount]int{
		SingleTurret: 1, DualTurret: 2, TripleTurret: 3, QuadTurret: 4,
		SingleBarbette: 1, DualBarbette: 2,
	}
	for m, want := range cases {
		if got := m.Mod(); got != want {
			t.Errorf("Mount %d Mod = %d, want %d", m, got, want)
		}
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
	// Guidance values (Book 2 p.197).
	if UnGuided.Value(8) != 0 || HardWired.Value(8) != 5 || OperatorGuided.Value(8) != 8 {
		t.Errorf("guidance values wrong")
	}
	if got := MissileTarget(10, 5, 1); got != 16 { // HardWired missile TL-10, mod +1
		t.Errorf("MissileTarget = %d, want 16", got)
	}
	// 5D at range, 6D when missile TL is below the range.
	if MissileDice(10, 5) != 5 || MissileDice(10, 12) != 6 {
		t.Errorf("MissileDice wrong")
	}
}
