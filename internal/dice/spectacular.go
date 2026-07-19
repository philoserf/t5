package dice

// A Spectacular classifies a roll's exceptional outcome (Book 1 p. 127): three
// or more ones force a Spectacular Success, three or more sixes a Spectacular
// Failure, and both together (only possible with six or more dice) is
// Spectacularly Interesting.
//
// This package classifies; it does not act. Applying the override to a task's
// outcome belongs to internal/task, which owns Book 1 pp.120-131 — see
// task.Resolve. CheckResult.Success here is the arithmetic result alone.
type Spectacular int

// Spectacular classifications of a check (Book 1 p. 127).
const (
	NotSpectacular           Spectacular = iota
	SpectacularSuccess                   // three or more ones
	SpectacularFailure                   // three or more sixes
	SpectacularlyInteresting             // three ones and three sixes at once
)

// String names the outcome for display.
func (s Spectacular) String() string {
	switch s {
	case SpectacularSuccess:
		return "Spectacular Success"
	case SpectacularFailure:
		return "Spectacular Failure"
	case SpectacularlyInteresting:
		return "Spectacularly Interesting"
	default:
		return "Not Spectacular"
	}
}

// minSpectacularDice is the smallest pool that can hold three of a kind, so the
// smallest that can be spectacular at all (Book 1 p. 127: three ones are "not
// possible on 1D or 2D").
const minSpectacularDice = 3

// Classify inspects the raw dice faces (Book 1 p. 127). Ones and sixes are
// counted across all dice; fewer than three dice can never be spectacular.
func Classify(faces []int) Spectacular {
	if len(faces) < minSpectacularDice {
		return NotSpectacular
	}

	ones, sixes := 0, 0

	for _, f := range faces {
		switch f {
		case 1:
			ones++
		case 6:
			sixes++
		}
	}

	switch {
	case ones >= 3 && sixes >= 3:
		return SpectacularlyInteresting
	case sixes >= 3:
		return SpectacularFailure
	case ones >= 3:
		return SpectacularSuccess
	default:
		return NotSpectacular
	}
}
