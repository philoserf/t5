package vehiclegen

import "github.com/philoserf/t5/internal/dice"

// Values are the calculated VehicleMaker columns. Cost is in KCr.
type Values struct {
	TechLevel int
	Tons      float64
	Speed     int
	Load      float64
	Armor     int
	Cage      int
	Flash     int
	Radiation int
	Sound     int
	Psi       int
	Insulated int
	Sealed    int
	CostKCr   float64
}

// Operation is one chart-cell operation: add, multiply, divide, or replace.
type Operation struct {
	Add float64
	Mul float64
	Set *float64
}

func (o Operation) apply(value float64) float64 {
	if o.Set != nil {
		value = *o.Set
	}

	value += o.Add
	if o.Mul != 0 {
		value *= o.Mul
	}

	return value
}

// Modifier is one row of Charts 10-13. Blank cells are zero Operations.
type Modifier struct {
	Code, Descriptor string
	Note             string
	TechLevel        Operation
	Tons             Operation
	Speed            Operation
	Load             Operation
	Armor            Operation
	Cage             Operation
	Flash            Operation
	Radiation        Operation
	Sound            Operation
	Psi              Operation
	Insulated        Operation
	Sealed           Operation
	CostKCr          Operation
}

// Enhancer returns a named Chart 12 Bulk, Stage, Environment, Option, or More
// Option row. Codes are the chart codes where printed and stable descriptive
// codes otherwise (for example "Protected" and "VTOL").
func Enhancer(code string) (Modifier, bool) {
	row, ok := enhancerRows[code]

	return row, ok
}

// Apply performs chart rows in selection order and returns their column totals.
// It is deterministic and contains no dice; Generate only selects rows.
func Apply(rows ...Modifier) Values {
	var v Values

	for _, row := range rows {
		v.TechLevel = int(row.TechLevel.apply(float64(v.TechLevel)))
		v.Tons = row.Tons.apply(v.Tons)
		v.Speed = int(row.Speed.apply(float64(v.Speed)))
		v.Load = row.Load.apply(v.Load)
		v.Armor = int(row.Armor.apply(float64(v.Armor)))
		v.Cage = int(row.Cage.apply(float64(v.Cage)))
		v.Flash = int(row.Flash.apply(float64(v.Flash)))
		v.Radiation = int(row.Radiation.apply(float64(v.Radiation)))
		v.Sound = int(row.Sound.apply(float64(v.Sound)))
		v.Psi = int(row.Psi.apply(float64(v.Psi)))
		v.Insulated = int(row.Insulated.apply(float64(v.Insulated)))
		v.Sealed = int(row.Sealed.apply(float64(v.Sealed)))
		v.CostKCr = row.CostKCr.apply(v.CostKCr)
	}

	return v
}

func add(n float64) Operation { return Operation{Add: n} }
func mul(n float64) Operation { return Operation{Mul: n} }
func set(n float64) Operation { return Operation{Set: new(n)} }

// Motive returns a Chart 10/11 motive row. Category is G (civil ground), M
// (military), F (flyer), or W (watercraft).
func Motive(category, code string) (Modifier, bool) {
	row, ok := motiveRows[category+":"+code]

	return row, ok
}

// Spec selects the three creation-chart rows and optional enhancer rows.
type Spec struct {
	Category, Type, Mission, Motive string
	Enhancers                       []Modifier
}

// Vehicle is a completed VehicleMaker calculation.
type Vehicle struct {
	Spec

	Values

	Beastpower float64
	Problems   []string
}

// Design applies a chart selection. Motive supplies the base mechanism values;
// Type and Mission then modify them, followed by enhancers in listed order.
func Design(spec Spec) Vehicle {
	motive, mok := Motive(spec.Category, spec.Motive)
	typeRow, tok := vehicleType(spec.Category, spec.Type)
	mission, sok := vehicleMission(spec.Category, spec.Mission)

	v := Vehicle{Spec: spec}
	if !mok {
		v.Problems = append(v.Problems, "unknown motive")
	}

	if !tok {
		v.Problems = append(v.Problems, "unknown vehicle type")
	}

	if !sok {
		v.Problems = append(v.Problems, "unknown mission")
	}

	if len(v.Problems) != 0 {
		return v
	}

	rows := make([]Modifier, 0, 3+len(spec.Enhancers))
	if spec.Category == "F" {
		// Flyer Mission rows multiply the Motive's base columns; the chart prints
		// Mission before Motive, but the arithmetic is evaluated on the completed
		// column. Put the base first so the multiplication has its stated operand.
		rows = append(rows, motive, typeRow, mission)
	} else {
		rows = append(rows, typeRow, motive, mission)
	}

	rows = append(rows, spec.Enhancers...)
	v.Values = Apply(rows...)
	v.Beastpower = Beastpower(v.Tons, v.Speed)

	return v
}

