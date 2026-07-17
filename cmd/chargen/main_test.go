package main

import (
	"strings"
	"testing"

	"github.com/philoserf/t5/internal/chargen"
	"github.com/philoserf/t5/internal/dice"
)

// served builds a Character who ran one career, for exercising the sheet's
// rendering branches without generating a whole life.
func served(rec chargen.CareerRecord) chargen.Character {
	return chargen.Character{Age: 26, Careers: []chargen.CareerRecord{rec}}
}

// TestRankOf covers the rank title the sheet reports — including the two cases
// it deliberately stays silent about.
func TestRankOf(t *testing.T) {
	soldier := func(rank int, officer bool) chargen.CareerRecord {
		return chargen.CareerRecord{Career: chargen.Soldier, Terms: 3, Rank: rank, Officer: officer}
	}

	cases := []struct {
		name string
		c    chargen.Character
		rec  chargen.CareerRecord
		want string
	}{
		// The entry enlisted rank is not worth reporting: "mustered out as
		// Private after 1 term" says nothing "after 1 term" hasn't already said.
		{"entry enlisted rank is silent", served(soldier(1, false)), soldier(1, false), ""},
		{"promoted enlisted", served(soldier(3, false)), soldier(3, false), "Sergeant"},
		{
			"officer track reads the other ladder",
			served(soldier(1, true)),
			soldier(1, true),
			"2nd Lieutenant",
		},
		{"officer promoted", served(soldier(3, true)), soldier(3, true), "Captain"},
		// A rankless career has no ladder to read at all.
		{
			"rankless career", served(chargen.CareerRecord{Career: chargen.Scout, Rank: 0}),
			chargen.CareerRecord{Career: chargen.Scout, Rank: 0},
			"",
		},
		// A rank past the end of the ladder must not index out of bounds.
		{"rank past the ladder", served(soldier(99, false)), soldier(99, false), ""},
	}
	for _, c := range cases {
		if got := rankOf(c.c, c.rec); got != c.want {
			t.Errorf("%s: rankOf = %q, want %q", c.name, got, c.want)
		}
	}

	// A Noble's rank is their Social Standing, not a ladder position. Roll a
	// character whose Soc is 11 (2D = 6+5) — a Knight.
	noble := chargen.Generate(dice.NewScripted(1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 6, 5))

	rec := chargen.CareerRecord{Career: chargen.Noble, Terms: 2}
	if got := rankOf(noble, rec); got != "Knight" {
		t.Errorf("Noble with Soc 11: rankOf = %q, want Knight", got)
	}
}

// TestSummaryLine covers the sheet's headline in each of its three shapes, and
// the deceased suffix it must not double up.
func TestSummaryLine(t *testing.T) {
	// No career at all.
	if got := summaryLine(
		chargen.Character{Age: 18},
	); got != "Age 18 — did not qualify for a career\n" {
		t.Errorf("careerless = %q", got)
	}

	if got := summaryLine(
		chargen.Character{Age: 18, Dead: true},
	); !strings.Contains(
		got,
		", deceased",
	) {
		t.Errorf("a dead careerless character should say so: %q", got)
	}

	// One career: the headline carries the outcome, rank, and term count.
	one := served(chargen.CareerRecord{
		Career: chargen.Soldier, Terms: 3, Rank: 3, Outcome: chargen.MusteredOut,
	})

	want := "Soldier — age 26, mustered out as Sergeant after 3 terms\n"
	if got := summaryLine(one); got != want {
		t.Errorf("single career = %q, want %q", got, want)
	}

	// "died" already conveys death, so the deceased suffix must not repeat it.
	died := served(chargen.CareerRecord{Career: chargen.Soldier, Terms: 1, Outcome: chargen.Died})
	died.Dead = true

	got := summaryLine(died)
	if !strings.Contains(got, "died") {
		t.Errorf("a fatal career should say died: %q", got)
	}

	if strings.Contains(got, "deceased") {
		t.Errorf("%q says both died and deceased", got)
	}

	// Several careers: they are named in order, with no single-career detail.
	many := chargen.Character{Age: 42, Careers: []chargen.CareerRecord{
		{Career: chargen.Scout, Terms: 2, Outcome: chargen.MusteredOut},
		{Career: chargen.Citizen, Terms: 1, Outcome: chargen.MusteredOut},
		{Career: chargen.Merchant, Terms: 2, Outcome: chargen.MusteredOut},
	}}
	if got := summaryLine(many); got != "Scout, Citizen, then Merchant — age 42\n" {
		t.Errorf("multi-career = %q", got)
	}
}

