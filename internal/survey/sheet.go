package survey

// The system sheet: a full rendering of one surveyed hex. The canonical Second
// Survey line carries only the mainworld, so the great bulk of what the
// generators compute — the stellar family, the orbit map, every secondary world
// and moon, the mainworld's port facilities, native status, and Resource Units —
// has no other renderer. Sheet is that renderer. It lays out; each generator
// still owns the meaning of its own data (systemgen.System.Stars,
// worldgen.Facilities.Services, worldgen.ZoneName, worldgen.World.BaseNames).

import (
	"fmt"
	"strings"

	"github.com/philoserf/t5/internal/route"
	"github.com/philoserf/t5/internal/systemgen"
	"github.com/philoserf/t5/internal/tradecode"
	"github.com/philoserf/t5/internal/uwp"
	"github.com/philoserf/t5/internal/worldgen"
)

// doublePlanetTag names the designation once. bodyLabel and moonLabel are fifty
// lines apart and the invariant is that they read alike: an equal-size
// mainworld/host pair and an equal-size sibling moon describe the same physical
// situation, so a reworded tag must not apply to only one of them.
const doublePlanetTag = "  [double planet]"

// rule separates the sheet's header from its body.
const rule = "────────────────────────────────────────────────────────────────"

// Sheet renders everything known about a surveyed hex: the mainworld in full
// (profile, extensions with Resource Units, nobility, bases, zone, native status,
// and starport facilities), the stellar family, and the complete orbit map with
// each world's UWP and each body's moons.
func (rec Record) Sheet() string {
	var b strings.Builder
	b.Grow(2048)

	s := rec.System
	mw := s.Mainworld

	// field writes one "label   value" row of the mainworld block.
	field := func(label, format string, a ...any) {
		fmt.Fprintf(&b, "  %-11s %s\n", label, fmt.Sprintf(format, a...))
	}

	fmt.Fprintf(&b, "%s  %s%s\n%s\n", rec.Hex, rec.Name, capitalTitle(mw), rule)

	// The mainworld's codes are finalized in phases (base+zone at generation, climate
	// and satellite at placement, a capital in the region survey), so unlike a
	// non-mainworld they are ordered here, at render, not at the source.
	field(
		"Mainworld",
		"%s  %s",
		mw.Profile,
		tradecode.Join(worldgen.OrderTradeCodes(mw.TradeCodes), " "),
	)
	field("Extensions", "%s   RU %d", mw.Extensions(), mw.Economic().RU())
	field("Traffic", "~%s/week", plural(route.ExpectedTraffic(mw.Importance()), "ship", "ships"))

	if mw.Nobility() != "" {
		field("Nobility", "%s", mw.Nobility())
	}

	field("Bases", "%s", orNone(mw.BaseNames()))
	field("Travel Zone", "%s", worldgen.ZoneName(mw.Zone))

	if mw.NativeStatus != "" {
		field("Natives", "%s", mw.NativeStatus)
	}

	if f, ok := worldgen.PortFacilities(mw.Profile, mw.Belt); ok {
		field("Starport", "%c — %s", f.Class, f.Quality)
		// Most no-port worlds (X, Y) list nothing — but a class-X world with water or
		// ice still offers local unrefined fuel (Book 3 p.24), so the guard is on the
		// service list being non-empty, not on the class.
		if svc := f.Services(); len(svc) > 0 {
			fmt.Fprintf(&b, "              %s\n", strings.Join(svc, " · "))
		}
	}

	b.WriteString("\n  Stars\n")

	for _, sl := range s.Stars() {
		fmt.Fprintf(&b, "    %-18s %s%s\n", sl.Label, sl.Star, starOrbit(sl))
	}

	fmt.Fprintf(&b, "\n  Orbits — %s · PBG %s · %s · %s\n",
		plural(s.Worlds, "world", "worlds"), s.PBG(),
		plural(s.GasGiants, "gas giant", "gas giants"), plural(s.Belts, "belt", "belts"))
	writeOrbits(&b, s)
	writeUnplacedMainworld(&b, s)

	return strings.TrimRight(b.String(), "\n")
}

// plural renders a count with the right form of its noun.
func plural(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, one)
	}

	return fmt.Sprintf("%d %s", n, many)
}

// writeUnplacedMainworld names the mainworld when the orbit map has no place for
// it: a primary with no habitable zone gives systemgen no orbit to put it in, so
// it would otherwise vanish from a map that claims to be complete.
func writeUnplacedMainworld(b *strings.Builder, s systemgen.System) {
	for _, o := range s.Orbits {
		if o.Kind == systemgen.KindMainworld {
			return
		}
	}

	fmt.Fprintf(b, "    *  --  %-14s %s  — unplaced (the primary has no habitable zone)\n",
		"Mainworld", s.Mainworld.Profile)
}

