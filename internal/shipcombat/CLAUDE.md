# shipcombat

The Space Combat resolution engine (Book 2 pp. 193-204). The Space Weapon, Missile, and Defensive
Fire tasks are roll-low over `task.ResolveDice` (range-band or 5D dice; targets
`weaponTL+C+S+K+mods`, `missileTL+guidance+mods`, `defenseTL−attackTL+mountMod`); plus
`HitCompartment` (Flux + targeting), `Penetrate` (layered armor), the L1 damage-location table,
damage/diagnosis `Severity`, the missile `MassiveExplosion` proximity table, and movement
(`Agility`, `RammingHits`, the p.200 range-change grid).

The ShipCard compartment model (`HullLocations`, p.95 Table H — compartments/span/subcompartments
by hull ordinal) backs hit location and `SubCompartmentsKnockedOut` damage spread; the missile and
weapons-task Massive Explosion tables are both present.

The tasks still take primitives (TLs/mods/AV/compartment numbers) because that is what the book's
tables are — but nobody has to invent them: `ship.go` bridges the two packages, so `Attack` takes a
designed `shipgen.Weapon`, `Defend` a `Defense`, `AttackWithMissile` a round, and
`ArmorLayers`/`Card`/`ShipAgility` a `Ship`. **A generated ship can fight.** Golden-locked to the
Murphy/Gryphon, Vanguard/Antares, Joshua, and Vigilant worked examples.

Note `Armor.AV` is the **per-layer** value, not a total (`ArmorLayers` repeats it, and must not
divide by the layer count), and `ShipAgility` is capped by `Hull.MaxG`, not just the drive.

Out of scope: the p.201 interference/clustering tactical options.
