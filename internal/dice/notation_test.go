package dice

import "testing"

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want Roll
	}{
		{"1D", Roll{Kind: Standard, N: 1}},
		{"2D", Roll{Kind: Standard, N: 2}},
		{"D", Roll{Kind: Standard, N: 1}},
		{"2D-2", Roll{Kind: Standard, N: 2, Mod: -2}},
		{"2D+2", Roll{Kind: Standard, N: 2, Mod: 2}},
		{"8D", Roll{Kind: Standard, N: 8}},
		{"flux", Roll{Kind: FluxRoll}},
		{"FLUX", Roll{Kind: FluxRoll}},
		{" 2d - 2 ", Roll{Kind: Standard, N: 2, Mod: -2}},
		{"D/2", Roll{Kind: Half}},
	}
	for _, c := range cases {
		got, err := Parse(c.in)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", c.in, err)

			continue
		}

		if got != c.want {
			t.Errorf("Parse(%q) = %+v, want %+v", c.in, got, c.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	for _, in := range []string{"", "2", "xD", "2D+", "0D", "-1D", "1D6", "2D2"} {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) = nil error, want error", in)
		}
	}
}

func TestEval(t *testing.T) {
	r := scripted(3, 4)

	roll, _ := Parse("2D+2")
	if got := roll.Eval(r); got != 9 { // 3+4+2
		t.Fatalf("Eval(2D+2) = %d, want 9", got)
	}

	if got := (Roll{Kind: FluxRoll}).Eval(scripted(6, 1)); got != 5 {
		t.Fatalf("Eval(Flux) = %d, want 5", got)
	}
}

func TestString(t *testing.T) {
	cases := map[string]Roll{
		"2D":   {Kind: Standard, N: 2},
		"2D+2": {Kind: Standard, N: 2, Mod: 2},
		"2D-2": {Kind: Standard, N: 2, Mod: -2},
		"Flux": {Kind: FluxRoll},
		"D/2":  {Kind: Half},
	}
	for want, roll := range cases {
		if got := roll.String(); got != want {
			t.Errorf("String() = %q, want %q", got, want)
		}
	}
}
