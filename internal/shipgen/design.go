package shipgen

import "fmt"

// Design builds a complete ship from a spec (Book 2 Starship Design Checklist).
// It is total — never an error: infeasibility is reported in Ship.Problems (a
// power plant weaker than its drives, a TL-capped drive, a weapon the hull has no
// mount point for or the yard cannot build, or an over-budget hull). The tonnage
// budget's Payload is the residual left for what the engine does not yet model —
// accommodations and cargo — so a clean ship still lists that tonnage as Payload
// rather than fully spent.
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
	// The hull's configuration caps its thrust however big the drive is: a Cluster
	// is rated for 1G and a Braced hull for 3 (Book 2 p.71). Hull.MaxG has been
	// computed since the first commit and read by nothing, so a Cluster with a huge
	// maneuver drive designed clean and then flew at 8G in a fight.
	if ship.Maneuver != nil && ship.Maneuver.Potential > h.MaxG {
		problems = append(problems, fmt.Sprintf("maneuver drive rated %dG but a %s hull is capped at %dG",
			ship.Maneuver.Potential, h.Config, h.MaxG))
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

	// Weapons: each is designed on its own (DesignWeapon), then the hull has to
	// have somewhere to put them and the yard has to be able to build them.
	for _, ws := range spec.Weapons {
		w := DesignWeapon(ws)
		ship.Weapons = append(ship.Weapons, w)
		problems = append(problems, w.Problems...)
		// A yard cannot build above its own tech level: the weapon's TL is its
		// base shifted by the range and the stage, so a long-range or advanced
		// model can outrun the ship that carries it.
		if w.TL > spec.TL {
			problems = append(problems, fmt.Sprintf("%s is TL-%d, above the ship's TL-%d",
				w.Name(), w.TL, spec.TL))
		}
	}
	// Defenses are built the same way, and mostly ride Bolt-Ins, which need no
	// mount point — a ship can carry screens without giving up a gun.
	for _, ds := range spec.Defenses {
		d := DesignDefense(ds)
		ship.Defenses = append(ship.Defenses, d)
		problems = append(problems, d.Problems...)
		if d.TL > spec.TL {
			problems = append(problems, fmt.Sprintf("%s is TL-%d, above the ship's TL-%d",
				d.Name(), d.TL, spec.TL))
		}
	}
	if p := mountPoints(h, ship.Weapons, ship.Defenses); p != "" {
		problems = append(problems, p)
	}
	used += armamentTonnage(ship.Weapons, ship.Defenses)

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
	for _, w := range ship.Weapons {
		ship.Cost += w.Cost
	}
	for _, d := range ship.Defenses {
		ship.Cost += d.Cost
	}

	ship.Problems = problems
	return ship
}
