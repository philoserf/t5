// Package personals implements Traveller5's social-interaction system, the
// Personals (Book 1 pp. 180-185). A Personal is resolved roll-low: the Purpose
// sets the dice, and Target = Strategy value × Tactic multiplier + the best
// applicable Law + up to two situational Mods.
package personals

import (
	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/task"
)

// Purpose is the objective of a Personal (Book 1 p.182); its difficulty sets the
// dice rolled — Carouse 1D, Query 2D, Persuade 3D, Command 4D.
type Purpose int

const (
	Carouse Purpose = iota
	Query
	Persuade
	Command
)

// Dice is the number of dice the Purpose rolls (Carouse 1 … Command 4).
func (p Purpose) Dice() int { return int(p) + 1 }

func (p Purpose) String() string {
	switch p {
	case Query:
		return "Query"
	case Persuade:
		return "Persuade"
	case Command:
		return "Command"
	default:
		return "Carouse"
	}
}

// Strategy is the approach an Actor takes (Book 1 p.183). A strategy's base
// point value depends on the Purpose it is used for.
type Strategy int

const (
	Casual Strategy = iota
	Enjoyment
	Discussion
	ActiveListening
	AppealTo
	ForceOfWill
	Charming
	Angry
)

// strategyValues[purpose][strategy] is the strategy's base point value 1-5
// (Book 1 p.183 P1 table); a strategy absent from a purpose's map is not
// available for that purpose.
var strategyValues = map[Purpose]map[Strategy]int{
	Carouse:  {Casual: 1, Enjoyment: 2, Discussion: 3, ActiveListening: 4, AppealTo: 5},
	Query:    {Enjoyment: 1, Discussion: 2, ActiveListening: 3, AppealTo: 4, ForceOfWill: 5},
	Persuade: {Discussion: 1, ActiveListening: 2, AppealTo: 3, ForceOfWill: 4, Charming: 5},
	Command:  {ActiveListening: 1, AppealTo: 2, ForceOfWill: 3, Charming: 4, Angry: 5},
}

// StrategyValue returns a strategy's base point value for a purpose and whether
// the strategy is available for that purpose (Book 1 p.183).
func StrategyValue(p Purpose, s Strategy) (int, bool) {
	v, ok := strategyValues[p][s]
	return v, ok
}

// Law is one of the Five Laws of Personal Interaction (Book 1 p.183). The best
// applicable Law supplies a Mod to the Target.
type Law int

const (
	Similarity Law = iota
	Superiority
	Inferiority
	Comfort
	Violence
)

// lawMods[law] is the Law's unconditional Mod per Purpose (Carouse, Query,
// Persuade, Command); 0 means the Law contributes nothing there (Book 1 p.183).
// Inferiority grants only Query +1 unconditionally — its Persuade +2 is
// conditional (see InferiorityAppeal) and so is not in this table.
var lawMods = map[Law][4]int{
	Similarity:  {1, 1, 1, 0},
	Superiority: {0, 1, 2, 3},
	Inferiority: {0, 1, 0, 0},
	Comfort:     {2, 1, 1, 0},
	Violence:    {0, 1, 2, 3},
}

// InferiorityAppeal is the extra Persuade Mod the Inferiority Law grants only
// when the Actor supports it with Begging, Flattery, or Politeness (Book 1 p.183,
// the "+2*" entry). Apply it as a situational Mod when that condition holds;
// LawMod cannot, as it does not see the Tactic.
const InferiorityAppeal = 2

// LawMod returns the Mod the Law contributes for the given Purpose (0 if it does
// not apply, or if the Purpose is out of range).
func LawMod(l Law, p Purpose) int {
	if p < Carouse || p > Command {
		return 0
	}
	return lawMods[l][p]
}

// Situational Mods (Book 1 p.185). Repeat is applied per repeated Strategy or
// Tactic after its first use (required); Brazen is Query/Persuade only; Urgent
// caps the Personal at a single attempt.
const (
	Repeat = -1
	Brazen = 3
	Urgent = 2
)

// Resolve resolves a Personal Interaction (Book 1 p.184): Target = strategyValue
// × tacticMult + lawMod + the situational mods; the Purpose's dice are rolled at
// or under the Target. A tacticMult of 1 means no Tactic multiplier is applied.
func Resolve(
	r *dice.Roller,
	purpose Purpose,
	strategyValue, tacticMult, lawMod int,
	mods ...int,
) dice.CheckResult {
	return task.ResolveDice(r, purpose.Dice(), strategyValue*tacticMult+lawMod, mods...)
}

// Camaraderie counts an Actor's successful Carouses with a Target; each adds +1
// to later Personals with that Target, to a maximum of 6 (Book 1 p.182).
type Camaraderie int

// MaxCamaraderie caps the casual-friendship bonus (Book 1 p.182).
const MaxCamaraderie = 6

// Gain records one successful Carouse, up to the maximum.
func (c *Camaraderie) Gain() {
	if *c < MaxCamaraderie {
		*c++
	}
}

// Mod is the Camaraderie bonus to add as a Personal Mod.
func (c Camaraderie) Mod() int { return int(c) }
