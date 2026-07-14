package trade

// Passengers, freight, and mail (Book 2 p.220). A merchant ship earns income
// each jump from passages, freight tonnage, speculative cargo (the pricing
// engine), and mail. Availability is rolled per world; the rates and broker
// terms are fixed tables.

// Per-jump revenue rates (Book 2 p.220 Merchant Ship Revenues).
const (
	FreightRatePerTon = 1_000  // Cr per ton of freight
	MailRatePerTon    = 15_000 // Cr per ton of mail (ship must carry a Mail Vault)
)

// Passage is a class of passenger travel (Book 2 p.220).
type Passage int

const (
	High Passage = iota
	Mid
	Low
)

// baseFare is each passage class's Cr fare at neutral Passage Demand (Book 2
// p.220 Merchant Ship Revenues).
var baseFare = map[Passage]int{High: 10_000, Mid: 8_000, Low: 1_000}

// PassageFare is the fare for one passenger of the class at the ship's Passage
// Demand (Book 2 p.220 Premium Passage Pricing): High and Mid move Cr1,000 per
// point of Demand, Low moves Cr100. Floored at zero.
func PassageFare(class Passage, demand int) int {
	step := 1_000
	if class == Low {
		step = 100
	}
	return max(baseFare[class]+demand*step, 0)
}

// AvailablePassengers is the number of passengers of a class offering to board
// on departure day (Book 2 p.220 C): Flux + the world population digit + the
// attending skill's modifier (Steward for High, Admin for Mid, Streetwise for
// Low), floored at zero.
func AvailablePassengers(flux, populationDigit, skillMod int) int {
	return max(flux+populationDigit+skillMod, 0)
}

// AvailableFreight is the tons of freight offered for shipment (Book 2 p.220 D):
// (Flux + population digit + Liaison) times one more than the count of the
// world's value trade classes, floored at zero. Pass valueTCs as len(ValueClasses(tcs)).
func AvailableFreight(flux, populationDigit, valueTCs, liaison int) int {
	return max(flux+populationDigit+liaison, 0) * (valueTCs + 1)
}

// BrokerCommissionPercent is the cut a broker of the given skill takes from a
// sale (Book 2 p.220 Brokers): 5% per point of Broker DM.
func BrokerCommissionPercent(brokerSkill int) int {
	return BrokerDM(brokerSkill) * 5
}

// BrokerAvailable reports whether a broker of the given skill works at a starport
// of the given class (Book 2 p.220): skill 7+ only at A, 5-6 at A-B, 3-4 at A-C,
// 1-2 at A-D. A skill below 1 has no broker.
func BrokerAvailable(brokerSkill int, starport byte) bool {
	var worst byte
	switch {
	case brokerSkill >= 7:
		worst = 'A'
	case brokerSkill >= 5:
		worst = 'B'
	case brokerSkill >= 3:
		worst = 'C'
	case brokerSkill >= 1:
		worst = 'D'
	default:
		return false
	}
	return starport >= 'A' && starport <= worst
}

// NetSale is a sale's proceeds after the broker's commission (Book 2 pp. 220-221):
// the realized selling price less the broker's percentage cut.
func NetSale(sellingPrice, brokerSkill int) int {
	return sellingPrice - sellingPrice*BrokerCommissionPercent(brokerSkill)/100
}
