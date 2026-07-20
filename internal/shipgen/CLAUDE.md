# shipgen

Adventure Class Ship design (Book 2 pp. 30-95). Ship design is **deterministic**, not rolled: the
designer chooses tonnage/TL/config/structure/armor/drives and `Design(spec) ShipSpec` computes the
derived performance. `Design` is total — never an error; infeasibility (over-budget, underpowered
plant, TL-capped drive) is reported in `Ship.Problems`. The core insight: the p.78 Z1 Drive
Potential grid is a clean formula, `drivePotential = min(2*driveOrd/hullOrd, 9)`
(`DriveForPotential` is the Z2 inverse). Hull/drive letters are the eHex letters (A-Z, no I/O) as
an ordinal 1..24 (Hull A=100t … Z=2400t), distinct from an eHex value. `Ship.QSP()` renders the
compact profile (`S-AL22`, the ship's UWP analog); golden-locked to the Murphy Scout and Beowulf.
Costs are plain int Cr.

**Armament** (`weapon.go`/`weapon_data.go`, `defense.go`, `missile.go`, `mounts.go`) is the same
shape: Book 2 p.83 carries the whole weapon-design system as six tables (devices, TL stage
effects, mounts, space/world range effects), and p.174 prints the same six for defenses — so one
rule, `install`, scales both. The stage shifts the device's TL and prices it; the range shifts the
TL again and scales the **mount's** tonnage and cost ("Range Effects apply to the Mount but not the
Weapon" — a weapon's own tonnage is zero, so the mount is what takes up room). The mount also
supplies the damage dice and the attack Mod, and the two mount tables run **opposite ways**: a
Single Turret attacks at −2 (p.83) and defends at +1 (p.174) — a bigger mount aims worse but
defends better. `DesignWeapon` (23 models), `DesignDefense` (10 screens, plus the nine weapons
p.174 allows in the Anti-Missile Defensive Fire mode), and `DesignMissile` (size 1–7 × warhead ×
guidance, each constraining the others) render the book's `LongName` — a weapon's UWP analog.

`MinTL`/`MaxTL` bound the designable Tech Level (0..21, Book 2 p.51 — 21 is the design system's own
ceiling, distinct from the TL-15 Imperial shipyard limit). `Design` does not enforce them (it is
total); they exist for callers taking a TL from outside, since a negative one renders a ship card
whose TL field is not an eHex value at all.

`Design` mounts them against the hull: one HardPoint per 100t, or three FirmPoints instead (sub-ton
mounts only), with the Bolt-In needing neither — which is why `Tonnage` is fixed-point hundredths.
Golden-locked to the p.167 and p.176 catalogs, every row.

Gotchas: the book divides by three by multiplying by **0.33** (a 200t Main at Vlong is "66 tons"),
so range multipliers are hundredths — do not "fix" this to exact division; and the drive stage
table and the weapon one stay **separate tables that happen to agree**, because each is printed
both ways in the book and each is settled by its own worked examples (**Modified** costs /2 on both
sides: pp.104/127/134/190 for drives against the x1 of pp.63/76, pp.83/225/226/251 for weapons
against the p.279 appendix). The drive side is a **majority** reading, not a clean one: p.48's
sample-ship notes work Modified at x1, note 14 saying "same pricing per ton" in prose. Four
printings and two self-reconciling worked columns outweigh two printings and two notes — but the
book does not agree with itself, so do not re-open this on finding p.48 (#300 was mis-resolved
twice that way). The cell is asserted in **both** catalogs — `TestDesignDriveStageCatalogP127`
(MCr4) and `…P134` (Cr2,000,000) — so a revert to x1 fails two tests, not one. That is deliberate:
the note above cites p.127 and p.134 as two independent columns, and for a while only p.127
actually held the cell, so a revert would have failed once while the note went on claiming two
backed it. Do not "simplify" either assertion away, and never re-introduce an `mcr == 0`-style
sentinel that opts the row out of the check — that hatch existed, and it was a documented way to
silently un-resolve the conflict.

Drive stage tonnage **rounds up** and there is no tonnage floor — p.77's "no drive may be smaller
than the Drive-A of the class" is a floor on the size **letter**, which is the only reading under
which the worked tables' seven sub-Drive-A rows reproduce. Book conflicts are resolved against the
design tables and documented at the point of transcription; the p.127 and p.134 stage columns are
golden-locked cell by cell.

Deferred: sensors, crew/accommodations, Quality, Batteries, world-surface defenses, and the
pp.168–169 interference grid.
