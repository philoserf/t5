package worldgen

// Starport and spaceport facilities (Book 2 p.24). Every world's port is one of
// the mainworld starport classes A-E/X or, for a non-mainworld, a spaceport
// class F/G/H/Y. Its class fixes the shipyard, repair, fuel, and downport
// services; only the highport depends on population.

// FuelKind is the hydrogen fuel a port offers (Book 2 p.24).
type FuelKind int

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

// RepairLevel is the heaviest repair a port supports (Book 2 p.24).
type RepairLevel int

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

// Facilities describes a port's services (Book 2 p.24). Downport reports whether
// there is any surface port; Beacon marks an unmanned beacon-only downport
// (classes E/H). Highport is population-dependent and set by PortFacilities.
type Facilities struct {
	Class       byte
	Quality     string
	Shipyard    string // "", "Spacecraft", or "Starships"
	Repairs     RepairLevel
	Fuel        FuelKind
	Downport    bool
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
	}
	switch {
	case f.Beacon:
		out = append(out, "beacon-only downport")
	case f.Downport:
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
	'A': {Class: 'A', Quality: "Excellent", Shipyard: "Starships", Repairs: Overhaul, Fuel: RefinedFuel, Downport: true, RefuelHours: "2D"},
	'B': {Class: 'B', Quality: "Good", Shipyard: "Spacecraft", Repairs: Overhaul, Fuel: RefinedFuel, Downport: true, RefuelHours: "2D"},
	'C': {Class: 'C', Quality: "Routine", Repairs: MajorRepairs, Fuel: UnrefinedFuel, Downport: true, RefuelHours: "4D"},
	'D': {Class: 'D', Quality: "Poor", Repairs: MinorRepairs, Fuel: UnrefinedFuel, Downport: true, RefuelHours: "4D"},
	'E': {Class: 'E', Quality: "Frontier", Repairs: NoRepairs, Fuel: NoFuel, Downport: true, Beacon: true},
	'X': {Class: 'X', Quality: "None", Repairs: NoRepairs, Fuel: NoFuel},
	'F': {Class: 'F', Quality: "Good", Repairs: MinorRepairs, Fuel: UnrefinedFuel, Downport: true, RefuelHours: "4D"},
	'G': {Class: 'G', Quality: "Poor", Repairs: SuperficialRepairs, Fuel: UnrefinedFuel, Downport: true, RefuelHours: "4D"},
	'H': {Class: 'H', Quality: "Basic", Repairs: NoRepairs, Fuel: NoFuel, Downport: true, Beacon: true},
	'Y': {Class: 'Y', Quality: "None", Repairs: NoRepairs, Fuel: NoFuel},
}

// highportPop is the mainworld population at or above which a class-A/B/C
// starport has a highport (Book 2 p.24). Other classes never have one.
var highportPop = map[byte]int{'A': 7, 'B': 8, 'C': 9}

// PortFacilities returns the services for a port class and the world's
// population (population matters only for the A/B/C highport). It reports false
// for an unknown class letter.
func PortFacilities(class byte, population int) (Facilities, bool) {
	f, ok := portTable[class]
	if !ok {
		return Facilities{}, false
	}
	if threshold, gated := highportPop[class]; gated {
		f.Highport = population >= threshold
	}
	return f, true
}
