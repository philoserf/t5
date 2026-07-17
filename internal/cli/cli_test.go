package cli

import (
	"errors"
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
