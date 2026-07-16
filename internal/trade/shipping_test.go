package trade

import "testing"

func TestPassageFare(t *testing.T) {
	// Premium Passage Pricing table endpoints (Book 2 p.220).
	cases := []struct {
		class  Passage
		demand int
		want   int
	}{
		{High, 0, 10_000},
		{Mid, 0, 8_000},
		{Low, 0, 1_000},
		{High, 5, 15_000},
		{Mid, 5, 13_000},
		{Low, 5, 1_500},
		{High, -5, 5_000},
		{Mid, -5, 3_000},
		{Low, -5, 500},
		{High, 3, 13_000},
		{Mid, 3, 11_000},
		{Low, 3, 1_300},
	}
	for _, c := range cases {
		if got := c.class.Fare(c.demand); got != c.want {
			t.Errorf("Passage(%d).Fare(%d) = %d, want %d", c.class, c.demand, got, c.want)
		}
	}
}

func TestAvailablePassengers(t *testing.T) {
	// Flux + Pop + skill mod, floored at zero.
	if got := AvailablePassengers(2, 8, 1); got != 11 {
		t.Errorf("AvailablePassengers(2,8,1) = %d, want 11", got)
	}
	if got := AvailablePassengers(-5, 1, 0); got != 0 {
		t.Errorf("AvailablePassengers(-5,1,0) = %d, want 0 (floored)", got)
	}
}

func TestAvailableFreight(t *testing.T) {
	// (Flux + Pop + Liaison) x (value trade classes + 1). Regina-ish: pop 8, and
	// a world whose codes carry 3 value classes (Hi, Ri, and Po; An is not one).
	tcs := []string{"Hi", "Ri", "Po", "An"}
	if got := AvailableFreight(0, 8, 0, tcs); got != 32 { // 8 x (3+1)
		t.Errorf("AvailableFreight(0,8,0) = %d, want 32", got)
	}
	if got := AvailableFreight(2, 8, 2, tcs); got != 48 { // (2+8+2) x 4
		t.Errorf("AvailableFreight(2,8,2) = %d, want 48", got)
	}
	if got := AvailableFreight(-9, 1, 0, tcs); got != 0 { // floored at zero
		t.Errorf("AvailableFreight negative = %d, want 0", got)
	}
	// A world with no value classes still offers freight at the x1 multiplier.
	if got := AvailableFreight(0, 5, 0, []string{"An"}); got != 5 {
		t.Errorf("AvailableFreight(no value classes) = %d, want 5", got)
	}
}

// TestPassageFareOutOfRange guards the bounds check: an out-of-range Passage must
// return zero, not panic on an array index.
func TestPassageFareOutOfRange(t *testing.T) {
	if got := Passage(3).Fare(0); got != 0 {
		t.Errorf("Passage(3).Fare(0) = %d, want 0", got)
	}
	if got := Passage(-1).Fare(0); got != 0 {
		t.Errorf("Passage(-1).Fare(0) = %d, want 0", got)
	}
}

func TestPassageAttendingSkill(t *testing.T) {
	cases := map[Passage]string{High: "Steward", Mid: "Admin", Low: "Streetwise"}
	for class, want := range cases {
		if got := class.AttendingSkill(); got != want {
			t.Errorf("Passage(%d).AttendingSkill() = %q, want %q", class, got, want)
		}
	}
}

func TestBrokerCommissionAndAvailability(t *testing.T) {
	// Commission = 5% per DM (Book 2 p.220): Broker-4 -> +2 -> 10%; Broker-7 -> +4 -> 20%.
	if got := BrokerCommissionPercent(4); got != 10 {
		t.Errorf("BrokerCommissionPercent(4) = %d, want 10", got)
	}
	if got := BrokerCommissionPercent(7); got != 20 {
		t.Errorf("BrokerCommissionPercent(7) = %d, want 20", got)
	}
	// Availability by starport class.
	if !BrokerAvailable(7, 'A') || BrokerAvailable(7, 'B') {
		t.Errorf("Broker-7 should work only at A")
	}
	if !BrokerAvailable(4, 'C') || BrokerAvailable(4, 'D') {
		t.Errorf("Broker-4 should work at A-C, not D")
	}
	if !BrokerAvailable(1, 'D') || BrokerAvailable(1, 'E') {
		t.Errorf("Broker-1 should work at A-D, not E")
	}
	if BrokerAvailable(0, 'A') {
		t.Errorf("no broker below skill 1")
	}
}

func TestNetSale(t *testing.T) {
	// Book 2 p.221 Beowulf: a Cr7,350 sale via a Broker-4 nets Cr6,615 after the
	// 10% commission.
	if got := NetSale(7350, 4); got != 6615 {
		t.Errorf("NetSale(7350, Broker-4) = %d, want 6615", got)
	}
	if got := NetSale(7800, 0); got != 7800 {
		t.Errorf("NetSale with no broker = %d, want 7800", got)
	}
}
