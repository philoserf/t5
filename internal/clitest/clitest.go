// Package clitest is the end-to-end harness the generator commands share for
// testing the CLI contract internal/cli defines: generated records go to stdout,
// everything else to stderr, and bad input exits cli.FailureCode without naming a
// seed for records it never generated.
//
// A command's main calls os.Exit, so it cannot be exercised in the test process:
// each case re-executes the test binary as a child that runs main instead of the
// tests. Every command needs the same child, so it lives here once rather than
// being copied into six main_test.go files.
//
// A command wires it up in two lines:
//
//	var shipgen = clitest.Command{Name: "shipgen", Main: main}
//
//	func TestMain(m *testing.M) { shipgen.TestMain(m) }
//
// and then runs it: shipgen.Run(t, "-tl", "-5").AssertRejected(t).
package clitest

import (
	"errors"
	"flag"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/philoserf/t5/internal/cli"
)

// childEnv marks a process as the child. It is only a marker: the command line
// rides argv, because the interception below happens before m.Run and so before
// anything has looked at os.Args. Passing the args out of band through the
// environment would need an encoding — a separator that cannot occur inside a real
// flag value, which rules out space and comma (several commands take list-valued
// flags like "-weapon beamlaser:T1:orbit,sandcaster") and NUL (os/exec refuses it).
// Using argv needs none of that.
const childEnv = "T5_CLITEST_CHILD"

// Command binds a generator command's name and main for end-to-end testing.
type Command struct {
	// Name is the command as it is invoked ("shipgen"). It becomes os.Args[0] in
	// the child, which is where cli's diagnostics and flag's own usage errors
	// both read the command name from.
	Name string
	// Main is the command's main function.
	Main func()
}

// TestMain runs the command instead of the package's tests when this process is
// the child. Call it from the test package's TestMain.
//
// The interception happens before m.Run, so the child never enters the testing
// framework at all: no test flags are parsed, no test is selected, and the child
// cannot be perturbed by which test happened to spawn it. That is what lets Run
// stay a plain function call rather than a test that re-enters itself.
func (c Command) TestMain(m *testing.M) {
	if os.Getenv(childEnv) == "1" {
		c.runChild(os.Args[1:]) // never returns
	}

	os.Exit(m.Run())
}

// Run executes the command in a subprocess with the given command line and
// returns its streams and exit code. No -seed is passed unless the caller passes
// one, so a run exercises the drawn-seed path by default.
func (c Command) Run(t *testing.T, args ...string) Result {
	t.Helper()

	// The subprocess is this very test binary, re-executed so that TestMain can
	// hand it to the command's main instead of the tests.
	// t.Context is cancelled when the test ends, so a child that hangs cannot
	// outlive the run that spawned it.
	cmd := exec.CommandContext( //nolint:gosec // os.Args[0] is the test binary itself
		t.Context(),
		os.Args[0],
		args...,
	)

	cmd.Env = append(os.Environ(), childEnv+"=1")

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

	return Result{Stdout: stdout.String(), Stderr: stderr.String(), Code: code}
}

// runChild runs main under the command line the parent asked for.
func (c Command) runChild(args []string) {
	// The test binary's flags are already registered on the default FlagSet, and
	// the command is about to register its own. Start it from a clean one, named
	// as the command is, so a flag error reports the command's usage.
	flag.CommandLine = flag.NewFlagSet(c.Name, flag.ExitOnError)

	os.Args = append([]string{c.Name}, args...)

	c.Main()
	os.Exit(0) // main returned, so the run succeeded
}

// Result is one child run: the two streams kept apart (the split is the thing
// under test) and the exit code.
type Result struct {
	Stdout, Stderr string
	Code           int
}

// AssertRejected checks the whole bad-input contract: exit cli.FailureCode, a
// reason on stderr, nothing on stdout, and — the obligation nothing else
// enforces — no seed named for a run that generated nothing.
func (r Result) AssertRejected(t *testing.T) {
	t.Helper()

	if r.Code != cli.FailureCode {
		t.Errorf("exit code = %d, want %d (stderr %q)", r.Code, cli.FailureCode, r.Stderr)
	}

	if strings.TrimSpace(r.Stderr) == "" {
		t.Error("nothing on stderr; rejected input must say why")
	}

	if out := strings.TrimSpace(r.Stdout); out != "" {
		t.Errorf("wrote %q to stdout, want nothing on the record stream", out)
	}

	// cli owns what a seed line looks like, and this is the direction that fails
	// open: a matcher that drifted off the real wording would stop matching and
	// wave a leaked seed through. Asking cli means there is no second wording to
	// drift from.
	if cli.HasSeedReport(r.Stderr) {
		t.Errorf("named a seed for a run that generated nothing: %q", r.Stderr)
	}

	// An unrecovered panic ALSO exits 2, writes to stderr, writes nothing to
	// stdout, and names no seed — so it satisfies every clause above and a crash
	// on a bad-input path would read as a clean rejection. That is the fail-open
	// shape this repo has now found three times; here it would hide the very
	// crashes these tables exist to catch.
	if strings.Contains(r.Stderr, "panic:") {
		t.Errorf("child panicked rather than rejecting the input: %q", r.Stderr)
	}
}

// AssertReportedSeed checks the good-run contract: exit 0, records on stdout, and
// the drawn seed named on stderr and only there. A dropped reportSeed() call
// fails here — which is the whole reason every command has one of these.
func (r Result) AssertReportedSeed(t *testing.T) {
	t.Helper()

	if r.Code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr %q)", r.Code, r.Stderr)
	}

	if !cli.HasSeedReport(r.Stderr) {
		t.Errorf("an unseeded run must name its drawn seed on stderr, got %q", r.Stderr)
	}

	if strings.TrimSpace(r.Stdout) == "" {
		t.Error("nothing on stdout; a successful run must print its records there")
	}

	// Two checks, because they fail in opposite directions and neither subsumes
	// the other. cli.HasSeedReport is the canonical Notef line, routed through the
	// owner of the wording so a reword cannot leave a stale matcher behind. The
	// broad substring catches a seed reaching stdout by any OTHER path — a record
	// header like `fmt.Println("# seed", n)`, or a sheet field — which the anchored
	// form cannot see.
	//
	// Replacing the substring with the anchored matcher alone was a mistake made in
	// this very wave: it closed a drift hole and opened a coverage hole, in the one
	// assertion whose whole job is catching a leak onto the record stream.
	if cli.HasSeedReport(r.Stdout) || strings.Contains(strings.ToLower(r.Stdout), "seed") {
		t.Errorf("the seed leaked onto the record stream: %q", r.Stdout)
	}
}
