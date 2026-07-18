// Package cli holds command-line setup shared by the generator commands, and the
// convention they follow for reporting: generated records go to stdout, anything
// else goes to stderr (Note, Fatalf), and bad input exits non-zero. A caller can
// then pipe stdout as data — a command that printed "unknown density" onto its
// record stream and exited 0 would be indistinguishable from one that worked.
package cli

import (
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"

	"github.com/philoserf/t5/internal/dice"
)

// FailureCode is the exit status for bad input. It is flag's own usage-error
// code, so a rejected flag value and a rejected flag both exit the same way.
const FailureCode = 2

// Fatalf reports bad input on stderr, prefixed with the command name, and exits
// with FailureCode. Use it for anything the caller must fix; a legitimate empty
// result is not an error (see Note).
func Fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", commandName(), fmt.Sprintf(format, args...))
	os.Exit(FailureCode)
}

// Notef reports on stderr, prefixed with the command name, and returns — for a
// true but empty result ("no star systems at this density"), which is a success
// worth saying out loud rather than printing nothing at all. It stays off stdout
// so it never lands in a piped record stream.
func Notef(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", commandName(), fmt.Sprintf(format, args...))
}

// commandName is the running command's name, as flag reports it in its own usage
// errors ("sectorgen"), so our diagnostics and flag's look alike.
func commandName() string {
	return filepath.Base(os.Args[0])
}

// Roller defines the shared -seed flag, parses the command line, and returns a
// roller. When -seed is given the roller is seeded — so any value, including 0,
// is reproducible — and otherwise a fresh seed is drawn and reported on stderr
// (via Notef), so every run is reproducible after the fact: re-run with the
// printed -seed to get the same records. The report stays off stdout, so a
// piped record stream is unaffected. A command that needs extra flags defines
// them (via the flag package) before calling Roller, which parses the command
// line.
func Roller() *dice.Roller {
	seed := flag.Uint64("seed", 0, "random seed; if omitted, a fresh random seed is used")

	flag.Parse()

	seeded := false

	flag.Visit(func(f *flag.Flag) {
		if f.Name == "seed" {
			seeded = true
		}
	})

	if seeded {
		return dice.NewWithSeed(*seed)
	}

	// Draw the seed here rather than in dice.New, so we can say it out loud —
	// a record nobody can regenerate is the bug this avoids.
	fresh := rand.Uint64()
	Notef("seed %d", fresh)

	return dice.NewWithSeed(fresh)
}

// SeededRoller defines the shared -n and -seed flags (naming the item in the -n
// help text), parses, and returns the requested count and a roller (see Roller
// for the seeding contract). A count below 1 is rejected rather than silently
// generating nothing.
func SeededRoller(item string) (int, *dice.Roller) {
	count := flag.Int("n", 1, fmt.Sprintf("number of %s to generate", item))
	r := Roller() // parses the command line before we read *count

	if *count < 1 {
		Fatalf("-n must be at least 1, got %d", *count)
	}

	return *count, r
}
