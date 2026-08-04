package vehiclegen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestVehicleSpeeds(t *testing.T) {
	want := []int{0, 5, 10, 20, 30, 50, 100, 300, 500, 700, 1000, 2000, 3000, 5000, 10000, 20000, 30000}
	for speed, kph := range want {
		if got, ok := SpeedKPH(speed); !ok || got != kph {
			t.Errorf("SpeedKPH(%d) = %d, %v; want %d, true", speed, got, ok, kph)
		}
	}

	for _, speed := range []int{-1, 17, 19, 20} {
		if _, ok := SpeedKPH(speed); ok {
			t.Errorf("SpeedKPH(%d) found a value absent from the table", speed)
		}
	}
}

func TestCollisionDamage(t *testing.T) {
	for _, speed := range []int{0, 1, 7, 13, 19} {
		got, ok := CollisionDicePerTon(speed)
		if !ok || got != speed*speed {
			t.Errorf("CollisionDicePerTon(%d) = %d, %v", speed, got, ok)
		}
	}
}

func TestBeastpower(t *testing.T) {
	for _, c := range []struct {
		tons  float64
		speed int
		want  float64
	}{
		{1, 1, 1},
		{1, 6, 216},
		{0.5, 4, 32}, // one Design Box square at Speed 4, p.146
		{10, 13, 21_970},
	} {
		if got := Beastpower(c.tons, c.speed); got != c.want {
			t.Errorf("Beastpower(%g, %d) = %g, want %g", c.tons, c.speed, got, c.want)
		}
	}
}

func TestEnduranceRange(t *testing.T) {
	for _, c := range []struct {
		endurance Endurance
		speed     int
		want      Range
	}{
		{Minutes, 13, Local},
		{Hours, 4, Local},
		{Hours, 5, Regional},
		{Hours, 7, Continental},
		{Hours, 10, World},
		{Days, 3, Regional},
		{Days, 4, Continental},
		{Days, 7, World},
		{Weeks, 1, Continental},
		{Weeks, 2, World},
		{Months, 1, Continental},
		{Year, 1, World},
	} {
		if got, ok := EnduranceRange(c.endurance, c.speed); !ok || got != c.want {
			t.Errorf("EnduranceRange(%d, %d) = %d, %v; want %d", c.endurance, c.speed, got, ok, c.want)
		}
	}
}

func TestOccupants(t *testing.T) {
	for _, c := range []struct {
		bulk      Bulk
		tons      float64
		endurance Endurance
		want      int
	}{
		{VLight, 0.5, Hours, 1},
		{VLight, 0.5, Days, 0},
		{Light, 2, Hours, 2},
		{Light, 2, Days, 1},
		{Light, 2, Weeks, 0},
		{Medium, 10, Hours, 10},
		{Heavy, 10, Days, 5},
		{VHeavy, 10, Weeks, 3},
		{Medium, 10, Months, 2},
		{Medium, 10, Year, 2},
	} {
		if got := Occupants(c.bulk, c.tons, c.endurance); got != c.want {
			t.Errorf("Occupants(%d, %g, %d) = %d, want %d", c.bulk, c.tons, c.endurance, got, c.want)
		}
	}
}

func TestDesignGroundVehicle(t *testing.T) {
	// Book 3 p.150: GroundCar Tons 2 / Load 1 / KCr 20; Passenger Armor 5 /
	// Insulated 12 / KCr 10; Wheeled TL 6 / Speed 5 / x1.
	v := Design(Spec{Category: "G", Type: "GC", Mission: "P", Motive: "W"})
	if len(v.Problems) != 0 {
		t.Fatalf("Design: %v", v.Problems)
	}

	want := Values{TechLevel: 6, Tons: 2, Speed: 5, Load: 1, Armor: 5, Insulated: 12, CostKCr: 30}
	if v.Values != want {
		t.Errorf("GroundCar = %+v, want %+v", v.Values, want)
	}

	if v.Beastpower != 250 {
		t.Errorf("GroundCar BP = %g, want 250", v.Beastpower)
	}
}