// TestEducationField: degrees and subjects are independent — a Trade School
// graduate has a Major but no degree, and each combination renders on its own.
func TestEducationField(t *testing.T) {
	cases := []struct {
		name string
		c    chargen.Character
		want string
	}{
		{"uneducated", chargen.Character{}, ""},
		{"degree only", chargen.Character{Degrees: []string{"BA"}}, "BA"},
		{"trade school: a major, no degree", chargen.Character{Major: "Trader"}, "Trader (major)"},
		{
			"degree and subjects",
			chargen.Character{Degrees: []string{"BA"}, Major: "History", Minor: "Art"},
			"BA — History (major), Art (minor)",
		},
		{
			"a major with no minor",
			chargen.Character{Degrees: []string{"BA", "MA"}, Major: "History"},
			"BA, MA — History (major)",
		},
	}
	for _, c := range cases {
		if got := educationField(c.c); got != c.want {
			t.Errorf("%s: educationField = %q, want %q", c.name, got, c.want)
		}
	}
}

// TestRenderOmitsEmptyFields: the sheet lists a field only when the character has
// it, so a plain character's sheet is not padded with zeroes and empty labels.
func TestRenderOmitsEmptyFields(t *testing.T) {
	c := served(chargen.CareerRecord{Career: chargen.Scout, Terms: 1, Outcome: chargen.MusteredOut})

	out := render(c)
	if !strings.Contains(out, "Scout") || !strings.Contains(out, "UPP") {
		t.Fatalf("sheet is missing its headline or UPP:\n%s", out)
	}

	for _, absent := range []string{"Wealth", "Benefits", "Education", "Masterpieces", "Publications", "Service"} {
		if strings.Contains(out, absent) {
			t.Errorf("a character with no %s should not show that field:\n%s", absent, out)
		}
	}

	// Given the things, it shows them — including Service, which appears only
	// for a life of more than one career.
	rich := chargen.Character{
		Age: 40, Credits: 50000, Benefits: []string{"TAS"}, Publications: 2,
		Careers: []chargen.CareerRecord{
			{Career: chargen.Scholar, Terms: 2, Outcome: chargen.MusteredOut},
			{Career: chargen.Citizen, Terms: 1, Outcome: chargen.MusteredOut},
		},
	}

	out = render(rich)
	for _, want := range []string{"Service", "Publications", "Wealth", "Cr50,000", "Benefits", "TAS"} {
		if !strings.Contains(out, want) {
			t.Errorf("sheet is missing %s:\n%s", want, out)
		}
	}
}

func TestCommas(t *testing.T) {
	cases := map[int]string{
		0: "0", 5: "5", 100: "100", 1000: "1,000",
		2000000: "2,000,000", 230000: "230,000", -12345: "-12,345",
	}
	for n, want := range cases {
		if got := commas(n); got != want {
			t.Errorf("commas(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestPlural(t *testing.T) {
	cases := []struct {
		n    int
		noun string
		want string
	}{
		{1, "term", "1 term"},
		{2, "term", "2 terms"},
		{0, "term", "0 terms"},
		{1, "discovery", "1 discovery"},
		{3, "discovery", "3 discoveries"}, // consonant + y -> ies
		{2, "wound badge", "2 wound badges"},
	}
	for _, c := range cases {
		if got := plural(c.n, c.noun); got != c.want {
			t.Errorf("plural(%d, %q) = %q, want %q", c.n, c.noun, got, c.want)
		}
	}
}

func TestOutcomePhrase(t *testing.T) {
	cases := map[chargen.TermOutcome]string{
		chargen.MusteredOut: "mustered out",
		chargen.Disabled:    "disabled",
		chargen.Died:        "died",
		chargen.Ongoing:     "still serving",
	}
	for o, want := range cases {
		if got := outcomePhrase(o); got != want {
			t.Errorf("outcomePhrase(%v) = %q, want %q", o, got, want)
		}
	}
}
