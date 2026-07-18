package shipgen

// Fuel and armor (Book 2 pp. 75, 79).

// Fitting costs (Book 2 p.79 F): the fuel scoop and purifier a wilderness
// refueler carries, each 1 ton.
const (
	fuelScoopCost    = 100_000   // KCr100
	fuelPurifierCost = 1_000_000 // MCr1
	liquidHydrogenCr = 500       // Cr500 per ton
)

// fuel computes a ship's fuel tankage and cost (Book 2 p.79). Jump fuel is
// P*hull/10 and power-plant operations fuel is P*hull/100 (one month's supply),
// each scaled by its drive's stage fuel multiplier. Maneuver fuel is included in
// operations. Cost is the liquid hydrogen plus any scoop/purifier fittings.
func fuel(hullTons int, jump, power *Drive, scoop, purifier bool) Fuel {
	tankage := 0
	if jump != nil {
		tankage += fuelMul(jump.Potential*hullTons/10, jump.Stage)
	}

	if power != nil {
		tankage += fuelMul(power.Potential*hullTons/100, power.Stage)
	}

	cost := tankage * liquidHydrogenCr
	if scoop {
		cost += fuelScoopCost
	}

	if purifier {
		cost += fuelPurifierCost
	}

	return Fuel{Tons: tankage, Cost: cost}
}

// fuelMul scales a base fuel amount by a drive stage's fuel multiplier (Book 2
// p.76 X, e.g. Experimental x2.0, Advanced x0.8).
func fuelMul(base int, stage Stage) int {
	return base * stageData[stageIndex(stage)].fuelPct / 100
}

// structureAV is the per-layer armor value of a hull structure (Book 2 p.75 B):
// Plate = TL, Shell/Polymer/Organic = TL/2, FeNi = 20, Charged = TL*2.
func structureAV(structure Structure, tl int) int {
	switch structure {
	case Shell, Polymer, Organic:
		return tl / 2
	case FeNi:
		return 20
	case Charged:
		return tl * 2
	default: // FramePlate -> Plate
		return tl
	}
}

// armor computes a hull's armor (Book 2 p.75): layer 1 is integral to the hull
// (no tonnage), and each further layer costs 4% of the hull's tonnage (Shell
// 2%). The armor value is the per-layer value (identical layers are not summed
// on the record); armor imposes no monetary cost.
func armor(tl, hullTons int, structure Structure, layers int) Armor {
	if layers < 1 {
		layers = 1
	}

	pct := 4
	if structure == Shell {
		pct = 2
	}

	return Armor{
		Layers: layers,
		AV:     structureAV(structure, tl),
		Tons:   hullTons * (layers - 1) * pct / 100,
	}
}
