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

// TestRiskDescriptors covers every row of table 10b (Book 1 p.36), transcribed
// from the reference extract.
func TestRiskDescriptors(t *testing.T) {
	cases := []struct {
		flux int
		want RiskDescriptor
	}{
		{-6, RiskDescriptor{"Impossible", "None", "Far Future"}},
		{-5, RiskDescriptor{"Highly Improbable", "Trivial", "Centuries"}},
		{-4, RiskDescriptor{"Improbable", "Negligible", "Lifetime"}},
		{-3, RiskDescriptor{"Highly Unlikely", "Very Minor", "Generation"}},
		{-2, RiskDescriptor{"Unlikely", "Minor", "Decades"}},
		{-1, RiskDescriptor{"Not Likely", "Mild", "Years"}},
		{0, RiskDescriptor{"Either Way", "Temporary", "Months"}},
		{1, RiskDescriptor{"Possible", "Strong", "Weeks"}},
		{2, RiskDescriptor{"Likely", "Major", "Days"}},
		{3, RiskDescriptor{"Probable", "Severe", "Hours"}},
		{4, RiskDescriptor{"Very Probable", "Very Severe", "Minutes"}},
		{5, RiskDescriptor{"Almost Certain", "Devastating", "Seconds"}},
		{6, RiskDescriptor{"Certain", "Total", "Now"}},
	}

	for _, c := range cases {
		got, ok := RiskDescriptors(c.flux)
		if !ok || got != c.want {
			t.Errorf("RiskDescriptors(%d) = (%+v, %v), want %+v", c.flux, got, ok, c.want)
		}
	}

	for _, flux := range []int{-7, 7} {
		if _, ok := RiskDescriptors(flux); ok {
			t.Errorf("RiskDescriptors(%d) unexpectedly succeeded", flux)
		}
	}
}

