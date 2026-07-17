package dice

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Kind identifies how a parsed Roll is evaluated.
type Kind int

const (
	// Standard is nD with an optional constant: N dice summed, plus Mod.
	Standard Kind = iota
	// FluxRoll is the D-D roll, ranging -5..+5. N and Mod are ignored.
	FluxRoll
	// Half is D/2 rounded up, ranging 1..3. N and Mod are ignored.
	Half
)

// A Roll is a parsed die instruction such as "2D", "2D-2", "Flux", or "D/2".
// Charts throughout T5 give rolls in this form; Parse turns the text into a
// Roll and Eval executes it against a Roller.
type Roll struct {
	Kind Kind
	N    int // number of dice (Standard only)
	Mod  int // constant added to the sum (Standard only), may be negative
}

// Parse reads a T5 die instruction. It accepts, case-insensitively and
// ignoring spaces: "Flux"; "D/2"; and the standard form "nD", "nD+k", or
// "nD-k" where the leading count defaults to 1 (so "D" means "1D").
func Parse(s string) (Roll, error) {
	t := strings.ToUpper(strings.ReplaceAll(s, " ", ""))
	switch t {
	case "":
		return Roll{}, errors.New("dice: empty roll")
	case "FLUX":
		return Roll{Kind: FluxRoll}, nil
	case "D/2":
		return Roll{Kind: Half}, nil
	}

	d := strings.IndexByte(t, 'D')
	if d < 0 {
		return Roll{}, fmt.Errorf("dice: %q is not a die roll", s)
	}

	n := 1

	if d > 0 {
		v, err := strconv.Atoi(t[:d])
		if err != nil {
			return Roll{}, fmt.Errorf("dice: bad dice count in %q", s)
		}

		n = v
	}

	if n < 1 {
		return Roll{}, fmt.Errorf("dice: dice count must be positive in %q", s)
	}

	mod := 0

	if rest := t[d+1:]; rest != "" {
		if rest[0] != '+' && rest[0] != '-' {
			return Roll{}, fmt.Errorf("dice: modifier must be signed (+k or -k) in %q", s)
		}

		v, err := strconv.Atoi(rest)
		if err != nil {
			return Roll{}, fmt.Errorf("dice: bad modifier in %q", s)
		}

		mod = v
	}

	return Roll{Kind: Standard, N: n, Mod: mod}, nil
}

// Eval executes the roll against r.
func (roll Roll) Eval(r *Roller) int {
	switch roll.Kind {
	case FluxRoll:
		return r.Flux()
	case Half:
		return r.HalfDie()
	default:
		return r.Dice(roll.N) + roll.Mod
	}
}

// String renders the roll back into T5 notation.
func (roll Roll) String() string {
	switch roll.Kind {
	case FluxRoll:
		return "Flux"
	case Half:
		return "D/2"
	default:
		if roll.Mod > 0 {
			return fmt.Sprintf("%dD+%d", roll.N, roll.Mod)
		}

		if roll.Mod < 0 {
			return fmt.Sprintf("%dD%d", roll.N, roll.Mod)
		}

		return fmt.Sprintf("%dD", roll.N)
	}
}
