package worldgen

import "github.com/philoserf/t5/internal/uwp"

// ClimateCodes returns the climate / orbit-dependent trade classifications a
// mainworld earns from where it sits relative to its star's habitable zone
// (Book 3 Chart D, p.26; the HZ offset comes from Chart B, p.24). These are the
// codes TradeClassifications intentionally omits because they need a concrete
// orbit — systemgen supplies it once the mainworld is placed.
//
// Tr (Tropic) and Tu (Tundra) each need the world one orbit inside/outside the
// habitable zone AND a temperate-band UWP; Fr (Frozen) needs it two-plus orbits
// out AND a water-bearing UWP; Tz (Twilight Zone) is any world in orbit 0 or 1.
// This edition's Chart B does not emit the broader T5SS Hot/Cold pair.
func ClimateCodes(p uwp.Profile, orbit, hzOrbit int) []string {
	var out []string
	switch offset := orbit - hzOrbit; {
	case offset == -1 && tropicUWP(p):
		out = append(out, "Tr")
	case offset == 1 && tropicUWP(p):
		out = append(out, "Tu")
	case offset >= 2 && frozenUWP(p):
		out = append(out, "Fr")
	}
	if orbit == 0 || orbit == 1 {
		out = append(out, "Tz")
	}
	return out
}

// tropicUWP reports the temperate-band shape Tr/Tu share (Book 3 p.26): a
// mid-to-large world with a breathable-ish atmosphere and moderate hydrographics.
func tropicUWP(p uwp.Profile) bool {
	return allows("6789", p.Size) && allows("456789", p.Atmosphere) && allows("34567", p.Hydrographics)
}

// frozenUWP reports the Frozen shape (Book 3 p.26): any sized world that still
// carries water/ice.
func frozenUWP(p uwp.Profile) bool {
	return allows("23456789", p.Size) && allows("123456789A", p.Hydrographics)
}
