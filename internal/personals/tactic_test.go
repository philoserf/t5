package personals

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func effectCode(effect TacticEffect, ok bool) string {
	if !ok {
		return "-"
	}

	if effect.Mod != 0 {
		code := fmt.Sprintf("%+d", effect.Mod)
		if effect.InflictsPain {
			code += "*"
		}

		return code
	}

	if effect.Multiplier != 1 {
		return strconv.Itoa(effect.Multiplier)
	}

	return "."
}

func TestTacticMatrixExhaustive(t *testing.T) {
	// Every one of the 20 printed strategy rows, across all 16 tactic columns.
	rows := []struct {
		purpose  Purpose
		strategy Strategy
		want     string
	}{
		{Carouse, Casual, "+3 +2 . - . . . . - - . 2 2 2 - -8"},
		{Carouse, Enjoyment, ". . . . . . . . . - . 2 2 2 - -6"},
		{Carouse, Discussion, ". . 2 2 2 2 2 . . . . . 2 2 - -6"},
		{Carouse, ActiveListening, ". . . . . . . . . . . . . . . -6"},
		{Carouse, AppealTo, ". . 2 . 2 2 2 . 2 2 2 2 . . . +4*"},
		{Query, Enjoyment, "+2 +1 . . . . . . . - . 2 2 2 - -6"},
		{Query, Discussion, ". . 2 2 2 2 2 . . . . . 2 2 - -6"},
		{Query, ActiveListening, ". . . . . . . . . . . . . . . -6"},
		{Query, AppealTo, ". . 2 . 2 2 2 . 2 2 2 2 . . . +4*"},
		{Query, ForceOfWill, ". . 2 2 2 2 2 . . - . . . . 2 +6*"},
		{Persuade, Discussion, "+1 . 2 2 2 2 2 . . . . . 2 2 - -6"},
		{Persuade, ActiveListening, ". . . . . . . . . . . . . . . -6"},
		{Persuade, AppealTo, ". . 2 . 2 2 2 . 2 2 2 2 . . . +4*"},
		{Persuade, ForceOfWill, ". . 2 2 2 2 2 . . - . . . . 2 +6*"},
		{Persuade, Charming, ". 3 . 2 2 2 2 2 . 2 . 2 2 . . -4*"},
		{Command, ActiveListening, ". . . . . . . . . . . . . . . -6"},
		{Command, AppealTo, ". . 2 . 2 2 2 . 2 2 2 2 . . . +4*"},
		{Command, ForceOfWill, ". . 2 2 2 2 2 . . - . . . . 2 +6*"},
		{Command, Charming, ". 3 . 2 2 2 2 2 . 2 . 2 2 . . -4*"},
		{Command, Angry, ". 3 2 2 2 2 2 2 . . . . . . 2 +6*"},
	}

	for _, row := range rows {
		want := strings.Fields(row.want)
		if len(want) != 16 {
			t.Fatalf("bad test row %s/%d: %d cells", row.purpose, row.strategy, len(want))
		}

		for tactic, code := range want {
			effect, ok := TacticEffectFor(row.purpose, row.strategy, Tactic(tactic))
			if got := effectCode(effect, ok); got != code {
				t.Errorf("%s/%d tactic %d = %s, want %s", row.purpose, row.strategy, tactic, got, code)
			}
		}
	}
}

func TestTacticMatrixRejectsUnavailableStrategy(t *testing.T) {
	if _, ok := TacticEffectFor(Carouse, Angry, Logic); ok {
		t.Error("Carouse/Angry unexpectedly available")
	}

	if _, ok := TacticEffectFor(Query, AppealTo, Tactic(99)); ok {
		t.Error("out-of-range tactic unexpectedly available")
	}
}

func TestResolveSelectionAndFightHook(t *testing.T) {
	// Persuade Charming (5) with Common Enemies (x3), no law: target 15.
	result, ok := ResolveSelection(dice.NewScripted(5, 5, 5), Persuade, Charming, CommonEnemies, 0)
	if !ok || !result.Success || result.Target != 15 || result.Fight {
		t.Errorf("Charming/Common Enemies = %+v, %v", result, ok)
	}

	// Query Force of Will (5) with Pain (+6): target 11. Failure requests a
	// combat handoff and the starred cell records inflicted pain.
	violent, ok := ResolveSelection(dice.NewScripted(6, 6), Query, ForceOfWill, Pain, 0)
	if !ok || violent.Success || !violent.Fight || !violent.InflictsPain || violent.Target != 11 {
		t.Errorf("failed violent Personal = %+v, %v", violent, ok)
	}
}

func TestResolveSelectionRejectsBeforeRolling(t *testing.T) {
	r := dice.NewScripted(2)
	if _, ok := ResolveSelection(r, Carouse, Casual, Authority, 0); ok {
		t.Error("incompatible Casual/Authority resolved")
	}

	// The rejected selection consumed nothing.
	if got := r.Die(); got != 2 {
		t.Errorf("die after rejection = %d, want 2", got)
	}
}
