# Master Mod tables

`mastermod` contains named random-reference lookups from Book 1 pp.264–269.
`Table.Lookup` accepts the already-rolled total; it never rolls dice itself.
Tables retain the source's roll notation and exact contiguous bounds. Blank
source cells are not invented. Rows are book-literal text — the seven
`<none>` entries transcribe p.267's "Treat blank entries as `<none>`."
literally; there is no placeholder-substitution engine (removed as YAGNI: no
real `<…>` tokens exist in the data, and substitution would corrupt `<none>`).

Table files are split by printed page: `tables_p264.go` … `tables_p269.go`. **`docs/rules/mastermod.md`**
is the verified, page-by-page source of truth these files mirror (#358), including a full chart
index and cell-by-cell transcription of every quirk below — check it, not docs/reference/, when a
cell's correctness is in question.

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
- `Notation` strings not parseable by `dice.Parse` are allowlisted by exact
  string in the tests: `Bad Flux`, `2x1D`, `Hits/2`. A typo fails the suite.
- **Printed typos are preserved, not corrected** (verified against the
  rendered PDF pages, not the text extract): `Truthfullness` (p.267 — also
  printed that way in Book 3 p.279), `Cacaphony` (p.268), `More Burdensom`
  and `Very Burdensom` (p.268), and `Vfast Land` (p.266, directly under
  "Fast Lane").
  Rows are book-literal; a consumer wanting normalized English must map it
  itself. Note the divergence: `internal/epic` carries the same Themes table
  and deliberately normalizes `Truthfullness` → `Truthfulness` (see its
  CLAUDE.md) — that is epic's settled call, and this is mastermod's.
