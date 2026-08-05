package main

import (
	"bufio"
	"encoding/json"
	"strings"
	"testing"
)

// TestGoldenN3Seed42 pins worldgen's rendered stdout for a seeded run,
// byte for byte, closing the #321 hazard: nothing else in this package would
// notice a reordered field or a shifted dice stream that changed every
// generated UWP without changing the exit code or the seed report. It also
// checks the same run against the README's sample block, sharing the one
// subprocess invocation rather than re-running the command a second time.
// Run `go test ./cmd/worldgen/... -update` to regenerate after a deliberate
// rules change — eyeball the diff before committing it.
func TestGoldenN3Seed42(t *testing.T) {
	got := command.Run(t, "-n", "3", "-seed", "42")

	t.Run("golden", func(t *testing.T) { got.AssertGolden(t, "n3seed42") })
	t.Run("readme", func(t *testing.T) {
		got.AssertReadmeUpToDate(t, "../../README.md", "$ go run ./cmd/worldgen -n 3 -seed 42")
	})
}

// TestGoldenN3Seed42JSON pins -format json's rendered output, byte for byte —
// same reasoning as TestGoldenN3Seed42, extended to the JSON record shape.
func TestGoldenN3Seed42JSON(t *testing.T) {
	command.Run(t, "-n", "3", "-seed", "42", "-format", "json").AssertGolden(t, "n3seed42-json")
}

// TestFormatIsDiceInvariant is the one check that distinguishes "-format
// re-renders the same generated worlds" from "-format generates a different,
// richer record": the same seed must draw the same dice regardless of
// -format, so the UWPs in the JSON records must equal the UWPs in the text
// lines, in the same order. A -format that called a roll-hungrier generator
// path only for JSON would pass every other test here while silently making
// "-seed 42 -format json" describe different worlds than "-seed 42 -format
// text" — exactly the reproducibility break #361 exists to catch structurally.
func TestFormatIsDiceInvariant(t *testing.T) {
	const n, seed = "5", "42"

	text := command.Run(t, "-n", n, "-seed", seed, "-format", "text").Stdout
	jsonOut := command.Run(t, "-n", n, "-seed", seed, "-format", "json").Stdout

	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	textUWPs := make([]string, 0, len(lines))

	for _, line := range lines {
		textUWPs = append(textUWPs, strings.Fields(line)[0])
	}

	jsonUWPs := make([]string, 0, len(lines))

	scanner := bufio.NewScanner(strings.NewReader(jsonOut))
	for scanner.Scan() {
		var rec jsonWorld
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			t.Fatalf("decoding JSON record %q: %v", scanner.Text(), err)
		}

		jsonUWPs = append(jsonUWPs, rec.UWP)
	}

	if len(textUWPs) != len(jsonUWPs) {
		t.Fatalf("text produced %d worlds, json produced %d", len(textUWPs), len(jsonUWPs))
	}

	for i := range textUWPs {
		if textUWPs[i] != jsonUWPs[i] {
			t.Errorf("world %d: text UWP %q != json UWP %q — -format changed what was generated",
				i, textUWPs[i], jsonUWPs[i])
		}
	}
}
