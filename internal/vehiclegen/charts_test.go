package vehiclegen

import (
	"math"
	"reflect"
	"testing"
)

func TestCreationChartRegistriesComplete(t *testing.T) {
	wantTypes := []string{
		"G:GC", "G:U", "G:T", "G:V", "G:M", "G:H", "G:R",
		"M:T", "M:C", "M:V", "M:R",
		"F:F", "F:G", "F:B",
		"W:S", "W:U", "W:B",
	}
	wantMissions := []string{
		"G:RO", "G:P", "G:C", "G:MP", "G:OR",
		"M:W", "M:T", "M:S", "M:R",
		"F:A", "F:B", "F:C", "F:P", "F:S", "F:U",
		"W:C", "W:P", "W:E", "W:T",
	}
	wantMotives := []string{
		"G:AC", "G:W", "G:Z", "G:G", "G:T", "G:L",
		"M:AC", "M:W", "M:Z", "M:G", "M:T", "M:L",
		"F:W", "F:R", "F:F", "F:LTA", "F:Z", "F:G",
		"W:S", "W:U", "W:H", "W:G",
	}

	assertRegistry(t, "type", typeRows, wantTypes)
	assertRegistry(t, "mission", missionRows, wantMissions)
	assertRegistry(t, "motive", motiveRows, wantMotives)
}

func TestDesignMilitaryWeaponMission(t *testing.T) {
	// Book 3 p.150: Weapon is a Mission row (section B); a Tank Type with the
	// Weapon Mission is the book's Combat Tank ("TW" in the p.140 catalog).
	v := Design(Spec{Category: "M", Type: "T", Mission: "W", Motive: "W"})
	if len(v.Problems) != 0 {
		t.Fatalf("Combat Tank design problems = %v", v.Problems)
	}

	if v.TechLevel != 6 || v.Tons != 7 || v.Speed != 4 || v.Armor != 50 || v.CostKCr != 600 {
		t.Errorf("Combat Tank values = %+v", v.Values)
	}
}

func assertRegistry(t *testing.T, name string, rows map[string]Modifier, want []string) {
	t.Helper()

	if len(rows) != len(want) {
		t.Errorf("%s registry has %d rows, want %d", name, len(rows), len(want))
	}

	for _, key := range want {
		if row, ok := rows[key]; !ok || row.Code == "" || row.Descriptor == "" {
			t.Errorf("%s registry row %q = %+v, %v", name, key, row, ok)
		}
	}
}

func TestEnhancerRegistryComplete(t *testing.T) {
	codes := []string{
		"Vl", "L", "M", "H", "Vh",
		"Fos", "PC", "Ren", "Pro", "Ear", "Std", "Imp", "Adv",
		"Air", "Enclosed", "Sealed", "DoubleSealed", "Insulated", "Protected", "Armored",
		"UpArmored", "AltArmored",
		"HighPowered", "Slave", "Remote", "WeaponMount", "Luxury", "Fast", "PassengerModule",
		"CargoModule", "Redundancy", "OffRoad", "Mole", "Hydrofoils", "Stubs", "VTOL", "STOL",
		"LiftingBody", "Wings1", "Wings2", "Wings3", "Floats", "ParasiteNipple",
	}

	if len(enhancerRows) != len(codes) {
		t.Fatalf("enhancer registry has %d rows, want %d", len(enhancerRows), len(codes))
	}

	for _, code := range codes {
		row, ok := Enhancer(code)
		if !ok || row.Code != code || row.Descriptor == "" {
			t.Errorf("Enhancer(%q) = %+v, %v", code, row, ok)
		}
	}

	if _, ok := Enhancer("not-a-row"); ok {
		t.Error("unknown enhancer found")
	}
}

