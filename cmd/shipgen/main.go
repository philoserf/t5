// Command shipgen designs Traveller5 Adventure Class Ships and prints their ship
// card and Quick Ship Profile.
//
// Usage:
//
//	shipgen [-hull L -tl N -config C -maneuver L -jump L -power L -mission S]
//	        [-weapon "name[:mount[:range]],..."] [-n N] [-seed V]
//
// With -hull, a specific ship is designed from the flags; without it, a random
// ship is generated (-n and -seed apply). Drive/hull letters are A-Z; config is
// one of C B P U S A L. Weapons name a model (and optionally a mount and a range
// to build it for), e.g. -weapon "beamlaser:T1:orbit,sandcaster" — a hull carries
// one mount per 100 tons.
package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/philoserf/t5/internal/cli"
	"github.com/philoserf/t5/internal/shipgen"
)

func main() {
	hull := flag.String("hull", "", "hull size letter A-Z (blank = random ship)")
	tl := flag.Int("tl", 12, "tech level")
	configLetter := flag.String("config", "S", "hull config letter C/B/P/U/S/A/L")
	structure := flag.String(
		"structure",
		"plate",
		"hull structure (plate/shell/polymer/feni/organic/charged)",
	)
	armorLayers := flag.Int("armor", 1, "armor layers")
	maneuver := flag.String("maneuver", "A", "maneuver drive letter (blank = none)")
	jump := flag.String("jump", "A", "jump drive letter (blank = none)")
	power := flag.String("power", "A", "power plant letter (blank = none)")
	mission := flag.String("mission", "X", "QSP mission code")
	weapons := flag.String(
		"weapon",
		"",
		`weapon(s) to mount, comma-separated, each "name[:mount[:range]]" (e.g. "beamlaser:T1:orbit,sandcaster")`,
	)
	defenses := flag.String(
		"defense",
		"",
		`defense(s) to install, comma-separated, each "name[:mount[:range]]" (e.g. "blackglobe,nucleardamper")`,
	)
	n, r := cli.SeededRoller("ships")

	if *hull == "" {
		for i := range n {
			if i > 0 {
				fmt.Println()
			}

			fmt.Println(shipgen.Generate(r))
		}

		return
	}

	spec, err := specFromFlags(flags{
		hull: *hull, tl: *tl, config: *configLetter, structure: *structure,
		armor: *armorLayers, maneuver: *maneuver, jump: *jump, power: *power, mission: *mission,
		weapons: *weapons, defenses: *defenses,
	})
	if err != nil {
		cli.Fatalf("%v", err)
	}
	// The spec is fixed, so the ship is too: design it once and print it n times.
	ship := shipgen.Design(spec)

	for i := range n {
		if i > 0 {
			fmt.Println()
		}

		fmt.Println(ship)
	}
}

type flags struct {
	hull                           string
	tl                             int
	config, structure              string
	armor                          int
	maneuver, jump, power, mission string
	weapons, defenses              string
}

// specFromFlags builds a ShipSpec from the CLI flags, reporting an unknown hull,
// config, or structure as an error for the caller to fail on. (An infeasible but
// well-formed design is not an error: Design is total and reports it in
// Ship.Problems.)
func specFromFlags(f flags) (shipgen.ShipSpec, error) {
	hullOrd := letterOrdinal(f.hull)
	if hullOrd == 0 {
		return shipgen.ShipSpec{}, fmt.Errorf(
			"invalid hull %q (want a letter A-Z, no I or O)",
			f.hull,
		)
	}

	config, ok := shipgen.ConfigByLetter(f.config)
	if !ok {
		return shipgen.ShipSpec{}, fmt.Errorf("invalid config %q (want C/B/P/U/S/A/L)", f.config)
	}

	structure, ok := shipgen.StructureByName(f.structure)
	if !ok {
		return shipgen.ShipSpec{}, fmt.Errorf(
			"invalid structure %q (want plate/shell/polymer/feni/organic/charged)",
			f.structure,
		)
	}

	maneuver, err := driveSpec("maneuver", f.maneuver)
	if err != nil {
		return shipgen.ShipSpec{}, err
	}

	jump, err := driveSpec("jump", f.jump)
	if err != nil {
		return shipgen.ShipSpec{}, err
	}

	power, err := driveSpec("power", f.power)
	if err != nil {
		return shipgen.ShipSpec{}, err
	}

	weapons, err := weaponSpecs(f.weapons)
	if err != nil {
		return shipgen.ShipSpec{}, err
	}

	defenseList, err := defenseSpecs(f.defenses)
	if err != nil {
		return shipgen.ShipSpec{}, err
	}

	return shipgen.ShipSpec{
		Mission: f.mission, TL: f.tl, HullLetter: hullOrd, Config: config,
		Structure: structure, ArmorLayers: f.armor,
		Maneuver:  maneuver,
		Jump:      jump,
		Power:     power,
		Weapons:   weapons,
		Defenses:  defenseList,
		FuelScoop: true,
	}, nil
}

