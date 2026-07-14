// Command sectorgen surveys a randomly generated Traveller5 sector. By default it
// prints one subsector of it as canonical Second Survey lines — a full star system
// generated for every stellar hex.
//
// Usage:
//
//	sectorgen [-density name] [-subsector A] [-sector] [-hex CCRR] [-seed value]
//
// density is one of the names sectorgen.DensityNames reports (default standard);
// with -seed, output is reproducible. Every view surveys the same sector and then
// shows part of it, so they agree with each other: the default lists one
// subsector's worlds, -sector lists them all and adds the trade routes, and -hex
// prints one system's full sheet — the stellar family, the orbit map with every
// secondary world and moon, and the mainworld's facilities — the detail the
// one-line survey record cannot carry. (A sector's systems share one dice stream,
// so a hex's world depends on every hex before it; surveying the whole sector each
// time is what keeps the views consistent. It costs milliseconds.)
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
	sector := flag.Bool("sector", false, "survey the whole sector, with trade routes and way stations")
	hex := flag.String("hex", "", "print the full system sheet for one hex, e.g. 0436 (surveys the sector)")
	r := cli.Roller()

	d, ok := sectorgen.DensityByName(*densityName)
	if !ok {
		fmt.Printf("unknown density %q (known: %s)\n", *densityName, strings.Join(sectorgen.DensityNames(), ", "))
		return
	}

	// Validate the view's argument before surveying, so a typo costs nothing.
	var target sectorgen.Hex
	letter := byte(0)
	switch {
	case *hex != "":
		h, ok := sectorgen.ParseHex(*hex)
		if !ok {
			fmt.Printf("invalid hex %q (want CCRR, columns 1-%d, rows 1-%d, e.g. 0436)\n",
				*hex, sectorgen.Columns, sectorgen.Rows)
			return
		}
		target = h
	case !*sector:
		if *subsector == "" {
			fmt.Println("invalid subsector \"\" (want a letter A-P)")
			return
		}
		letter = strings.ToUpper(*subsector)[0]
		if letter < 'A' || letter > 'P' {
			fmt.Printf("invalid subsector %q (want a letter A-P)\n", *subsector)
			return
		}
	}

	// Every view is a slice of the same surveyed sector, so they agree on what sits
	// in any given hex — and each world already carries the capitals, trade routes,
	// and way stations that only a whole-region pass can assign.
	sv := survey.Sector(r, d)

	switch {
	case *hex != "":
		rec, found := sv.At(target)
		if !found {
			fmt.Printf("hex %s holds no star system at %s density (try another hex or seed)\n", target, d)
			return
		}
		fmt.Println(rec.Sheet())
	case *sector:
		fmt.Println(sv)
	default:
		for _, rec := range sv.Records {
			if rec.Hex.Subsector() == letter {
				fmt.Println(rec.SecondSurvey())
			}
		}
	}
}
