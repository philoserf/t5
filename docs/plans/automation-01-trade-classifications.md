# Plan: Finish Traveller5 Trade Classifications (catalog #1)

Chart D, Book 3 p.26 (predicates + comments); Trade/Commerce usage, Book 2 pp.204-207;
Habitable-Zone climate offsets, Book 3 p.24 (Ho=HZ-1, Co=HZ+1, Fr=HZ+2, Tz=Orbit 0-1).

Target: `internal/worldgen/tradeclass.go` (+ `tradeclass_test.go`).

## 0. Current state (what is already implemented)

`tcRules` in `tradeclass.go` covers 21 UWP-only codes, in table order:

- Planetary (9): `As De Fl Ga He Ic Oc Va Wa`
- Population (7): `Lo Ni Ph Hi Pa Ag Na`
- Economic (5): `Pi In Po Pr Ri`

The `tcRule` struct carries only `siz, atm, hyd, pop` eHex-digit sets; `TradeClassifications(p uwp.Profile)`
loops the table and emits every code whose sets all match via `allows()`. Government, Law, and
Starport are **not** yet consulted, and there is no per-world orbit/climate context.

Verified against Chart D image (p.26): all 21 existing predicates are transcribed correctly.

## 1. Every missing code -> predicate -> required input -> disposition

Predicate columns are eHex-digit sets from Chart D; `--` = unconstrained. "HZ offset" = orbit minus
the primary's Habitable-Zone orbit (negative = inner/hotter).

### Do-now (UWP-only; needs only fields already on `uwp.Profile`)

| Code | Name              | Predicate (Chart D p.26)                               | New input needed   | Disposition           |
| ---- | ----------------- | ------------------------------------------------------ | ------------------ | --------------------- |
| `Di` | Dieback           | Pop 0, Gov 0, Law 0 ("000-T")                          | Gov, Law, Starport | do-now                |
| `Ba` | Barren            | Pop 0, Gov 0, Law 0, **Starport E or X**               | Gov, Law, Starport | do-now                |
| `Px` | Prison/Exile Camp | Atm 23AB, Hyd 12345, Pop 3456, Law 6789 (comment "MW") | Law                | do-now                |
| `Re` | Reserve           | Pop 01234, Gov 6, Law 045                              | Gov, Law           | do-now (caveat below) |

Notes:

- `Di`/`Ba` share the same 000 core; Chart D adds the Starport E/X constraint only to `Ba`. Book text
  (p.26 header): "Ba requires Starport E or X." Recommended partition so exactly one fires:
  **Di = 000 + Starport A-D**, **Ba = 000 + Starport E/X**. This also keeps the existing
  asteroid-belt test green (its hand-built 000 profile has Starport `0x00`, which is in neither set).
- `Px` is the mainworld ("MW") twin of the non-MW `Pe` Penal Colony; only `Px` applies to a mainworld.
- `Re` Reserve sits in the "Secondary" band, **not** Political/Special, and has a full UWP predicate,
  so it is mechanically generatable. Caveat: a Reserve is often a referee designation; flag it and let
  the user prune if undesired. Recommendation: include it (deterministic predicate, cheap).

### Blocked on #15 (need per-world orbit / HZ offset, or non-mainworld status)

| Code | Name          | Predicate + comment (p.26 / p.24)                        | Blocking input                       |
| ---- | ------------- | -------------------------------------------------------- | ------------------------------------ |
| `Fr` | Frozen        | Siz 23456789, Hyd 123456789A; "HZ +2 or outer"           | HZ offset >= +2                      |
| `Ho` | Hot           | (no UWP); "HZ-1" (not a proper TC)                       | HZ offset == -1                      |
| `Co` | Cold          | (no UWP); "HZ+1" (not a proper TC)                       | HZ offset == +1                      |
| `Tr` | Tropic        | Siz 6789, Atm 456789, Hyd 34567; "HZ-1"                  | HZ offset == -1                      |
| `Tu` | Tundra        | Siz 6789, Atm 456789, Hyd 34567; "HZ+1"                  | HZ offset == +1                      |
| `Tz` | Twilight Zone | (no UWP); "Orbit 0-1" (not a proper TC)                  | absolute orbit in {0,1}              |
| `Fa` | Farming       | Atm 456789, Hyd 45678, Pop 23456; "HZ but not MW"        | HZ orbit AND not mainworld           |
| `Mi` | Mining        | Pop 23456; "Not MW. MW=In"                               | not mainworld AND system MW has `In` |
| `Pe` | Penal Colony  | Atm 23AB, Hyd 12345, Pop 3456, Gov 6, Law 6789; "Not MW" | not mainworld                        |
| `Sa` | Satellite     | (no UWP); "Far Satellite"                                | world is a far satellite             |
| `Lk` | Locked        | (no UWP); "Close Satellite" (not a proper TC)            | world is a close satellite           |

