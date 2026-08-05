package main

import "testing"

// TestGoldenCareerScoutSeed7 pins chargen's rendered character sheet for a
// seeded career run, byte for byte — see cmd/worldgen/golden_test.go for why.
// It also checks the same run against the README's sample block, sharing the
// one subprocess invocation rather than re-running the command a second time.
// Run `go test ./cmd/chargen/... -update` to regenerate after a deliberate
// rules change.
func TestGoldenCareerScoutSeed7(t *testing.T) {
	got := command.Run(t, "-career", "scout", "-seed", "7")

	t.Run("golden", func(t *testing.T) { got.AssertGolden(t, "careerScoutSeed7") })
	t.Run("readme", func(t *testing.T) {
		got.AssertReadmeUpToDate(t, "../../README.md", "$ go run ./cmd/chargen -career scout -seed 7")
	})
}
