// Command chargen prints randomly generated Traveller5 character profiles
// (UPP).
//
// Usage:
//
//	chargen [-n count] [-seed value]
//
// With -seed, output is reproducible; without it, each run is freshly seeded.
package main

import (
	"flag"
	"fmt"

	"github.com/philoserf/t5/internal/chargen"
	"github.com/philoserf/t5/internal/dice"
)

func main() {
	n := flag.Int("n", 1, "number of characters to generate")
	seed := flag.Uint64("seed", 0, "random seed; 0 uses a fresh random seed")
	flag.Parse()

	r := dice.New()
	if *seed != 0 {
		r = dice.NewWithSeed(*seed)
	}

	for range *n {
		fmt.Println(chargen.Generate(r))
	}
}
