# Where `t5` will not stretch

_A design review, written against the working tree at `b84d273` (2026-07-20).
Companion to `THEORY.md` (what the system means) and `walkthrough.md` (how it
reads). This document covers only what the current shape cannot absorb._

The repo is 183 commits old, all of them inside 60 days. Everything below is
cheaper now than after four more Makers. That is the whole argument for reading
it today rather than filing it.

Seven findings. Five are structural; two are refactors. Section 8 lists what
should not be touched.

---

## 1. The reproducibility guarantee rests on a mechanism that cannot deliver it

`THEORY.md` calls the dice stream "the deepest, least obvious, and most
load-bearing invariant in the system," and it is right that most of the
surprising code descends from it. It does not hold at the level users care
about, and the code already says so in three places:

- `internal/survey/survey.go:58` — "a hex's world depends on every hex rolled
  before it, but **that is an artifact of the roller**… the region-wide passes
  are the reason that would survive any redesign."
- `internal/systemgen/systemgen.go:154` — "Across multiple systems from one
  roller these rolls **do advance the shared stream**… each system stays
  reproducible for a given seed **and version**."
- `internal/systemgen/systemgen.go:205` — the `ggPresent`-over-rolled-0 case
  "**cannot align**." The same comment measures its own frequency at **28%**,
  "common, not a corner."

`README.md:23` promises "pass `-seed` and you get the same result every time."
What is actually true is: the same result every time, against one build. The
project's purpose is correcting the transcription, so every rule fix invalidates
every seed anyone has kept. #321 did exactly that and reached review undetected;
`CLAUDE.md:131` records it.

### What the discipline costs

- **255 `dice.NewScripted` call sites across 46 test files, ~690 lines of
  hand-enumerated die faces.** Every one is a tripwire on draw order.
- Three roll-and-discard sites in production code, existing only to hold
  alignment: `systemgen/systemgen.go:217`, `worldgen/worldgen.go:37`,
  `chargen/career.go:753`.
- Eleven documented _don't roll here_ sites with the same motive, among them
  `worldgen/otherworld.go:135`, `systemgen/orbitmap.go:326`,
  `chargen/career.go:973`, `chargen/career.go:1763`, `chargen/sophont.go:30`.
- A hand-rolled `scriptCounter` at `systemgen_test.go:259`, because
  "there is no way to count draws through `dice.NewScripted`" (`:247`).
- `chargen/career.go:1411` (`undercoverAssignment`) rerolls until 1-3, so its
  draw count is unbounded by construction and cannot be aligned at all.
- **Zero goroutines, zero `t.Parallel()`, no `sync` import in 196 files.**
  Sequential draw order _is_ the contract, so 1280-hex sector generation can
  never be parallel.

### The redesign

`internal/dice` has four constructors — `New`, `NewWithSeed`, `NewSource`,
`NewScripted` — and no `Fork`, `Derive`, `Child`, `Split`, or `Spawn`. There is
no reachable RNG state to re-key from; `Roller` holds an opaque `d6 func() int`
and an unexported `seed` with no setter.

Per-entity seed derivation — `NewWithSeed(hash(sectorSeed, col, row))` — buys
what the current design is trying to buy and failing to:

- a belt-rule fix changes belt worlds and nothing else;
- a hex is regenerable in isolation, which is what `survey.Subsector` wants;
- a sector parallelizes;
- the discard sites, the alignment comments, and most of the brittleness in
  those 690 lines go away.

The region-wide passes (`markSectorCapitals`, `placeNavalDepots`,
`placeWayStations`) remain the real reason a sector is atomic, exactly as
`survey.go:58` already argues. Those survive untouched.

---

## 2. Everything renders; nothing parses

`THEORY.md:28` treats the compact code-string as "the product rather than …
serialization of some richer inner model." Taken seriously, that makes the
product write-only.

The complete set of parsers in the repo:

| parser                                         | scope                              |
| ---------------------------------------------- | ---------------------------------- |
| `ehex/ehex.go:45` `ParseDigit`                 | one character                      |
| `worldgen/allegiance.go:72` `ParseAllegiance`  | a 2-char code                      |
| `sectorgen/sectorgen.go:59` `ParseHex`         | a 4-digit coord                    |
| `sectorgen/sectorgen.go:99` `ParseSubsector`   | one letter                         |
| `dice/notation.go:34` `Parse`                  | dice notation, not a domain object |
| `cmd/shipgen/main.go:267` `parseInstallations` | a CLI flag value                   |

