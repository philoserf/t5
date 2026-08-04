package mastermod

import (
	"slices"
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestLookupBoundaries(t *testing.T) {
	tbl := table("Example", "1D", 1, "one", "two")
	if _, ok := tbl.Lookup(0); ok {
		t.Error("below range succeeded")
	}

	if _, ok := tbl.Lookup(3); ok {
		t.Error("above range succeeded")
	}

	if got, ok := tbl.Lookup(1); !ok || got != "one" {
		t.Errorf("lookup = %q,%v", got, ok)
	}

	if got, ok := tbl.Lookup(2); !ok || got != "two" {
		t.Errorf("lookup = %q,%v", got, ok)
	}
}

func TestInventoryAndGoldenRows(t *testing.T) {
	if got := Names(); !slices.Equal(got, manifest) {
		t.Fatalf("registry names diverge from the manifest:\ngot  %q\nwant %q", got, manifest)
	}

	cases := map[string]struct {
		roll int
		want string
	}{
		"Device Damage Location":      {6, "Processor"},
		"Anatomical Damage Location":  {9, "Limb-Grip-3"},
		"Theme4":                      {4, "Revenge"},
		"Space Sensors":               {6, "Distress Call"},
		"Reported Fault":              {3, "Power System"},
		"Probability":                 {-6, "Impossible"},
		"Major Races":                 {5, "Droyne"},
		"Terrain2":                    {5, "Frozen Lands"},
		"Supply":                      {6, "Unique"},
		"Large Groups":                {4, "10,000"},
		"QREBS":                       {15, "E+2 B+1 S-1"},
		"Technology Fantastic (Low)":  {1, "27"},
		"Technology Fantastic (High)": {6, "28"},
		"Walk":                        {5, "Ssonic"},
		"NewSpeak":                    {-5, "Worst"},
		"Truth":                       {-5, "GAEWK"},
	}
	for name, tc := range cases {
		tbl, ok := Get(name)
		if !ok {
			t.Fatalf("missing %q", name)
		}

		if got, ok := tbl.Lookup(tc.roll); !ok || got != tc.want {
			t.Errorf("%s[%d] = %q,%v; want %q", name, tc.roll, got, ok, tc.want)
		}
	}
}

func TestSparseLookup(t *testing.T) {
	tbl, _ := Get("Armor Mods")
	if _, ok := tbl.Lookup(-4); ok {
		t.Error("blank source row became an entry")
	}

	if got, ok := tbl.Lookup(-2); !ok || got != "Armor" {
		t.Errorf("Armor Mods[-2] = %q,%v", got, ok)
	}
}

// knownSpecialDice are the roll notations the appendix uses that
// dice.Parse does not accept, allowlisted by exact string: "Bad Flux"
// (roll-low flux), "2x1D" (two 1D rolls read as digits, 11..66), and
// "Hits/2" (a derived value, not a die roll).
var knownSpecialDice = map[string]bool{
	"Bad Flux": true,
	"2x1D":     true,
	"Hits/2":   true,
}

func TestDiceNotationParseable(t *testing.T) {
	for _, name := range Names() {
		tbl, _ := Get(name)
		if knownSpecialDice[tbl.Dice] {
			continue
		}

		if _, err := dice.Parse(tbl.Dice); err != nil {
			t.Errorf("%q: dice notation %q is neither Parse-able nor allowlisted: %v", name, tbl.Dice, err)
		}
	}
}

func TestGetReturnsCopy(t *testing.T) {
	tbl, ok := Get("Armor Mods")
	if !ok {
		t.Fatal("missing Armor Mods")
	}

	tbl.Rows[0] = "poisoned"
	tbl.Rolls[0] = -99

	again, _ := Get("Armor Mods")
	if again.Rows[0] != "Hvy Armor" || again.Rolls[0] != -3 {
		t.Errorf("registry mutated through Get: Rows[0]=%q Rolls[0]=%d", again.Rows[0], again.Rolls[0])
	}
}

func TestSparseEmptyRollsPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("sparse with no rolls did not panic")
		}
	}()

	sparse("Empty", "1D", nil)
}

// manifest is the full sorted inventory of registered table names — a
// swapped, renamed, or dropped table fails here even if the count survives.
var manifest = []string{
	"Acceleration",
	"Anatomical Damage Location",
	"Anglic Brands",
	"Armor Mods",
	"Aroma",
	"Attitude",
	"Beauty",
	"Biological Damage Location",
	"Bulk/Burden",
	"Careers",
	"Climate",
	"Comms",
	"Comparatives",
	"Conformity",
	"Crime",
	"Damage Severity",
	"Degree",
	"Demand",
	"Device Damage Location",
	"Diagnosis",
	"Diagnosis Severity",
	"Doctrine",
	"Drive",
	"Ease Of Use",
	"Economic",
	"Emotion",
	"Environment",
	"Environmental Conditions",
	"Environmental Crime",
	"Evasion Mods",
	"Evidence and Proof",
	"Fraction",
	"Friends",
	"Gravity",
	"Gravity G",
	"Groups1",
	"Groups2",
	"Heavy Weapon Damage Location",
	"Highway",
	"Historical Ruins",
	"Humaniti",
	"Idea",
	"Imagination",
	"Imminence",
	"Imperial Brands",
	"Injury",
	"Injury Severity",
	"Justice",
	"Large Groups",
	"Light",
	"Local Anomaly",
	"Local Observation",
	"Local Unrest",
	"Logic",
	"Major Races",
	"MegaCorp Brands",
	"MegaCorporations",
	"Minor Planets",
	"Morality",
	"Naval Time In Jump",
	"NewSpeak",
	"Nobility",
	"Number of Careers",
	"Order and Chaos",
	"Outer System",
	"Planetary",
	"Political",
	"Population",
	"Potential",
	"Probability",
	"Property Crime",
	"QREBS",
	"Quality",
	"Quality Mod",
	"Quality Period",
	"Reliability",
	"Remote System",
	"Reported Fault",
	"Respect",
	"Rewards",
	"Safety",
	"Secondary",
	"Sensation",
	"Sense Emotion",
	"Sense Visibility",
	"Sensor Responding",
	"Severity",
	"Society Crime",
	"Sophont Crime",
	"Sophonts1",
	"Sophonts2",
	"Sound",
	"Sounds",
	"Space Sensors",
	"Special",
	"Specialized Sensors",
	"Speed kph",
	"Stability",
	"Standard Time In Jump",
	"Starport Situations",
	"Startown Situations",
	"Stellar Anomalies",
	"Strange Warnings",
	"Supply",
	"Surprise",
	"Technology Fantastic (High)",
	"Technology Fantastic (Low)",
	"Technology High",
	"Technology Low",
	"Technology Med",
	"Technology UHigh",
	"Technology VHigh",
	"Technology VLow",
	"Technology XHigh",
	"Terrain1",
	"Terrain2",
	"Theme1",
	"Theme2",
	"Theme3",
	"Theme4",
	"Theme5",
	"Theme6",
	"Tool Damage Location",
	"Touch",
	"Truth",
	"Typical BR",
	"Typical DH",
	"Uncertainty",
	"Vehicle or Armor Damage Location",
	"Vilani Brands",
	"Vilani Comparatives",
	"Visibility",
	"Walk",
	"Weapon Damage Location",
	"Weather",
	"World Sensors",
	"World Size",
	"Xeno-Med",
	"Zero-G",
}