// starOrbit notes where a star sits: a secondary holds a numbered orbit around
// the primary, while a companion orbits inside its own star.
func starOrbit(sl systemgen.StarSlot) string {
	switch {
	case sl.Companion:
		return "  (companion)"
	case sl.Orbit >= 0:
		return fmt.Sprintf("  (Orbit %d)", sl.Orbit)
	default:
		return ""
	}
}

// capitalTitle names the capital a world is, if any. The names come from worldgen
// (the one source of truth for the codes), so this cannot drift from what the
// survey stamps — it once did, labelling a Sector Capital "Subsector" after the
// codes were corrected here but not there.
func capitalTitle(w worldgen.World) string {
	if code := w.CapitalCode(); code != "" {
		return "   — " + worldgen.CapitalName(code)
	}

	return ""
}

func orNone(items []string) string {
	if len(items) == 0 {
		return "none"
	}

	return strings.Join(items, ", ")
}

// writeOrbits renders the orbit map as a tree: one line per placed body, with its
// moons indented beneath it. Orbits are grouped under their host star.
func writeOrbits(b *strings.Builder, s systemgen.System) {
	host := ""
	for _, o := range s.Orbits {
		if o.Host != host {
			host = o.Host
			fmt.Fprintf(b, "    [%s]\n", host)
		}

		marker := " "
		if o.Kind == systemgen.KindMainworld {
			marker = "*"
		}

		fmt.Fprintf(b, "   %s%3d  %s\n", marker, o.Orbit, bodyLabel(o, s.Mainworld.Profile))

		// A satellite mainworld's orbit is held by its parent body, so the moons
		// listed under it are the mainworld's siblings around that parent, not its
		// own (Book 3 p.21/p.29; see systemgen's satelliteParent).
		sibling := o.Kind == systemgen.KindMainworld && (o.Giant != nil || o.Parent != nil)
		for _, m := range o.Satellites {
			fmt.Fprintf(b, "          · %s\n", moonLabel(m, sibling))
		}
	}
}

// bodyLabel describes what occupies an orbit. The mainworld's own profile is
// passed in, since its PlacedOrbit records only that it is the mainworld.
func bodyLabel(o systemgen.PlacedOrbit, mainworld uwp.Profile) string {
	switch {
	case o.Kind == systemgen.KindMainworld:
		s := fmt.Sprintf("%-14s %s", "Mainworld", mainworld)

		switch {
		case o.Giant != nil:
			s += fmt.Sprintf("  — moon of Gas Giant %s", o.Giant)
		case o.Parent != nil:
			s += fmt.Sprintf("  — moon of %s %s", o.Parent.Type(), o.Parent.Profile())
			// The same words moonLabel gives an equal-size moon: on this sheet a
			// double planet is a double planet, whichever body is the mainworld.
			if o.DoublePlanet {
				s += doublePlanetTag
			}
		}

		return s
	case o.Giant != nil:
		return fmt.Sprintf("%-14s %s", "Gas Giant", o.Giant)
	case o.Kind == systemgen.KindBelt:
		return o.Kind.String() // systemgen names the kinds; do not restate them here
	case o.World != nil:
		s := fmt.Sprintf("%-14s %s", o.World.Type(), o.World.Profile())
		// A non-mainworld's codes are stored in Chart D order by the assembler
		// (worldgen.TradeClassificationsWithContext), so they render as-is.
		if tcs := tradecode.Join(o.World.TradeCodes(), " "); tcs != "" {
			s += "  " + tcs
		}

		return s
	default:
		return o.Kind.String()
	}
}

// moonLabel describes one satellite: a ring, or a moon with its own world type
// and UWP, its orbit name, and whether it forms a double planet with its parent.
// A sibling moon shares its parent with the mainworld rather than orbiting the
// body on the line above it.
func moonLabel(m systemgen.Satellite, sibling bool) string {
	if m.Ring {
		return "Ring"
	}

	orbit := "close"
	if m.Far {
		orbit = "far"
	}

	kind := "moon"
	if sibling {
		kind = "sibling moon"
	}

	s := fmt.Sprintf("%s %-5s %-12s %s", kind, m.OrbitLetter, m.Type(), m.Profile())
	// Stored in Chart D order by the assembler, so no render-time sort here.
	if tcs := tradecode.Join(m.TradeCodes(), " "); tcs != "" {
		s += " " + tcs
	}

	s += fmt.Sprintf("  (%s orbit)", orbit)
	if m.DoublePlanet {
		s += doublePlanetTag
	}

	return s
}