No reader exists for `uwp.Profile`, `worldgen.World`, `systemgen.System`,
`shipgen.Ship`, or `survey.Record`. There are **zero struct tags and zero
`encoding/*` imports** repo-wide. Rendering is hand-built per package: 52
`Sprintf` calls in `shipgen`, 22 in `survey`, 18 in `systemgen`.

Worse than "the parser was never written": `uwp.Profile.String()` is **lossy by
design**. `formatStarport` (`uwp.go:75`) and `ehex.Format` both emit `?` for
out-of-domain values, so a rendered profile cannot round-trip even in principle.
A parser cannot be bolted on; the render path needs a non-lossy mode or a
separate wire format first. This is the strict/total split — correct on its own
terms — colliding with serialization.

`worldgen/world.go:154` reasons explicitly about "a reader, **or a parser**"
mis-reading a column-shifted record. The format was designed to be parsed.

### What this blocks

- **Persistence.** Seeds are the only save format, and §1 showed they expire on
  every rule fix.
- **Referee edit.** The one thing every Traveller tool must support.
- **T5SS `.sec` import** — deferred in #195.
- **#183** ("static item/ship catalogs… used both as content and as golden
  fixtures for the corresponding makers") is blocked behind it and is not
  scoped as such.
- **Round-trip tests** — the strongest available check on a positional record
  format, unavailable by construction.

---

## 3. The mainworld has no type

`THEORY.md:153` gets close — "the ruling names the wrong invariant" — and stops
one level short. The invariant is not misnamed; the field it should read does
not exist.

`worldgen.OtherWorldType` (`otherworld.go:11`) names nine types including
`Planetoids`. **Every secondary world is typed. The mainworld is not.**
`worldgen.World` (`world.go:16`) has 13 fields and no type. So a mainworld's
belt-ness has nowhere to live except `Size == 0`, and `uwp.Profile.IsBelt` is a
rename of the banned comparison because there is no other field to ask.

Sharper still: `worldgen` **defines** the type and then discards it.
`GenerateOtherWorld`, `GenerateHostWorld`, and `GenerateSatelliteWorld`
(`otherworld.go:74/90/112`) all return a bare `uwp.Profile`. `systemgen`
re-attaches the type alongside the profile (`systemgen/otherworld.go:11`,
`satellites.go:21`) and makes it load-bearing at `satellites.go:193`. The
concept is defined in the package that throws it away.

### The settled ruling is already dead

`CLAUDE.md` says: ask `Profile.IsBelt`, never compare Size to 0. Three of its
readers already ignore it:

- `systemgen/satellites.go:191` explicitly **refuses** `IsBelt` for secondary
  worlds in favor of `Type == worldgen.Planetoids`, with a comment explaining
  why. So `IsBelt()` returns true for bodies `systemgen` classifies `KindWorld`.
- `worldgen/tradeclass.go` never calls `IsBelt` **or** `BeltSize`. The `As` rule
  (`:32`) is a raw three-digit conjunction on Size/Atm/Hyd.
- `systemgen/orbitmap.go:240` reads Size as a dimension behind a comment
  conceding **"the enforcement belongs in the type (a Size that cannot be
  silently read as a dimension)."**

That last line is this section's recommendation, already written in the
codebase. The ruling should be retired and replaced by it, not re-defended.

### It reaches stdout

`typeSize` (`otherworld.go:144`) returns 0 for a `Worldlet` on a 1-3;
`chartProfile` then forces Atm 0 and Hyd 0. The profile matches the `As` rule
exactly, so a solid tiny world is stamped an asteroid belt — and keeps its
moons:

```
$ go run ./cmd/systemgen -n 40 -seed 5
7: Worldlet Y000233-A As Va Lo (3 moons: Ell Y575030-6; Ess HFFA669-9; Arr Y410041-6)
```

Nine such lines in 40 systems on that seed. (`THEORY.md:169` reports five.)

