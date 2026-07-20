package worldgen

import "github.com/philoserf/t5/internal/uwp"

// Starport and spaceport facilities (Book 2 p.24; the fuel table is Book 3 p.24).
// Every world's port is one of the mainworld starport classes A-E/X or, for a
// non-mainworld, a spaceport class F/G/H/Y. Its class fixes the shipyard, repair,
// hydrogen fuel, and downport services; the highport depends on population, the
// exotic fuels on tech level, and the local-fuel fallback on the world itself.

// FuelKind is the hydrogen fuel a port offers (Book 2 p.24).
type FuelKind int

// Starport fuel availability.
const (
	NoFuel        FuelKind = iota // none at the port
	UnrefinedFuel                 // raw/unrefined only
	RefinedFuel                   // refined and unrefined ("Both")
)

func (f FuelKind) String() string {
	switch f {
	case RefinedFuel:
		return "Refined+Unrefined"
	case UnrefinedFuel:
		return "Unrefined"
	default:
		return "None"
	}
}

// exoticFuelMinTL is the tech level at which a class-A or -B starport begins to
// offer each exotic fuel (Book 3 p.24 "Fuel at Starports and Spaceports"):
// Radioactives at TL 8, Collector at TL 14, Anti-Matter at TL 18. No other class
// offers any of them.
var exoticFuelMinTL = []struct {
	name  string
	minTL int
}{
	{"Radioactives", 8},
	{"Collector", 14},
	{"Anti-Matter", 18},
}

// RepairLevel is the heaviest repair a port supports (Book 2 p.24).
type RepairLevel int

// Starport repair capability.
const (
	NoRepairs RepairLevel = iota
	SuperficialRepairs
	MinorRepairs
	MajorRepairs
	Overhaul
)

func (r RepairLevel) String() string {
	switch r {
	case Overhaul:
		return "Overhaul"
	case MajorRepairs:
		return "Major"
	case MinorRepairs:
		return "Minor"
	case SuperficialRepairs:
		return "Superficial"
	default:
		return "None"
	}
}

// portKind is a port's surface presence: none, a downport, or — for an asteroid
// belt — a beltport that replaces the downport (Book 2 p.24, "An Asteroid Mainworld
// has a Beltport instead"). It is one field, not a Downport/Beltport bool pair,
// because a port that is both is not a real port; the pair could represent that and
// this cannot (#330). Set only by PortFacilities.
type portKind uint8

const (
	portNone portKind = iota
	portDown
	portBelt
)

// Facilities describes a port's services (Book 2 p.24). Beacon marks an unmanned
// beacon-only downport (classes E/H). Highport is population-dependent; ExoticFuels
// are tech-level-gated; LocalFuel is the environmental fallback; port is the
// surface-port kind (downport or, for a belt, beltport) — all set by PortFacilities.
type Facilities struct {
	Class       byte
	Quality     string
	Shipyard    string // "", "Spacecraft", or "Starships"
	Repairs     RepairLevel
	Fuel        FuelKind
	ExoticFuels []string // Radioactives / Collector / Anti-Matter, TL-gated (A/B only)
	LocalFuel   bool     // no port fuel, but unrefined is skimmable from the world
	port        portKind
	Beacon      bool
	Highport    bool
	RefuelHours string // "2D", "4D", or "" when no fuel
}

// Services lists a port's services in prose (Book 2 p.24): what it builds, how
// far it can repair, its fuel and refuelling time, and its surface and orbital
// facilities. Each service is named only when the port offers it — so a fuel-less
// port omits fuel, and classes X and Y, which are no port at all, return nothing
// rather than advertising a repair capability they do not have. A beacon-only
// downport (classes E and H) is named as such rather than claiming a staffed one.
func (f Facilities) Services() []string {
	var out []string
	if f.Shipyard != "" {
		out = append(out, "builds "+f.Shipyard)
	}

	if f.Repairs != NoRepairs {
		out = append(out, "repairs: "+f.Repairs.String())
	}

	if f.Fuel != NoFuel {
		fuel := "fuel: " + f.Fuel.String()
		if f.RefuelHours != "" {
			fuel += " (" + f.RefuelHours + " hours)"
		}

		out = append(out, fuel)
	} else if f.LocalFuel {
		// No port fuel, but the world's own water or ice can be skimmed unrefined
		// (Book 3 p.24, the ** note) — the difference between "stranded" and "slow".
		out = append(out, "fuel: Unrefined (local water/ice)")
	}

	for _, e := range f.ExoticFuels {
		out = append(out, "exotic fuel: "+e)
	}

	switch {
	case f.port == portBelt:
		out = append(out, "beltport")
	case f.Beacon:
		out = append(out, "beacon-only downport")
	case f.port == portDown:
		out = append(out, "downport")
	}

	if f.Highport {
		out = append(out, "highport")
	}

	return out
}

