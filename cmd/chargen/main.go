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
	"strconv"
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
		cli.Fatalf("%v", err)
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

// careerByName resolves a -career flag value to its career data.
func careerByName(name string) (chargen.Career, error) {
	if c, ok := chargen.CareerByName(name); ok {
		return c, nil
	}
	known := strings.ToLower(strings.Join(chargen.CareerNames(), ", "))
	return chargen.Career{}, fmt.Errorf("unknown career %q (known: %s)", name, known)
}

// labelWidth aligns the field labels of the character sheet (the widest is
// "Commendations").
const labelWidth = 13

// render formats a careered character as a labeled, multi-line sheet.
func render(c chargen.Character) string {
	var b strings.Builder
	b.WriteString(summaryLine(c))
	field(&b, "UPP", c.UPP())
	if hw := homeworldField(c); hw != "" {
		field(&b, "Homeworld", hw)
	}
	if ed := educationField(c); ed != "" {
		field(&b, "Education", ed)
	}
	if len(c.Careers) > 1 { // one career is summarized in the header line
		field(&b, "Service", serviceField(c))
	}
	if skills := c.Skills.List(); len(skills) > 0 {
		field(&b, "Skills", strings.Join(skills, ", "))
	}
	renderAchievements(&b, c)
	if c.Credits > 0 {
		field(&b, "Wealth", "Cr"+commas(c.Credits))
	}
	if len(c.Benefits) > 0 {
		field(&b, "Benefits", strings.Join(c.Benefits, ", "))
	}
	if len(c.Entitlements) > 0 {
		field(&b, "Entitlements", entitlementsField(c))
	}
	return b.String()
}

// summaryLine is the sheet's headline: the career(s), age, and — for a single
// career — how and after how long it ended.
func summaryLine(c chargen.Character) string {
	deceased := ""
	if c.Dead {
		deceased = ", deceased"
	}
	switch len(c.Careers) {
	case 0:
		return fmt.Sprintf("Age %d — did not qualify for a career%s\n", c.Age, deceased)
	case 1:
		rec := c.Careers[0]
		detail := outcomePhrase(rec.Outcome)
		if rank := rankOf(c, rec); rank != "" {
			detail += " as " + rank // rank binds to the muster-out, not the term count
		}
		detail += " after " + plural(rec.Terms, "term")
		if rec.Outcome == chargen.Died {
			deceased = "" // "died" already conveys it
		}
		return fmt.Sprintf("%s — age %d, %s%s\n", chargen.CareerByID(rec.Career).Name, c.Age, detail, deceased)
	default:
		return fmt.Sprintf("%s — age %d%s\n", careerNames(c), c.Age, deceased)
	}
}

// serviceField lists each career of a multi-career life on its own line.
func serviceField(c chargen.Character) string {
	var lines []string
	for _, rec := range c.Careers {
		s := fmt.Sprintf("%s: %s, %s", chargen.CareerByID(rec.Career).Name, plural(rec.Terms, "term"), outcomePhrase(rec.Outcome))
		if rank := rankOf(c, rec); rank != "" {
			s += " as " + rank
		}
		lines = append(lines, s)
	}
	return strings.Join(lines, "\n")
}

// renderAchievements appends a labeled line for each career token the character
// accumulated.
func renderAchievements(b *strings.Builder, c chargen.Character) {
	if c.Fame > 0 || c.Talent > 0 {
		var parts []string
		if c.Fame > 0 {
			parts = append(parts, fmt.Sprintf("Fame %d", c.Fame))
		}
		if c.Talent > 0 {
			parts = append(parts, fmt.Sprintf("Talent %d", c.Talent))
		}
		field(b, "Reputation", strings.Join(parts, ", "))
	}
	if c.Masterpieces > 0 {
		field(b, "Masterpieces", fmt.Sprintf("%d (Cr%s)", c.Masterpieces, commas(c.MasterpieceValue)))
	}
	// Plain counted awards, each a labeled line only when non-zero.
	for _, a := range []struct {
		n     int
		label string
	}{
		{c.Publications, "Publications"},
		{c.Commendations, "Commendations"},
		{c.Discoveries, "Discoveries"},
		{c.LandGrants, "Land grants"},
		{c.ShipShares, "Ship shares"},
		{c.WoundBadges, "Wound badges"},
	} {
		if a.n > 0 {
			field(b, a.label, strconv.Itoa(a.n))
		}
	}
}

