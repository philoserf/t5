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
