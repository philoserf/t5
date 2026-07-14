// Command sectorgen surveys a randomly generated Traveller5 sector and prints a
// view of it. By default it lists one subsector's worlds as canonical Second
// Survey lines — a full star system generated for every stellar hex.
//
// Usage:
//
//	sectorgen [-density name] [-subsector A] [-sector] [-hex CCRR] [-seed value]
//
// density is one of the names sectorgen.DensityNames reports (default standard);
// with -seed, output is reproducible. Each view is a selection from one survey of
// the whole sector (see survey.Sector), so they agree about what sits in a hex:
// the default lists one subsector, -sector lists the whole sector and its trade
// routes, and -hex prints one system's full sheet — the stellar family, the orbit
// map with every secondary world and moon, and the mainworld's facilities, the
// detail a one-line survey record cannot carry.
package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/philoserf/t5/internal/cli"
	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/sectorgen"
	"github.com/philoserf/t5/internal/survey"
)

func main() {
	densityName := flag.String("density", "standard", "stellar density (extragalactic…core)")
	subsector := flag.String("subsector", "A", "subsector letter A-P")
	sector := flag.Bool("sector", false, "survey the whole sector, with trade routes and way stations")
	hex := flag.String("hex", "", "print the full system sheet for one hex, e.g. 0436")
	r := cli.Roller()

	d, ok := sectorgen.DensityByName(*densityName)
	if !ok {
		fmt.Printf("unknown density %q (known: %s)\n", *densityName, strings.Join(sectorgen.DensityNames(), ", "))
		return
	}

	switch {
	case *hex != "":
		printHex(r, d, *hex)
	case *sector:
		fmt.Println(survey.Sector(r, d))
	default:
		printSubsector(r, d, *subsector)
	}
}

// printHex prints one system's full sheet.
func printHex(r *dice.Roller, d sectorgen.Density, hex string) {
	h, ok := sectorgen.ParseHex(hex)
	if !ok {
		fmt.Printf("invalid hex %q (want CCRR, columns 1-%d, rows 1-%d, e.g. 0436)\n",
			hex, sectorgen.Columns, sectorgen.Rows)
		return
	}
	rec, found := survey.Sector(r, d).At(h)
	if !found {
		fmt.Printf("hex %s holds no star system at %s density (try another hex or seed)\n", h, d)
		return
	}
	fmt.Println(rec.Sheet())
}

// printSubsector lists one subsector's worlds as Second Survey lines.
func printSubsector(r *dice.Roller, d sectorgen.Density, subsector string) {
	letter, ok := sectorgen.ParseSubsector(subsector)
	if !ok {
		fmt.Printf("invalid subsector %q (want a letter A-P)\n", subsector)
		return
	}
	for _, rec := range survey.Sector(r, d).Subsector(letter) {
		fmt.Println(rec.SecondSurvey())
	}
}
