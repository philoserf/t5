package systemgen

import (
	"reflect"
	"testing"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/tradecode"
	"github.com/philoserf/t5/internal/uwp"
	"github.com/philoserf/t5/internal/worldgen"
)

// reginaSystem is a constructed system carrying Regina's mainworld record and a
// small stellar set, used to pin the String and SecondSurvey formats.
func reginaSystem() System {
	closeStar := Star{Type: "K", Decimal: 8, Size: "VI"}
	companion := Star{Type: "M", Decimal: 6, Size: "V"}

	return System{
		Primary:        Star{Type: "F", Decimal: 8, Size: "V"},
		Close:          &closeStar,
		CloseOrbit:     3,
		CloseCompanion: &companion,
		GasGiants:      3,
		Belts:          0,
		Worlds:         8,
		Mainworld: worldgen.World{
			Profile: uwp.Profile{
				Starport:      'A',
				Size:          7,
				Atmosphere:    8,
				Hydrographics: 8,
				Population:    8,
				Government:    9,
				Law:           9,
				TechLevel:     12,
			},
			TradeCodes: []tradecode.Code{"Ph", "Pa", "Ri"},
			Importance: 4,
			Economic: worldgen.Economic{
				Resources:      13,
				Labor:          7,
				Infrastructure: 14,
				Efficiency:     4,
			},
			Cultural: worldgen.Cultural{
				Heterogeneity: 9,
				Acceptance:    12,
				Strangeness:   6,
				Symbols:       13,
			},
			Nobility:        "BcCeF",
			NavalBase:       true,
			ScoutBase:       true,
			Zone:            'G',
			PopulationDigit: 7,
		},
	}
}

func TestSystemString(t *testing.T) {
	want := "Primary: F8 V\n" +
		"Close: K8 VI (Orbit 3)\n" +
		"Close Companion: M6 V\n" +
		"Worlds: 8  PBG: 703\n" +
		"Mainworld: A788899-C Ph Pa Ri {+4}(D7E+4)[9C6D] BcCeF NS -"
	if got := reginaSystem().String(); got != want {
		t.Fatalf("String() =\n%q\nwant\n%q", got, want)
	}
}

func TestSystemSecondSurvey(t *testing.T) {
	want := "1910 Regina A788899-C Ph Pa Ri {+4}(D7E+4)[9C6D] BcCeF NS - 703 8 Im F8 V K8 VI M6 V"
	if got := reginaSystem().SecondSurvey("1910", "Regina", ""); got != want {
		t.Fatalf("SecondSurvey() =\n%q\nwant\n%q", got, want)
	}
}

func TestOrbitFormulas(t *testing.T) {
	// Close = 1D-1 (0-5), Near = 5+1D (6-11), Far = 11+1D (12-17).
	if got := closeOrbit(dice.NewScripted(1)); got != 0 {
		t.Errorf("closeOrbit(1D=1) = %d, want 0", got)
	}

	if got := closeOrbit(dice.NewScripted(6)); got != 5 {
		t.Errorf("closeOrbit(1D=6) = %d, want 5", got)
	}

	if got := nearOrbit(dice.NewScripted(1)); got != 6 {
		t.Errorf("nearOrbit(1D=1) = %d, want 6", got)
	}

	if got := nearOrbit(dice.NewScripted(6)); got != 11 {
		t.Errorf("nearOrbit(1D=6) = %d, want 11", got)
	}

	if got := farOrbit(dice.NewScripted(1)); got != 12 {
		t.Errorf("farOrbit(1D=1) = %d, want 12", got)
	}
	// Regina's worked example: Far star in Orbit 11 + 1D(=5) = 16.
	if got := farOrbit(dice.NewScripted(5)); got != 16 {
		t.Errorf("farOrbit(1D=5) = %d, want 16 (Regina)", got)
	}
}

func TestGasGiants(t *testing.T) {
	// 2D/2 - 2, floored at 0.
	cases := map[int]int{2: 0, 4: 0, 8: 2, 12: 4} // 2D total -> gas giants
	for total, want := range cases {
		// Two dice summing to total (use total/2 twice, adjusting for odd).
		a := total / 2

		b := total - a
		if got := gasGiants(dice.NewScripted(a, b)); got != want {
			t.Errorf("gasGiants(2D=%d) = %d, want %d", total, got, want)
		}
	}
}

func TestBelts(t *testing.T) {
	cases := map[int]int{1: 0, 3: 0, 4: 1, 6: 3} // 1D -> belts
	for die, want := range cases {
		if got := belts(dice.NewScripted(die)); got != want {
			t.Errorf("belts(1D=%d) = %d, want %d", die, got, want)
		}
	}
}

