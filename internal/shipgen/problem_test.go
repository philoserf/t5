package shipgen

import "testing"

// TestProblemKindNamesCoverEveryKind guards the hand-kept problemKindNames map: a
// new ProblemKind added to the iota block without a name entry would silently
// render as "ProblemKind(?)" in a test-failure message. Iterating the whole enum
// range and asserting each names itself catches that drift.
func TestProblemKindNamesCoverEveryKind(t *testing.T) {
	for k := ConfigInvalid; k <= MissileGuidanceUnavailable; k++ {
		if got := k.String(); got == "ProblemKind(?)" {
			t.Errorf("ProblemKind %d has no name in problemKindNames", int(k))
		}
	}

	// And no stray names for values outside the enum.
	if len(problemKindNames) != int(MissileGuidanceUnavailable)+1 {
		t.Errorf("problemKindNames has %d entries, want %d (one per kind)",
			len(problemKindNames), int(MissileGuidanceUnavailable)+1)
	}
}
