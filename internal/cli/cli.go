// Package cli holds command-line setup shared by the generator commands, and the
// convention they follow for reporting: generated records go to stdout, anything
// else goes to stderr (Note, Fatalf), and bad input exits non-zero. A caller can
// then pipe stdout as data — a command that printed "unknown density" onto its
// record stream and exited 0 would be indistinguishable from one that worked.
package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

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
// roller together with reportSeed, the function that names the drawn seed on
// stderr. When -seed is given the roller is seeded — so any value, including 0,
// is reproducible — and otherwise a fresh seed is drawn, so every run is
// reproducible after the fact: re-run with the reported -seed to get the same
// records. The report goes through Notef, staying off stdout, so a piped record
// stream is unaffected. A command that needs extra flags defines them (via the
// flag package) before calling Roller, which parses the command line.
//
// Reporting is a step of its own because Roller runs before the command has
// checked its own input. Roller only parses; everything a command validates
// itself — an unknown density, a bad hull letter — it validates after Roller
// returns. Naming the seed at construction therefore announced a seed and then
// died on bad input, so the seed named a run that generated nothing. The caller
// instead calls reportSeed once its own flags are known good, which is the last
// point before records start, so a run rejected by Fatalf never claims a seed.
//
// reportSeed is idempotent, and says nothing on a seeded run — the caller
// supplied that seed and does not need it read back.
func Roller() (*dice.Roller, func()) {
	seed := flag.Uint64("seed", 0, "random seed; if omitted, a fresh random seed is used")

	flag.Parse()

	seeded := false

	flag.Visit(func(f *flag.Flag) {
		if f.Name == "seed" {
			seeded = true
		}
	})

	if seeded {
		return dice.NewWithSeed(*seed), func() {}
	}

	// dice.New keeps the seed it drew, so it can be said out loud — a record
	// nobody can regenerate is the bug this avoids.
	r := dice.New()
	fresh, _ := r.Seed() // always set: dice.New routes through NewWithSeed
	reported := false

	// Capture the seed, not the roller. The report is a fact about this run
	// decided here; deferring only WHEN it is said, not what.
	//
	// The guard is the idempotence the doc promises, and it is load-bearing:
	// chargen and shipgen each call reportSeed from two places, and today only
	// avoid saying the seed twice because control flow happens to return between
	// them. A command that grows a third path should not have to know that.
	return r, func() {
		if reported {
			return
		}

		reported = true

		Notef("seed %d", fresh)
	}
}

// SeededRoller defines the shared -n and -seed flags (naming the item in the -n
// help text), parses, and returns the requested count, a roller, and the seed
// report (see Roller for the seeding and reporting contract). A count below 1 is
// rejected rather than silently generating nothing — and rejected here, before
// the roller is handed back, so a bad -n never names a seed either.
func SeededRoller(item string) (int, *dice.Roller, func()) {
	count := flag.Int("n", 1, fmt.Sprintf("number of %s to generate", item))
	r, reportSeed := Roller() // parses the command line before we read *count

	if *count < 1 {
		Fatalf("-n must be at least 1, got %d", *count)
	}

	return *count, r, reportSeed
}

// RejectUnusable exits 2 if the command line set any flag the chosen code path
// cannot honor. usable names the flags that path does read; everything else the
// caller typed is reported and rejected. what describes the path in the message,
// and should tell the caller how to get the behaviour they asked for
// ("a random ship; give -hull to design one").
//
// A flag a path silently discards is bad input, not a no-op. Both generator CLIs
// had a path that swallowed one: shipgen's random ship ignored every design flag,
// so "-tl 99" produced a TL-14 ship at exit 0 (#315), and sectorgen's -hex view
// ignored -subsector, so "-hex 0436 -subsector Q" printed a hex at exit 0 while
// the same "-subsector Q" is rejected on the default path. The second is the worse
// shape: input the command's own validator would refuse, waved through because
// another flag outranked it.
//
// It asks flag.Visit — which walks only what was actually set — rather than
// comparing values against defaults, because most defaults are legal values a
// caller may well type. "-tl 0" is a real Tech Level and "-armor 1" is the hull's
// integral layer; a value test would read both as unset, and "-tl 0" alongside a
// random ship is exactly the discarded input this exists to catch.
//
// Call it before reporting the seed, so a run rejected here never names a seed
// for records it did not generate.
func RejectUnusable(what string, usable ...string) {
	var unusable []string

	flag.Visit(func(f *flag.Flag) {
		if !slices.Contains(usable, f.Name) {
			unusable = append(unusable, "-"+f.Name)
		}
	})

	if len(unusable) > 0 {
		Fatalf("%s cannot be combined with %s", strings.Join(unusable, ", "), what)
	}
}