// TestGenerateDeterministic confirms a seeded system is reproducible and
// internally consistent: the primary always exists and world/count invariants
// hold across many seeds.
func TestGenerateDeterministic(t *testing.T) {
	for seed := uint64(1); seed <= 200; seed++ {
		a := Generate(dice.NewWithSeed(seed))
		b := Generate(dice.NewWithSeed(seed))
		// Compare the generated values, not their formatting, so a future
		// String() tweak can't fail a determinism test.
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("seed %d not reproducible:\n%s\n---\n%s", seed, a, b)
		}

		if a.Primary.Type == "" {
			t.Fatalf("seed %d: primary has no spectral type", seed)
		}

		if a.GasGiants < 0 || a.GasGiants > 4 {
			t.Fatalf("seed %d: gas giants %d out of range 0-4", seed, a.GasGiants)
		}

		if a.Belts < 0 || a.Belts > 3 {
			t.Fatalf("seed %d: belts %d out of range 0-3", seed, a.Belts)
		}
		// Worlds = 1 + GG + Belts + 2D, so at least 1+GG+Belts+2.
		if lo := 1 + a.GasGiants + a.Belts + 2; a.Worlds < lo {
			t.Fatalf("seed %d: worlds %d below floor %d", seed, a.Worlds, lo)
		}
		// Present secondary stars sit in their orbit bands.
		if a.Close != nil && (a.CloseOrbit < 0 || a.CloseOrbit > 5) {
			t.Fatalf("seed %d: close orbit %d out of band 0-5", seed, a.CloseOrbit)
		}

		if a.Near != nil && (a.NearOrbit < 6 || a.NearOrbit > 11) {
			t.Fatalf("seed %d: near orbit %d out of band 6-11", seed, a.NearOrbit)
		}

		if a.Far != nil && (a.FarOrbit < 12 || a.FarOrbit > 17) {
			t.Fatalf("seed %d: far orbit %d out of band 12-17", seed, a.FarOrbit)
		}
		// A companion only exists when its star does.
		if a.CloseCompanion != nil && a.Close == nil {
			t.Fatalf("seed %d: close companion without a close star", seed)
		}

		if a.NearCompanion != nil && a.Near == nil {
			t.Fatalf("seed %d: near companion without a near star", seed)
		}

		if a.FarCompanion != nil && a.Far == nil {
			t.Fatalf("seed %d: far companion without a far star", seed)
		}
	}
}

// TestPrimaryNeverBDorOB documents that a primary's single Flux (-5..+5) can
// never reach the OB (Flux -6) or BD (Flux +6..+8) rows.
func TestPrimaryNeverBDorOB(t *testing.T) {
	for seed := uint64(1); seed <= 500; seed++ {
		p := Generate(dice.NewWithSeed(seed)).Primary
		if p.Type == "BD" || p.Type == "O" || p.Type == "B" {
			t.Fatalf("seed %d: primary is %s, which should be unreachable", seed, p.Type)
		}
	}
}

func TestClampGasGiants(t *testing.T) {
	cases := []struct {
		rolled int
		gg     ggConstraint
		want   int
	}{
		{3, ggAny, 3},
		{0, ggPresent, 1}, // a map symbol forces at least one
		{2, ggPresent, 2},
		{4, ggAbsent, 0}, // no symbol forces none
		{0, ggAbsent, 0},
	}
	for _, c := range cases {
		if got := clampGasGiants(c.rolled, c.gg); got != c.want {
			t.Errorf("clampGasGiants(%d, %v) = %d, want %d", c.rolled, c.gg, got, c.want)
		}
	}
}

func TestGenerateForMap(t *testing.T) {
	// The coarse map symbols govern generation across seeds: a gas-giant symbol
	// forces at least one (its absence none), and an asteroid symbol forces a
	// Size-0 belt mainworld.
	for seed := uint64(1); seed <= 25; seed++ {
		if s := GenerateForMap(dice.NewWithSeed(seed), true, false); s.GasGiants < 1 {
			t.Errorf("seed %d: gas giant present but count = %d", seed, s.GasGiants)
		}

		if s := GenerateForMap(dice.NewWithSeed(seed), false, false); s.GasGiants != 0 {
			t.Errorf("seed %d: no gas giant but count = %d", seed, s.GasGiants)
		}

		if s := GenerateForMap(dice.NewWithSeed(seed), false, true); s.Mainworld.Profile.Size != 0 {
			t.Errorf(
				"seed %d: asteroid symbol but mainworld size = %d",
				seed,
				s.Mainworld.Profile.Size,
			)
		}
	}
	// The unconstrained Generate is unchanged: same seed, same gas-giant count as
	// the internal any-constraint path.
	if a, b := Generate(
		dice.NewWithSeed(9),
	).GasGiants, generate(
		dice.NewWithSeed(9),
		ggAny,
		false,
	).GasGiants; a != b {
		t.Errorf("Generate diverged from ggAny: %d vs %d", a, b)
	}
}

