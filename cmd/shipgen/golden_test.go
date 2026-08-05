package main

import "testing"

// murphyReadmeArgs are the exact flags the README's "Sample output" section
// runs: the Murphy-ish design plus a Beam Laser and a Black Globe defense
// (deliberately above the ship's own TL, to also show the problem line).
// Design is pure — no dice — so this needs no -seed to be reproducible.
func murphyReadmeArgs() []string {
	return []string{
		"-hull", "A", "-tl", "12", "-config", "S", "-maneuver", "A", "-jump", "A",
		"-weapon", "beamlaser:T1:orbit", "-defense", "blackglobe",
	}
}

// TestGoldenReadmeDesign pins shipgen's rendered ship card for the README's
// own design flags, byte for byte — see cmd/worldgen/golden_test.go for why.
// It also checks the same run against the README's sample block, sharing the
// one subprocess invocation rather than re-running the command a second
// time; that readme check caught a real instance of drift on first write —
// the README was missing the Defenses/Fuel/problem lines the command
// actually prints. Run `go test ./cmd/shipgen/... -update` to regenerate
// after a deliberate rules change.
func TestGoldenReadmeDesign(t *testing.T) {
	got := command.Run(t, murphyReadmeArgs()...)

	t.Run("golden", func(t *testing.T) { got.AssertGolden(t, "readmeDesign") })
	t.Run("readme", func(t *testing.T) {
		got.AssertReadmeUpToDate(
			t,
			"../../README.md",
			`$ go run ./cmd/shipgen -hull A -tl 12 -config S -maneuver A -jump A -weapon "beamlaser:T1:orbit" -defense blackglobe`,
		)
	})
}
