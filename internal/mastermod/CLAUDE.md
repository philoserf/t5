# Master Mod tables

`mastermod` contains named random-reference lookups from Book 1 pp.264–269.
`Table.Lookup` accepts the already-rolled total; it never rolls dice itself.
Tables retain the source's roll notation and exact contiguous bounds. Blank
source cells are not invented. Rows are book-literal text — the seven
`<none>` entries transcribe p.267's "Treat blank entries as `<none>`."
literally; there is no placeholder-substitution engine (removed as YAGNI: no
real `<…>` tokens exist in the data, and substitution would corrupt `<none>`).

Table files are split by printed page: `tables_p264.go` … `tables_p269.go`.

## Deliberate exclusions — do not re-add

- **Barrier Height / Barrier Width / Barrier Depth** (chart 14, p.268): their
  cells are blank in the printed appendix; there is nothing to transcribe.
- **Scene Mods** (chart 15, p.268): a formula ("Scene = Flux + Mods"), not a
  die table.
- **Imperiallines and Hortalez** (chart 21, p.269): printed in the
  MegaCorporations table without 2x1D keys, so they cannot be rolled.

## Quirks that are faithful to the book — do not "fix"

- **Anatomical Damage Location** rows 9–10 are `Limb-Grip-3/4` (p.264 prints
  "Grip" there; the combat chapter prints "Limb-Group-3/4"). The appendix
  reading was chosen deliberately.
- **Technology Fantastic** (p.265) is one printed header over TWO value
  sub-columns (roll-1 row ends "… 22 27 33"); registered as
  "Technology Fantastic (Low)" (22–27) and "(High)" (28–33).
- **Groups1/Groups2** (p.266) have SEVEN rows under a "1D" header;
  **Diagnosis Severity** (p.269) has NINE.
- The chart 03/07/12 "Flux" tables (pp.265–267) run **-6..+6** — 13 rows,
  wider than plain Flux.
- `Dice` strings not parseable by `dice.Parse` are allowlisted by exact
  string in the tests: `Bad Flux`, `2x1D`, `Hits/2`. A typo fails the suite.