// Generate selects a valid ground, flyer, or watercraft design with injected
// dice, then delegates to deterministic Design.
func Generate(r *dice.Roller) Vehicle {
	categories := []string{"G", "F", "W"}
	category := categories[r.Index(len(categories))]
	choices := generationChoices[category]

	return Design(Spec{
		Category: category,
		Type:     choices.types[r.Index(len(choices.types))],
		Mission:  choices.missions[r.Index(len(choices.missions))],
		Motive:   choices.motives[r.Index(len(choices.motives))],
	})
}

type choices struct{ types, missions, motives []string }

var generationChoices = map[string]choices{
	"G": {
		[]string{"GC", "U", "T", "V", "M", "H", "R"},
		[]string{"RO", "P", "C", "MP", "OR"},
		[]string{"AC", "W", "Z", "G", "T"},
	},
	"F": {[]string{"F", "G", "B"}, []string{"A", "B", "C", "P", "S", "U"}, []string{"W", "R", "F", "LTA", "Z", "G"}},
	"W": {[]string{"S", "U", "B"}, []string{"C", "P", "E", "T"}, []string{"S", "U", "H", "G"}},
}

func vehicleType(category, code string) (Modifier, bool) {
	r, ok := typeRows[category+":"+code]

	return r, ok
}

func vehicleMission(category, code string) (Modifier, bool) {
	r, ok := missionRows[category+":"+code]

	return r, ok
}

var typeRows = map[string]Modifier{
	"G:GC": {Code: "GC", Descriptor: "GroundCar", Tons: add(2), Speed: add(-1), CostKCr: add(20)},
	"G:U":  {Code: "U", Descriptor: "Utility", Tons: add(3), Speed: add(-2), CostKCr: add(30)},
	"G:T":  {Code: "T", Descriptor: "Truck", Tons: add(4), Speed: add(-3), CostKCr: add(50)},
	"G:V":  {Code: "V", Descriptor: "Vehicle", Tons: add(5), Speed: add(-3), CostKCr: add(60)},
	"G:M":  {Code: "M", Descriptor: "Mover", Tons: add(3), CostKCr: add(50)},
	"G:H":  {Code: "H", Descriptor: "Hauler", Tons: add(5), Speed: add(-4), CostKCr: add(40)},
	"G:R":  {Code: "R", Descriptor: "Trailer", Tons: add(5), Speed: add(4), CostKCr: add(40)},
	"M:T": {
		Code:       "T",
		Descriptor: "Tank",
		Tons:       add(5),
		Speed:      add(-1),
		Armor:      add(50),
		Cage:       add(10),
		Flash:      add(10),
		Radiation:  add(10),
		Sound:      add(20),
		Insulated:  add(20),
		Sealed:     add(20),
		CostKCr:    add(500),
	},
	"M:C": {
		Code:       "C",
		Descriptor: "Carrier",
		Tons:       add(4),
		Load:       add(2),
		Armor:      add(40),
		Cage:       add(10),
		Flash:      add(10),
		Radiation:  add(10),
		Sound:      add(20),
		Insulated:  add(20),
		Sealed:     add(20),
		CostKCr:    add(300),
	},
	"M:V": {
		Code:       "V",
		Descriptor: "Vehicle",
		Tons:       add(2),
		Speed:      add(1),
		Load:       add(1),
		Armor:      add(30),
		Cage:       add(10),
		Flash:      add(10),
		Radiation:  add(10),
		Sound:      add(20),
		Insulated:  add(20),
		Sealed:     add(20),
		CostKCr:    add(100),
	},
	"M:R": {
		Code: "R", Descriptor: "Trailer", Tons: add(5), Load: add(4), Armor: add(30), Cage: add(10),
		Flash: add(10), Radiation: add(10), Sound: add(20), Insulated: add(20), Sealed: add(20), CostKCr: add(50),
	},
	"F:F": {Code: "F", Descriptor: "Flyer"},
	"F:G": {Code: "G", Descriptor: "Glider", Note: "requires Winged motive and unpowered"},
	"F:B": {Code: "B", Descriptor: "Balloon", Note: "requires LTA motive and unpowered"},
	"W:S": {
		Code:       "S",
		Descriptor: "Ship",
		TechLevel:  add(5),
		Tons:       add(1000),
		Speed:      add(4),
		Load:       add(600),
		Armor:      add(10),
		Sealed:     add(0),
		CostKCr:    add(1000),
	},
	"W:U": {
		Code:       "U",
		Descriptor: "Sub",
		TechLevel:  add(6),
		Tons:       add(100),
		Speed:      add(4),
		Load:       add(60),
		Armor:      add(20),
		Sealed:     add(20),
		CostKCr:    add(1000),
	},
	"W:B": {
		Code:       "B",
		Descriptor: "Boat",
		TechLevel:  add(5),
		Tons:       add(10),
		Speed:      add(4),
		Load:       add(6),
		Armor:      add(5),
		CostKCr:    add(100),
	},
}

