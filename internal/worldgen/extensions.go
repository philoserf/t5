package worldgen

import (
	"fmt"
	"slices"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/ehex"
	"github.com/philoserf/t5/internal/uwp"
)

// The Extensions summarize a world beyond its UWP, in the notation
// {Ix}(Ex)[Cx] (Book 3 pp. 18, 27, Chart E). Importance feeds both the others,
// so compute it first.

// importanceTradeCodes each add +1 to Importance when present (Chart E, B3 p.27:
// "Per Ag Hi In Ri +1" — Pre-Agricultural is not on the list).
var importanceTradeCodes = []string{"Ag", "Hi", "In", "Ri"}

// Importance returns the Importance Extension (Ix), a signed rank of a world
// within its region. It is a pure additive sum over the UWP, trade codes, and
// bases; "Important" is +4 or greater. tcs is the world's trade-classification
// codes (see TradeClassifications).
func Importance(p uwp.Profile, tcs []string, navalBase, scoutBase, wayStation bool) int {
	ix := 0
	switch p.Starport {
	case 'A', 'B':
		ix++
	case 'D', 'E', 'X':
		ix--
	}
	if p.TechLevel >= 16 { // G
		ix++
	}
	if p.TechLevel >= 10 { // A
		ix++
	}
	if p.TechLevel <= 8 {
		ix--
	}
	for _, code := range importanceTradeCodes {
		if slices.Contains(tcs, code) {
			ix++
		}
	}
	if p.Population <= 6 {
		ix--
	}
	if navalBase && scoutBase {
		ix++
	}
	if wayStation {
		ix++
	}
	return ix
}

// An Economic is the Economic Extension (Ex): the strength of a world's
// economy, rendered as "(RLI±E)" — Resources, Labor, and Infrastructure as eHex
// digits, then a signed Efficiency.
type Economic struct {
	Resources      int
	Labor          int
	Infrastructure int
	Efficiency     int
}

// RollEconomic rolls the Economic Extension. Resources = 2D, plus gas giants and
// belts when TL is 8 or more; Labor = Pop-1; Infrastructure depends on the
// population band and Importance; Efficiency = Flux. Resources, Labor, and
// Infrastructure are floored at 0; Efficiency may be negative.
func RollEconomic(r *dice.Roller, p uwp.Profile, ix, gasGiants, belts int) Economic {
	resources := r.Dice(2)
	if p.TechLevel >= 8 {
		resources += gasGiants + belts
	}

	var infrastructure int
	switch {
	case p.Population == 0:
		infrastructure = 0
	case p.Population <= 3:
		infrastructure = ix
	case p.Population <= 6:
		infrastructure = r.Die() + ix
	default:
		infrastructure = r.Dice(2) + ix
	}

	return Economic{
		Resources:      max(resources, 0),
		Labor:          max(p.Population-1, 0),
		Infrastructure: max(infrastructure, 0),
		Efficiency:     r.Flux(),
	}
}

// RU is the Resource Units, the product of the four factors. Any factor of 0 is
// treated as 1 ("If any value = 0, use 1 (to avoid multiplying by zero)", B3
// p.27); Efficiency may be negative, so RU may be negative.
func (e Economic) RU() int {
	return nonZeroFactor(e.Resources) * nonZeroFactor(e.Labor) *
		nonZeroFactor(e.Infrastructure) * nonZeroFactor(e.Efficiency)
}

// nonZeroFactor returns v, substituting 1 for 0 (the RU zero rule, B3 p.27).
func nonZeroFactor(v int) int {
	if v == 0 {
		return 1
	}
	return v
}

// String renders the extension as "(RLI±E)", e.g. "(D7E+4)".
func (e Economic) String() string {
	return fmt.Sprintf("(%s%s%s%+d)",
		ehex.Format(e.Resources), ehex.Format(e.Labor), ehex.Format(e.Infrastructure), e.Efficiency)
}

// A Cultural is the Cultural Extension (Cx), rendered as "[HASS]" —
// Heterogeneity, Acceptance, Strangeness, and Symbols, each an eHex digit.
type Cultural struct {
	Heterogeneity int
	Acceptance    int
	Strangeness   int
	Symbols       int
}

// RollCultural rolls the Cultural Extension: Heterogeneity = Pop+Flux,
// Acceptance = Pop+Ix, Strangeness = Flux+5, Symbols = Flux+TL. Every value is
// floored at 1, except that a population-0 world has all values 0.
func RollCultural(r *dice.Roller, p uwp.Profile, ix int) Cultural {
	if p.Population == 0 {
		return Cultural{}
	}
	return Cultural{
		Heterogeneity: max(p.Population+r.Flux(), 1),
		Acceptance:    max(p.Population+ix, 1),
		Strangeness:   max(r.Flux()+5, 1),
		Symbols:       max(r.Flux()+p.TechLevel, 1),
	}
}

// String renders the extension as "[HASS]", e.g. "[9C6D]".
func (c Cultural) String() string {
	return fmt.Sprintf("[%s%s%s%s]",
		ehex.Format(c.Heterogeneity), ehex.Format(c.Acceptance), ehex.Format(c.Strangeness), ehex.Format(c.Symbols))
}
