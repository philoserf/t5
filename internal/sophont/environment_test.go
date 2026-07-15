package sophont

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/uwp"
)

// dryWorld is a homeworld that triggers no locomotion DMs (Atm < 8, Size > 5,
// Hyd < 6), so a locomotion die maps straight to its column.
var dryWorld = uwp.Profile{Size: 8, Atmosphere: 6, Hydrographics: 3, Population: 8}

// TestEnvironmentClean traces a baseline environment: Clear terrain (Environ DM
// 0), a Walker, an Omnivore Hunter/Gatherer.
func TestEnvironmentClean(t *testing.T) {
	seq := append(fluxSeq(0), 3)        // Environ DM Flux 0; locomotion die 3
	seq = append(seq, fluxSeq(0, 0)...) // niche class Flux 0; sub-niche Flux 0
	env := rollEnvironment(dice.NewScripted(seq...), dryWorld)

	if env.EnvironDM != 0 || env.Terrain != "Clear" {
		t.Errorf("terrain = %q (DM %d), want Clear (0)", env.Terrain, env.EnvironDM)
	}
	if env.Locomotion != Walker {
		t.Errorf("locomotion = %v, want Walker", env.Locomotion)
	}
	if env.Class != "Omnivore" || env.Niche != "H/G" {
		t.Errorf("niche = %s/%s, want Omnivore/H/G", env.Class, env.Niche)
	}
}

// TestLocomotionDMs confirms the cumulative homeworld DMs shift the locomotion
// column (chart 05B): on an Ocean world a high-atmosphere die is pulled down.
func TestLocomotionDMs(t *testing.T) {
	// Ocean terrain (Environ DM +4). Homeworld Atm 8 (-2) and Hyd 9 (+2) net 0,
	// so die 6 -> column 5 -> Diver.
	wet := uwp.Profile{Size: 8, Atmosphere: 8, Hydrographics: 9, Population: 8}
	seq := append(fluxSeq(4), 6)
	seq = append(seq, fluxSeq(0, 0)...)
	env := rollEnvironment(dice.NewScripted(seq...), wet)
	if env.Terrain != "Ocean" || env.Locomotion != Diver {
		t.Errorf("got %s/%v, want Ocean/Diver", env.Terrain, env.Locomotion)
	}
}

// TestNicheEnvironDM confirms the Environ DM shifts the sub-niche but not the
// basic class: a +2 DM turns a mid-Flux carnivore into a Chaser.
func TestNicheEnvironDM(t *testing.T) {
	seq := append(fluxSeq(2), 3)        // Environ DM +2; locomotion die
	seq = append(seq, fluxSeq(3, 0)...) // class Flux +3 -> Carnivore; sub Flux 0 (+2 DM)
	env := rollEnvironment(dice.NewScripted(seq...), dryWorld)
	if env.Class != "Carnivore" || env.Niche != "Chaser" {
		t.Errorf("niche = %s/%s, want Carnivore/Chaser", env.Class, env.Niche)
	}
}
