package trade

// Delivery terms and mail (Book 2 pp.209, 210, 220).

// MailRatePerTon is the premium rate for a ton of mail (Book 2 p.209, "Each ton
// of mail is shipped at a premium rate of Cr15,000").
//
// Per jump is this package's reading, not the book's words: p.209 states the rate
// without a period, and "Bid is per Jump" is said on p.220 of the CONTRACT table
// only. Freight's "to the starship's next port of call" makes per-jump the natural
// reading for a shipment too, but it is an inference.
// Mail is always incidental — never a major or minor lot — and every lot is at
// least one ton, so this is also the smallest mail payday. On the p.220
// shipment side the payment is a voucher for the same Cr15,000, redeemable at
// any class-A starport.
//
// Carrying mail at all is conditional, and the book states three separate
// conditions in two places:
//
//   - "To be allowed to carry mail, the ship must be armed and the crew must
//     include a gunner" (p.209).
//   - "The ship must have an installed 1-ton Mail Vault" (p.220).
//   - "Destination World must be at least Importance-2 less than current world"
//     (p.220). This one is *directional* — mail flows from the more important
//     world to the less important one — and so is not the same test as the
//     long-term contract's symmetric Importance difference at mailContractBid.
//
// None of the three is enforced here. This package prices mail; it holds no
// ship, no crew, and no pair of worlds to test them against, so a caller
// building a mail workflow must gate on them itself before charging this rate.
// Mail has no availability roll — that is the book's answer, not a gap here. The
// p.220 checklist lists "D.3 Find Mail availability" beside D.1 Freight and D.2
// Cargo, so it reads like a third rollable quantity; but where D.1 prints a formula
// ("Freight = (Flux + Pop) x (total TCs +1)"), D.3 prints only "Possibly 1 ton.
// Requires Mail Vault", and the MAIL SHIPMENTS sidebar hands the question to the
// referee ("The steward can inquire at the starport about availability") with no
// target, difficulty or mod. pp.209 and 220 are the only pages that price or gate mail
// CARRIAGE; the Vault component itself is specified around pp.42-48 and p.89.
// So there is no AvailableMail beside AvailablePassengers and
// AvailableFreight, and inventing one would be our formula wearing a citation.
const MailRatePerTon = 15_000 // Cr per ton

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
// p.220). The ship names a route "between two worlds with an Importance
// difference of at least 2" — symmetric, unlike the per-shipment destination
// rule at MailRatePerTon — and "must have an installed 1-ton Mail Vault"; the
// armed-and-gunner condition of p.209 applies to carrying mail at all. The
// contract goes to a low bid: roll 2D, and if the bid this table returns is
// acceptable to the captain, the ship wins it.
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
//
// Renewal is not modelled. The book adds that "10 Round Trips in a calendar
// year allows negotiating a similar contract (at one level of bid higher) for
// the same route in the next year" — but a bid level is a step along the 2D
// axis, so a renewal is this table read one row up, and that needs a calendar
// and a per-route trip tally that nothing in this package keeps. A caller
// tracking those can renew by calling this with the previous roll plus one.
func MailContractBid(twoD, roundTrips int) int {
	bid := mailContractBid[clamp(twoD, 2, 12)-2]
	if roundTrips >= 10 {
		return bid.tenTrip
	}

	return bid.fiveTrip
}
