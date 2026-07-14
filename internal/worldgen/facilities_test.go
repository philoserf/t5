package worldgen

import (
	"strings"
	"testing"
)

func TestPortFacilities(t *testing.T) {
	cases := []struct {
		class    byte
		pop      int
		quality  string
		shipyard string
		repairs  RepairLevel
		fuel     FuelKind
		downport bool
		beacon   bool
		highport bool
		refuel   string
	}{
		{'A', 7, "Excellent", "Starships", Overhaul, RefinedFuel, true, false, true, "2D"},
		{'A', 6, "Excellent", "Starships", Overhaul, RefinedFuel, true, false, false, "2D"}, // Pop below highport threshold
		{'B', 8, "Good", "Spacecraft", Overhaul, RefinedFuel, true, false, true, "2D"},
		{'C', 9, "Routine", "", MajorRepairs, UnrefinedFuel, true, false, true, "4D"},
		{'C', 8, "Routine", "", MajorRepairs, UnrefinedFuel, true, false, false, "4D"},
		{'D', 12, "Poor", "", MinorRepairs, UnrefinedFuel, true, false, false, "4D"},
		{'E', 12, "Frontier", "", NoRepairs, NoFuel, true, true, false, ""},
		{'X', 12, "None", "", NoRepairs, NoFuel, false, false, false, ""},
		{'F', 5, "Good", "", MinorRepairs, UnrefinedFuel, true, false, false, "4D"},
		{'G', 5, "Poor", "", SuperficialRepairs, UnrefinedFuel, true, false, false, "4D"},
		{'H', 5, "Basic", "", NoRepairs, NoFuel, true, true, false, ""},
		{'Y', 5, "None", "", NoRepairs, NoFuel, false, false, false, ""},
	}
	for _, c := range cases {
		f, ok := PortFacilities(c.class, c.pop)
		if !ok {
			t.Errorf("PortFacilities(%c) not found", c.class)
			continue
		}
		if f.Quality != c.quality || f.Shipyard != c.shipyard || f.Repairs != c.repairs ||
			f.Fuel != c.fuel || f.Downport != c.downport || f.Beacon != c.beacon ||
			f.Highport != c.highport || f.RefuelHours != c.refuel {
			t.Errorf("PortFacilities(%c, pop %d) = %+v", c.class, c.pop, f)
		}
	}
	if _, ok := PortFacilities('Z', 5); ok {
		t.Errorf("PortFacilities(Z) should be unknown")
	}
}

func TestFuelAndRepairStrings(t *testing.T) {
	if RefinedFuel.String() != "Refined+Unrefined" || UnrefinedFuel.String() != "Unrefined" || NoFuel.String() != "None" {
		t.Errorf("FuelKind.String mismatch")
	}
	if Overhaul.String() != "Overhaul" || MajorRepairs.String() != "Major" || NoRepairs.String() != "None" {
		t.Errorf("RepairLevel.String mismatch")
	}
}

// TestServices golden-locks the prose a port advertises (Book 2 p.24): a class-A
// port lists everything, a beacon port (E) names its beacon-only downport, and a
// class-X world — no port at all — advertises nothing.
func TestServices(t *testing.T) {
	a, _ := PortFacilities('A', 8)
	got := strings.Join(a.Services(), " · ")
	want := "builds Starships · repairs: Overhaul · fuel: Refined+Unrefined (2D hours) · downport · highport"
	if got != want {
		t.Errorf("class-A services =\n%q\nwant\n%q", got, want)
	}
	// E is a beacon-only downport with no fuel and no repairs.
	e, _ := PortFacilities('E', 5)
	if got := strings.Join(e.Services(), " · "); got != "beacon-only downport" {
		t.Errorf("class-E services = %q, want %q", got, "beacon-only downport")
	}
	// X is no port at all: it must advertise nothing, not "repairs: None".
	x, _ := PortFacilities('X', 5)
	if svc := x.Services(); len(svc) != 0 {
		t.Errorf("class-X advertises services it does not have: %v", svc)
	}
}
