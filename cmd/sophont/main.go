// Command sophont prints randomly generated Traveller5 non-human species.
//
// Usage:
//
//	sophont [-n count] [-seed value] [-char]
//
// Each line is a species record: Genetic Profile, homeworld UWP, evolutionary
// environment, Size, lifespan, gender structure, and caste (if any). With -char
// it also prints one generated individual of each species — its UPP and assigned
// gender and caste. With -seed, output is reproducible; without it, each run is
// freshly seeded.
package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/philoserf/t5/internal/chargen"
	"github.com/philoserf/t5/internal/cli"
	"github.com/philoserf/t5/internal/ehex"
	"github.com/philoserf/t5/internal/sophont"
)

func main() {
	showChar := flag.Bool("char", false, "also print one individual member of each species")
	n, r := cli.SeededRoller("species")

	for range n {
		s := sophont.Generate(r)
		fmt.Println(s)
		if *showChar {
			fmt.Println("  " + memberLine(s, chargen.GenerateSophont(r, s)))
		}
	}
}

// memberLine renders one generated individual: its characteristic profile and
// assigned gender/caste.
func memberLine(s sophont.Species, c chargen.Character) string {
	parts := []string{"UPP " + profile(s, c)}
	if c.Gender != "" {
		parts = append(parts, c.Gender)
	}
	if c.Caste != "" {
		parts = append(parts, c.Caste)
	}
	return strings.Join(parts, " ")
}

// profile is the sophont's characteristic UPP. It is not chargen's eHex UPP: a
// sophont characteristic can exceed the eHex range (an 8D Strength reaches 48),
// so an out-of-range value is shown in brackets as its decimal rather than lost.
// The characteristic identities are carried by the species' Genetic Profile. When
// C6 is the Caste characteristic its slot holds no rolled value, so it is dropped
// from the profile — the caste appears by name instead.
func profile(s sophont.Species, c chargen.Character) string {
	var b strings.Builder
	for ch := chargen.Strength; ch <= chargen.Social; ch++ {
		if ch == chargen.Social && s.Caste != nil {
			break // C6 is the caste, reported by name, not a placeholder digit
		}
		v := c.Score(ch)
		if v >= 0 && v <= ehex.Max {
			b.WriteString(ehex.Format(v))
		} else {
			fmt.Fprintf(&b, "[%d]", v)
		}
	}
	return b.String()
}
