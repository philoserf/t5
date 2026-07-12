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
task check      # gofmt -l, go vet, go test — the pre-commit gate
task test       # go test ./...
task cover      # go test -cover ./...
task deps       # brew bundle — install tooling from the Brewfile
```

Or drive `go` directly (e.g. `go test ./internal/dice` for a single package).
Tooling (go, go-task, poppler's `pdftotext`) is pinned in `Brewfile`.

## Code

- `internal/dice/` — the T5 dice engine, faithful to Book 1 pp. 18-19 and the Dice
  Appendix pp. 253-260. A `Roller` is the single, seedable source of randomness (inject a
  scripted `d6` for deterministic tests, as the tests do). It provides the primitives (`Dice`,
  `Flux`/`GoodFlux`/`BadFlux`, `HalfDie`, even distributions), the roll-low `Check`/`Resolve`
  mechanic (Mod adjusts the Target, DM adjusts the roll), and a `Parse`/`Eval` for chart
  notation like `2D-2` and `Flux`. Build generators on top of this rather than calling
  `math/rand` directly. `dice.NewSource(func() int)` supplies a custom/scripted die source,
  which is how cross-package tests pin exact rolls.
- `internal/ehex/` — Traveller extended-hex digits (0-9, A-Z omitting I and O). `Digit`
  encodes, `ParseDigit` decodes. Every UWP characteristic is an eHex value.
- `internal/uwp/` — the `Profile` type and its `String` in StSAHPGL-T form (e.g. `A788899-C`).
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
  `World.SecondSurvey` renders the canonical world-record line (golden: Regina's
  `A788899-C Ph Pa Ri {+4}(D7E+4)[9C6D] BcCeF NS -`).
- `internal/systemgen/` — star system creation (Book 3 pp. 16-17, 28): the stars (spectral
  type/decimal/size via the p. 28 table, transcribed from a rendered image since the dense
  grid does not survive text extraction), gas giants, planetoid belts, world count, and the
  mainworld via `worldgen`. `classify` is the pure table lookup; `rollStar`/`Generate` roll and
  compose. Secondary stars (Close/Near/Far) are placed in orbit bands. The mainworld is a full
  `worldgen.World` (UWP + all derived data); `System.SecondSurvey(hex, name, allegiance)` renders
  the canonical one-line record and `String` shows it with PBG. Per-world orbital detail is deferred.
- `internal/chargen/` — character creation (Book 1, Characteristics pp. 47+, careers pp. 63-79,
  Master Chargen Checklist p. 72). Generates the six-characteristic UPP (Str/Dex/End/Int/Edu/Soc,
  each 2D, eHex) at age 18, offers `Check`, and `AgingCheck` (Book 1 p. 89: `2D < LifeStage`,
  physical from 34 / mental from 66, zero-cascade to illness/death). `GenerateCareered` runs the
  checklist lifecycle: UPP → homeworld skills (`ApplyHomeworldSkills`, one skill per Trade
  Classification, Book 1 p. 56 — the homeworld is a `worldgen.World` input) → optional education
  (`educate`, Book 1 pp. 59-60: remedial ED5 then the best-qualifying academic program — College
  or University, a shared `academicProgram` differing only in prereq/graduation-Edu — via
  apply/pass-fail/waiver, Major+1 per pass and Minor+1 per 2 passes, BA + Edu bump on graduation;
  golden-locked to the book's Eneri Dinsha College example `9AB58A`) → qualify → term loop →
  muster-out. Education is gated on `Policy.PursueEducation`, so a
  no-education policy leaves any dice trace (e.g. the golden Scout's) untouched. The term engine
  (`career.go`) is career-agnostic with pluggable seams (`CCMode` Rotate/Fixed — under FixedCC the
  policy picks one CC that serves the whole career, `AdvanceRule` RollLow/RollHigh, `Qualification`
  char-set, `ContinueRule` fixed/char/UseCC/TermsMod, `MusterBenefitDMT` for a +Terms Benefit-column DM, and
  the rank ladders `EnlistedRanks`/`OfficerRanks` + `Commission`/`EnlistedPromote`/`OfficerPromote`
  rules for armed-forces careers); a `Policy` (with `DefaultPolicy`) supplies every player choice so
  generation is deterministic and testable. The rank step (`resolveRank`) runs after Risk & Reward
  for a surviving armed-forces character: an enlisted soldier rolls Commission (success → officer
  track) else Enlisted Promotion, an officer rolls Officer Promotion; promotion targets are raised
  by `Character.Medals` (a Reward success) and `WoundBadges`, and each rank grants its automatic
  skill. Careers are data, each a file + hand-traced golden: `ScoutCareer` (`scout.go`, p. 79),
  `RogueCareer` (`rogue.go`, p. 84 — FixedCC), `SoldierCareer` (`soldier.go`, p. 82 — the first
  ranked career), `MarineCareer` (`marine.go`, p. 86), `SpacerCareer` (`spacer.go`, p. 81 — the
  naval career, whose Rating ladder uses the engine's EnlistedPromote), `AgentCareer`
  (`agent.go`, p. 83 — a rankless career whose Continue eases with terms served via
  `ContinueRule.TermsMod`), `CitizenCareer` (`citizen.go`, p. 78 — an `AutoBegin` career whose
  `CitizenLife` term (`runCitizenTerm`) replaces Risk & Reward with a benign roll that grants a
  Job/Hobby skill and never injures), `EntertainerCareer` (`entertainer.go`, p. 77 — a
  `FameCareer` whose `runFameTerm` shifts `Character.Fame` by a Flux roll, granting Talent +1 and
  two extra skills on a rise, and Continues vs Fame via `ContinueRule.UseFame`), `CraftsmanCareer` (`craftsman.go`, p. 75 — a `Masterpiece` career whose `runCraftsmanTerm`
  attempts a Masterpiece from Master Points [CC + Craftsman skill + `skill.Set.TopLevels`], raises
  the Craftsman skill each term, and Continues vs Craftsman×2 via `ContinueRule.UseSkill`), `ScholarCareer` (`scholar.go`, p. 76 — standard Risk & Reward where a Reward is a Publication
  [`RewardKind`], with a single rank ladder [`resolveRank` skips Commission when there is no
  officer track] and Publication-boosted promotion/continue [`PromotionRule.PubsMod`,
  `ContinueRule.PubsMod`]), and `FunctionaryCareer` (`functionary.go`, p. 87 — an `OfficePolitics`
  career whose `runPoliticsTerm` is two unmodified rolls: a failed Risk ends the career as a job
  loss [`MusteredOut` from the term, handled in `RunCareer`], a Reward success is a promotion). The Academic grid column uses `AwardMajor`
  / `AwardMinor` cells that raise the character's College Major/Minor (lost if uneducated, per the
  page footnote); `DefaultPolicy.ChooseSkillColumn` is character-aware, so a graduate specializes
  in the Academic column while an uneducated Scout falls through to Courier. Deferred: the rest of
  the education institutions (Trade School, higher/military), the other 2 careers (Merchant, Noble), the Scout's
  Courier/Explorer duty and R&R reward, the Rogue's Scheme mechanic (payoff/prison/infamy, its
  Scheme-driven eligibility, +Terms on R&R/Continue, the "12 is always failure" rule), the
  armed-forces Branch/Operations R&R mods and commission/promotion skill eligibility (Soldier and
  Marine), and Fame.
- `internal/calendar/` — the Imperial Calendar (Book 1 Appendix 02, p. 262): a 365-day `Date` (day 1 is
  Holiday, then 52 weeks Wonday..Senday), with `Weekday`, `Add` (year rollover), and `String`
  (`001-1105`). Pure date math, no dice.
- `internal/task/` — the Universal Task Format (Book 1 pp. 120-131). A `Difficulty` ladder (Easy
  1D … Beyond Impossible 8D, with `Hasty`/`Cautious` pace) over the dice engine's roll-low
  `Resolve`: `task.Resolve(r, difficulty, target, mods...)`, target being characteristic+skill.
  The play systems (senses, combat, personals) build on this.
- `internal/skill/` — a character's skills and knowledges (Book 1 pp. 132-171), a pure
  inventory (no dice). Cascade skills (Pilot/Gunner/Engineer/…) hold Knowledges; `GrantCascade`
  applies the Knowledge-Knowledge-Skill career progression; `TaskLevel` stacks parent+knowledge.
  Levels cap at 15, knowledges at 6. Used by chargen careers.
- `cmd/worldgen/`, `cmd/systemgen/`, `cmd/chargen/` — CLIs, each taking `-n` and `-seed`, e.g.
  `go run ./cmd/systemgen -n 10 -seed 42`.

When adding a generator, transcribe the rule tables/formulas from `docs/reference/` and lock
them with a golden test built from a worked example in the books.

## Current state

Early. Beyond the engine (dice + worldgen + systemgen + chargen) the repo contains only the source PDFs (`docs/pdf/`), a
README, and `.gitignore`.

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
