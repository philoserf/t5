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

// placeMainworld rolls the orbit the mainworld wants around the primary and its
// satellite record (Book 3 p.24). It rolls, in order: the HZ-variance Flux (DM +2
// for an M primary, −2 for O/B, Chart B) and the mainworld-type Flux (Chart C,
// plus a satellite orbit-letter Flux when it is a satellite). A Far/Close
// satellite mainworld earns Sa/Lk. It returns the wanted orbit (−1 when the
// primary has no habitable zone, so the mainworld cannot be placed) and the
// satellite record.
//
// The climate codes are deliberately NOT tagged here. The wanted orbit is not yet
// the orbit the mainworld gets: placeOrbits claims it, and claim may nudge it off
// a reserved secondary-star orbit or clamp it to the star's precluded floor.
// Keying the codes to the pre-claim orbit left the one-line record describing a
// different orbit than the orbit map showed, so tagMainworldClimate runs on the
// claimed orbit instead.
func placeMainworld(
	r *dice.Roller,
	primary Star,
	mainworld *worldgen.World,
) (int, MainworldSatellite) {
	// An asteroid-belt mainworld (Size 0) is placed using the Belt column of the
	// P2 chart as an absolute orbit, without regard to the habitable zone (Book 3
	// p.21): no HZ variance, no gas-giant-satellite roll, and no climate codes —
	// and it is placed even when the primary has no habitable zone.
	//
	// The Belt column is an offset table, so its low rows are negative (−1 at 2D=2,
	// the only negative row a 2D roll reaches). An orbit is never negative, and −1
	// is this function's "no orbit" sentinel, so emitting the raw offset dropped the
	// belt mainworld from the orbit map altogether. Floor it at the innermost orbit
	// the primary allows — the clamp claim would apply anyway — which keeps a belt
	// inside the orbit domain and clear of the sentinel.
	if mainworld.Belt {
		return max(p2(r.Dice(2)).belt, firstOrbit(primary)), MainworldSatellite{}
	}

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

	return max(hz+hzVar, 0), sat
}

// tagMainworldClimate appends the climate / orbit-dependent trade codes the
// mainworld earns from the concrete orbit it was finally placed in (Book 3 Chart
// D, p.26). It runs after placeOrbits claims that orbit, so the mainworld's codes,
// s.MainworldOrbit, and the orbit map always describe one orbit rather than three.
// An asteroid-belt mainworld, placed without regard to the habitable zone, takes
// none.
func tagMainworldClimate(mainworld *worldgen.World, orbit, hz int, hasHZ bool) {
	if mainworld.Belt {
		return
	}

	if codes := worldgen.ClimateCodes(mainworld.Profile, orbit, hz, hasHZ); len(codes) > 0 {
		mainworld.TradeCodes = append(mainworld.TradeCodes, codes...)
	}
}

// satellite orbit-letter names by Flux −6..+6 (Book 3 p.24 Chart C).
var (
	closeOrbitLetters = [13]string{
		"Ay",
		"Bee",
		"Cee",
		"Dee",
		"Ee",
		"Eff",
		"Gee",
		"Aitch",
		"Eye",
		"Jay",
		"Kay",
		"Ell",
		"Em",
	}
	farOrbitLetters = [13]string{
		"En",
		"Oh",
		"Pee",
		"Que",
		"Arr",
		"Ess",
		"Tee",
		"Yu",
		"Vee",
		"Dub",
		"Ex",
		"Wye",
		"Zee",
	}
)

// rollMainworldSatellite rolls the mainworld type (Book 3 p.24 Chart 2C): a
// Far Satellite (Flux −5/−4), a Close Satellite (−3), or a Planet (everything
// else). Chart 2C is a single row per Flux value — the same Flux that names the
// type names the orbit letter in that row's Close or Far column — so it is one
// roll, not two. Rolling the letter separately (as this once did) both consumed an
// extra die, shifting the whole system's stream, and paired a satellite with a
// letter from an unrelated row.
func rollMainworldSatellite(r *dice.Roller) MainworldSatellite {
	flux := r.Flux()
	switch flux {
	case -5, -4:
		return MainworldSatellite{
			IsSatellite: true,
			Far:         true,
			OrbitLetter: farOrbitLetters[dice.FluxIndex(flux)],
		}
	case -3:
		return MainworldSatellite{
			IsSatellite: true,
			OrbitLetter: closeOrbitLetters[dice.FluxIndex(flux)],
		}
	default:
		return MainworldSatellite{}
	}
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
