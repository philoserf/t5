# vehiclegen

VehicleMaker (Book 3 pp.132-158). `Design` is the deterministic path: select Type, Mission, and
Motive rows for ground (`G`), military (`M`), flyer (`F`), or watercraft (`W`) and apply named
Chart 12 rows from `Enhancer`. `Generate` only chooses valid rows with an injected `dice.Roller`
and delegates to `Design`.

The calculated `Values` retain TL, tons, Speed, Load, Armor, Cage, Flash/Radiation/Sound/Psi,
Insulated, Sealed, and KCr. `SpeedKPH`, `CollisionDicePerTon`, `Beastpower`, `EnduranceRange`, and
`Occupants` are pure transcriptions of Charts 06-07. `DesignBoxRows`, `SurfaceAccess`,
`FlyerAccess`, `SeafaringAccess`, `Altitudes`, and `Depths` expose the readable pp.143-148
operating/design tables.

The dense tables contain blanks and conditional notes. A blank remains no operation; it is not
silently inferred, and Chart 12 restrictions remain in `Modifier.Note` rather than being silently
enforced without the additional state they require. Weapon creation and on-board brains belong to
their own makers; QREBS ageing and the non-human occupant ratios require user/body context and are
not part of the vehicle's scoped calculated columns.
