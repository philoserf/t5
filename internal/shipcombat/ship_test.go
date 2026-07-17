package shipcombat

import (
	"testing"

	"github.com/philoserf/t5/internal/shipgen"
)

// TestArmorLayersDoesNotDivide is the bug this file used to have. shipgen's
// Armor.AV is the PER-LAYER value — "identical layers are not summed on the
// record" — and ArmorLayers divided it by the layer count, so a ship's armor got
// thinner the more layers it bought. Two AV-6 layers were modelled as two AV-3
// ones, and an 8-point hit that the real ship stops dead went straight through.
func TestArmorLayersDoesNotDivide(t *testing.T) {
	// The Murphy: a Shell hull at TL 12, two layers, AV 6 each.
	ship := shipgen.Design(shipgen.ShipSpec{
		TL: 12, HullLetter: 1, Config: shipgen.Lifting, Structure: shipgen.Shell, ArmorLayers: 2,
		Maneuver: &shipgen.DriveSpec{Letter: 1}, Power: &shipgen.DriveSpec{Letter: 1},
	})

	layers := ArmorLayers(ship)
	if len(layers) != 2 {
		t.Fatalf("armor = %v, want 2 layers", layers)
	}

	for i, av := range layers {
		if av != ship.Armor.AV {
			t.Errorf("layer %d has AV %d, want the ship's per-layer AV %d", i, av, ship.Armor.AV)
		}
	}
	// An 8-point hit does not get through two AV-6 layers.
	if pen, _ := Penetrate(8, layers); pen {
		t.Errorf("8 damage should not penetrate two AV-%d layers", ship.Armor.AV)
	}
	// Buying more armor must never make a ship softer.
	four := shipgen.Design(shipgen.ShipSpec{
		TL: 12, HullLetter: 1, Config: shipgen.Lifting, Structure: shipgen.Shell, ArmorLayers: 4,
		Maneuver: &shipgen.DriveSpec{Letter: 1}, Power: &shipgen.DriveSpec{Letter: 1},
	})
	if got := ArmorLayers(four); got[0] < layers[0] {
		t.Errorf(
			"a 4-layer ship has thinner layers (%d) than a 2-layer one (%d)",
			got[0],
			layers[0],
		)
	}
}

// TestShipAgilityRespectsTheHull: a hull's configuration caps its thrust whatever
// drive is bolted to it (Book 2 p.71) — a Cluster is rated for 1G. It used to dodge
// at whatever the drive said, which for a big drive on a small hull was 8G.
func TestShipAgilityRespectsTheHull(t *testing.T) {
	cluster := shipgen.Design(shipgen.ShipSpec{
		TL:          15,
		HullLetter:  1,
		Config:      shipgen.Cluster,
		Structure:   shipgen.FramePlate,
		ArmorLayers: 1,
		Maneuver:    &shipgen.DriveSpec{Letter: 4},
		Power:       &shipgen.DriveSpec{Letter: 4},
	})
	if cluster.Hull.MaxG != 1 {
		t.Fatalf("a Cluster hull is capped at 1G, got %d", cluster.Hull.MaxG)
	}

	if cluster.Maneuver.Potential <= cluster.Hull.MaxG {
		t.Fatalf(
			"this test needs a drive that outruns the hull, got %dG",
			cluster.Maneuver.Potential,
		)
	}
	// Design says so...
	if len(cluster.Problems) == 0 {
		t.Errorf("a drive that outruns its hull's config should be a Problem")
	}
	// ...and the ship flies the hull it actually has.
	if got := ShipAgility(cluster, 0); got != 1 {
		t.Errorf("Cluster agility = %d, want 1 (the hull's cap, not the drive's %d)",
			got, cluster.Maneuver.Potential)
	}
}
