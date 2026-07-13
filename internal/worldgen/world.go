package worldgen

import (
	"fmt"
	"slices"
	"strings"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/uwp"
)

// A World is a fully-detailed mainworld: its Universal World Profile plus the
// derived data from Book 3 Charts C-F — trade classifications, the {Ix}(Ex)[Cx]
// Extensions, nobility, bases, travel zone, native status, and the population
// multiplier digit.
type World struct {
	Profile         uwp.Profile
	TradeCodes      []string
	Importance      int
	Economic        Economic
	Cultural        Cultural
	Nobility        string
	NavalBase       bool
	ScoutBase       bool
	Zone            byte
	NativeStatus    string
	PopulationDigit int
}

// PopulationDigit returns the population multiplier digit (Book 3 PBG): an
// even-distribution 1-9 value for an inhabited world, else 0.
func PopulationDigit(r *dice.Roller, population int) int {
	if population == 0 {
		return 0
	}
	return r.EvenDist1to9()
}

// GenerateWorld rolls a complete mainworld: the UWP and every derived attribute.
// gasGiants and belts come from the enclosing system (they feed the Economic
// Extension and the PBG); isCapital marks a subsector or sector capital, which
// affects nobility. Bases are rolled before Importance because Importance
// depends on them.
func GenerateWorld(r *dice.Roller, gasGiants, belts int, isCapital bool) World {
	p := Generate(r)
	tcs := TradeClassifications(p)
	naval, scout := RollBases(r, p.Starport)
	ix := Importance(p, tcs, naval, scout, false)
	return World{
		Profile:         p,
		TradeCodes:      tcs,
		Importance:      ix,
		Economic:        RollEconomic(r, p, ix, gasGiants, belts),
		Cultural:        RollCultural(r, p, ix),
		Nobility:        Nobility(tcs, ix, isCapital),
		NavalBase:       naval,
		ScoutBase:       scout,
		Zone:            TravelZone(p),
		NativeStatus:    NativeStatus(p),
		PopulationDigit: PopulationDigit(r, p.Population),
	}
}

// SetCapital marks the world a capital carrying the given capital trade code —
// Cs (subsector), Cx (sector), or Cp (Imperial) — and recomputes its nobility as
// a capital (Book 3 Chart D p.26). Which world is the capital is a whole-region
// decision the caller makes; this only encodes the result on the world.
func (w *World) SetCapital(code string) {
	if !slices.Contains(w.TradeCodes, code) {
		w.TradeCodes = append(w.TradeCodes, code)
	}
	w.Nobility = Nobility(w.TradeCodes, w.Importance, true)
}

// SecondSurvey renders the world portion of the Second Survey record:
//
//	UWP  TCs  {Ix}(Ex)[Cx]  Nobility  Bases  Zone
//
// e.g. "A788899-C Ph Pa Ri {+4}(D7E+4)[9C6D] BcCeF NS -". A dash marks an empty
// bases or (Green) zone field. The system-level fields (hex, PBG belts/giants,
// world count, stellar data) are appended by the caller.
func (w World) SecondSurvey() string {
	fields := []string{
		w.Profile.String(),
		strings.Join(w.TradeCodes, " "),
		importance(w.Importance) + w.Economic.String() + w.Cultural.String(),
		w.Nobility,
		dashIfEmpty(w.bases()),
		dashIfEmpty(w.zone()),
	}
	return strings.Join(slices.DeleteFunc(fields, isEmpty), " ")
}

func isEmpty(s string) bool { return s == "" }

// importance renders the Importance Extension {Ix}. It is signed when non-zero
// but bare at zero, matching the book's Importance table (0, +4, -2).
func importance(ix int) string {
	if ix == 0 {
		return "{0}"
	}
	return fmt.Sprintf("{%+d}", ix)
}

// bases renders the base codes: "N" for a Naval base, "S" for a Scout base.
func (w World) bases() string {
	var b strings.Builder
	if w.NavalBase {
		b.WriteByte('N')
	}
	if w.ScoutBase {
		b.WriteByte('S')
	}
	return b.String()
}

// zone renders the travel zone, leaving Green blank (the common case).
func (w World) zone() string {
	if w.Zone == 'A' || w.Zone == 'R' {
		return string(w.Zone)
	}
	return ""
}

func dashIfEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
