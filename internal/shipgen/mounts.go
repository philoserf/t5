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

// mountPoints reports whether a hull of the given tonnage can carry the weapons,
// and the shortfall message if it cannot. Weapons of one ton or more need a
// HardPoint each; sub-ton weapons ride FirmPoints, three to a block.
//
// The cheapest allocation gives every full mount its own HardPoint block and
// packs the sub-ton mounts three to a block, so the hull needs
// hardMounts + ceil(subTonMounts/3) blocks. Fewer blocks than that cannot be made
// to work by any other split, and more are never required.
func mountPoints(hullTons int, weapons []Weapon) string {
	hard, firm := 0, 0
	for _, w := range weapons {
		if w.Tons.SubTon() {
			firm++
		} else {
			hard++
		}
	}
	if hard == 0 && firm == 0 {
		return ""
	}
	blocks := hullTons / 100
	need := hard + (firm+firmPointsPerBlock-1)/firmPointsPerBlock
	if need <= blocks {
		return ""
	}
	return fmt.Sprintf("%s need %d mount blocks but a %dt hull has %d (one Hardpoint, or %d Firmpoints, per 100t)",
		mountPhrase(hard, firm), need, hullTons, blocks, firmPointsPerBlock)
}

// mountPhrase names what is being mounted, for the shortfall message.
func mountPhrase(hard, firm int) string {
	switch {
	case firm == 0:
		return fmt.Sprintf("%d weapons", hard)
	case hard == 0:
		return fmt.Sprintf("%d sub-ton weapons", firm)
	default:
		return fmt.Sprintf("%d weapons and %d sub-ton weapons", hard, firm)
	}
}

// weaponTonnage is the whole tons the weapons occupy in the ship's budget. A
// mount on a HardPoint "is at least 1 ton (round up)" (Book 2 p.83), so a full
// mount is charged whole tons; the sub-ton mounts on FirmPoints keep their
// fractions, and the total rounds up once at the end.
func weaponTonnage(weapons []Weapon) int {
	var total Tonnage
	for _, w := range weapons {
		if w.Tons.SubTon() {
			total += w.Tons
			continue
		}
		total += Tons(w.Tons.Ceil())
	}
	return total.Ceil()
}
