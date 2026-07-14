package survey

// The system sheet: a full rendering of one surveyed hex. The canonical Second
// Survey line carries only the mainworld, so the great bulk of what the
// generators compute — the stellar family, the orbit map, every secondary world
// and moon, the mainworld's port facilities, native status, and Resource Units —
// has no other renderer. Sheet is that renderer.

import (
	"fmt"
	"strings"

	"github.com/philoserf/t5/internal/route"
	"github.com/philoserf/t5/internal/systemgen"
	"github.com/philoserf/t5/internal/uwp"
	"github.com/philoserf/t5/internal/worldgen"
)

// Sheet renders everything known about a surveyed hex: the mainworld in full
// (profile, extensions with Resource Units, nobility, bases, zone, native status,
// and starport facilities), the stellar family, and the complete orbit map with
// each world's UWP and each body's moons.
func (rec Record) Sheet() string {
	var b strings.Builder
	s := rec.System
	mw := s.Mainworld

	fmt.Fprintf(&b, "%s  %s%s\n", rec.Hex, rec.Name, capitalTitle(mw))
	fmt.Fprintf(&b, "%s\n", strings.Repeat("─", 64))

	// The mainworld, expanded past what the one-line record can hold.
	fmt.Fprintf(&b, "  Mainworld   %s", mw.Profile)
	if tcs := strings.Join(mw.TradeCodes, " "); tcs != "" {
		fmt.Fprintf(&b, "  %s", tcs)
	}
	b.WriteByte('\n')
	fmt.Fprintf(&b, "  Extensions  {%+d}%s%s   RU %d\n",
		mw.Importance, mw.Economic, mw.Cultural, mw.Economic.RU())
	fmt.Fprintf(&b, "  Traffic     ~%d ships/week\n", route.ExpectedTraffic(mw.Importance))
	if mw.Nobility != "" {
		fmt.Fprintf(&b, "  Nobility    %s\n", mw.Nobility)
	}
	fmt.Fprintf(&b, "  Bases       %s\n", basesOf(mw))
	fmt.Fprintf(&b, "  Travel Zone %s\n", zoneName(mw.Zone))
	if mw.NativeStatus != "" {
		fmt.Fprintf(&b, "  Natives     %s\n", mw.NativeStatus)
	}
	writeFacilities(&b, mw)

	// The stellar family.
	b.WriteString("\n  Stars\n")
	for _, e := range stellarFamily(s) {
		fmt.Fprintf(&b, "    %-18s %s\n", e.label, e.star)
	}

	// The orbit map — the bulk of the hidden detail.
	fmt.Fprintf(&b, "\n  Orbits — %d worlds · PBG %s · %d gas giants · %d belts\n",
		s.Worlds, s.PBG(), s.GasGiants, s.Belts)
	writeOrbits(&b, s)

	return strings.TrimRight(b.String(), "\n")
}

// capitalTitle names the capital a world is, if any.
func capitalTitle(w worldgen.World) string {
	switch {
	case hasCode(w, "Cx"):
		return "   — Sector Capital"
	case hasCode(w, "Cs"):
		return "   — Subsector Capital"
	default:
		return ""
	}
}

func hasCode(w worldgen.World, code string) bool {
	for _, c := range w.TradeCodes {
		if c == code {
			return true
		}
	}
	return false
}

// basesOf spells out the base codes the Second Survey line abbreviates.
func basesOf(w worldgen.World) string {
	var out []string
	if w.NavalBase {
		out = append(out, "Naval")
	}
	if w.ScoutBase {
		out = append(out, "Scout")
	}
	if w.WayStation {
		out = append(out, "Way Station")
	}
	if len(out) == 0 {
		return "none"
	}
	return strings.Join(out, ", ")
}

func zoneName(z byte) string {
	switch z {
	case 'A':
		return "Amber"
	case 'R':
		return "Red"
	default:
		return "Green"
	}
}

// writeFacilities renders the starport's services (Book 2 p.24), which the
// survey line has no room for.
func writeFacilities(b *strings.Builder, w worldgen.World) {
	f, ok := worldgen.PortFacilities(w.Profile.Starport, w.Profile.Population)
	if !ok {
		return
	}
	fmt.Fprintf(b, "  Starport    %c — %s\n", f.Class, f.Quality)
	var svc []string
	if f.Shipyard != "" {
		svc = append(svc, "builds "+f.Shipyard)
	}
	svc = append(svc, "repairs: "+f.Repairs.String())
	if f.Fuel.String() != "" {
		fuel := "fuel: " + f.Fuel.String()
		if f.RefuelHours != "" {
			fuel += " (" + f.RefuelHours + " hours)"
		}
		svc = append(svc, fuel)
	}
	if f.Downport {
		svc = append(svc, "downport")
	}
	if f.Beacon {
		svc = append(svc, "beacon only")
	}
	if f.Highport {
		svc = append(svc, "highport")
	}
	fmt.Fprintf(b, "              %s\n", strings.Join(svc, " · "))
}

// a starEntry is one member of the stellar family, with its role and orbit.
type starEntry struct{ label, star string }

// stellarFamily lists the system's stars in rotation order, noting the orbit each
// secondary holds around the primary. Companions orbit inside their own star.
func stellarFamily(s systemgen.System) []starEntry {
	out := []starEntry{{"Primary", s.Primary.String()}}
	add := func(label string, st *systemgen.Star, orbit int, companion bool) {
		if st == nil {
			return
		}
		if companion {
			out = append(out, starEntry{label, st.String() + "  (companion)"})
			return
		}
		out = append(out, starEntry{label, fmt.Sprintf("%s  (Orbit %d)", st, orbit)})
	}
	add("Primary companion", s.PrimaryCompanion, 0, true)
	add("Close", s.Close, s.CloseOrbit, false)
	add("Close companion", s.CloseCompanion, 0, true)
	add("Near", s.Near, s.NearOrbit, false)
	add("Near companion", s.NearCompanion, 0, true)
	add("Far", s.Far, s.FarOrbit, false)
	add("Far companion", s.FarCompanion, 0, true)
	return out
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
	case o.Kind == systemgen.KindMainworld && o.Giant != nil:
		return fmt.Sprintf("%-14s %s  — moon of Gas Giant %s", "Mainworld", mainworld, o.Giant)
	case o.Kind == systemgen.KindMainworld && o.Parent != nil:
		return fmt.Sprintf("%-14s %s  — moon of %s %s", "Mainworld", mainworld, o.Parent.Type, o.Parent.Profile)
	case o.Kind == systemgen.KindMainworld:
		return fmt.Sprintf("%-14s %s", "Mainworld", mainworld)
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
