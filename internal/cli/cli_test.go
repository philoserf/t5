package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestFatalf checks the contract the commands rely on: bad input goes to stderr,
// never stdout, and exits non-zero. Fatalf calls os.Exit, so it runs in a
// subprocess (the standard Go idiom) rather than killing the test binary.
func TestFatalf(t *testing.T) {
	if os.Getenv("CLI_TEST_FATAL") == "1" {
		Fatalf("unknown density %q", "bogus")

		return // not reached
	}

	cmd := exec.Command(
		os.Args[0],
		"-test.run=TestFatalf",
	)

	cmd.Env = append(os.Environ(), "CLI_TEST_FATAL=1")

	var stdout, stderr strings.Builder

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		t.Fatalf("Fatalf should exit non-zero, got err %v", err)
	}

	if code := exit.ExitCode(); code != FailureCode {
		t.Errorf("exit code = %d, want %d", code, FailureCode)
	}

	if !strings.Contains(stderr.String(), `unknown density "bogus"`) {
		t.Errorf("message should reach stderr, got %q", stderr.String())
	}
	// The whole point: a caller piping stdout gets data or nothing, never an
	// error message masquerading as a record.
	if out := strings.TrimSpace(stdout.String()); out != "" {
		t.Errorf("Fatalf wrote %q to stdout, want nothing", out)
	}
}

// TestRollerReportsFreshSeed checks the reproducibility contract: an unseeded
// run must say which seed it drew, on stderr only, and re-running with that
// seed must reproduce the records byte for byte. Roller reads the real command
// line and the flag package is global, so each run happens in a subprocess.
func TestRollerReportsFreshSeed(t *testing.T) {
	if os.Getenv("CLI_TEST_ROLL") == "1" {
		rollChild()

		return // not reached
	}

	// An unseeded run: records on stdout, the drawn seed on stderr.
	firstOut, firstErr := runRollChild(t)

	if strings.TrimSpace(firstErr) == "" {
		t.Fatal("an unseeded run reported nothing; the drawn seed must reach stderr")
	}

	seed, ok := parseSeed(firstErr)
	if !ok {
		t.Fatalf("no %q line in stderr %q", "seed <n>", firstErr)
	}
	// The seed must not ride along on the record stream a caller pipes.
	if strings.Contains(firstOut, "seed") {
		t.Errorf("seed leaked onto stdout: %q", firstOut)
	}

	// Replaying that seed must reproduce the records exactly.
	secondOut, _ := runRollChild(t, "-seed", seed)
	if secondOut != firstOut {
		t.Errorf("replay with -seed %s gave %q, want %q", seed, secondOut, firstOut)
	}
}

// rollChild is the subprocess half of TestRollerReportsFreshSeed: it builds a
// roller from its own command line and prints a short roll sequence to stdout,
// standing in for a generator's records.
func rollChild() {
	// Drop the test binary's own flags (everything up to the "--" the parent
	// passes); Roller parses what follows, as it would a real command line.
	rollerArgs := flag.Args() // read before the reset discards them
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = append([]string{"rollgen"}, rollerArgs...)

	r := Roller()

	var rolls []string
	for range 10 {
		rolls = append(rolls, fmt.Sprint(r.Dice(2)))
	}

	fmt.Println(strings.Join(rolls, " "))
	os.Exit(0)
}

// runRollChild runs the subprocess with the given roller flags and returns its
// stdout and stderr separately — the split is the thing under test.
func runRollChild(t *testing.T, args ...string) (string, string) {
	t.Helper()

	cmd := exec.Command(os.Args[0], append([]string{"-test.run=TestRollerReportsFreshSeed", "--"}, args...)...)
	cmd.Env = append(os.Environ(), "CLI_TEST_ROLL=1")

	var stdout, stderr strings.Builder

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("child failed: %v (stderr %q)", err, stderr.String())
	}

	return stdout.String(), stderr.String()
}

// parseSeed pulls the seed value out of the reported "…: seed <n>" line.
func parseSeed(stderr string) (string, bool) {
	for line := range strings.SplitSeq(stderr, "\n") {
		_, rest, found := strings.Cut(line, "seed ")
		if !found {
			continue
		}

		if seed := strings.TrimSpace(rest); seed != "" {
			return seed, true
		}
	}

	return "", false
}
