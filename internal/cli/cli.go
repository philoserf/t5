// Package cli holds command-line setup shared by the generator commands.
package cli

import (
	"flag"
	"fmt"

	"github.com/philoserf/t5/internal/dice"
)

// SeededRoller defines the shared -n and -seed flags (naming the item in the
// -n help text), parses the command line, and returns the requested count and
// a roller. When -seed is given the roller is seeded — so any value, including
// 0, is reproducible — and otherwise it is freshly random.
func SeededRoller(item string) (n int, r *dice.Roller) {
	count := flag.Int("n", 1, fmt.Sprintf("number of %s to generate", item))
	seed := flag.Uint64("seed", 0, "random seed; if omitted, a fresh random seed is used")
	flag.Parse()

	seeded := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "seed" {
			seeded = true
		}
	})
	if seeded {
		return *count, dice.NewWithSeed(*seed)
	}
	return *count, dice.New()
}
