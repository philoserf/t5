package mastermod

// Damage location tables (Book 1 p.264): charts L1 (Thing), L2 (Vehicle or
// Equipment), and L3 (Character).
func init() { //nolint:gochecknoinits // declarative source data and registry validation
	register(
		table("Device Damage Location", "1D", 1, "Case", "Power", "Input", "Output", "Controls", "Processor"),
		table("Tool Damage Location", "1D", 1, "Case", "Power", "Adjuster", "Toolhead", "Grip", "Safety"),
		table("Weapon Damage Location", "1D", 1, "Frame", "Ammunition", "Sights", "Barrel", "Grip", "Mechanism"),
		table(
			"Heavy Weapon Damage Location",
			"2D",
			2,
			"Controls",
			"Mount",
			"Sights",
			"Shields",
			"Stocks",
			"Barrel",
			"Power",
			"Frame",
			"Ammunition",
			"Mechanism",
			"Computer",
		),
		table(
			"Vehicle or Armor Damage Location",
			"2D",
			2,
			"Controls",
			"Interior",
			"Visor",
			"Protections",
			"Life Support",
			"Locomotion",
			"Power",
			"Torso",
			"Manipulators",
			"Navigation",
			"Computer",
		),
		// Book 1 p.264 prints rows 9-10 as "Limb-Grip-3" and "Limb-Grip-4"
		// even though rows 3-4 are "Limb-Group-1/2" and the combat chapter
		// prints "Limb-Group-3/4". The appendix reading was chosen
		// deliberately — do not "fix" Grip to Group here.
		table(
			"Anatomical Damage Location",
			"2D",
			2,
			"Head",
			"Head",
			"Limb-Group-1",
			"Limb-Group-2",
			"Torso",
			"Torso",
			"Torso",
			"Limb-Grip-3",
			"Limb-Grip-4",
			"Graze",
			"Graze",
		),
		table(
			"Biological Damage Location",
			"2D",
			2,
			"Brain",
			"Senses",
			"Circulation",
			"Skeleton",
			"Respiration",
			"Skin",
			"Digestion",
			"Elimination",
			"Muscle",
			"Skin",
			"Skin",
		),
	)
}
