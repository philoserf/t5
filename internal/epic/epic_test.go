package epic

import (
	"testing"

	"github.com/philoserf/t5/internal/dice"
)

func TestScaffoldMatchesTypicalEPIC(t *testing.T) {
	adventure := Scaffold()
	wantPurposes := []string{
		"Who What Why Where When How",
		"Meeting The Key Non-Players",
		"Pursuing The Key Goal",
		"Approaching Resolution",
	}

	if adventure.Theme != "" {
		t.Errorf("Theme = %q, want empty bare scaffold", adventure.Theme)
	}

	if len(adventure.Acts) != 4 {
		t.Fatalf("len(Acts) = %d, want 4", len(adventure.Acts))
	}

	for i, act := range adventure.Acts {
		if act.Number != i+1 || act.Purpose != wantPurposes[i] {
			t.Errorf("Act %d = {%d %q}, want {%d %q}", i, act.Number, act.Purpose, i+1, wantPurposes[i])
		}

		if len(act.Scenes) != 5 {
			t.Fatalf("len(Act %d Scenes) = %d, want 5", i+1, len(act.Scenes))
		}

		for j, scene := range act.Scenes {
			if scene != (Scene{Number: j + 1}) {
				t.Errorf("Act %d Scene %d = %+v, want unfilled numbered slot", i+1, j+1, scene)
			}
		}
	}

	if adventure.Climax.Purpose != "The Final Resolution" || len(adventure.Climax.Scenes) != 0 {
		t.Errorf("Climax = %+v, want unfilled final resolution", adventure.Climax)
	}
}

func TestScaffoldReturnsIndependentScenes(t *testing.T) {
	first := Scaffold()
	first.Acts[0].Scenes[0].Purpose = "changed"

	if got := Scaffold().Acts[0].Scenes[0].Purpose; got != "" {
		t.Errorf("new Scaffold scene purpose = %q, want empty", got)
	}
}

func TestThemeTable(t *testing.T) {
	want := [6][6]string{
		{"Justice", "Loyalty", "Awe", "Danger", "Laziness", "Betrayal"},
		{"Happiness", "Cheerful", "Human Frailty", "Paranoia", "Heroism", "Disappointing"},
		{"Kindness", "Trustworthy", "Brave", "Pursuit", "Escape", "Unreliable"},
		{"Honesty", "Admiration", "Bizarre", "Revenge", "Deception", "Stupidity"},
		{"Truthfulness", "Friendly", "Thrifty", "Humiliation", "Conformity", "Confusion"},
		{"Cleanliness", "Novelty", "Profitable", "Improbable", "Extremes", "Chaos"},
	}

	for row := 1; row <= 6; row++ {
		for column := 1; column <= 6; column++ {
			got, ok := Theme(column, row)
			if !ok || got != want[row-1][column-1] {
				t.Errorf("Theme(%d, %d) = %q, %v; want %q, true", column, row, got, ok, want[row-1][column-1])
			}
		}
	}
}

func TestThemeRejectsNonDieCoordinates(t *testing.T) {
	for _, tc := range []struct{ column, row int }{{0, 1}, {7, 1}, {1, 0}, {1, 7}} {
		if got, ok := Theme(tc.column, tc.row); ok || got != "" {
			t.Errorf("Theme(%d, %d) = %q, %v; want empty, false", tc.column, tc.row, got, ok)
		}
	}
}

func TestGenerateRollsThemeAndScaffold(t *testing.T) {
	adventure := Generate(dice.NewScripted(6, 5))
	if adventure.Theme != "Confusion" {
		t.Errorf("Theme = %q, want Confusion", adventure.Theme)
	}

	if len(adventure.Acts) != 4 || len(adventure.Acts[0].Scenes) != 5 {
		t.Errorf("Generate did not include canonical scaffold: %+v", adventure)
	}
}
