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
// Run `go test ./cmd/shipgen/... -update` to regenerate after a deliberate
// rules change.
func TestGoldenReadmeDesign(t *testing.T) {
	command.Run(t, murphyReadmeArgs()...).AssertGolden(t, "readmeDesign")
}

// TestReadmeSampleUpToDate fails if the README's shipgen sample ever
// diverges from a fresh run of the exact invocation it shows. It caught a
// real instance of this on first write: the README was missing the
// Defenses/Fuel/problem lines the command actually prints.
func TestReadmeSampleUpToDate(t *testing.T) {
	got := command.Run(t, murphyReadmeArgs()...)
	got.AssertReadmeUpToDate(
		t,
		"../../README.md",
		`$ go run ./cmd/shipgen -hull A -tl 12 -config S -maneuver A -jump A -weapon "beamlaser:T1:orbit" -defense blackglobe`,
	)
}
