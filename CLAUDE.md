# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

A multipurpose Traveller5 (T5) workspace holding three kinds of work in one place:

- **Go code** — generators (planned: `chargen`, `worldgen`, `systemgen`, and more) over T5 rules.
- **Rules reference** — the T5 core rulebooks extracted from PDF to markdown, then summarized and synthesized.
- **Worldbuilding** — setting notes and fiction that draw on the above.

Go module path: `github.com/philoserf/t5`.

This directory is a git repo sitting under the `philoserf` umbrella workspace
(`source/philoserf/`), which discovers and syncs it via `gh repo list philoserf`.

## Commands

Machine-level workflow runs through `task` (go-task, version 3; see `Taskfile.yml`):

```sh
task            # = task test
task check      # golangci-lint run (subsumes format + vet), go test — the pre-commit gate
task test       # go test ./...
task cover      # go test -cover ./...
task deps       # brew bundle — install tooling from the Brewfile
```

Or drive `go` directly (e.g. `go test ./internal/dice` for a single package).
Tooling (go, go-task, golangci-lint, poppler's `pdftotext`) is pinned in `Brewfile`.

## Code

- `internal/dice/` — the T5 dice engine, faithful to Book 1 pp. 18-19 and the Dice
  Appendix pp. 253-260. A `Roller` is the single, seedable source of randomness (inject a
  scripted `d6` for deterministic tests, as the tests do). It provides the primitives (`Dice`,
  `DiceFaces` for the individual dice, `Flux`/`GoodFlux`/`BadFlux`, `HalfDie`, even distributions),
  the roll-low `Check`/`Resolve` mechanic (Mod adjusts the Target, DM adjusts the roll; the result
  carries `Faces` and a `Spectacular()` classifier for three-1s/three-6s, Book 1 p.127 — but
  `Resolve` only **classifies**, it does not act: `Success` here is plain arithmetic, and the
  p.127 override is applied one layer up in `task`, which owns the chapter that states it), the
  Many-Dice fast methods for large pools (`ManyDice10`/`ManyDice2D`/`Average35`/`ManyDice35Flux`,
  Book 1 p.260), and a `Parse`/`Eval` for chart notation like `2D-2` and `Flux`. Build generators on top of this rather than calling
  `math/rand` directly. `dice.NewSource(func() int)` supplies a custom/scripted die source,
  which is how cross-package tests pin exact rolls. `NewScripted` is **exact, not cyclic**:
  it validates every face is a real 1..6 at construction and panics rather than wrapping when
  a test outruns its script. That is what makes a green suite evidence that the dice stream is
  unchanged — a script must enumerate **every** die the code under test draws, so a change in
  dice consumption fails loudly instead of being served recycled faces.
  `Roller.Derive(discriminators...)` mints an independent child stream from the parent's seed
  and the discriminators alone — **not** from how many rolls the parent has made (splitmix64
  fold; panics on an unseeded parent). It is the substream primitive from #326:
  `sectorgen.DeriveHex` keys a hex on `(seed, col, row)`, so a hex is regenerable in
  isolation and a rule fix touching one hex's draws cannot shift any other. The reach is
  currently the **sector→hex boundary only** — the entity generators (`worldgen`, `systemgen`,
  `chargen`) still run a single sub-entity on one stream, so their intra-entity roll-and-discard
  alignment sites (`worldgen.go`, `systemgen.go` `giantsFor`, `career.go` reward) remain
  load-bearing until each adopts `Derive` internally.
- `internal/ehex/` — Traveller extended-hex digits (0-9, A-Z omitting I and O). `Digit`
  encodes, `ParseDigit` decodes. Every UWP characteristic is an eHex value.
- `internal/uwp/` — the `Profile` type and its `String` in StSAHPGL-T form (e.g. `A788899-C`).
  `BeltSize` (0) is the one Size digit that is a **code, not a dimension** — a field of asteroids.
  Whether Size 0 _means_ a belt depends on the body, which is why the reader is **not** a `Profile`
  method: a **mainworld** with Size 0 is a belt (Book 3 p.16, "determined when World Size is
  generated"), but a **secondary** Size-0 world usually is not — a Worldlet is a tiny solid world that
  renders the same `St000...`, and only a `Planetoids` body is a belt. A Profile alone cannot tell
  them apart, so `uwp.Profile.IsBelt` was **removed** in #328 (it was `Size == BeltSize`: right for a
  mainworld, wrong for a Worldlet). Ask the body fact: `worldgen.World.Belt` for a mainworld (which
  _does_ derive from Size 0), `OtherWorldType.IsBelt()` for a secondary. `BeltSize` remains only for
  genuine dimension uses (the rendered digit; the Size-0 floor below which nothing caps a satellite).
- `internal/worldgen/` — mainworld UWP creation (Book 3 pp. 16-25). The characteristic
  formulas are **pure functions** taking their rolls as arguments (test them at their edges);
  `Generate` rolls in checklist order and composes them. Validated against the book's Regina
  worked example (golden test → `A788899-C`). Fold new generators into this shape.
  `TradeClassifications` returns the UWP-determinable trade codes (Book 3 p. 25); codes needing
  climate, orbit, mainworld status, or referee input are intentionally excluded. `Importance`,
  `RollEconomic`, and `RollCultural` compute the {Ix}(Ex)[Cx] Extensions (Book 3 Chart E,
  validated against Regina's `{+4}(D7E+4)[9C6D]`); Importance feeds the other two. `Nobility`,
  `RollBases`, `TravelZone`, and `NativeStatus` add the Chart F world data (p. 28).
  `GenerateWorld` composes all of it into a `World` (UWP + every derived attribute), and
  `World.SecondSurvey` renders the world-record line (golden: Regina's
  `A788899-C Ph Pa Ri {+4}(D7E+4)[9C6D] BcCeF NS -`). It is the **generatable subset** of the
  book's own Regina line, not a copy: the book also prints `An` (Ancient Site) and `Cp` (Subsector
  Capital), both referee-assigned, and it spaces the extensions (`{+4} (D7E+4) [9C6D]`). The record
  is **positional**, so every empty field is dashed rather than dropped — a world matching no trade
  code at all is real, and collapsing its TC column shifts every later field one place left.
- `internal/systemgen/` — star system creation (Book 3 pp. 16-17, 28): stars, gas giants,
  planetoid belts, world count, the orbit map, and satellites. See
  `internal/systemgen/CLAUDE.md`.
- `internal/chargen/` — character creation (Book 1, pp. 47+/63-79): UPP, education, the
  career-agnostic term engine, and all 13 careers as data. See `internal/chargen/CLAUDE.md`.
- `internal/sophont/` — the Sophont Creation System (Book 3 pp. 217-239), the core spine that
  makes chargen work for aliens; bridged by `chargen.GenerateSophont`. See
  `internal/sophont/CLAUDE.md`.
- `internal/calendar/` — the Imperial Calendar (Book 1 Appendix 02, p. 262): a 365-day `Date` (day 1 is
  Holiday, then 52 weeks Wonday..Senday), with `Weekday`, `Add` (year rollover), and `String`
  (`001-1105`). Pure date math, no dice.
- `internal/task/` — the Universal Task Format (Book 1 pp. 120-131). A `Difficulty` ladder (Easy
  1D … Beyond Impossible 8D, with `Hasty`/`Cautious` pace) over the dice engine's roll-low
  `Resolve`: `task.Resolve(r, difficulty, target, mods...)`, target being characteristic+skill.
  The play systems (senses, combat, personals) build on this. **`task` owns the difficulty
  vocabulary; `dice` speaks only dice counts** — a `Difficulty` is a ladder _index_ (Average = 1)
  and a `dice.Check{Dice: …}` is a _count_ (an Average check is 2D), so never pass one as the
  other: convert with `Difficulty.Dice()`. `dice` names exactly one count, `DefaultCheckDice`,
  the 2D `Resolve` falls back to. `Dice()` **panics** off-ladder rather than returning a count
  that would silently resolve as an ordinary check; `String()` stays total and renders `?`.
  `task` also owns the **p.127 Spectacular override** (`applySpectacular`, applied by both
  `Resolve` and `ResolveDice`): three 1s force `Success` true "even if the result would otherwise
  be a failure", three 6s force it false, `Effect` stays arithmetic, and Spectacularly Interesting
  (both at once, 6D+) leaves the outcome to the referee. It lives here, not in `dice`, because
  p.127 states the rule about _tasks_ ("Sometimes the task result is Spectacular") — `dice` keeps
  the dice observation (`Classify` over `Faces`) and `task` keeps the consequence, so a caller
  rolling a non-task `dice.Check` gets arithmetic with no opt-out flag to remember. This is why
  `chargen.Character.Check` routes through `task.ResolveDice`: a 3D Hard Check is a "very hard
  task" (p.47) and must stay Spectacular-eligible.
- `internal/skill/` — a character's skills and knowledges (Book 1 pp. 132-171), a pure
  inventory (no dice). Cascade skills (Pilot/Gunner/Engineer/…) hold Knowledges; `GrantCascade`
  applies the Knowledge-Knowledge-Skill career progression; `TaskLevel` stacks parent+knowledge.
  Levels cap at 15, knowledges at 6. Used by chargen careers.
- `internal/shipgen/` — Adventure Class Ship design (Book 2 pp. 30-95), deterministic rather
  than rolled; plus the weapon/defense/missile armament system. See
  `internal/shipgen/CLAUDE.md`.
- `internal/trade/` — the Trade & Commerce pricing engine (Book 2 pp. 209-221): speculative
  cargo, shipping, random trade goods, contracts. See `internal/trade/CLAUDE.md`.
- `internal/shipcombat/` — the Space Combat resolution engine (Book 2 pp. 193-204); bridges to
  `shipgen` so a generated ship can fight. See `internal/shipcombat/CLAUDE.md`.
- `internal/sectorgen/` + `internal/survey/` + `internal/route/` — interstellar mapping (Book 3
  pp. 12-15, 21, 27-28): the hex grid, sector survey, trade routes, and the deep per-system
  sheet. See `internal/survey/CLAUDE.md`.
- `internal/senses/`, `internal/personals/`, `internal/combat/` — the play tier (Book 1): sense
  Actions, social Personals, and personal combat, all roll-low via `task.ResolveDice`.
- `internal/rangeband/` — the world/space range ladder (Book 1 pp. 24-29), shared by the play tier.
- `cmd/worldgen/`, `cmd/systemgen/`, `cmd/chargen/`, `cmd/sectorgen/`, `cmd/shipgen/`, `cmd/sophont/` — CLIs
  (most take `-n` and `-seed`; sectorgen/shipgen take their own design flags), e.g.
  `go run ./cmd/shipgen -hull A -tl 12 -config L -structure shell -maneuver A -jump A
-weapon beamlaser:T1:orbit -defense blackglobe`.
  They follow one convention, owned by `internal/cli`: **generated records go to stdout, everything
  else to stderr**. Bad input is `cli.Fatalf` (exit 2, the code `flag` itself uses); a true-but-empty
  result is `cli.Notef` (exit 0, still off stdout, so a piped record stream stays clean).
  **Every run is reproducible**: `cli.Roller` draws the fresh seed itself when `-seed` is omitted and
  reports it via `Notef` ("`sectorgen: seed 16919235832026294750`"), so a run worth keeping can always
  be replayed — re-run with that `-seed` for byte-identical records, or to select another view of the
  same survey (`-hex`, `-sector`). The report is on stderr, so piped records are unaffected.
  **`README.md` pins actual generated records** (the `sectorgen -seed 42` sample), and nothing in
  the suite checks them — no test pins CLI output. So a change that shifts a seeded dice stream
  silently falsifies the README's own reproducibility promise. #321 did exactly that and it
  survived to review. When a generator's draw sequence changes, grep for pinned sample output and
  regenerate it, then diff the regenerated block against a fresh run.

  It is **deferred, not printed at construction**: `Roller`/`SeededRoller` hand back a `reportSeed`
  func alongside the roller, and each command calls it only once its own flags validate, so a run
  that dies on bad input never names a seed for records it did not generate. `Roller` merely
  parses — every per-command check (an unknown density, a bad hull letter) happens after it returns,
  which is why the ordering has to be the caller's. A command with a view flag validates it up front
  for the same reason (`sectorgen.selectView` resolves `-hex`/`-subsector` before the survey runs).
  A flag the chosen path **cannot honor** is bad input, not a no-op — `cli.RejectUnusable`, which
  every such path calls before reporting its seed. It is named the way round that stays correct:
  the caller lists the flags its path _reads_, so a flag added later is covered the day it is added
  rather than the day someone remembers to list it. `shipgen` without `-hull` rolls a random ship
  that reads none of the design flags, and `sectorgen`'s views are exclusive, so the losing view's
  flag is rejected too — `-hex 0436 -subsector Q` used to print a hex at exit 0 while that same
  `-subsector Q` is refused on the default path. "Was it set" is asked of `flag.Visit`, never of the
  value — `-tl 0` is a real Tech Level and `-armor 1` the hull's integral layer, so a
  default-comparison would wave the caller's own input through as unset. A bool explicitly set false
  is the exception: `-sector=false` asks for the path _not_ to be taken, so it is not a conflict.

- `internal/clitest/` — the end-to-end harness all six CLIs share. `main` calls `os.Exit`, so each
  case re-executes the test binary as a child that runs `main` instead of the tests; `Command.TestMain`
  intercepts before `m.Run` — which is what frees argv: nothing has parsed `os.Args` yet, so the
  child's command line rides it and the environment variable is only a marker. `Command.Run` returns
  the two streams and the exit code apart. `AssertRejected`
  / `AssertReportedSeed` assert the whole `internal/cli` contract in one call. The seed line is matched
  as a **whole line** (`^cmd: seed \d+$`), not as the substring `"seed "` — `flag`'s usage text
  describes the `-seed` flag and would otherwise read as a seed report on every run `flag` rejects.
  `reportSeed` is an obligation nothing in the compiler enforces, so every command has a test that
  fails if the call is dropped. **`cli` owns the seed-line format**: `cli.seedNote` is the `Notef`
  format `Roller` prints, and the matcher behind `cli.HasSeedReport`/`cli.ReportedSeed` is _built
  from_ that constant, so the seed wording cannot drift between writer and reader (the command
  prefix is spelled by hand and is not derived — see the note at `seedNote`). That matters
  because `AssertRejected` uses it for a **negative** assertion ("this run named no seed"), which
  fails open: a matcher drifted off the real wording would stop matching and wave a leaked seed
  through.

  The stdout leak check in `AssertReportedSeed` therefore runs **two** checks, and both must stay:
  `cli.HasSeedReport` for the canonical `Notef` line, plus a broad `"seed"` substring for a seed
  reaching the record stream by any other path (a header like `fmt.Println("# seed", n)`, a sheet
  field). They fail in opposite directions and neither subsumes the other — the anchored form
  cannot see a non-`Notef` leak, and the substring cannot survive a reword. Collapsing them to the
  anchored matcher alone looks like a tidy-up and is a coverage regression; that mistake shipped
  once and was caught in review. `internal/cli`'s own tests are an external `package cli_test` — `clitest` imports `cli`,
  so testing from inside would cycle — and they drive `clitest.Command` like any `cmd`, with a
  stand-in `rollgen` whose `Main` dispatches on an env var to the fatal/roll/quiet child.

When adding a generator, transcribe the rule tables/formulas from `docs/reference/` and lock
them with a golden test built from a worked example in the books.

### Settled rulings — do not re-open

These have each been re-litigated at least once. The full reasoning lives in the package's own
`CLAUDE.md`; read it there before touching any of them.

- **shipgen, the p.300 drive-cost ruling.** Modified drive stages cost **/2**, not x1. Finding
  p.48's x1 sample-ship notes is not new evidence — it is the losing side of a conflict the book
  has with itself, and #300 was mis-resolved that way twice. The cell is deliberately asserted in
  two catalogs (`TestDesignDriveStageCatalogP127`, `…P134`); do not collapse either assertion, and
  never add an `mcr == 0`-style sentinel that opts the **Modified `/2` cell** out of the check. (The
  `P134` table _does_ carry an `mcr == 0` sentinel — for rows whose printed cost derives from
  unrounded tonnage and so is not reproducible under the final-tonnage rule; that is legitimate, and
  the Modified cell's `P134`-side corroboration lives in the separate exact-cost loop, not the main
  table. The prohibition is against sentinelling away the `/2` reading, not against that mechanism.)
- **uwp, belt-ness is a body fact — Size 0 means belt only for a mainworld.** A mainworld with
  Size 0 IS a belt (Book 3 p.16); a **secondary** Size-0 world usually is not (a Worldlet is a tiny
  solid world rendering the same `St000...`, only `Planetoids` is a belt). A Profile cannot tell them
  apart, so `Profile.IsBelt` — `Size == 0`, right for a mainworld, wrong for a Worldlet — was
  **removed** in #328. Ask the body fact: `worldgen.World.Belt` for a mainworld (it derives from
  Size 0), `OtherWorldType.IsBelt()` for a secondary. Reading Size-as-belt for a _secondary_ shipped
  four times (#213, #200, #309, and #324's phantom As-with-moons). `BeltSize` (0) stays only for
  genuine dimension uses (the belt's rendered digit; the no-cap floor for satellites). The `As` trade
  code is a Chart-D UWP code (correct for a mainworld), but `TradeClassificationsWithContext` **strips**
  it from a secondary world that is not a `Planetoids` belt (`WorldContext.Belt`).
- **clitest, the two stdout leak checks.** `AssertReportedSeed` runs both an anchored seed-line
  matcher and a broad `"seed"` substring check. They fail in opposite directions and neither
  subsumes the other; collapsing them to the anchored form alone looks like a tidy-up and is a
  coverage regression that already shipped once.
- **README sample records are unpinned by any test.** A change that shifts a seeded dice stream
  silently falsifies the README's reproducibility promise (#321). When a generator's draw sequence
  changes, regenerate the pinned block and diff it against a fresh run.

## Current state

The world/system/character census is complete, and so is the starship tier: a sector can be
surveyed, its worlds and systems detailed, characters generated to crew a ship, the ship designed
and armed, and the ship flown into a fight. Sophont creation (#17) now has its **core spine**
(`internal/sophont` + the `chargen.GenerateSophont` bridge), so chargen works for aliens; its
physical/flavor tier is deferred. The Tier-5 content makers are now the largest open pieces. The
backlog lives in **GitHub issues** (the "Triage and Tracking" project) — one issue per unstarted
generator/primitive and one per deferred piece, each carrying its book page refs, scope, and
dependencies. (The former `docs/automation-catalog.md` planning doc has been retired.)

- **Source of truth** for rules is `docs/pdf/` (T5 Core Rules Books 1–3 + Read Me). These PDFs
  are **git-ignored and not distributed** (copyrighted Far Future Enterprises material) — each
  user supplies their own copies locally; see the README. Do not commit or otherwise redistribute
  them. Extracted markdown is derived output, not authoritative — regenerate rather than
  hand-patch when the extraction pipeline changes.
- **PDF-to-markdown extraction approach is undecided.** Confirm the chosen tool before doing
  or scripting an extraction; don't assume one.

## Working conventions

- **YAGNI.** The user actively prunes speculative scaffolding — do not create empty directories,
  placeholder packages, or `.gitkeep` trees ahead of real content. Add structure when something
  goes in it.
- Commit only when asked (see the user's global git guidance).
