package sophont

// Sophont gender, transcribed from Book 3 p.230 (Sophont Creation step 08),
// verified against a rendered image of the page. Every species has a gender
// structure; its Gender Determination Table maps a 2D roll to the gender an
// individual is born into, and each non-base gender carries characteristic
// differences. Assignment timing (08C), gender shifts (08D), and the
// caste-gender relation (08E) are life-cycle detail, deferred until a consumer
// needs them.

import "github.com/philoserf/t5/internal/dice"

// A GenderStructure is one of the five reproductive structures (Book 3 p.230).
type GenderStructure int

const (
	Solitaire GenderStructure = iota // one gender, self-reproducing
	Dual                             // Female/Male
	FMN                              // Female/Male/Neuter (two reproduce)
	EAB                              // Egg Donor/Activator/Bearer (all reproduce)
	Group                            // many numbered genders
)

func (g GenderStructure) String() string {
	return [...]string{"Solitaire", "Dual", "FMN", "EAB", "Group"}[g]
}

// A Difference is a caste- or gender-based adjustment to an individual's
// characteristics (Book 3 pp.229-230), applied when the caste or gender is
// assigned. C2..C5 and C1's low end are flat modifiers in Mods; a large positive
// C1 adjustment is extra Strength dice instead.
type Difference struct {
	C1Dice int    // extra dice added to C1 Strength
	Mods   [5]int // flat modifiers to C1..C5 (Mods[0] is C1's flat part)
}

// A Gender is a species' gender structure: the genders that exist, the
// Determination Table (a 2D roll, index 2-12, gives an individual's gender), and
// the per-gender characteristic differences (the base gender has none).
type Gender struct {
	Structure   GenderStructure
	Genders     []string
	Table       [13]string
	Differences map[string]Difference
}

// gendersFor lists the gender types of a structure, base gender first.
var gendersFor = map[GenderStructure][]string{
	Solitaire: {"Solo"},
	Dual:      {"Female", "Male"},
	FMN:       {"Female", "Male", "Neuter"},
	EAB:       {"Egg Donor", "Activator", "Bearer"},
	Group:     {"One", "Two", "Three", "Four", "Five", "Six"},
}

// structureByFlux is the Structure column of chart 08A, indexed by Flux+5.
var structureByFlux = []GenderStructure{
	Solitaire, Solitaire, EAB, EAB, Dual, Dual, Dual, FMN, FMN, Group, Group,
}

// genderColumns are the per-structure gender columns of chart 08A, each indexed
// by Flux+5.
var genderColumns = map[GenderStructure][]string{
	Solitaire: {
		"Solo",
		"Solo",
		"Solo",
		"Solo",
		"Solo",
		"Solo",
		"Solo",
		"Solo",
		"Solo",
		"Solo",
		"Solo",
	},
	Dual: {
		"Female",
		"Female",
		"Female",
		"Female",
		"Male",
		"Female",
		"Male",
		"Male",
		"Male",
		"Female",
		"Male",
	},
	EAB: {
		"Egg Donor",
		"Egg Donor",
		"Egg Donor",
		"Activator",
		"Egg Donor",
		"Activator",
		"Bearer",
		"Bearer",
		"Bearer",
		"Activator",
		"Bearer",
	},
	FMN: {
		"Female",
		"Female",
		"Female",
		"Male",
		"Female",
		"Male",
		"Neuter",
		"Neuter",
		"Neuter",
		"Male",
		"Neuter",
	},
	Group: {"Six", "Six", "Four", "Four", "Two", "One", "Three", "Five", "Five", "Six", "Six"},
}

// genderDifference reads the Gender-Based Differences table (chart 08B) at a
// Flux. Note C1 is flat at +1/+2 but gains dice at +3/+4/+5.
func genderDifference(flux int) Difference {
	flux = clamp(flux, -5, 5)
	switch {
	case flux <= -2:
		return Difference{Mods: [5]int{flux, flux, flux, flux, flux}}
	case flux <= 0:
		return Difference{} // -1, 0: no change
	case flux == 1:
		return Difference{Mods: [5]int{1, 0, 0, 0, 0}}
	case flux == 2:
		return Difference{Mods: [5]int{2, 2, 2, 2, 2}}
	default: // +3/+4/+5: C1 gains (flux-2) dice, C2-C5 flat +flux
		return Difference{C1Dice: flux - 2, Mods: [5]int{0, flux, flux, flux, flux}}
	}
}

// rollGender rolls a species' gender structure and builds its Determination
// Table: entry 2 is Gender 1, entry 3 is Gender 2 (for multi-gender structures),
// and entries 4-12 are rolled on the structure's column. Each non-base gender
// then rolls its characteristic differences.
func rollGender(r *dice.Roller) Gender {
	structure := structureByFlux[clamp(r.Flux(), -5, 5)+5]
	genders := gendersFor[structure]
	col := genderColumns[structure]

	var table [13]string
	table[2] = genders[0]
	if len(genders) > 1 {
		table[3] = genders[1]
	} else {
		table[3] = genders[0]
	}
	for entry := 4; entry <= 12; entry++ {
		table[entry] = col[clamp(r.Flux(), -5, 5)+5]
	}

	diffs := make(map[string]Difference, len(genders)-1)
	for _, g := range genders[1:] {
		diffs[g] = genderDifference(r.Flux())
	}
	return Gender{Structure: structure, Genders: genders, Table: table, Differences: diffs}
}