`WorldContext` (`context.go:16`) is the workaround — an out-of-band channel for
the facts the model cannot hold, at 7 fields and growing. #324 fixes the symptom
by adding an eighth.

---

## 4. Trade codes are bare strings, and the boundary has already broken

`.golangci.yml:11` disables `goconst` with the reason "domain vocabulary
(skills/careers/trade codes) is string map keys, not consts." That was a
defensible call for one package. It is no longer one package.

**46 distinct code literals, 122 occurrences, 6 packages, no named type
anywhere.** The only registry is `chartDOrder` (`tradeclass.go:69`), a sort-rank
list; unknown codes silently sort last (`rankOf`, `:110`), so it validates
nothing.

`THEORY.md:148` calls this "latent rather than live, because nothing calls
[`trade`]." It is live, in a pair of packages that do talk:

- `chargen/homeworld.go:35` keys homeworld skills on **`"Ds"`** — a code no
  `worldgen` path emits and that appears nowhere in `chartDOrder`. A dead map
  key that grants nothing, silently. Precisely the failure mode `THEORY.md`
  predicted for `trade`, already realized in `chargen`.
- `chargen` omits 14 codes `worldgen` can emit (`Ba Di Ph Px Pe Re Sa Lk Mr Cy
Fo Pz Ab An`). Each is a homeworld that grants no skill, with nothing
  distinguishing "no skill by rule" from "not in the map."
- Set sizes disagree across the repo: `trade`'s value set is 14
  (`trade.go:30`), its goods columns 21 (`goods_data.go`), `worldgen` produces 46.

So the fix is not blocked on wiring `trade` to a consumer. It is a small
mechanical change today and a six-package change after `thinggen`.

### The same pressure, generalized

"Package boundaries follow the rulebook's chapters, not software concerns" is
well-argued for `dice`/`task`. It stops working the moment two chapters need the
same primitive, and the whole remaining roadmap is that case:

- **#176:** "Build here first: the shared `geom` size/profile/density/volume/mass
  primitive, **reused verbatim** by ThingMaker, BeastMaker (#24), and Sophont
  creation (#17)."
- **#172 Money/Value:** deferred pending a consumer; costs are bare `int` Cr
  across `shipgen` and `trade`.
- **#181:** "~90 named Nd→row tables on **one trivial lookup engine**" — an
  admission that per-package hand-transcription does not scale.

Trade codes are the instance that has already failed. `geom` and Money/Value are
the next two.

---

## 5. Derived fields are publicly writable, and one blessed mutator is already stale

This is a correctness issue, not a style one.

`worldgen.World.Importance`, `Nobility`, `Economic`, and `Cultural` are all
derived at `generateWorld` (`world.go:58`) from `Profile`, `TradeCodes`, and the
bases — and all four are public. The package supplies recomputing mutators
(`SetCapital` `:84`, `SetNavalDepot` `:97`, `SetWayStation` `:107`), but:

- `w.WayStation = true` compiles and skips the `Importance` recompute
  (`world.go:113`) and the `Nobility` recompute (`world.go:114`).
- `SetWayStation`'s own doc (`world.go:105`) concedes it leaves `Economic` and
  `Cultural` un-cascaded. **The blessed path is already inconsistent.**

The same shape recurs:

- `systemgen.System` — 20 exported fields carrying four documented but
  unenforceable invariants, among them `len(Giants) == GasGiants`
  (`systemgen.go:51`) and "`Orbits` sorted by orbit" (`:44`).
- `systemgen.OtherWorld` / `Satellite` / `PlacedOrbit` — `Type` must agree with
  `Profile` must agree with `TradeCodes`; `satellites.go:193` treats `Type` as
  the sole authority on belt-ness, so a caller-set mismatch misclassifies the
  body.
- `shipgen.Ship` — a caller can clear `Problems` without touching anything that
  produced them, and `aboard()` gates four separate accountings on its length.
- `worldgen.Facilities` — `PortFacilities` enforces `Downport XOR Beltport`
  (`facilities.go:251`); both are public bools.
- `survey.Survey` — the at-most-one-`Cs`-per-sector / one-`Cp`-per-subsector
  constraint from `markSectorCapitals` (`survey.go:174`) lives nowhere on the
  type.

`uwp.go:47` concedes the general case: "Profile is an exported model whose
fields a caller may set directly, and a String method must not crash." The
type's answer to a broken invariant is to render `?`.

The counterexample matters. `chargen.Character.scores` (`chargen.go:53`) is
unexported, and both `skill.Set` maps (`skill.go:22`) are too. The encapsulated
pattern exists in this repo and works. The world/system/ship model simply does
not use it.

---

## 6. `chargen` has outgrown the shape it is written in

- `career.go` is **1990 lines** — 6.5× the next-largest file in the package,
  holding 45 functions.
- `runTerm` (`career.go:695`, 96 lines, `//nolint:cyclop`) dispatches eight
  career-variant bools as a flat `if`-ladder. The flags are exclusive **by
  convention only** — no validator, no constructor, no test. A `Career` literal
  setting two compiles, generates, and silently runs whichever the ladder
  reaches first. Ordering is load-bearing in at least one place: `FameCareer`
  returns before `selectCC` (`:700`), so the Entertainer never draws a
  Controlling Characteristic.
- `Policy` has **17 methods** (`policy.go:7`, `//nolint:interfacebloat`) and
  **no interactive implementation has ever been written** — no `bufio`, no
  `Stdin`, anywhere. `DefaultPolicy` plus one `NextCareer` override in
  `cmd/chargen/main.go:84`. Seventeen player-choice hooks serving a player that
  does not exist.
- `Character` has no `Species` field. Slot 5 is hardcoded `Social` in five
  places (`chargen.go:26/36/39/53/124`), while `sophont/characteristics.go:36`
  says C6 may be `Soc`, `Cha`, or `Cas`. `GenerateSophont` writes a Chaser score
  into a slot the type system calls Social; `applyDifference`
  (`sophont.go:67`) loops `for i := range 5` and silently skips it.
- `Character.Check`'s `numDice` parameter: **all 10 production call sites pass
  literal `2`** (`education.go:168/188/218/234/251`,
  `career.go:1085/1196/1199/1232/1235`). The `task.Difficulty` integration it
  exists for has no production consumer; the defaulting branch is exercised by
  one test.

#191, #192, and #193 all land here. A fourteenth career _shape_ means a ninth
bool.

---

## 7. `Problems []string` is stringly-typed, and tests already depend on the prose

24 producers across four types (`Ship`, `Weapon`, `Defense`, `Missile`), every
one a free-prose `Sprintf`. Production reads only `len()`, via
`aboard()` (`mounts.go:29`), which gates mount points, tonnage, cost, and TL
suppression.

But the wording is a cross-package test contract: `mounts_test.go:60/175/262`
and `design_test.go:131/192` match it with `strings.Contains`, and
`shipcombat/ship_test.go:117` constructs a `Problems` literal to fake a failed
design. Rewording a `Sprintf` is a silent test break in another package. A typed
problem code costs nothing here and is the difference between infeasibility
being data — as `generate.go:44` claims — and being a printed string.

---

## 8. Adjacent: ~17% of source has no path to a user

Packages with no non-test importer:

| package      | src LOC |
| ------------ | ------- |
| `trade`      | 1310    |
| `shipcombat` | 583     |
| `personals`  | 253     |
| `combat`     | 131     |
| `calendar`   | 104     |
| `senses`     | 76      |

Plus `rangeband` (222), whose only consumer is `senses`. **2679 of ~15,750
source lines**, with ~1700 lines of tests proving they work in isolation.

This is not dead code — it is the campaign tier with no campaign layer, and
`THEORY.md:140` is right that most of it is parked behind deferral issues. The
consequence that matters for §4: `trade` is the package whose interface most
needs the shared types, and it is the one with no consumer to force the
question. It will stay designed wrong until something calls it.

---

## What should not be touched

- **The CLI contract.** `cli.RejectUnusable`, the deferred `reportSeed`, the
  `flag.Visit`-not-value-comparison reasoning, the bool-set-false exemption.
  This is the best-defended boundary in the repo and it is genuinely principled.
- **The fail-open discipline in `clitest`.** The paired stdout leak checks stay.
  The `panic:` clause in `AssertRejected` stays. The derived-not-restated seed
  matcher stays.
- **The `dice`/`task` jurisdictional split.** Correct as argued; p.127 is a
  statement about tasks.
- **`route`** as a dice-free pure graph over its own minimal `World`. Exactly
  right, and the model for how a package should decline to import the world.
- **The comment register.** Comments here adjudicate rules rather than gloss
  code, and with the spec unavailable they are the only durable record of what
  the source text says. They are the repo's real asset. Nothing below asks for
  fewer of them.

---

## Ordering

Each section is tracked as a GitHub issue; all seven are on Project #6.

| §   | issue                                              |                                  |
| --- | -------------------------------------------------- | -------------------------------- |
| §1  | [#326](https://github.com/philoserf/t5/issues/326) | dice substreams                  |
| §2  | [#327](https://github.com/philoserf/t5/issues/327) | serialization / no parser        |
| §3  | [#328](https://github.com/philoserf/t5/issues/328) | the mainworld has no body type   |
| §4  | [#329](https://github.com/philoserf/t5/issues/329) | trade codes have no type         |
| §5  | [#330](https://github.com/philoserf/t5/issues/330) | publicly-writable derived fields |
| §6  | [#331](https://github.com/philoserf/t5/issues/331) | the `chargen` term engine        |
| §7  | [#332](https://github.com/philoserf/t5/issues/332) | typed `shipgen` problems         |

Suggested order:

1. **Dice substreams** (§1, #326). Foundational, worsens monotonically, and makes
   everything after it easier.
2. **Type the mainworld** (§3, #328). Small. Retire the `IsBelt` ruling rather
   than reinforce it — `orbitmap.go:240` already says how. Unblocks #324 properly
   and stops `WorldContext` growing.
3. **A trade-code type** (§4, #329). Moved ahead of the general primitive tier:
   it is not blocked on `trade` getting a consumer, because `chargen`↔`worldgen`
   is already broken and `"Ds"` is the proof.
4. **Encapsulate the derived fields** (§5, #330). Start with `worldgen.World`,
   since `SetWayStation` is already inconsistent. Follow `skill.Set`'s pattern.
5. **Serialization layer** (§2, #327). Needs a non-lossy render decision first,
   so it depends on 2 and 4. Unblocks #183 and #195, and is the only way a
   generated sector outlives a rule fix.
6. **The shared primitive tier** (§4, general). Do it when `geom`/#176 forces
   it.
7. `chargen` shape (§6, #331) and `Problems` typing (§7, #332) are refactors,
   not redesigns. Schedule against the deferral issues.

---

## Provenance

§1–§4 and §8 were verified directly against the code and by running the
generators. §5–§7 came from a subagent sweep and were re-verified afterwards; two
counts were wrong on first writing and are corrected here — `Policy` has **16**
methods, not 17, and the scripted-dice surface is **256** `NewScripted` call
sites across 46 test files (249 of them `dice.NewScripted` from outside the
`dice` package, in 45 files). One sweep also misreported `.golangci.yml` as
absent, which it is not.

The §3 figures are reproducible with `go run ./cmd/systemgen -n 40 -seed 5`:
**15** `As`-stamped Size-0 secondary worlds, **9** of them carrying moons, across
14 Worldlets and 1 Iceworld.

## A note on `.issues/` — cleared 2026-07-20

The git-ignored `.issues/` directory held 34 markdown files. **All 34 described
problems that no longer existed** — the directory was dated Jul 16 and waves
5–10 landed Jul 19–20, fixing every one. Two of its notes backed the losing side
of disagreements that had since been settled the other way
(`shipcombat-table-h-page-ref-disagreement.md` argued for p.86; the tree
normalized to p.95).

It was not a backlog. It was an unswept output buffer from a prior `code-audit`
run, and reading it as open work would have sent someone to re-fix nine
`chargen` bugs that were already fixed. It has been removed.

**The live recommendation is the second half:** `code-audit` writes there by
default and will recreate it on its next run. Point it somewhere obviously
transient, or sweep the directory as part of whatever closes the issues it
raises. A findings buffer that outlives its findings is worse than no buffer —
it reads exactly like a backlog, and this repo already has a documented habit of
filing issues for problems that do not exist.
