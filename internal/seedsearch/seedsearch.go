// Package seedsearch finds a real dice.Roller seed producing a rare or
// otherwise hard-to-hand-construct generated outcome, for tests whose fixture
// needs an actual result from the top-level generation path — where the path
// length or shape varies enough (a career of unknown length, a variable term
// count) that there is no fixed dice.NewScripted sequence to write instead.
//
// A fixture that needs such an outcome used to hard-code a seed found by
// hand, each pinned with a comment like "confirmed by direct inspection".
// That holds only until something shifts the dice stream a generator draws
// from — a rule fix, a reordered roll — at which point the seed produces an
// ordinary result and the test fails on an unrelated assumption instead of
// asserting anything about the behavior it exists to exercise. Searching at
// run time costs a handful of cheap generations and never needs re-deriving.
package seedsearch

import "testing"

// Find returns the first seed in 1..limit for which matches reports true, and
// fails t if none does.
//
// matches's ONLY job is testing the precondition the fixture needs — the rare
// or hard-to-construct situation under test — never the behavior the fixture
// goes on to assert. Folding the assertion itself into matches makes the
// search find a seed that already passes and the test then merely confirm it
// again: vacuous. Keep the two separate even when it is tempting to combine
// them for one term of a compound condition.
//
// Calibrate limit from the rarest outcome the specific caller needs, measured
// (e.g. by a throwaway frequency probe), not guessed: a bound picked too
// small fails intermittently as a generator's dice consumption shifts over
// time even when nothing about the mechanic itself is wrong.
func Find(t *testing.T, limit uint64, what string, matches func(seed uint64) bool) uint64 {
	t.Helper()

	for seed := uint64(1); seed <= limit; seed++ {
		if matches(seed) {
			return seed
		}
	}

	t.Fatalf("no seed in 1..%d produced %s", limit, what)

	return 0
}