var missionRows = map[string]Modifier{
	"G:RO": {Code: "RO", Descriptor: "Road Only"},
	"G:P":  {Code: "P", Descriptor: "Passenger", Load: add(5), Insulated: add(12), CostKCr: add(10)},
	"G:C":  {Code: "C", Descriptor: "Cargo", Load: add(5), Sealed: add(6), CostKCr: add(10)},
	"G:MP": {Code: "MP", Descriptor: "Multi-Purpose", Load: add(5), Sealed: add(6), CostKCr: add(10)},
	"G:OR": {
		Code:       "OR",
		Descriptor: "Off Road",
		Load:       add(20),
		Armor:      add(10),
		Cage:       add(10),
		Flash:      add(10),
		Radiation:  add(10),
		Insulated:  add(20),
		Sealed:     add(20),
		CostKCr:    add(100),
	},
	"M:W": {Code: "W", Descriptor: "Weapon", Tons: add(2), Speed: add(-1), CostKCr: add(100)},
	"M:T": {Code: "T", Descriptor: "Troop", Tons: add(1)},
	"M:S": {Code: "S", Descriptor: "Supply", Tons: add(3), Load: add(1), Armor: add(-10)},
	"M:R": {Code: "R", Descriptor: "Recon", Tons: add(-1), Speed: add(1), Armor: add(-10), CostKCr: add(100)},
	"F:A": {
		Code:       "A",
		Descriptor: "Attack",
		TechLevel:  add(2),
		Tons:       mul(2),
		Speed:      add(1),
		Load:       mul(2),
		Armor:      add(20),
		Flash:      add(20),
		Radiation:  add(1),
		Sound:      add(20),
		Insulated:  add(10),
		Sealed:     add(1),
		CostKCr:    mul(3),
	},
	"F:B": {
		Code:       "B",
		Descriptor: "Bomber",
		TechLevel:  add(1),
		Tons:       mul(3),
		Load:       mul(3),
		Armor:      add(10),
		Flash:      add(20),
		Radiation:  add(1),
		Sound:      add(20),
		Insulated:  add(10),
		Sealed:     add(1),
		CostKCr:    mul(2),
	},
	"F:C": {
		Code:       "C",
		Descriptor: "Cargo",
		Tons:       mul(4),
		Load:       mul(2),
		Armor:      add(5),
		Flash:      add(20),
		Radiation:  add(1),
		Sound:      add(20),
		Insulated:  add(10),
		Sealed:     add(1),
		CostKCr:    mul(1),
	},
	"F:P": {
		Code:       "P",
		Descriptor: "Protector",
		TechLevel:  add(1),
		Tons:       mul(2),
		Speed:      add(1),
		Load:       mul(1),
		Armor:      add(10),
		Flash:      add(20),
		Radiation:  add(1),
		Sound:      add(20),
		Insulated:  add(10),
		Sealed:     add(1),
		CostKCr:    mul(3),
	},
	"F:S": {
		Code:       "S",
		Descriptor: "Scientific",
		TechLevel:  add(-1),
		Tons:       mul(4),
		Load:       mul(2),
		Armor:      add(5),
		Flash:      add(20),
		Radiation:  add(1),
		Sound:      add(20),
		Insulated:  add(10),
		Sealed:     add(1),
		CostKCr:    mul(2),
	},
	"F:U": {
		Code:       "U",
		Descriptor: "Utility",
		Tons:       mul(3),
		Flash:      add(20),
		Radiation:  add(1),
		Sound:      add(20),
		Insulated:  add(10),
		Sealed:     add(1),
		CostKCr:    mul(10),
	},
	"W:C": {Code: "C", Descriptor: "Cargo", Speed: add(-1)},
	"W:P": {Code: "P", Descriptor: "Patrol", TechLevel: add(2), Speed: add(1), Armor: mul(2)},
	"W:E": {Code: "E", Descriptor: "Explorer", TechLevel: add(2), Sealed: mul(2)},
	"W:T": {Code: "T", Descriptor: "Transport"},
}

