package main

import (
	"testing"

	"github.com/philoserf/t5/internal/chargen"
)

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
