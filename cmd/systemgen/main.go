// Command systemgen prints randomly generated Traveller5 star systems.
//
// Usage:
//
//	systemgen [-n count] [-seed value]
//
// With -seed, output is reproducible; without it, each run is freshly seeded.
package main

import (
	"fmt"

	"github.com/philoserf/t5/internal/cli"
	"github.com/philoserf/t5/internal/systemgen"
)

func main() {
	n, r := cli.SeededRoller("systems")
	for i := range n {
		if i > 0 {
			fmt.Println()
		}
		fmt.Println(systemgen.Generate(r))
	}
}
