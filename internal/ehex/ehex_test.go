package ehex

import "testing"

func TestDigitRoundTrip(t *testing.T) {
	for v := 0; v <= Max; v++ {
		c := Digit(v)
		got, err := ParseDigit(c)
		if err != nil {
			t.Fatalf("ParseDigit(%q) error: %v", c, err)
		}
		if got != v {
			t.Fatalf("round trip %d -> %q -> %d", v, c, got)
		}
	}
}

func TestDigitKnownValues(t *testing.T) {
	cases := map[int]byte{0: '0', 9: '9', 10: 'A', 12: 'C', 17: 'H', 18: 'J', Max: 'Z'}
	for v, want := range cases {
		if got := Digit(v); got != want {
			t.Errorf("Digit(%d) = %q, want %q", v, got, want)
		}
	}
}

func TestAlphabetOmitsIandO(t *testing.T) {
	for _, c := range []byte{'I', 'O'} {
		if _, err := ParseDigit(c); err == nil {
			t.Errorf("ParseDigit(%q) should fail; I and O are not eHex digits", c)
		}
	}
}

func TestParseDigitCaseInsensitive(t *testing.T) {
	up, _ := ParseDigit('A')
	low, err := ParseDigit('a')
	if err != nil || up != 10 || low != 10 {
		t.Fatalf("case-insensitive parse failed: up=%d low=%d err=%v", up, low, err)
	}
}

func TestFormat(t *testing.T) {
	if got := Format(12); got != "C" {
		t.Errorf("Format(12) = %q, want C", got)
	}
	for _, v := range []int{-1, Max + 1, 40} {
		if got := Format(v); got != "?" {
			t.Errorf("Format(%d) = %q, want ? (must not panic)", v, got)
		}
	}
}

func TestDigitPanicsOutOfRange(t *testing.T) {
	for _, v := range []int{-1, Max + 1} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Digit(%d) did not panic", v)
				}
			}()
			Digit(v)
		}()
	}
}
