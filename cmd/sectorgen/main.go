// Command sectorgen prints a randomly generated Traveller5 subsector map: the
// star systems present in one 8x10 subsector at a chosen stellar density, with
// gas-giant and asteroid-belt flags.
//
// Usage:
//
//	sectorgen [-density name] [-subsector A] [-seed value]
//
// density is one of extragalactic, rift, sparse, scattered, standard, dense,
// cluster, core (default standard). With -seed, output is reproducible.
package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/sectorgen"
)

var densities = map[string]sectorgen.Density{
	"extragalactic": sectorgen.ExtraGalactic,
	"rift":          sectorgen.Rift,
	"sparse":        sectorgen.Sparse,
	"scattered":     sectorgen.Scattered,
	"standard":      sectorgen.Standard,
	"dense":         sectorgen.Dense,
	"cluster":       sectorgen.Cluster,
	"core":          sectorgen.Core,
}

func main() {
	densityName := flag.String("density", "standard", "stellar density (extragalactic…core)")
	subsector := flag.String("subsector", "A", "subsector letter A-P")
	seed := flag.Uint64("seed", 0, "random seed; if omitted, a fresh random seed is used")
	flag.Parse()

	d, ok := densities[strings.ToLower(*densityName)]
	if !ok {
		fmt.Printf("unknown density %q\n", *densityName)
		return
	}
	letter := byte('A')
	if len(*subsector) > 0 {
		letter = strings.ToUpper(*subsector)[0]
	}

	r := dice.New()
	seeded := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "seed" {
			seeded = true
		}
	})
	if seeded {
		r = dice.NewWithSeed(*seed)
	}

	systems := sectorgen.GenerateSubsector(r, d, letter)
	fmt.Printf("Subsector %c — %s density — %d systems\n", letter, d, len(systems))
	for _, s := range systems {
		flags := []string{}
		if s.GasGiant {
			flags = append(flags, "GG")
		}
		if s.AsteroidMainworld {
			flags = append(flags, "Asteroid")
		}
		fmt.Printf("  %s  %s\n", s.Hex, strings.Join(flags, " "))
	}
}
