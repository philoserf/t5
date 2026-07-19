package chargen

import "testing"

// cascadeParents is the book's list of the skills that contain Knowledges. The
// book prints the list twice and the two disagree; this is the inclusive one,
// from the Skills chapter's "Knowledge is the Foundation of Skill" (Book 1
// p.133): "Some skills include within them several Knowledges (Animals, Driver,
// Engineer, Fighter, Flyer, Gunner, Heavy Weapons, Language, Musician, Pilot,
// Seafarer)."
//
// Two notes on its authority:
//
//   - The other list, "The Knowledges-Only Skills" in the careers chapter (Book 1
//     p.63), omits Language and Musician. The inclusive one is used because it is
//     a superset, and because Language is plainly a cascade — the careers-chapter
//     box is about what Education and Training can impart, and Language is the
//     one cascade whose acquisition is "handled differently".
//   - The parent space is closed but the *knowledge* space is open: the Skills
//     chapter's "More Knowledges" says "Many other Knowledges are possible: one
//     for every career; one for every world; one for every branch of science".
//     So this test constrains cascade cells' parents only, never the knowledges
//     beneath them.
var cascadeParents = map[string]bool{
	"Animals": true, "Driver": true, "Engineer": true, "Fighter": true,
	"Flyer": true, "Gunner": true, "Heavy Weapons": true, "Language": true,
	"Musician": true, "Pilot": true, "Seafarer": true,
}

// TestCareerGridCascadeParents walks every career's skill grid and asserts that
// each cascade cell names a real cascade parent skill.
//
// A cascade cell routes through skill.Set.GrantCascade, which files the award as
// a Knowledge under cell.Skill and registers that parent at level 0. Nothing on
// the write path validates the parent, so a transcription slip — a misspelled
// parent, or a pluralized one such as "Pilots" — mints a phantom cascade: the
// character still renders plausibly ("Pilots-0 Pilots/Small Craft-1"), but the
// knowledge is filed under a parent no rule ever reads, so it never stacks with
// the real skill it was meant to sit under and TaskLevel silently undercounts.
// This test is the check that the deleted skill.IsCascade write-path guard never
// actually performed.
func TestCareerGridCascadeParents(t *testing.T) {
	for _, career := range allCareers {
		for col := range career.Skills {
			for row, cell := range career.Skills[col] {
				parent := cell.cascadeParent()
				if parent == "" {
					continue
				}

				if !cascadeParents[parent] {
					t.Errorf(
						"%s grid col %d row %d: cascade parent %q is not a cascade skill (Book 1 p.133)",
						career.Name, col, row+1, parent,
					)
				}
			}
		}
	}
}

// TestCareerGridsHaveCascadeCells guards the guard: if the grids ever stop
// carrying cascade cells, TestCareerGridCascadeParents would pass vacuously.
func TestCareerGridsHaveCascadeCells(t *testing.T) {
	found := 0

	for _, career := range allCareers {
		for col := range career.Skills {
			for _, cell := range career.Skills[col] {
				if cell.cascadeParent() != "" {
					found++
				}
			}
		}
	}

	if found == 0 {
		t.Fatal("no cascade cells in any career grid: TestCareerGridCascadeParents is vacuous")
	}
}
