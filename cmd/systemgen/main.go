// Command systemgen prints randomly generated Traveller5 star systems.
//
// Usage:
//
//	systemgen [-n count] [-seed value]
//
// With -seed, output is reproducible; without it, each run is freshly seeded.
package main

import (
	"flag"
	"fmt"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/systemgen"
)

func main() {
	n := flag.Int("n", 1, "number of systems to generate")
	seed := flag.Uint64("seed", 0, "random seed; 0 uses a fresh random seed")
	flag.Parse()

	r := dice.New()
	if *seed != 0 {
		r = dice.NewWithSeed(*seed)
	}

	for i := range *n {
		if i > 0 {
			fmt.Println()
		}
		fmt.Println(systemgen.Generate(r))
	}
}
