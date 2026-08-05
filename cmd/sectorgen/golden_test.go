package main

import (
	"strings"
	"testing"

	"github.com/philoserf/t5/internal/clitest"
)

// TestGoldenSeed42 pins sectorgen's default subsector listing for a seeded
// run, byte for byte — see cmd/worldgen/golden_test.go for why. It also
// checks the same run against the README's sample block, sharing the one
// subprocess invocation rather than re-running the command a second time. The
// README pipes the command through `head -3`, so only the first three lines
// are shown — comparing the whole stdout against that block would fail on
// line count alone, not content, so the readme check is a prefix comparison
// instead of full equality (the one command among the four goldens where
// that distinction matters). Run `go test ./cmd/sectorgen/... -update` to
// regenerate after a deliberate rules change.
func TestGoldenSeed42(t *testing.T) {
	got := command.Run(t, "-seed", "42")

	t.Run("golden", func(t *testing.T) { got.AssertGolden(t, "seed42") })
	t.Run("readme", func(t *testing.T) {
		want := clitest.ReadmeBlock(t, "../../README.md", "$ go run ./cmd/sectorgen -seed 42 | head -3")

		if !strings.HasPrefix(got.Stdout, want) {
			t.Errorf(
				"README sample is stale.\n--- README shows (head -3) ---\n%s--- actual output's first lines ---\n%.*s",
				want,
				len(want),
				got.Stdout,
			)
		}
	})
}
