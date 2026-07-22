package task

import "github.com/philoserf/t5/internal/dice"

const assumedUncertainFace = 3

// An UncertainResult separates what the player can infer from the result the
// referee secretly administers (Book 1 pp. 126-127).
type UncertainResult struct {
	Actual             dice.CheckResult
	ApparentRoll       int
	ApparentSuccess    bool
	VisibleSpectacular dice.Spectacular
	PlayerFaces        []int
	RefereeFaces       []int
}

// ResolveUncertainDice resolves a task whose uncertainDice are rolled secretly
// by the referee. The player rolls the remaining dice first. The announced
// result substitutes 3 for each hidden die, while Actual uses every real face.
// Visible spectaculars are imposed on the announced result; spectaculars that
// involve hidden faces remain visible only in Actual for secret administration.
func ResolveUncertainDice(
	r *dice.Roller,
	numDice, uncertainDice, target int,
	mods ...int,
) UncertainResult {
	numDice = max(numDice, 1)
	if uncertainDice < 0 || uncertainDice > numDice {
		panic("task: uncertainty must be between zero and the task dice count")
	}

	playerFaces := r.DiceFaces(numDice - uncertainDice)
	refereeFaces := r.DiceFaces(uncertainDice)
	actualFaces := append(append([]int(nil), playerFaces...), refereeFaces...)
	adjustedTarget := target + sum(mods)
	actual := resultFromFaces(actualFaces, adjustedTarget)
	actual = applySpectacular(actual)

	visible := dice.Classify(playerFaces)
	apparentRoll := faceSum(playerFaces) + assumedUncertainFace*uncertainDice
	apparentSuccess := apparentRoll <= adjustedTarget

	switch visible {
	case dice.SpectacularSuccess:
		apparentSuccess = true
	case dice.SpectacularFailure:
		apparentSuccess = false
	case dice.NotSpectacular, dice.SpectacularlyInteresting:
	}

	return UncertainResult{
		Actual:             actual,
		ApparentRoll:       apparentRoll,
		ApparentSuccess:    apparentSuccess,
		VisibleSpectacular: visible,
		PlayerFaces:        playerFaces,
		RefereeFaces:       refereeFaces,
	}
}

// UncertaintyDice applies pace to the hidden portion of an Uncertain task. Each
// harder level (Hasty +1, Extra Hasty +2) adds that many hidden dice; Cautious
// does not reduce uncertainty. The result cannot exceed the paced task's dice.
func UncertaintyDice(base, harderLevels, pacedTaskDice int) int {
	return min(max(base+max(harderLevels, 0), 0), max(pacedTaskDice, 1))
}

func resultFromFaces(faces []int, target int) dice.CheckResult {
	roll := faceSum(faces)

	return dice.CheckResult{
		Roll: roll, Total: roll, Target: target,
		Success: roll <= target, Effect: target - roll,
		Faces: append([]int(nil), faces...),
	}
}

func faceSum(faces []int) int {
	total := 0
	for _, face := range faces {
		total += face
	}

	return total
}
