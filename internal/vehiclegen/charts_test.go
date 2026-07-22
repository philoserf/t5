package vehiclegen

import (
	"math"
	"testing"
)

func TestCreationChartRegistriesComplete(t *testing.T) {
	wantTypes := []string{
		"G:GC", "G:U", "G:T", "G:V", "G:M", "G:H", "G:R",
		"M:T", "M:C", "M:V", "M:R", "M:W",
		"F:F", "F:G", "F:B",
		"W:S", "W:U", "W:B",
	}
	wantMissions := []string{
		"G:RO", "G:P", "G:C", "G:MP", "G:OR",
		"M:T", "M:S", "M:R",
		"F:A", "F:B", "F:C", "F:P", "F:S", "F:U",
		"W:C", "W:P", "W:E", "W:T",
	}
	wantMotives := []string{
		"G:AC", "G:W", "G:Z", "G:G", "G:T",
		"M:AC", "M:W", "M:Z", "M:G", "M:T",
		"F:W", "F:R", "F:F", "F:LTA", "F:Z", "F:G",
		"W:S", "W:U", "W:H", "W:G",
	}

	assertRegistry(t, "type", typeRows, wantTypes)
	assertRegistry(t, "mission", missionRows, wantMissions)
	assertRegistry(t, "motive", motiveRows, wantMotives)
}

func TestDesignMilitaryWeaponType(t *testing.T) {
	v := Design(Spec{Category: "M", Type: "W", Mission: "T", Motive: "W"})
	if len(v.Problems) != 0 {
		t.Fatalf("military Weapon design problems = %v", v.Problems)
	}

	if v.TechLevel != 6 || v.Tons != 3 || v.Speed != 5 || v.CostKCr != 100 {
		t.Errorf("military Weapon values = %+v", v.Values)
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

func TestEnhancerColumns(t *testing.T) {
	base := Values{
		TechLevel: 8, Tons: 12, Speed: 6, Load: 4, Armor: 10, Cage: 10,
		Flash: 10, Radiation: 10, Sound: 10, Psi: 10, Insulated: 10, Sealed: 10, CostKCr: 100,
	}

	for _, c := range []struct {
		code string
		want Values
	}{
		{"H", Values{
			TechLevel: 9, Tons: 24, Speed: 5, Load: 6, Armor: 20, Cage: 20,
			Flash: 10, Radiation: 10, Sound: 10, Psi: 10, Insulated: 20, Sealed: 20, CostKCr: 300,
		}},
		{"Adv", Values{
			TechLevel: 11, Tons: 10, Speed: 7, Load: 5, Armor: 30, Cage: 30,
			Flash: 10, Radiation: 10, Sound: 10, Psi: 10, Insulated: 10, Sealed: 10, CostKCr: 140,
		}},
		{"Armored", Values{
			TechLevel: 10, Tons: 13, Speed: 6, Load: 4, Armor: 30, Cage: 20,
			Flash: 20, Radiation: 20, Sound: 22, Psi: 0, Insulated: 30, Sealed: 30, CostKCr: 130,
		}},
		{"HighPowered", Values{
			TechLevel: 9, Tons: 13, Speed: 7, Load: 3, Armor: 10, Cage: 10,
			Flash: 10, Radiation: 10, Sound: 10, Psi: 10, Insulated: 10, Sealed: 10, CostKCr: 200,
		}},
	} {
		row, _ := Enhancer(c.code)

		got := applyTo(base, row)
		if got != c.want {
			t.Errorf("%s applied = %+v, want %+v", c.code, got, c.want)
		}
	}
}

func applyTo(base Values, row Modifier) Values {
	return Values{
		TechLevel: int(row.TechLevel.apply(float64(base.TechLevel))),
		Tons:      row.Tons.apply(base.Tons), Speed: int(row.Speed.apply(float64(base.Speed))),
		Load: row.Load.apply(base.Load), Armor: int(row.Armor.apply(float64(base.Armor))),
		Cage: int(row.Cage.apply(float64(base.Cage))), Flash: int(row.Flash.apply(float64(base.Flash))),
		Radiation: int(row.Radiation.apply(float64(base.Radiation))),
		Sound:     int(row.Sound.apply(float64(base.Sound))), Psi: int(row.Psi.apply(float64(base.Psi))),
		Insulated: int(row.Insulated.apply(float64(base.Insulated))),
		Sealed:    int(row.Sealed.apply(float64(base.Sealed))), CostKCr: row.CostKCr.apply(base.CostKCr),
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
