package sophont

import (
	"strings"
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

// TestGenerateInvariants runs Generate across many seeds and checks the
// structural guarantees: six characteristics, a six-letter Genetic Profile of
// valid letters, and a plausible homeworld (Atmosphere 2-9, Population 7+).
func TestGenerateInvariants(t *testing.T) {
	const validGP = "SDAGEVITK C" // the letters gpLetter can emit
	for seed := uint64(1); seed <= 50; seed++ {
		s := Generate(dice.NewWithSeed(seed))

		if len(s.GeneticProfile) != 6 {
			t.Fatalf("seed %d: GP %q, want 6 letters", seed, s.GeneticProfile)
		}
		for _, ch := range s.GeneticProfile {
			if !strings.ContainsRune(validGP, ch) {
				t.Errorf("seed %d: GP %q has invalid letter %c", seed, s.GeneticProfile, ch)
			}
		}
		for i, c := range s.Chars {
			if c.Dice < 1 || c.Dice > 8 {
				t.Errorf("seed %d: char %d die count %d out of range", seed, i, c.Dice)
			}
		}
		if a := s.Homeworld.Profile.Atmosphere; a < 2 || a > 9 {
			t.Errorf("seed %d: homeworld Atm %d, want 2-9", seed, a)
		}
		if p := s.Homeworld.Profile.Population; p < 7 {
			t.Errorf("seed %d: homeworld Pop %d, want 7+", seed, p)
		}
	}
}

// TestGenerateFixesC1C4 confirms the always-present slots: every species has
// C1 = Str and C4 = Int (Book 3 p.228).
func TestGenerateFixesC1C4(t *testing.T) {
	for seed := uint64(1); seed <= 20; seed++ {
		s := Generate(dice.NewWithSeed(seed))
		if s.Chars[0].Name != Str || s.Chars[3].Name != Int {
			t.Errorf("seed %d: C1/C4 = %v/%v, want Str/Int", seed, s.Chars[0].Name, s.Chars[3].Name)
		}
	}
}

// TestHumanPredicate confirms the Human predicate: an all-standard profile is
// Human, one analog slot is not.
func TestHumanPredicate(t *testing.T) {
	human := Species{Chars: [6]CharSpec{{Str, 2}, {Dex, 2}, {End, 2}, {Int, 2}, {Edu, 2}, {Soc, 2}}}
	if !human.Human() {
		t.Error("all-standard profile should be Human")
	}
	alien := human
	alien.Chars[5].Name = Cas
	if alien.Human() {
		t.Error("a Caste C6 should not be Human")
	}
}
