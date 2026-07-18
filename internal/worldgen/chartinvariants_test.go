package worldgen

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/uwp"
)

// allOtherWorldTypes is every non-mainworld type. The tripwire below fails when
// a type is added without extending this list, so the invariant tests here
// cannot silently stop covering the full set.
var allOtherWorldTypes = []OtherWorldType{
	Hospitable, InnerWorld, BigWorld, StormWorld,
	RadWorld, Inferno, Worldlet, Iceworld, Planetoids,
}

// TestAllOtherWorldTypesCovered guards the list above: Planetoids is the last
// declared type, so a new one lands past it and trips this.
func TestAllOtherWorldTypesCovered(t *testing.T) {
	if got, want := len(allOtherWorldTypes), int(Planetoids)+1; got != want {
		t.Fatalf("allOtherWorldTypes has %d entries, want %d — a world type was "+
			"added; extend the list so the chart-invariant tests cover it", got, want)
	}
}

// TestChartProfileNormalizesAnyProfile pins the funnel itself. chartProfile is
// the single return path of GenerateSatelliteWorld, so whatever a case body
// assembles — including a future type that never heard of the chart's Size
// rules — is corrected here. Feeding it deliberately out-of-chart profiles is
// the direct test that bypass is not available.
func TestChartProfileNormalizesAnyProfile(t *testing.T) {
	cases := []struct {
		name             string
		size, atm, hyd   int
		wantAtm, wantHyd int
	}{
		// The #210 shape: a Size-0 body with an atmosphere and oceans.
		{"size 0 keeps neither air nor water", 0, 8, 8, 0, 0},
		{"size 1 keeps air but no water", 1, 8, 8, 8, 0},
		{"size 2 keeps both", 2, 8, 8, 8, 8},
		// Out-of-range values are floored and capped, not passed through.
		{"negative values floor at 0", 5, -6, -6, 0, 0},
		{"oversized values clamp to F and A", 5, 99, 99, maxAtmosphere, maxHydrographics},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := chartProfile(uwp.Profile{
				Size:          c.size,
				Atmosphere:    c.atm,
				Hydrographics: c.hyd,
			})
			if got.Atmosphere != c.wantAtm {
				t.Errorf("Atmosphere = %d, want %d", got.Atmosphere, c.wantAtm)
			}

			if got.Hydrographics != c.wantHyd {
				t.Errorf("Hydrographics = %d, want %d", got.Hydrographics, c.wantHyd)
			}
		})
	}
}

// TestGeneratedWorldsObeyChartRules is the end-to-end property: across every
// type, every mainworld population, and every satellite size cap, no generated
// profile violates the World Creation chart's structural rules (Book 3 p.24) or
// the satellite size rule (p.21).
//
// This is the test a new world type gets for free. Because Size is rolled and
// capped above the switch and every profile returns through chartProfile, a new
// case body cannot break these invariants — but if someone reroutes around the
// pipeline, this catches it for all nine types at once.
func TestGeneratedWorldsObeyChartRules(t *testing.T) {
	caps := []int{NoSizeCap, 0, 1, 2, 5, 9, 15}

	for _, wt := range allOtherWorldTypes {
		for seed := range uint64(300) {
			for mwPop := range 12 {
				for _, maxSize := range caps {
					p := GenerateSatelliteWorld(dice.NewWithSeed(seed), wt, mwPop, maxSize)

					ctx := func() string {
						return wt.String() + " seed/pop/cap profile " + p.String()
					}

					// Book 3 p.24: "If Atm<0 or Siz=0, Atm=0".
					if p.Size == 0 && p.Atmosphere != 0 {
						t.Fatalf("%s: Size 0 with Atmosphere %d, want 0", ctx(), p.Atmosphere)
					}

					// Book 3 p.24: "If Siz <2, Hyd =0".
					if p.Size < 2 && p.Hydrographics != 0 {
						t.Fatalf("%s: Size %d with Hydrographics %d, want 0",
							ctx(), p.Size, p.Hydrographics)
					}

					// Book 3 p.21: a satellite is never larger than its parent.
					if maxSize != NoSizeCap && p.Size > maxSize {
						t.Fatalf("%s: Size %d exceeds parent cap %d", ctx(), p.Size, maxSize)
					}

					// The characteristics stay inside their eHex ranges.
					if p.Atmosphere < 0 || p.Atmosphere > maxAtmosphere {
						t.Fatalf("%s: Atmosphere %d out of range", ctx(), p.Atmosphere)
					}

					if p.Hydrographics < 0 || p.Hydrographics > maxHydrographics {
						t.Fatalf("%s: Hydrographics %d out of range", ctx(), p.Hydrographics)
					}
				}
			}
		}
	}
}

// TestPlanetoidsSpendNoSizeDie pins the one place hoisting Size could have
// shifted the dice stream: a belt has no Size formula, so typeSize must return 0
// without rolling. If it ever spends a die, every later roll for a Planetoids
// world shifts and this diverges.
func TestPlanetoidsSpendNoSizeDie(t *testing.T) {
	var spent int

	r := dice.NewSource(func() int {
		spent++

		return 4
	})

	if got := typeSize(r, Planetoids); got != 0 {
		t.Errorf("typeSize(Planetoids) = %d, want 0", got)
	}

	if spent != 0 {
		t.Errorf("typeSize(Planetoids) spent %d dice, want 0 — the dice stream has shifted", spent)
	}
}
