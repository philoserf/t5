package sophont

// Sophont characteristics, transcribed from Book 3 p.228 (Sophont Creation
// step 06), verified against a rendered image of the page.
//
// A sophont species always has exactly six characteristics C1..C6, but the
// identity of four of them varies: C1 is always Str and C4 is always Int, while
// C2/C3/C5/C6 are each chosen by a Flux roll (chart 06A). A second Flux roll per
// characteristic (chart 06B) fixes how many dice that characteristic rolls when
// an individual of the species is generated, so the species template records a
// die count, not a value. Humans come out all-2D with the Genetic Profile
// "SDEIES" — the reference every other species is measured against.

import "github.com/philoserf/t5/internal/dice"

// A CharName is one of the fourteen characteristic identities a sophont slot may
// take (Book 3 p.228). The four Human-standard names (Dex, End, Edu, Soc) sit
// beside their non-human analogs; a species is Human only if all six slots hold
// Human-standard names.
type CharName int

// The sophont characteristic slots C1..C6.
const (
	Str CharName = iota // C1, always
	Dex                 // C2 Human
	Agi                 // C2 analog
	Gra                 // C2 analog
	End                 // C3 Human
	Sta                 // C3 analog
	Vig                 // C3 analog
	Int                 // C4, always
	Edu                 // C5 Human
	Tra                 // C5 analog
	Ins                 // C5 analog
	Soc                 // C6 Human
	Cha                 // C6 analog
	Cas                 // C6 analog (Caste)
)

// charInfo carries a characteristic's abbreviation and its Genetic Profile
// letter. The GP letter is position-independent here; the two C6 names that
// begin with "C" (Cha, Cas) use distinct letters — Cha=C, Cas=K.
var charInfo = map[CharName]struct {
	abbrev string
	gp     byte
}{
	Str: {"Str", 'S'},
	Dex: {"Dex", 'D'},
	Agi: {"Agi", 'A'},
	Gra: {"Gra", 'G'},
	End: {"End", 'E'},
	Sta: {"Sta", 'S'},
	Vig: {"Vig", 'V'},
	Int: {"Int", 'I'},
	Edu: {"Edu", 'E'},
	Tra: {"Tra", 'T'},
	Ins: {"Ins", 'I'},
	Soc: {"Soc", 'S'},
	Cha: {"Cha", 'C'},
	Cas: {"Cas", 'K'},
}

// String returns the characteristic's three-letter abbreviation, e.g. "Str".
func (c CharName) String() string {
	if info, ok := charInfo[c]; ok {
		return info.abbrev
	}

	return "?"
}

// gpLetter returns the characteristic's Genetic Profile letter.
func (c CharName) gpLetter() byte {
	if info, ok := charInfo[c]; ok {
		return info.gp
	}

	return '?'
}

// A CharSpec is one slot of a species' characteristic profile: which
// characteristic it is, and how many dice it rolls at character generation.
type CharSpec struct {
	Name CharName
	Dice int
}

// nameCols holds the variable-slot name tables (chart 06A), indexed by Flux+5
// (Flux -5..+5). C1 and C4 are omitted — they are always Str and Int.
var nameCols = map[int][]CharName{
	2: {Agi, Agi, Agi, Agi, Dex, Dex, Dex, Gra, Gra, Gra, Gra}, // C2
	3: {Sta, Sta, Sta, Sta, End, End, End, Vig, Vig, Vig, Vig}, // C3
	5: {Ins, Ins, Ins, Ins, Edu, Edu, Edu, Tra, Tra, Tra, Tra}, // C5
	6: {Cas, Cas, Cas, Soc, Soc, Soc, Soc, Cha, Cha, Cha, Cha}, // C6
}

