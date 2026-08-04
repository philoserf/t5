package personals

import "github.com/philoserf/t5/internal/dice"

// Tactic is the concrete means used to implement a Strategy (Book 1 p.185).
type Tactic int

// The sixteen tactics in the P1 Personal Interactions grid (Book 1 p.184).
const (
	CommonInterests Tactic = iota
	CommonEnemies
	Logic
	Authority
	Morality
	Culture
	Emotion
	Indebted
	Payment
	Begging
	Politeness
	Flattery
	Referral
	Familiarity
	Insult
	Pain
)

// TacticEffect is one cell of the P1 grid. A compatible blank cell has
// Multiplier 1 and Mod 0; printed x2/x3 cells change Multiplier, while the
// Interests, Enemies, and Pain columns can instead add a Mod. Violent marks
// Insult and Pain, whose failed Personal becomes a fight.
type TacticEffect struct {
	Multiplier   int
	Mod          int
	Violent      bool
	InflictsPain bool
}

type matrixKey struct {
	Purpose  Purpose
	Strategy Strategy
	Tactic   Tactic
}

// TacticEffectFor returns the complete P1 effect and whether the combination is
// allowed. It also rejects a Strategy not available for the Purpose.
func TacticEffectFor(p Purpose, s Strategy, t Tactic) (TacticEffect, bool) {
	if _, ok := StrategyValue(p, s); !ok || t < CommonInterests || t > Pain {
		return TacticEffect{}, false
	}

	key := matrixKey{p, s, t}
	if incompatible[key] {
		return TacticEffect{}, false
	}

	effect := TacticEffect{Multiplier: 1, Violent: t == Insult || t == Pain}
	if value, ok := multipliers[key]; ok {
		effect.Multiplier = value
	}

	if value, ok := tacticMods[key]; ok {
		effect.Mod = value
	}

	if painCells[key] {
		effect.InflictsPain = true
	}

	return effect, true
}

// PersonalResult is a resolved Strategy/Tactic selection. Fight is an explicit
// handoff to the caller's combat orchestration, not combat performed here.
type PersonalResult struct {
	dice.CheckResult

	Fight        bool
	InflictsPain bool
}

// ResolveSelection validates and resolves one complete P1 Strategy/Tactic
// selection. It returns ok=false and consumes no dice for an unavailable pair.
// A failed Insult or Pain tactic sets Fight for the caller to act upon.
func ResolveSelection(
	r *dice.Roller,
	p Purpose,
	s Strategy,
	t Tactic,
	lawMod int,
	mods ...int,
) (PersonalResult, bool) {
	strategy, ok := StrategyValue(p, s)
	if !ok {
		return PersonalResult{}, false
	}

	effect, ok := TacticEffectFor(p, s, t)
	if !ok {
		return PersonalResult{}, false
	}

	result := Resolve(r, p, strategy, effect.Multiplier, lawMod, append([]int{effect.Mod}, mods...)...)

	return PersonalResult{
		CheckResult:  result,
		Fight:        effect.Violent && !result.Success,
		InflictsPain: effect.InflictsPain,
	}, true
}

func keys(p Purpose, s Strategy, tactics ...Tactic) []matrixKey {
	result := make([]matrixKey, len(tactics))
	for i, tactic := range tactics {
		result[i] = matrixKey{p, s, tactic}
	}

	return result
}

type matrixEntry struct {
	keys  []matrixKey
	value int
}

func values(entries ...matrixEntry) map[matrixKey]int {
	result := make(map[matrixKey]int)

	for _, entry := range entries {
		for _, key := range entry.keys {
			result[key] = entry.value
		}
	}

	return result
}

func entry(keys []matrixKey, value int) matrixEntry {
	return matrixEntry{keys, value}
}

