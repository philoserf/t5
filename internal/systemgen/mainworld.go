package systemgen

import (
	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

// MainworldSatellite records whether the mainworld is a satellite of a gas
// giant and, if so, whether it is Far or Close (tidally locked) and its orbit
// letter (Book 3 p.24 Chart C "2C Satellite?").
type MainworldSatellite struct {
	IsSatellite bool
	Far         bool   // Far satellite; else Close (tidally locked to the primary)
	OrbitLetter string // the satellite's orbit name, e.g. "Arr"
}

// placeMainworld gives the mainworld a concrete orbit around the primary and
// tags its trade codes (Book 3 p.24). It rolls, in order: the HZ-variance Flux
// (DM +2 for an M primary, −2 for O/B, Chart B) and the mainworld-type Flux
// (Chart C, plus a satellite orbit-letter Flux when it is a satellite). A Far/
// Close satellite mainworld earns Sa/Lk, and its climate codes (Tr/Tu/Fr/Tz)
// are appended. It returns the mainworld orbit (−1 when the primary has no HZ)
// and the satellite record.
func placeMainworld(r *dice.Roller, primary Star, mainworld *worldgen.World) (int, MainworldSatellite) {
	hzVar := hzVarFromFlux(r.Flux() + hzVarDM(primary))
	sat := rollMainworldSatellite(r)
	if sat.IsSatellite {
		if sat.Far {
			mainworld.TradeCodes = append(mainworld.TradeCodes, "Sa")
		} else {
			mainworld.TradeCodes = append(mainworld.TradeCodes, "Lk")
		}
	}
	hz, ok := HZOrbit(primary)
	if !ok {
		return -1, sat
	}
	orbit := max(hz+hzVar, 0)
	if codes := worldgen.ClimateCodes(mainworld.Profile, orbit, hz); len(codes) > 0 {
		mainworld.TradeCodes = append(mainworld.TradeCodes, codes...)
	}
	return orbit, sat
}

// satellite orbit-letter names by Flux −6..+6 (Book 3 p.24 Chart C).
var (
	closeOrbitLetters = [13]string{"Ay", "Bee", "Cee", "Dee", "Ee", "Eff", "Gee", "Aitch", "Eye", "Jay", "Kay", "Ell", "Em"}
	farOrbitLetters   = [13]string{"En", "Oh", "Pee", "Que", "Arr", "Ess", "Tee", "Yu", "Vee", "Dub", "Ex", "Wye", "Zee"}
)

// rollMainworldSatellite rolls the mainworld type (Book 3 p.24 Chart C): a
// Far Satellite (type Flux −5/−4), a Close Satellite (−3), or a Planet
// (everything else). A satellite additionally rolls its orbit letter.
func rollMainworldSatellite(r *dice.Roller) MainworldSatellite {
	switch r.Flux() {
	case -5, -4:
		return MainworldSatellite{IsSatellite: true, Far: true, OrbitLetter: farOrbitLetters[fluxIndex(r.Flux())]}
	case -3:
		return MainworldSatellite{IsSatellite: true, OrbitLetter: closeOrbitLetters[fluxIndex(r.Flux())]}
	default:
		return MainworldSatellite{}
	}
}

// fluxIndex maps a Flux value (−6..+6) to a 0..12 table index, clamping.
func fluxIndex(flux int) int {
	return min(max(flux+6, 0), 12)
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
