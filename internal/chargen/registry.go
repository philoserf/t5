package chargen

import "strings"

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

// CareerByName returns the career with the given name (case-insensitive) and
// whether one matched.
func CareerByName(name string) (Career, bool) {
	for _, c := range allCareers {
		if strings.EqualFold(c.Name, name) {
			return c, true
		}
	}

	return Career{}, false
}

// CareerNames lists every career's name, for a "known careers" message.
func CareerNames() []string {
	names := make([]string, len(allCareers))
	for i, c := range allCareers {
		names[i] = c.Name
	}

	return names
}
