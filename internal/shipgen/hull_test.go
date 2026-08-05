package shipgen

import "testing"

func TestHullCostLinear(t *testing.T) {
	// Spot-check the p.70 grid against the linear formulas.
	cases := []struct {
		ordinal int
		config  HullConfig
		mcr     int
	}{
		{1, Lifting, 16},
		{2, Lifting, 28},
		{24, Lifting, 292}, // Murphy is Lifting-A = 16
		{1, Cluster, 2},
		{24, Cluster, 48},
		{1, Streamlined, 8},
		{2, Streamlined, 14},
		{24, Streamlined, 146},
		{1, Airframe, 9},
		{1, Unstreamlined, 5},
		{1, Braced, 3},
		{1, Planetoid, 1},
	}
	for _, c := range cases {
		if got := hullCostMCr(c.ordinal, c.config); got != c.mcr {
			t.Errorf("hullCostMCr(%d, %s) = %d, want %d", c.ordinal, c.config, got, c.mcr)
		}
	}
	// Structure multipliers on cost.
	if hullCost(1, Lifting, Organic) != 8_000_000 || hullCost(1, Lifting, Charged) != 32_000_000 {
		t.Errorf("structure cost multipliers wrong")
	}
}

func TestHullMurphy(t *testing.T) {
	// Murphy Scout: 100t Hull-A Lifting Body, TL-12, Shell (Book 2 p.42).
	h := hull(12, 1, 0, Lifting, Shell)
	if h.Letter != 1 || h.Tons != 100 || h.Hardpoints != 1 {
		t.Errorf("hull = %+v, want A/100t/1 hardpoint", h)
	}

	if h.Friction != -5 || h.MaxG != 9 || h.Agility != 1 || h.Stability != 3 || !h.LandCapable {
		t.Errorf("hull attrs = %+v, want Friction /5, Agility +1, Stability +3, lands", h)
	}

	if h.BaseArmor != 6 { // Shell = TL/2 = 6
		t.Errorf("Shell base AV = %d, want 6", h.BaseArmor)
	}

	if h.Cost != 16_000_000 { // Lifting-A = 16 MCr
		t.Errorf("hull cost = %d, want 16 MCr", h.Cost)
	}
}

func TestHullBeowulfOvertonnage(t *testing.T) {
	// Beowulf Free Trader: 222t on a 200t Hull-B, Streamlined, TL-10 (Book 2
	// p.45). +22t is minor overtonnage: Agility -1, Stability -1.
	h := hull(10, 2, 222, Streamlined, FramePlate)
	if h.Tons != 222 || h.Letter != 2 {
		t.Errorf("hull = %+v, want 222t Hull-B", h)
	}

	if h.Agility != -1 || h.Stability != 0 { // Streamlined 0/+1, minus overtonnage 1/1
		t.Errorf("overtonnage agility/stability = %d/%d, want -1/0", h.Agility, h.Stability)
	}

	if h.BaseArmor != 10 || h.Hardpoints != 2 { // Frame-Plate AV = TL; 222/100 = 2
		t.Errorf("hull = %+v, want AV 10, 2 hardpoints", h)
	}
}

func TestHullGrossOvertonnageRoundsUp(t *testing.T) {
	// 260t specced as Hull-B (nominal 200) is >49t over, so it rounds up to Hull-C.
	h := hull(12, 2, 260, Streamlined, FramePlate)
	if h.Letter != 3 || h.Tons != 260 {
		t.Errorf("gross overtonnage: hull = %c/%dt, want C/260t", ordinalLetter(h.Letter), h.Tons)
	}
}

func TestHullUndertonnage(t *testing.T) {
	// 180t on a 200t Hull-B is minor undertonnage: Agility +1, no stability change.
	h := hull(12, 2, 180, Streamlined, FramePlate)
	if h.Agility != 1 || h.Stability != 1 {
		t.Errorf("undertonnage agility/stability = %d/%d, want 1/1", h.Agility, h.Stability)
	}
}
