# Master Mod Tables (Book 1 pp.264-269)

> **Verified against the rendered PDF** (`pdftoppm -r 200`, Book 1 pp.264-269), 2026-08-05.
> `docs/reference/`'s plain-text extract garbles these pages' columns out of order — do not use
> it to re-check a cell; render the page instead (see vehiclegen's
> [Extraction workflow](vehiclegen.md#extraction-workflow), which applies unchanged here). This
> file is the source of truth `internal/mastermod`'s `tables_p264.go` … `tables_p269.go` mirror.
>
> **Scope check performed while writing this doc:** the appendix's own index (chart "00", p.264)
> numbers its charts 00-21, which could be misread as 21 separate _pages_. It is not — charts are
> packed 4-to-a-page (p.264: index + L1/L2/L3; p.265: charts 01-04; p.266: 05-08; p.267: 09-12;
> p.268: 13-16; p.269: 17-21). The whole appendix is fully contained in pp.264-269, so
> `internal/mastermod`'s stated scope is already complete — there is no pp.270+ remainder deferred
> anywhere. Worth recording since the "00-21" numbering invites exactly this misreading.

## Chart index

Every chart implemented by `internal/mastermod`, by printed page and registered table name(s).
Dice notation and row span are the chart's own header (e.g. "1D" = 6 rows keyed 1-6; "Flux" here
always means the standard -5..+5 range unless noted **wide** below).

| Page | Charts                  | Dice                    | Registered table name(s)                                                                                                                          |
| ---- | ----------------------- | ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| 264  | 00 (index, not a table) | -                       | -                                                                                                                                                 |
| 264  | L1                      | 1D                      | Device/Tool/Weapon Damage Location                                                                                                                |
| 264  | L2                      | 2D                      | Heavy Weapon/Vehicle or Armor Damage Location                                                                                                     |
| 264  | L3                      | 2D                      | Anatomical/Biological Damage Location                                                                                                             |
| 265  | 01                      | Flux                    | Visibility, Environ, Attitude, Conformity, Imagination, Logic, Idea                                                                               |
| 265  | 02                      | 1D                      | Typical Mods (VLow…Uhigh, **Technology Fantastic**), Gravity, Zero-G, Acceleration, Environ (Bad Flux)                                            |
| 265  | 03                      | Flux **(wide, -6..+6)** | Probability, Severity, Imminence, Order and Chaos, Morality, Evidence and Proof                                                                   |
| 265  | 04                      | Flux                    | Comparatives, Vilani, NewSpeak, Imperial/Anglic/Vilani/MegaCorp Brand Names                                                                       |
| 266  | 05                      | 1D **(7 rows)**         | Groups1, Groups2; 6D/5D/4D/3D/2D/1D/0 Injury, Severity, Diagnosis, Mods, Mods (Bad Flux)                                                          |
| 266  | 06                      | 1D                      | Crime, Property, Environ, Sophonts, Society, Crime versus Justice, Doctrine (Good Flux)                                                           |
| 266  | 07                      | Flux **(wide, -6..+6)** | Careers, Nobility, Friends, Humaniti, Sophonts1, Sophonts2, Major Races                                                                           |
| 266  | 08                      | Flux                    | Gravity Size/G/Speed, Descriptive Speeds (Walk/Drive/Highway), Comms, Weather, Sounds                                                             |
| 267  | 09                      | 1D                      | Theme1-6                                                                                                                                          |
| 267  | 10                      | Flux                    | Planetary, Population, Economic, Climate, Secondary, Political, Special, Terrain1, Terrain2 (all `<none>`-at-zero)                                |
| 267  | 11                      | Flux                    | Emotion, Degree, Potential, Fraction, Rewards, Standard/Tuned-Naval Time In Jump                                                                  |
| 267  | 12                      | Flux **(wide, -6..+6)** | Supply, Demand, Uncertainty, Beauty, Truth, Respect                                                                                               |
| 268  | 13                      | Flux                    | Visibility, Light, Sound, Aroma, Sensation, Touch, Emotion, Large Groups                                                                          |
| 268  | 14                      | 1D                      | Barrier Height/Width/Depth (**blank, not transcribed**), Stability, Typical BR, Typical DH, Xeno-Med (Bad Flux)                                   |
| 268  | 15                      | 2D-2                    | QREBS Quality/Mod/Period; **Scene Mods is a formula, not a table**                                                                                |
| 268  | 16                      | Flux                    | Reliability, Ease Of Use, Bulk/Burden, Safety, Surprise                                                                                           |
| 269  | 17                      | 1D                      | Sensor Responding, Space/World/Specialized Sensors (Good Flux)                                                                                    |
| 269  | 18                      | 1D                      | Local Observation, Local Anomaly, Strange Warnings, Local Unrest (Good Flux)                                                                      |
| 269  | 19                      | 1D                      | Starport/Startown Situations, Stellar Anomalies, Minor Planets (Good Flux)                                                                        |
| 269  | 20                      | 1D                      | Outer System, Remote System, Historical Ruins, Reported Fault (Good Flux)                                                                         |
| 269  | 21                      | 2x1D                    | MegaCorporations (Q/R/E/B/S mods; **Imperiallines/Hortalez have no roll key**); Damage Severity (Hits/2) and Diagnosis Severity (1D) **(9 rows)** |

## Fully transcribed: the quirky tables

Every row below was read directly off the rendered page. These are the tables
`internal/mastermod/CLAUDE.md`'s "Quirks" section calls out, so they carry the most risk of a
future mistaken "fix" — full transcription puts the page evidence next to the registry it backs.

**L1/L3 Damage Location, rows 9-10 (p.264)** — Anatomical column: `Head, Head, Limb-Group-1,
Limb-Group-2, Torso, Torso, Torso, Limb-Grip-3, Limb-Grip-4, Graze, Graze` (2D, rows 2-12). Rows 9
and 10 print **"Limb-Grip-3/4"**, not "Limb-Group-3/4" as the combat chapter elsewhere spells it.

**Technology (chart 02, p.265)** — one printed header "Fantastic" over two roll-1 sub-columns
ending `… 22 27` and `… 28 33`: registered as two tables, "Technology Fantastic (Low)" (22-27)
and "Technology Fantastic (High)" (28-33).

**Groups1/Groups2 (chart 05, p.266)** — SEVEN rows (1-7) under a "1D" header, not six:
Individuals/Groups/Hundreds/Thousands/Ten Thousands/Hundred Thousands/Millions (Groups1);
Millions/Hundred Thousands/Ten Thousands/Thousands/Hundreds/Groups/Individuals (Groups2, the
mirror order).

**Diagnosis Severity (chart 21, p.269)** — NINE rows (1D key 1-9): Easy 1D, Average 2D, Difficult
3D, Formidable 4D, Staggering 5D, Hopeless 6D, Impossible 7D, Beyond 8D, Destroyed. (Damage
Severity, same page, uses the identical nine-row list keyed by Hits/2 instead of 1D.)

**Wide Flux tables (-6..+6, 13 rows instead of the usual -5..+5):** chart 03 (Probability et al.,
p.265), chart 07 (Careers et al., p.266), chart 12 (Supply et al., p.267). Confirmed by row count
on the rendered page for all three.

**Barrier Height/Width/Depth (chart 14, p.268)** — genuinely blank in the printed appendix; only
the Stability/Typical BR/Typical DH/Xeno-Med columns carry values. Nothing to transcribe, and
nothing is missing from the registry.

**Scene Mods (chart 15, p.268)** — not a die table at all: the page prints a boxed formula,
verbatim `Scene= Flux + Mods (ignore 0 or greater)`, with Mods C/E/EOU/Surprise pointing to other
charts (Starship Construction Chart 26, QREBS). `mastermod` correctly has no table for this.

**MegaCorporations (chart 21, p.269)** — every row except **Imperiallines** and **Hortalez**
carries a 2×1D key (e.g. `1-1 General`, `2-6 Sternmetal`); those two rows print Q/R/E/B/S mods
with both key columns blank, so they cannot be rolled and are correctly excluded from the table.

**Five preserved print typos**, each confirmed on its page: `Truthfullness` (chart 09 Theme1 row
5, p.267 — also Book 3 p.279), `Cacaphony` (chart 13 Sound row +3, p.268), `More Burdensom` and
`Very Burdensom` (chart 16 Bulk/Burden rows +4/+5, p.268), `Vfast Land` directly under "Fast Lane"
(chart 08 Highway row +2, p.266).

## Draw order

**None.** `mastermod.Table.Lookup` accepts an already-rolled total (the caller's own dice, rolled
however the caller's context requires) — it never calls a `*dice.Roller` itself, so there is
nothing to audit here the way `vehiclegen.Generate`'s roll sequence needs auditing. Dispatching a
table's own `Dice` notation (e.g. "2x1D", "Hits/2") through `dice.Parse`/`Eval` to actually
produce that total is the deferred roll bridge, #366 — deliberately not built ahead of a
consumer (YAGNI), per that issue's own text.

## Deferred / out of scope

Confirmed against the rendered pages in this pass — all three are genuine exclusions, not gaps:

- **Barrier Height/Width/Depth** (chart 14, p.268): blank cells in the printed appendix.
- **Scene Mods** (chart 15, p.268): a formula, not a die table.
- **Imperiallines and Hortalez** (chart 21, p.269): no 2×1D key printed, so unrollable.
- **The roll bridge** (dispatching a table's `Dice` string through `dice.Parse`/`Eval` to actually
  produce a Lookup-ready total) — tracked as #366, waiting on a first consumer.

## Errata / resolved conflicts

Full reasoning lives in `internal/mastermod/CLAUDE.md`'s "Quirks" section; this is a pointer-only
ledger co-located with the table data, per #358. Every item below was independently re-verified
against the rendered page while writing this doc — no new disagreements found.

- Anatomical Damage Location's Limb-Grip-3/4 reading (p.264) — deliberate, not a typo fix target.
- Technology Fantastic's Low/High split (p.265).
- Groups1/Groups2's seven rows (p.266) and Diagnosis/Damage Severity's nine rows (p.269).
- The three wide (-6..+6) Flux tables: charts 03/07/12.
- The five preserved print typos listed above, plus the divergence with `internal/epic`'s Themes
  table (which normalizes `Truthfullness` → `Truthfulness` as its own settled call — see that
  package's CLAUDE.md — while `mastermod` preserves the book-literal spelling).

No new citation or transcription errors were found in this pass (contrast `vehiclegen.md`, where
one turned up). `internal/mastermod`'s existing page citations (pp.264-269) are accurate as
written.
