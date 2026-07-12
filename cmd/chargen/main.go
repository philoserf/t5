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
	"github.com/philoserf/t5/internal/worldgen"
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
		// A homeworld is an input to character generation (selected, assigned, or
		// rolled). Here it is a freshly generated world with no system context.
		homeworld := worldgen.GenerateWorld(r, 0, 0, false)
		fmt.Println(render(chargen.GenerateCareered(r, chargen.DefaultPolicy{}, homeworld, career), career))
	}
}

// careerByName resolves a -career flag value to its career data.
func careerByName(name string) (chargen.Career, error) {
	switch strings.ToLower(name) {
	case "scout":
		return chargen.ScoutCareer, nil
	case "rogue":
		return chargen.RogueCareer, nil
	case "soldier":
		return chargen.SoldierCareer, nil
	case "marine":
		return chargen.MarineCareer, nil
	case "spacer":
		return chargen.SpacerCareer, nil
	default:
		return chargen.Career{}, fmt.Errorf("unknown career %q (known: scout, rogue, soldier, marine, spacer)", name)
	}
}

// render formats a careered character as a one-line sheet.
func render(c chargen.Character, career chargen.Career) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s  age %d  homeworld %s", c.UPP(), c.Age, c.Homeworld.Profile)
	if len(c.Degrees) > 0 {
		fmt.Fprintf(&b, "  %s", strings.Join(c.Degrees, " "))
		if c.Major != "" {
			fmt.Fprintf(&b, " (%s)", strings.Join(subjects(c), "/"))
		}
	}
	if len(c.Careers) == 0 {
		fmt.Fprintf(&b, "  did not qualify for %s", career.Name)
		renderTail(&b, c)
		return b.String()
	}
	rec := c.Careers[len(c.Careers)-1]
	fmt.Fprintf(&b, "  %s: %d terms, %s", career.Name, rec.Terms, rec.Outcome)
	if title := rankTitle(career, rec); title != "" {
		fmt.Fprintf(&b, ", %s", title)
	}
	if c.WoundBadges > 0 {
		fmt.Fprintf(&b, ", %d wound badges", c.WoundBadges)
	}
	renderTail(&b, c)
	return b.String()
}

// rankTitle returns the character's final rank title, or "" for a rankless career.
func rankTitle(career chargen.Career, rec chargen.CareerRecord) string {
	ranks := career.EnlistedRanks
	if rec.Officer {
		ranks = career.OfficerRanks
	}
	if rec.Rank >= 1 && rec.Rank <= len(ranks) {
		return ranks[rec.Rank-1].Title
	}
	return ""
}

// subjects lists a graduate's Major and, if declared, Minor.
func subjects(c chargen.Character) []string {
	s := []string{c.Major}
	if c.Minor != "" {
		s = append(s, c.Minor)
	}
	return s
}

// renderTail appends the skills, credits, benefits, and deceased marker common
// to careered and non-qualified characters (both carry homeworld skills).
func renderTail(b *strings.Builder, c chargen.Character) {
	if s := c.Skills.String(); s != "" {
		fmt.Fprintf(b, "  [%s]", s)
	}
	if c.Credits > 0 {
		fmt.Fprintf(b, "  Cr%d", c.Credits)
	}
	if len(c.Benefits) > 0 {
		fmt.Fprintf(b, "  benefits: %s", strings.Join(c.Benefits, ", "))
	}
	if c.Dead {
		b.WriteString("  DECEASED")
	}
}
