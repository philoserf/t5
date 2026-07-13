# Plan — Catalog #11: Expose individual die faces + Many-Dice / Good-Bad Flux

Book 1 Dice Appendix (chapters 7-8, printed pp. 259-261; PDF page == printed page).
Package: `internal/dice/`. Effort: **S**. Tracks catalog item #11.

## Research findings — what already exists in `internal/dice/`

`dice.go`:

- `Die() int` — one D6 (1..6).
- `Dice(n int) int` — **sum only** of n D6 (0 for n<=0). Individual faces are discarded.
- `Flux() int` (D-D, -5..+5), `GoodFlux() int` (0..+5), `BadFlux() int` (-5..0) — **already
  correct and complete**. Verified against the p. 261 Flux appendix: Good = High-Low avg +2
  range 0..+5; Bad = Low-High avg -2 range -5..0. The "Good-Bad Flux" half of this catalog
  item's title is **already done** — no new Flux work is required.
- `HalfDie() int` (D/2 round up, 1..3).
- `New`/`NewWithSeed`/`NewSource`/`NewScripted` — construction. `NewScripted(faces...)` cycles
  a fixed face sequence: the deterministic-test seam used everywhere.

`check.go`:

- `Check{Dice,Target,Mod,DM}` and `CheckResult{Roll,Total,Target,Success,Effect}`.
- **`CheckResult` exposes only `Roll` (the summed 2D/nD), NOT the individual die faces.**
  `Resolve` calls `r.Dice(n)` and throws the faces away. This is the gap blocking Spectacular
  detection.

`even.go`: `EvenDist1to9`, `EvenDist0to9`, `EvenDist1to10`. `notation.go`: `Parse`/`Eval`/`Roll`.

