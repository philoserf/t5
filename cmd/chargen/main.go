// Command chargen prints randomly generated Traveller5 characters.
//
// Usage:
//
//	chargen [-n count] [-seed value] [-career name]
//
// Without -career it prints a bare UPP. With -career (e.g. "scout") it runs the
// named career — qualification, terms, skills, aging, and mustering out — using
// the default policy, and prints the resulting character sheet. With -seed,
// output is reproducible; without it, each run is freshly seeded.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/philoserf/t5/internal/chargen"
	"github.com/philoserf/t5/internal/cli"
)

func main() {
	careerName := flag.String("career", "", `career to run (e.g. "scout"); empty prints UPP only`)
	n, r := cli.SeededRoller("characters")

	if *careerName == "" {
		for range n {
			fmt.Println(chargen.Generate(r))
		}
		return
	}

	career, err := careerByName(*careerName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "chargen:", err)
		os.Exit(2)
	}
	for range n {
		fmt.Println(render(chargen.GenerateCareered(r, chargen.DefaultPolicy{}, career), career))
	}
}

// careerByName resolves a -career flag value to its career data.
func careerByName(name string) (chargen.Career, error) {
	switch strings.ToLower(name) {
	case "scout":
		return chargen.ScoutCareer, nil
	default:
		return chargen.Career{}, fmt.Errorf("unknown career %q (known: scout)", name)
	}
}

// render formats a careered character as a one-line sheet.
func render(c chargen.Character, career chargen.Career) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s  age %d", c.UPP(), c.Age)
	if len(c.Careers) == 0 {
		fmt.Fprintf(&b, "  did not qualify for %s", career.Name)
		return b.String()
	}
	rec := c.Careers[len(c.Careers)-1]
	fmt.Fprintf(&b, "  %s: %d terms, %s", career.Name, rec.Terms, rec.Outcome)
	if c.WoundBadges > 0 {
		fmt.Fprintf(&b, ", %d wound badges", c.WoundBadges)
	}
	if s := c.Skills.String(); s != "" {
		fmt.Fprintf(&b, "  [%s]", s)
	}
	if c.Credits > 0 {
		fmt.Fprintf(&b, "  Cr%d", c.Credits)
	}
	if len(c.Benefits) > 0 {
		fmt.Fprintf(&b, "  benefits: %s", strings.Join(c.Benefits, ", "))
	}
	if c.Dead {
		b.WriteString("  DECEASED")
	}
	return b.String()
}
