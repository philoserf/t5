# Implementation plans

Saved plans for the outstanding backlog: fixes for the open book-fidelity issues, and
build-out plans for the automations the [automation catalog](../automation-catalog.md)
marks 🟡 begun-but-incomplete. Each plan cites the book pages, names the code it touches,
and follows the house method — transcribe dense tables from a rendered page image,
golden-lock against a book worked example (Regina for worlds/systems), one reviewed PR per
slice.

## Open-issue fixes

[`open-issue-fixes.md`](open-issue-fixes.md) — resolution plans for issues **#62–69** from the
worldgen and chargen audits.

| Issue | Area     | Fix                                                      | Effort |
| ----- | -------- | -------------------------------------------------------- | ------ |
| #62   | worldgen | Importance drops the spurious `Pa` code                  | S      |
| #63   | worldgen | `Economic.RU()` applies the 0→1 substitution             | S      |
| #64   | chargen  | Knighthood raises Soc (muster-out fidelity PR)           | M ‡    |
| #65   | chargen  | extra muster rolls for Commendations / medals / Fame 19+ | M ‡    |
| #67   | chargen  | Scout/Agent "Fame +2" actually raises Fame               | M ‡    |
| #66   | chargen  | Noble "When Elevated 2" bonus skills                     | S      |
| #68   | chargen  | Entertainer starts Term 1 at 2D (Flux from Term 2)       | S      |
| #69   | chargen  | per-career Retry characteristic (enhancement)            | S      |

‡ #64/#65/#67 ship as one PR (shared muster-out goldens).

## Incomplete automations (catalog 🟡)

| Catalog | Plan                                                                               | Status                                                          |
| ------- | ---------------------------------------------------------------------------------- | --------------------------------------------------------------- |
| #1      | [`automation-01-trade-classifications.md`](automation-01-trade-classifications.md) | 21/39 codes done; +4 UWP-only now (S), 11 climate need #15      |
| #11     | [`automation-11-dice-faces-manydice.md`](automation-11-dice-faces-manydice.md)     | Good/Bad Flux done; die-faces + Many-Dice one PR (S)            |
| #15     | [`automation-15-systemgen-placement.md`](automation-15-systemgen-placement.md)     | stars/counts done; per-world orbit/climate/scheduler (L, 7 PRs) |

**Key cross-dependency:** systemgen #15 **Phase 2** (mainworld orbit + climate) produces the
`worldgen.ClimateCodes` that unblocks the 11 climate trade codes in #1. Die-faces (#11) unblocks
the Genetics gene (#29) and Uncertain/Spectacular tasks.