**Not exposed (the actual work of #11):**

1. Individual die faces of a multi-die roll.
2. Spectacular classification (three 1s / three 6s / both).
3. The four Many-Dice fast methods (p. 260).

## Rule extracts (cited)

### Spectacular Results — B1 p. 127

- **Three Ones** (impossible on 1D/2D): roll includes **>= 3 ones** -> Spectacular Success
  (even if it would otherwise fail).
- **Three Sixes**: roll includes **>= 3 sixes** -> Spectacular Failure.
- **Spectacularly Interesting** (needs 6+ dice): includes both 6-6-6 and 1-1-1.
- Threshold is "includes 3" = **>= 3**, counted across **all** dice in the roll. Confirmed by
  the Astrogator example: Universe4 player rolls 4D (6,6,5,1) + referee uncertain 1D (6) =
  three sixes total across the 5D -> Spectacular Failure. So counting must span player +
  uncertain dice together.

### Uncertain Tasks — B1 p. 126

- `Uncertain (nD)`: referee rolls n of the difficulty's dice, player rolls the rest. For the
  players' deduction, each uncertain die is **assumed to be 3**. If player-dice-sum + 3*n
  yields success -> announced success; if it yields failure -> announced failure; otherwise
  the outcome stays hidden. Referee secretly resolves the true roll (all faces).

### Genetics gene — B1 p. 102 (relevant to separate catalog #29)

- The **first die** rolled for a genetic characteristic is the Gene; remaining dice are
  developed value. Human Str = 2D -> `faces[0]` is the gene. `DiceFaces(n)[0]` supplies this.

### Many Dice — B1 p. 260 (chapter 7)

- **Many Dice Defined:** used only when the count exceeds the largest Dice-Table entry —
  "**at least 11 Dice**".
- **MANY DICE 10 (reuse-10):** roll ten D6, then reuse them cyclically — die roll 1 serves
  rolls 11, 21, 31, ... Sum over n. Example: 100D, ten dice = 1 2 3 1 2 3 1 2 3 1 (sum 19),
  reused through 100D = **190**.
- **MANY DICE 2D (2D-subsample):** roll `k = 2D` (2..12) to pick a subsample size, roll kD,
  reuse those k faces cyclically over n. Example: 100D, 2D=5, roll 5D = 1 2 4 5 5 (=17),
  reuse through 100D = 17*20 = **340**. 75D, 2D=2, roll 2D = 1 3; series 1,3,1,3... over 75 =
  **149** (38 ones + 37 threes).
- **MANY DICE 3.5 (average):** result = **n * 3.5** (3.5 = expected single-die roll). Example:
  50D -> 50*3.5 = **175**. No dice rolled.
- **MANY DICE 3.5 FLUX:** roll one Flux, look up the Dice-3.5-Flux value, multiply by n.
  Table (p. 260): `value = (7 + Flux) / 2 = 3.5 + Flux*0.5`. Flux -5->1, -4->1.5, -3->2,
  -2->2.5, -1->3, 0->3.5, +1->4, +2->4.5, +3->5, +4->5.5, +5->6. Example: 100D -> Flux value
  - 100.
- Who chooses: player vs non-player -> player picks; player vs player -> the **recipient** of
  the result picks. (Policy note only; the engine exposes all four, caller selects.)

## (1) New API on `dice.Roller` and helpers

### a. Faces primitive — add to `dice.go`

```go
// DiceFaces rolls nD and returns the individual faces in roll order (nil for
// n <= 0). Dice(n) == sum of DiceFaces(n); DiceFaces additionally preserves
// each face, needed for Spectacular detection (B1 p.127) and the Genetics gene
// (the first face, B1 p.102).
func (r *Roller) DiceFaces(n int) []int
```

Keep `Dice(n)` as-is (hot path in worldgen/chargen — avoid per-roll allocation). `DiceFaces`
is a separate small loop; the duplication is one `for range n`.

### b. Spectacular classifier — new file `dice/spectacular.go`

```go
type Spectacular int
const (
    NotSpectacular Spectacular = iota
    SpectacularSuccess          // >=3 ones
    SpectacularFailure          // >=3 sixes
    SpectacularlyInteresting    // both (needs >=6 dice)
)
func (s Spectacular) String() string

// Classify inspects raw faces (B1 p.127). Three-or-more ones -> Success,
// three-or-more sixes -> Failure, both -> Interesting. Fewer than 3 dice can
// never be spectacular.
func Classify(faces []int) Spectacular
```

Implementation: count ones and sixes; `ones>=3 && sixes>=3 -> Interesting`; else
`sixes>=3 -> Failure`; else `ones>=3 -> Success`; else `NotSpectacular`.
(Order: Interesting first so both-present wins.)

### c. Expose faces + spectacular on the check path — edit `check.go`

- Add field `Faces []int` to `CheckResult` (doc: the raw dice; `Roll == sum(Faces)`).
- In `Resolve`, replace `roll := r.Dice(n)` with `faces := r.DiceFaces(n); roll := sumFaces(faces)`
  and set `Faces: faces`. Purely additive — existing readers of `Roll`/`Total`/`Effect`
  unchanged. Only `Resolve` pays a slice alloc; worldgen/chargen call `Dice` directly and are
  unaffected.
- Add method `func (c CheckResult) Spectacular() Spectacular { return Classify(c.Faces) }`.

### d. Many-Dice methods — new file `dice/manydice.go`

```go
const ManyDiceMin = 11 // "at least 11 Dice" (B1 p.260); methods are permissive
                       // but intended for n >= ManyDiceMin.

// ManyMethod selects a fast large-pool resolution (B1 p.260).
type ManyMethod int
const (
    Reuse10 ManyMethod = iota // Many Dice 10
    Sample2D                  // Many Dice 2D
    Average35                 // Many Dice 3.5 (no roll)
    FluxAverage35             // Many Dice 3.5 Flux
)

func (r *Roller) ManyDice10(n int) int         // reuse-10, exact int
func (r *Roller) ManyDice2D(n int) int         // 2D subsample, exact int
func Average35(n int) float64                   // n*3.5, no Roller needed
func (r *Roller) ManyDice35Flux(n int) float64  // ((7+Flux)/2)*n

// ManyDice dispatches; returns float64 so all four unify (Reuse10/Sample2D are
// integral). Caller rounds as the situation (damage) requires.
func (r *Roller) ManyDice(n int, m ManyMethod) float64
```

Design notes:

- `Average35` is a free function (no randomness) — do not fake it as a method; document it in
  the same file so the four live together.
- Rounding is left to the caller (e.g. damage). Return `float64` from the averaging methods;
  `int` from the two reuse methods since they are exact.
- Guard: `n<=0 -> 0`. Do **not** panic below 11; the "at least 11" is a usage guideline, so
  keep permissive and document. Caller (task/damage layer) decides when to switch to a fast
  method.

## (2) Precise formulas/procedures (all B1 p.260)

- `ManyDice10(n)`: `f := DiceFaces(10); sum += f[i%10] for i in 0..n-1`. Verify 100D golden:
  faces 1,2,3,1,2,3,1,2,3,1 -> 190.
- `ManyDice2D(n)`: `k := Dice(2)` (2..12); `f := DiceFaces(k)`; `sum += f[i%k] for i in 0..n-1`.
  Goldens: (k=5 faces 1,2,4,5,5) 100D -> 340; (k=2 faces 1,3) 75D -> 149.
- `Average35(n)`: `float64(n) * 3.5`. Golden: 50 -> 175.0.
- `ManyDice35Flux(n)`: `flux := Flux(); value := (7.0+float64(flux))/2.0; return value*float64(n)`.
  Goldens per table row, e.g. flux 0 -> 3.5*n; flux +5 -> 6*n; flux -5 -> 1*n; 100D at flux +2
  -> 4.5*100 = 450.

## (3) Integration with Check/Resolve and the `task` layer

- **Check/Resolve**: covered by (1c). `dice.CheckResult` gains `Faces` + `Spectacular()`.
  `task.Resolve` / `task.ResolveDice` delegate to `r.Resolve`, so both **automatically** carry
  faces and can report Spectacular with no signature change. Callers read
  `res.Spectacular()`.
- **Spectacular tasks (consumer)**: any task caller can now branch on
  `res.Spectacular()`. "Spectacularly Stupid" (C+S < numDice) and "Easy tasks can still fail"
  are already handled by the roll-low mechanic; only the 3-ones/3-sixes detection was missing.
  No new task API strictly required for the basic case — document the pattern. Optional thin
  helper `task.Spectacular(res) dice.Spectacular` for discoverability.
- **Uncertain tasks (consumer, B1 p.126)**: needs the _full_ face set (player + uncertain) for
  Spectacular per the Astrogator example, plus the assumed-3 deduction. Recommend a follow-on
  `task.ResolveUncertain(r, d Difficulty, uncertainDice, target int, mods...)` returning the
  true `CheckResult` (with all Faces, so `Spectacular()` sees referee dice) plus an
  announced-visibility flag computed from player-sum + 3*uncertainDice. This is enabled by
  DiceFaces but is a separate item — flag it, don't necessarily build it under #11.
- **Genetics #29 (consumer)**: `DiceFaces(2)[0]` is the Str gene (B1 p.102). #11 unblocks it;
  no genetics code in this item.
- **Many-Dice consumers**: no current caller (large damage pools like Suitcase Nuke 180D are
  future combat/damage work). Ship the four methods as engine primitives now; wire a
  `damage`/combat layer to them later. `Average35`/`ManyDice*` do not touch Check/Resolve.

## (4) Tests (deterministic via `NewScripted`)

New `dice/manydice_test.go`, `dice/spectacular_test.go`, and additions to `dice_test.go` /
`check_test.go`. Style matches existing `scripted(...)` helper and table tests.

- `DiceFaces`: `scripted(1,2,3).DiceFaces(3)` -> `[]int{1,2,3}`; `DiceFaces(0)` -> nil;
  `sum(DiceFaces(n)) == Dice(n)` invariant on a seeded roller.
- `Classify`: 3 ones -> Success; 3 sixes -> Failure; both (6D `1,1,1,6,6,6`) -> Interesting;
  2 ones -> NotSpectacular; `<3` dice never spectacular; 4+ sixes still Failure.
- `CheckResult.Spectacular` via `Resolve`: script a 3D check rolling `1,1,1` -> Faces set,
  Roll 3, Spectacular()==SpectacularSuccess even when Target would fail; `6,6,6` -> Failure.
- `ManyDice10`: golden `scripted(1,2,3,1,2,3,1,2,3,1).ManyDice10(100)` == 190.
- `ManyDice2D`: script the leading 2D then the subsample so k is fixed —
  `scripted(2,3, 1,2,4,5,5).ManyDice2D(100)` (2D=5 from 2+3, faces 1,2,4,5,5) == 340;
  `scripted(1,1, 1,3).ManyDice2D(75)` (2D=2, faces 1,3) == 149.
- `Average35`: `Average35(50)==175`; odd n -> `.5` (e.g. 51 -> 178.5).
- `ManyDice35Flux`: `scripted(6,1).ManyDice35Flux(100)` (Flux=+5 -> value 6) == 600;
  `scripted(1,6)` (Flux=-5 -> value 1) == 100; `scripted(4,4)` (Flux 0) 100D == 350;
  table-drive all 11 flux rows against `(7+flux)/2`.
- Range/invariant test folded into `TestPrimitiveRanges` style: over a seeded roller,
  `ManyDice10(n)`/`ManyDice2D(n)` land within `[n, 6n]`.

## (5) Effort and sequencing

**Effort: S** (all in `internal/dice`, one small additive field on `CheckResult`, no consumer
rewrites; Good/Bad Flux already done).

Sequence:

1. `DiceFaces` in `dice.go` + test (also proves the `sum==Dice` invariant).
2. `spectacular.go` (`Spectacular`, `Classify`, `String`) + test.
3. `check.go`: add `Faces` field, populate in `Resolve`, add `Spectacular()` method + test.
4. `manydice.go` (four methods + dispatcher + `ManyDiceMin`) + goldens.
5. Run `task check` (gofmt -l, go vet, go test).
6. Update the `internal/dice/` bullet in the repo `CLAUDE.md` to mention DiceFaces, the
   Spectacular classifier, and the Many-Dice methods.

Deferred (flagged, not in #11): `task.ResolveUncertain` (consumer of Faces, B1 p.126);
Genetics gene extraction (#29); a combat/damage layer that calls the Many-Dice methods.
