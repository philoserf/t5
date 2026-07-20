// The tests are an EXTERNAL package (cli_test, not cli) so they can use the very
// harness the commands use: internal/clitest imports cli, so a test inside package
// cli could not import it back. Nothing here needs unexported access — the contract
// under test is entirely exported — so the external package costs nothing and buys
// the one subprocess implementation that six cmd packages already share.
package cli_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/philoserf/t5/internal/cli"
	"github.com/philoserf/t5/internal/clitest"
)

// childKind selects which stand-in command the child runs. clitest.Run hands the
// child the parent's environment, so a test picks its child with t.Setenv and the
// command line stays pure flags — which matters for the roller child, whose whole
// point is that Roller parses a real -seed off argv.
const childKind = "T5_CLI_TEST_CHILD"

const (
	kindFatal = "fatal"
	kindQuiet = "quiet"
	kindRoll  = "roll" // the default, so a missing selector is the ordinary case
)

// command is a stand-in generator: it follows the same contract the six real ones
// do, so it can be driven by the same harness. Its main is a dispatcher because one
// TestMain can intercept for only one Command.
var command = clitest.Command{Name: "rollgen", Main: childMain}

func TestMain(m *testing.M) { command.TestMain(m) }

// childMain is the subprocess half of every test below.
func childMain() {
	switch os.Getenv(childKind) {
	case kindFatal:
		cli.Fatalf("unknown density %q", "bogus")
	case kindQuiet:
		quietChild()
	default:
		rollChild()
	}
}

// rollChild stands in for a well-behaved command: it builds a roller from its own
// command line, reports the seed once its (nonexistent) flags check out, and prints
// a short roll sequence to stdout as a generator would print records.
func rollChild() {
	r, reportSeed := cli.Roller()
	reportSeed()

	rolls := make([]string, 0, 10)

	for range 10 {
		rolls = append(rolls, strconv.Itoa(r.Dice(2)))
	}

	//nolint:forbidigo // this subprocess emulates a command's record stream
	fmt.Println(strings.Join(rolls, " "))
}

// quietChild builds an unseeded roller and rolls with it, but never calls
// reportSeed — standing in for a command that rejects a flag after Roller returns.
// Nothing about a seed may reach either stream.
func quietChild() {
	r, _ := cli.Roller()
	_ = r.Dice(2)
}

// child selects the stand-in and returns the runner for it.
func child(t *testing.T, kind string) clitest.Command {
	t.Helper()
	t.Setenv(childKind, kind)

	return command
}

// TestFatalf checks the contract the commands rely on: bad input goes to stderr,
// never stdout, and exits FailureCode. Fatalf calls os.Exit, so it runs in a
// subprocess rather than killing the test binary — the same subprocess every
// cmd/*/main_test.go uses. AssertRejected is that contract in one call.
func TestFatalf(t *testing.T) {
	got := child(t, kindFatal).Run(t)

	got.AssertRejected(t)

	if !strings.Contains(got.Stderr, `unknown density "bogus"`) {
		t.Errorf("message should reach stderr, got %q", got.Stderr)
	}
}

// TestRollerReportsFreshSeed checks the reproducibility contract: an unseeded run
// must say which seed it drew, on stderr only, and re-running with that seed must
// reproduce the records byte for byte.
//
// It is also what keeps cli.HasSeedReport honest. clitest uses that matcher for a
// negative assertion, which fails open if it drifts off the real wording; here it
// is exercised POSITIVELY against a seed line an actual Roller wrote, so drift
// fails here loudly instead of passing there silently.
func TestRollerReportsFreshSeed(t *testing.T) {
	rollgen := child(t, kindRoll)

	// An unseeded run: records on stdout, the drawn seed on stderr and only there.
	first := rollgen.Run(t)
	first.AssertReportedSeed(t)

	seed, ok := cli.ReportedSeed(first.Stderr)
	if !ok {
		t.Fatalf("no seed line in stderr %q", first.Stderr)
	}

	// Replaying that seed must reproduce the records exactly.
	second := rollgen.Run(t, "-seed", seed)
	if second.Stdout != first.Stdout {
		t.Errorf("replay with -seed %s gave %q, want %q", seed, second.Stdout, first.Stdout)
	}
}

// TestRollerDefersSeedReport is the other half of the contract, and the reason the
// report is a returned function at all (#293): Roller runs before the command has
// checked its own flags, so it must stay silent until the command says its input is
// good. A command that rejects a flag and exits never calls reportSeed, and so
// never names a seed for a run that generated nothing.
func TestRollerDefersSeedReport(t *testing.T) {
	got := child(t, kindQuiet).Run(t)

	if got.Code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr %q)", got.Code, got.Stderr)
	}

	if cli.HasSeedReport(got.Stderr) {
		t.Errorf("Roller named a seed before the caller reported it: %q", got.Stderr)
	}
}

// TestSeedReportMatching pins the two ends of the format cli owns: a line Notef
// actually produces is recognized, and the near-misses are not. The second half is
// the one that protects clitest's negative assertion — flag's own usage text for
// -seed must never read as a seed report.
func TestSeedReportMatching(t *testing.T) {
	cases := map[string]struct {
		stderr string
		want   string // the seed, or "" for no match
	}{
		"reported seed":     {"rollgen: seed 12345\n", "12345"},
		"among other notes": {"rollgen: no systems\nrollgen: seed 7\n", "7"},
		"flag usage text": {
			"  -seed uint\n    \trandom seed; if omitted, a fresh random seed is used\n",
			"",
		},
		"prose mentioning a seed": {"rollgen: could not use seed 4 twice\n", ""},
		"no digits":               {"rollgen: seed\n", ""},
		"empty":                   {"", ""},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, ok := cli.ReportedSeed(tc.stderr)
			if ok != (tc.want != "") {
				t.Fatalf("ReportedSeed(%q) matched = %v, want %v", tc.stderr, ok, tc.want != "")
			}

			if got != tc.want {
				t.Errorf("ReportedSeed(%q) = %q, want %q", tc.stderr, got, tc.want)
			}

			if has := cli.HasSeedReport(tc.stderr); has != ok {
				t.Errorf("HasSeedReport = %v but ReportedSeed matched = %v", has, ok)
			}
		})
	}
}
