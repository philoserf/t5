package skill

import (
	"strconv"
	"testing"
)

func TestRaiseAndCap(t *testing.T) {
	var s Set
	s.Raise("Pilot", 1)
	s.Raise("Pilot", 2)

	if got := s.Level("Pilot"); got != 3 {
		t.Errorf("Pilot level = %d, want 3", got)
	}

	s.Raise("Pilot", 20)

	if got := s.Level("Pilot"); got != Max {
		t.Errorf("Pilot level = %d, want cap %d", got, Max)
	}

	if s.Level("Gunner") != 0 {
		t.Errorf("absent skill level = %d, want 0", s.Level("Gunner"))
	}
	// A negative change cannot drive a level below 0.
	s.Raise("Comms", -3)

	if got := s.Level("Comms"); got != 0 {
		t.Errorf("Comms after negative raise = %d, want 0", got)
	}

	s.RaiseKnowledge("Pilot", "Small Craft", -3)

	if got := s.KnowledgeLevel("Pilot", "Small Craft"); got != 0 {
		t.Errorf("knowledge after negative raise = %d, want 0", got)
	}
}

// TestRaiseNonPositiveDoesNotRegister pins the difference between the two
// level-0 entries. A non-positive Raise of a skill the character does not hold
// is a no-op — it must not leave a phantom "Comms-0" in List/String — while
// RaiseKnowledge's parent registration at level 0 is the deliberate cascade
// marker (Knowledge-Knowledge-Skill) and must survive.
func TestRaiseNonPositiveDoesNotRegister(t *testing.T) {
	var s Set
	s.Raise("Comms", -3)
	s.Raise("Steward", 0)

	if got := s.String(); got != "" {
		t.Errorf("String() after non-positive raises of absent skills = %q, want empty", got)
	}

	if got := len(s.List()); got != 0 {
		t.Errorf("List() = %v, want no entries", s.List())
	}
	// A skill the character does hold keeps its entry when decremented to 0.
	s.Raise("Comms", 1)
	s.Raise("Comms", -5)

	if got := s.String(); got != "Comms-0" {
		t.Errorf("String() after decrementing a held skill = %q, want %q", got, "Comms-0")
	}
	// The cascade parent is still registered at 0 by RaiseKnowledge.
	var c Set
	c.RaiseKnowledge("Pilot", "Small Craft", 1)

	if got := c.String(); got != "Pilot-0 Pilot/Small Craft-1" {
		t.Errorf("cascade parent registration = %q, want %q", got, "Pilot-0 Pilot/Small Craft-1")
	}
}

func TestRaiseKnowledgeAndCap(t *testing.T) {
	var s Set
	s.RaiseKnowledge("Engineer", "J-Drive", 1)

	if got := s.KnowledgeLevel("Engineer", "J-Drive"); got != 1 {
		t.Errorf("J-Drive = %d, want 1", got)
	}
	// Raising a knowledge registers the parent skill at level 0.
	if _, present := s.skills["Engineer"]; !present {
		t.Error("parent skill Engineer not registered")
	}

	s.RaiseKnowledge("Engineer", "J-Drive", 20)

	if got := s.KnowledgeLevel("Engineer", "J-Drive"); got != KnowledgeMax {
		t.Errorf("J-Drive = %d, want cap %d", got, KnowledgeMax)
	}
}

func TestTaskLevelStacks(t *testing.T) {
	var s Set
	s.Raise("Engineer", 1)
	s.RaiseKnowledge("Engineer", "J-Drive", 1)

	if got := s.TaskLevel("Engineer", "J-Drive"); got != 2 {
		t.Errorf("TaskLevel = %d, want 2 (1 skill + 1 knowledge)", got)
	}
}