var motiveRows = map[string]Modifier{
	"G:AC": {Code: "AC", Descriptor: "Air Cushion", TechLevel: set(8), Tons: add(2), Speed: add(6), CostKCr: mul(2)},
	"G:W":  {Code: "W", Descriptor: "Wheeled", TechLevel: set(6), Speed: add(5), CostKCr: mul(1)},
	"G:Z":  {Code: "Z", Descriptor: "Lifter", TechLevel: set(9), Tons: add(1), Speed: add(3), CostKCr: mul(2)},
	"G:G":  {Code: "G", Descriptor: "Grav", TechLevel: set(10), Tons: add(-1), Speed: add(5), CostKCr: mul(3)},
	"G:T":  {Code: "T", Descriptor: "Tracked", TechLevel: set(7), Tons: add(2), Speed: add(4), CostKCr: mul(2)},
	"M:AC": {Code: "AC", Descriptor: "Air Cushion", TechLevel: set(8), Tons: add(2), Speed: add(6), CostKCr: mul(2)},
	"M:W":  {Code: "W", Descriptor: "Wheeled", TechLevel: set(6), Speed: add(5), CostKCr: mul(1)},
	"M:Z":  {Code: "Z", Descriptor: "Lifter", TechLevel: set(9), Tons: add(1), Speed: add(3), CostKCr: mul(2)},
	"M:G":  {Code: "G", Descriptor: "Grav", TechLevel: set(10), Tons: add(-1), Speed: add(5), CostKCr: mul(3)},
	"M:T":  {Code: "T", Descriptor: "Tracked", TechLevel: set(7), Tons: add(2), Speed: add(4), CostKCr: mul(2)},
	"F:W": {
		Code:       "W",
		Descriptor: "Winged",
		TechLevel:  set(7),
		Tons:       add(10),
		Speed:      add(8),
		Load:       add(2),
		CostKCr:    add(300),
	},
	"F:R": {
		Code:       "R",
		Descriptor: "Rotor",
		TechLevel:  set(8),
		Tons:       add(10),
		Speed:      add(7),
		Load:       add(.5),
		CostKCr:    add(400),
	},
	"F:F": {
		Code:       "F",
		Descriptor: "Flapper",
		TechLevel:  set(10),
		Tons:       add(10),
		Speed:      add(6),
		Load:       add(.5),
		CostKCr:    add(500),
	},
	"F:LTA": {
		Code:       "LTA",
		Descriptor: "Lighter-Than-Air",
		TechLevel:  set(6),
		Tons:       add(40),
		Speed:      add(5),
		Load:       add(10),
		CostKCr:    add(600),
	},
	"F:Z": {
		Code:       "Z",
		Descriptor: "Lifter",
		TechLevel:  set(9),
		Tons:       add(8),
		Speed:      add(2),
		Load:       add(1),
		CostKCr:    add(600),
	},
	"F:G": {
		Code:       "G",
		Descriptor: "Grav",
		TechLevel:  set(10),
		Tons:       add(9),
		Speed:      add(4),
		Load:       add(3),
		CostKCr:    add(700),
	},
	"W:U": {Code: "U", Descriptor: "Unpowered", TechLevel: add(-3), Speed: set(0), CostKCr: mul(.5)},
	"W:S": {Code: "S", Descriptor: "Standard"},
	"W:H": {
		Code:       "H",
		Descriptor: "Hovercraft",
		TechLevel:  set(6),
		Tons:       mul(2),
		Speed:      add(5),
		Load:       add(3),
		CostKCr:    add(200),
	},
	"W:G": {Code: "G", Descriptor: "Grav", TechLevel: set(10), Tons: mul(.2), Speed: add(4), CostKCr: mul(2)},
}

