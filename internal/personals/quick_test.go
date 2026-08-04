package personals

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestGenerateQuickNPC(t *testing.T) {
	npc := GenerateQuickNPC(dice.NewScripted(3, 4, 2, 5, 6, 1, 4, 3))

	want := map[Purpose]int{Carouse: 7, Query: 7, Persuade: 7, Command: 7}
	for purpose, value := range want {
		if got, ok := npc.Value(purpose); !ok || got != value {
			t.Errorf("Value(%s) = %d, %v; want %d, true", purpose, got, ok, value)
		}
	}

	if _, ok := npc.Value(Purpose(99)); ok {
		t.Error("out-of-range purpose returned a Quick NPC value")
	}
}

func TestQuickNPCCheck(t *testing.T) {
	npc := GenerateQuickNPC(dice.NewScripted(3, 4, 3, 4, 3, 4, 3, 4))

	result := npc.Check(dice.NewScripted(3, 4), Query, LawMod(Similarity, Query))
	if !result.Success || result.Roll != 7 || result.Target != 8 {
		t.Errorf("Quick NPC Query = %+v, want roll 7 against 8", result)
	}
}

func TestQuickNPCSeeded(t *testing.T) {
	a := GenerateQuickNPC(dice.NewWithSeed(188))
	b := GenerateQuickNPC(dice.NewWithSeed(188))

	for purpose := Carouse; purpose <= Command; purpose++ {
		av, _ := a.Value(purpose)
		bv, _ := b.Value(purpose)

		if av != bv {
			t.Errorf("seeded %s values differ: %d != %d", purpose, av, bv)
		}
	}
}
