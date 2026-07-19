package main

import (
	"errors"
	"flag"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/philoserf/t5/internal/cli"
	"github.com/philoserf/t5/internal/shipgen"
)

// murphyFlags are the Murphy-class Scout's design flags (Book 2 pp.42-43), the
// package's golden ship: 100t Hull-A, TL-12, Lifting Body, Maneuver-A + Jump-A +
// Power-A. `shipgen -hull A -tl 12 -config L` must reach exactly this spec.
func murphyFlags() flags {
	return flags{
		hull: "A", tl: 12, config: "L", structure: "shell", armor: 2,
		maneuver: "A", jump: "A", power: "A", mission: "S",
	}
}

func TestSpecFromFlags(t *testing.T) {
	spec, err := specFromFlags(murphyFlags())
	if err != nil {
		t.Fatalf("the Murphy's own flags were rejected: %v", err)
	}

	if spec.HullLetter != 1 || spec.TL != 12 || spec.ArmorLayers != 2 {
		t.Errorf(
			"hull/TL/armor = %d/%d/%d, want 1/12/2",
			spec.HullLetter,
			spec.TL,
			spec.ArmorLayers,
		)
	}

	if spec.Maneuver == nil || spec.Jump == nil || spec.Power == nil {
		t.Fatalf(
			"all three drives were requested, got %+v/%+v/%+v",
			spec.Maneuver,
			spec.Jump,
			spec.Power,
		)
	}

	if spec.Maneuver.Letter != 1 {
		t.Errorf("Maneuver-A ordinal = %d, want 1", spec.Maneuver.Letter)
	}
	// End to end: the flags a user actually types must design the golden ship.
	if qsp := shipgen.Design(spec).QSP(); qsp != "S-AL22" {
		t.Errorf("QSP = %q, want the Murphy's S-AL22", qsp)
	}
}

// TestSpecFromFlagsOmittedDrives: a blank drive letter means the ship has no such
// drive (nil), not a size-A one — a non-starship has no jump drive at all.
func TestSpecFromFlagsOmittedDrives(t *testing.T) {
	f := murphyFlags()
	f.jump = ""

	spec, err := specFromFlags(f)
	if err != nil {
		t.Fatalf("a blank jump drive is legal, got %v", err)
	}

	if spec.Jump != nil {
		t.Errorf("blank -jump should give no jump drive, got %+v", spec.Jump)
	}

	if spec.Maneuver == nil {
		t.Errorf("the other drives should survive an omitted jump drive")
	}

	// All three are independently omittable — blank is the only "none".
	for name, blank := range map[string]func(*flags){
		"maneuver": func(f *flags) { f.maneuver = "" },
		"jump":     func(f *flags) { f.jump = "" },
		"power":    func(f *flags) { f.power = "" },
	} {
		g := murphyFlags()
		blank(&g)

		if _, err := specFromFlags(g); err != nil {
			t.Errorf("a blank %s drive is legal, got %v", name, err)
		}
	}
}

// TestSpecFromFlagsRejects: every bad value is rejected rather than silently
// coerced. The command turns each of these into a stderr message and exit 2.
func TestSpecFromFlagsRejects(t *testing.T) {
	cases := map[string]func(*flags){
		"empty hull":      func(f *flags) { f.hull = "" },
		"non-letter hull": func(f *flags) { f.hull = "?" },
		"multi-char hull": func(f *flags) { f.hull = "AB" },
		// I and O are not eHex letters: "I" must not slide to some other hull.
		"hull I":            func(f *flags) { f.hull = "I" },
		"hull O":            func(f *flags) { f.hull = "O" },
		"unknown config":    func(f *flags) { f.config = "Q" },
		"unknown structure": func(f *flags) { f.structure = "adamantium" },
		// A bad drive letter is a typo, not a request for no drive: only a blank
		// means "none" (see TestSpecFromFlagsOmittedDrives). Silently dropping the
		// drive would hand back a driveless ship and exit 0.
		"non-letter maneuver": func(f *flags) { f.maneuver = "3" },
		"multi-char jump":     func(f *flags) { f.jump = "AB" },
		"jump I":              func(f *flags) { f.jump = "I" },
		"power O":             func(f *flags) { f.power = "O" },
		"non-letter power":    func(f *flags) { f.power = "?" },
		// Layer 1 is integral to the hull, so the library floors a low count — but
		// a typed 0 or negative is a mistake, not a request for the minimum.
		"zero armor":     func(f *flags) { f.armor = 0 },
		"negative armor": func(f *flags) { f.armor = -3 },
	}
	for name, mutate := range cases {
		f := murphyFlags()
		mutate(&f)

		if _, err := specFromFlags(f); err == nil {
			t.Errorf("%s: should be rejected, got a spec", name)
		}
	}
}