// TestEnhancerColumnsAgainstPage asserts every Chart 12 cell against literals
// transcribed from the rendered Book 3 p.152 (independent of the registry —
// a misread column must fail here, not be mirrored). Blank and printed-0
// cells are zero Operations. Column order matches the chart: TL, Tons, Speed,
// Load, Armor, Cage, FlashProof, RadProof, SoundProof, PsiShield, Insulated,
// Sealed, KCr.
func TestEnhancerColumnsAgainstPage(t *testing.T) {
	non := Operation{}

	want := map[string][13]Operation{
		// D Bulk.
		"Vl": {
			add(-1), mul(1.0 / 3), add(1), add(-2), mul(1.0 / 3), non, non, mul(1.0 / 3), non, non,
			mul(1.0 / 3), mul(1.0 / 3), mul(1.0 / 3),
		},
		"L": {add(-1), mul(.5), add(1), add(-1), mul(.5), non, non, mul(.5), non, non, mul(.5), mul(.5), mul(.5)},
		"M": {},
		"H": {add(1), mul(2), add(-1), add(2), mul(2), non, non, mul(2), mul(2), non, mul(2), mul(2), mul(3)},
		"Vh": {
			add(2), mul(3), add(-2), add(3), mul(3), non, non, mul(2), mul(2), non, mul(3), mul(3),
			mul(9),
		},
		// E Stage.
		"Fos": {add(-2), add(2), non, non, add(-10), non, non, non, add(-10), non, non, non, non},
		"PC":  {add(-1), add(1), add(-2), add(-2), add(-5), non, non, non, add(-5), non, non, non, add(10)},
		"Ren": {add(-1), add(1), add(-1), add(-1), non, non, non, non, non, non, non, non, add(20)},
		"Pro": {add(-2), add(1), add(-1), add(-1), non, non, non, non, non, non, non, non, add(20)},
		"Ear": {add(-1), add(1), non, non, add(-10), non, non, non, add(-10), non, non, non, add(10)},
		"Std": {},
		"Imp": {add(1), add(-1), non, non, add(10), non, non, non, add(10), non, non, non, add(20)},
		"Adv": {add(3), add(-2), add(1), add(1), add(20), non, non, non, add(20), non, non, non, add(40)},
		// F Environ.
		"Air": {add(-2), non, non, non, non, non, non, non, non, non, non, non, non},
		"Enclosed": {
			add(-1), non, non, non, add(4), non, add(4), non, add(4), non, add(12), non,
			non,
		},
		"Sealed": {non, non, non, non, add(6), add(2), add(6), non, add(8), non, add(16), add(20), add(2)},
		"DoubleSealed": {
			non, add(1), non, non, add(8), add(4), add(6), non, add(12), non, add(30),
			add(20), add(5),
		},
		"Insulated": {
			non, non, non, non, add(8), add(4), add(6), non, add(12), non, add(30), add(20),
			add(10),
		},
		"Protected": {
			add(1), add(1), non, non, add(10), add(10), add(10), add(10), add(12), non,
			add(10), add(20), add(20),
		},
		"Armored": {
			add(2), add(1), non, non, add(20), add(10), add(10), add(10), add(12), non,
			add(20), add(20), add(30),
		},
		"UpArmored": {
			add(3), add(2), non, non, add(30), add(20), add(20), add(20), add(20), non,
			add(30), add(20), add(40),
		},
		"AltArmored": {
			add(3), add(2), non, non, add(60), add(20), add(30), add(30), add(30), non,
			add(30), add(30), add(50),
		},
		// G Options.
		"HighPowered": {add(1), add(1), add(1), add(-1), non, non, non, non, non, non, non, non, add(100)},
		"Slave":       {add(1), add(-1), non, non, non, non, non, non, non, non, non, non, add(10)},
		"Remote":      {add(1), add(-2), non, non, non, non, non, non, non, non, non, non, add(20)},
		"WeaponMount": {non, non, non, add(-1), non, non, non, non, non, non, non, non, non},
		"Luxury":      {non, non, non, non, non, non, non, non, non, non, non, non, mul(2)},
		"Fast":        {add(1), add(1), add(1), add(-2), non, non, non, non, non, non, non, non, add(30)},
		"PassengerModule": {
			non, non, non, add(-3), non, non, non, non, non, non, non, non,
			add(100),
		},
		"CargoModule": {non, add(1), add(-1), add(1), non, non, non, non, non, non, non, non, add(20)},
		"Redundancy":  {add(1), add(1), non, non, non, non, non, non, non, non, non, non, add(60)},
		// J More Options.
		"OffRoad":    {non, non, non, non, non, non, non, non, non, non, non, non, add(30)},
		"Mole":       {add(1), mul(3), set(1), non, non, non, non, non, non, non, non, non, add(400)},
		"Hydrofoils": {add(1), add(1), add(1), non, non, non, non, non, non, non, non, non, add(30)},
		"Stubs":      {non, non, non, non, non, non, non, non, non, non, non, non, add(20)},
		"VTOL":       {non, non, add(-1), add(-2), non, non, non, non, non, non, non, non, add(100)},
		"STOL":       {non, non, non, add(-1), non, non, non, non, non, non, non, non, add(50)},
		"LiftingBody": {
			non, add(4), add(1), mul(2), non, non, non, non, non, non, non, non,
			add(200),
		},
		"Wings1": {non, mul(2), add(1), mul(1), non, non, non, non, non, non, non, non, add(100)},
		"Wings2": {non, mul(3), add(2), mul(2), non, non, non, non, non, non, non, non, add(200)},
		"Wings3": {non, mul(4), add(3), mul(3), non, non, non, non, non, non, non, non, add(300)},
		"Floats": {non, add(-1), add(-1), non, non, non, non, non, non, non, non, non, add(100)},
		"ParasiteNipple": {
			add(1), non, non, add(-1), non, non, non, non, non, non, non, non,
			add(100),
		},
	}

	if len(want) != len(enhancerRows) {
		t.Fatalf("page transcription covers %d rows, registry has %d", len(want), len(enhancerRows))
	}

	for code, cols := range want {
		row, ok := Enhancer(code)
		if !ok {
			t.Errorf("Enhancer(%q) missing", code)

			continue
		}

		got := [13]Operation{
			row.TechLevel, row.Tons, row.Speed, row.Load, row.Armor, row.Cage, row.Flash,
			row.Radiation, row.Sound, row.Psi, row.Insulated, row.Sealed, row.CostKCr,
		}
		if !reflect.DeepEqual(got, cols) {
			t.Errorf("Enhancer(%q) columns = %+v, want %+v", code, got, cols)
		}
	}
}

