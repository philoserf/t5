package dice

import "testing"

func TestResolveRollLow(t *testing.T) {
	// 2D = 5 against Target 7: success with Effect +2.
	got := scripted(2, 3).Resolve(Check{Dice: Average, Target: 7})
	if !got.Success || got.Effect != 2 || got.Roll != 5 || got.Total != 5 {
		t.Fatalf("Resolve = %+v, want success roll 5 effect 2", got)
	}
}

func TestResolveDefaultsToAverage(t *testing.T) {
	// Dice unset -> 2D. Two dice consumed.
	got := scripted(4, 4).Resolve(Check{Target: 9})
	if got.Roll != 8 || !got.Success {
		t.Fatalf("Resolve default dice = %+v, want roll 8 success", got)
	}
}

func TestResolveModRaisesTarget(t *testing.T) {
	// Roll 8 vs Target 7 fails; Mod +2 raises Target to 9 and it succeeds.
	base := scripted(5, 3).Resolve(Check{Target: 7})
	if base.Success {
		t.Fatalf("expected base failure, got %+v", base)
	}

	modded := scripted(5, 3).Resolve(Check{Target: 7, Mod: 2})
	if !modded.Success || modded.Target != 9 {
		t.Fatalf("Mod result = %+v, want success target 9", modded)
	}
}

func TestResolveDMRaisesRoll(t *testing.T) {
	// Roll 6 vs Target 7 succeeds; DM +2 raises the roll to 8 and it fails.
	dm := scripted(3, 3).Resolve(Check{Target: 7, DM: 2})
	if dm.Success || dm.Total != 8 || dm.Effect != -1 {
		t.Fatalf("DM result = %+v, want failure total 8 effect -1", dm)
	}
}
