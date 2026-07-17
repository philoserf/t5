package shipgen

import "fmt"

// Mount allocation (Book 2 pp.83, 156). A hull carries weapons at its mount
// points, and how many it has is a function of its size alone: "A hull allows one
// Hardpoint per 100 tons. One Mount may be installed at each Hardpoint."
//
// The wrinkle is FirmPoints: "A hull may allocate, instead of HardPoints, three
// FirmPoints per 100 tons. A Firmpoint will accept any mount which is less than
// one ton (a full-ton mount does not fit on a FirmPoint)." So each 100-ton block
// is spent one of two ways — as a single HardPoint carrying any mount, or as
// three FirmPoints each carrying a sub-ton one. A ship that wants many small
// weapons trades big-mount capacity for count.

// firmPointsPerBlock is how many FirmPoints a 100-ton block yields when it is
// allocated to them instead of to its single HardPoint (Book 2 p.156).
const firmPointsPerBlock = 3

// mountPoints reports whether a hull of the given tonnage can carry the weapons
// and defenses, and the shortfall message if it cannot. Mounts of one ton or more
// need a HardPoint each; sub-ton mounts ride FirmPoints, three to a block.
//
// Bolt-In defenses are free of all this: a Bolt-In "can be placed anywhere within
// a ship and does not require a HardPoint or FirmPoint, nor access to the exterior
// of the hull" (Book 2 p.176). It still weighs three tons, but it competes for
// nothing — which is exactly why it is the defenses' natural mount, and why a ship
// can carry screens without giving up a single gun.
//
// The cheapest allocation gives every full mount its own HardPoint block and packs
// the sub-ton mounts three to a block, so the hull needs
// hardMounts + ceil(subTonMounts/3) blocks. Fewer blocks cannot be made to work by
// any other split, and more are never required.
func mountPoints(h Hull, weapons []Weapon, defenses []Defense) string {
	hard, firm := 0, 0
	// A component that failed to design occupies no mount. Counting it would invent
	// a mount shortfall on top of the real problem that stopped it being built — two
	// complaints for one mistake, one of them false.
	count := func(t Tonnage, problems []string) {
		if len(problems) > 0 || t == 0 {
			return
		}

		if t.SubTon() {
			firm++
		} else {
			hard++
		}
	}
	for _, w := range weapons {
		count(w.Tons, w.Problems)
	}

	for _, d := range defenses {
		if d.Spec.Mount == BoltIn {
			continue // needs no mount point at all
		}

		count(d.Tons, d.Problems)
	}

	if hard == 0 && firm == 0 {
		return ""
	}
	// Hull.Hardpoints is the hull's mount capacity, one per 100 tons. Recomputing
	// the division here would give the field a second home and let the two drift.
	need := hard + (firm+firmPointsPerBlock-1)/firmPointsPerBlock
	if need <= h.Hardpoints {
		return ""
	}

	return fmt.Sprintf(
		"%s need %d mount blocks but a %dt hull has %d (one Hardpoint, or %d Firmpoints, per 100t)",
		mountPhrase(hard, firm),
		need,
		h.Tons,
		h.Hardpoints,
		firmPointsPerBlock,
	)
}

// mountPhrase names what is being mounted, for the shortfall message.
func mountPhrase(hard, firm int) string {
	switch {
	case firm == 0:
		return fmt.Sprintf("%d mounts", hard)
	case hard == 0:
		return fmt.Sprintf("%d sub-ton mounts", firm)
	default:
		return fmt.Sprintf("%d mounts and %d sub-ton mounts", hard, firm)
	}
}

// armamentTonnage is the whole tons the weapons and defenses occupy in the ship's
// budget. A mount on a HardPoint "is at least 1 ton (round up)" (Book 2 p.83), so
// a full mount is charged whole tons; sub-ton mounts on FirmPoints keep their
// fractions, and the total rounds up once at the end.
func armamentTonnage(weapons []Weapon, defenses []Defense) int {
	var total Tonnage

	charge := func(t Tonnage) {
		if t.SubTon() {
			total += t

			return
		}

		total += t.RoundUp()
	}
	for _, w := range weapons {
		charge(w.Tons)
	}

	for _, d := range defenses {
		charge(d.Tons)
	}

	return total.Ceil()
}
