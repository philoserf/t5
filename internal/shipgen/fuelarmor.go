package shipgen

// Fuel and armor (Book 2 pp.75, 79).

// Fitting costs (Book 2 p.79 F): the fuel scoop and purifier a wilderness
// refueler carries, each 1 ton.
const (
	fuelScoopCost    = 100_000   // KCr100
	fuelPurifierCost = 1_000_000 // MCr1
	liquidHydrogenCr = 500       // Cr500 per ton
)

// driveFuel is one drive's own fuel demand in tons (Book 2 p.79): jump fuel is
// P*hull/10 per jump and power-plant operations fuel is P*hull/100 (one month's
// supply), each scaled by the stage's fuel multiplier — p.127 states the same
// thing the other way round, "(H x P/10) divided by Efficiency", the Fuel column
// being the reciprocal of the Efficiency column. Maneuver drives draw no fuel of
// their own; their consumption is part of power-plant operations.
func driveFuel(hullTons int, d *Drive) int {
	if d == nil {
		return 0
	}

	switch d.Kind {
	case Jump:
		return fuelMul(d.Potential*hullTons/10, d.Stage)
	case Power:
		return fuelMul(d.Potential*hullTons/100, d.Stage)
	default:
		// Maneuver burns no fuel of its own: p.79 folds its consumption into the
		// power plant's operations fuel, which the Power case already charges.
		return 0
	}
}

// fuel computes a ship's fuel tankage and cost (Book 2 p.79) as the sum of what
// each drive demands. It records each drive's share on the Drive itself, since
// the split is real information — the p.127 example table prints a Fuel Tons
// column per drive — and the ship total is the only other place it appears.
// Cost is the liquid hydrogen plus any scoop/purifier fittings.
func fuel(hullTons int, jump, power *Drive, scoop, purifier bool) Fuel {
	tankage := 0

	for _, d := range []*Drive{jump, power} {
		if d == nil {
			continue
		}

		d.Fuel = driveFuel(hullTons, d)
		tankage += d.Fuel
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
func structureAV(structure HullMaterial, tl int) int {
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
func armor(tl, hullTons int, structure HullMaterial, layers int) Armor {
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
