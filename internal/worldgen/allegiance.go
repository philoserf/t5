package worldgen

// Allegiance is the polity a world belongs to (Book 3 Chart F, p.28, "A"). It is
// imposed by the referee, not generated — the checklist on p.23 says so outright —
// so this is a validated code with an Imperial default, not a table roll.
//
// The book lists Im (Imperial), Cs (Client State), Na (Non-Aligned), Va (Vargr),
// As (Aslan), Zh (Zhodani), So (Solomani), Kk (K'kree), and Hv (Hiver), and adds
// that "Other abbreviations are possible" — so Valid checks the shape of a code
// (two letters), not membership in that list. Only Imperial has a named constant,
// because it is the only one the code path uses (the default); the rest are named
// where a referee supplies them. If per-world allegiance is ever generated or a
// renderer needs the full polity names, the enumeration and a name table belong
// here then, with a consumer.
type Allegiance string

// Imperial is the default polity, used when the referee names none.
const (
	Imperial          Allegiance = "Im"
	DefaultAllegiance            = Imperial
)

// Valid reports whether the code is well-formed: exactly two characters. The book
// permits codes beyond the nine it lists, so this checks shape, not membership.
func (a Allegiance) Valid() bool { return len(a) == 2 }

// ParseAllegiance reads a referee-supplied allegiance code, defaulting an empty one
// to Imperial. It reports whether the code was well-formed; a malformed one returns
// the default too, so a caller that ignores the bool (as the survey line does)
// always renders a valid two-letter code rather than a broken record.
func ParseAllegiance(code string) (Allegiance, bool) {
	if a := Allegiance(code); a.Valid() {
		return a, true
	}
	return DefaultAllegiance, code == ""
}
