package vehiclegen

import "testing"

// TestStdMilitaryCatalogGolden reproduces Std military vehicles from the Book 3
// p.140 catalog, transcribed from the rendered page. Each catalog line prints
// TL Tons Speed Load Armor Cage Flash Rad Sound Psi Ins Sealed Note KCr.
//
// Chart-reproducible columns (asserted exactly): TechLevel, Speed, Load, Armor,
// Cage, FlashProof, PsiShield. The remaining printed columns carry two
// catalog-wide constants the creation charts do not produce, asserted here as
// invariants rather than silently skipped:
//
//   - Tons prints chart value +1, and RadProof/SoundProof/Insulated/Sealed each
//     print chart value +1, uniformly across all 25 military rows. The catalog
//     vehicles carry an installed weapon (p.153 "Additional Steps: Create
//     weapons for Vehicle Weapons Mounts"; NoteT/NoteC/NoteV on p.150), which
//     is created outside VehicleMaker's charts — the +1 ton is its hull space
//     and the +1 protections ride with it.
//   - KCr is NOT asserted: it embeds the WeaponMaker cost of the installed
//     weapons (for example TW-G prints 6620 against a chart cost of 1870), which
//     is out of scope for this package.
//
// The picks isolate specific chart cells: MVR-T vs MVR-W differ only by the
// motive multiplier (KCr 200 = (100+100)x2 - (100+100)x1, proving Type+Mission
// then Motive order), MCS-W exercises Supply's Speed -1, MCT-L and the -L
// designs exist only under the Legged reading of the duplicated "W Wheeled"
// motive row, and TW-G exercises the Weapon Mission with the Grav x3 motive.
func TestStdMilitaryCatalogGolden(t *testing.T) {
	protected, _ := Enhancer("Protected")
	armored, _ := Enhancer("Armored")
	upArmored, _ := Enhancer("UpArmored")
	altArmored, _ := Enhancer("AltArmored")

	pa := []Modifier{protected, armored}
	pUpA := []Modifier{protected, upArmored}
	pAltA := []Modifier{protected, altArmored}

	for _, c := range []struct {
		name                string
		spec                Spec
		tl, speed           int
		load                float64
		armor, cage, flash  int
		tons                float64 // catalog Tons (chart value +1)
		rad, sound, ins, sl int     // catalog values (chart value +1)
	}{
		// Std MVR-T Recon MilVeh -T (P)(A) 10 6 6 1 50 30 30 31 45 0 51 61 0 2020
		{
			"MVR-T",
			Spec{Category: "M", Type: "V", Mission: "R", Motive: "T", Enhancers: pa},
			10, 6, 1, 50, 30, 30, 6, 31, 45, 51, 61,
		},
		// Std MVR-W Recon MilVeh -W (P)(A) 9 4 7 1 50 30 30 31 45 0 51 61 0 1820
		{
			"MVR-W",
			Spec{Category: "M", Type: "V", Mission: "R", Motive: "W", Enhancers: pa},
			9, 7, 1, 50, 30, 30, 4, 31, 45, 51, 61,
		},
		// Std MCS-W Supply MilCarrier -W (P)(AltA) 10 11 4 3 100 40 50 51 63 0 61 71 0 2720
		{
			"MCS-W",
			Spec{Category: "M", Type: "C", Mission: "S", Motive: "W", Enhancers: pAltA},
			10, 4, 3, 100, 40, 50, 11, 51, 63, 61, 71,
		},
		// Std MCT-L Troop MilCarrier -L (P)(A) 9 8 5 2 70 30 30 31 45 0 51 61 0 2720
		{
			"MCT-L",
			Spec{Category: "M", Type: "C", Mission: "T", Motive: "L", Enhancers: pa},
			9, 5, 2, 70, 30, 30, 8, 31, 45, 51, 61,
		},
		// Std MCT-W Troop MilCarrier -W (P)(UpA) 10 9 5 2 80 40 40 41 53 0 61 61 0 2720
		{
			"MCT-W",
			Spec{Category: "M", Type: "C", Mission: "T", Motive: "W", Enhancers: pUpA},
			10, 5, 2, 80, 40, 40, 9, 41, 53, 61, 61,
		},
		// Std TW-G Combat Tank -Grav (P)(AltA) 14 10 4 0 120 40 50 51 63 0 61 71 0 6620
		{
			"TW-G",
			Spec{Category: "M", Type: "T", Mission: "W", Motive: "G", Enhancers: pAltA},
			14, 4, 0, 120, 40, 50, 10, 51, 63, 61, 71,
		},
		// Std RS-W Supply MilTrlr -W (P)(AltA) 10 12 4 5 90 40 50 51 63 0 61 71 0 470
		{
			"RS-W",
			Spec{Category: "M", Type: "R", Mission: "S", Motive: "W", Enhancers: pAltA},
			10, 4, 5, 90, 40, 50, 12, 51, 63, 61, 71,
		},
	} {
		v := Design(c.spec)
		if len(v.Problems) != 0 {
			t.Fatalf("%s: problems %v", c.name, v.Problems)
		}

		if v.TechLevel != c.tl || v.Speed != c.speed || v.Load != c.load ||
			v.Armor != c.armor || v.Cage != c.cage || v.Flash != c.flash || v.Psi != 0 {
			t.Errorf("%s = %+v, want TL %d Speed %d Load %g Armor %d Cage %d Flash %d Psi 0",
				c.name, v.Values, c.tl, c.speed, c.load, c.armor, c.cage, c.flash)
		}

		// Catalog-wide installed-weapon invariant: printed value = chart value +1.
		if v.Tons+1 != c.tons || v.Radiation+1 != c.rad || v.Sound+1 != c.sound ||
			v.Insulated+1 != c.ins || v.Sealed+1 != c.sl {
			t.Errorf("%s = %+v, want catalog Tons %g Rad %d Sound %d Ins %d Sealed %d at chart+1",
				c.name, v.Values, c.tons, c.rad, c.sound, c.ins, c.sl)
		}
	}

	// The motive-multiplier proof from the catalog pair: MVR-T minus MVR-W is
	// exactly one extra (Type+Mission) cost, KCr 200.
	tracked := Design(Spec{Category: "M", Type: "V", Mission: "R", Motive: "T"})

	wheeled := Design(Spec{Category: "M", Type: "V", Mission: "R", Motive: "W"})
	if tracked.CostKCr-wheeled.CostKCr != 200 {
		t.Errorf("MVR-T minus MVR-W cost = %g, want 200", tracked.CostKCr-wheeled.CostKCr)
	}
}
