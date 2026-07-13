// Command sectorgen prints a randomly generated Traveller5 subsector map: the
// star systems present in one 8x10 subsector at a chosen stellar density, with
// gas-giant and asteroid-belt flags.
//
// Usage:
//
//	sectorgen [-density name] [-subsector A] [-seed value]
//
// density is one of the names sectorgen.DensityNames reports (default standard);
// with -seed, output is reproducible.
package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/philoserf/t5/internal/cli"
	"github.com/philoserf/t5/internal/sectorgen"
	"github.com/philoserf/t5/internal/survey"
)

func main() {
	densityName := flag.String("density", "standard", "stellar density (extragalactic…core)")
	subsector := flag.String("subsector", "A", "subsector letter A-P")
	detail := flag.Bool("detail", false, "generate a full system per hex and print Second Survey lines")
	r := cli.Roller()

	d, ok := sectorgen.DensityByName(*densityName)
	if !ok {
		fmt.Printf("unknown density %q (known: %s)\n", *densityName, strings.Join(sectorgen.DensityNames(), ", "))
		return
	}
	letter := byte('A')
	if len(*subsector) > 0 {
		letter = strings.ToUpper(*subsector)[0]
	}
	if letter < 'A' || letter > 'P' {
		fmt.Printf("invalid subsector %q (want a letter A-P)\n", *subsector)
		return
	}

	if *detail {
		for _, rec := range survey.Subsector(r, d, letter) {
			fmt.Println(rec.SecondSurvey())
		}
		return
	}

	systems := sectorgen.GenerateSubsector(r, d, letter)
	fmt.Printf("Subsector %c — %s density — %d systems\n", letter, d, len(systems))
	for _, s := range systems {
		var flags []string
		if s.GasGiant {
			flags = append(flags, "GG")
		}
		if s.AsteroidMainworld {
			flags = append(flags, "Asteroid")
		}
		fmt.Printf("  %s  %s\n", s.Hex, strings.Join(flags, " "))
	}
}
