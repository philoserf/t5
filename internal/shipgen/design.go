package shipgen

import "fmt"

// Design builds a complete ship from a spec (Book 2 Starship Design Checklist).
// It is total — never an error: infeasibility is reported in Ship.Problems (a
// power plant weaker than its drives, a TL-capped drive, or an over-budget
// hull). The tonnage budget's Payload is the residual left for the payload the
// core engine does not yet model — accommodations, weapons, and cargo — so a
// clean core ship still lists that tonnage as Payload rather than fully spent.
func Design(spec ShipSpec) Ship {
	h := hull(spec.TL, spec.HullLetter, spec.Tons, spec.Config, spec.Structure)
	ship := Ship{Spec: spec, Hull: h}
	var problems []string
	used := 0

	addDrive := func(kind DriveKind, ds *DriveSpec) *Drive {
		if ds == nil {
			return nil
		}
		d, problem := designDrive(kind, *ds, h.Letter, spec.TL)
		used += d.Tons
		if problem != "" {
			problems = append(problems, problem)
		}
		return d
	}
	ship.Maneuver = addDrive(Maneuver, spec.Maneuver)
	ship.Jump = addDrive(Jump, spec.Jump)
	ship.Power = addDrive(Power, spec.Power)

	// The power plant must be at least as capable as the drives it feeds
	// (Book 2 p.78); a jump/maneuver drive with no plant cannot run.
	powered := 0
	if ship.Maneuver != nil {
		powered = max(powered, ship.Maneuver.Potential)
	}
	if ship.Jump != nil {
		powered = max(powered, ship.Jump.Potential)
	}
	switch {
	case powered > 0 && ship.Power == nil:
		problems = append(problems, "drives require a power plant")
	case ship.Power != nil && ship.Power.Potential < powered:
		problems = append(problems, fmt.Sprintf("power plant potential %d is below the drives it feeds (%d)", ship.Power.Potential, powered))
	}

	ship.Fuel = fuel(h.Tons, ship.Jump, ship.Power, spec.FuelScoop, spec.FuelPurifier)
	used += ship.Fuel.Tons
	if spec.FuelScoop { // the scoop and purifier fittings are 1 ton each (Book 2 p.79)
		used++
	}
	if spec.FuelPurifier {
		used++
	}

	ship.Armor = armor(spec.TL, h.Tons, spec.Structure, spec.ArmorLayers)
	used += ship.Armor.Tons

	ship.Tonnage = Budget{Hull: h.Tons, Used: used, Payload: h.Tons - used}
	if ship.Tonnage.Payload < 0 {
		problems = append(problems, fmt.Sprintf("over budget by %dt", -ship.Tonnage.Payload))
	}

	ship.Cost = h.Cost + ship.Fuel.Cost
	for _, d := range []*Drive{ship.Maneuver, ship.Jump, ship.Power} {
		if d != nil {
			ship.Cost += d.Cost
		}
	}

	ship.Problems = problems
	return ship
}
