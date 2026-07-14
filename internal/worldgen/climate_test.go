package worldgen

import (
	"slices"
	"testing"

	"github.com/philoserf/t5/internal/uwp"
)

func TestClimateCodes(t *testing.T) {
	temperate := uwp.Profile{Size: 7, Atmosphere: 6, Hydrographics: 5} // Tr/Tu shape
	cases := []struct {
		name    string
		p       uwp.Profile
		orbit   int
		hzOrbit int
		want    []string
	}{
		{"temperate at HZ", temperate, 4, 4, nil},
		// Ho/Co depend on the orbit alone (Chart D gives them no UWP columns), so a
		// world one orbit inside the HZ is Hot whether or not it is also Tropic.
		{"tropic (HZ-1) is also hot", temperate, 3, 4, []string{"Ho", "Tr"}},
		{"tundra (HZ+1) is also cold", temperate, 5, 4, []string{"Co", "Tu"}},
		{"HZ-1 with a non-tropic UWP is still hot", uwp.Profile{Size: 2, Atmosphere: 0}, 3, 4, []string{"Ho"}},
		{"HZ+1 with a non-tropic UWP is still cold", uwp.Profile{Size: 2, Atmosphere: 0}, 5, 4, []string{"Co"}},
		{"frozen (HZ+2)", uwp.Profile{Size: 5, Hydrographics: 6}, 6, 4, []string{"Fr"}},
		{"HZ+2 but no water", uwp.Profile{Size: 5, Hydrographics: 0}, 6, 4, nil},
		{"twilight orbit 0", temperate, 0, 4, []string{"Tz"}},
		{"twilight orbit 1", temperate, 1, 4, []string{"Tz"}},
		{"tropic and twilight", temperate, 1, 2, []string{"Ho", "Tr", "Tz"}},
	}
	for _, c := range cases {
		if got := ClimateCodes(c.p, c.orbit, c.hzOrbit); !slices.Equal(got, c.want) {
			t.Errorf("%s: ClimateCodes = %v, want %v", c.name, got, c.want)
		}
	}
}
