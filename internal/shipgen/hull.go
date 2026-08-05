package shipgen

// Hull design (Book 2 pp.51-53, 70-71). All hull attributes are deterministic
// lookups/formulas over the hull size ordinal, configuration, structure, and TL.

// configCost is each configuration's hull cost in MCr as a linear function of the
// size ordinal: cost = slope*ordinal + intercept. Verified against the "A Hull
// Costs" grid (Book 2 p.70) — every config column is exactly linear (e.g. Hull-A
// Lifting = 12*1+4 = 16, Hull-Z Streamlined = 6*24+2 = 146).
var configCost = [...]struct{ slope, intercept int }{
	Cluster:       {2, 0},
	Braced:        {3, 0},
	Planetoid:     {1, 0},
	Unstreamlined: {3, 2},
	Streamlined:   {6, 2},
	Airframe:      {7, 2},
	Lifting:       {12, 4},
}

// configAttr is each configuration's performance characteristics (Book 2 p.71).
// Friction is a code: >0 means x N (heating multiplier), <0 means / -N (divisor).
// Agility and Stability are base mods before over/undertonnage; land is whether
// the hull can make world landings (Planetoid's "?" is treated as no).
var configAttr = [...]struct {
	friction, agility, maxG, stability int
	land                               bool
}{
	Cluster:       {2, -5, 1, -3, false},
	Braced:        {2, -4, 3, -2, false},
	Planetoid:     {1, -2, 9, -1, false},
	Unstreamlined: {-2, -1, 9, 0, true},
	Streamlined:   {-3, 0, 9, 1, true},
	Airframe:      {-4, 1, 9, 2, true},
	Lifting:       {-5, 1, 9, 3, true},
}

// validConfig reports whether a HullConfig names a real row of the p.70/p.71 tables.
// Design uses it to raise a Problem; configIndex is what keeps the lookups safe.
func validConfig(c HullConfig) bool { return c >= Cluster && int(c) < len(configAttr) }

// configIndex bounds a HullConfig to the tables, so an out-of-range one reads as
// Cluster rather than panicking — the same guard-of-last-resort stageIndex gives
// a Stage. It is not a silent default: Design reports the substitution as a
// Problem, because a hull nobody asked for is its own kind of infeasible.
func configIndex(c HullConfig) HullConfig {
	if !validConfig(c) {
		return Cluster
	}

	return c
}

// hullCostMCr returns a hull's cost in MCr for the given size ordinal and config
// (Book 2 p.70). Structure multipliers are applied by hullCost.
func hullCostMCr(ordinal int, config HullConfig) int {
	c := configCost[configIndex(config)]

	return c.slope*ordinal + c.intercept
}

// hullCost returns a hull's cost in Cr, applying the structure multiplier
// (Organic x0.5, Charged x2; Book 2 p.52).
func hullCost(ordinal int, config HullConfig, structure HullMaterial) int {
	cr := hullCostMCr(ordinal, config) * 1_000_000

	switch structure {
	case Organic:
		cr /= 2
	case Charged:
		cr *= 2
	default:
		// FramePlate, Shell, Polymer, and FeNi carry no cost multiplier.
	}

	return cr
}

// hull designs a hull (Book 2 pp.51-53, 70-71). A tons of 0 uses the ordinal's
// nominal tonnage; a non-zero tons applies the over/undertonnage rules — gross
// overtonnage (>49t over) rounds up to the next hull size, and lesser variance
// adjusts Agility (and, when over, Stability). Base armor value is the TL
// (Shell: TL/2). Hardpoints are one per 100 tons.
func hull(tl, ordinal, tons int, config HullConfig, structure HullMaterial) Hull {
	nominal := HullTons(ordinal)
	if tons == 0 {
		tons = nominal
	}
	// Gross overtonnage rounds up to the next hull size (Book 2 p.70).
	for tons-nominal > 49 && ordinal < maxLetter {
		ordinal++
		nominal = HullTons(ordinal)
	}

	config = configIndex(config)
	attr := configAttr[config]
	agility, stability := attr.agility, attr.stability

	switch delta := tons - nominal; {
	case delta >= 25:
		agility -= 2
		stability -= 2
	case delta >= 1:
		agility--
		stability--
	case delta <= -25:
		agility += 2
	case delta <= -1:
		agility++
	}

	return Hull{
		Letter: ordinal, Tons: tons, Config: config, Structure: structure,
		Friction: attr.friction, MaxG: attr.maxG,
		Agility: agility, Stability: stability, LandCapable: attr.land,
		Hardpoints: tons / 100,
		BaseArmor:  structureAV(structure, tl),
		Cost:       hullCost(ordinal, config, structure),
	}
}
