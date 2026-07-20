// Command worldgen prints randomly generated Traveller5 mainworld profiles,
// each followed by its trade classifications.
//
// Usage:
//
//	worldgen [-n count] [-seed value]
//
// With -seed, output is reproducible; without it, each run is freshly seeded.
package main

import (
	"fmt"

	"github.com/philoserf/t5/internal/cli"
	"github.com/philoserf/t5/internal/tradecode"
	"github.com/philoserf/t5/internal/worldgen"
)

func main() {
	n, r, reportSeed := cli.SeededRoller("worlds")
	reportSeed() // no flags of its own to check, so the input is already good

	for range n {
		p := worldgen.Generate(r)
		if tc := worldgen.TradeClassifications(p); len(tc) > 0 {
			fmt.Printf("%s  %s\n", p, tradecode.Join(tc, " "))
		} else {
			fmt.Println(p)
		}
	}
}
