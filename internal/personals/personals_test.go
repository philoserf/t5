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

func TestCamaraderieCap(t *testing.T) {
	var c Camaraderie
	for range 10 {
		c.Gain()
	}
	if c.Mod() != MaxCamaraderie {
		t.Errorf("Camaraderie capped at %d, got %d", MaxCamaraderie, c.Mod())
	}
}
