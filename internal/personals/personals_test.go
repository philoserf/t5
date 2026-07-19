package personals

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestPurposeDice(t *testing.T) {
	cases := map[Purpose]int{Carouse: 1, Query: 2, Persuade: 3, Command: 4}
	for p, want := range cases {
		if p.Dice() != want {
			t.Errorf("%s.Dice() = %d, want %d", p, p.Dice(), want)
		}
	}
}

// Issue #308: a Purpose outside the closed four-constant enum has no dice count.
// Dice must not invent one — Purpose(-1) gave 0, which dice.Resolve then treats
// as its 2D default, and Purpose(99) gave 100D. Same shape, and the same fix, as
// task.Difficulty.Dice (#242).
func TestPurposeDiceOffLadderPanics(t *testing.T) {
	for _, p := range []Purpose{-1, Command + 1, 99} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Purpose(%d).Dice() = %d, want panic", int(p), p.Dice())
				}
			}()

			_ = p.Dice()
		}()
	}
}

// The other half of the strict/display split: String stays total, and reports a
// bad value rather than naming it Carouse.
func TestPurposeStringOffLadder(t *testing.T) {
	if got := Purpose(99).String(); got != "?" {
		t.Errorf("Purpose(99).String() = %q, want %q", got, "?")
	}

	if got := Carouse.String(); got != "Carouse" {
		t.Errorf("Carouse.String() = %q, want %q", got, "Carouse")
	}
}

func TestStrategyValue(t *testing.T) {
	// The same strategy is valued differently per purpose (Book 1 p.183).
	if v, ok := StrategyValue(Persuade, Charming); !ok || v != 5 {
		t.Errorf("Persuade/Charming = %d,%v, want 5,true", v, ok)
	}

	if v, ok := StrategyValue(Command, Charming); !ok || v != 4 {
		t.Errorf("Command/Charming = %d,%v, want 4,true", v, ok)
	}

	if v, ok := StrategyValue(Carouse, AppealTo); !ok || v != 5 {
		t.Errorf("Carouse/AppealTo = %d,%v, want 5,true", v, ok)
	}
	// Charming is not available for Carouse or Query.
	if _, ok := StrategyValue(Carouse, Charming); ok {
		t.Errorf("Carouse/Charming should be unavailable")
	}
}

func TestLawMod(t *testing.T) {
	if LawMod(Superiority, Command) != 3 || LawMod(Superiority, Carouse) != 0 {
		t.Errorf("Superiority mods wrong")
	}

	if LawMod(Comfort, Carouse) != 2 || LawMod(Violence, Persuade) != 2 {
		t.Errorf("Comfort/Violence mods wrong")
	}
	// Inferiority's Persuade +2 is conditional (Begging/Flattery/Politeness) and
	// so is not in the unconditional table; only Query +1 is.
	if LawMod(Inferiority, Persuade) != 0 || LawMod(Inferiority, Query) != 1 {
		t.Errorf(
			"Inferiority mods = %d/%d, want 0/1",
			LawMod(Inferiority, Persuade),
			LawMod(Inferiority, Query),
		)
	}
	// An out-of-range Purpose returns 0 rather than panicking.
	if LawMod(Similarity, Purpose(99)) != 0 {
		t.Errorf("out-of-range Purpose should return 0")
	}
}

func TestResolveWorkedExample(t *testing.T) {
	// Book 1 p.184: Persuade, Charming (5), Flattery (x2), Similarity (+1),
	// Camaraderie-1 (+1) -> Target = 5*2 + 1 + 1 = 12 on 3D.
	sv, _ := StrategyValue(Persuade, Charming)
	law := LawMod(Similarity, Persuade)

	var cam Camaraderie
	cam.Gain() // one prior successful Carouse -> +1

	res := Resolve(dice.NewScripted(4, 3, 3), Persuade, sv, 2, law, cam.Mod()) // 3D = 10
	if res.Target != 12 {
		t.Errorf("target = %d, want 12", res.Target)
	}

	if res.Roll != 10 || !res.Success {
		t.Errorf("roll 10 vs 12 should succeed: %+v", res)
	}
	// A roll of 13 (over 12) fails.
	if res := Resolve(dice.NewScripted(5, 5, 3), Persuade, sv, 2, law, cam.Mod()); res.Success {
		t.Errorf("roll 13 vs 12 should fail: %+v", res)
	}
}

// The full Mods for Personals sidebar (Book 1 p.185).
func TestSituationalMods(t *testing.T) {
	cases := map[string]struct{ got, want int }{
		"By Communicator. Voice":            {VoiceOnly, -4},
		"By Communicator. Voice + Visual":   {VoiceAndVisual, -2},
		"Brazen (Query or Persuade)":        {Brazen, 3},
		"Urgent (only once)":                {Urgent, 2},
		"Subsequent Use of Strategies, per": {Repeat, -1},
	}
	for name, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", name, c.got, c.want)
		}
	}
}

// "Threat of Violence = +Fighter Skill" (Book 1 p.185).
// "Bluff (once) Flux" (Book 1 p.185) — a Flux roll, so it can help or hurt.
func TestBluff(t *testing.T) {
	if got := Bluff(dice.NewScripted(6, 1)); got != 5 {
		t.Errorf("Bluff = %d, want 5", got)
	}

	if got := Bluff(dice.NewScripted(1, 6)); got != -5 {
		t.Errorf("Bluff = %d, want -5", got)
	}
}

// Resolve sums whatever mods it is handed: the two-mod cap (three with
// Deliberate) is the caller's to enforce, not Resolve's.
func TestResolveDoesNotCapMods(t *testing.T) {
	res := Resolve(
		dice.NewScripted(1, 1, 1),
		Persuade,
		5,
		1,
		1,
		Brazen,
		Urgent,
		2, /* Threat of Violence = +Fighter Skill */
	)
	if res.Target != 5+1+Brazen+Urgent+2 {
		t.Errorf("target = %d, want %d", res.Target, 5+1+Brazen+Urgent+2)
	}
}

func TestCamaraderieCap(t *testing.T) {
	var c Camaraderie
	for range 10 {
		c.Gain()
	}

	if c.Mod() != MaxCamaraderie {
		t.Errorf("Camaraderie capped at %d, got %d", MaxCamaraderie, c.Mod())
	}
}