var multipliers = values(
	entry(keys(Carouse, Casual, Flattery, Referral, Familiarity), 2),
	entry(keys(Carouse, Enjoyment, Flattery, Referral, Familiarity), 2),
	entry(keys(Carouse, Discussion, Logic, Authority, Morality, Culture, Emotion, Referral, Familiarity), 2),
	entry(keys(Carouse, AppealTo, Logic, Morality, Culture, Emotion, Payment, Begging, Politeness, Flattery), 2),
	entry(keys(Query, Enjoyment, Flattery, Referral, Familiarity), 2),
	entry(keys(Query, Discussion, Logic, Authority, Morality, Culture, Emotion, Referral, Familiarity), 2),
	entry(keys(Query, AppealTo, Logic, Morality, Culture, Emotion, Payment, Begging, Politeness, Flattery), 2),
	entry(keys(Query, ForceOfWill, Logic, Authority, Morality, Culture, Emotion, Insult), 2),
	entry(keys(Persuade, Discussion, Logic, Authority, Morality, Culture, Emotion, Referral, Familiarity), 2),
	entry(keys(Persuade, AppealTo, Logic, Morality, Culture, Emotion, Payment, Begging, Politeness, Flattery), 2),
	entry(keys(Persuade, ForceOfWill, Logic, Authority, Morality, Culture, Emotion, Insult), 2),
	entry(keys(Persuade, Charming, Authority, Morality, Culture, Emotion, Indebted, Begging, Flattery, Referral), 2),
	entry(keys(Persuade, Charming, CommonEnemies), 3),
	entry(keys(Command, AppealTo, Logic, Morality, Culture, Emotion, Payment, Begging, Politeness, Flattery), 2),
	entry(keys(Command, ForceOfWill, Logic, Authority, Morality, Culture, Emotion, Insult), 2),
	entry(keys(Command, Charming, Authority, Morality, Culture, Emotion, Indebted, Begging, Flattery, Referral), 2),
	entry(keys(Command, Charming, CommonEnemies), 3),
	entry(keys(Command, Angry, Logic, Authority, Morality, Culture, Emotion, Indebted, Insult), 2),
	entry(keys(Command, Angry, CommonEnemies), 3),
)

var tacticMods = values(
	entry(keys(Carouse, Casual, CommonInterests), 3),
	entry(keys(Carouse, Casual, CommonEnemies), 2),
	entry(keys(Query, Enjoyment, CommonInterests), 2),
	entry(keys(Query, Enjoyment, CommonEnemies), 1),
	entry(keys(Persuade, Discussion, CommonInterests), 1),
	entry(keys(Carouse, Casual, Pain), -8),
	entry(keys(Carouse, Enjoyment, Pain), -6),
	entry(keys(Carouse, Discussion, Pain), -6),
	entry(keys(Carouse, ActiveListening, Pain), -6),
	entry(keys(Carouse, AppealTo, Pain), 4),
	entry(keys(Query, Enjoyment, Pain), -6),
	entry(keys(Query, Discussion, Pain), -6),
	entry(keys(Query, ActiveListening, Pain), -6),
	entry(keys(Query, AppealTo, Pain), 4),
	entry(keys(Query, ForceOfWill, Pain), 6),
	entry(keys(Persuade, Discussion, Pain), -6),
	entry(keys(Persuade, ActiveListening, Pain), -6),
	entry(keys(Persuade, AppealTo, Pain), 4),
	entry(keys(Persuade, ForceOfWill, Pain), 6),
	entry(keys(Persuade, Charming, Pain), -4),
	entry(keys(Command, ActiveListening, Pain), -6),
	entry(keys(Command, AppealTo, Pain), 4),
	entry(keys(Command, ForceOfWill, Pain), 6),
	entry(keys(Command, Charming, Pain), -4),
	entry(keys(Command, Angry, Pain), 6),
)

var incompatible = func() map[matrixKey]bool {
	result := make(map[matrixKey]bool)

	rows := []struct {
		p Purpose
		s Strategy
		t []Tactic
	}{
		{Carouse, Casual, []Tactic{Authority, Payment, Begging, Insult}},
		{Carouse, Enjoyment, []Tactic{Begging, Insult}},
		{Carouse, Discussion, []Tactic{Insult}},
		{Query, Enjoyment, []Tactic{Begging, Insult}},
		{Query, Discussion, []Tactic{Insult}},
		{Query, ForceOfWill, []Tactic{Begging}},
		{Persuade, Discussion, []Tactic{Insult}},
		{Persuade, ForceOfWill, []Tactic{Begging}},
		{Command, ForceOfWill, []Tactic{Begging}},
	}
	for _, row := range rows {
		for _, tactic := range row.t {
			result[matrixKey{row.p, row.s, tactic}] = true
		}
	}

	return result
}()

var painCells = func() map[matrixKey]bool {
	result := make(map[matrixKey]bool)
	for _, key := range []matrixKey{
		{Carouse, AppealTo, Pain},
		{Query, AppealTo, Pain},
		{Query, ForceOfWill, Pain},
		{Persuade, AppealTo, Pain},
		{Persuade, ForceOfWill, Pain},
		{Persuade, Charming, Pain},
		{Command, AppealTo, Pain},
		{Command, ForceOfWill, Pain},
		{Command, Charming, Pain},
		{Command, Angry, Pain},
	} {
		result[key] = true
	}

	return result
}()
