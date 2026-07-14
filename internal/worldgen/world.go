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
	NavalDepot      bool
	WayStation      bool
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
	return generateWorld(r, gasGiants, belts, isCapital, false)
}

// GenerateBeltWorld is GenerateWorld for an asteroid-belt mainworld (Book 3
// p.13): the UWP is a belt (Size 0), and every derived attribute follows from
// it. Used when the sector map's asteroid symbol fixes the mainworld.
func GenerateBeltWorld(r *dice.Roller, gasGiants, belts int, isCapital bool) World {
	return generateWorld(r, gasGiants, belts, isCapital, true)
}

func generateWorld(r *dice.Roller, gasGiants, belts int, isCapital, belt bool) World {
	p := generate(r, belt)
	tcs := append(TradeClassifications(p), ZoneCodes(p)...)
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

// SetNavalDepot places a Naval Depot on the world (Book 3 p.28), a system-
// encompassing naval fleet base. Unlike a Way Station it confers no Importance
// bonus — Chart E's Ix table does not list one — so it only sets the base. Depots
// are sited by the region survey (1 per 1000 worlds, Starport A only), not rolled
// per world, so this is a setter rather than part of RollBases.
func (w *World) SetNavalDepot() {
	w.NavalDepot = true
}

// SetWayStation places a Scout Way Station on the world (Book 3 p.28), a
// communications relay sited on a trade route. It sets the base and applies the
// Importance bonus the station confers (Chart E: "+1 If Way Station"),
// recomputing nobility from the new Importance. The already-rolled Economic and
// Cultural extensions are left unchanged (the +1 is not cascaded), mirroring
// SetCapital.
func (w *World) SetWayStation() {
	if w.WayStation {
		return
	}
	w.WayStation = true
	w.Importance = Importance(w.Profile, w.TradeCodes, w.NavalBase, w.ScoutBase, true)
	w.Nobility = Nobility(w.TradeCodes, w.Importance, hasCapitalCode(w.TradeCodes))
}

// hasCapitalCode reports whether the trade codes already mark the world a
// capital (Cs subsector, Cx sector, or Cp Imperial).
func hasCapitalCode(tcs []string) bool {
	return slices.Contains(tcs, "Cs") || slices.Contains(tcs, "Cx") || slices.Contains(tcs, "Cp")
}

// SecondSurvey renders the world portion of the Second Survey record:
//
//	UWP  TCs  {Ix}(Ex)[Cx]  Nobility  Bases  Zone
//
// e.g. "A788899-C Ph Pa Ri {+4}(D7E+4)[9C6D] BcCeF NS -". A dash marks an empty
// trade-code, bases, or (Green) zone field. The system-level fields (hex, PBG
// belts/giants, world count, stellar data) are appended by the caller.
//
// Every field is dashed rather than dropped, because this is a POSITIONAL record: a
// world matching no trade code at all (C539700-8 is one — its atmosphere, hydro-
// graphics, and population between them exclude every rule) used to render with the
// TC column simply gone, silently shifting the extensions, nobility, bases, and
// zone one column to the left. A reader, or a parser, would have read its
// extensions as its trade codes.
func (w World) SecondSurvey() string {
	fields := []string{
		w.Profile.String(),
		dashIfEmpty(strings.Join(w.TradeCodes, " ")),
		w.Extensions(),
		w.Nobility,
		dashIfEmpty(w.bases()),
		dashIfEmpty(w.zone()),
	}
	return strings.Join(fields, " ")
}

// Extensions renders the world's three extensions as they appear in the Second
// Survey record: "{Ix}(Ex)[Cx]", e.g. "{+4}(D7E+4)[9C6D]".
func (w World) Extensions() string {
	return importance(w.Importance) + w.Economic.String() + w.Cultural.String()
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

// BaseNames names the bases the world hosts, in the order their codes appear in
// the Second Survey record (Book 3 p.28). Empty when the world hosts none.
func (w World) BaseNames() []string {
	var out []string
	for _, b := range w.baseList() {
		out = append(out, b.name)
	}
	return out
}

// baseList is the world's bases as (record code, display name) pairs, in Chart F
// order (Book 3 p.28). The code is not the name's initial: a Naval Depot's code is
// "D", not "N", and it would otherwise collide with the Naval Base.
func (w World) baseList() []struct{ code, name string } {
	var out []struct{ code, name string }
	if w.NavalBase {
		out = append(out, struct{ code, name string }{"N", "Naval"})
	}
	if w.ScoutBase {
		out = append(out, struct{ code, name string }{"S", "Scout"})
	}
	if w.NavalDepot {
		out = append(out, struct{ code, name string }{"D", "Naval Depot"})
	}
	if w.WayStation {
		out = append(out, struct{ code, name string }{"W", "Way Station"})
	}
	return out
}

// bases renders the base codes: N Naval, S Scout, D Naval Depot, W Way Station
// (Book 3 p.28), in Chart F order.
func (w World) bases() string {
	var b strings.Builder
	for _, base := range w.baseList() {
		b.WriteString(base.code)
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
