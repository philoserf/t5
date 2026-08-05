# vehiclegen

VehicleMaker (Book 3 pp.132-158). `Design` is the deterministic path: select Type, Mission, and
Motive rows for ground (`G`), military (`M`), flyer (`F`), or watercraft (`W`) and apply named
Chart 12 rows from `Enhancer`, then the Chart 13 Endurance row. `Generate` only chooses valid
rows with an injected `dice.Roller` and delegates to `Design`.

The calculated `Values` retain TL, tons, Speed, Load, Armor, Cage, Flash/Radiation/Sound/Psi,
Insulated, Sealed, and KCr. `SpeedKPH`, `CollisionDicePerTon`, `Beastpower`, `EnduranceRange`, and
`Occupants` are pure transcriptions of Charts 06-07. `DesignBoxRows`, `SurfaceAccess`,
`FlyerAccess`, `SeafaringAccess`, `Altitudes`, and `Depths` expose the readable pp.143-148
operating/design tables.

The creation charts (**pp.150-153** — corrected from a previously-cited pp.150-152 that missed
Chart 13's Endurance table on p.153, found while writing `docs/rules/vehiclegen.md`) are dense
and the docs/reference text extract is **column-lossy** for them — every cell here was
transcribed from the rendered PDF pages, and `TestEnhancerColumnsAgainstPage` pins all 43 Chart 12
rows with literals transcribed independently of the registry. `TestStdMilitaryCatalogGolden`
reproduces seven vehicles from the p.140 Std military catalog. **`docs/rules/vehiclegen.md`** is
the verified, cell-by-cell source of truth these registries mirror (#358) — check it, not
docs/reference/, when a cell's correctness is in question.

## Settled rulings (each verified against the rendered pages — do not re-open)

- **Apply order is per category.** Ground and military apply chart order A Type, B Mission,
  C Motive, so the Motive cost multiplier covers Type+Mission — proven by the catalog pair
  MVR-T − MVR-W = KCr 200 = (100+100)×2 − (100+100)×1. Flyers apply Motive first because the
  Mission rows are multipliers over the Motive's base columns (the Type rows are empty).
  Watercraft apply Type, Motive, Mission: the W Missions carry no Tons/KCr operation so cost is
  order-independent, but Patrol/Explorer TL +2 must land after the Motive TL replacement
  (Hovercraft 6 / Grav 10) or it would be erased.
- **Weapon is a military Mission, not a Type** (p.150 section B: W/T/S/R; Types are T/C/V/R).
  The catalog corroborates: "TW" Combat Tank = Tank Type + Weapon Mission, and p.150's footnote
  "Tank is only used with Weapon" only parses under this reading.
- **The duplicated "W Wheeled 6 0 5 ×1" motive row is read as L Legged** (printed twice in both
  the G and M motive lists — an erratum). Evidence: the OR* footnote references "-Legged", and the
  catalog's -L designs (MVR-L, MCS-L, MCT-L, RS-L, TW-L) reproduce exactly under TL 6 / Speed 5 /
  ×1.
- **A printed 0 is a no-op, never a reset.** The charts use 0 for "no change" (compare p.151's
  "** = No Change"); reading Chart 12 Sealed's Rad/Psi 0 as `set(0)` would destroy a Tank's
  Rad 10, and the catalog's Protected+armored rows keep their Radiation values.
- **LTA final tonnage is 10× the calculated tonnage** (p.151 footnote). "Final" means after every
  chart row including Endurance, so Beastpower and Occupants both see the ×10 figure.
- **Chart 13 Endurance rows are applied last.** Days/Weeks/Months/Year add TL +1..+4 and
  1..4 × Vehicle Speed tons, KCr 20/50/100 for Days/Weeks/Months; the printed Year row has no KCr.
  p.153 lists H between G Options and J Special Options, but `Spec.Enhancers` is a flat list, so
  the engine applies H after all enhancer rows — visible only when a J row changes Speed, which no
  book example exercises.
- **Beastpower column shifts** (p.146 box) apply to the BP _lookup_ speed only, never to the
  vehicle's Speed value: Vheavy/High Power/Protected +1, Armored +2, Vlite/Hydrofoil −1, and
  watercraft Ship +3 / Sub +2 / Boat +1.
- **The military catalog prints two constants the charts do not produce**: every p.140 row is
  chart value +1 in Tons and +1 in Rad/Sound/Insulated/Sealed (the installed weapon from the
  p.153 "Additional Steps", outside these charts), and the KCr column embeds WeaponMaker weapon
  costs. The golden asserts the +1s as catalog-wide invariants and does not assert KCr.

The dense tables contain blanks and conditional notes. A blank remains no operation; it is not
silently inferred, and Chart 12 restrictions remain in `Modifier.Note` rather than being silently
enforced without the additional state they require — except in `Generate`, which honors p.151's
Note G/Note B by pairing Glider only with the Winged motive and Balloon only with LTA. Weapon
creation and on-board brains belong to their own makers; QREBS ageing and the non-human occupant
ratios require user/body context and are not part of the vehicle's scoped calculated columns.
