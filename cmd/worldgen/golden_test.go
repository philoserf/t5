package main

import "testing"

// TestGoldenN3Seed42 pins worldgen's rendered stdout for a seeded run,
// byte for byte, closing the #321 hazard: nothing else in this package would
// notice a reordered field or a shifted dice stream that changed every
// generated UWP without changing the exit code or the seed report. Run
// `go test ./cmd/worldgen/... -update` to regenerate after a deliberate
// rules change — eyeball the diff before committing it.
func TestGoldenN3Seed42(t *testing.T) {
	command.Run(t, "-n", "3", "-seed", "42").AssertGolden(t, "n3seed42")
}

// TestReadmeSampleUpToDate fails if the README's "Sample output" block for
// this command ever diverges from a fresh run of the exact invocation it
// shows.
func TestReadmeSampleUpToDate(t *testing.T) {
	got := command.Run(t, "-n", "3", "-seed", "42")
	got.AssertReadmeUpToDate(t, "../../README.md", "$ go run ./cmd/worldgen -n 3 -seed 42")
}
