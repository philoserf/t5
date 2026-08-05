package worldgen

import (
	"fmt"
	"slices"
	"strings"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/tradecode"
	"github.com/philoserf/t5/internal/uwp"
)

// A World is a fully-detailed mainworld: its Universal World Profile plus the
// derived data from Book 3 Charts C-F — trade classifications, the {Ix}(Ex)[Cx]
// Extensions, nobility, bases, travel zone, native status, and the population
// multiplier digit.
type World struct { //nolint:recvcheck // deliberate value-reader / pointer-mutator split
	Profile    uwp.Profile
	TradeCodes []tradecode.Code
	// Belt reports whether the mainworld is an asteroid belt. For a mainworld — and
	// only a mainworld — belt-ness is Size 0: Book 3 p.16 makes it a belt "when
	// World Size is generated", whether the sector map forced it (GenerateBeltWorld)
	// or the 2D-2 roll came up 0. The field carries that fact so systemgen asks it
	// (placement, satellites, port) rather than re-reading the Size digit, and so a
	// secondary world — where Size 0 does NOT mean belt (a Worldlet is not a belt,
	// #324) — is never mistaken for one by sharing this accessor.
	Belt bool
	// The four derived Extensions are unexported behind readers (Importance,
	// Economic, Cultural, Nobility) so a caller cannot set them out of step with
	// the Profile/TradeCodes/bases they descend from. They change only through the
	// Set* mutators, which recompute them. This follows skill.Set's encapsulation
	// (chargen.go), the pattern this repo already uses where the world/system model
	// did not (#330).
	importance      int
	economic        Economic
	cultural        Cultural
	nobility        string
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

// NewWorld assembles a World from already-computed parts, the path for building a
// world from known values rather than rolling it — a test fixture, or a record
// parsed back in (#327). base carries the exported inputs and flags (Profile,
// TradeCodes, Belt, bases, Zone, NativeStatus, PopulationDigit); the four derived
// Extensions, unexported so a caller cannot leave them stale against those inputs
// (#330), are supplied here and stamped in atomically. GenerateWorld is the rolled
// path; after construction the Extensions change only through the Set* mutators.
func NewWorld(base World, importance int, economic Economic, cultural Cultural, nobility string) World {
	base.importance = importance
	base.economic = economic
	base.cultural = cultural
	base.nobility = nobility

	return base
}

func generateWorld(r *dice.Roller, gasGiants, belts int, isCapital, belt bool) World {
	p := generate(r, belt)
	tcs := TradeClassificationsWithContext(p, WorldContext{IsMainworld: true})
	naval, scout := RollBases(r, p.Starport)
	ix := Importance(p, tcs, naval, scout, false)

	return World{
		Profile:    p,
		TradeCodes: tcs,
		// A mainworld is a belt iff its Size is 0 (Book 3 p.16) — whether the belt
		// param forced it or the 2D-2 roll produced it. See World.Belt.
		Belt:            p.Size == uwp.BeltSize,
		importance:      ix,
		economic:        RollEconomic(r, p, ix, gasGiants, belts),
		cultural:        RollCultural(r, p, ix),
		nobility:        Nobility(tcs, ix, isCapital),
		NavalBase:       naval,
		ScoutBase:       scout,
		Zone:            TravelZone(p),
		NativeStatus:    NativeStatus(p),
		PopulationDigit: PopulationDigit(r, p.Population),
	}
}

// Importance reports the world's Importance Extension {Ix} (Book 3 Chart E).
func (w World) Importance() int { return w.importance }

// Economic reports the world's Economic Extension (Ex).
func (w World) Economic() Economic { return w.economic }

// Cultural reports the world's Cultural Extension [Cx].
func (w World) Cultural() Cultural { return w.cultural }

// Nobility reports the world's nobility codes (Book 3 Chart D).
func (w World) Nobility() string { return w.nobility }

// SetCapital marks the world a capital carrying the given capital trade code —
// Cp (Subsector Capital), Cs (Sector Capital), or Cx (Imperial Capital) — and
// recomputes its nobility as a capital (Book 3 Chart D p.26). Which world is the
// capital is a whole-region decision the caller makes; this only encodes the
// result on the world.
func (w *World) SetCapital(code tradecode.Code) {
	w.addTradeCode(code)

	w.nobility = Nobility(w.TradeCodes, w.importance, true)
}

// SetColony marks the mainworld as a colony owned by another world (Book 3
// Chart D, p.26). Ownership itself is a relationship between sector hexes and is
// retained by survey; the world record carries the Cy classification.
func (w *World) SetColony() {
	w.addTradeCode(tradecode.Cy)
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
	w.importance = Importance(w.Profile, w.TradeCodes, w.NavalBase, w.ScoutBase, true)
	w.nobility = Nobility(w.TradeCodes, w.importance, hasCapitalCode(w.TradeCodes))
}

// capitalNames maps each capital trade code to the office it marks (Book 3 Chart D
// p.26). This is the one source of truth for the codes, so a renderer and the
// nobility check cannot disagree about which is which — they did once, when the
// codes were corrected in markSectorCapitals but the sheet's labels were not.
var capitalNames = map[tradecode.Code]string{
	tradecode.Cp: "Subsector Capital",
	tradecode.Cs: "Sector Capital",
	tradecode.Cx: "Imperial Capital",
}

// CapitalName returns the office a capital trade code marks, or "" if the code is
// not a capital.
func CapitalName(code tradecode.Code) string { return capitalNames[code] }

// CapitalCode returns the capital code the world carries, or "" if it is not a
// capital.
func (w World) CapitalCode() tradecode.Code {
	for _, tc := range w.TradeCodes {
		if _, ok := capitalNames[tc]; ok {
			return tc
		}
	}

	return ""
}

// hasCapitalCode reports whether the trade codes already mark the world a capital.
func hasCapitalCode(tcs []tradecode.Code) bool {
	for _, tc := range tcs {
		if _, ok := capitalNames[tc]; ok {
			return true
		}
	}

	return false
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
		dashIfEmpty(tradecode.Join(OrderTradeCodes(w.TradeCodes), " ")),
		w.Extensions(),
		w.nobility,
		dashIfEmpty(w.bases()),
		dashIfEmpty(w.zone()),
	}

	return strings.Join(fields, " ")
}