// scriptCounter replays faces while counting the draws. It exists only because
// there is no way to count draws through dice.NewScripted, and it reproduces BOTH
// of that constructor's safety properties, not just the obvious one:
//
//   - it panics past the end of the script, so a case that outruns its faces fails
//     loudly rather than being served recycled ones;
//   - it validates every face is a real 1..6, so a fixture typo ("44" for "4") is
//     a loud failure rather than a silent impossible die.
//
// The second is easy to leave out, and leaving it out is what would make a green
// suite stop being evidence about the dice stream — the property CLAUDE.md names
// as the reason NewScripted is exact rather than cyclic.
func scriptCounter(t *testing.T, faces []int, drawn *int) *dice.Roller {
	t.Helper()

	for i, f := range faces {
		if f < 1 || f > 6 {
			t.Fatalf("scriptCounter: face %d at index %d is not a die face 1..6", f, i)
		}
	}

	return dice.NewSource(func() int {
		if *drawn >= len(faces) {
			panic("script exhausted")
		}

		f := faces[*drawn]
		*drawn++

		return f
	})
}

// TestGiantsDiceBudget locks the dice budget of the gas-giant segment, which is
// where GenerateForMap's alignment with Generate is either kept or lost (#321).
// The count roll is 2D/2-2 and each detailed giant is one 2D, so the budget is
// the whole claim: identical for every constraint over the same faces, save the
// documented ggPresent-over-a-rolled-0 exception, which is exactly one 2D more.
func TestGiantsDiceBudget(t *testing.T) {
	// 5,5 -> 2D=10 -> 10/2-2 = 3 giants, then three 2D sizes (4,4 -> 8 -> eHex
	// 20+8-1 = 27, a Large Gas Giant). 2 + 6 = 8 dice.
	rolledThree := []int{5, 5, 4, 4, 4, 4, 4, 4}
	// 2,2 -> 2D=4 -> max(4/2-2, 0) = 0 giants, so no size rolls follow: 2 dice.
	rolledZero := []int{2, 2}
	// The exception's extra giant needs a 2D of its own beyond those 2.
	rolledZeroPlusExtra := []int{2, 2, 4, 4}

	cases := []struct {
		name      string
		faces     []int
		gg        ggConstraint
		wantCount int
		wantDice  int
	}{
		{"any over a rolled 3", rolledThree, ggAny, 3, 8},
		// The fix: the giants are rolled and detailed, then discarded, so the
		// budget matches ggAny's rather than collapsing to the count roll alone.
		{"absent over a rolled 3", rolledThree, ggAbsent, 0, 8},
		{"present over a rolled 3", rolledThree, ggPresent, 3, 8},
		{"any over a rolled 0", rolledZero, ggAny, 0, 2},
		{"absent over a rolled 0", rolledZero, ggAbsent, 0, 2},
		// The one unalignable case: a giant with no roll to re-derive from.
		{"present over a rolled 0", rolledZeroPlusExtra, ggPresent, 1, 4},
	}

	for _, c := range cases {
		drawn := 0

		count, detailed := giantsFor(scriptCounter(t, c.faces, &drawn), c.gg)
		if count != c.wantCount {
			t.Errorf("%s: count = %d, want %d", c.name, count, c.wantCount)
		}

		if len(detailed) != c.wantCount {
			t.Errorf("%s: detailed %d giants, want %d", c.name, len(detailed), c.wantCount)
		}

		if drawn != c.wantDice {
			t.Errorf("%s: drew %d dice, want %d", c.name, drawn, c.wantDice)
		}
	}
}