var enhancerRows = map[string]Modifier{ //nolint:gochecknoglobals // immutable chart registry
	"Vl": {
		Code: "Vl", Descriptor: "Vlight", TechLevel: add(-1), Tons: mul(1.0 / 3), Speed: add(1),
		Load: add(-2), Armor: mul(1.0 / 3), Insulated: mul(1.0 / 3), Sealed: mul(1.0 / 3),
		CostKCr: mul(1.0 / 3),
	},
	"L": {
		Code: "L", Descriptor: "Light", TechLevel: add(-1), Tons: mul(.5), Speed: add(1),
		Load: add(-1), Armor: mul(.5), Insulated: mul(.5), Sealed: mul(.5), CostKCr: mul(.5),
	},
	"M": {Code: "M", Descriptor: "Medium"},
	"H": {
		Code: "H", Descriptor: "Heavy", TechLevel: add(1), Tons: mul(2), Speed: add(-1),
		Load: add(2), Armor: mul(2), Cage: mul(2), Insulated: mul(2), Sealed: mul(2), CostKCr: mul(3),
	},
	"Vh": {
		Code: "Vh", Descriptor: "VHeavy", TechLevel: add(2), Tons: mul(3), Speed: add(-2),
		Load: add(3), Armor: mul(3), Cage: mul(2), Insulated: mul(3), Sealed: mul(3), CostKCr: mul(9),
	},
	"Fos": {
		Code: "Fos", Descriptor: "Fossil", TechLevel: add(-2), Tons: add(2), Armor: add(-10),
		Cage: add(-10), Note: "not Grav or Lifter",
	},
	"PC": {
		Code: "PC", Descriptor: "PowerCell", TechLevel: add(-1), Tons: add(1), Speed: add(-2),
		Load: add(-2), Armor: add(-5), Cage: add(-5), CostKCr: add(10), Note: "not Grav or Lifter",
	},
	"Ren": {
		Code: "Ren", Descriptor: "Renewable", TechLevel: add(-1), Tons: add(1), Speed: add(-1),
		Load: add(-1), CostKCr: add(20), Note: "not Grav or Lifter",
	},
	"Pro": {
		Code: "Pro", Descriptor: "Prototype", TechLevel: add(-2), Tons: add(1), Speed: add(-1),
		Load: add(-1), CostKCr: add(20),
	},
	"Ear": {
		Code: "Ear", Descriptor: "Early", TechLevel: add(-1), Tons: add(1), Armor: add(-10),
		Cage: add(-10), CostKCr: add(10),
	},
	"Std": {Code: "Std", Descriptor: "Standard"},
	"Imp": {
		Code: "Imp", Descriptor: "Improved", TechLevel: add(1), Tons: add(-1), Armor: add(10),
		Cage: add(10), CostKCr: add(20),
	},
	"Adv": {
		Code: "Adv", Descriptor: "Advanced", TechLevel: add(3), Tons: add(-2), Speed: add(1),
		Load: add(1), Armor: add(20), Cage: add(20), CostKCr: add(40),
	},
	"Air": {Code: "Air", Descriptor: "Air (Open)", TechLevel: add(-2), Load: set(0)},
	"Enclosed": {
		Code: "Enclosed", Descriptor: "Enclosed", TechLevel: add(-1), Armor: add(4),
		Flash: add(4), Sound: add(4), Insulated: add(12),
	},
	"Sealed": {
		Code: "Sealed", Descriptor: "Sealed", Armor: add(6), Cage: add(2), Flash: add(6),
		Radiation: set(0), Sound: add(8), Psi: set(0), Insulated: add(16), Sealed: add(20), CostKCr: add(2),
	},
	"DoubleSealed": {
		Code: "DoubleSealed", Descriptor: "Double Sealed", Tons: add(1), Armor: add(8), Cage: add(4),
		Flash: add(6), Radiation: set(0), Sound: add(12), Psi: set(0), Insulated: add(30),
		Sealed: add(20), CostKCr: add(5),
	},
	"Insulated": {
		Code: "Insulated", Descriptor: "Insulated", Armor: add(8), Cage: add(4), Flash: add(6),
		Sound: add(12), Psi: set(0), Insulated: add(30), Sealed: add(10), CostKCr: add(10),
	},
	"Protected": {
		Code: "Protected", Descriptor: "Protected", TechLevel: add(1), Tons: add(1), Armor: add(10),
		Cage: add(10), Flash: add(10), Radiation: add(10), Sound: add(12), Psi: set(0),
		Insulated: add(30), Sealed: add(20), CostKCr: add(20),
	},
	"Armored": {
		Code: "Armored", Descriptor: "Armored", TechLevel: add(2), Tons: add(1), Armor: add(20),
		Cage: add(10), Flash: add(10), Radiation: add(10), Sound: add(12), Psi: set(0),
		Insulated: add(20), Sealed: add(20), CostKCr: add(30),
	},
	"UpArmored": {
		Code: "UpArmored", Descriptor: "UpArmored", TechLevel: add(3), Tons: add(2), Armor: add(30),
		Cage: add(20), Flash: add(20), Radiation: add(20), Sound: add(20), Psi: set(0),
		Insulated: add(30), Sealed: add(20), CostKCr: add(40),
	},
	"AltArmored": {
		Code: "AltArmored", Descriptor: "AltArmored", TechLevel: add(3), Tons: add(2), Armor: add(60),
		Cage: add(20), Flash: add(30), Radiation: add(30), Sound: add(30), Psi: set(0),
		Insulated: add(30), Sealed: add(30), CostKCr: add(50),
	},
	"HighPowered": {
		Code: "HighPowered", Descriptor: "High Powered", TechLevel: add(1), Tons: add(1), Speed: add(1),
		Load: add(-1), CostKCr: add(100),
	},
	"Slave":       {Code: "Slave", Descriptor: "Slave", TechLevel: add(1), Tons: add(-1), CostKCr: add(10)},
	"Remote":      {Code: "Remote", Descriptor: "Remote", TechLevel: add(1), Tons: add(-2), CostKCr: add(20)},
	"WeaponMount": {Code: "WeaponMount", Descriptor: "Weapon Mount", Load: add(-1)},
	"Luxury":      {Code: "Luxury", Descriptor: "Luxury", CostKCr: mul(2), Note: "Q=9 or A"},
	"Fast": {
		Code: "Fast", Descriptor: "Fast", TechLevel: add(1), Tons: add(1), Speed: add(1),
		Load: add(-2), CostKCr: add(30),
	},
	"PassengerModule": {
		Code: "PassengerModule", Descriptor: "Passenger Module", Load: add(-3), CostKCr: add(100),
		Note: "20 passengers",
	},
	"CargoModule": {
		Code: "CargoModule", Descriptor: "Cargo Module", Tons: add(1), Speed: add(-1), Load: add(1),
		CostKCr: add(20), Note: "one ton",
	},
	"Redundancy": {Code: "Redundancy", Descriptor: "Redundancy", TechLevel: add(1), Tons: add(1), CostKCr: add(60)},
	"OffRoad":    {Code: "OffRoad", Descriptor: "OffRoad", CostKCr: add(30), Note: "ground; Wheeled"},
	"Mole": {
		Code: "Mole", Descriptor: "Mole", TechLevel: add(1), Tons: mul(3), Speed: set(1),
		CostKCr: add(400), Note: "Ground Explorer, not ACV",
	},
	"Hydrofoils": {
		Code: "Hydrofoils", Descriptor: "Hydrofoils", TechLevel: add(1), Tons: add(1), Speed: add(1),
		CostKCr: add(30), Note: "watercraft",
	},
	"Stubs": {Code: "Stubs", Descriptor: "Stubs", CostKCr: add(20), Note: "flyer; Grav or Lifter"},
	"VTOL": {
		Code: "VTOL", Descriptor: "VTOL Mod", Speed: add(-1), Load: add(-2), CostKCr: add(100),
		Note: "flyer; Medium or less",
	},
	"STOL": {Code: "STOL", Descriptor: "STOL Mod", Load: add(-1), CostKCr: add(50), Note: "flyer; Heavy or less"},
	"LiftingBody": {
		Code: "LiftingBody", Descriptor: "Lifting Body Hull", Tons: add(4), Speed: add(1), Load: mul(2),
		CostKCr: add(200), Note: "flyer",
	},
	"Wings1": {
		Code: "Wings1", Descriptor: "Add-On Wings-1", Tons: mul(2), Speed: add(1), Load: mul(1),
		CostKCr: add(100), Note: "flyer; B +1",
	},
	"Wings2": {
		Code: "Wings2", Descriptor: "Add-On Wings-2", Tons: mul(3), Speed: add(2), Load: mul(2),
		CostKCr: add(200), Note: "flyer; B +1",
	},
	"Wings3": {
		Code: "Wings3", Descriptor: "Add-On Wings-3", Tons: mul(4), Speed: add(3), Load: mul(3),
		CostKCr: add(300), Note: "flyer; B +1",
	},
	"Floats": {
		Code: "Floats", Descriptor: "Float Landing Gear", Tons: add(-1), Speed: add(-1), CostKCr: add(100),
		Note: "flyer",
	},
	"ParasiteNipple": {
		Code: "ParasiteNipple", Descriptor: "Parasite Nipple", TechLevel: add(1), Load: add(-1),
		CostKCr: add(100), Note: "flyer",
	},
}
