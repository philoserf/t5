package worldgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/philoserf/t5/internal/uwp"
)

// port builds a profile for facilities tests: a solid world (Size 5, so it is not
// mistaken for an asteroid belt), dry (Hyd 0, so no local-fuel fallback), at a low
// tech level (TL 5, below every exotic-fuel threshold). Tests that exercise those
// features set the relevant field.
func port(class byte, pop int) uwp.Profile {
	return uwp.Profile{Starport: class, Size: 5, Population: pop, TechLevel: 5}
}

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
		{
			'A',
			6,
			"Excellent",
			"Starships",
			Overhaul,
			RefinedFuel,
			true,
			false,
			false,
			"2D",
		}, // Pop below highport threshold
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
		f, ok := PortFacilities(port(c.class, c.pop), false)
		if !ok {
			t.Errorf("PortFacilities(%c) not found", c.class)

			continue
		}

		if f.Quality != c.quality || f.Shipyard != c.shipyard || f.Repairs != c.repairs ||
			f.Fuel != c.fuel || (f.port == portDown) != c.downport || f.Beacon != c.beacon ||
			f.Highport != c.highport || f.RefuelHours != c.refuel {
			t.Errorf("PortFacilities(%c, pop %d) = %+v", c.class, c.pop, f)
		}
	}

	if _, ok := PortFacilities(port('Z', 5), false); ok {
		t.Errorf("PortFacilities(Z) should be unknown")
	}
}

func TestFuelAndRepairStrings(t *testing.T) {
	if RefinedFuel.String() != "Refined+Unrefined" || UnrefinedFuel.String() != "Unrefined" ||
		NoFuel.String() != "None" {
		t.Errorf("FuelKind.String mismatch")
	}

	if Overhaul.String() != "Overhaul" || MajorRepairs.String() != "Major" ||
		NoRepairs.String() != "None" {
		t.Errorf("RepairLevel.String mismatch")
	}
}

// TestServices golden-locks the prose a port advertises (Book 2 p.24): a class-A
// port lists everything, a beacon port (E) names its beacon-only downport, and a
// class-X world — no port at all — advertises nothing.
func TestServices(t *testing.T) {
	a, _ := PortFacilities(port('A', 8), false)
	got := strings.Join(a.Services(), " · ")

	want := "builds Starships · repairs: Overhaul · fuel: Refined+Unrefined (2D hours) · downport · highport"
	if got != want {
		t.Errorf("class-A services =\n%q\nwant\n%q", got, want)
	}
	// E is a beacon-only downport with no fuel and no repairs.
	e, _ := PortFacilities(port('E', 5), false)
	if got := strings.Join(e.Services(), " · "); got != "beacon-only downport" {
		t.Errorf("class-E services = %q, want %q", got, "beacon-only downport")
	}
	// X is no port at all: it must advertise nothing, not "repairs: None".
	x, _ := PortFacilities(port('X', 5), false)
	if svc := x.Services(); len(svc) != 0 {
		t.Errorf("class-X advertises services it does not have: %v", svc)
	}
}

// TestExoticFuels: class A and B carry exotic fuels as their tech level reaches
// each threshold (Book 3 p.24) — Radioactives TL 8, Collector TL 14, Anti-Matter
// TL 18 — and no other class offers any.
func TestExoticFuels(t *testing.T) {
	cases := []struct {
		tl   int
		want []string
	}{
		{7, nil},
		{8, []string{"Radioactives"}},
		{14, []string{"Radioactives", "Collector"}},
		{18, []string{"Radioactives", "Collector", "Anti-Matter"}},
	}
	for _, c := range cases {
		p := port('A', 8)
		p.TechLevel = c.tl

		f, _ := PortFacilities(p, false)
		if !slices.Equal(f.ExoticFuels, c.want) {
			t.Errorf("class-A TL%d exotic fuels = %v, want %v", c.tl, f.ExoticFuels, c.want)
		}
	}
	// A class-C starport offers no exotic fuel, however high its tech level.
	c := port('C', 9)

	c.TechLevel = 18
	if f, _ := PortFacilities(c, false); len(f.ExoticFuels) != 0 {
		t.Errorf("class-C offers exotic fuel: %v", f.ExoticFuels)
	}
}

// TestLocalFuel: a fuel-less port on a world with water or ice can still be
// refuelled unrefined from the environment (Book 3 p.24, the ** note) — the
// difference between a slow refuel and being stranded.
func TestLocalFuel(t *testing.T) {
	// A class-E frontier port with oceans: unrefined local fuel is available.
	wet := port('E', 4)
	wet.Hydrographics = 6

	f, _ := PortFacilities(wet, false)
	if !f.LocalFuel {
		t.Errorf("a class-E port on a wet world should have local fuel")
	}

	if !slices.Contains(f.Services(), "fuel: Unrefined (local water/ice)") {
		t.Errorf("local fuel not advertised: %v", f.Services())
	}
	// A dry class-E world (Hyd 0): nothing to skim.
	if f, _ := PortFacilities(port('E', 4), false); f.LocalFuel {
		t.Errorf("a dry class-E world should have no local fuel")
	}
	// A port that already has fuel does not gain a redundant local-fuel line.
	if f, _ := PortFacilities(port('C', 9), false); f.LocalFuel {
		t.Errorf("a class-C port already has fuel; local fuel should be false")
	}
}

// TestBeltport: an asteroid-belt mainworld (Size 0) has a Beltport in place of a
// downport (Book 2 p.24).
func TestBeltport(t *testing.T) {
	profile := uwp.Profile{Starport: 'C', Size: 0, Population: 9, TechLevel: 5}

	f, _ := PortFacilities(profile, true)
	if f.port != portBelt {
		t.Errorf("an asteroid mainworld should have a beltport, not a downport: %+v", f)
	}

	if !slices.Contains(f.Services(), "beltport") {
		t.Errorf("beltport not advertised: %v", f.Services())
	}
	// A class-X asteroid belt has no port at all, so no beltport either.
	if f, _ := PortFacilities(uwp.Profile{Starport: 'X', Size: 0}, true); f.port == portBelt {
		t.Errorf("a class-X belt has no port, so no beltport")
	}
	// #324: the same Size-0 profile that is NOT a belt (a tiny solid world) keeps
	// its downport. Belt-ness is the caller's fact, not the Size digit — passing
	// belt=false must not earn a beltport.
	if f, _ := PortFacilities(profile, false); f.port != portDown {
		t.Errorf("a non-belt Size-0 world must keep its downport, not get a beltport: %+v", f)
	}
}
