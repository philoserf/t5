// Package sophont creates non-human species with the Traveller5 Sophont Creation
// System (Book 3 pp.217-239). Generate produces a Species template — the
// "Sophont Creation Card" — that character generation later consumes to make
// individuals of the species.
//
// This is the core spine: the evolutionary Environment (step 05), the six-slot
// Characteristics profile and Genetic Profile (step 06). The homeworld reuses
// the world generator (the book, p.217, blesses substituting it), rerolled until
// it is a plausible cradle — Atmosphere 2-9 and Population 7+. Later phases add
// Size, Life Stages, Caste, and Gender; the physical/appearance layers (senses,
// body structure, special abilities) are background detail deferred until a
// consumer needs them.
package sophont

import (
	"fmt"
	"strings"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// A Species is a generated sophont template: its homeworld, evolutionary
// environment, and the six-characteristic profile with its Genetic Profile. A
// species is Human when every characteristic is Human-standard, giving the
// reference Genetic Profile "SDEIES".
type Species struct {
	Homeworld      worldgen.World
	Environment    Environment
	Chars          [6]CharSpec
	GeneticProfile string
	Size           int // average volume in liters, ~mass in kg (Human = 72)
	LifeCycle      LifeCycle
	Gender         Gender
	Caste          *Caste // nil unless C6 is the Caste characteristic
}

// Generate rolls a complete species: a plausible homeworld, the environment it
// evolved in, its characteristic profile, and its average Size.
func Generate(r *dice.Roller) Species {
	home := plausibleHomeworld(r)
	env := rollEnvironment(r, home.Profile)
	chars, gp := rollCharacteristics(r, env)
	gender := rollGender(r)

	s := Species{
		Homeworld:      home,
		Environment:    env,
		Chars:          chars,
		GeneticProfile: gp,
		Size:           Size(chars),
		LifeCycle:      rollLifeCycle(r),
		Gender:         gender,
	}
	if chars[5].Name == Cas { // C6 is Caste
		caste := rollCaste(r, gender)
		s.Caste = &caste
	}

	return s
}

// clamp constrains v to the inclusive range [lo, hi]. Flux-indexed table
// lookups use it to bound a signed Flux (plus modifiers) before offsetting into
// a slice.
func clamp(v, lo, hi int) int {
	return min(max(v, lo), hi)
}

// plausibleHomeworld rolls mainworlds until one is a plausible sophont cradle:
// Atmosphere 2-9 and Population 7+ (Book 3 pp.217-218).
func plausibleHomeworld(r *dice.Roller) worldgen.World {
	for {
		w := worldgen.GenerateWorld(r, 0, 0, false)
		if a := w.Profile.Atmosphere; a >= 2 && a <= 9 && w.Profile.Population >= 7 {
			return w
		}
	}
}

// Human reports whether every characteristic is Human-standard.
func (s Species) Human() bool {
	for _, c := range s.Chars {
		switch c.Name {
		case Str, Dex, End, Int, Edu, Soc:
		default:
			return false
		}
	}

	return true
}

// String renders a one-line species summary: Genetic Profile, homeworld UWP,
// and the evolutionary environment.
func (s Species) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s", s.GeneticProfile, s.Homeworld.Profile)
	fmt.Fprintf(
		&b,
		" %s/%s/%s %s Size %d Lifespan %d %s",
		s.Environment.Terrain,
		s.Environment.Locomotion,
		s.Environment.Class,
		s.Environment.Niche,
		s.Size,
		s.LifeCycle.Lifespan,
		s.Gender.Structure,
	)

	if s.Caste != nil {
		fmt.Fprintf(&b, " %s-caste", s.Caste.Structure)
	}

	return b.String()
}
