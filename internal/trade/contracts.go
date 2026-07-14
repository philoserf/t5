package trade

// Delivery terms and long-term mail contracts (Book 2 pp.210, 220).

// StandardDeliveryDays is the customary window local merchants have to deliver
// accepted goods to a waiting ship (Book 2 p.210).
const StandardDeliveryDays = 4

// AcceleratedDeliveryPremium is the surcharge to speed delivery (Book 2 p.210):
// 10% of base cost per day earlier than the standard window. A non-positive
// daysEarly costs nothing.
func AcceleratedDeliveryPremium(baseCost, daysEarly int) int {
	if daysEarly <= 0 {
		return 0
	}
	return baseCost * daysEarly / 10
}

// NonStandardTermsSurcharge is the 10% premium for delivering against changed
// terms — OTO delivered to Surface, or STS delivered to Orbit (Book 2 p.210).
func NonStandardTermsSurcharge(amount int) int {
	return amount / 10
}

// mailContractBid holds the long-term mail-contract low bids in Cr per jump for a
// 10-round-trip and a 5-round-trip commitment, indexed by 2D roll - 2 (Book 2
// p.220). A contract runs between two worlds whose Importance differs by at least
// 2, and the ship must carry a 1-ton Mail Vault.
var mailContractBid = [11]struct{ tenTrip, fiveTrip int }{
	{8_000, 4_000},   // 2D = 2
	{10_000, 6_000},  // 3
	{12_000, 8_000},  // 4
	{13_000, 10_000}, // 5
	{14_000, 13_000}, // 6
	{15_000, 15_000}, // 7
	{16_000, 18_000}, // 8
	{18_000, 22_000}, // 9
	{20_000, 24_000}, // 10
	{22_000, 28_000}, // 11
	{24_000, 30_000}, // 12
}

// MailContractBid returns the low bid, in Cr per jump, for a long-term mail
// contract at the given 2D roll (Book 2 p.220). roundTrips of 10 or more uses the
// ten-round-trip column; fewer uses the five-round-trip column. The 2D roll is
// clamped to the table's 2-12 range.
func MailContractBid(twoD, roundTrips int) int {
	bid := mailContractBid[clamp(twoD, 2, 12)-2]
	if roundTrips >= 10 {
		return bid.tenTrip
	}
	return bid.fiveTrip
}
