// Command worldgen prints randomly generated Traveller5 mainworld profiles,
// each followed by its trade classifications.
//
// Usage:
//
//	worldgen [-n count] [-seed value] [-format text|json]
//
// With -seed, output is reproducible; without it, each run is freshly seeded.
// -format never changes what is generated or how many dice a world draws —
// only how the same records are rendered, so the same -seed names the same
// worlds under either format.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/philoserf/t5/internal/cli"
	"github.com/philoserf/t5/internal/tradecode"
	"github.com/philoserf/t5/internal/worldgen"
)

// jsonWorld is one -format json record: the same Profile and trade codes the
// text format renders, structured instead of fixed-width — raw ints for
// every UWP characteristic plus the rendered "uwp" convenience string, so a
// consumer needs neither a fixed-width parser nor to re-implement rendering.
//
// This is deliberately a CLI-local shape, not worldgen.World: this command
// only ever builds a bare uwp.Profile (worldgen.Generate), never a full
// World — GenerateWorld needs system context (gas giants, belts, capital
// status) this command doesn't have and would have to fabricate. A World-
// level JSON shape waits for a consumer that actually builds one (systemgen,
// survey).
type jsonWorld struct {
	UWP           string           `json:"uwp"`
	Starport      string           `json:"starport"`
	Size          int              `json:"size"`
	Atmosphere    int              `json:"atmosphere"`
	Hydrographics int              `json:"hydrographics"`
	Population    int              `json:"population"`
	Government    int              `json:"government"`
	Law           int              `json:"law"`
	TechLevel     int              `json:"techLevel"`
	TradeCodes    []tradecode.Code `json:"tradeCodes"`
}

func main() {
	format := flag.String("format", "text", `output format: "text" or "json"`)
	n, r, reportSeed := cli.SeededRoller("worlds")

	// -format is the one flag this command checks itself, so the "already
	// good" reportSeed used to rely on no longer holds — validate before
	// naming a seed for a run that might not generate anything.
	if *format != "text" && *format != "json" {
		cli.Fatalf("unknown -format %q (want text or json)", *format)
	}

	reportSeed()

	out := bufio.NewWriter(os.Stdout)
	enc := json.NewEncoder(out)

	for i := range n {
		p := worldgen.Generate(r)
		tc := worldgen.TradeClassifications(p)

		if *format == "json" {
			// uwpStr is used for both UWP and Starport: p.String() already
			// guards an invalid Starport byte down to '?' rather than the raw
			// (possibly NUL) byte, so reading Starport out of it keeps that
			// guard rather than bypassing it via string(p.Starport) directly.
			uwpStr := p.String()

			err := enc.Encode(jsonWorld{
				UWP:           uwpStr,
				Starport:      uwpStr[:1],
				Size:          p.Size,
				Atmosphere:    p.Atmosphere,
				Hydrographics: p.Hydrographics,
				Population:    p.Population,
				Government:    p.Government,
				Law:           p.Law,
				TechLevel:     p.TechLevel,
				TradeCodes:    tc,
			})
			if err != nil {
				cli.Fatalf("writing world %d: %v", i, err)
			}

			continue
		}

		if len(tc) > 0 {
			fmt.Fprintf(out, "%s  %s\n", p, tradecode.Join(tc, " "))
		} else {
			fmt.Fprintln(out, p)
		}
	}

	if err := out.Flush(); err != nil {
		cli.Fatalf("writing output: %v", err)
	}
}