// portTable is the fixed, population-independent part of each port class
// (Book 2 p.24). Highport is filled in by PortFacilities.
var portTable = map[byte]Facilities{
	'A': {
		Class:       'A',
		Quality:     "Excellent",
		Shipyard:    "Starships",
		Repairs:     Overhaul,
		Fuel:        RefinedFuel,
		port:        portDown,
		RefuelHours: "2D",
	},
	'B': {
		Class:       'B',
		Quality:     "Good",
		Shipyard:    "Spacecraft",
		Repairs:     Overhaul,
		Fuel:        RefinedFuel,
		port:        portDown,
		RefuelHours: "2D",
	},
	'C': {
		Class:       'C',
		Quality:     "Routine",
		Repairs:     MajorRepairs,
		Fuel:        UnrefinedFuel,
		port:        portDown,
		RefuelHours: "4D",
	},
	'D': {
		Class:       'D',
		Quality:     "Poor",
		Repairs:     MinorRepairs,
		Fuel:        UnrefinedFuel,
		port:        portDown,
		RefuelHours: "4D",
	},
	'E': {
		Class:   'E',
		Quality: "Frontier",
		Repairs: NoRepairs,
		Fuel:    NoFuel,
		port:    portDown,
		Beacon:  true,
	},
	'X': {Class: 'X', Quality: "None", Repairs: NoRepairs, Fuel: NoFuel},
	'F': {
		Class:       'F',
		Quality:     "Good",
		Repairs:     MinorRepairs,
		Fuel:        UnrefinedFuel,
		port:        portDown,
		RefuelHours: "4D",
	},
	'G': {
		Class:       'G',
		Quality:     "Poor",
		Repairs:     SuperficialRepairs,
		Fuel:        UnrefinedFuel,
		port:        portDown,
		RefuelHours: "4D",
	},
	'H': {
		Class:   'H',
		Quality: "Basic",
		Repairs: NoRepairs,
		Fuel:    NoFuel,
		port:    portDown,
		Beacon:  true,
	},
	'Y': {Class: 'Y', Quality: "None", Repairs: NoRepairs, Fuel: NoFuel},
}

// highportPop is the mainworld population at or above which a class-A/B/C
// starport has a highport (Book 2 p.24). Other classes never have one.
var highportPop = map[byte]int{'A': 7, 'B': 8, 'C': 9}

// PortFacilities returns the services for a world's port, from its whole profile:
// the class fixes most of it, but the highport depends on population, the exotic
// fuels on tech level (Book 3 p.24), the local-fuel fallback on hydrographics, and
// the beltport on whether the mainworld is an asteroid belt. belt is a body fact
// the caller supplies (worldgen.World.Belt) rather than a Size read, since a
// tiny Size-0 world is not a belt and keeps its downport (#324). It reports false
// for an unknown class letter.
func PortFacilities(p uwp.Profile, belt bool) (Facilities, bool) {
	f, ok := portTable[p.Starport]
	if !ok {
		return Facilities{}, false
	}

	if threshold, gated := highportPop[p.Starport]; gated {
		f.Highport = p.Population >= threshold
	}
	// Class A and B carry the exotic fuels their tech level has reached (Book 3
	// p.24); no other class offers any.
	if p.Starport == 'A' || p.Starport == 'B' {
		for _, e := range exoticFuelMinTL {
			if p.TechLevel >= e.minTL {
				f.ExoticFuels = append(f.ExoticFuels, e.name)
			}
		}
	}
	// A port with no hydrogen of its own can still be refuelled from the world's
	// own water or ice, unrefined (Book 3 p.24, the ** note): any Hydrographics at
	// all means there is something to skim. Gas-giant skimming is a system fact, not
	// a world one, so it is left to the referee.
	if f.Fuel == NoFuel && p.Hydrographics >= 1 {
		f.LocalFuel = true
	}
	// An asteroid-belt mainworld has a Beltport in place of a downport (Book 2
	// p.24). A tiny Size-0 world that is not a belt keeps its downport.
	if belt && f.port == portDown {
		f.port = portBelt
	}

	return f, true
}
