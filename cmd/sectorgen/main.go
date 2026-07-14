// Command sectorgen prints a randomly generated Traveller5 subsector map: the
// star systems present in one 8x10 subsector at a chosen stellar density, with
// gas-giant and asteroid-belt flags.
//
// Usage:
//
//	sectorgen [-density name] [-subsector A] [-detail] [-sector] [-hex CCRR] [-seed value]
//
// density is one of the names sectorgen.DensityNames reports (default standard);
// with -seed, output is reproducible. -detail prints Second Survey lines for the
// subsector; -sector surveys the full sector with trade routes and way stations.
// -hex prints the full system sheet for one hex — the stellar family, the orbit
// map with every secondary world and moon, and the mainworld's facilities — the
// detail the one-line survey record cannot carry.
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
	sector := flag.Bool("sector", false, "survey the full sector with trade routes and way stations")
	hex := flag.String("hex", "", "print the full system sheet for one hex, e.g. 0436 (surveys the sector)")
	r := cli.Roller()

	d, ok := sectorgen.DensityByName(*densityName)
	if !ok {
		fmt.Printf("unknown density %q (known: %s)\n", *densityName, strings.Join(sectorgen.DensityNames(), ", "))
		return
	}

	// A hex drills into one system of the surveyed sector, so the sector must be
	// generated the same way -sector generates it.
	if *hex != "" {
		sv := survey.Sector(r, d)
		rec, found := sv.At(*hex)
		if !found {
			fmt.Printf("hex %s holds no star system at %s density (try another hex or seed)\n", *hex, d)
			return
		}
		fmt.Println(rec.Sheet())
		return
	}

	if *sector {
		fmt.Println(survey.Sector(r, d))
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
