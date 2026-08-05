# sophont

The Sophont Creation System (Book 3 pp. 217-239), the **core spine** that makes chargen work for
aliens. `Generate` composes Flux-driven sub-generators into a `Species` template: a plausible
homeworld (**reuses `worldgen`**, rerolled to Atm 2-9 / Pop 7+, rather than a parallel world
generator — the book blesses the substitution), the evolutionary `Environment` (terrain → Environ
DM, locomotion, ecological niche; p.227), the six-slot characteristic profile (`CharSpec{Name,
Dice}` — C1=Str/C4=Int fixed, C2/C3/C5/C6 rolled) and its `GeneticProfile` string (p.228;
`RollValue` applies the ≥4D "Rolling Higher" scaling `12+(n-2)D`), `Size` (weighted physical-dice
total × multiplier, Human=72) and a closed-form `Height` that reproduces the pp.236-237 grids, the
`LifeCycle` (per-stage durations → lifespan, Human=74; p.231), the `Gender` structure +
determination table (p.230), and — only when C6 is Caste — the `Caste` generation table (p.229,
resolving the `=Gender`/`=Special`/Unique substitutions). Gender and caste each carry a
`map[string]Difference` (the C1-dice/C1-C5-flat adjustments a member takes on at assignment).

Golden-locked to the **Human reference** (`SDEIES`, all-2D, Size 72, lifespan 74) and the p.218/219
**Ay Flux-0 fixtures** — the book prints no dice-traced worked species, so dense-table interiors are
locked cell-by-cell in `tables_test.go` instead.

The chargen bridge is `chargen.GenerateSophont` (`internal/chargen/sophont.go`): it rolls an
individual's six characteristics per the species' die counts, assigns a gender (and caste, if any)
by a 2D roll on the species tables, and applies their `Difference`s (no upper cap — an 8D Str
reaches 48, so `Character.UPP` now renders out-of-eHex-range values as `?` via `ehex.Format`).

Deferred until a consumer needs them: the physical/flavor tier (senses, body structure,
manipulators, special abilities, size-BFP body form, uniques, psionics), the caste/gender life-cycle
sub-mechanics (shift, assignment timing, caste-gender relation), the Skilled-caste skill lists
(Chart 12), sophont career service, and species-driven aging.

## Settled: `CharName` is not a naming inconsistency with `chargen.Characteristic`

`CharName` has 14 members, not 6 — 8 (Gra, Sta, Vig, Ins, Tra, Cha, Cas) are alien-analog
identities from chart 06A with no `chargen.Characteristic` counterpart at all. It was never "the
same six characteristics named on two conventions"; the two types overlap on six slots and
diverge past that. Do not rename `CharName`'s constants to full words:

- The abbreviations are the book's own chart-06A column identity — `gpLetter()` derives the
  Genetic Profile letters directly from them (`Cha`→C, `Cas`→K, etc.). The abbreviation **is**
  the data, not a display shortcut over a "real" long form.
- Renaming would mean inventing 8 canonical long-form names attested nowhere in the book or the
  code, for zero benefit.
- Both `Characteristic.String()` and `CharName.String()` already return identical hardcoded
  three-letter strings independent of the Go constant's identifier, so renaming changes no
  rendered output either way — and neither `String()` is on any production/golden path
  (`cmd/sophont` has no `testdata/` at all).

The real, narrower finding: `chargen.GenerateSophont` trusts `species.Chars[i]` to be slot
C(i+1), matching `chargen.Characteristic(i)`'s own C1..C6 order — documented at both
`chargen/sophont.go`'s `GenerateSophont` and this package's `CharSpec`, since nothing checks it
in code.
