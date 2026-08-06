// Package epic builds Traveller5 EPIC adventure checklists and rolls their
// session themes (Book 3 pp.274-279).
//
// An EPIC is deliberately a scaffold rather than a plotted story: four ordered
// acts each contain five scenes that may be played in any order, followed by a
// climax. The referee supplies each scene's concrete cast, stage, purpose, and
// supporting data.
package epic

import "github.com/philoserf/t5/internal/dice"

const (
	actCount      = 4
	scenesPerAct  = 5
	climaxPurpose = "The Final Resolution"
)

// A Scene is one checklist slot. Number is its position within the act. Purpose
// is intentionally empty for ordinary scenes: an EPIC author fills in the need
// the scene resolves, and scenes within an act need not occur in number order.
type Scene struct {
	Number  int
	Purpose string
}

// An Act is one ordered phase of the adventure and its five scene slots.
type Act struct {
	Number  int
	Purpose string
	Scenes  []Scene
}

// A Climax is the freeform final block of scenes after Act 4. Unlike an act,
// it has as many scenes as the authored resolution needs, so a scaffold
// begins with no fixed slots.
type Climax struct {
	Purpose string
	Scenes  []Scene
}

// An Adventure is the canonical EPIC checklist plus the unannounced theme that
// guides the referee. Theme is empty in a bare Scaffold and set by Generate.
type Adventure struct {
	Theme  string
	Acts   []Act
	Climax Climax
}

var actPurposes = [...]string{
	"Who What Why Where When How",
	"Meeting The Key Non-Players",
	"Pursuing The Key Goal",
	"Approaching Resolution",
}

// Scaffold returns a fresh canonical EPIC structure: four acts of five scenes,
// then a climactic final resolution (Book 3 p.275). Each call owns its slices,
// so callers may fill in scenes without changing later scaffolds.
func Scaffold() Adventure {
	acts := make([]Act, actCount)
	for i := range acts {
		scenes := make([]Scene, scenesPerAct)
		for j := range scenes {
			scenes[j].Number = j + 1
		}

		acts[i] = Act{Number: i + 1, Purpose: actPurposes[i], Scenes: scenes}
	}

	return Adventure{
		Acts:   acts,
		Climax: Climax{Purpose: climaxPurpose},
	}
}

// Generate returns a canonical scaffold with a random session theme. It draws
// exactly two dice: the first selects the table column and the second its row.
func Generate(r *dice.Roller) Adventure {
	adventure := Scaffold()
	adventure.Theme, _ = Theme(r.Die(), r.Die())

	return adventure
}

// Theme returns the Typical Mods theme at a 1D-by-1D coordinate (Book 3
// p.279). column and row must each be a die face in 1..6.
func Theme(column, row int) (string, bool) {
	if column < 1 || column > 6 || row < 1 || row > 6 {
		return "", false
	}

	return themes[row-1][column-1], true
}

var themes = [6][6]string{
	{"Justice", "Loyalty", "Awe", "Danger", "Laziness", "Betrayal"},
	{"Happiness", "Cheerful", "Human Frailty", "Paranoia", "Heroism", "Disappointing"},
	{"Kindness", "Trustworthy", "Brave", "Pursuit", "Escape", "Unreliable"},
	{"Honesty", "Admiration", "Bizarre", "Revenge", "Deception", "Stupidity"},
	{"Truthfulness", "Friendly", "Thrifty", "Humiliation", "Conformity", "Confusion"},
	{"Cleanliness", "Novelty", "Profitable", "Improbable", "Extremes", "Chaos"},
}
