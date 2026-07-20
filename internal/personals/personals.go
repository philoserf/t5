// Package personals implements Traveller5's social-interaction system, the
// Personals (Book 1 pp. 180-185). A Personal is resolved roll-low: the Purpose
// sets the dice, and Target = Strategy value × Tactic multiplier + the best
// applicable Law + up to two situational Mods (three if the Personal is
// Deliberate). That allowance is the caller's to keep: Resolve takes unbounded
// variadic mods and sums them all.
package personals

import (
	"fmt"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/task"
)

// Purpose is the objective of a Personal (Book 1 p.182); its difficulty sets the
// dice rolled — Carouse 1D, Query 2D, Persuade 3D, Command 4D.
type Purpose int

// Purposes of a Personals interaction.
const (
	Carouse Purpose = iota
	Query
	Persuade
	Command
)

var purposeNames = [...]string{
	Carouse:  "Carouse",
	Query:    "Query",
	Persuade: "Persuade",
	Command:  "Command",
}

// Dice is the number of dice the Purpose rolls (Carouse 1 … Command 4).
//
// Purpose is a closed four-constant enum, so a value outside it is a programming
// error and Dice panics rather than inventing a count. This is the strict half of
// the same split task.Difficulty.Dice takes: a dice count feeds a mechanic and
// has nowhere to put a "?" sentinel, and the unguarded int(p)+1 was quietly
// plausible in both directions — Purpose(-1) yielded 0, which dice.Resolve treats
// as its 2D default, and Purpose(99) yielded 100D. String below is the display
// half, and stays total.
func (p Purpose) Dice() int {
	if !p.valid() {
		panic(fmt.Sprintf("personals: purpose %d out of range %d..%d",
			int(p), int(Carouse), int(Command)))
	}

	return int(p) + 1
}

// String returns the purpose's name, or "?" for a value outside the enum — a
// Stringer must be able to report a bad value rather than name it Carouse.
func (p Purpose) String() string {
	if !p.valid() {
		return "?"
	}

	return purposeNames[p]
}

// valid reports whether p names one of the four Purposes. Written once and
// asked three times — Dice, String and LawMod each need it, and the enum's
// bounds should not be restated at every use.
func (p Purpose) valid() bool { return p >= Carouse && p <= Command }

// Strategy is the approach an Actor takes (Book 1 p.185, STRATEGIES). A strategy's base
// point value depends on the Purpose it is used for.
type Strategy int

// Interaction strategies.
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

// The laws of social attraction.
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
	if !p.valid() {
		return 0
	}

	return lawMods[l][p]
}

// The flat Situational Mods of the Mods for Personals sidebar (Book 1 p.183).
// Bluff and Threat of Violence are not flat and have their own functions below;
// Deliberate is not a Mod at all — it is the permission to apply a third one.
const (
	// VoiceOnly is the Mod for a Personal conducted by communicator, voice only:
	// the Actor forfeits every visual cue ("By Communicator. Voice. -4").
	VoiceOnly = -4
	// VoiceAndVisual is the lighter penalty when the communicator carries picture
	// as well as sound ("By Communicator. Voice + Visual -2").
	VoiceAndVisual = -2
	// Brazen is the bonus for a bold approach, available to Query and Persuade
	// only ("Brazen (Query or Persuade) +3").
	Brazen = 3
	// Urgent presses the Target for an immediate answer and may be applied only
	// once ("Urgent (only once) +2").
	Urgent = 2
	// Repeat is the penalty for reusing a Strategy or a Tactic, applied once per
	// subsequent use ("Subsequent Use of Strategies, per -1" and the identical
	// line for Tactics). It is required, not optional.
	Repeat = -1
)

// Bluff rolls the Flux Mod a bluff contributes ("Bluff (once) Flux", Book 1
// p.183). Being Flux rather than a flat value, a bluff is a gamble: it ranges
// -5..+5, so it can cost the Actor as easily as it pays. A Personal admits at
// most one bluff.
func Bluff(r *dice.Roller) int { return r.Flux() }

// Threat of Violence is a Mod of +Fighter Skill ("Threat of Violence = +Fighter
// Skill", Book 1 p.183) — pass the character's Fighter level (or a subordinate
// Knowledge) straight to Resolve as a mod; there is no helper, because there is
// nothing to compute. The page attaches a consequence this package does not
// model — "failure converts the personal into a fight" — so a caller applying it
// is responsible for opening combat on a failed Resolve.
//
// It is not the Violence entry in lawMods. That is the Law of Violence, whose
// per-Purpose Mod comes through LawMod; this is the p.183 situational Mod. A
// Personal backed by menace may well draw both.

// Resolve resolves a Personal Interaction (Book 1 p.184): Target = strategyValue
// × tacticMult + lawMod + the situational mods; the Purpose's dice are rolled at
// or under the Target. A tacticMult of 1 means no Tactic multiplier is applied.
//
// Resolve sums every Mod it is handed and enforces no ceiling on their number.
// The book allows two situational Mods, or three when the Personal is Deliberate
// — the sidebar's one non-numeric entry, which carries no value of its own and
// so is documented here rather than declared. Choosing which Mods apply, and
// keeping to that allowance, is the caller's job.
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
type Camaraderie int //nolint:recvcheck // deliberate value-reader / pointer-mutator split

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