func TestDesignEnduranceRows(t *testing.T) {
	// Chart 13 (Book 3 p.153): Days..Year add TL +1..+4 and 1..4 x Speed tons;
	// KCr 20/50/100 for Days/Weeks/Months, none printed for Year.
	base := Design(Spec{Category: "G", Type: "T", Mission: "C", Motive: "W"})

	for _, c := range []struct {
		endurance Endurance
		tl        int
		tons      float64
		kcr       float64
	}{
		{Minutes, 0, 0, 0},
		{Hours, 0, 0, 0},
		{Days, 1, 5, 20},
		{Weeks, 2, 10, 50},
		{Months, 3, 15, 100},
		{Year, 4, 20, 0},
	} {
		v := Design(Spec{Category: "G", Type: "T", Mission: "C", Motive: "W", Endurance: c.endurance})
		if v.TechLevel != base.TechLevel+c.tl || v.Tons != base.Tons+c.tons ||
			v.CostKCr != base.CostKCr+c.kcr {
			t.Errorf("endurance %d = %+v, want +%d TL, +%g tons, +%g KCr over %+v",
				c.endurance, v.Values, c.tl, c.tons, c.kcr, base.Values)
		}
	}
}

func TestDesignLTATonnageAndBeastpowerShift(t *testing.T) {
	// Book 3 p.151: "LTA final tonnage equals 10x the calculated tonnage."
	lta := Design(Spec{Category: "F", Type: "B", Mission: "C", Motive: "LTA"})
	if lta.Tons != 1600 { // (40 x4) x10
		t.Errorf("Balloon Cargo tons = %g, want 1600", lta.Tons)
	}

	// Book 3 p.146 column shift: Watercraft Ship reads BP at Speed +3.
	ship := Design(Spec{Category: "W", Type: "S", Mission: "T", Motive: "S"})
	if want := 1000 * float64(7*7*7); ship.Beastpower != want {
		t.Errorf("Ship BP = %g, want %g (Speed 4 shifted +3)", ship.Beastpower, want)
	}

	armored, _ := Enhancer("Armored")

	tank := Design(Spec{Category: "M", Type: "T", Mission: "W", Motive: "T", Enhancers: []Modifier{armored}})
	if want := tank.Tons * float64(5*5*5); tank.Beastpower != want {
		t.Errorf("Armored tank BP = %g, want %g (Speed 3 shifted +2)", tank.Beastpower, want)
	}
}

func TestDesignFlyerAndWatercraft(t *testing.T) {
	flyer := Design(Spec{Category: "F", Type: "F", Mission: "A", Motive: "W"})
	if len(flyer.Problems) != 0 || flyer.TechLevel != 9 || flyer.Tons != 20 || flyer.Speed != 9 ||
		flyer.Load != 4 || flyer.Armor != 20 || flyer.CostKCr != 900 {
		t.Errorf("Attack Winged Flyer = %+v", flyer)
	}

	boat := Design(Spec{Category: "W", Type: "B", Mission: "P", Motive: "H"})
	if len(boat.Problems) != 0 || boat.TechLevel != 8 || boat.Tons != 20 || boat.Speed != 10 ||
		boat.Load != 9 || boat.Armor != 10 || boat.CostKCr != 300 {
		t.Errorf("Patrol Hovercraft Boat = %+v", boat)
	}
}

func TestGenerateDeterministicAndValid(t *testing.T) {
	for seed := uint64(1); seed <= 30; seed++ {
		a := Generate(dice.NewWithSeed(seed))

		b := Generate(dice.NewWithSeed(seed))
		if a.Category != b.Category || a.Values != b.Values || len(a.Problems) != 0 {
			t.Fatalf("seed %d: generated designs differ or invalid: %+v / %+v", seed, a, b)
		}
	}
}

func TestDesignRejectsUnknownRows(t *testing.T) {
	v := Design(Spec{Category: "?", Type: "?", Mission: "?", Motive: "?"})
	if len(v.Problems) != 3 {
		t.Errorf("unknown spec problems = %v, want three", v.Problems)
	}
}
