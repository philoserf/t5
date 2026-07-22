package mastermod

import "testing"

func TestLookupBoundaries(t *testing.T) {
	tbl := table("Example", "1D", 1, "one", "<Thing> two")
	if _, ok := tbl.Lookup(0, nil); ok {
		t.Error("below range succeeded")
	}

	if _, ok := tbl.Lookup(3, nil); ok {
		t.Error("above range succeeded")
	}

	if got, ok := tbl.Lookup(1, nil); !ok || got != "one" {
		t.Errorf("lookup = %q,%v", got, ok)
	}

	if got, ok := tbl.Lookup(2, map[string]string{"Thing": "stable"}); !ok || got != "stable two" {
		t.Errorf("substitution = %q,%v", got, ok)
	}
}

func TestInventoryAndGoldenRows(t *testing.T) {
	if len(Names()) != 138 {
		t.Fatalf("inventory = %d, want 138", len(Names()))
	}

	cases := map[string]struct {
		roll int
		want string
	}{
		"Device Damage Location":     {6, "Processor"},
		"Anatomical Damage Location": {9, "Limb-Grip-3"},
		"Theme4":                     {4, "Revenge"},
		"Space Sensors":              {6, "Distress Call"},
		"Reported Fault":             {3, "Power System"},
		"Probability":                {-6, "Impossible"},
		"Major Races":                {5, "Droyne"},
		"Terrain2":                   {5, "Frozen Lands"},
		"Supply":                     {6, "Unique"},
		"Large Groups":               {4, "10,000"},
		"QREBS":                      {15, "E+2 B+1 S-1"},
	}
	for name, tc := range cases {
		tbl, ok := Get(name)
		if !ok {
			t.Fatalf("missing %q", name)
		}

		if got, ok := tbl.Lookup(tc.roll, nil); !ok || got != tc.want {
			t.Errorf("%s[%d] = %q,%v; want %q", name, tc.roll, got, ok, tc.want)
		}
	}
}

func TestSparseLookup(t *testing.T) {
	tbl, _ := Get("Armor Mods")
	if _, ok := tbl.Lookup(-4, nil); ok {
		t.Error("blank source row became an entry")
	}

	if got, ok := tbl.Lookup(-2, nil); !ok || got != "Armor" {
		t.Errorf("Armor Mods[-2] = %q,%v", got, ok)
	}
}

func TestRegistryValid(t *testing.T) {
	for _, name := range Names() {
		tbl, _ := Get(name)
		if !tbl.Valid() {
			t.Errorf("%q invalid", name)
		}
	}
}
