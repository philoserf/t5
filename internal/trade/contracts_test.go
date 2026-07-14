package trade

import "testing"

func TestEstimateActualValue(t *testing.T) {
	// Book 2 p.210: a first die of 6 (no broker) bounds the sale to 100%-170%.
	if lo, hi := EstimateActualValue(6, 0); lo != 100 || hi != 170 {
		t.Errorf("EstimateActualValue(6,0) = %d-%d, want 100-170", lo, hi)
	}
	// A first die of 1 spans the low end: Flux 0 down to -5 -> 100% down to 40%.
	if lo, hi := EstimateActualValue(1, 0); lo != 40 || hi != 100 {
		t.Errorf("EstimateActualValue(1,0) = %d-%d, want 40-100", lo, hi)
	}
	// A Broker-4 (+2) shifts the whole range up: Flux 0..5 -> effective +2..+7 ->
	// 120%..300%.
	if lo, hi := EstimateActualValue(6, 4); lo != 120 || hi != 300 {
		t.Errorf("EstimateActualValue(6,Broker-4) = %d-%d, want 120-300", lo, hi)
	}
}

func TestDeliveryPremiums(t *testing.T) {
	// 10% of base cost per day of accelerated delivery.
	if got := AcceleratedDeliveryPremium(3000, 2); got != 600 {
		t.Errorf("AcceleratedDeliveryPremium(3000,2) = %d, want 600", got)
	}
	if got := AcceleratedDeliveryPremium(3000, 0); got != 0 {
		t.Errorf("AcceleratedDeliveryPremium(_,0) = %d, want 0", got)
	}
	// 10% surcharge for non-standard OTO/STS terms.
	if got := NonStandardTermsSurcharge(1000); got != 100 {
		t.Errorf("NonStandardTermsSurcharge(1000) = %d, want 100", got)
	}
}

func TestMailContractBid(t *testing.T) {
	// Table endpoints (Book 2 p.220), both columns.
	cases := []struct {
		twoD, roundTrips, want int
	}{
		{2, 10, 8_000}, {7, 10, 15_000}, {12, 10, 24_000},
		{2, 5, 4_000}, {7, 5, 15_000}, {12, 5, 30_000},
	}
	for _, c := range cases {
		if got := MailContractBid(c.twoD, c.roundTrips); got != c.want {
			t.Errorf("MailContractBid(%d, %d RT) = %d, want %d", c.twoD, c.roundTrips, got, c.want)
		}
	}
	// Out-of-range 2D clamps into the table.
	if got := MailContractBid(1, 10); got != 8_000 {
		t.Errorf("MailContractBid(1,10) = %d, want 8000 (clamped)", got)
	}
}
