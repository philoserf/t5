package shipgen

import "testing"

func TestFuelMurphy(t *testing.T) {
	// Murphy Scout: Jump-2 in 100t = 20t jump fuel; Power potential 2 = 2t
	// operations fuel; total 22t. Scoops + purifier fittings add cost.
	jump := &Drive{Kind: Jump, Potential: 2}
	power := &Drive{Kind: Power, Potential: 2}
	f := fuel(100, jump, power, true, true)
	if f.Tons != 22 {
		t.Errorf("fuel tankage = %d, want 22", f.Tons)
	}
	// 22t x Cr500 + KCr100 scoop + MCr1 purifier.
	if f.Cost != 22*500+100_000+1_000_000 {
		t.Errorf("fuel cost = %d, want %d", f.Cost, 22*500+1_100_000)
	}
}

func TestFuelNoJumpNoFittings(t *testing.T) {
	// An in-system ship with only a power plant carries just operations fuel.
	f := fuel(200, nil, &Drive{Kind: Power, Potential: 3}, false, false)
	if f.Tons != 6 || f.Cost != 3000 { // 3*200/100 = 6t; 6*Cr500
		t.Errorf("fuel = %+v, want 6t / Cr3000", f)
	}
}

func TestFuelStageMultiplier(t *testing.T) {
	// An Experimental jump drive burns x2.0 fuel (Book 2 p.76 X).
	f := fuel(100, &Drive{Kind: Jump, Potential: 2, Stage: Experimental}, nil, false, false)
	if f.Tons != 40 { // 20 x 2.0
		t.Errorf("experimental jump fuel = %d, want 40", f.Tons)
	}
}

func TestArmorMurphy(t *testing.T) {
	// Murphy: 2 layers of Shell armor on a 100t TL-12 hull.
	a := armor(12, 100, Shell, 2)
	if a.Layers != 2 || a.AV != 6 || a.Tons != 2 {
		t.Errorf("armor = %+v, want 2 layers / AV 6 / 2t", a)
	}
}

func TestArmorLayersAndStructures(t *testing.T) {
	// Plate (Frame-Plate) armor is 4% per extra layer, AV = TL.
	if a := armor(10, 100, FramePlate, 2); a.Tons != 4 || a.AV != 10 {
		t.Errorf("Plate armor = %+v, want 4t / AV 10", a)
	}
	// 6 layers of Plate on 100t = 4% x 5 = 20t.
	if a := armor(12, 100, FramePlate, 6); a.Tons != 20 {
		t.Errorf("6-layer armor = %dt, want 20", a.Tons)
	}
	// One layer is integral: no tonnage.
	if a := armor(12, 100, FramePlate, 1); a.Tons != 0 {
		t.Errorf("single-layer armor should be free, got %dt", a.Tons)
	}
	// Structure armor values (Book 2 p.75 B).
	if structureAV(FeNi, 12) != 20 || structureAV(Charged, 12) != 24 || structureAV(FramePlate, 12) != 12 {
		t.Errorf("structure AV wrong")
	}
}
