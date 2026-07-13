# Open-issue fixes — resolution plans (#62–69)

Resolution plans for the eight book-fidelity / correctness issues opened after the
worldgen and chargen audits. Grouped so overlapping golden re-traces are done once; each
group ships as its own reviewed PR. Convention: golden-lock against a book worked example,
run `task check`, one PR per group.

## Suggested order

1. worldgen #62, #63 — independent, small, quick wins.
2. chargen muster-out fidelity PR (#64 + #65 + #67) — one PR (shared muster goldens).
3. chargen #66, #68 — small per-term fixes.
4. chargen #69 — enhancement, lowest priority.

---

## worldgen (2 issues — independent, small, do first)

### #62 — Importance Extension wrongly counts Pre-Agricultural (Pa)

`internal/worldgen/extensions.go`: drop `"Pa"` from `importanceTradeCodes` (Chart E, B3 p.27,
lists only **Ag/Hi/In/Ri**). Re-trace the Importance tests — the isolated `bases=false` case
becomes **+3** (was asserting +4), `both bases` becomes **+4** (was +5); confirm the full Regina
`GenerateWorld` golden still yields `{+4}` with its rolled bases (correct Regina = Starport A +1,
TL≥10 +1, Ri +1, Naval&Scout +1 = +4). Effort **S**.

### #63 — `Economic.RU()` omits the 0→1 substitution

`Economic.RU()`: substitute 1 for any of R/L/I/E that is 0 before multiplying (B3 p.27, "If any
value = 0, use 1"); Efficiency may still be negative (RU may be negative). Add a table test
covering E=0 (e.g. 13·7·14·1 = 1274, not 0) and a negative-E case. Effort **S**.

---

## chargen — muster-out fidelity (one PR: #64 + #65 + #67)

These three share `MusterOut`/`applyBenefit` and re-trace overlapping muster goldens, so do them
together. Effort **M** (logic small; the golden re-traces across scout/agent/spacer/soldier/
marine/rogue/merchant/functionary/entertainer are the work).

### #64 — Knighthood muster benefit never raises Soc (B1 p.68)

Add a `Knighthood` `BenefitKind` (or special-case the named benefit) that sets Soc to
`max(Soc, 11)`, or `Soc+1` if already ≥11. Armed-forces (Spacer/Soldier/Marine) variant:
Knighthood is officer-only; a non-officer gets `Soc+1` instead — so the handler needs the
career's officer flag / record. Touches `applyBenefit` and the ~17 `named("Knighthood")` muster
rows. Audit and re-trace any golden whose reachable muster rows include Knighthood.

### #65 — extra muster-out rolls not granted (B1 p.67)

In `MusterOut`: `rolls += c.Commendations + qualifying medals + (c.Fame >= 19 ? 1 : 0)`. Confirm
from p.67 which medal tiers (MCG/SEH) count and whether plain Medals do. Agent (Commendations)
and Entertainer (Fame) goldens gain rolls.

### #67 — Scout/Agent "Fame +2" muster benefit is inert

Replace `named("Fame +2")` at `scout.go:97` and `agent.go:58` with
`Benefit{Kind: FameBump, Value: 2}`. Re-trace the Scout golden (Fame now 2 → the Scout's
`DMFameHalf` muster DM shifts) and the Agent golden.

---

## chargen — per-career term fixes (small, independent PRs)

### #66 — Noble "When Elevated 2" bonus skills not awarded (B1 p.85)

In `runIntrigueTerm`, when the Elevation succeeds (Soc raised, Land Grant awarded) award
`EligPerTerm + 2` skills that term instead of `EligPerTerm`. Re-trace the Noble golden. Effort **S**.

### #68 — Entertainer applies a Fame Flux in Term 1 (B1 p.77: "Term 1 = 2D", Flux from Term 2)

`runFameTerm` must skip the Flux on the first served term (the initial 2D stands) and apply it
from Term 2. `runFameTerm(r, p, c, career)` currently receives no term index, so thread it in —
pass `*careerRun` (RunCareer already sets `run.terms` before each `runTerm`) and skip the Flux
when `run.terms == 0`. Re-trace the Entertainer golden. Effort **S**.

### #69 — first-career Retry ignores each career's Retry characteristic (enhancement, B1 p.63/79)

Add a per-career `Retry` field naming the Retry characteristic (empty = career grants no Retry;
Scout = Education/C5). `beginCareer` uses it for the second Begin roll instead of the qualify
target, and only retries careers that declare one. Low priority — only fires on a failed first
Begin. Effort **S**.
