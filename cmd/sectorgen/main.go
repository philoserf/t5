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
	sector := flag.Bool(
		"sector",
		false,
		"survey the whole sector, with trade routes and way stations",
	)
	hex := flag.String("hex", "", "print the full system sheet for one hex, e.g. 0436")
	r := cli.Roller()

	d, ok := sectorgen.DensityByName(*densityName)
	if !ok {
		cli.Fatalf(
			"unknown density %q (known: %s)",
			*densityName,
			strings.Join(sectorgen.DensityNames(), ", "),
		)
	}

	switch {
	case *hex != "":
		printHex(r, d, *hex)
	case *sector:
		printSector(r, d)
	default:
		printSubsector(r, d, *subsector)
	}
}

// printSector lists every world in the sector, with its trade routes.
func printSector(r *dice.Roller, d sectorgen.Density) {
	sv := survey.Sector(r, d)
	if len(sv.Records) == 0 {
		cli.Note("no star systems in this sector at %s density", d)
		return
	}
	fmt.Println(sv)
}

// printHex prints one system's full sheet.
func printHex(r *dice.Roller, d sectorgen.Density, hex string) {
	h, ok := sectorgen.ParseHex(hex)
	if !ok {
		cli.Fatalf("invalid hex %q (want CCRR, columns 1-%d, rows 1-%d, e.g. 0436)",
			hex, sectorgen.Columns, sectorgen.Rows)
	}
	rec, found := survey.Sector(r, d).At(h)
	if !found {
		cli.Note("hex %s holds no star system at %s density (try another hex or seed)", h, d)
		return
	}
	fmt.Println(rec.Sheet())
}

// printSubsector lists one subsector's worlds as Second Survey lines. An empty
// subsector is a true result at a low density, not a failure — it says so on
// stderr, so stdout stays a clean record stream either way.
func printSubsector(r *dice.Roller, d sectorgen.Density, subsector string) {
	letter, ok := sectorgen.ParseSubsector(subsector)
	if !ok {
		cli.Fatalf("invalid subsector %q (want a letter A-P)", subsector)
	}
	records := survey.Sector(r, d).Subsector(letter)
	if len(records) == 0 {
		cli.Note("no star systems in subsector %c at %s density", letter, d)
		return
	}
	for _, rec := range records {
		fmt.Println(rec.SecondSurvey())
	}
}