// TestLetterOrdinal: hull and drive letters are the eHex letters used as an
// ordinal 1..24 — A=1 (Hull A = 100t) through Z=24, skipping I and O. This is
// NOT an eHex value (where A=10), which is the easy way to get it wrong.
func TestLetterOrdinal(t *testing.T) {
	cases := map[string]int{
		"A": 1, "B": 2, "H": 8,
		"J": 9,  // I is skipped, so J follows H
		"N": 13, // ...and O is skipped after N
		"P": 14,
		"Z": 24, // the last hull, 2400t
		"a": 1,  // letters are case-insensitive
		"z": 24,
		// I and O are not eHex letters, so they name no hull — they must be
		// rejected outright, not slid onto a neighboring size.
		"I": 0, "O": 0,
		"": 0, "?": 0, "AB": 0,
	}
	for s, want := range cases {
		if got := letterOrdinal(s); got != want {
			t.Errorf("letterOrdinal(%q) = %d, want %d", s, got, want)
		}
	}
}

// TestMainRejectsBadFlags is the end-to-end half of TestSpecFromFlagsRejects: a
// rejected flag must reach the caller as the shared CLI contract says — the
// message on stderr, nothing on stdout (a piped record stream stays clean), and
// exit cli.FailureCode. main calls os.Exit, so each case runs in a subprocess
// (the idiom internal/cli's own tests use).
func TestMainRejectsBadFlags(t *testing.T) {
	if os.Getenv("SHIPGEN_TEST_MAIN") == "1" {
		mainChild()

		return // not reached
	}

	cases := map[string][]string{
		"invalid jump drive":     {"-jump", "I"},
		"multi-char jump drive":  {"-jump", "AB"},
		"numeric maneuver drive": {"-maneuver", "3"},
		"zero armor":             {"-armor", "0"},
		"negative armor":         {"-armor", "-3"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			stdout, stderr, code := runMainChild(t, args...)

			if code != cli.FailureCode {
				t.Errorf("exit code = %d, want %d (stderr %q)", code, cli.FailureCode, stderr)
			}

			if strings.TrimSpace(stderr) == "" {
				t.Errorf("nothing on stderr; a rejected flag must say why")
			}
			// The whole point of the convention: no ship record on stdout.
			if out := strings.TrimSpace(stdout); out != "" {
				t.Errorf("wrote %q to stdout, want nothing", out)
			}
		})
	}
}

// mainChild is the subprocess half of TestMainRejectsBadFlags: it rebuilds a
// clean command line from the args after "--" and runs main as shipgen would.
func mainChild() {
	args := flag.Args() // read before the reset discards them
	flag.CommandLine = flag.NewFlagSet("shipgen", flag.ExitOnError)

	os.Args = append([]string{"shipgen", "-hull", "A", "-seed", "1"}, args...)

	main()
	os.Exit(0)
}

// runMainChild runs shipgen's main in a subprocess with the given extra flags
// and returns its stdout, stderr, and exit code separately.
func runMainChild(t *testing.T, args ...string) (string, string, int) {
	t.Helper()

	cmd := exec.Command(
		os.Args[0],
		append([]string{"-test.run=TestMainRejectsBadFlags", "--"}, args...)...)
	cmd.Env = append(os.Environ(), "SHIPGEN_TEST_MAIN=1")

	var stdout, stderr strings.Builder

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	code := 0

	var exit *exec.ExitError
	if errors.As(err, &exit) {
		code = exit.ExitCode()
	} else if err != nil {
		t.Fatalf("child failed: %v (stderr %q)", err, stderr.String())
	}

	return stdout.String(), stderr.String(), code
}