// An installation is one entry of -weapon or -defense: "name[:mount[:range]]". The
// two flags have the same shape, so they share a parser; only the lookup of the
// name and the default spec differ.
type installation struct {
	name  string
	mount Mount
	rng   Range
}

// Mount and Range here are the optional halves of an entry — "unset" is a real
// state, meaning "use the component's own default".
type (
	Mount struct {
		set   bool
		value shipgen.Mount
	}
	Range struct {
		set   bool
		value shipgen.Range
	}
)

// parseInstallations splits a comma-separated list of "name[:mount[:range]]"
// entries and resolves the mount and range of each, leaving the name to the caller
// (which knows whether it is looking up a weapon or a defense).
func parseInstallations(list, kind string) ([]installation, error) {
	if list == "" {
		return nil, nil
	}

	var out []installation

	for entry := range strings.SplitSeq(list, ",") {
		parts := strings.Split(strings.TrimSpace(entry), ":")
		if len(parts) > 3 {
			return nil, fmt.Errorf(
				"%s %q has too many fields (want name[:mount[:range]])",
				kind,
				entry,
			)
		}

		inst := installation{name: parts[0]}
		if len(parts) > 1 && parts[1] != "" {
			m, ok := shipgen.MountByCode(parts[1])
			if !ok {
				return nil, fmt.Errorf("unknown mount %q (known: %s)",
					parts[1], strings.Join(shipgen.MountCodes(), ", "))
			}

			inst.mount = Mount{set: true, value: m}
		}

		if len(parts) > 2 && parts[2] != "" {
			r, ok := shipgen.RangeByName(parts[2])
			if !ok {
				return nil, fmt.Errorf("unknown range %q (known: %s)",
					parts[2], strings.Join(shipgen.RangeNames(), ", "))
			}

			inst.rng = Range{set: true, value: r}
		}

		out = append(out, inst)
	}

	return out, nil
}

// weaponSpecs parses the -weapon flag. The mount defaults to the weapon's own
// minimum and the range to the standard rung of its ladder, so "beamlaser" alone is
// a valid weapon and "beamlaser:T1:orbit" spells the installation out.
//
// An infeasible but well-formed weapon is not an error here — Design reports it in
// Ship.Problems, the same as an underpowered plant.
func weaponSpecs(list string) ([]shipgen.WeaponSpec, error) {
	entries, err := parseInstallations(list, "weapon")
	if err != nil {
		return nil, err
	}

	var specs []shipgen.WeaponSpec

	for _, e := range entries {
		model, ok := shipgen.WeaponByName(e.name)
		if !ok {
			return nil, fmt.Errorf("unknown weapon %q (known: %s)",
				e.name, strings.Join(shipgen.WeaponNames(), ", "))
		}

		spec := shipgen.DefaultWeapon(model)
		if e.mount.set {
			spec.Mount = e.mount.value
		}

		if e.rng.set {
			spec.Range = e.rng.value
		}

		specs = append(specs, spec)
	}

	return specs, nil
}

// defenseSpecs parses the -defense flag, the same shape as -weapon, defaulting to
// the Bolt-In mount (which needs no hardpoint) at the standard range.
func defenseSpecs(list string) ([]shipgen.DefenseSpec, error) {
	entries, err := parseInstallations(list, "defense")
	if err != nil {
		return nil, err
	}

	var specs []shipgen.DefenseSpec

	for _, e := range entries {
		model, ok := shipgen.DefenseByName(e.name)
		if !ok {
			return nil, fmt.Errorf("unknown defense %q (known: %s)",
				e.name, strings.Join(shipgen.DefenseNames(), ", "))
		}

		spec := shipgen.DefaultDefense(model)
		if e.mount.set {
			spec.Mount = e.mount.value
		}

		if e.rng.set {
			spec.Range = e.rng.value
		}

		specs = append(specs, spec)
	}

	return specs, nil
}

// driveSpec returns a standard DriveSpec for a size letter. Only a blank means
// "no such drive" — a non-starship has no jump drive at all. Anything else that
// is not a hull/drive letter is a typo, and a typo must not quietly design a
// driveless ship: it is an error, exactly as the same bad value is for -hull.
func driveSpec(name, letter string) (*shipgen.DriveSpec, error) {
	if letter == "" {
		return nil, nil //nolint:nilnil // a blank drive letter legitimately means "no drive"
	}

	ord := letterOrdinal(letter)
	if ord == 0 {
		return nil, fmt.Errorf(
			"invalid %s drive %q (want a letter A-Z, no I or O, or blank for none)",
			name,
			letter,
		)
	}

	return &shipgen.DriveSpec{Letter: ord}, nil
}

// letterOrdinal maps a hull/drive letter A-Z (no I/O) to its ordinal 1..24, or 0.
func letterOrdinal(s string) int {
	if len(s) != 1 {
		return 0
	}

	return shipgen.LetterOrdinal(s[0])
}
