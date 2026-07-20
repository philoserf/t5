package worldgen

import (
	"slices"
	"testing"

	"github.com/philoserf/t5/internal/tradecode"
	"github.com/philoserf/t5/internal/uwp"
)

func TestClimateCodes(t *testing.T) {
	temperate := uwp.Profile{Size: 7, Atmosphere: 6, Hydrographics: 5} // Tr/Tu shape

	cases := []struct {
		name    string
		p       uwp.Profile
		orbit   int
		hzOrbit int
		hasHZ   bool
		want    []tradecode.Code
	}{
		{"temperate at HZ", temperate, 4, 4, true, nil},
		// Ho/Co depend on the orbit alone (Chart D gives them no UWP columns), so a
		// world one orbit inside the HZ is Hot whether or not it is also Tropic.
		{"tropic (HZ-1) is also hot", temperate, 3, 4, true, []tradecode.Code{"Ho", "Tr"}},
		{"tundra (HZ+1) is also cold", temperate, 5, 4, true, []tradecode.Code{"Co", "Tu"}},
		{
			"HZ-1 with a non-tropic UWP is still hot",
			uwp.Profile{Size: 2, Atmosphere: 0},
			3,
			4,
			true,
			[]tradecode.Code{"Ho"},
		},
		{
			"HZ+1 with a non-tropic UWP is still cold",
			uwp.Profile{Size: 2, Atmosphere: 0},
			5,
			4,
			true,
			[]tradecode.Code{"Co"},
		},
		{"frozen (HZ+2)", uwp.Profile{Size: 5, Hydrographics: 6}, 6, 4, true, []tradecode.Code{"Fr"}},
		{"HZ+2 but no water", uwp.Profile{Size: 5, Hydrographics: 0}, 6, 4, true, nil},
		{"twilight orbit 0", temperate, 0, 4, true, []tradecode.Code{"Tz"}},
		{"twilight orbit 1", temperate, 1, 4, true, []tradecode.Code{"Tz"}},
		{"tropic and twilight", temperate, 1, 2, true, []tradecode.Code{"Ho", "Tr", "Tz"}},
		// Tz needs no habitable zone: a world in orbit 0-1 of a no-HZ star still
		// earns it, and earns nothing else (the offset codes have no zone to work
		// from).
		{"twilight without a habitable zone", temperate, 1, 0, false, []tradecode.Code{"Tz"}},
		{"no habitable zone, not orbit 0-1", temperate, 3, 0, false, nil},
	}
	for _, c := range cases {
		if got := ClimateCodes(c.p, c.orbit, c.hzOrbit, c.hasHZ); !slices.Equal(got, c.want) {
			t.Errorf("%s: ClimateCodes = %v, want %v", c.name, got, c.want)
		}
	}
}