### Referee-assigned / sector context (NOT generated -- out of scope)

Chart D header: "Politicals and Specials are assigned by Referee (not generated)."

| Code                | Name                                                | Why excluded                                                   |
| ------------------- | --------------------------------------------------- | -------------------------------------------------------------- |
| `Cp`/`Cs`/`Cx`      | Subsector/Sector/Capital                            | Regional designation (Starport A) -- needs sectorgen, not #15  |
| `Cy`                | Colony                                              | Owner = most-important world within 6 hexes -- needs sectorgen |
| `Mr`                | Military Rule                                       | Assigned by regional Allegiance power (no predicate)           |
| `Fo`                | Forbidden                                           | Red Zone -- referee                                            |
| `Pz` `Da` `Ab` `An` | Puzzle / Dangerous / Data Repository / Ancient Site | Special -- referee                                             |

## 2. Concrete do-now code changes (`internal/worldgen/tradeclass.go`)

### 2a. Extend `tcRule` with Gov, Law, and Starport

```go
type tcRule struct {
    code                         string
    siz, atm, hyd, pop, gov, law string // eHex-digit sets; "" = unconstrained
    port                         string // allowed Starport letters; "" = any
}
```

Existing rows keep `gov`/`law`/`port` as `""` (zero value) -> behavior unchanged. Because the struct
is filled positionally, either (a) convert `tcRules` to keyed/field-named literals, or (b) append the
two/three new fields to every existing row. Named fields are safer against future column drift;
recommend switching the literal to named fields in the same edit.

### 2b. New rows (Chart D order)

Insert `Di`, `Ba` at the head of the Population band (before `Lo`); insert `Px` after `Na`; append
`Re` after `Ri` (its Secondary-band slot follows the skipped Climate band, so trailing emission
preserves relative table order):

```go
// Population (000 pair -- Di/Ba partitioned by starport so exactly one fires)
{code: "Di", pop: "0", gov: "0", law: "0", port: "ABCD"},
{code: "Ba", pop: "0", gov: "0", law: "0", port: "EX"},
...
// Economic
{code: "Px", atm: "23AB", hyd: "12345", pop: "3456", law: "6789"},
...
// Secondary
{code: "Re", pop: "01234", gov: "6", law: "045"},
```

### 2c. Thread Gov/Law/Starport through `TradeClassifications`

Extend the match to the two new eHex characteristics plus a starport-letter check:

```go
for _, r := range tcRules {
    if allows(r.siz, p.Size) && allows(r.atm, p.Atmosphere) &&
        allows(r.hyd, p.Hydrographics) && allows(r.pop, p.Population) &&
        allows(r.gov, p.Government) && allows(r.law, p.Law) &&
        portAllows(r.port, p.Starport) {
        out = append(out, r.code)
    }
}
```

Add a byte-set helper (Starport is a literal letter, not eHex):

```go
// portAllows reports whether starport letter s is in the allowed set (""=any).
func portAllows(set string, s byte) bool {
    if set == "" {
        return true
    }
    return strings.IndexByte(set, s) >= 0
}
```

No signature change to `TradeClassifications`; `GenerateWorld` / `World` / `SecondSurvey` are untouched.
Update the package doc block on `tcRules` (lines 20-28) to note that Di/Ba/Px/Re are now included and
that Starport E/X + Gov/Law are consulted, and to move Di/Ba out of the "intentionally excluded" list.

## 3. Climate interface for #15 (design now, implement with systemgen per-world placement)

