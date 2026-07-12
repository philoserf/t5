package chargen

// allCareers indexes every career by its CareerID, so an ID — stored in a
// CareerRecord, or produced by the Agent's Undercover Assignment — resolves to
// its full data. The composite-literal keys are the CareerID constants, so the
// array stays aligned with them by construction.
var allCareers = [...]Career{
	Scout:       ScoutCareer,
	Rogue:       RogueCareer,
	Soldier:     SoldierCareer,
	Marine:      MarineCareer,
	Spacer:      SpacerCareer,
	Agent:       AgentCareer,
	Citizen:     CitizenCareer,
	Entertainer: EntertainerCareer,
	Craftsman:   CraftsmanCareer,
	Scholar:     ScholarCareer,
	Functionary: FunctionaryCareer,
	Noble:       NobleCareer,
	Merchant:    MerchantCareer,
}

// CareerByID returns the career data for a CareerID.
func CareerByID(id CareerID) Career { return allCareers[id] }