func TestDesignBoxRows(t *testing.T) {
	rows := DesignBoxRows()
	if len(rows) != 20 {
		t.Fatalf("DesignBoxRows = %d, want 20", len(rows))
	}

	for _, row := range rows {
		cubes := row.LengthSquares * row.WidthSquares * row.HeightSquares
		if math.Abs(row.Tons-float64(cubes)/4) > .5 {
			t.Errorf("dimensions %+v = %d cubes = %g tons", row, cubes, float64(cubes)/4)
		}
	}

	rows[0].Tons = 99
	if DesignBoxRows()[0].Tons == 99 {
		t.Error("DesignBoxRows exposed mutable chart storage")
	}
}

func TestTerrainTables(t *testing.T) {
	for _, c := range []struct {
		terrain, vehicle string
		want             Access
	}{
		{"Highway", "Cars", Accessible},
		{"Grid", "Cars", GridOnly},
		{"Ocean", "Cars", Prohibited},
		{"Mountain", "Track", Accessible},
		{"Wetland", "ACV", Accessible},
	} {
		if got, ok := SurfaceAccess(c.terrain, c.vehicle); !ok || got != c.want {
			t.Errorf("SurfaceAccess(%q,%q) = %v,%v want %v", c.terrain, c.vehicle, got, ok, c.want)
		}
	}

	if got, ok := SeafaringAccess("Ocean", "Ship"); !ok || got != Accessible {
		t.Errorf("Ocean Ship = %v,%v", got, ok)
	}

	if got, ok := SeafaringAccess("River", "Sub"); !ok || got != Disallowed {
		t.Errorf("River Sub = %v,%v", got, ok)
	}

	for _, c := range []struct {
		terrain, flyer string
		want           Access
	}{
		{"Orbit", "Wing", Prohibited},
		{"Orbit", "Lifter", Accessible},
		{"Atm2", "Wing", Accessible},
		{"Atm2", "Rotor", Prohibited},
		{"Atm8", "Flap", Accessible},
		{"AtmD", "LTA", Conditional},
		{"Under5m", "Rotor", Temporary},
	} {
		if got, ok := FlyerAccess(c.terrain, c.flyer); !ok || got != c.want {
			t.Errorf("FlyerAccess(%q,%q) = %v,%v want %v", c.terrain, c.flyer, got, ok, c.want)
		}
	}
}

func TestAltitudeAndDepthTables(t *testing.T) {
	altitudes := Altitudes()
	if len(altitudes) != 15 || altitudes[0].Range != "10" || altitudes[len(altitudes)-1].Level != "Surface" {
		t.Errorf("altitude table boundaries = %+v ... %+v", altitudes[0], altitudes[len(altitudes)-1])
	}

	depths := Depths()
	if len(depths) != 17 || depths[8].Pressure != 1 || depths[15].Pressure != 50_000 {
		t.Errorf("depth pressure rows incorrect: %+v", depths)
	}
}