systemgen already tracks star orbits (`CloseOrbit`, etc.) but does **not** yet compute each world's
orbit or the primary's HZ orbit. #15 (per-world orbital detail) must produce, per world, its absolute
orbit and HZ offset. Feed that back into a context-aware classifier that is a **superset** of the
UWP-only one (leave `TradeClassifications` as the pure-UWP entry point; the mainworld default keeps
working before #15 lands).

Proposed additions (do not build yet):

```go
// WorldContext carries the system-level facts Chart D's climate/secondary codes
// need beyond the UWP. #15 populates it from systemgen orbital placement.
type WorldContext struct {
    Orbit               int  // absolute orbit number (for Tz, orbit 0-1)
    HZOffset            int  // Orbit - HabitableZoneOrbit (neg = inner/hotter)
    HasHZOrbit          bool // world sits in a numbered orbit with a known HZ
    IsMainworld         bool
    Satellite           byte // 0 none, 'C' close (-> Lk), 'F' far (-> Sa)
    MainworldIndustrial bool // system mainworld carries In (for Mi)
}

// TradeClassificationsWithContext returns the UWP codes plus the climate and
// secondary codes decidable only with orbital/system context (Chart D p.26).
func TradeClassificationsWithContext(p uwp.Profile, ctx WorldContext) []string
```

Climate resolution from `HZOffset` (Book 3 p.24): `-1` -> `Ho` (+ `Tr` if Siz/Atm/Hyd match);
`+1` -> `Co` (+ `Tu` if match); `>= +2` -> `Fr` if Siz/Hyd match; `Orbit in {0,1}` -> `Tz`.
Secondary: `Fa` when in an HZ orbit and `!IsMainworld` and Atm/Hyd/Pop match; `Mi` when `!IsMainworld`
and `MainworldIndustrial` and Pop match; `Pe` when `!IsMainworld` and predicate matches; `Sa`/`Lk`
from `Satellite`. Once #15 lands, systemgen computes the mainworld's own orbit/HZ and calls this
variant so the mainworld also gains `Fr/Ho/Co/Tr/Tu/Tz` where applicable.

Keep `Ho`, `Co`, `Lk`, `Tz` emittable but documented as "not properly TCs" (Chart D header) -- a
follow-up may gate them behind a flag if the user wants strict TC-only output.

## 4. Tests (`internal/worldgen/tradeclass_test.go`)

Existing tests stay green (verified): Regina (`A788899-C`) matches none of Di/Ba/Px/Re; the
asteroid-belt 000 profile has Starport `0x00` so hits neither Di nor Ba; other cases don't match the
new predicates. Add golden worlds exercising each do-now code:

- **Di Dieback**: `Profile{Starport:'C', Size:5, Atmosphere:4, Hydrographics:3, Population:0, Government:0, Law:0}` -> includes `Di`, excludes `Ba`.
- **Ba Barren**: same but `Starport:'E'` (and an `'X'` variant) -> includes `Ba`, excludes `Di`.
- **Di/Ba mutual exclusion**: assert a 000 world never emits both, across Starports A-E and X.
- **Px Prison/Exile**: `Atm:0xA (10), Hyd:3, Pop:4, Law:7` with Starport set -> includes `Px`; a Law-5 twin excludes it.
- **Re Reserve**: `Pop:2, Government:6, Law:4` -> includes `Re`; Gov-5 and Law-6 twins exclude it.
- **Ordering**: assert emission order matches Chart D table order (Di/Ba before Lo; Px after Na; Re last).
- **Regression**: re-assert Regina and the asteroid belt unchanged.

If #15 is landed later, add a `TradeClassificationsWithContext` table test with worlds at HZ offsets
-1/0/+1/+2, orbit 0-1, and satellite/non-MW variants, each locked to a hand-traced Chart D expectation.

## 5. Effort and sequencing

1. **Do-now codes (Di, Ba, Px, Re)** -- Size **S**. Single-file change to `tradeclass.go` (struct +
   4 rows + `portAllows` + two extra `allows` calls + doc update) plus `tradeclass_test.go` goldens.
   No downstream API change. Ship independently of #15. Run `task check` (gofmt/vet/test) as the gate.
2. **Climate/secondary interface** -- Size **M**, and **blocked on #15**. Design lands now (section 3);
   implementation waits for systemgen to compute per-world orbit + HZ orbit. Deliver `WorldContext` +
   `TradeClassificationsWithContext` together with #15, then have systemgen call the context variant for
   the mainworld and every other world.
3. **Out of scope**: Political (`Cp/Cs/Cx/Cy`), `Mr`, and Special (`Fo/Pz/Da/Ab/An`) -- referee/sector
   assigned; revisit only if/when a sectorgen exists.
