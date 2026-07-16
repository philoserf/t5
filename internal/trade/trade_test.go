package trade

import (
	"slices"
	"testing"
)

// TestBeowulfJourney golden-locks the whole pricing engine to the Free Trader
// Beowulf worked example (Book 2 p.221).
func TestBeowulfJourney(t *testing.T) {
	// Cargo bought at Efate A646930-D Hi In An (TL 13): Cr3,000 -1,000 (Hi)
	// -1,000 (In) + 13x100 = Cr2,300; the Cargo ID drops the non-value An.
	if got := Cost(13, []string{"Hi", "In", "An"}); got != 2300 {
		t.Errorf("Efate cost = %d, want 2300", got)
	}
	if got := CargoID(13, []string{"Hi", "In", "An"}, ""); got != "D-Hi In Cr2,300" {
		t.Errorf("Efate Cargo ID = %q, want %q", got, "D-Hi In Cr2,300")
	}
	// Carried to Alell (market Ri, TL 10): source In matches market Ri (+1,000)
	// -> Cr6,000, x(1 + 10%x(13-10)) = Cr7,800; sold at Flux 0 -> 100%.
	priceAlell := Price(13, 10, []string{"Hi", "In"}, []string{"Ri"})
	if priceAlell != 7800 {
		t.Errorf("Alell price = %d, want 7800", priceAlell)
	}
	if got := SellingPrice(priceAlell, 0, 0); got != 7800 {
		t.Errorf("Alell sale at Flux 0 = %d, want 7800", got)
	}

	// Cargo A-Ri Cr5,000 bought at a TL-10 Rich world: 3,000 +1,000 (Ri) + 1,000.
	if got := Cost(10, []string{"Ri"}); got != 5000 {
		t.Errorf("Rich-world cost = %d, want 5000", got)
	}
	if got := CargoID(10, []string{"Ri"}, ""); got != "A-Ri Cr5,000" {
		t.Errorf("Cargo ID = %q, want %q", got, "A-Ri Cr5,000")
	}
	// Carried to Efate (market Hi In, TL 13): source Ri matches Hi and In
	// (+2,000) -> Cr7,000, x(1 + 10%x(10-13)) = Cr4,900.
	priceEfate := Price(10, 13, []string{"Ri"}, []string{"Hi", "In", "An"})
	if priceEfate != 4900 {
		t.Errorf("Efate price = %d, want 4900", priceEfate)
	}
	// Sold with a Broker-4 (DM +2) and a net Flux of +2 -> effective +4 -> 150%.
	if got := SellingPrice(priceEfate, 2, 4); got != 7350 {
		t.Errorf("Efate sale with Broker-4 = %d, want 7350", got)
	}
}

func TestValueClasses(t *testing.T) {
	// Only value classes survive, and they come out in chart order.
	got := ValueClasses([]string{"Po", "Na", "An", "De", "Cp", "Hi", "In"})
	want := []string{"De", "Hi", "In", "Na", "Po"}
	if !slices.Equal(got, want) {
		t.Errorf("ValueClasses = %v, want %v", got, want)
	}
}

func TestCostChartExample(t *testing.T) {
	// The p.221 chart example: TL 8, De Hi In Na Po -> Cr1,800.
	if got := Cost(8, []string{"De", "Hi", "In", "Na", "Po"}); got != 1800 {
		t.Errorf("cost = %d, want 1800", got)
	}
	if got := CargoID(
		8,
		[]string{"De", "Hi", "In", "Na", "Po"},
		"",
	); got != "8-De Hi In Na Po Cr1,800" {
		t.Errorf("Cargo ID = %q, want %q", got, "8-De Hi In Na Po Cr1,800")
	}
}

func TestBrokerDM(t *testing.T) {
	cases := map[int]int{0: 0, 1: 1, 2: 1, 3: 2, 4: 2, 8: 4, 9: 4, 12: 4}
	for skill, want := range cases {
		if got := BrokerDM(skill); got != want {
			t.Errorf("BrokerDM(%d) = %d, want %d", skill, got, want)
		}
	}
}

func TestActualValueRangeAndClamp(t *testing.T) {
	// Table endpoints and the -5..+8 clamp.
	cases := map[int]int{-5: 40, 0: 100, 5: 170, 8: 400}
	for flux, want := range cases {
		if got := ActualValuePercent(flux, 0); got != want {
			t.Errorf("ActualValuePercent(%d, 0) = %d, want %d", flux, got, want)
		}
	}
	// Below -5 clamps to 40%; a huge broker + high Flux clamps to +8 (400%).
	if got := ActualValuePercent(-9, 0); got != 40 {
		t.Errorf("ActualValuePercent(-9) = %d, want 40 (clamped)", got)
	}
	if got := ActualValuePercent(5, 12); got != 400 {
		t.Errorf("ActualValuePercent(+5, Broker 12) = %d, want 400 (clamped)", got)
	}
}

// TestImbalanceBonus locks the Book 2 p.211 rule: Imbalance goods sold into a
// market carrying the class whose oversupply produced them earn +Cr1,000.
func TestImbalanceBonus(t *testing.T) {
	knorbes := TradeGood{Name: "Pelts", Type: "Rares", Imbalance: "Na"}
	if got := ImbalanceBonus(knorbes, []string{"Na", "Ri"}); got != 1_000 {
		t.Errorf("ImbalanceBonus into an Na market = %d, want 1000", got)
	}
	if got := ImbalanceBonus(knorbes, []string{"Ag", "Ri"}); got != 0 {
		t.Errorf("ImbalanceBonus into a non-Na market = %d, want 0", got)
	}
	// Ordinary (non-Imbalance) goods never earn the bonus.
	if got := ImbalanceBonus(TradeGood{Name: "Antibiotics"}, []string{"Na"}); got != 0 {
		t.Errorf("ordinary goods earned an imbalance bonus: %d", got)
	}
}

// TestCargoIDNoValueClasses guards against the dangling separator a world with no
// value trade class used to produce ("8- Cr3,800").
func TestCargoIDNoValueClasses(t *testing.T) {
	if got := CargoID(8, []string{"Wa", "An"}, "Im"); got != "8 Cr3,800" {
		t.Errorf("CargoID(no value classes) = %q, want %q", got, "8 Cr3,800")
	}
	if got := CargoID(8, nil, "Im"); got != "8 Cr3,800" {
		t.Errorf("CargoID(nil) = %q, want %q", got, "8 Cr3,800")
	}
}

func TestCargoIDAllegiance(t *testing.T) {
	// A non-Imperial source appends its allegiance; Imperial and empty do not.
	if got := CargoID(13, []string{"Hi"}, "Zh"); got != "D-Hi Cr3,300 Zh" {
		t.Errorf("Cargo ID = %q, want %q", got, "D-Hi Cr3,300 Zh")
	}
	if got := CargoID(13, []string{"Hi"}, "Im"); got != "D-Hi Cr3,300" {
		t.Errorf("Imperial Cargo ID = %q, want %q", got, "D-Hi Cr3,300")
	}
}