func TestGrantCascadeKKS(t *testing.T) {
	// Knowledge-Knowledge-Skill: awards 1-2 raise the knowledge, 3+ the skill.
	var s Set
	s.GrantCascade("Pilot", "Small Craft") // 1: knowledge 1, skill 0

	if s.KnowledgeLevel("Pilot", "Small Craft") != 1 || s.Level("Pilot") != 0 {
		t.Fatalf(
			"after 1st: K=%d S=%d, want 1/0",
			s.KnowledgeLevel("Pilot", "Small Craft"),
			s.Level("Pilot"),
		)
	}

	s.GrantCascade("Pilot", "Small Craft") // 2: knowledge 2

	if s.KnowledgeLevel("Pilot", "Small Craft") != 2 || s.Level("Pilot") != 0 {
		t.Fatalf(
			"after 2nd: K=%d S=%d, want 2/0",
			s.KnowledgeLevel("Pilot", "Small Craft"),
			s.Level("Pilot"),
		)
	}

	s.GrantCascade("Pilot", "Small Craft") // 3: skill 1

	if s.KnowledgeLevel("Pilot", "Small Craft") != 2 || s.Level("Pilot") != 1 {
		t.Fatalf(
			"after 3rd: K=%d S=%d, want 2/1",
			s.KnowledgeLevel("Pilot", "Small Craft"),
			s.Level("Pilot"),
		)
	}

	s.GrantCascade("Pilot", "Small Craft") // 4: skill 2

	if s.Level("Pilot") != 2 {
		t.Fatalf("after 4th: S=%d, want 2", s.Level("Pilot"))
	}
}

// Book 1 p. 75 counts "up to FIVE Skills at level 6+ (or Knowledges at level-6)
// (but not languages)" toward Master Points, so a level-6 knowledge counts and a
// Language knowledge does not. KnowledgeMax is 6, so a qualifying knowledge is
// always exactly at the threshold.
func TestTopLevelsCountsKnowledges(t *testing.T) {
	var s Set

	s.Raise("Craftsman", 9)
	s.RaiseKnowledge("Engineer", "J-Drive", KnowledgeMax)
	s.RaiseKnowledge("Gunner", "Turret", KnowledgeMax)
	s.RaiseKnowledge("Language", "Vilani", KnowledgeMax) // excluded: not languages
	s.RaiseKnowledge("Pilot", "Small Craft", 5)          // below the threshold

	// Engineer/J-Drive 6 + Gunner/Turret 6; Craftsman and Language are excluded
	// and Pilot/Small Craft is under level 6.
	if got := s.TopLevels(5, 6, "Craftsman", "Language"); got != 12 {
		t.Errorf("TopLevels = %d, want 12", got)
	}
}

// The five slots are shared between skills and knowledges, highest first.
func TestTopLevelsSharesSlotsWithKnowledges(t *testing.T) {
	var s Set

	for _, lvl := range []int{10, 9, 8, 7} {
		s.Raise("S"+strconv.Itoa(lvl), lvl)
	}

	s.RaiseKnowledge("Engineer", "J-Drive", KnowledgeMax) // 6, takes the 5th slot
	s.RaiseKnowledge("Gunner", "Turret", KnowledgeMax)    // 6, crowded out

	if got := s.TopLevels(5, 6); got != 10+9+8+7+6 {
		t.Errorf("TopLevels = %d, want %d", got, 10+9+8+7+6)
	}
}

func TestString(t *testing.T) {
	var s Set
	s.Raise("Navigation", 2)
	s.Raise("Vacc Suit", 1)
	s.RaiseKnowledge("Pilot", "Small Craft", 1)
	// Sorted: Navigation-2, Pilot-0 + Pilot/Small Craft-1, Vacc Suit-1.
	want := "Navigation-2 Pilot-0 Pilot/Small Craft-1 Vacc Suit-1"
	if got := s.String(); got != want {
		t.Fatalf("String() =\n%q\nwant\n%q", got, want)
	}
}

// TestTopLevelsCountsParentAndKnowledgeOnce: a cascade parent and a Knowledge
// beneath it are one competency, not two. Counting both filled two of the five
// Master Point slots from a single skill, inflating a Craftsman's Masterpoint
// total and with it his Masterpiece chances.
func TestTopLevelsCountsParentAndKnowledgeOnce(t *testing.T) {
	var s Set

	s.Raise("Engineer", 6)
	s.RaiseKnowledge("Engineer", "J-Drive", 6)
	s.Raise("Animals", 6)

	// Two competencies at 6, not three: Engineer contributes once.
	if got, want := s.TopLevels(5, 6), 12; got != want {
		t.Errorf("TopLevels = %d, want %d (Engineer counted once, plus Animals)", got, want)
	}

	// And the knowledge alone still fills a slot when the parent is below the bar.
	var t2 Set

	t2.RaiseKnowledge("Gunner", "Turret", 6)

	if got, want := t2.TopLevels(5, 6), 6; got != want {
		t.Errorf("knowledge-only TopLevels = %d, want %d", got, want)
	}
}