// diceCols holds the die-count table (chart 06B), indexed by Flux+5 (the table
// runs Flux -5..+6, so twelve entries). Only Str scales past 3D; the analog and
// mental columns cap at 3D and Soc stays 2D.
var diceCols = map[CharName][]int{
	Str: {1, 1, 2, 2, 2, 2, 3, 4, 5, 6, 7, 8}, // C1
	Dex: {1, 1, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3}, // C2 column
	End: {1, 1, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3}, // C3 column
	Int: {1, 1, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3}, // C4 column
	Edu: {1, 1, 2, 2, 2, 2, 2, 2, 2, 3, 3, 3}, // C5 column
	Soc: {1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}, // C6 column
}

// rollCharacteristics rolls a species' six-slot characteristic profile and its
// Genetic Profile. C1=Str and C4=Int are fixed; the other four names come from
// chart 06A (with the locomotion DM: Flyer favors the analogs Agi/Sta/Ins,
// Swimmer/Diver favors Gra/Vig/Tra). Each die count comes from chart 06B: the
// physical slots (C1/C2/C3) are indexed by Flux+EnvironDM, C3 further shifted by
// the carnivore sub-niche (+2 Chaser, -2 Pouncer); mental slots by Flux alone.
// C5 Edu/Tra are a flat 2D (no table roll); C6 Caste is left at a placeholder 2D
// until caste generation (a later phase) fixes it.
func rollCharacteristics(r *dice.Roller, env Environment) ([6]CharSpec, string) {
	var names [6]CharName

	names[0] = Str // C1, always
	names[1] = rollName(r, 2, env.locomotionMod())
	names[2] = rollName(r, 3, env.locomotionMod())
	names[3] = Int // C4, always
	names[4] = rollName(r, 5, env.locomotionMod())
	names[5] = rollName(r, 6, 0) // C6 takes no locomotion DM

	var chars [6]CharSpec

	gp := make([]byte, 6)

	for i, name := range names {
		chars[i] = CharSpec{Name: name, Dice: rollDice(r, i, name, env)}
		gp[i] = name.gpLetter()
	}

	return chars, string(gp)
}

// rollName rolls one variable name slot (2, 3, 5, or 6) with a name DM.
func rollName(r *dice.Roller, slot, mod int) CharName {
	idx := clamp(r.Flux()+mod, -5, 5) + 5

	return nameCols[slot][idx]
}

// rollDice returns the die count for slot i holding characteristic name.
func rollDice(r *dice.Roller, i int, name CharName, env Environment) int {
	switch i {
	case 0: // C1 Str — physical
		return diceLookup(Str, r.Flux()+env.EnvironDM)
	case 1: // C2 — physical
		return diceLookup(Dex, r.Flux()+env.EnvironDM)
	case 2: // C3 — physical, shifted by the carnivore sub-niche
		return diceLookup(End, r.Flux()+env.EnvironDM+env.c3NicheMod())
	case 3: // C4 Int — mental
		return diceLookup(Int, r.Flux())
	case 4: // C5
		// Chart 06B is explicit and asymmetric: "C5 (Ins): Use this table; for
		// Edu or Tra, use 2D." Only the Instinct analog rolls the column; the two
		// schooling-based names are a flat 2D regardless of Flux.
		if name == Ins {
			return diceLookup(Edu, r.Flux())
		}

		return 2
	default: // C6
		if name == Cas {
			return 2 // placeholder until caste generation
		}

		return diceLookup(Soc, r.Flux())
	}
}

// diceLookup reads the die count for a column at a Flux index (clamped -5..+6).
// The column is keyed by its Human-standard name (Str/Dex/End/Int/Edu/Soc); the
// first argument selects the column, not the rolled characteristic.
func diceLookup(column CharName, flux int) int {
	return diceCols[column][clamp(flux, -5, 6)+5]
}

// RollValue rolls a characteristic value for the given die count, applying the
// "Rolling Higher Characteristics" rule (Book 3 p.228): 1-3 dice roll normally;
// 4+ dice roll 12 + (dice-2)D so a large pool stays bounded (6D = 12+4D). Used
// when an individual of the species is generated.
func RollValue(r *dice.Roller, numDice int) int {
	if numDice <= 3 {
		return r.Dice(numDice)
	}

	return 12 + r.Dice(numDice-2)
}
