// Command shipgen designs Traveller5 Adventure Class Ships and prints their ship
// card and Quick Ship Profile.
//
// Usage:
//
//	shipgen [-hull L -tl N -config C -maneuver L -jump L -power L -mission S] [-n N] [-seed V]
//
// With -hull, a specific ship is designed from the flags; without it, a random
// ship is generated (-n and -seed apply). Drive/hull letters are A-Z; config is
// one of C B P U S A L.
package main

import (
	"flag"
	"fmt"

	"github.com/philoserf/t5/internal/cli"
	"github.com/philoserf/t5/internal/shipgen"
)

func main() {
	hull := flag.String("hull", "", "hull size letter A-Z (blank = random ship)")
	tl := flag.Int("tl", 12, "tech level")
	configLetter := flag.String("config", "S", "hull config letter C/B/P/U/S/A/L")
	structure := flag.String("structure", "plate", "hull structure (plate/shell/polymer/feni/organic/charged)")
	armorLayers := flag.Int("armor", 1, "armor layers")
	maneuver := flag.String("maneuver", "A", "maneuver drive letter (blank = none)")
	jump := flag.String("jump", "A", "jump drive letter (blank = none)")
	power := flag.String("power", "A", "power plant letter (blank = none)")
	mission := flag.String("mission", "X", "QSP mission code")
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
	})
	if err != nil {
		cli.Fatalf("%v", err)
	}
	for i := range n {
		if i > 0 {
			fmt.Println()
		}
		fmt.Println(shipgen.Design(spec))
	}
}

type flags struct {
	hull                           string
	tl                             int
	config, structure              string
	armor                          int
	maneuver, jump, power, mission string
}

// specFromFlags builds a ShipSpec from the CLI flags, reporting an unknown hull,
// config, or structure as an error for the caller to fail on. (An infeasible but
// well-formed design is not an error: Design is total and reports it in
// Ship.Problems.)
func specFromFlags(f flags) (shipgen.ShipSpec, error) {
	hullOrd := letterOrdinal(f.hull)
	if hullOrd == 0 {
		return shipgen.ShipSpec{}, fmt.Errorf("invalid hull %q (want a letter A-Z, no I or O)", f.hull)
	}
	config, ok := shipgen.ConfigByLetter(f.config)
	if !ok {
		return shipgen.ShipSpec{}, fmt.Errorf("invalid config %q (want C/B/P/U/S/A/L)", f.config)
	}
	structure, ok := shipgen.StructureByName(f.structure)
	if !ok {
		return shipgen.ShipSpec{}, fmt.Errorf("invalid structure %q (want plate/shell/polymer/feni/organic/charged)", f.structure)
	}
	return shipgen.ShipSpec{
		Mission: f.mission, TL: f.tl, HullLetter: hullOrd, Config: config,
		Structure: structure, ArmorLayers: f.armor,
		Maneuver:  driveSpec(f.maneuver),
		Jump:      driveSpec(f.jump),
		Power:     driveSpec(f.power),
		FuelScoop: true,
	}, nil
}

// driveSpec returns a standard DriveSpec for a size letter, or nil for a blank.
func driveSpec(letter string) *shipgen.DriveSpec {
	if ord := letterOrdinal(letter); ord != 0 {
		return &shipgen.DriveSpec{Letter: ord}
	}
	return nil
}

// letterOrdinal maps a hull/drive letter A-Z (no I/O) to its ordinal 1..24, or 0.
func letterOrdinal(s string) int {
	if len(s) != 1 {
		return 0
	}
	return shipgen.LetterOrdinal(s[0])
}
