package worldgen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/philoserf/t5/internal/ehex"
	"github.com/philoserf/t5/internal/tradecode"
	"github.com/philoserf/t5/internal/uwp"
)

// Serialization for the world portion of a Second Survey record (REDESIGN.md §2,
// #327). World.SecondSurvey renders six space-separated fields —
//
//	UWP  TCs  {Ix}(Ex)[Cx]  Nobility  Bases  Zone
//
// e.g. "A788899-C Ph Pa Ri {+4}(D7E+4)[9C6D] BcCeF NS -". ParseWorld reverses it,
// building a World through NewWorld that re-renders the identical line. The result
// is a *record* World, not a fully generated one: the line does not carry Belt,
// NativeStatus, or PopulationDigit, and SecondSurvey does not render them, so they
// are absent and unneeded. Strict throughout — a malformed record errors rather
// than yielding a plausible-but-wrong world, and the reader is the exact inverse of
// SecondSurvey over a valid record.

// ParseWorld decodes a World record line (the world portion of a Second Survey
// record). It errors on a line that is not the six-field StSAHPGL form.
func ParseWorld(s string) (World, error) {
	fields := strings.Fields(s)

	// UWP, then the trade codes, then the Extensions token (the only field that
	// starts with '{'), then exactly nobility, bases, and zone.
	ext := -1

	for i, f := range fields {
		if strings.HasPrefix(f, "{") {
			ext = i

			break
		}
	}

	if ext < 1 || ext+3 != len(fields)-1 {
		return World{}, fmt.Errorf(
			"worldgen: %q is not a world record (want UWP TCs {Ix}(Ex)[Cx] Nobility Bases Zone)", s,
		)
	}

	profile, err := uwp.Parse(fields[0])
	if err != nil {
		return World{}, fmt.Errorf("worldgen: %q: %w", s, err)
	}

	tcs, err := parseTradeCodes(fields[1:ext])
	if err != nil {
		return World{}, fmt.Errorf("worldgen: %q: %w", s, err)
	}

	ix, ex, cx, err := parseExtensions(fields[ext])
	if err != nil {
		return World{}, fmt.Errorf("worldgen: %q: %w", s, err)
	}

	naval, scout, depot, wayStation, err := parseBases(fields[ext+2])
	if err != nil {
		return World{}, fmt.Errorf("worldgen: %q: %w", s, err)
	}

	zone, err := parseZone(fields[ext+3])
	if err != nil {
		return World{}, fmt.Errorf("worldgen: %q: %w", s, err)
	}

	base := World{
		Profile:    profile,
		TradeCodes: tcs,
		NavalBase:  naval, ScoutBase: scout, NavalDepot: depot, WayStation: wayStation,
		Zone: zone,
	}

	return NewWorld(base, ix, ex, cx, fields[ext+1]), nil
}

// parseTradeCodes reverses the trade-code field: a lone "-" is no codes, otherwise
// each token is a Chart D code, rejected if it is not one.
func parseTradeCodes(tokens []string) ([]tradecode.Code, error) {
	if len(tokens) == 1 && tokens[0] == "-" {
		return nil, nil
	}

	codes := make([]tradecode.Code, len(tokens))

	for i, t := range tokens {
		c := tradecode.Code(t)
		if !tradecode.Valid(c) {
			return nil, fmt.Errorf("%q is not a trade classification", t)
		}

		codes[i] = c
	}

	return codes, nil
}

