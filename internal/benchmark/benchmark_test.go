package benchmark

import "testing"

func TestScale(t *testing.T) {
	cases := []struct {
		level      int
		descriptor string
		ok         bool
	}{
		{-1, "", false},
		{0, "Unknown", true},
		{10, "World", true},
		{20, "Sector", true},
		{36, "All Reality", true},
		{37, "", false},
	}

	for _, c := range cases {
		got, ok := Scale(c.level)
		if ok != c.ok || got.Descriptor != c.descriptor {
			t.Errorf("Scale(%d) = (%+v, %v), want descriptor %q, ok %v", c.level, got, ok, c.descriptor, c.ok)
		}
	}
}

func TestScaleRowsAreIndexedByLevel(t *testing.T) {
	for level := range 37 {
		entry, ok := Scale(level)
		if !ok || entry.Level != level || entry.Descriptor == "" {
			t.Errorf("Scale(%d) = (%+v, %v), want populated entry at its level", level, entry, ok)
		}
	}
}

func TestRisk(t *testing.T) {
	cases := []struct {
		risk           Risk
		level          int
		requiresAction bool
	}{
		{Risk{-6, -6, -6}, -18, false},
		{Risk{1, -1, 0}, 0, false},
		{Risk{3, 4, 3}, 10, true},
		{Risk{6, 6, 6}, 18, true},
	}

	for _, c := range cases {
		if got := c.risk.Level(); got != c.level {
			t.Errorf("%+v.Level() = %d, want %d", c.risk, got, c.level)
		}

		if got := c.risk.RequiresAction(); got != c.requiresAction {
			t.Errorf("%+v.RequiresAction() = %v, want %v", c.risk, got, c.requiresAction)
		}
	}
}

func TestRiskDescriptors(t *testing.T) {
	probability, severity, imminence, ok := RiskDescriptors(6)
	if !ok || probability != "Certain" || severity != "Total" || imminence != "Now" {
		t.Errorf("RiskDescriptors(6) = (%q, %q, %q, %v)", probability, severity, imminence, ok)
	}

	if _, _, _, ok := RiskDescriptors(7); ok {
		t.Error("RiskDescriptors(7) unexpectedly succeeded")
	}
}

func TestImpactSpeed(t *testing.T) {
	cases := []struct {
		speed int
		kph   int
		hits  int
		ok    bool
	}{
		{-1, 0, 0, false},
		{0, 0, 0, true},
		{10, 1_000, 100, true},
		{16, 30_000, 256, true},
		{20, 0, 400, true},
		{21, 0, 0, false},
	}

	for _, c := range cases {
		got, ok := ImpactSpeed(c.speed)
		if ok != c.ok || got.KPH != c.kph || got.HitsPerTon != c.hits {
			t.Errorf("ImpactSpeed(%d) = (%+v, %v), want kph %d hits %d ok %v", c.speed, got, ok, c.kph, c.hits, c.ok)
		}
	}
}

func TestImpactHits(t *testing.T) {
	cases := []struct {
		speed int
		tons  float64
		want  float64
	}{
		{-1, 1, 0}, {5, 0, 0}, {5, -1, 0}, {5, 1, 25}, {5, 0.5, 12.5}, {10, 3, 300},
	}

	for _, c := range cases {
		if got := ImpactHits(c.speed, c.tons); got != c.want {
			t.Errorf("ImpactHits(%d, %v) = %v, want %v", c.speed, c.tons, got, c.want)
		}
	}
}

