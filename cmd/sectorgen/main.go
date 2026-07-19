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
	r, reportSeed := cli.Roller()

	d, ok := sectorgen.DensityByName(*densityName)
	if !ok {
		cli.Fatalf(
			"unknown density %q (known: %s)",
			*densityName,
			strings.Join(sectorgen.DensityNames(), ", "),
		)
	}
	// The view's own argument is resolved before the survey runs, so every flag
	// is known good by the time the seed is named: a seed reported for a run that
	// then dies on a bad hex would name a run that surveyed nothing.
	show, err := selectView(*hex, *sector, *subsector)
	if err != nil {
		cli.Fatalf("%v", err)
	}

	reportSeed()

	show(r, d)
}

// A view prints one selection from a survey of the sector. Each view is chosen
// and its argument validated up front (see selectView), so the survey itself is
// reached only once the whole command line is good.
type view func(*dice.Roller, sectorgen.Density)

// selectView resolves the view flags to the single view to print, checking that
// view's own argument. The views are exclusive and ordered: -hex wins over
// -sector, which wins over the default subsector listing.
func selectView(hex string, sector bool, subsector string) (view, error) {
	switch {
	case hex != "":
		h, ok := sectorgen.ParseHex(hex)
		if !ok {
			return nil, fmt.Errorf(
				"invalid hex %q (want CCRR, columns 1-%d, rows 1-%d, e.g. 0436)",
				hex, sectorgen.Columns, sectorgen.Rows)
		}

		return func(r *dice.Roller, d sectorgen.Density) { printHex(r, d, h) }, nil
	case sector:
		return printSector, nil
	default:
		letter, ok := sectorgen.ParseSubsector(subsector)
		if !ok {
			return nil, fmt.Errorf("invalid subsector %q (want a letter A-P)", subsector)
		}

		return func(r *dice.Roller, d sectorgen.Density) { printSubsector(r, d, letter) }, nil
	}
}

// printSector lists every world in the sector, with its trade routes.
func printSector(r *dice.Roller, d sectorgen.Density) {
	sv := survey.Sector(r, d)
	if len(sv.Records) == 0 {
		cli.Notef("no star systems in this sector at %s density", d)

		return
	}

	fmt.Println(sv)
}

// printHex prints one system's full sheet. The hex is already parsed, so an
// empty hex here is a true result, never a bad flag.
func printHex(r *dice.Roller, d sectorgen.Density, h sectorgen.Hex) {
	rec, found := survey.Sector(r, d).At(h)
	if !found {
		cli.Notef("hex %s holds no star system at %s density (try another hex or seed)", h, d)

		return
	}

	fmt.Println(rec.Sheet())
}

// printSubsector lists one subsector's worlds as Second Survey lines. An empty
// subsector is a true result at a low density, not a failure — it says so on
// stderr, so stdout stays a clean record stream either way.
func printSubsector(r *dice.Roller, d sectorgen.Density, letter byte) {
	records := survey.Sector(r, d).Subsector(letter)
	if len(records) == 0 {
		cli.Notef("no star systems in subsector %c at %s density", letter, d)

		return
	}

	for _, rec := range records {
		fmt.Println(rec.SecondSurvey())
	}
}
