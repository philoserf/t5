package shipgen

import "testing"

// TestProblemKindNamesCoverEveryKind guards the hand-kept problemKindNames map: a
// new ProblemKind added to the iota block without a name entry would silently
// render as "ProblemKind(?)" in a test-failure message. The bound is the
// numProblemKinds sentinel, not the last named constant, so a kind appended at the
// END of the enum is covered too — the sentinel tracks the count automatically.
func TestProblemKindNamesCoverEveryKind(t *testing.T) {
	for k := range numProblemKinds {
		if k.String() == "ProblemKind(?)" {
			t.Errorf("ProblemKind %d has no name in problemKindNames", int(k))
		}
	}

	// And no stray names beyond the enum (the sentinel is not itself named).
	if len(problemKindNames) != int(numProblemKinds) {
		t.Errorf("problemKindNames has %d entries, want %d (one per kind)",
			len(problemKindNames), int(numProblemKinds))
	}
}
