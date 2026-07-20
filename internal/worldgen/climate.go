package worldgen

import (
	"github.com/philoserf/t5/internal/tradecode"
	"github.com/philoserf/t5/internal/uwp"
)

// ClimateCodes returns the climate / orbit-dependent trade classifications a world
// earns from where it sits relative to its star's habitable zone (Book 3 Chart D,
// p.26; the HZ offset comes from Chart B, p.24). These are the codes
// TradeClassifications intentionally omits because they need a concrete orbit —
// systemgen supplies it once the world is placed.
//
// The chart splits them into two kinds, and the distinction is the whole point:
//
//   - Ho (Hot) and Co (Cold) depend on the ORBIT ALONE. Chart D gives them no Size,
//     Atmosphere, Hydrographics, Population, Government, or Law constraint — every
//     column is "--". A world one orbit inside the habitable zone is Hot and one
//     orbit outside it is Cold, whatever else it is.
//   - Tr (Tropic), Tu (Tundra), and Fr (Frozen) need the orbit AND a UWP that can
//     carry the climate: a temperate-band world for Tr/Tu, a water-bearing one for
//     Fr.
//
// So Ho and Tr describe the same orbit, and a world there is Hot whether or not it
// is also Tropic — which is why these are not mutually exclusive cases. Ho and Co
// were previously omitted on the grounds that "this edition's Chart B does not emit
// the broader T5SS Hot/Cold pair"; Chart D, the chart this function implements,
// defines both.
//
// Tz (Twilight Zone) is the exception that does not need a habitable zone at all:
// Chart D defines it as "Orbit 0-1", full stop. So a world in orbit 0 or 1 of a
// star with no habitable zone still earns Tz, even though hasHZ is false and the
// offset codes above cannot be computed (there is no zone to offset from).
func ClimateCodes(p uwp.Profile, orbit, hzOrbit int, hasHZ bool) []tradecode.Code {
	var out []tradecode.Code

	if hasHZ {
		switch offset := orbit - hzOrbit; {
		case offset == -1:
			out = append(out, tradecode.Ho) // orbit alone
			if tropicUWP(p) {
				out = append(out, tradecode.Tr)
			}
		case offset == 1:
			out = append(out, tradecode.Co) // orbit alone
			if tropicUWP(p) {
				out = append(out, tradecode.Tu)
			}
		case offset >= 2 && frozenUWP(p):
			out = append(out, tradecode.Fr)
		}
	}

	if orbit == 0 || orbit == 1 {
		out = append(out, tradecode.Tz)
	}

	return out
}

// tropicUWP reports the temperate-band shape Tr/Tu share (Book 3 p.26): a
// mid-to-large world with a breathable-ish atmosphere and moderate hydrographics.
func tropicUWP(p uwp.Profile) bool {
	return allows("6789", p.Size) && allows("456789", p.Atmosphere) &&
		allows("34567", p.Hydrographics)
}

// frozenUWP reports the Frozen shape (Book 3 p.26): any sized world that still
// carries water/ice.
func frozenUWP(p uwp.Profile) bool {
	return allows("23456789", p.Size) && allows("123456789A", p.Hydrographics)
}
