package shipgen

// Problem is a single reported reason a designed ship, weapon, defense, or
// missile could not be built the way its spec asked. Design and its component
// designers are total (never an error, never a panic): infeasibility is reported
// here rather than refused.
//
// Kind is the machine-readable failure category — the stable contract a caller
// (and a cross-package test) keys on. Detail is the human-readable prose that the
// ship card prints. Problem.String() returns Detail so the rendered record is
// byte-identical to when Problems was a []string.
type Problem struct {
	Kind   ProblemKind
	Detail string
}

// String returns the display prose, so a Problem renders exactly as its old
// []string form did (fmt "%s"/"%v"/"%q" all honour this Stringer).
func (p Problem) String() string { return p.Detail }

// reported reports whether p carries a real failure. designDrive returns a single
// Problem by value — the zero Problem (empty Detail) when the drive is buildable —
// rather than a slice, so callers ask this instead of the slice-form aboard/len.
func (p Problem) reported() bool { return p.Detail != "" }

// ProblemKind names a distinct design-failure category. Giving the failure a type
// means the failure prose is no longer a cross-package test contract: tests assert
// the Kind, and the Detail wording is free to change (#332).
type ProblemKind int

// The problem kinds, one per distinct design failure. They group by where the
// failure arises — spec fields, drives, whole-ship assembly, weapon/defense
// installation, then missile rounds — with the weapon/defense group's first four
// shared, since a weapon standing in as a defense fails the same ways.
const (
	ConfigInvalid     ProblemKind = iota // configuration ordinal names no config
	HullSizeInvalid                      // hull letter outside A..Z
	DriveStageInvalid                    // drive stage ordinal names no stage
	DriveSizeInvalid                     // drive size ordinal names no drive

	DrivePotentialZero // drive too small for the hull, or rounded to Potential 0
	DriveAboveTL       // drive Potential capped by the yard's tech level

	DriveExceedsHullG    // maneuver drive rated above the hull config's G cap
	NoPowerPlant         // drives present but no power plant to feed them
	PowerBelowDrives     // power plant weaker than the drives it feeds
	ComponentAboveShipTL // a weapon/defense TL exceeds the ship's TL
	OverBudget           // components exceed the hull's tonnage
	MountsExceedHull     // more mount blocks needed than the hull has

	UnknownWeapon           // weapon/mount/range ordinal names nothing
	WeaponMountIncompatible // a weapon placed in a mount that cannot fire out
	MountTooSmall           // mount below the weapon's minimum
	RangeIncompatible       // device built for a range on the wrong scale ladder

	UnknownDefenseMountOrRange // defense mount/range ordinal names nothing
	DefenseRangeIncrease       // a defense range increased above the standard
	WeaponNotADefense          // a weapon the book does not allow as point defence
	UnknownDefense             // defense ordinal names no screen

	MissileRoundSizeMismatch    // launcher does not throw the round's size
	WarheadUnavailable          // the size carries no such warhead
	MissileRoundTypeMismatch    // launcher does not throw the round's type
	MissileGuidanceIncompatible // the round cannot carry the chosen guidance
	MissileGuidanceUnavailable  // guidance too advanced for this size/TL

	numProblemKinds // sentinel: the count of kinds, kept last so it tracks the enum
)

// String names the kind for readable test-failure messages. It is not the display
// prose (that is Problem.Detail) — it is the enum's own label.
func (k ProblemKind) String() string {
	if name, ok := problemKindNames[k]; ok {
		return name
	}

	return "ProblemKind(?)"
}

var problemKindNames = map[ProblemKind]string{
	ConfigInvalid:               "ConfigInvalid",
	HullSizeInvalid:             "HullSizeInvalid",
	DriveStageInvalid:           "DriveStageInvalid",
	DriveSizeInvalid:            "DriveSizeInvalid",
	DrivePotentialZero:          "DrivePotentialZero",
	DriveAboveTL:                "DriveAboveTL",
	DriveExceedsHullG:           "DriveExceedsHullG",
	NoPowerPlant:                "NoPowerPlant",
	PowerBelowDrives:            "PowerBelowDrives",
	ComponentAboveShipTL:        "ComponentAboveShipTL",
	OverBudget:                  "OverBudget",
	MountsExceedHull:            "MountsExceedHull",
	UnknownWeapon:               "UnknownWeapon",
	WeaponMountIncompatible:     "WeaponMountIncompatible",
	MountTooSmall:               "MountTooSmall",
	RangeIncompatible:           "RangeIncompatible",
	UnknownDefenseMountOrRange:  "UnknownDefenseMountOrRange",
	DefenseRangeIncrease:        "DefenseRangeIncrease",
	WeaponNotADefense:           "WeaponNotADefense",
	UnknownDefense:              "UnknownDefense",
	MissileRoundSizeMismatch:    "MissileRoundSizeMismatch",
	WarheadUnavailable:          "WarheadUnavailable",
	MissileRoundTypeMismatch:    "MissileRoundTypeMismatch",
	MissileGuidanceIncompatible: "MissileGuidanceIncompatible",
	MissileGuidanceUnavailable:  "MissileGuidanceUnavailable",
}
