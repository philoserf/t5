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
	"slices"
	"strings"

	"github.com/philoserf/t5/internal/route"
	"github.com/philoserf/t5/internal/systemgen"
	"github.com/philoserf/t5/internal/uwp"
	"github.com/philoserf/t5/internal/worldgen"
)

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

	field("Mainworld", "%s  %s", mw.Profile, strings.Join(mw.TradeCodes, " "))
	field("Extensions", "{%+d}%s%s   RU %d", mw.Importance, mw.Economic, mw.Cultural, mw.Economic.RU())
	field("Traffic", "~%d ships/week", route.ExpectedTraffic(mw.Importance))
	if mw.Nobility != "" {
		field("Nobility", "%s", mw.Nobility)
	}
	field("Bases", "%s", orNone(mw.BaseNames()))
	field("Travel Zone", "%s", worldgen.ZoneName(mw.Zone))
	if mw.NativeStatus != "" {
		field("Natives", "%s", mw.NativeStatus)
	}
	if f, ok := worldgen.PortFacilities(mw.Profile.Starport, mw.Profile.Population); ok {
		field("Starport", "%c — %s", f.Class, f.Quality)
		fmt.Fprintf(&b, "              %s\n", strings.Join(f.Services(), " · "))
	}

	b.WriteString("\n  Stars\n")
	for _, sl := range s.Stars() {
		fmt.Fprintf(&b, "    %-18s %s%s\n", sl.Label, sl.Star, starOrbit(sl))
	}

	fmt.Fprintf(&b, "\n  Orbits — %d worlds · PBG %s · %d gas giants · %d belts\n",
		s.Worlds, s.PBG(), s.GasGiants, s.Belts)
	writeOrbits(&b, s)

	return strings.TrimRight(b.String(), "\n")
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

// capitalTitle names the capital a world is, if any.
func capitalTitle(w worldgen.World) string {
	switch {
	case slices.Contains(w.TradeCodes, "Cx"):
		return "   — Sector Capital"
	case slices.Contains(w.TradeCodes, "Cs"):
		return "   — Subsector Capital"
	default:
		return ""
	}
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
		for _, m := range o.Satellites {
			fmt.Fprintf(b, "          · %s\n", moonLabel(m))
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
			s += fmt.Sprintf("  — moon of %s %s", o.Parent.Type, o.Parent.Profile)
		}
		return s
	case o.Giant != nil:
		return fmt.Sprintf("%-14s %s", "Gas Giant", o.Giant)
	case o.Kind == systemgen.KindBelt:
		return "Planetoid Belt"
	case o.World != nil:
		s := fmt.Sprintf("%-14s %s", o.World.Type, o.World.Profile)
		if tcs := strings.Join(o.World.TradeCodes, " "); tcs != "" {
			s += "  " + tcs
		}
		return s
	default:
		return o.Kind.String()
	}
}

// moonLabel describes one satellite: a ring, or a moon with its own world type
// and UWP, its orbit name, and whether it forms a double planet with its parent.
func moonLabel(m systemgen.Satellite) string {
	if m.Ring {
		return "Ring"
	}
	orbit := "close"
	if m.Far {
		orbit = "far"
	}
	s := fmt.Sprintf("moon %-5s %-12s %s  (%s orbit)", m.OrbitLetter, m.Type, m.Profile, orbit)
	if m.DoublePlanet {
		s += "  [double planet]"
	}
	return s
}
