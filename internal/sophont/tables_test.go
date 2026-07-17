package sophont

import "testing"

// These tests lock the interiors of the dense transcribed grids at non-zero Flux
// cells — the Human and Ay goldens only exercise Flux 0, so a swapped or
// shifted cell elsewhere would otherwise ship undetected (the extraction-accuracy
// risk this project treats as its top threat). Each cell below is checked against
// the rendered Book 3 page and chosen to differ from its neighbors.

// idx5 maps a Flux (-5..+5) to an 11-wide table index; idx6 a Flux (-6..+6) to a
// 13-wide one.
func idx5(flux int) int { return flux + 5 }
func idx6(flux int) int { return flux + 6 }

func TestTerrainTableTranscription(t *testing.T) {
	cases := []struct {
		flux int
		want string
	}{{-5, "Mountain"}, {-3, "Exotic"}, {-1, "Rough"}, {2, "Wetlands"}, {5, "Ocean Depths"}}
	for _, c := range cases {
		if got := terrainNames[idx5(c.flux)]; got != c.want {
			t.Errorf("terrain at Flux %+d = %q, want %q", c.flux, got, c.want)
		}
	}
}

func TestLocomotionGridTranscription(t *testing.T) {
	cases := []struct {
		flux, die int // terrain Flux, 1D
		want      Locomotion
	}{
		{-5, 6, Flyer},     // Mountain, die 6
		{-3, 1, Amphibian}, // Exotic, die 1
		{-3, 5, Flyphib},   // Exotic, die 5
		{2, 2, Aquatic},    // Wetlands, die 2
		{2, 5, Triphibian}, // Wetlands, die 5
		{4, 1, Flyphib},    // Ocean, die 1
		{4, 6, Diver},      // Ocean, die 6
		{5, 1, Aquatic},    // Ocean Depths, die 1
	}
	for _, c := range cases {
		if got := locomotionGrid[idx5(c.flux)][c.die-1]; got != c.want {
			t.Errorf("locomotion at Flux %+d die %d = %v, want %v", c.flux, c.die, got, c.want)
		}
	}
}

func TestNicheTablesTranscription(t *testing.T) {
	classes := []struct {
		flux int
		want string
	}{{-6, "Producer"}, {-4, "Herbivore"}, {-2, "Omnivore"}, {3, "Carnivore"}, {5, "Scavenger"}}
	for _, c := range classes {
		if got := nicheClasses[idx6(c.flux)]; got != c.want {
			t.Errorf("niche class at Flux %+d = %q, want %q", c.flux, got, c.want)
		}
	}

	subs := []struct {
		class string
		flux  int
		want  string
	}{
		{"Carnivore", -6, "Pouncer"},
		{"Carnivore", 0, "Chaser"},
		{"Carnivore", 4, "Trapper"},
		{"Carnivore", 5, "Siren"},
		{"Carnivore", 6, "Killer"},
		{"Omnivore", -1, "Gatherer"},
		{"Omnivore", 6, "Eater"},
		{"Scavenger", -3, "Hijacker"},
		{"Scavenger", 5, "Reducer"},
		{"Herbivore", -3, "Intermittent"},
		{"Herbivore", 6, "Filter"},
		{"Producer", 0, "Basker"},
	}
	for _, c := range subs {
		if got := nicheCols[c.class][idx6(c.flux)]; got != c.want {
			t.Errorf("%s sub-niche at Flux %+d = %q, want %q", c.class, c.flux, got, c.want)
		}
	}
}

func TestDiceColumnsTranscription(t *testing.T) {
	cases := []struct {
		column CharName
		flux   int
		want   int
	}{
		{Str, -5, 1}, {Str, 4, 6}, {Str, 5, 7}, {Str, 6, 8}, // only Str scales past 3D
		{Dex, 2, 3}, {Dex, 6, 3}, // analog columns cap at 3D
		{Edu, 3, 2}, {Edu, 4, 3}, // Edu column turns 3D at +4
		{Soc, 6, 2}, // Soc stays 2D throughout
	}
	for _, c := range cases {
		if got := diceLookup(c.column, c.flux); got != c.want {
			t.Errorf("%v column at Flux %+d = %dD, want %dD", c.column, c.flux, got, c.want)
		}
	}
}

func TestCasteColumnsTranscription(t *testing.T) {
	cases := []struct {
		structure CasteStructure
		flux      int
		want      string
	}{
		{Body, -5, "Healer"},
		{Body, -3, "Antibody"},
		{Body, 3, "Voice"},
		{Body, 5, "Claw"},
		{Military, -5, "Medic"},
		{Military, -2, "Scout"},
		{Military, 2, "Warrior"},
		{Economic, 5, "Entrepreneur"},
		{Social, 3, "Patron"},
	}
	for _, c := range cases {
		if got := casteColumns[c.structure][idx5(c.flux)]; got != c.want {
			t.Errorf("%v caste at Flux %+d = %q, want %q", c.structure, c.flux, got, c.want)
		}
	}

	specials := []struct {
		flux int
		want string
	}{{-5, "DeMinimis"}, {-3, "Advisor Minus"}, {4, "Sport"}, {5, "Vice-Leader"}}
	for _, c := range specials {
		if got := specialColumn[idx5(c.flux)]; got != c.want {
			t.Errorf("special caste at Flux %+d = %q, want %q", c.flux, got, c.want)
		}
	}

	uniques := map[CasteStructure]string{Body: "Brain", Military: "General", Social: "Ruler"}
	for structure, want := range uniques {
		if got := casteUnique[structure]; got != want {
			t.Errorf("%v Unique caste = %q, want %q", structure, got, want)
		}
	}
}

func TestGenderColumnsTranscription(t *testing.T) {
	cases := []struct {
		structure GenderStructure
		flux      int
		want      string
	}{
		{Dual, -1, "Male"},
		{Dual, 0, "Female"},
		{Dual, 4, "Female"},
		{Dual, 5, "Male"},
		{EAB, -2, "Activator"},
		{EAB, 1, "Bearer"},
		{FMN, 1, "Neuter"},
		{FMN, 4, "Male"},
		{Group, -1, "Two"},
		{Group, 0, "One"},
		{Group, 5, "Six"},
	}
	for _, c := range cases {
		if got := genderColumns[c.structure][idx5(c.flux)]; got != c.want {
			t.Errorf("%v gender at Flux %+d = %q, want %q", c.structure, c.flux, got, c.want)
		}
	}
}
