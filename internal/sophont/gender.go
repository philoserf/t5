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

// Gender structures (Book 3 p. 230).
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

// genderC1 reads the C1 column of the Gender-Based Differences table (chart 08B)
// at a Flux, returning extra Strength dice and a flat modifier (in that order)
// — never both.
// C1 is flat at +1/+2 but gains dice at +3/+4/+5 (the printed cells are "+1D",
// "+2D", "+3D", i.e. Flux-2 dice).
func genderC1(flux int) (int, int) {
	flux = clamp(flux, -5, 5)
	switch {
	case flux <= -2:
		return 0, flux
	case flux <= 0:
		return 0, 0 // -1, 0: no change
	case flux <= 2:
		return 0, flux // +1, +2: flat
	default:
		return flux - 2, 0 // +3/+4/+5: +1D/+2D/+3D
	}
}

// rollGenderDifference rolls one gender's characteristic differences on chart
// 08B: "Roll once within each Gender for each Characteristic. C5 is Ins (not
// Edu or Tra)." (Book 3 p.230) — an independent Flux roll per characteristic,
// each read from its own column, except that the C5 column applies only to a
// species whose C5 is the Instinct analog. The five columns are printed
// separately precisely because C1 differs.
//
// Chart 06B carries the identical asymmetry for the die counts — "C5 (Ins): Use
// this table; for Edu or Tra, use 2D" — and rolls nothing for an Edu/Tra C5, so
// this follows suit and draws no C5 Flux for one. The book does not say outright
// whether the roll is skipped or merely discarded; only the dice stream, not the
// outcome, distinguishes the two readings.
func rollGenderDifference(r *dice.Roller, c5IsIns bool) Difference {
	var d Difference

	d.C1Dice, d.Mods[0] = genderC1(r.Flux())
	for i := 1; i < 4; i++ { // C2..C4
		d.Mods[i] = differenceMod(r.Flux())
	}

	if c5IsIns {
		d.Mods[4] = differenceMod(r.Flux())
	}

	return d
}

// autoFillsGenderTwo reports whether chart 08A pins Gender 2 at table entry 3.
// The footnote conditions it on the three named multi-gender structures — "If
// Dual, FMN, or EAB, enter Gender 2 (Male, Activator) on entry line 3" (Book 3
// p.230) — and on nothing else. Solitaire has no Gender 2, and for Group the
// exemption is load-bearing: the book's own Group example, the Rem, "have a
// gender structure 36145 ... Note that Gender Two has evolutionarily dropped
// out" (p.219). A pinned entry 3 gives Two a floor of 2/36 and makes the Rem
// unreachable, so a Group rolls entry 3 like every other entry.
func autoFillsGenderTwo(s GenderStructure) bool {
	return s == Dual || s == FMN || s == EAB
}

// rollGender rolls a species' gender structure and builds its Determination
// Table: entry 2 is Gender 1, entry 3 is Gender 2 for Dual/FMN/EAB, and every
// remaining entry is rolled on the structure's column. Each non-base gender then
// rolls its characteristic differences; c5IsIns says whether the species' C5 is
// the Instinct analog, which chart 08B requires for a C5 difference.
func rollGender(r *dice.Roller, c5IsIns bool) Gender {
	structure := structureByFlux[clamp(r.Flux(), -5, 5)+5]
	genders := gendersFor[structure]
	col := genderColumns[structure]

	var table [13]string

	table[2] = genders[0]
	first := 3

	if autoFillsGenderTwo(structure) {
		table[3] = genders[1]
		first = 4
	}

	for entry := first; entry <= 12; entry++ {
		table[entry] = col[clamp(r.Flux(), -5, 5)+5]
	}

	diffs := make(map[string]Difference, len(genders)-1)
	for _, g := range genders[1:] {
		diffs[g] = rollGenderDifference(r, c5IsIns)
	}

	return Gender{Structure: structure, Genders: genders, Table: table, Differences: diffs}
}