// parseExtensions reverses "{Ix}(Ex)[Cx]" into its three parts (Book 3 Chart E).
func parseExtensions(s string) (int, Economic, Cultural, error) {
	ixStr, rest, ok := bracketed(s, '{', '}')
	if !ok {
		return 0, Economic{}, Cultural{}, fmt.Errorf("%q: missing {Ix}", s)
	}

	exStr, rest, ok := bracketed(rest, '(', ')')
	if !ok {
		return 0, Economic{}, Cultural{}, fmt.Errorf("%q: missing (Ex)", s)
	}

	cxStr, rest, ok := bracketed(rest, '[', ']')
	if !ok || rest != "" {
		return 0, Economic{}, Cultural{}, fmt.Errorf("%q: missing or trailing [Cx]", s)
	}

	ix, err := strconv.Atoi(ixStr)
	if err != nil {
		return 0, Economic{}, Cultural{}, fmt.Errorf("Importance %q: %w", ixStr, err)
	}

	ex, err := parseEconomic(exStr)
	if err != nil {
		return 0, Economic{}, Cultural{}, err
	}

	cx, err := parseCultural(cxStr)
	if err != nil {
		return 0, Economic{}, Cultural{}, err
	}

	return ix, ex, cx, nil
}

// bracketed pulls a leading open..closer-delimited span off s, returning its inner
// text and the remainder. It reports false unless s starts with open and holds a
// matching closer.
func bracketed(s string, open, closer byte) (string, string, bool) {
	if len(s) == 0 || s[0] != open {
		return "", s, false
	}

	end := strings.IndexByte(s, closer)
	if end < 0 {
		return "", s, false
	}

	return s[1:end], s[end+1:], true
}

// parseEconomic reverses "(Ex)" content "RLIe": three eHex digits (Resources,
// Labor, Infrastructure) then a signed Efficiency.
func parseEconomic(s string) (Economic, error) {
	if len(s) < 4 {
		return Economic{}, fmt.Errorf("Economic %q too short", s)
	}

	d, err := ehex.ParseDigits(s[:3])
	if err != nil {
		return Economic{}, fmt.Errorf("Economic %q: %w", s, err)
	}

	eff, err := strconv.Atoi(s[3:])
	if err != nil {
		return Economic{}, fmt.Errorf("Economic efficiency %q: %w", s[3:], err)
	}

	return Economic{Resources: d[0], Labor: d[1], Infrastructure: d[2], Efficiency: eff}, nil
}

// parseCultural reverses "[Cx]" content "HASy": four eHex digits.
func parseCultural(s string) (Cultural, error) {
	if len(s) != 4 {
		return Cultural{}, fmt.Errorf("Cultural %q is not four digits", s)
	}

	d, err := ehex.ParseDigits(s)
	if err != nil {
		return Cultural{}, fmt.Errorf("Cultural %q: %w", s, err)
	}

	return Cultural{Heterogeneity: d[0], Acceptance: d[1], Strangeness: d[2], Symbols: d[3]}, nil
}

// parseBases reverses the bases field: "-" is none, otherwise a run of N/S/D/W in
// that order (baseList). An unknown letter, or an out-of-order/duplicate run,
// errors — the reader accepts only what bases() emits.
func parseBases(s string) (bool, bool, bool, bool, error) {
	if s == "-" {
		return false, false, false, false, nil
	}

	var naval, scout, depot, wayStation bool

	for i := range len(s) {
		switch s[i] {
		case 'N':
			naval = true
		case 'S':
			scout = true
		case 'D':
			depot = true
		case 'W':
			wayStation = true
		default:
			return false, false, false, false, fmt.Errorf("%q is not a base code (N/S/D/W)", s)
		}
	}

	// Reject any spelling the writer would not emit (wrong order, repeats) by
	// comparing against bases() itself, so the reader is provably its inverse.
	canonical := World{NavalBase: naval, ScoutBase: scout, NavalDepot: depot, WayStation: wayStation}.bases()
	if canonical != s {
		return false, false, false, false, fmt.Errorf("%q is not canonical bases (want %q)", s, canonical)
	}

	return naval, scout, depot, wayStation, nil
}

// parseZone reverses the zone field: "-" is Green (zone() renders A and R only, so
// anything else dashes), "A" Amber, "R" Red.
func parseZone(s string) (byte, error) {
	switch s {
	case "-":
		return 'G', nil
	case "A":
		return 'A', nil
	case "R":
		return 'R', nil
	default:
		return 0, fmt.Errorf("%q is not a travel zone (A/R/-)", s)
	}
}
