package shipcombat

import "testing"

// TestMissileMassiveExplosion golden-locks the Book 2 p.197 Sz+1D proximity table.
func TestMissileMassiveExplosion(t *testing.T) {
	// 6 or less: Direct Hit, Vaporized (100D).
	if e := MissileMassiveExplosion(6); !e.Vaporized || e.Blast != 100 || e.Proximity != "Direct Hit" {
		t.Errorf("MissileMassiveExplosion(6) = %+v, want Direct Hit / Vaporized / 100D", e)
	}
	// 7: a Hit — Blast 90D, BFE 20D, Rad 10D, Burn 30D.
	if e := MissileMassiveExplosion(7); e.Blast != 90 || e.BFE != 20 || e.Rad != 10 || e.Burn != 30 || e.Vaporized {
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
