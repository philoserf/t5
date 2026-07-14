package shipgen

import "testing"

// murphyScout is a constructed Ship carrying the Book 2 p.42 worked-example
// numbers, used to pin the QSP and ship-card format before the design math
// exists (the shipgen analog of systemgen's reginaSystem fixture).
func murphyScout() Ship {
	return Ship{
		Spec: ShipSpec{
			Name: "Murphy-class Scout", Mission: "S", TL: 12,
			HullLetter: 1, Config: Lifting, Structure: Shell, ArmorLayers: 2,
			Maneuver: &DriveSpec{Letter: 1}, Jump: &DriveSpec{Letter: 1}, Power: &DriveSpec{Letter: 1},
			FuelScoop: true, FuelPurifier: true,
		},
		Hull:     Hull{Letter: 1, Tons: 100, Config: Lifting, Structure: Shell, BaseArmor: 6, Hardpoints: 1, Agility: 1, Stability: 3},
		Maneuver: &Drive{Kind: Maneuver, Letter: 1, Potential: 2},
		Jump:     &Drive{Kind: Jump, Letter: 1, Potential: 2},
		Power:    &Drive{Kind: Power, Letter: 1, Potential: 2},
		Armor:    Armor{Layers: 2, AV: 6},
		Fuel:     Fuel{Tons: 22},
		Cost:     70_300_000,
		Tonnage:  Budget{Hull: 100, Payload: 4},
	}
}

func TestQSP(t *testing.T) {
	if got := murphyScout().QSP(); got != "S-AL22" {
		t.Errorf("QSP = %q, want S-AL22", got)
	}
}

func TestShipCard(t *testing.T) {
	want := "Murphy-class Scout  S-AL22\n" +
		"Hull:    A 100t Lifting Body · TL-12 · Shell\n" +
		"Drives:  Maneuver-A 2G · Jump-A J-2 · Power-A\n" +
		"Armor:   2 layers AV-6\n" +
		"Fuel:    22t · Cost MCr 70.3 · Payload 4t"
	if got := murphyScout().String(); got != want {
		t.Errorf("String() =\n%q\nwant\n%q", got, want)
	}
}

func TestLetterOrdinal(t *testing.T) {
	// Hull letters are the eHex letters (no I/O) as ordinals 1..24 -> 100..2400t.
	cases := []struct {
		letter  byte
		ordinal int
		tons    int
	}{
		{'A', 1, 100}, {'B', 2, 200}, {'H', 8, 800}, {'J', 9, 900}, // I is skipped
		{'K', 10, 1000}, {'N', 13, 1300}, {'P', 14, 1400}, // O is skipped
		{'Z', 24, 2400},
	}
	for _, c := range cases {
		if got := LetterOrdinal(c.letter); got != c.ordinal {
			t.Errorf("LetterOrdinal(%c) = %d, want %d", c.letter, got, c.ordinal)
		}
		if got := ordinalLetter(c.ordinal); got != c.letter {
			t.Errorf("ordinalLetter(%d) = %c, want %c", c.ordinal, got, c.letter)
		}
		if got := HullTons(c.ordinal); got != c.tons {
			t.Errorf("HullTons(%d) = %d, want %d", c.ordinal, got, c.tons)
		}
	}
	// 'I' and 'O' are not letters in the scale; digits and junk map to 0.
	for _, c := range []byte{'I', 'O', '0', '9', '?'} {
		if got := LetterOrdinal(c); got != 0 {
			t.Errorf("LetterOrdinal(%c) = %d, want 0", c, got)
		}
	}
}

func TestConfigAndStructureLetters(t *testing.T) {
	if Lifting.Letter() != 'L' || Streamlined.Letter() != 'S' || Cluster.Letter() != 'C' {
		t.Errorf("config letters wrong")
	}
	if Lifting.String() != "Lifting Body" || Shell.String() != "Shell" {
		t.Errorf("names wrong")
	}
	if Config(99).Letter() != '?' || Structure(99).String() != "?" {
		t.Errorf("out-of-range guards wrong")
	}
}
