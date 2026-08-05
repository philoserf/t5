# VehicleMaker creation charts (Book 3 pp.150-153)

> **Verified against the rendered PDF** (`pdftoppm -r 200`, Book 3 pp.150-153), 2026-08-05.
> `docs/reference/`'s plain-text extract is column-lossy for these four dense pages — its table
> cells are scrambled out of column order — so it must not be used to re-check them; render the
> pages instead (see [Extraction workflow](#extraction-workflow) below). This file is the source
> of truth `internal/vehiclegen`'s `typeRows`/`missionRows`/`motiveRows`/`enhancerRows` and
> `enduranceModifier` mirror. Every cell below was independently transcribed from the rendered
> page image, then cross-checked against the registry — zero cell-value disagreements found. The
> [Errata](#errata--resolved-conflicts) section below records every settled ruling for these
> charts, most already known from `internal/vehiclegen/CLAUDE.md`; only its first entry, a stale
> page citation, is new to this pass.

A blank cell (`-`) is a genuine no-op in the printed chart: the book itself uses blank and a
literal `0` interchangeably to mean "no change," never "set to zero" (see Errata). `×N`/`/N`
notation is the book's own multiplier form.

## Chart 10 — Ground Vehicles (p.150)

### G — Ground Vehicles

**A. Type**

| Code | Descriptor | TL  | Tons | Speed | Load | KCr,000 |
| ---- | ---------- | --- | ---- | ----- | ---- | ------- |
| GC   | GroundCar  | -   | 2    | -     | 1    | 20      |
| U    | Utility    | -   | 3    | -     | 2    | 30      |
| T    | Truck      | -   | 4    | -     | 3    | 50      |
| V    | Vehicle    | -   | 5    | -     | 3    | 60      |
| M    | Mover      | -   | 3    | -     | -    | 50      |
| H    | Hauler     | -   | 5    | -     | 4    | 40      |
| R    | Trailer    | -   | 5    | -     | 4    | 40      |

**B. Mission**

| Code | Descriptor    | Armor | Insulated | KCr,000 |
| ---- | ------------- | ----- | --------- | ------- |
| RO   | Road Only     | -     | -         | -       |
| P    | Passenger     | 5     | 12        | 10      |
| C    | Cargo         | 5     | 6         | 10      |
| MP   | Multi-Purpose | 5     | 6         | 10      |
| OR\* | Off Road      | 20    | 20        | 100     |

OR also sets Cage 10, FlashProof 10, RadProof 10, SoundProof 10, Sealed 20.
\*OR Off Road: if Type=Vehicle, substitute ST (Some Terrain) if Motive=Air Cushion, RT (Rough
Terrain) if Motive=Legged, AT (All Terrain) if Motive=Tracked, or MT (Most Terrain) if
Motive=Wheeled.

**C. Motive**

| Code | Descriptor  | TL (set) | Tons | Speed | KCr,000 |
| ---- | ----------- | -------- | ---- | ----- | ------- |
| AC   | Air Cushion | 8        | +2   | 6     | ×2      |
| W    | Wheeled     | 6        | 0    | 5     | ×1      |
| Z    | Lifter      | 9        | +1   | 3     | ×2      |
| G    | Grav        | 10       | -1   | 5     | ×3      |
| T    | Tracked     | 7        | +2   | 4     | ×2      |
| W    | Wheeled     | 6        | 0    | 5     | ×1      |

Quality = 5 + Motive TL − Actual TL. Tech Level is determined by Motive. The printed Motive list
repeats "W Wheeled" as its sixth row — see [Errata](#errata--resolved-conflicts).

### M — Military Vehicles

**A. Type**

| Code | Descriptor | Tons | Speed | Load | Armor | Cage | FlashProof | RadProof | SoundProof | PsiShield | Insulated | Sealed | Note | KCr,000 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| T (also AV, AFV) | Tank | 5 | -1 | - | 50 | 10 | 10 | 10 | 20 | 0 | 20 | 20 | NoteT | 500 |
| C | Carrier | 4 | 0 | 2 | 40 | 10 | 10 | 10 | 20 | 0 | 20 | 20 | NoteC | 300 |
| V | Vehicle | 2 | +1 | 1 | 30 | 10 | 10 | 10 | 20 | 0 | 20 | 20 | NoteV | 100 |
| R | Trailer | 5 | 0 | 4 | 30 | 10 | 10 | 10 | 20 | 0 | 20 | 20 | - | 50 |

**B. Mission**

| Code | Descriptor | Tons | Speed | Load | Armor | Note  | KCr,000 |
| ---- | ---------- | ---- | ----- | ---- | ----- | ----- | ------- |
| W    | Weapon     | +2   | -     | -    | -     | NoteV | 100     |
| T    | Troop      | +1   | -     | -    | -     | -     | -       |
| S    | Supply     | +3   | -1    | +1   | -10   | -     | -       |
| R    | Recon      | -1   | +1    | -    | -10   | -     | 100     |

**C. Motive** — identical to G's Motive table (AC/W/Z/G/T/W, same values).

Quality = 5 + Motive TL − Actual TL. A Military vehicle automatically has weapons mount
capabilities. Technically, Tank is only used with Weapon; otherwise it is called Armored Vehicle
or Armored Fighting Vehicle. NoteT: install TWO weapons (one Vehicle-Mount, one Turret-Mount).
NoteC: install ONE turret mount weapon. NoteV: install ONE fixed mount weapon (supersedes NoteT
or NoteC).

## Chart 11 — Flyers and Boats (p.151)

### F — Flyers

**A. Type**

| Code | Descriptor | Note   |
| ---- | ---------- | ------ |
| F    | Flyer      | -      |
| G    | Glider     | Note G |
| B    | Balloon    | Note B |

**B. Mission**

| Code | Descriptor | TL  | Tons | Speed | Load | Armor | FlashProof | RadProof | SoundProof | Insulated | Sealed | KCr,000 |
| ---- | ---------- | --- | ---- | ----- | ---- | ----- | ---------- | -------- | ---------- | --------- | ------ | ------- |
| A    | Attack     | +2  | ×2   | +1    | ×2   | 20    | 20         | 1        | 20         | 10        | 1      | ×3      |
| B    | Bomber     | +1  | ×3   | -     | ×3   | 10    | 20         | 1        | 20         | 10        | 1      | ×2      |
| C    | Cargo      | 0   | ×4   | 0     | ×2   | 5     | 20         | 1        | 20         | 10        | 1      | ×1      |
| P    | Protector  | +1  | ×2   | +1    | ×1   | 10    | 20         | 1        | 20         | 10        | 1      | ×3      |
| S    | Scientific | -1  | ×4   | 0     | ×2   | 5     | 20         | 1        | 20         | 10        | 1      | ×2      |
| U    | Utility    | -   | -    | -     | ×3   | 0     | 20         | 1        | 20         | 10        | 1      | ×10     |

**C. Motive**

| Code | Descriptor       | TL (set) | Tons | Speed | Load | KCr,000 |
| ---- | ---------------- | -------- | ---- | ----- | ---- | ------- |
| W    | Winged           | 7        | 10   | 8     | 2    | 300     |
| R    | Rotor            | 8        | 10   | 7     | 0.5  | 400     |
| F    | Flapper          | 10       | 10   | 6     | 0.5  | 500     |
| LTA  | Lighter-Than-Air | 6        | 40   | 5     | 10   | 600     |
| Z    | Lifter           | 9        | 8    | 2     | 1    | 600     |
| G    | Grav             | 10       | 9    | 4     | 3    | 700     |

Quality = 5 + Motive TL − Actual TL, determined by Motive selection. **Lighter-Than-Atmosphere.**
LTA final tonnage equals 10× the calculated tonnage. Note G: requires Motive=Winged and
unpowered. Note B: requires Lighter-Than-Air and unpowered.

### W — Watercraft

**A. Type**

| Code | Descriptor | TL  | Tons | Speed | Load | Armor | KCr,000 |
| ---- | ---------- | --- | ---- | ----- | ---- | ----- | ------- |
| S    | Ship       | 5   | 1000 | 4     | 600  | 10    | 1,000   |
| U    | Sub        | 6   | 100  | 4     | 60   | 20    | 1,000   |
| B    | Boat       | 5   | 10   | 4     | 6    | 5     | 100     |

(Sub also sets Sealed 20; Ship sets Sealed 0/no-op.)

**B. Mission**

| Code | Descriptor | TL  | Speed | Armor | Insulated |
| ---- | ---------- | --- | ----- | ----- | --------- |
| C    | Cargo      | -   | -1    | -     | -         |
| P    | Patrol     | +2  | +1    | ×2    | -         |
| E    | Explorer   | +2  | -     | -     | ×2        |
| T    | Transport  | -   | -     | -     | -         |

**C. Motive**

| Code | Descriptor | TL  | Speed   | Load | KCr,000 |
| ---- | ---------- | --- | ------- | ---- | ------- |
| S    | Standard   | -   | -       | -    | -       |
| U    | Unpowered  | -3  | 0 (set) | -    | /2      |
| H    | Hovercraft | 6   | ×2, +5  | 3    | 200     |
| G    | Grav       | 10  | /5, +4  | \*\* | ×2      |

Quality = 5 + Motive TL − Actual TL. \*\* = No Change.

## Chart 12 — Vehicle Enhancers (p.152)

All 43 rows below are pinned cell-by-cell (independent of the registry) by
`TestEnhancerColumnsAgainstPage` in `internal/vehiclegen/charts_test.go`; this table mirrors that
test's `want` map, cross-checked here against the rendered page image. Column order: TL, Tons,
Speed, Load, Armor, Cage, FlashProof, RadProof, SoundProof, PsiShield, Insulated, Sealed, KCr.

**D. Bulk**

| Code | Descriptor | TL                           | Tons | Speed | Load | Armor | Cage | FlashProof | RadProof | SoundProof | PsiShield | Insulated | Sealed | KCr,000 |
| ---- | ---------- | ---------------------------- | ---- | ----- | ---- | ----- | ---- | ---------- | -------- | ---------- | --------- | --------- | ------ | ------- |
| Vl   | Vlight     | -1                           | /3   | +1    | -2   | /3    | -    | -          | /3       | -          | -         | /3        | /3     | /3      |
| L    | Light      | -1                           | /2   | +1    | -1   | /2    | -    | -          | /2       | -          | -         | /2        | /2     | /2      |
| M    | Medium     | (blank — the reference bulk) |      |       |      |       |      |            |          |            |           |           |        |         |
| H    | Heavy      | +1                           | ×2   | -1    | +2   | ×2    | -    | -          | ×2       | ×2         | -         | ×2        | ×2     | ×3      |
| Vh   | VHeavy     | +2                           | ×3   | -2    | +3   | ×3    | -    | -          | ×2       | ×2         | -         | ×3        | ×3     | ×9      |

**E. Stage**

| Code | Descriptor | TL  | Tons | Speed | Load | Armor | SoundProof | KCr,000 |
| ---- | ---------- | --- | ---- | ----- | ---- | ----- | ---------- | ------- |
| Fos  | Fossil     | -2  | +2   | -     | -    | -10   | -10        | -       |
| PC   | PowerCell  | -1  | +1   | -2    | -2   | -5    | -5         | 10      |
| Ren  | Renewable  | -1  | +1   | -1    | -1   | -     | -          | 20      |
| Pro  | Prototype  | -2  | +1   | -1    | -1   | -     | -          | 20      |
| Ear  | Early      | -1  | +1   | -     | -    | -10   | -10        | 10      |
| Std  | Standard   | 0   | 0    | -     | -    | -     | -          | -       |
| Imp  | Improved   | +1  | -1   | -     | -    | +10   | +10        | 20      |
| Adv  | Advanced   | +3  | -2   | +1    | +1   | +20   | +20        | 40      |

**F. Environ**

| Code         | Descriptor   | TL  | Tons | Armor | Cage | FlashProof | SoundProof | PsiShield | Insulated | Sealed | KCr,000 |
| ------------ | ------------ | --- | ---- | ----- | ---- | ---------- | ---------- | --------- | --------- | ------ | ------- |
| Air (Open)   | Air (Open)   | -2  | 0    | -     | -    | -          | -          | -         | -         | -      | -       |
| Enclosed     | Enclosed     | -1  | -    | 4     | -    | 4          | 4          | -         | 12        | -      | -       |
| Sealed       | Sealed       | -   | 0    | 6     | 2    | 6          | 8          | 0         | 16        | 20     | 2       |
| DoubleSealed | DoubleSealed | -   | +1   | 8     | 4    | 6          | 12         | 0         | 30        | 20     | 5       |
| Insulated    | Insulated    | -   | -    | 8     | 4    | 6          | 12         | -         | 30        | 20     | 10      |
| Protected    | Protected    | +1  | +1   | 10    | 10   | 10         | 12         | 0         | 10        | 20     | 20      |
| Armored      | Armored      | +2  | +1   | 20    | 10   | 10         | 12         | 0         | 20        | 20     | 30      |
| UpArmored    | UpArmored    | +3  | +2   | 30    | 20   | 20         | 20         | 0         | 30        | 20     | 40      |
| AltArmored   | AltArmored   | +3  | +2   | 60    | 20   | 30         | 30         | 0         | 30        | 30     | 50      |

Sealed's PsiShield/RadProof cells are printed `0` (no-op), not blank — see
[Errata](#errata--resolved-conflicts).

**G. Options**

| Code            | Descriptor       | TL  | Tons | Speed | Load | Note      | KCr,000 |
| --------------- | ---------------- | --- | ---- | ----- | ---- | --------- | ------- |
| HighPowered     | High Powered     | +1  | +1   | +1    | -1   | -         | 100     |
| Slave           | Slave            | +1  | -1   | -     | -    | -         | 10      |
| Remote          | Remote           | +1  | -2   | -     | -    | -         | 20      |
| WeaponMount     | Weapon Mount     | -   | -    | -     | -1   | -         | -       |
| Luxury          | Luxury           | -   | -    | -     | -    | Q= 9 or A | ×2      |
| Fast            | Fast             | +1  | +1   | +1    | -2   | -         | 30      |
| PassengerModule | Passenger Module | -   | -    | -     | -3   | 20 pass   | 100     |
| CargoModule     | Cargo Module     | -   | +1   | -1    | +1   | one ton   | 20      |
| Redundancy      | Redundancy       | +1  | +1   | -     | -    | -         | 60      |

**J. More Options**

| Code           | Descriptor         | Applies to | TL  | Tons | Speed    | Load | Note           | KCr,000 |
| -------------- | ------------------ | ---------- | --- | ---- | -------- | ---- | -------------- | ------- |
| OffRoad        | OffRoad            | ground     | -   | -    | -        | -    | Wheeled        | 30      |
| Mole           | Mole               | ground     | +1  | ×3   | =1 (set) | -    | Note 2         | 400     |
| Hydrofoils     | HydroFoils         | water      | +1  | +1   | +1       | -    | -              | 30      |
| Stubs          | Stubs              | flyer      | -   | -    | -        | -    | Grav or Lifter | 20      |
| VTOL           | VTOL Mod           | flyer      | -   | -    | -1       | -2   | Med or less    | 100     |
| STOL           | STOL Mod           | flyer      | -   | -    | -        | -1   | Hvy or less    | 50      |
| LiftingBody    | Lifting Body Hull  | flyer      | -   | +4   | +1       | ×2   | -              | 200     |
| Wings1         | Add-On Wings-1     | flyer      | -   | ×2   | +1       | ×1   | B= +1          | 100     |
| Wings2         | Add-On Wings-2     | flyer      | -   | ×3   | +2       | ×2   | B= +1          | 200     |
| Wings3         | Add On Wings-3     | flyer      | -   | ×4   | +3       | ×3   | B= +1          | 300     |
| Floats         | Float Landing Gear | flyer      | -   | -1   | -1       | -    | -              | 100     |
| ParasiteNipple | Parasite Nipple    | flyer      | +1  | -    | -        | -1   | -              | 100     |

Note1 (Stage rows Fossil/PowerCell/Renewable): may not be Grav or Lifter. Note 2 (Mole): only if
Ground Vehicle, Explorer, not ACV.

## Chart 13 — Vehicle Design Checklist / Endurance (p.153)

**H. Endurance** (default is Hours)

| Code | Descriptor | TL  | Tons (×Speed) | KCr,000 |
| ---- | ---------- | --- | ------------- | ------- |
| -    | Hours      | -   | 0\*           | -       |
| -    | Days       | +1  | 1\*           | 20      |
| -    | Weeks      | +2  | 2\*           | 50      |
| LR   | Months     | +3  | 3\*           | 100     |
| VLR  | Year       | +4  | 4\*           | -       |

\*this value times Vehicle Speed. The printed Year row has no KCr value.

**Convert Endurance to Range** (Speed 1-13 ↔ Kph 5-5000)

| Speed → | 1           | 2        | 3        | 4           | 5           | 6           | 7           | 8           | 9           | 10    | 11    | 12    | 13    |
| ------- | ----------- | -------- | -------- | ----------- | ----------- | ----------- | ----------- | ----------- | ----------- | ----- | ----- | ----- | ----- |
| Kph     | 5           | 10       | 20       | 30          | 50          | 100         | 300         | 500         | 700         | 1000  | 2000  | 3000  | 5000  |
| Minutes | Local       | Local    | Local    | Local       | Local       | Local       | Local       | Local       | Local       | Local | Local | Local | Local |
| Hours   | Local       | Local    | Local    | Local       | Regional    | Regional    | Continental | Continental | Continental | World | World | World | World |
| Days    | Regional    | Regional | Regional | Continental | Continental | Continental | World       | World       | World       | World | World | World | World |
| Weeks   | Continental | World    | World    | World       | World       | World       | World       | World       | World       | World | World | World | World |
| Months  | Continental | World    | World    | World       | World       | World       | World       | World       | World       | World | World | World | World |
| Year    | World       | World    | World    | World       | World       | World       | World       | World       | World       | World | World | World | World |

Minutes always renders Local regardless of speed (the printed grid leaves every Minutes cell
blank; `EnduranceRange` returns Local unconditionally for it, matching "the default reading of an
otherwise-unfilled row").

**Vehicle Category (Charts 10-11):** G Ground Vehicle, M Military Vehicle, F Flyer, W Water Craft.
**Type-Mission-Motive (Charts 10-11):** A Type, B Mission, C Motive.
**Vehicle Enhancers (Chart 12-13):** D Bulk, E Stage, F Environ, G Options, H Endurance (default
Hours), J Special Options.

## Draw order (`internal/vehiclegen.Generate`)

`Generate` draws exactly 3 or 4 dice, via `dice.Roller.Index`, in this order:

1. **category** — `r.Index(3)` over `{G, F, W}` (never M: `Generate` never rolls a Military
   vehicle, since nothing selects it — see [Deferred](#deferred--out-of-scope)).
2. **type** — `r.Index(len(types))` over the category's Chart 10/11 A rows.
3. **mission** — `r.Index(len(missions))` over the category's B rows.
4. **motive** — `r.Index(len(motives))` over the category's C rows, **skipped** when category=F
   and type=G (Note G forces Motive=Winged) or type=B (Note B forces Motive=LTA): those two cases
   draw only 3 dice total, not 4.

`Design` itself draws no dice. Bulk/Stage/Environ/Options/Endurance (Charts 12-13) are never
randomly rolled by this package — `Generate` never populates `Spec.Enhancers` or `Spec.Endurance`,
so a `Generate`d vehicle always has the Standard stage/no enhancers/Hours endurance implicitly.
Enhancers and Endurance are supplied by the caller's `Spec` and applied deterministically by
`Design`; there is no draw order to audit for them because nothing draws for them.

## Deferred / out of scope

- **Military category (M) is never rolled by `Generate`** — only reachable via an explicit
  `Spec{Category: "M", ...}`. `Generate`'s category roll only covers `{G, F, W}`. Not a bug (M
  requires a Type/Mission/Motive combination a random roll has no policy for choosing sensibly —
  e.g. Weapon-as-Mission needing a compatible Type), but worth naming here since a reader of the
  chart might expect all four categories to be reachable from `Generate` alone.
- Weapon creation for Vehicle Weapon Mounts, on-board brain installation (Vehicle Operations),
  and QREBS aging are Chart 13's "Additional Steps" — explicitly out of scope for this package
  per its own CLAUDE.md ("Weapon creation and on-board brains belong to their own makers").
  Depends on GunMaker/ArmorMaker (#177).
- Vehicle FillForm/HitForm (Charts 14-15) are paper-form bookkeeping, not calculated data — no
  Go representation is needed.
- p.153's Additional Steps (installed weapon tonnage/protections/cost onto a generated vehicle)
  is tracked separately as #365.

## Errata / resolved conflicts

Full reasoning for each of these lives in `internal/vehiclegen/CLAUDE.md`'s "Settled rulings"
section; this is a pointer-only ledger co-located with the table data itself, per #358.

- **Citation fix (found while writing this doc):** `internal/vehiclegen/CLAUDE.md` cited "the
  creation charts" as **pp.150-152**. Chart 13's H Endurance table (and the Convert-Endurance-to-
  Range grid) is on **p.153**, one page outside that range. Corrected to pp.150-153 in this pass
  — the same class of inaccuracy #312 tracks, found incidentally rather than by that issue's
  planned repo-wide sweep.
- **The duplicated "W Wheeled" Motive row** (Chart 10, both G and M) is read as **L Legged** —
  an erratum, not a transcription error. See CLAUDE.md for the full evidence (the OR\* footnote's
  "-Legged" reference; the Std military catalog's -L designs reproducing exactly under this
  reading).
- **A printed `0` is a no-op, never a reset** (Chart 12 Sealed's RadProof/PsiShield cells, Chart
  10 Motive Wheeled's Tons cell). Reading a printed 0 as "set to 0" would erase values a later row
  needs preserved (e.g. a Tank's Radiation 10).
- **Weapon is a Mission, not a Type** (Chart 10 M section B) — confirmed by the p.140 catalog's
  "TW" Combat Tank (Tank Type + Weapon Mission).
- **LTA final tonnage is ×10** the calculated tonnage (p.151 footnote), applied after every chart
  row including Endurance.
- **Apply order is per category** (Type→Mission→Motive for Ground/Military, Motive→Mission for
  Flyers, Type→Motive→Mission for Watercraft) — proven against the catalog's own cost arithmetic;
  see CLAUDE.md.

## Extraction workflow

`docs/reference/`'s plain `pdftotext` extract merges these charts' columns out of order — its
text is useful for locating a section by keyword search, never for reading a dense table's cell
values. For any chart this dense:

1. Find the true PDF page index for a cited printed page number — **do not assume they match**.
   `pdftotext -f N -l N` a candidate page and look for the "- NNN -" footer; Book 3's pages
   happened to align 1:1 in this pass (`docs/reference` split index N ⇔ printed page N), but that
   is a fact to verify per-book, not to assume.
2. Render the page(s) as an image: `pdftoppm -f N -l N -png -r 200 "<book>.pdf" /tmp/out`. 200 DPI
   is legible for dense tables; raise it if a column's text is still ambiguous.
3. Read the rendered image directly and transcribe cell by cell, in the chart's own column order.
4. Cross-check against the current Go registry (or an existing independently-transcribed test
   like `TestEnhancerColumnsAgainstPage`) rather than trusting either source alone — a mismatch
   between the two is exactly the signal worth investigating before touching either side.
