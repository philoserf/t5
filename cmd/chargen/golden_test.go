package main

import "testing"

// TestGoldenCareerScoutSeed7 pins chargen's rendered character sheet for a
// seeded career run, byte for byte — see cmd/worldgen/golden_test.go for why.
// Run `go test ./cmd/chargen/... -update` to regenerate after a deliberate
// rules change.
func TestGoldenCareerScoutSeed7(t *testing.T) {
	command.Run(t, "-career", "scout", "-seed", "7").AssertGolden(t, "careerScoutSeed7")
}

// TestReadmeSampleUpToDate fails if the README's chargen sample ever diverges
// from a fresh run of the exact invocation it shows.
func TestReadmeSampleUpToDate(t *testing.T) {
	got := command.Run(t, "-career", "scout", "-seed", "7")
	got.AssertReadmeUpToDate(t, "../../README.md", "$ go run ./cmd/chargen -career scout -seed 7")
}
