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
		{"tropic (HZ-1)", temperate, 3, 4, []string{"Tr"}},
		{"tundra (HZ+1)", temperate, 5, 4, []string{"Tu"}},
		{"HZ-1 but wrong UWP", uwp.Profile{Size: 2, Atmosphere: 0}, 3, 4, nil},
		{"frozen (HZ+2)", uwp.Profile{Size: 5, Hydrographics: 6}, 6, 4, []string{"Fr"}},
		{"HZ+2 but no water", uwp.Profile{Size: 5, Hydrographics: 0}, 6, 4, nil},
		{"twilight orbit 0", temperate, 0, 4, []string{"Tz"}},
		{"twilight orbit 1", temperate, 1, 4, []string{"Tz"}},
		{"tropic and twilight", temperate, 1, 2, []string{"Tr", "Tz"}},
	}
	for _, c := range cases {
		if got := ClimateCodes(c.p, c.orbit, c.hzOrbit); !slices.Equal(got, c.want) {
			t.Errorf("%s: ClimateCodes = %v, want %v", c.name, got, c.want)
		}
	}
}
