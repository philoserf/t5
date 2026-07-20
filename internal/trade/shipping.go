package trade

// Passengers, freight, and mail (Book 2 p.220). A merchant ship earns income
// each jump from passages, freight tonnage, speculative cargo (the pricing
// engine), and mail. Availability is rolled per world; the rates and broker
// terms are fixed tables.

// FreightRatePerTon is the per-jump freight rate (Book 2 p.220 Merchant Ship
// Revenues); the mail rate lives with the other mail rules in contracts.go.
const FreightRatePerTon = 1_000 // Cr per ton of freight

// Passage is a class of passenger travel (Book 2 p.220).
type Passage int

// Passenger passage classes.
const (
	High Passage = iota
	Mid
	Low
)

// passageRate is each class's Cr fare at neutral Passage Demand and the Cr its
// fare moves per point of Demand (Book 2 p.220 Merchant Ship Revenues and
// Premium Passage Pricing).
var passageRate = [...]struct{ fare, step int }{
	High: {10_000, 1_000},
	Mid:  {8_000, 1_000},
	Low:  {1_000, 100},
}

// Fare is the price of one passage of this class at the ship's Passage Demand
// (Book 2 p.220 Premium Passage Pricing), floored at zero. An out-of-range class
// has no fare.
func (c Passage) Fare(demand int) int {
	if c < High || int(c) >= len(passageRate) {
		return 0
	}

	rate := passageRate[c]

	return max(rate.fare+demand*rate.step, 0)
}

// AttendingSkill names the skill whose modifier applies to this class's
// availability roll (Book 2 p.220 C). An out-of-range class has no attending
// skill and returns "", the same posture Fare takes with its zero fare: a class
// that does not exist gets no answer rather than a plausible wrong one.
func (c Passage) AttendingSkill() string {
	switch c {
	case High:
		return "Steward"
	case Mid:
		return "Admin"
	case Low:
		return "Streetwise"
	default:
		return ""
	}
}

// AvailablePassengers is the number of passengers of a class offering to board on
// departure day (Book 2 p.220 C): Flux + the world population digit + the
// attending skill's modifier (see Passage.AttendingSkill), floored at zero.
func AvailablePassengers(flux, populationDigit, skillMod int) int {
	return max(flux+populationDigit+skillMod, 0)
}

// AvailableFreight is the tons of freight offered for shipment (Book 2 p.220 D):
// (Flux + population digit + Liaison) times one more than the count of the
// world's value trade classes, floored at zero.
func AvailableFreight(flux, populationDigit, liaison int, worldTCs []string) int {
	return max(flux+populationDigit+liaison, 0) * (len(ValueClasses(worldTCs)) + 1)
}
