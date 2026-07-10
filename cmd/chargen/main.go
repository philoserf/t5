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
	"fmt"

	"github.com/philoserf/t5/internal/chargen"
	"github.com/philoserf/t5/internal/cli"
)

func main() {
	n, r := cli.SeededRoller("characters")
	for range n {
		fmt.Println(chargen.Generate(r))
	}
}
