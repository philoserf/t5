package worldgen

// Allegiance is the polity a world belongs to (Book 3 Chart F, p.28, "A"). It is
// imposed by the referee, not generated — the checklist on p.23 says so outright —
// so this is a validated code with an Imperial default, not a table roll.
//
// The core rules Chart F lists two-letter codes — Im (Imperial), Cs (Client
// State), Na (Non-Aligned), Va (Vargr), As (Aslan), Zh (Zhodani), So (Solomani),
// Kk (K'kree), Hv (Hiver) — and add that "Other abbreviations are possible". The
// Second Survey / T5SS convention also uses four-letter codes that name the
// specific polity within a race (ImDd, the Domain of Deneb; NaHu, non-aligned
// human). So Valid checks the shape of a code — two or four letters — not
// membership in any list. Only Imperial has a named constant, because it is the
// only one the code path uses (the default); the rest are supplied by the referee.
type Allegiance string

// Imperial is the default polity, used when the referee names none.
const (
	Imperial          Allegiance = "Im"
	DefaultAllegiance            = Imperial
)

// Valid reports whether the code is well-formed: a two-letter Chart F code or a
// four-letter Second Survey code. The book permits codes beyond the ones it lists,
// so this checks shape, not membership.
func (a Allegiance) Valid() bool { return len(a) == 2 || len(a) == 4 }

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
