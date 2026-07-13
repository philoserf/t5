package systemgen

import (
	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// placeMainworld gives the mainworld a concrete orbit around the primary and
// tags its climate trade codes (Book 3 p.24 Chart B "2B Mainworld Orbit"). It
// rolls the HZ-variance Flux (DM +2 for an M primary, −2 for O/B) once, so the
// dice stream is stable whether or not the primary has a habitable zone. It
// returns the mainworld orbit (−1 when the primary has no HZ) and appends any
// climate codes (Tr/Tu/Fr/Tz) to the mainworld's trade codes.
func placeMainworld(r *dice.Roller, primary Star, mainworld *worldgen.World) int {
	hzVar := hzVarFromFlux(r.Flux() + hzVarDM(primary))
	hz, ok := HZOrbit(primary)
	if !ok {
		return -1
	}
	orbit := max(hz+hzVar, 0)
	if codes := worldgen.ClimateCodes(mainworld.Profile, orbit, hz); len(codes) > 0 {
		mainworld.TradeCodes = append(mainworld.TradeCodes, codes...)
	}
	return orbit
}

// hzVarFromFlux maps the HZ-variance Flux to an orbit offset (Book 3 p.24 Chart
// B): −6 → −2 (inferno), −5..−3 → −1 (hot/tropic), −2..+2 → 0 (temperate),
// +3..+5 → +1 (cold/tundra), +6 → +2 (frozen). DM-adjusted Flux beyond ±6 clamps.
func hzVarFromFlux(flux int) int {
	switch {
	case flux <= -6:
		return -2
	case flux <= -3:
		return -1
	case flux <= 2:
		return 0
	case flux <= 5:
		return 1
	default:
		return 2
	}
}

// hzVarDM is the placement Flux modifier (Book 3 p.24): +2 for an M primary,
// −2 for an O or B primary, 0 otherwise.
func hzVarDM(primary Star) int {
	switch primary.Type {
	case "M":
		return 2
	case "O", "B":
		return -2
	}
	return 0
}
