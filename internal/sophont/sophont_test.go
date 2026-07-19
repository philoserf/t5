package sophont

import (
	"strings"
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

// validGP is exactly the set of letters gpLetter can emit; TestValidGPIsExact
// keeps it honest, so a Genetic Profile check against it is a real check.
const validGP = "SDAGEVITKC"

// TestValidGPIsExact pins the validGP set to charInfo: every emitted letter must
// appear in the set, and every letter in the set must be emitted. The second
// direction is what catches a stray character (a space, a typo) silently
// widening the Genetic Profile validation in TestGenerateInvariants.
func TestValidGPIsExact(t *testing.T) {
	emitted := map[byte]bool{}

	for name, info := range charInfo {
		if !strings.ContainsRune(validGP, rune(info.gp)) {
			t.Errorf("%v emits GP letter %q, absent from validGP %q", name, info.gp, validGP)
		}

		emitted[info.gp] = true
	}

	for i := range len(validGP) {
		if !emitted[validGP[i]] {
			t.Errorf("validGP %q contains %q, which no characteristic emits", validGP, validGP[i])
		}
	}
}

// TestGenerateInvariants runs Generate across many seeds and checks the
// structural guarantees: six characteristics, a six-letter Genetic Profile of
// valid letters, and a plausible homeworld (Atmosphere 2-9, Population 7+).
func TestGenerateInvariants(t *testing.T) {
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
			// A Caste C6 is the one slot with no rolled value: exactly 0 dice.
			if c.Name == Cas {
				if c.Dice != 0 {
					t.Errorf("seed %d: Caste C6 die count %d, want 0", seed, c.Dice)
				}

				continue
			}

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
	// Die counts vary independently of the names: chart 06B can hand an
	// all-standard-name species a 6D Strength (average 21). That is not Human.
	strong := human

	strong.Chars[0].Dice = 6
	if strong.Human() {
		t.Error("an SDEIES species with 6D Strength should not be Human")
	}

	frail := human

	frail.Chars[2].Dice = 1
	if frail.Human() {
		t.Error("an SDEIES species with 1D Endurance should not be Human")
	}
}