// Extensions renders the world's three extensions as they appear in the Second
// Survey record: "{Ix}(Ex)[Cx]", e.g. "{+4}(D7E+4)[9C6D]".
func (w World) Extensions() string {
	return importance(w.importance) + w.economic.String() + w.cultural.String()
}

// importance renders the Importance Extension {Ix}. It is signed when non-zero
// but bare at zero, matching the book's Importance table (0, +4, -2).
func importance(value int) string {
	if value == 0 {
		return "{0}"
	}

	return fmt.Sprintf("{%+d}", value)
}

// BaseNames names the bases the world hosts, in the order their codes appear in
// the Second Survey record (Book 3 p.28). Empty when the world hosts none.
func (w World) BaseNames() []string {
	out := make([]string, 0, len(w.baseList()))
	for _, b := range w.baseList() {
		out = append(out, b.name)
	}

	return out
}

// A baseEntry pairs a base's record code with its display name (Book 3 Chart F,
// p.28). The code is not the name's initial: a Naval Depot's code is "D", not "N",
// and it would otherwise collide with the Naval Base.
type baseEntry struct{ code, name string }

// baseList is the world's bases in Chart F order.
func (w World) baseList() []baseEntry {
	var out []baseEntry
	if w.NavalBase {
		out = append(out, baseEntry{"N", "Naval"})
	}

	if w.ScoutBase {
		out = append(out, baseEntry{"S", "Scout"})
	}

	if w.NavalDepot {
		out = append(out, baseEntry{"D", "Naval Depot"})
	}

	if w.WayStation {
		out = append(out, baseEntry{"W", "Way Station"})
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

// addTradeCode adds code to the world's trade codes if it is not already
// present.
func (w *World) addTradeCode(code tradecode.Code) {
	if !slices.Contains(w.TradeCodes, code) {
		w.TradeCodes = append(w.TradeCodes, code)
	}
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