// field writes one "  Label   value" row; a value with embedded newlines has its
// continuation lines aligned under the first.
func field(b *strings.Builder, label, value string) {
	lines := strings.Split(value, "\n")
	fmt.Fprintf(b, "  %-*s  %s\n", labelWidth, label, lines[0])
	pad := strings.Repeat(" ", 2+labelWidth+2)
	for _, line := range lines[1:] {
		fmt.Fprintf(b, "%s%s\n", pad, line)
	}
}

// entitlementsField lists each recurring annual entitlement on its own line, as
// "Source: CrN/year", noting the starting age for one that has not begun yet
// (a pension the character will collect only at retirement).
func entitlementsField(c chargen.Character) string {
	lines := make([]string, len(c.Entitlements))
	for i, e := range c.Entitlements {
		line := fmt.Sprintf("%s: Cr%s/year", e.Source, commas(e.PerYear))
		if e.FromAge > c.Age {
			line += fmt.Sprintf(" (from age %d)", e.FromAge)
		}
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

// homeworldField is the homeworld's UWP and any trade classifications.
func homeworldField(c chargen.Character) string {
	uwp := c.Homeworld.Profile.String()
	if uwp == "" {
		return ""
	}
	if len(c.Homeworld.TradeCodes) > 0 {
		uwp += "   " + strings.Join(c.Homeworld.TradeCodes, " ")
	}
	return uwp
}

// educationField renders the character's degrees and, if declared, their Major
// and Minor subjects. A Trade School graduate has a Major but no degree, so the
// subjects show on their own.
func educationField(c chargen.Character) string {
	if len(c.Degrees) == 0 && c.Major == "" {
		return ""
	}
	s := strings.Join(c.Degrees, ", ")
	if c.Major != "" {
		subjects := c.Major + " (major)"
		if c.Minor != "" {
			subjects += ", " + c.Minor + " (minor)"
		}
		if s == "" {
			s = subjects
		} else {
			s += " — " + subjects
		}
	}
	return s
}

// careerNames joins a multi-career life's career names, e.g. "Citizen, then
// Merchant".
func careerNames(c chargen.Character) string {
	names := make([]string, len(c.Careers))
	for i, rec := range c.Careers {
		names[i] = chargen.CareerByID(rec.Career).Name
	}
	last := len(names) - 1
	return strings.Join(names[:last], ", ") + ", then " + names[last]
}

// rankOf returns a rank title worth reporting: the Noble's Social-Standing
// title, an officer's rank, or an enlisted rank the character was promoted into.
// It returns "" for a rankless career and for the entry enlisted rank (rank 1) a
// character never rose above — "mustered out as Temp/Private" is just noise, as
// "after 1 term" already implies no promotion.
func rankOf(c chargen.Character, rec chargen.CareerRecord) string {
	career := chargen.CareerByID(rec.Career)
	if career.ReturnIntrigue {
		return chargen.NobleTitle(c.Score(chargen.Social))
	}
	ranks := career.EnlistedRanks
	if rec.Officer {
		ranks = career.OfficerRanks
	} else if rec.Rank <= 1 {
		return "" // never promoted above the entry rank
	}
	if rec.Rank >= 1 && rec.Rank <= len(ranks) {
		return ranks[rec.Rank-1].Title
	}
	return ""
}

// outcomePhrase renders a term outcome in plain words.
func outcomePhrase(o chargen.TermOutcome) string {
	switch o {
	case chargen.MusteredOut:
		return "mustered out"
	case chargen.Disabled:
		return "disabled"
	case chargen.Died:
		return "died"
	default:
		return "still serving"
	}
}

// plural renders a count with its noun, pluralized for counts other than one.
func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return strconv.Itoa(n) + " " + pluralize(noun)
}

// pluralize forms the plural of a noun (handling a trailing consonant + "y").
func pluralize(noun string) string {
	if strings.HasSuffix(noun, "y") && !strings.ContainsRune("aeiou", rune(noun[len(noun)-2])) {
		return noun[:len(noun)-1] + "ies"
	}
	return noun + "s"
}

// commas groups an integer with thousands separators, e.g. 2000000 -> "2,000,000".
func commas(n int) string {
	if n < 0 {
		return "-" + commas(-n)
	}
	s := strconv.Itoa(n)
	var out strings.Builder
	for i := range len(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out.WriteByte(',')
		}
		out.WriteByte(s[i])
	}
	return out.String()
}
