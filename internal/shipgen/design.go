package shipgen

import "fmt"

// specProblems reports the fields of a ShipSpec that name nothing in the book's
// tables. Design stays total by clamping them (configIndex, stageIndex) rather
// than indexing raw, but a clamp on its own would quietly hand the caller a
// different ship than they asked for — so the substitution is reported here.
// That is the same choice DesignWeapon makes for an unknown model, mount, or
// range: build something, and say what was wrong with the request.
func specProblems(spec ShipSpec) []Problem {
	var problems []Problem

	if !validConfig(spec.Config) {
		problems = append(problems, Problem{Kind: ConfigInvalid, Detail: fmt.Sprintf(
			"configuration %d is not one of Cluster..Lifting; built as %s",
			spec.Config, configIndex(spec.Config),
		)})
	}
	// Hull sizes are the eHex letters as an ordinal, A=1 .. Z=24 (Book 2 p.93).
	// An ordinal outside that range does not panic — it yields a nonsense
	// tonnage and prints "?" — so it is reported rather than substituted, there
	// being no defensible guess at what size was meant.
	if spec.HullLetter < 1 || spec.HullLetter > maxLetter {
		problems = append(problems, Problem{Kind: HullSizeInvalid, Detail: fmt.Sprintf(
			"hull size %d is outside A..Z (1..%d)", spec.HullLetter, maxLetter,
		)})
	}

	// Ordered, not a map: Problems is a rendered list, so its order is part of
	// what a caller sees.
	drives := []struct {
		kind DriveKind
		spec *DriveSpec
	}{{Maneuver, spec.Maneuver}, {Jump, spec.Jump}, {Power, spec.Power}}
	for _, d := range drives {
		if d.spec == nil {
			continue
		}

		if stageIndex(d.spec.Stage) != d.spec.Stage {
			problems = append(problems, Problem{Kind: DriveStageInvalid, Detail: fmt.Sprintf(
				"%s drive stage %d is not one of Standard..Ultimate; built as %s",
				d.kind, d.spec.Stage, Standard,
			)})
		}
		// Book 2 p.77: "No drive may be smaller than the Drive-A of the class."
		// Read as a floor on the size LETTER, not on tonnage — which is what
		// reconciles p.77 with the worked columns on pp.104/127/134, where seven
		// stage-reduced rows print below their class's Drive-A tonnage while all
		// of them remain a Drive-B.
		//
		// Enforced here rather than in designDrive because this is where the
		// package reports a spec field naming no row of the book's tables. A
		// sub-A ordinal is not merely wrong, it inverts: driveTonsBase runs
		// negative, so a Letter of 0 yields -1 tons and -MCr2 — a drive that
		// ADDS budget and frees hull space, which is the phantom the tonnage
		// accounting exists to prevent.
		if !buildableDriveOrd(d.spec.Letter) {
			problems = append(problems, Problem{Kind: DriveSizeInvalid, Detail: fmt.Sprintf(
				"%s drive size %d names no drive: sizes are A..Z (1..%d) and the "+
					"even Nexus pairs A2..Z2 (26..%d)",
				d.kind, d.spec.Letter, maxLetter, 2*maxLetter,
			)})
		}
	}

	return problems
}

// Design builds a complete ship from a spec (Book 2 Starship Design Checklist).
// It is total — never an error, and never a panic: infeasibility is reported in
// Ship.Problems (a power plant weaker than its drives, a TL-capped drive, a
// weapon the hull has no mount point for or the yard cannot build, an
// over-budget hull, or a spec field naming no row of the book's tables). The
// tonnage budget's Payload is the residual left for what the engine does not yet
// model — accommodations and cargo — so a clean ship still lists that tonnage as
// Payload rather than fully spent.
func Design(spec ShipSpec) Ship { //nolint:gocognit,cyclop,funlen // ship-design pipeline (Book 2 pp.30-95)
	h := hull(spec.TL, spec.HullLetter, spec.Tons, spec.Config, spec.Structure)
	ship := Ship{Spec: spec, Hull: h}

	problems := specProblems(spec)

	used := 0

	addDrive := func(kind DriveKind, ds *DriveSpec) *Drive {
		if ds == nil {
			return nil
		}

		d, problem := designDrive(kind, *ds, h.Letter, spec.TL)
		used += d.Tons

		if problem.reported() {
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
		problems = append(
			problems,
			Problem{Kind: DriveExceedsHullG, Detail: fmt.Sprintf(
				"maneuver drive rated %dG but a %s hull is capped at %dG",
				ship.Maneuver.Potential, h.Config, h.MaxG,
			)},
		)
	}

	switch {
	case powered > 0 && ship.Power == nil:
		problems = append(problems, Problem{Kind: NoPowerPlant, Detail: "drives require a power plant"})
	case ship.Power != nil && ship.Power.Potential < powered:
		problems = append(
			problems,
			Problem{Kind: PowerBelowDrives, Detail: fmt.Sprintf(
				"power plant potential %d is below the drives it feeds (%d)",
				ship.Power.Potential,
				powered,
			)},
		)
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
		//
		// Only for something the ship actually carries. A component that already
		// failed to design has its real problem recorded, and a second complaint
		// about the TL it would have had is the double-complaint aboard exists to
		// prevent — two complaints for one mistake, one of them about a ship that
		// was never built.
		if aboard(w.Problems) && w.TL > spec.TL {
			problems = append(problems, Problem{Kind: ComponentAboveShipTL, Detail: fmt.Sprintf(
				"%s is TL-%d, above the ship's TL-%d", w.Name(), w.TL, spec.TL,
			)})
		}
	}
	// Defenses are built the same way, and mostly ride Bolt-Ins, which need no
	// mount point — a ship can carry screens without giving up a gun.
	for _, ds := range spec.Defenses {
		d := DesignDefense(ds)
		ship.Defenses = append(ship.Defenses, d)

		problems = append(problems, d.Problems...)
		if aboard(d.Problems) && d.TL > spec.TL {
			problems = append(problems, Problem{Kind: ComponentAboveShipTL, Detail: fmt.Sprintf(
				"%s is TL-%d, above the ship's TL-%d", d.Name(), d.TL, spec.TL,
			)})
		}
	}

	if p := mountPoints(h, ship.Weapons, ship.Defenses); p != "" {
		problems = append(problems, Problem{Kind: MountsExceedHull, Detail: p})
	}

	used += armamentTonnage(ship.Weapons, ship.Defenses)

	ship.Tonnage = Budget{Hull: h.Tons, Used: used, Payload: h.Tons - used}
	if ship.Tonnage.Payload < 0 {
		problems = append(problems, Problem{Kind: OverBudget, Detail: fmt.Sprintf(
			"over budget by %dt", -ship.Tonnage.Payload,
		)})
	}

	ship.Cost = h.Cost + ship.Fuel.Cost
	for _, d := range []*Drive{ship.Maneuver, ship.Jump, ship.Power} {
		if d != nil {
			ship.Cost += d.Cost
		}
	}

	// Cost is the third accounting of the same component list, alongside
	// mountPoints and armamentTonnage, and asks aboard the same question they do.
	// A refused component is not free of charge by accident — install may already
	// have priced it before the refusal was recorded — so without this the ship
	// pays for armament it does not carry.
	for _, w := range ship.Weapons {
		if aboard(w.Problems) {
			ship.Cost += w.Cost
		}
	}

	for _, d := range ship.Defenses {
		if aboard(d.Problems) {
			ship.Cost += d.Cost
		}
	}

	ship.Problems = problems

	return ship
}
