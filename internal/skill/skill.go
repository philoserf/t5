// Package skill models a Traveller5 character's skills and knowledges (Book 1,
// Skills pp. 132-171). A plain skill has a level; a cascade skill (Pilot,
// Gunner, Engineer, …) contains named Knowledges, and a knowledge acquired
// through a career follows the Knowledge-Knowledge-Skill progression. The
// package holds no dice — it is a pure inventory.
package skill

import (
	"maps"
	"slices"
	"strconv"
	"strings"
)

// Max is the level cap for a skill; KnowledgeMax caps a knowledge (and science).
const (
	Max          = 15
	KnowledgeMax = 6
)

// cascades are the skills that contain Knowledges.
var cascades = map[string]bool{
	"Animals": true, "Driver": true, "Engineer": true, "Fighter": true,
	"Flyer": true, "Gunner": true, "Heavy Weapons": true, "Language": true,
	"Musician": true, "Pilot": true, "Seafarer": true,
}

// IsCascade reports whether a skill contains Knowledges.
func IsCascade(name string) bool { return cascades[name] }

// A Set is a character's skills and knowledges. The zero value is ready to use.
type Set struct { //nolint:recvcheck // deliberate value-reader / pointer-mutator split
	skills     map[string]int            // plain skill or cascade parent -> level
	knowledges map[string]map[string]int // parent -> knowledge -> level
}

// Level returns a skill's level (0 if absent, including a cascade parent that is
// present at level 0).
func (s Set) Level(skill string) int { return s.skills[skill] }

// KnowledgeLevel returns a knowledge's level under its parent skill (0 if absent).
func (s Set) KnowledgeLevel(parent, knowledge string) int {
	return s.knowledges[parent][knowledge]
}

// TaskLevel returns the effective level for a task performed with a knowledge:
// the parent skill and the knowledge stack (Book 1: Engineer-1 + J-Drive-1
// resolves J-Drive tasks at 2).
func (s Set) TaskLevel(parent, knowledge string) int {
	return s.Level(parent) + s.KnowledgeLevel(parent, knowledge)
}

// TopLevels returns the summed levels of the highest n skills *and knowledges* at
// or above minLevel, skipping any skill named in exclude (a knowledge is skipped
// when its parent skill is excluded — which is how "but not languages" is
// expressed, "Language" being both the parent skill's name and its cascade key).
// Used for Craftsman Master Points (Book 1 p. 75: the Craftsman skill and "up to
// FIVE Skills at level 6+ (or Knowledges at level-6) (but not languages)"). The
// n slots are shared between skills and knowledges, highest first. Since
// KnowledgeMax is 6, a qualifying knowledge sits exactly at the p. 75 threshold.
func (s Set) TopLevels(n, minLevel int, exclude ...string) int {
	skip := make(map[string]bool, len(exclude))
	for _, name := range exclude {
		skip[name] = true
	}

	levels := make([]int, 0, len(s.skills))
	for name, lvl := range s.skills {
		if lvl >= minLevel && !skip[name] {
			levels = append(levels, lvl)
		}
	}

	for parent, ks := range s.knowledges {
		if skip[parent] {
			continue
		}

		for _, lvl := range ks {
			if lvl >= minLevel {
				levels = append(levels, lvl)
			}
		}
	}

	slices.Sort(levels)
	slices.Reverse(levels)

	sum := 0
	for i := 0; i < n && i < len(levels); i++ {
		sum += levels[i]
	}

	return sum
}

// Raise changes a plain skill by n levels, clamped to 0..Max (n is normally
// positive; a negative n cannot drive the level below 0).
func (s *Set) Raise(skill string, n int) {
	if s.skills == nil {
		s.skills = make(map[string]int)
	}

	s.skills[skill] = clamp(s.skills[skill]+n, 0, Max)
}

// RaiseKnowledge changes a knowledge under its parent by n, clamped to
// 0..KnowledgeMax. The parent skill is registered (at least level 0) so the
// character is known to hold the cascade skill.
func (s *Set) RaiseKnowledge(parent, knowledge string, n int) {
	if s.skills == nil {
		s.skills = make(map[string]int)
	}

	if _, ok := s.skills[parent]; !ok {
		s.skills[parent] = 0
	}

	if s.knowledges == nil {
		s.knowledges = make(map[string]map[string]int)
	}

	if s.knowledges[parent] == nil {
		s.knowledges[parent] = make(map[string]int)
	}

	s.knowledges[parent][knowledge] = clamp(s.knowledges[parent][knowledge]+n, 0, KnowledgeMax)
}

func clamp(v, lo, hi int) int { return min(max(v, lo), hi) }

// GrantCascade applies one career award of a cascade parent's knowledge, per the
// Knowledge-Knowledge-Skill progression: the first two awards raise the
// knowledge (to 1, then 2), and further awards raise the parent skill (1, 2, …).
// A player who would rather keep advancing the knowledge past 2 calls
// RaiseKnowledge directly.
func (s *Set) GrantCascade(parent, knowledge string) {
	if s.KnowledgeLevel(parent, knowledge) < 2 {
		s.RaiseKnowledge(parent, knowledge, 1)
	} else {
		s.Raise(parent, 1)
	}
}

// List returns the skills and knowledges as sorted "Name-Level" entries (a
// cascade knowledge as "Parent/Knowledge-Level"). Callers that render skills —
// which can contain spaces, e.g. "Vacc Suit" — should join this slice rather
// than split String() on spaces.
func (s Set) List() []string {
	var entries []string
	for _, name := range slices.Sorted(maps.Keys(s.skills)) {
		entries = append(entries, name+"-"+strconv.Itoa(s.skills[name]))
		for _, k := range slices.Sorted(maps.Keys(s.knowledges[name])) {
			entries = append(entries, name+"/"+k+"-"+strconv.Itoa(s.knowledges[name][k]))
		}
	}

	return entries
}

// String renders the inventory as space-separated "Skill-level" entries in a
// stable order, with knowledges shown as "Parent/Knowledge-level".
func (s Set) String() string {
	return strings.Join(s.List(), " ")
}
