package worldgen

// Allegiance is the polity a world belongs to (Book 3 Chart F, p.28, "A"). It is
// imposed by the referee, not generated — the checklist on p.23 says so outright —
// so this is a validated code with an Imperial default, not a table roll. A world's
// allegiance is a fact about the setting, decided when a sector is placed, and the
// generators carry it through rather than inventing it.

// Allegiance is a two-letter allegiance code.
type Allegiance string

// The allegiance codes Chart F names. The book adds "Other abbreviations are
// possible", so an unlisted two-letter code is accepted rather than rejected — a
// setting may have polities the core rules do not enumerate.
const (
	Imperial    Allegiance = "Im"
	ClientState Allegiance = "Cs"
	NonAligned  Allegiance = "Na"
	Vargr       Allegiance = "Va"
	Aslan       Allegiance = "As"
	Zhodani     Allegiance = "Zh"
	Solomani    Allegiance = "So"
	Kkree       Allegiance = "Kk"
	Hiver       Allegiance = "Hv"
)

// DefaultAllegiance is the allegiance a world takes when the referee names none:
// Imperial, the setting's default polity.
const DefaultAllegiance = Imperial

// allegianceNames records the polity each Chart F code stands for, and doubles as
// the set of codes the book enumerates.
var allegianceNames = map[Allegiance]string{
	Imperial:    "Imperial",
	ClientState: "Client State",
	NonAligned:  "Non-Aligned",
	Vargr:       "Vargr",
	Aslan:       "Aslan",
	Zhodani:     "Zhodani",
	Solomani:    "Solomani",
	Kkree:       "K'kree",
	Hiver:       "Hiver",
}

// Name returns the polity's full name, or the bare code for a setting-specific one
// the core rules do not list.
func (a Allegiance) Name() string {
	if name, ok := allegianceNames[a]; ok {
		return name
	}
	return string(a)
}

// Valid reports whether the code is well-formed: exactly two characters. The book
// permits codes beyond the nine it lists ("Other abbreviations are possible"), so
// this checks shape, not membership.
func (a Allegiance) Valid() bool {
	return len(a) == 2
}

// ParseAllegiance reads a referee-supplied allegiance code, defaulting an empty one
// to Imperial. It reports whether the code is well-formed; a malformed one still
// returns the default so a caller always has something to render.
func ParseAllegiance(code string) (Allegiance, bool) {
	if code == "" {
		return DefaultAllegiance, true
	}
	a := Allegiance(code)
	return a, a.Valid()
}
