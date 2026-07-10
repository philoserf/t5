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
- `internal/systemgen/` — star system creation (Book 3 pp. 16-17, 28): the stars (spectral
  type/decimal/size via the p. 28 table, transcribed from a rendered image since the dense
  grid does not survive text extraction), gas giants, planetoid belts, world count, and the
  mainworld via `worldgen`. `classify` is the pure table lookup; `rollStar`/`Generate` roll and
  compose. Orbital placement and per-world detailing are deferred.
- `cmd/worldgen/`, `cmd/systemgen/` — CLIs, each taking `-n` and `-seed`:
  `go run ./cmd/worldgen -n 10 -seed 42` and `go run ./cmd/systemgen -n 10 -seed 42`.

When adding a generator, transcribe the rule tables/formulas from `docs/reference/` and lock
them with a golden test built from a worked example in the books.

## Current state

Early. Beyond the engine (dice + worldgen + systemgen) the repo contains only the source PDFs (`docs/pdf/`), a
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
