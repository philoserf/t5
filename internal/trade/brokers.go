package trade

// Broker terms (Book 2 pp. 220-221). A broker negotiates a sale, adding a DM on
// the Actual Value Table for a percentage commission, and is only available at
// starports of sufficient class.

// BrokerDM is the Actual Value bonus a broker of the given skill provides: half
// skill rounded up, capped at +4 (Book 2 p.221) and floored at 0. A broker is a
// service one hires, never a penalty: a skill below 1 has no broker at all (see
// BrokerAvailable), so it contributes nothing rather than shifting the sale the
// wrong way. The floor also keeps BrokerCommissionPercent and NetSale honest.
func BrokerDM(brokerSkill int) int {
	return min(max((brokerSkill+1)/2, 0), 4)
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
