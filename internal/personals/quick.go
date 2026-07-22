package personals

import (
	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/task"
)

// QuickNPC holds the referee-secret 2D base values for Carouse, Query,
// Persuade, and Command, in that order (Book 1 p.183).
type QuickNPC struct {
	values [4]int
}

// GenerateQuickNPC rolls the four values in Purpose order, consuming exactly
// eight dice.
func GenerateQuickNPC(r *dice.Roller) QuickNPC {
	var npc QuickNPC
	for purpose := Carouse; purpose <= Command; purpose++ {
		npc.values[purpose] = r.Dice(2)
	}

	return npc
}

// Value returns the NPC's hidden base value for a Purpose.
func (npc QuickNPC) Value(p Purpose) (int, bool) {
	if !p.valid() {
		return 0, false
	}

	return npc.values[p], true
}

// Check rolls the Purpose's normal dice against its Quick NPC base value and
// any applicable Law/situational Mods.
func (npc QuickNPC) Check(r *dice.Roller, p Purpose, mods ...int) dice.CheckResult {
	value, ok := npc.Value(p)
	if !ok {
		panic("personals: quick NPC purpose out of range")
	}

	return task.ResolveDice(r, p.Dice(), value, mods...)
}