// TestGenerateForMapDiceAlignment is the falsifiable form of GenerateForMap's
// alignment claim (#321). Detailing the rolled gas-giant count rather than the
// constrained one keeps the constrained stream aligned with Generate's, so a
// constraint that agrees with the roll yields the very same system, and one that
// disagrees still re-derives the mainworld's UWP and the stellar family from the
// same faces. The single documented exception — ggPresent over a rolled 0 — is
// reached here as well; its size is pinned by TestGiantsDiceBudget, which can
// see the segment in isolation.
func TestGenerateForMapDiceAlignment(t *testing.T) {
	var agreeAbsent, agreePresent, disagreeAbsent, exception int

	for seed := uint64(1); seed <= 300; seed++ {
		want := Generate(dice.NewWithSeed(seed))
		rolled := want.GasGiants

		if rolled == 0 {
			// ggAbsent agrees with a rolled 0: identical dice, identical system.
			if got := GenerateForMap(dice.NewWithSeed(seed), false, false); !reflect.DeepEqual(
				got, want,
			) {
				t.Fatalf("seed %d: ggAbsent agreed with a rolled 0 but diverged:\n%s\n---\n%s",
					seed, got, want)
			}

			agreeAbsent++

			// ggPresent over a rolled 0 is the documented exception: the constraint
			// is honoured with a giant the unconstrained stream never rolled. The
			// size of the divergence — exactly one extra 2D — is not assertable
			// from here, because that extra giant also gives the orbit map and the
			// satellite pass another body to roll for; it is locked at the segment
			// itself by TestGiantsDiceBudget.
			if got := GenerateForMap(dice.NewWithSeed(seed), true, false); got.GasGiants != 1 {
				t.Fatalf("seed %d: ggPresent over a rolled 0 gave %d giants, want 1",
					seed, got.GasGiants)
			}

			exception++

			continue
		}

		// ggPresent agrees with any rolled count of 1 or more: identical system.
		if got := GenerateForMap(dice.NewWithSeed(seed), true, false); !reflect.DeepEqual(
			got, want,
		) {
			t.Fatalf("seed %d: ggPresent agreed with a rolled %d but diverged:\n%s\n---\n%s",
				seed, rolled, got, want)
		}

		agreePresent++

		// ggAbsent disagreeing with a rolled 1+ is the case the fix addresses. The
		// count is a genuine input downstream, so the systems are not equal; what
		// must hold is that the stream did not shift, which shows up as the
		// mainworld's UWP and the whole stellar family re-deriving unchanged.
		got := GenerateForMap(dice.NewWithSeed(seed), false, false)
		if got.GasGiants != 0 {
			t.Fatalf("seed %d: ggAbsent gave %d giants, want 0", seed, got.GasGiants)
		}

		if got.Mainworld.Profile != want.Mainworld.Profile {
			t.Fatalf(
				"seed %d: ggAbsent dropped %d rolled giants and the mainworld UWP moved, "+
					"%s vs %s — the discarded giants' 2D sizes were not drawn",
				seed, rolled, got.Mainworld.Profile, want.Mainworld.Profile,
			)
		}

		if !reflect.DeepEqual(got.Stars(), want.Stars()) {
			t.Fatalf(
				"seed %d: ggAbsent dropped %d rolled giants and the stellar family moved, "+
					"%s vs %s — the discarded giants' 2D sizes were not drawn",
				seed, rolled, got.Stellar(), want.Stellar(),
			)
		}

		disagreeAbsent++
	}

	// Each branch above is only evidence if seeds actually reached it.
	// agreeAbsent and exception count the SAME seeds — both increment
	// unconditionally in the rolled-0 branch, one per assertion made there — so
	// listing both in the guard below reads like four populations when there are
	// three. Asserted explicitly rather than left to look like coverage it is not.
	if agreeAbsent != exception {
		t.Fatalf("rolled-0 counters disagree: agreeAbsent=%d exception=%d — "+
			"they count the same seeds and must move together",
			agreeAbsent, exception)
	}

	if agreeAbsent == 0 || agreePresent == 0 || disagreeAbsent == 0 {
		t.Fatalf(
			"alignment cases not all exercised: rolled0=%d agreePresent=%d disagreeAbsent=%d",
			agreeAbsent, agreePresent, disagreeAbsent,
		)
	}
}

// TestStars locks the single ordered enumeration of the stellar family that
// String, Stellar, and the survey sheet all consume.
func TestStars(t *testing.T) {
	slots := reginaSystem().Stars()
	if len(slots) != 3 {
		t.Fatalf("got %d star slots, want 3 (primary, close, close companion)", len(slots))
	}

	if slots[0].Label != "Primary" || slots[0].Orbit != -1 || slots[0].Companion {
		t.Errorf("slot 0 = %+v, want the Primary at no orbit", slots[0])
	}

	if slots[1].Label != "Close" || slots[1].Orbit != 3 || slots[1].Companion {
		t.Errorf("slot 1 = %+v, want Close at Orbit 3", slots[1])
	}

	if slots[2].Label != "Close Companion" || !slots[2].Companion || slots[2].Orbit != -1 {
		t.Errorf("slot 2 = %+v, want a companion at no orbit", slots[2])
	}
}

// TestSecondaryStarOrbitsAreReserved: a world may not be placed in the orbit a
// secondary star occupies around the primary (Book 3 p.21).
func TestSecondaryStarOrbitsAreReserved(t *testing.T) {
	for seed := uint64(1); seed <= 300; seed++ {
		s := Generate(dice.NewWithSeed(seed))
		taken := map[int]bool{}

		for _, sl := range s.Stars() {
			if !sl.Companion && sl.Orbit >= 0 {
				taken[sl.Orbit] = true
			}
		}

		for _, o := range s.Orbits {
			if o.Host == "Primary" && taken[o.Orbit] {
				t.Fatalf(
					"seed %d: %s sits in orbit %d, which a secondary star occupies",
					seed,
					o.Kind,
					o.Orbit,
				)
			}
		}
	}
}
