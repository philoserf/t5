package sophont

// Sophont life stages, transcribed from Book 3 p.231 (Sophont Creation step 09),
// verified against a rendered image of the page. Every sophont passes through
// ten life stages (0 Infant … 9 Retirement); their durations set the species'
// traditional lifespan. Humans spend a 2-year infancy and two 4-year terms in
// each of the nine later stages, for the benchmark 74-year lifespan.

import "github.com/philoserf/t5/internal/dice"

// The ten fixed life stages (Book 3 p.231), by number, are: 0 Infant, 1 Child,
// 2 Teen, 3 Young Adult, 4 Adult, 5 Peak, 6 Mid-Life, 7 Senior, 8 Elder,
// 9 Retirement.

// A LifeCycle is a species' life-stage structure: the number of 4-year terms
// spent in each stage (index = stage number; stage 0 is a fixed 2-year infancy,
// stored as 0 terms) and the resulting traditional lifespan in years.
type LifeCycle struct {
	Terms    [10]int
	Lifespan int
}

// durationByFlux is the uniform part of chart 09A (Flux -4..+5, same across all
// stages 1-9), indexed by Flux+5. The Flux -5 entry is unused — that row is
// irregular and handled separately.
var durationByFlux = [11]int{0, 1, 1, 1, 2, 2, 2, 3, 3, 4, 6}

// lifeStageDuration returns the number of 4-year terms for a stage (1-9) at a
// given Flux (chart 09A). The -5 row is the short-lived special case: only
// stages 1, 5, and 9 last a term; the rest are skipped.
func lifeStageDuration(flux, stage int) int {
	flux = clamp(flux, -5, 5)
	if flux == -5 {
		if stage == 1 || stage == 5 || stage == 9 {
			return 1
		}

		return 0
	}

	return durationByFlux[flux+5]
}

// rollLifeCycle rolls a species' life-stage durations (a Flux per stage 1-9) and
// sums the traditional lifespan: a fixed 2-year infancy plus four years per term.
func rollLifeCycle(r *dice.Roller) LifeCycle {
	var lc LifeCycle

	lc.Lifespan = 2 // stage 0: the automatic half-term infancy

	for stage := 1; stage <= 9; stage++ {
		terms := lifeStageDuration(r.Flux(), stage)
		lc.Terms[stage] = terms
		lc.Lifespan += terms * 4
	}

	return lc
}
