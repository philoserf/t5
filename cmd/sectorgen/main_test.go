package main

import (
	"errors"
	"flag"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/philoserf/t5/internal/cli"
)

// TestSelectView: the three views are exclusive and ordered — -hex wins over
// -sector, which wins over the default subsector listing.
func TestSelectView(t *testing.T) {
	cases := map[string]struct {
		hex       string
		sector    bool
		subsector string
	}{
		"hex beats sector":    {hex: "0436", sector: true, subsector: "A"},
		"hex beats subsector": {hex: "0436", subsector: "A"},
		"sector":              {sector: true, subsector: "A"},
		"subsector default":   {subsector: "P"},
	}
	for name, c := range cases {
		if v, err := selectView(c.hex, c.sector, c.subsector); err != nil || v == nil {
			t.Errorf("%s: got view %v, err %v; want a view", name, v, err)
		}
	}
}

// TestSelectViewRejects: a view's own argument is checked when the view is
// selected — before the survey runs, which is what lets the seed report wait
// until the whole command line is known good (#293).
func TestSelectViewRejects(t *testing.T) {
	cases := map[string]struct {
		hex       string
		sector    bool
		subsector string
	}{
		"hex out of range":   {hex: "9999"},
		"hex not four digit": {hex: "43"},
		"hex not numeric":    {hex: "abcd"},
		"subsector Q":        {subsector: "Q"},
		"subsector empty":    {subsector: ""},
		"subsector word":     {subsector: "AA"},
	}
	for name, c := range cases {
		if _, err := selectView(c.hex, c.sector, c.subsector); err == nil {
			t.Errorf("%s: should be rejected, got a view", name)
		}
	}
}

// TestMainSeedFollowsValidation is the end-to-end contract of #293: an unseeded
// run names the seed it drew only once its own flags check out. A rejected flag
// exits 2 with its reason on stderr and no seed line — the seed would name a run
// that surveyed nothing — and no records on stdout. main calls os.Exit, so each
// case runs in a subprocess (the idiom internal/cli's own tests use).
func TestMainSeedFollowsValidation(t *testing.T) {
	if os.Getenv("SECTORGEN_TEST_MAIN") == "1" {
		mainChild()

		return // not reached
	}

	bad := map[string][]string{
		"unknown density":   {"-density", "qqq"},
		"invalid subsector": {"-subsector", "Q"},
		"invalid hex":       {"-hex", "9999"},
	}
	for name, args := range bad {
		t.Run(name, func(t *testing.T) {
			stdout, stderr, code := runMainChild(t, args...)

			if code != cli.FailureCode {
				t.Errorf("exit code = %d, want %d (stderr %q)", code, cli.FailureCode, stderr)
			}

			if strings.Contains(stderr, "seed ") {
				t.Errorf("named a seed for a run that generated nothing: %q", stderr)
			}

			if strings.TrimSpace(stderr) == "" {
				t.Errorf("nothing on stderr; a rejected flag must say why")
			}

			if out := strings.TrimSpace(stdout); out != "" {
				t.Errorf("wrote %q to stdout, want nothing", out)
			}
		})
	}

	// The other half: a good command line still names its drawn seed, on stderr,
	// with the records alone on stdout.
	t.Run("valid run reports its seed", func(t *testing.T) {
		stdout, stderr, code := runMainChild(t, "-subsector", "A")

		if code != 0 {
			t.Errorf("exit code = %d, want 0 (stderr %q)", code, stderr)
		}

		if !strings.Contains(stderr, "seed ") {
			t.Errorf("an unseeded run must name its seed on stderr, got %q", stderr)
		}

		if strings.Contains(stdout, "seed") {
			t.Errorf("seed leaked onto the record stream: %q", stdout)
		}
	})
}

// mainChild is the subprocess half of TestMainSeedFollowsValidation: it rebuilds
// a clean command line from the args after "--" and runs main as sectorgen
// would. It passes no -seed, so every case exercises the drawn-seed path.
func mainChild() {
	args := flag.Args() // read before the reset discards them
	flag.CommandLine = flag.NewFlagSet("sectorgen", flag.ExitOnError)

	os.Args = append([]string{"sectorgen"}, args...)

	main()
	os.Exit(0)
}

// runMainChild runs sectorgen's main in a subprocess with the given flags and
// returns its stdout, stderr, and exit code separately — the split is the thing
// under test.
func runMainChild(t *testing.T, args ...string) (string, string, int) {
	t.Helper()

	cmd := exec.Command(
		os.Args[0],
		append([]string{"-test.run=^TestMainSeedFollowsValidation$", "--"}, args...)...)
	cmd.Env = append(os.Environ(), "SECTORGEN_TEST_MAIN=1")

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