func TestTemperatureAt(t *testing.T) {
	cases := []struct {
		level, kelvin, celsius, hits int
		ok                           bool
	}{
		{-13, 0, 0, 0, false},
		{-12, 0, -273, 144, true},
		{-1, 275, 0, 1, true},
		{0, 300, 25, 0, true},
		{12, 600, 325, 144, true},
		{17, 2_000, 1_725, 1_000, true},
		{19, 4_000, 3_725, 2_000, true},
		{20, 0, 0, 0, false},
	}

	for _, c := range cases {
		got, ok := TemperatureAt(c.level)
		if ok != c.ok || got.Kelvin != c.kelvin || got.Celsius != c.celsius || got.Hits != c.hits {
			t.Errorf(
				"TemperatureAt(%d) = (%+v, %v), want K/C/hits %d/%d/%d ok %v",
				c.level,
				got,
				ok,
				c.kelvin,
				c.celsius,
				c.hits,
				c.ok,
			)
		}
	}
}

func TestHeatHits(t *testing.T) {
	for _, c := range []struct {
		kelvin float64
		want   int
	}{{-1, 0}, {600, 0}, {601, 300}, {700, 350}, {1_000, 500}} {
		if got := HeatHits(c.kelvin); got != c.want {
			t.Errorf("HeatHits(%v) = %d, want %d", c.kelvin, got, c.want)
		}
	}
}

func TestInsulationAt(t *testing.T) {
	cases := []struct {
		rating, minC, maxC, shipAV int
		ok                         bool
	}{
		{144, -275, 325, 15, true},
		{49, -150, 200, 5, true},
		{1, 0, 50, 1, true},
		{2, 0, 0, 0, false},
	}

	for _, c := range cases {
		got, ok := InsulationAt(c.rating)
		if ok != c.ok || got.MinC != c.minC || got.MaxC != c.maxC || got.ShipAV != c.shipAV {
			t.Errorf(
				"InsulationAt(%d) = (%+v, %v), want range %d..%d AV %d ok %v",
				c.rating,
				got,
				ok,
				c.minC,
				c.maxC,
				c.shipAV,
				c.ok,
			)
		}
	}
}

func TestSizeAt(t *testing.T) {
	for _, c := range []struct {
		code       SizeCode
		descriptor string
		ok         bool
	}{{SizeR, "Reading", true}, {Size5, "Person", true}, {Size9, "Moonlet", true}, {"X", "", false}} {
		got, ok := SizeAt(c.code)
		if ok != c.ok || got.Descriptor != c.descriptor {
			t.Errorf("SizeAt(%q) = (%+v, %v), want descriptor %q ok %v", c.code, got, ok, c.descriptor, c.ok)
		}
	}
}

func TestDecimalSize(t *testing.T) {
	cases := []struct {
		code    SizeCode
		decimal int
		wantMM  float64
		ok      bool
	}{
		{SizeZero, 0, 0, false},
		{SizeZero, 5, 0.5, true},
		{SizeR, 4, 1.4, true},
		{Size3, 1, 250, true},
		{Size5, 3, 1_800, true},
		{Size7, 9, 700_000, true},
		{Size8, 0, 0, false},
		{Size5, 10, 0, false},
	}

	for _, c := range cases {
		got, ok := DecimalSize(c.code, c.decimal)
		if ok != c.ok || got != c.wantMM {
			t.Errorf("DecimalSize(%q, %d) = (%v, %v), want (%v, %v)", c.code, c.decimal, got, ok, c.wantMM, c.ok)
		}
	}
}

func TestRandomSize(t *testing.T) {
	cases := []struct {
		code   SizeCode
		flux   int
		wantMM float64
		ok     bool
	}{
		{SizeZero, 0, 0, false},
		{SizeZero, 1, 0.2, true},
		{SizeR, -5, 0.5, true},
		{Size5, -2, 1_350, true},
		{Size5, 4, 6_000, true},
		{Size7, 5, 750_000, true},
		{Size8, 0, 0, false},
		{Size5, 6, 0, false},
	}

	for _, c := range cases {
		got, ok := RandomSize(c.code, c.flux)
		if ok != c.ok || got != c.wantMM {
			t.Errorf("RandomSize(%q, %d) = (%v, %v), want (%v, %v)", c.code, c.flux, got, ok, c.wantMM, c.ok)
		}
	}
}
