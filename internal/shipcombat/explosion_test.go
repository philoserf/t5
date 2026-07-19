package shipcombat

import (
	"testing"

	"github.com/philoserf/t5/internal/shipgen"
)

// TestMissileMassiveExplosion golden-locks the Book 2 p.197 Sz+1D proximity table.
func TestMissileMassiveExplosion(t *testing.T) {
	// 6 or less: Direct Hit, Vaporized (100D).
	if e := MissileMassiveExplosion(
		6,
	); !e.Vaporized || e.Blast != 100 ||
		e.Proximity != "Direct Hit" {
		t.Errorf("MissileMassiveExplosion(6) = %+v, want Direct Hit / Vaporized / 100D", e)
	}
	// 7: a Hit — Blast 90D, BFE 20D, Rad 10D, Burn 30D.
	if e := MissileMassiveExplosion(
		7,
	); e.Blast != 90 || e.BFE != 20 || e.Rad != 10 || e.Burn != 30 ||
		e.Vaporized {
		t.Errorf("MissileMassiveExplosion(7) = %+v, want 90/20/10/30", e)
	}
	// 11: a Far Miss — 5D / 1D / 1D / 1D.
	if e := MissileMassiveExplosion(11); e.Blast != 5 || e.BFE != 1 || e.Rad != 1 || e.Burn != 1 {
		t.Errorf("MissileMassiveExplosion(11) = %+v, want 5/1/1/1", e)
	}
	// 12 or more: a clean Miss.
	if e := MissileMassiveExplosion(12); e.Blast != 0 || e.Vaporized || e.Proximity != "Miss" {
		t.Errorf("MissileMassiveExplosion(12) = %+v, want a clean Miss", e)
	}
}

// TestWeaponsMassiveExplosion golden-locks the Book 2 p.196 Weapons-Task Massive
// Explosion multiplier table, every row, keyed by the designed round the book's
// option names describe.
func TestWeaponsMassiveExplosion(t *testing.T) {
	cases := []struct {
		name string
		typ  shipgen.MissileType
		size int
		want MassiveExplosionMultiplier
	}{
		// AM Missile: 10x AV, 1x Flash. The Anti-Matter warhead is Size-4 only.
		{"AM Missile", shipgen.AntiMatter, 4, MassiveExplosionMultiplier{AV: 10, Flash: 1}},
		// Missile Nuke Option: the only row with all four columns.
		{"Nuke", shipgen.Nuke, 5, MassiveExplosionMultiplier{AV: 10, Rad: 1, Flash: 1, EMP: 1}},
		// Missile EMP Option: 1x AV, 1x EMP, no Rad and no Flash.
		{"EMP", shipgen.EMP, 5, MassiveExplosionMultiplier{AV: 1, EMP: 1}},
		// KK Missile: 1x AV alone — its damage is speed, not blast.
		{"KK Missile", shipgen.Kinetic, 7, MassiveExplosionMultiplier{AV: 1}},
		// DeadFall, the one option split by size: AV 1x/5x/10x, Flash 1x/2x/3x.
		{"DeadFall Size-4", shipgen.Deadfall, 4, MassiveExplosionMultiplier{AV: 1, Flash: 1}},
		{"DeadFall Size-5", shipgen.Deadfall, 5, MassiveExplosionMultiplier{AV: 5, Flash: 2}},
		{"DeadFall Size-6", shipgen.Deadfall, 6, MassiveExplosionMultiplier{AV: 10, Flash: 3}},
	}
	for _, c := range cases {
		m := shipgen.Missile{Spec: shipgen.MissileSpec{Type: c.typ, Size: c.size}}
		if got, ok := WeaponsMassiveExplosion(m); !ok || got != c.want {
			t.Errorf("WeaponsMassiveExplosion(%s) = %+v,%v, want %+v,true",
				c.name, got, ok, c.want)
		}
	}
	// Warheads the table does not carry: an Explosive round Penetrates and no
	// more, and a Decoy or a Sensor Package does not detonate at all.
	for _, typ := range []shipgen.MissileType{
		shipgen.Slug, shipgen.Explosive, shipgen.Decoy, shipgen.SensorPkg,
	} {
		m := shipgen.Missile{Spec: shipgen.MissileSpec{Type: typ, Size: 5}}
		if _, ok := WeaponsMassiveExplosion(m); ok {
			t.Errorf("a %s round should not be tabulated", typ)
		}
	}
	// DeadFall exists at Size 4-6 only (p.170); no other size is tabulated.
	for _, size := range []int{3, 7} {
		m := shipgen.Missile{Spec: shipgen.MissileSpec{Type: shipgen.Deadfall, Size: size}}
		if _, ok := WeaponsMassiveExplosion(m); ok {
			t.Errorf("a Size-%d DeadFall should not be tabulated", size)
		}
	}
}

// A round as DesignMissile actually builds it reaches the table — the drift this
// replaces meant a designed Kinetic or Deadfall round silently dropped its
// explosion, because nothing produced the "KK" and "DeadFall6" keys the table
// was written against.
func TestWeaponsMassiveExplosionFromDesignedRound(t *testing.T) {
	// A KK Missile launcher is base TL 10; Long Range shifts it to the TL 11 a
	// Self-Aware brain needs (Book 2 p.170).
	spec := shipgen.DefaultWeapon(shipgen.KKMissile)
	spec.Range = shipgen.LongRange
	launcher := shipgen.DesignWeapon(spec)

	round := shipgen.DesignMissile(launcher, shipgen.MissileSpec{
		Size: 7, Type: shipgen.Kinetic, Guidance: shipgen.SelfAware,
	})
	if len(round.Problems) > 0 {
		t.Fatalf("a Size-7 Self-Aware kinetic round is legal: %v", round.Problems)
	}

	got, ok := WeaponsMassiveExplosion(round)
	if want := (MassiveExplosionMultiplier{AV: 1}); !ok || got != want {
		t.Errorf("WeaponsMassiveExplosion(%s) = %+v,%v, want %+v,true",
			round.LongName(), got, ok, want)
	}
}
