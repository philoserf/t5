package task

import (
	"github.com/philoserf/t5/internal/dice"
)

// The phantom assets used when a task intentionally omits one half of the
// ordinary Characteristic + Skill pairing (Book 1 pp.123-124).
const (
	PhantomSkill          = 3
	PhantomCharacteristic = 7
)

// JackOfAllTradesSkill returns the skill level JOT substitutes for a missing
// skill: JOT-2, never below zero (Book 1 p.128).
func JackOfAllTradesSkill(level int) int {
	return max(level-2, 0)
}

// AnalogCharacteristic returns the half-value, rounded up, of an analogous
// characteristic substituted for one the sophont does not have (Book 1 p.123).
func AnalogCharacteristic(value int) int {
	return max((value+1)/2, 0)
}

// CooperativeTargetSkill sums every participant's applicable skill with the
// one characteristic selected for a Cooperative task (Book 1 p.125). The task
// statement determines and limits the participating skill values.
func CooperativeTargetSkill(characteristic int, participantSkills ...int) int {
	return characteristic + sum(participantSkills)
}

// CooperativeTargetCharacteristic sums participating characteristics with the
// one skill asset selected for a Cooperative task (Book 1 p.125). This form is
// normally limited to physical characteristics by the referee.
func CooperativeTargetCharacteristic(skill int, participantCharacteristics ...int) int {
	return skill + sum(participantCharacteristics)
}

// OpposedWinner returns the unique participant with the lowest successful raw
// roll. A tie has no winner, as in the p.130 dogfight example. Results should
// already have been resolved through this package so Spectacular overrides are
// reflected in Success.
func OpposedWinner(results []dice.CheckResult) (int, bool) {
	winner, lowest, tied := -1, 0, false

	for i, result := range results {
		if !result.Success {
			continue
		}

		if winner < 0 || result.Roll < lowest {
			winner, lowest, tied = i, result.Roll, false
		} else if result.Roll == lowest {
			tied = true
		}
	}

	return winner, winner >= 0 && !tied
}

// OpposedLoser returns the unique participant with the highest unsuccessful
// raw roll for an elimination round (Book 1 p.126). If everyone succeeds, or
// the highest failing rolls tie, there is no loser and the round repeats.
func OpposedLoser(results []dice.CheckResult) (int, bool) {
	loser, highest, tied := -1, 0, false

	for i, result := range results {
		if result.Success {
			continue
		}

		if loser < 0 || result.Roll > highest {
			loser, highest, tied = i, result.Roll, false
		} else if result.Roll == highest {
			tied = true
		}
	}

	return loser, loser >= 0 && !tied
}

// ResolveArcane attempts an owned Arcane task. A non-owner cannot attempt it,
// consumes no dice, and receives ok=false; ownership does not otherwise change
// ordinary task resolution (Book 1 p.126).
func ResolveArcane(
	r *dice.Roller,
	d Difficulty,
	target int,
	owned bool,
	mods ...int,
) (dice.CheckResult, bool) {
	if !owned {
		return dice.CheckResult{}, false
	}

	return Resolve(r, d, target, mods...), true
}

// ThisIsHardDice applies TIH!: when applicable Skill plus Jack-of-All-Trades is
// below the task's original dice count, difficulty increases by 1D. JOT shields
// against the increase but is not added to the target (Book 1 pp.123, 128).
func ThisIsHardDice(numDice, skill, jackOfAllTrades int) int {
	numDice = max(numDice, 1)
	if skill+jackOfAllTrades < numDice {
		return numDice + 1
	}

	return numDice
}

// CertificationRank returns the successful test's raw roll, where lower is the
// better rank. A failed test confers no certification (Book 1 pp.124, 129).
func CertificationRank(result dice.CheckResult) (int, bool) {
	if !result.Success {
		return 0, false
	}

	return result.Roll, true
}