// TestImpactSpeed covers every row of table 11a (Book 1 p.37), transcribed
// from the reference extract. KPH 0 for Speeds 17-20 is the book's blank cell,
// not zero speed.
func TestImpactSpeed(t *testing.T) {
	cases := []struct {
		speed      int
		kph        int
		hits       int
		descriptor string
	}{
		{0, 0, 0, "Still"},
		{1, 5, 1, "Creep"},
		{2, 10, 4, "Xslow"},
		{3, 20, 9, "Slow"},
		{4, 30, 16, "Standard"},
		{5, 50, 25, "Cruise"},
		{6, 100, 36, "Fast"},
		{7, 300, 49, "Vfast"},
		{8, 500, 64, ""},
		{9, 700, 81, ""},
		{10, 1_000, 100, "Sonic"},
		{11, 2_000, 121, "Supersonic"},
		{12, 3_000, 144, "Hypersonic"},
		{13, 5_000, 169, ""},
		{14, 10_000, 196, ""},
		{15, 20_000, 225, ""},
		{16, 30_000, 256, "Meteoric"},
		{17, 0, 289, ""},
		{18, 0, 324, ""},
		{19, 0, 361, ""},
		{20, 0, 400, ""},
	}

	for _, c := range cases {
		got, ok := ImpactSpeed(c.speed)
		if !ok || got.Speed != c.speed || got.KPH != c.kph || got.HitsPerTon != c.hits ||
			got.Descriptor != c.descriptor {
			t.Errorf(
				"ImpactSpeed(%d) = (%+v, %v), want kph %d hits %d descriptor %q",
				c.speed,
				got,
				ok,
				c.kph,
				c.hits,
				c.descriptor,
			)
		}
	}

	for _, speed := range []int{-1, 21} {
		if _, ok := ImpactSpeed(speed); ok {
			t.Errorf("ImpactSpeed(%d) unexpectedly succeeded", speed)
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

// TestTemperatureAt covers every row of table 11b (Book 1 p.37), transcribed
// from the reference extract. The printed damage-dice column is deliberately
// not modeled (see CLAUDE.md).
func TestTemperatureAt(t *testing.T) {
	cases := []struct {
		level, kelvin, celsius, hits int
		descriptor                   string
	}{
		{-12, 0, -273, 144, "Absolute Zero"},
		{-11, 25, -250, 121, "Hydrogen Ice; LHyd"},
		{-10, 50, -225, 100, "Oxygen Ice"},
		{-9, 75, -200, 81, "Nitrogen Ice"},
		{-8, 100, -175, 64, ""},
		{-7, 125, -150, 49, ""},
		{-6, 150, -125, 36, ""},
		{-5, 175, -100, 25, ""},
		{-4, 200, -75, 16, "Radon Ice"},
		{-3, 225, -50, 9, ""},
		{-2, 250, -25, 4, ""},
		{-1, 275, 0, 1, "Cold"},
		{0, 300, 25, 0, "Human Temperate Environment"},
		{1, 325, 50, 1, "Hot"},
		{2, 350, 75, 4, ""},
		{3, 375, 100, 9, "Water boils"},
		{4, 400, 125, 16, "Sulfur melts"},
		{5, 425, 150, 25, ""},
		{6, 450, 175, 36, ""},
		{7, 475, 200, 49, ""},
		{8, 500, 225, 64, "Tin melts"},
		{9, 525, 250, 81, "Fire"},
		{10, 550, 275, 100, ""},
		{11, 575, 300, 121, ""},
		{12, 600, 325, 144, ""},
		{13, 700, 425, 350, "Lead melts"},
		{14, 800, 525, 400, ""},
		{15, 900, 625, 450, "Aluminum melts"},
		{16, 1_000, 725, 500, ""},
		{17, 2_000, 1_725, 1_000, "Titanium melts"},
		{18, 3_000, 2_725, 1_500, "Spectral M Star surface"},
		{19, 4_000, 3_725, 2_000, "Spectral K Star surface"},
	}

	for _, c := range cases {
		got, ok := TemperatureAt(c.level)
		if !ok || got.Level != c.level || got.Kelvin != c.kelvin || got.Celsius != c.celsius ||
			got.Hits != c.hits || got.Descriptor != c.descriptor {
			t.Errorf(
				"TemperatureAt(%d) = (%+v, %v), want K/C/hits %d/%d/%d descriptor %q",
				c.level,
				got,
				ok,
				c.kelvin,
				c.celsius,
				c.hits,
				c.descriptor,
			)
		}
	}

	for _, level := range []int{-13, 20} {
		if _, ok := TemperatureAt(level); ok {
			t.Errorf("TemperatureAt(%d) unexpectedly succeeded", level)
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

// TestInsulationAt covers every row of table 11c (Book 1 p.37), transcribed
// from the reference extract.
func TestInsulationAt(t *testing.T) {
	cases := []struct {
		rating, minC, maxC, shipAV int
	}{
		{144, -275, 325, 15},
		{121, -250, 300, 12},
		{100, -225, 275, 10},
		{81, -200, 250, 8},
		{64, -175, 225, 7},
		{49, -150, 200, 5},
		{36, -125, 175, 4},
		{25, -100, 150, 3},
		{16, -75, 125, 2},
		{9, -50, 100, 1},
		{4, -25, 75, 1},
		{1, 0, 50, 1},
	}

	for _, c := range cases {
		got, ok := InsulationAt(c.rating)
		if !ok || got.Rating != c.rating || got.MinC != c.minC || got.MaxC != c.maxC || got.ShipAV != c.shipAV {
			t.Errorf(
				"InsulationAt(%d) = (%+v, %v), want range %d..%d AV %d",
				c.rating,
				got,
				ok,
				c.minC,
				c.maxC,
				c.shipAV,
			)
		}
	}

	if _, ok := InsulationAt(2); ok {
		t.Error("InsulationAt(2) unexpectedly succeeded")
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

// blank marks a cell the book leaves empty; the lookup must return ok=false.
const blank = -1

// TestDecimalSizeFullTable covers every cell of the Decimal Size table (Book 1
// p.43), transcribed from the reference extract and normalized to millimeters.
func TestDecimalSizeFullTable(t *testing.T) {
	rows := []struct {
		code SizeCode
		mm   [10]float64 // indexed by decimal digit 0-9
	}{
		{SizeZero, [10]float64{blank, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9}},
		{SizeR, [10]float64{1, 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 1.8, 1.9}},
		{SizeT, [10]float64{2, 2.5, 3, 3.5, 4, 4.5, 5, 5.5, 6, 6.5}},
		{Size1, [10]float64{7, 15, 20, 30, 35, 40, 50, 55, 60, 70}},
		{Size2, [10]float64{75, 90, 100, 110, 120, 140, 150, 160, 180, 190}},
		{Size3, [10]float64{200, 250, 300, 350, 400, 450, 500, 600, 650, 700}},
		{Size4, [10]float64{750, 800, 900, 1_000, 1_050, 1_100, 1_200, 1_300, 1_350, 1_400}},
		{Size5, [10]float64{1_500, 1_600, 1_700, 1_800, 1_900, 2_000, 5_000, 5_500, 6_000, 6_500}},
		{Size6, [10]float64{7_500, 15_000, 20_000, 30_000, 35_000, 40_000, 50_000, 55_000, 60_000, 70_000}},
		{Size7, [10]float64{75_000, 150_000, 200_000, 300_000, 350_000, 400_000, 500_000, 550_000, 600_000, 700_000}},
	}

	for _, row := range rows {
		for decimal, want := range row.mm {
			got, ok := DecimalSize(row.code, decimal)

			if want == blank {
				if ok {
					t.Errorf("DecimalSize(%q, %d) = (%v, true), want blank cell", row.code, decimal, got)
				}

				continue
			}

			if !ok || got != want {
				t.Errorf("DecimalSize(%q, %d) = (%v, %v), want %v", row.code, decimal, got, ok, want)
			}
		}
	}
}

// TestRandomSizeFullTable covers every cell of the Random Size Variation table
// (Book 1 p.43), transcribed from the reference extract and normalized to
// millimeters. Size 0 is blank through Flux 0.
func TestRandomSizeFullTable(t *testing.T) {
	rows := []struct {
		code SizeCode
		mm   [11]float64 // indexed by flux+5, for Flux -5..+5
	}{
		{SizeZero, [11]float64{blank, blank, blank, blank, blank, blank, 0.2, 0.4, 0.6, 0.8, 1}},
		{SizeR, [11]float64{0.5, 0.6, 0.7, 0.8, 0.9, 1, 1.2, 1.4, 1.6, 1.8, 2}},
		{SizeT, [11]float64{1.5, 1.6, 1.7, 1.8, 1.9, 2, 3, 4, 5, 6, 7}},
		{Size1, [11]float64{4.5, 5, 5.5, 6, 6.5, 7, 20, 35, 50, 60, 75}},
		{Size2, [11]float64{40, 50, 55, 60, 70, 75, 100, 120, 150, 180, 200}},
		{Size3, [11]float64{140, 150, 160, 180, 190, 200, 300, 400, 500, 600, 750}},
		{Size4, [11]float64{450, 500, 550, 600, 650, 750, 900, 1_050, 1_200, 1_350, 1_500}},
		{Size5, [11]float64{1_100, 1_200, 1_300, 1_350, 1_400, 1_500, 1_700, 1_900, 5_000, 6_000, 7_500}},
		{Size6, [11]float64{2_000, 5_000, 5_500, 6_000, 6_500, 7_500, 20_000, 35_000, 50_000, 60_000, 75_000}},
		{
			Size7,
			[11]float64{40_000, 50_000, 55_000, 60_000, 70_000, 75_000, 200_000, 350_000, 500_000, 600_000, 750_000},
		},
	}

	for _, row := range rows {
		for i, want := range row.mm {
			flux := i - 5
			got, ok := RandomSize(row.code, flux)

			if want == blank {
				if ok {
					t.Errorf("RandomSize(%q, %d) = (%v, true), want blank cell", row.code, flux, got)
				}

				continue
			}

			if !ok || got != want {
				t.Errorf("RandomSize(%q, %d) = (%v, %v), want %v", row.code, flux, got, ok, want)
			}
		}
	}
}
