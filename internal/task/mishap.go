package task

// MishapKind identifies which failure consequence is being checked.
type MishapKind int

const (
	// OrdinaryMishap checks Reliability and may injure the task user.
	OrdinaryMishap MishapKind = iota
	// DangerousMishap checks Safety and may cause a greater injury.
	DangerousMishap
	// DestructiveMishap checks Safety and may damage the associated device.
	DestructiveMishap
)

// A MishapCheck records the p. 131 failure-consequence roll. Ordinary mishaps
// use Reliability; Dangerous and Destructive mishaps use Safety. The printed
// test is Flux + rating, with a negative result triggering the consequence.
// Injury/damage location, diagnosis, and repair are separate tables and remain
// the caller's responsibility.
type MishapCheck struct {
	Kind      MishapKind
	Flux      int
	Rating    int
	Total     int
	Triggered bool
}

// EvaluateMishap evaluates one failure consequence. It performs no roll so a
// referee may keep the Flux private; pass dice.Roller.Flux() as flux when a
// random result is wanted.
func EvaluateMishap(kind MishapKind, flux, rating int) MishapCheck {
	total := flux + rating

	return MishapCheck{Kind: kind, Flux: flux, Rating: rating, Total: total, Triggered: total < 0}
}
