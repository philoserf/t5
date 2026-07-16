package sophont

// Sophont caste, transcribed from Book 3 p.229 (Sophont Creation step 07),
// verified against a rendered image of the page. Caste applies only to a species
// whose C6 is the Caste characteristic. Its Caste Generation Table maps a 2D roll
// to the caste an individual is born into, and each non-Common caste carries
// characteristic differences.
//
// The Skilled structure is a special case: each member instead draws a skill
// from the Caste Skills list (Chart 12), which is part of the deferred physical
// tier, so a Skilled species records only its structure here. Caste shift (07D)
// and assignment timing (07E) are life-cycle detail, likewise deferred.

import "github.com/philoserf/t5/internal/dice"

// A CasteStructure is one of the six caste organizations (Book 3 p.229).
type CasteStructure int

const (
	Body CasteStructure = iota
	Economic
	Family
	Military
	Social
	Skilled
)

func (c CasteStructure) String() string {
	return [...]string{"Body", "Economic", "Family", "Military", "Social", "Skilled"}[c]
}

// casteStructures maps a 1D roll (index = die-1) to a structure (chart 07A).
var casteStructures = []CasteStructure{Body, Economic, Family, Military, Social, Skilled}

// A Caste is a species' caste structure: the Generation Table (a 2D roll, index
// 2-12, gives an individual's caste) and the per-caste characteristic differences
// (Common has none). A Skilled species leaves Table and Differences empty.
type Caste struct {
	Structure   CasteStructure
	Table       [13]string
	Differences map[string]Difference
}

// casteColumns are the structure columns of chart 07B, indexed by Flux+5. The
// "=Gender" and "=Special" cells are substitution markers resolved when the table
// is built.
var casteColumns = map[CasteStructure][]string{
	Body: {
		"Healer",
		"=Gender",
		"Antibody",
		"Sensor",
		"Memory",
		"Muscle",
		"Muscle",
		"Muscle",
		"Voice",
		"=Special",
		"Claw",
	},
	Economic: {
		"Innovator",
		"=Gender",
		"Guard",
		"Researcher",
		"Artisan",
		"Laborer",
		"Craftsman",
		"Clerk",
		"Manager",
		"=Special",
		"Entrepreneur",
	},
	Family: {
		"Healer",
		"=Gender",
		"Defender",
		"Caregiver",
		"Caregiver",
		"Breadwinner",
		"Breadwinner",
		"Breadwinner",
		"Uncle",
		"=Special",
		"Leader",
	},
	Military: {
		"Medic",
		"=Gender",
		"Aide",
		"Scout",
		"Specialist",
		"Soldier",
		"Technician",
		"Warrior",
		"Leader",
		"=Special",
		"Staff",
	},
	Social: {
		"Artist",
		"=Gender",
		"Enforcer",
		"Drone",
		"Artist",
		"Unit",
		"Unit",
		"Unit",
		"Patron",
		"=Special",
		"Entertainer",
	},
}

// specialColumn is the "6 Special" column of chart 07B (Flux+5), used for
// =Special re-rolls.
var specialColumn = []string{
	"DeMinimis", "Useless", "Advisor Minus", "Instructor", "Shaman",
	"Expendable", "Defective", "Valuable", "Advisor Plus", "Sport", "Vice-Leader",
}

// casteUnique is the X-row Unique caste (chart 07B), auto-inserted at entry 12.
var casteUnique = map[CasteStructure]string{
	Body: "Brain", Economic: "Director", Family: "Archon", Military: "General", Social: "Ruler",
}

// casteDifference reads the Caste-Based Differences table (chart 07C) at a Flux.
// C1 gains dice on any positive roll; -1/0/+1 impose no change.
func casteDifference(flux int) Difference {
	flux = clamp(flux, -5, 5)
	switch {
	case flux <= -2:
		return Difference{Mods: [5]int{flux, flux, flux, flux, flux}}
	case flux <= 1:
		return Difference{} // -1, 0, +1: no change
	default: // +2..+5: C1 gains `flux` dice, C2-C5 flat +flux
		return Difference{C1Dice: flux, Mods: [5]int{0, flux, flux, flux, flux}}
	}
}

// rollCaste rolls a species' caste structure and builds its Generation Table:
// entry 7 is the Common caste, entry 12 the Unique caste, and every other entry
// is rolled on the structure column (resolving =Gender against the parallel
// gender table and =Special against the Special column). Each distinct non-Common
// caste then rolls its characteristic differences.
func rollCaste(r *dice.Roller, gender Gender) Caste {
	structure := casteStructures[r.Die()-1]
	if structure == Skilled {
		return Caste{Structure: Skilled} // per-member skills deferred (Chart 12)
	}
	col := casteColumns[structure]
	common := col[5] // the Flux-0 (Common) caste

	var table [13]string
	var distinct []string
	seen := map[string]bool{}
	for entry := 2; entry <= 12; entry++ {
		var name string
		switch entry {
		case 7:
			name = common
		case 12:
			name = casteUnique[structure]
		default:
			name = rollCasteCell(r, col, gender, entry)
		}
		table[entry] = name
		if name != common && !seen[name] {
			seen[name] = true
			distinct = append(distinct, name)
		}
	}

	diffs := make(map[string]Difference, len(distinct))
	for _, name := range distinct {
		diffs[name] = casteDifference(r.Flux())
	}
	return Caste{Structure: structure, Table: table, Differences: diffs}
}

// rollCasteCell rolls one caste-table entry, resolving the substitution markers.
func rollCasteCell(r *dice.Roller, col []string, gender Gender, entry int) string {
	switch cell := col[clamp(r.Flux(), -5, 5)+5]; cell {
	case "=Gender":
		return gender.Table[entry] // the parallel gender at the same 2D slot
	case "=Special":
		return specialColumn[clamp(r.Flux(), -5, 5)+5]
	default:
		return cell
	}
}
