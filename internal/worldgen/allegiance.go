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
// human). The domain is open-ended, so Valid checks the shape of a code — two or
// four letters — not membership in any list. Only Imperial has a named constant, because it is the
// only one the code path uses (the default); the rest are supplied by the referee.
type Allegiance string

// Imperial is the default polity, used when the referee names none.
const (
	Imperial          Allegiance = "Im"
	DefaultAllegiance            = Imperial
)

// Valid reports whether the code is well-formed: two or four ASCII letters. The
// book permits codes beyond the ones it lists ("Other abbreviations are possible",
// Chart F), so this checks shape, not membership — an over-strict validator would
// reject legitimate referee-defined polities.
//
// Length alone is not shape. The Second Survey record is positional and
// whitespace-delimited, so a two- or four-character code carrying a space ("Im d")
// adds a field token and shifts every later column, and one carrying a control byte
// puts it straight into a piped record stream. Requiring letters rules out both
// without narrowing the book's open domain. Case is deliberately not enforced:
// the listed codes are leading-capital (Im, ImDd), but that is convention, and
// rejecting "im" would fail a code no parser could misread.
func (a Allegiance) Valid() bool {
	if len(a) != 2 && len(a) != 4 {
		return false
	}

	for i := range len(a) {
		if c := a[i]; (c < 'A' || c > 'Z') && (c < 'a' || c > 'z') {
			return false
		}
	}

	return true
}

// InvalidAllegiance is the sentinel rendered in place of a malformed code. Like
// ehex's "?" it is visibly wrong and safe in a record stream: it is the right
// width for the two-letter form and cannot be mistaken for a polity.
const InvalidAllegiance = "??"

// String renders the code for the record, substituting InvalidAllegiance for a
// malformed one. Allegiance is an exported string type a caller may convert into
// directly, bypassing ParseAllegiance, and — as with uwp.Profile's String — a value
// reaching the display path must not corrupt the record. Unlike Valid, which is the
// strict predicate, String never fails; it is the never-panicking display half of
// the pair.
func (a Allegiance) String() string {
	if !a.Valid() {
		return InvalidAllegiance
	}

	return string(a)
}

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
