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
	careerName := flag.String("career", "", `career(s) to run, comma-separated for a sequence (e.g. "scout" or "scout,merchant"); empty prints UPP only`)
	n, r := cli.SeededRoller("characters")

	if *careerName == "" {
		for range n {
			fmt.Println(chargen.Generate(r))
		}
		return
	}

	careers, err := careersByNames(*careerName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "chargen:", err)
		os.Exit(2)
	}
	for range n {
		// A homeworld is an input to character generation (selected, assigned, or
		// rolled). Here it is a freshly generated world with no system context.
		homeworld := worldgen.GenerateWorld(r, 0, 0, false)
		p := &sequencePolicy{remaining: careers[1:]}
		fmt.Println(render(chargen.GenerateCareered(r, p, homeworld, careers[0])))
	}
}

// careersByNames resolves a comma-separated list of career names.
func careersByNames(list string) ([]chargen.Career, error) {
	names := strings.Split(list, ",")
	careers := make([]chargen.Career, 0, len(names))
	for _, name := range names {
		c, err := careerByName(strings.TrimSpace(name))
		if err != nil {
			return nil, err
		}
		careers = append(careers, c)
	}
	return careers, nil
}

// sequencePolicy is DefaultPolicy that serves a fixed list of subsequent careers.
type sequencePolicy struct {
	chargen.DefaultPolicy
	remaining []chargen.Career
	i         int
}

func (p *sequencePolicy) NextCareer(chargen.Character) (chargen.Career, bool) {
	if p.i < len(p.remaining) {
		c := p.remaining[p.i]
		p.i++
		return c, true
	}
	return chargen.Career{}, false
}

// allCareers is every career indexed by its CareerID, for looking up a career
// from a record (which stores only the ID).
var allCareers = []chargen.Career{
	chargen.ScoutCareer, chargen.RogueCareer, chargen.SoldierCareer, chargen.MarineCareer,
	chargen.SpacerCareer, chargen.AgentCareer, chargen.CitizenCareer, chargen.EntertainerCareer,
	chargen.CraftsmanCareer, chargen.ScholarCareer, chargen.FunctionaryCareer, chargen.NobleCareer,
	chargen.MerchantCareer,
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
	case "agent":
		return chargen.AgentCareer, nil
	case "citizen":
		return chargen.CitizenCareer, nil
	case "entertainer":
		return chargen.EntertainerCareer, nil
	case "craftsman":
		return chargen.CraftsmanCareer, nil
	case "scholar":
		return chargen.ScholarCareer, nil
	case "functionary":
		return chargen.FunctionaryCareer, nil
	case "noble":
		return chargen.NobleCareer, nil
	case "merchant":
		return chargen.MerchantCareer, nil
	default:
		return chargen.Career{}, fmt.Errorf("unknown career %q (known: scout, rogue, soldier, marine, spacer, agent, citizen, entertainer, craftsman, scholar, functionary, noble, merchant)", name)
	}
}

// render formats a careered character as a one-line sheet.
func render(c chargen.Character) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s  age %d  homeworld %s", c.UPP(), c.Age, c.Homeworld.Profile)
	if len(c.Degrees) > 0 {
		fmt.Fprintf(&b, "  %s", strings.Join(c.Degrees, " "))
		if c.Major != "" {
			fmt.Fprintf(&b, " (%s)", strings.Join(subjects(c), "/"))
		}
	}
	if len(c.Careers) == 0 {
		b.WriteString("  did not qualify")
		renderTail(&b, c)
		return b.String()
	}
	for _, rec := range c.Careers {
		career := allCareers[rec.Career]
		fmt.Fprintf(&b, "  %s: %d terms, %s", career.Name, rec.Terms, rec.Outcome)
		title := rankTitle(career, rec)
		if career.ReturnIntrigue {
			title = chargen.NobleTitle(c.Score(chargen.Social)) // the Noble's rank is their Social Standing
		}
		if title != "" {
			fmt.Fprintf(&b, " (%s)", title)
		}
	}
	if c.WoundBadges > 0 {
		fmt.Fprintf(&b, ", %d wound badges", c.WoundBadges)
	}
	if c.Fame > 0 || c.Talent > 0 {
		fmt.Fprintf(&b, ", Fame %d Talent %d", c.Fame, c.Talent)
	}
	if c.Masterpieces > 0 {
		fmt.Fprintf(&b, ", %d masterpieces (Cr%d)", c.Masterpieces, c.MasterpieceValue)
	}
	if c.Publications > 0 {
		fmt.Fprintf(&b, ", %d publications", c.Publications)
	}
	if c.Discoveries > 0 {
		fmt.Fprintf(&b, ", %d discoveries", c.Discoveries)
	}
	if c.LandGrants > 0 {
		fmt.Fprintf(&b, ", %d land grants", c.LandGrants)
	}
	if c.ShipShares > 0 {
		fmt.Fprintf(&b, ", %d ship shares", c.ShipShares)
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
