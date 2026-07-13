# Catalog #15 — Systemgen Per-World Detailing & Placement (PLAN)

Package: `internal/systemgen/` (reusing `internal/worldgen`, `internal/ehex`, `internal/uwp`, `internal/dice`).
Scope: B3 pp.20-29 (Chart G + Chart B orbit/climate) and B1 pp.31-33 (orbital distances, HZ, drive limits, satellite orbits). Size **L**, delivered as **7 ordered PRs**.

---

## 1. What already exists (baseline audit)

`internal/systemgen/star.go` + `systemgen.go` + tests. Present today:

- **Star model & classification** (`star.go`): `Star{Type,Decimal,Size}`; the B3 p.28 Spectral Type & Size table (`spectralTable`, indexed by Flux+6), `classify` (pure, with the forbidden-combo → V fallbacks and BD/OB special cases), `rollStar` (primary rolls Flux; others derive from primary Flux). Verified against a rendered p.28 image and locked by `star_test.go`.
- **Stars in a system** (`systemgen.go`): Primary always; `PrimaryCompanion`, `Close`/`Near`/`Far` and each of their companions, present when a presence-`Flux() >= 3`. Matches B3 p.28 "Generate System Stars".
- **Secondary-star ORBIT BANDS only**: `CloseOrbit = 1D-1` (0-5), `NearOrbit = 5+1D` (6-11), `FarOrbit = 11+1D` (12-17) — the _star's_ orbit around the Primary (B3 p.28 "Place Stars In Orbits"). These are coarse band positions, **not** the per-object orbit grid.
- **Counts, not contents**: `GasGiants = 2D/2-2` (0-4), `Belts = 1D-3` (0-3), `Worlds = 1+GG+Belts+2D`. Gas giants and belts are integers with no size/type/orbit; `Worlds` is a bare count with no worlds generated.
- **Mainworld**: full `worldgen.World` via `worldgen.GenerateWorld(r, GG, Belts, false)` — UWP, TCs, {Ix}(Ex)[Cx], nobility, bases, zone, native status, PBG population digit.
- **Rendering**: `String`, `PBG`, `Stellar`, `SecondSurvey(hex,name,allegiance)`. Golden-locked to Regina in `systemgen_test.go`.

## 2. What is NOT built (the gap #15 must close)

- **Mainworld orbit & climate**: no Mainworld **Type** (Planet vs Satellite), no Satellite orbit name, no **HZ variance**, no **climate** → the climate trade codes are still intentionally excluded from `worldgen/tradeclass.go` (Fr/Ho/Co/Tr/Tu/Tz/Fa). **This is catalog #1's blocker.**
- **HZ orbit** per spectral type/size — not computed anywhere.
- **Concrete orbit numbers** for the mainworld and every object — only the three star bands exist.
- **Orbital geometry** — no AU/distance, no precluded (stellar-surface) orbits, no drive limits.
- **Gas-giant detailing** — count only; no size letter/diameter/type/skimming; no SGG→IG conversion.
- **Belt placement** — count only.
- **Other worlds** — none created, detailed, or placed (the `2D` in the world count produces no worlds).
- **Satellites** — none.
- **The placement scheduler** — rotate placement, precluded-orbit and collision resolution: entirely absent.

## 3. Exact tables/formulas extracted (source of truth)

Confirmed from rendered page images (text extraction mangles these grids). **The Chart G table on B3 p.29 is authoritative where it disagrees with the B3 p.20 / B1 p.31,33 duplicates** — its HZ values match the Regina worked example (F V → orbit 4; the p.20 A1 chart wrongly shows F V → 5).

### 3a. HZ Habitable Zone Orbits — B3 p.29 "HZ" (dup: B1 p.33 7a, B3 p.20)

Rows = spectral type, cols = size `Ia Ib II III IV V VI D`. `-` = not applicable.

```
O: 15 15 14 13 12 11  -  1
B: 13 13 12 11 10  9  -  0
A: 12 11  9  7  7  7  -  0
F: 11 10  9  6  6  4  3  0
G: 12 10  9  7  5  3  2  0
K: 12 10  9  8  5  2  1  0
M: 12 11 10  9  -  0  0  0
```

Orbit 0 or 1 in a star's HZ ⇒ **Tz Twilight Zone**. **Golden: F7 V → HZ orbit 4 (Regina primary).**

### 3b. Orbital Distances (AU) — B1 p.31 Chart 05 (dup B3 p.20 A1); Decimal detail B1 p.32

Integer-orbit AU by orbit #: `0→0.2, 1→0.4, 2→0.7, 3→1.0, 4→1.6, 5→2.8, 6→5.2, 7→10, 8→20, 9→40, 10→77, 11→154, 12→308, 13→615, 14→1230, 15→2500, 16→4900, 17→9800, 18→19500, 19→39500, 20→78700`. Million-km and light-time columns available. Decimal-orbit AU grid (Flux -5..+5) on B1 p.32 — **defer** (MOARN; not needed for placement/climate).

### 3c. Precluded (stellar-surface) orbits — B1 p.31 surface columns (dup B3 p.20 "Sub-Orbit")

A star of a given size/spectral has its surface at a listed orbit; "the first (innermost) available orbit is the next greater orbit number." Supergiants/giants preclude several inner orbits; V/VI/D stars preclude ~none (surface at orbit 0 or "no"). Transcribe as `surfaceOrbit(size,type)→int`, default 0.

### 3d. Drive limits — B1 p.31 (10D gravitic, 100D jump, 1000D maneuver, by type×size) — **defer** to a later, separate catalog item; not required by #15.

### 3e. Mainworld Orbit / Climate — B3 p.24 Chart B "2B MAINWORLD ORBIT" ← **unblocks #1**

`Flux → (HZ Var, Climate, climate TC)`:

```
Flux  HZVar  Climate        TC
-6    -2     (inferno-hot)  --
-5    -1     Hot/Tropic     Tr
-4    -1     Hot/Tropic     Tr
-3    -1     Hot/Tropic     Tr
-2     0     Temperate      --
-1     0     Temperate      --
 0     0     Temperate      --
+1     0     Temperate      --
+2     0     Temperate      --
+3    +1     Cold/Tundra    Tu
+4    +1     Cold/Tundra    Tu
+5    +1     Cold/Tundra    Tu
+6    +2     Frozen         Fr
```

Placement note: _"Place Mainworld in HZ Orbit ± HZ Var. DM +2 if Spectral M. DM -2 if Spectral O or B."_ Final mainworld orbit = `HZorbit(primary) + HZVar`. **Golden: Regina HZ Var Flux = 0 ⇒ Temperate, no climate TC** (matches book "Climate (based on HZ=0) = Temperate").
Additional climate codes by final-orbit position (B3 p.20 text / B1 p.33): HZ ⇒ Temperate; final orbit 0 or 1 ⇒ **Tz**. Emit the table's TC (Tr/Tu/Fr) plus Tz. (Book's Chart B emits Tr/Tu/Fr, not the broader T5SS Ho/Co pair — follow the book and document the deviation.)

### 3f. Mainworld Type & Satellite — B3 p.24 "2C SATELLITE?" (+ checklist B, p.23)

Checklist B: `MainWorld Type = Flux` (Planet or Satellite); if Satellite, `Satellite Orbit name = Flux`. 2C table maps `Flux → {Far Satellite | Close Satellite | Planet}`, whether parent is GG or Planet, and the orbit letter (Close: Ay..Em; Far: En..Zee). **Golden: Regina Type Flux = -4 → Far Satellite; Satellite Orbit Flux = -2 → "Arr".**

### 3g. GG Gas Giants detail — B3 p.29 "GG"

`2D → size letter / diameter(mi) / type`:

```
1 L 20,000  |  2 M 30,000 (Neptune) | 3 N 40,000 | 4 P 50,000 | 5 Q 60,000
6 R 70,000 (Saturn) | 7 S 80,000 | 8 T 90,000 (Jupiter) | 9 U 125,000
10 V 180,000 | 11 W 220,000 | 12 X 250,000 | 13 Y 250,000
```

Rows 1-6 = **SGG** (Small Gas Giant), 7-13 = **LGG** (Large). Skimming G value column (.2–3.0). Rules: "All BD Brown Dwarfs are Siz=Y"; **"Convert every second SGG to IG Ice Giant (same size)."** **Golden (Regina): GG1 2D=7→Siz S (LGG); GG2 2D=2→Siz M (SGG); GG3 2D=5→Siz Q (SGG)→converted to IG.**

### 3h. P2 Basic Placement Chart — B3 p.29

`2D → orbit for LGG / SGG / IG / Belt / World1 / World2`. LGG/SGG/IG/Belt are **HZ-relative offsets** (e.g. `-4..+7`, `HZ`=0); World1/World2 are **absolute orbit numbers**:

```
2D  LGG SGG IG  Belt World1 World2
1   -4  -3  HZ  -2   11     18
2   -3  -2  +1  -1   10     17
3   -2  -1  +2  HZ    8     16
4   -1  HZ  +3  +1    6     15
5   HZ  +1  +4  +2    4     14
6   +1  +2  +5  +3    2     13
7   +2  +3  +6  +4    0     12
8   +3  +4  +7  +5    1     11
9   +4  +5  +8  +6    3     10
10  +5  +6  +9  +7    5      9
11  +6  +7 +10  +8    7      8
12  +7  +8 +11  +9    9      7
```

Note: "GG and Belt placement is relative to HZ. World placement is by Orbit. If an orbit is duplicated or precluded, adjust to an adjacent / closest possible orbit." **Golden (Regina): SGG placement 2D=2 → HZ-2 = orbit 4-2 = orbit 2; IG placement 2D=2 → HZ+2.**

### 3i. P1 Placing Worlds priority + Other Worlds formulas — B3 p.29

Priority: Mainworld → Gas Giants (rotate per star) → Planetoid Belts (rotate per star) → Other Worlds (rotate per star; World1 column, last world via World2). Mainworld special cases: satellite ⇒ GG in MW orbit (BigWorld if no GG); asteroid-belt MW ⇒ Belt column, ignore HZ.
Other-worlds partial UWPs (`Max Pop = MW Pop-1`, spaceport not starport):

```
Hospitable  StSAHPGL-T (full)
Planetoids  St000PGL-T (Siz/Atm/Hyd=0)
Iceworld    StSAHPGL-T, Pop = DM-6
RadWorld    StSAH000-0, Siz = 2D
Inferno     YSB0000-0, Siz = 6+1D
BigWorld    StSAHPGL-T, Siz = 2D+7 (any Siz B+ is BW)
Worldlet    StSAHPGL-T, Siz = 1D-3
InnerWorld  StSAHPGL-T, Pop = DM-4, Hyd = DM-4
Stormworld  StSAHPGL-T, Siz = 2D, Atm = DM+4, Hyd = DM-4, Pop = DM-6
```

World-type by 1D and orbit region:

```
Inner (inside HZ-1) & HZ Worlds : 1 Inferno 2 InnerWorld 3 BigWorld 4 Stormworld 5 RadWorld 6 Hospitable*
Outer (beyond HZ+1)            : 1 Worldlet 2 Iceworld 3 BigWorld 4 Iceworld 5 RadWorld 6 Iceworld
Inner/HZ Satellites            : 1 Inferno 2 InnerWorld 3 BigWorld 4 Stormworld 5 RadWorld 6 Hospitable*
Outer Satellites               : 1 Worldlet 2 Iceworld 3 BigWorld 4 Stormworld 5 RadWorld 6 Iceworld
```

Regions: Inner = orbits ≤ HZ-2; Hospitable = HZ-1..HZ+1; Outer = ≥ HZ+2.

### 3j. Satellites — B3 p.29 "S" + orbit tables B3 p.24 2C, B1 p.33 7b1/7b2

Count per world: `GG = 1D-1`, `Inner = 1D-5`, `Hospitable = 1D-4`, `Outer = 1D-3`. Result 0 ⇒ Ring, reroll; <0 ⇒ none. Each satellite: `2D 7- Close / 8+ Far`, Flux for orbit letter (7b1 Ay..Em locked, 7b2 En..Zee not-locked, with radius multipliers).

---

## 4. Phased delivery

### Phase 1 — Orbital geometry primitives · effort **M**

**Scope**: pure lookups the rest of the plan builds on; no System changes yet.
**Transcribes**: HZ table (3a, B3 p.29), integer AU distances (3b, B1 p.31), precluded/surface orbits (3c, B1 p.31).
**Code** — new `internal/systemgen/orbit.go`:

- `HZOrbit(spectralType, size string) (int, bool)` — table lookup; `ok=false` for `-` cells (e.g. size VI on O/B). Handles `D` and `BD`(→Y-equivalent, HZ 0).
- `OrbitAU(orbit int) float64` and `orbitDistances [21]float64`.
- `surfaceOrbit(size, spectralType string) int` (default 0) and `precluded(star Star, orbit int) bool`.
  **Reuse**: consumes `Star` from `star.go`.
  **Tests** (`orbit_test.go`): every HZ cell at its edges; **golden `HZOrbit("F","V") == 4`**; AU spot checks (orbit 3 = 1.0, orbit 5 = 2.8); precluded example for a supergiant.
  **Ships**: geometry library, exercised only by tests.

### Phase 2 — Mainworld orbit + CLIMATE · effort **M** · ⭐ UNBLOCKS CATALOG #1

**Scope**: give the mainworld a concrete orbit and emit its climate trade codes. This is the phase that closes the #1 gap (the climate TCs excluded in `worldgen/tradeclass.go`).
**Transcribes**: Mainworld Type/Satellite (3f, B3 p.24 2C + p.23), Mainworld Orbit/Climate table (3e, B3 p.24 2B).
**Code**:

- `internal/systemgen/mainworld.go`:
  - `type MainworldPlacement struct { IsSatellite bool; SatelliteFar bool; SatelliteOrbitName string; HZVariance int; Orbit int; Climate string }`.
  - `rollMainworldType(r) (isSat, far bool, orbitName string)` (Flux → 2C).
  - `rollHZVariance(r, primary Star) int` (Flux + DM: +2 if M, -2 if O/B → table 2B HZ Var).
  - `mainworldOrbit(primary Star, hzVar int) int` = `HZOrbit(primary) + hzVar`, clamped to available/non-precluded.
- `internal/worldgen/tradeclass.go` (or new `climate.go`): `func ClimateCodes(finalOrbit, hzOrbit int) []string` — pure; returns `Tr`/`Tu`/`Fr` per the offset (map 2B rows to offset: offset -1 → Tr, +1 → Tu, ≤-2 → (hot/inferno, no TC per book), +2 → Fr) plus `Tz` when finalOrbit ∈ {0,1}. Update the `tcRules` header comment to note climate is now supplied here.
- `worldgen.World`: add `Climate string` field (optional, descriptive); systemgen appends `ClimateCodes(...)` to `Mainworld.TradeCodes`.
- `systemgen.Generate`: after the primary is rolled, compute `MainworldPlacement`, append climate codes to the mainworld's TCs. Roll ordering: perform the Type/HZVar rolls in systemgen (not `worldgen.Generate`) so worldgen's Regina golden dice stream is untouched; lock the new order with a scripted golden.
  **Reuse**: `worldgen.World.TradeCodes`, `uwp.Profile`.
  **Tests**: table-driven 2B (each Flux → HZVar/TC); **golden: Regina HZVar Flux=0 ⇒ Temperate, no climate TC added, mainworld orbit = 4**; Regina Type Flux=-4 ⇒ Far Satellite, orbit name "Arr"; a Tr case (Flux -4), a Tu case (Flux +4), a Fr case (Flux +6), a Tz case (orbit 0/1). Confirm `worldgen` golden tests still pass unchanged.
  **Ships**: mainworlds now carry orbit + climate; #1's climate trade codes are live.

### Phase 3 — Placed-object model + mainworld placement · effort **S**

**Scope**: the System data model for concrete orbits; place only the mainworld (and its GG/BigWorld parent if satellite). No multi-object scheduler yet.
**Transcribes**: P1 mainworld rules (3i, B3 p.29) — satellite ⇒ GG (or BigWorld Siz=-2D+7) in MW orbit; asteroid-belt MW ⇒ Belt column.
**Code** — `internal/systemgen/placement.go`:

- `type ObjectKind int` (Mainworld, GasGiant, IceGiant, Belt, OtherWorld, BigWorld, Satellite).
- `type Placed struct { Kind ObjectKind; Orbit int; World *worldgen.World; GG *GasGiant; ... }`.
- `type StarSystem struct { Host *Star; HZ int; Orbits map[int]*Placed }` (one per hosting star: Primary/Close/Near/Far).
- Extend `System` with `Systems []StarSystem` (Primary first).
- `placeMainworld(...)`: seat the mainworld at its Phase-2 orbit on the Primary; if satellite, create+seat the parent GG/BigWorld in that orbit and hang the mainworld off it.
  **Reuse**: `worldgen.GenerateWorld` output stays the mainworld payload.
  **Tests**: satellite Regina ⇒ Primary orbit 4 holds a GG, mainworld is its satellite; planet-mainworld ⇒ mainworld seated directly; asteroid-belt mainworld ⇒ Belt-column seat.
  **Ships**: system carries a real orbit map with the mainworld placed.

### Phase 4 — Gas giants & belts: detail + the ROTATE SCHEDULER (core) · effort **L** · ⭐ STATEFUL HARD PART

**Scope**: detail every gas giant, convert every 2nd SGG→IG, place all GGs and belts across stars via rotate placement with precluded/collision resolution. **This phase builds the scheduler that Phase 5 reuses.**
**Transcribes**: GG table (3g, B3 p.29), P2 Basic Placement Chart LGG/SGG/IG/Belt columns (3h, B3 p.29), P1 rotate-per-star (3i).
**Code**:

- `internal/systemgen/gasgiant.go`: `type GasGiant struct { Size string; DiameterMi int; Large bool; Ice bool; SkimG float64 }`; `rollGasGiant(r) GasGiant` (2D→row); SGG→IG conversion counter (every second SGG in generation order becomes an IceGiant of the same size).
- `internal/systemgen/scheduler.go` — **the placer**:
  ```
  type host struct { sys *StarSystem; hz, minOrbit, maxOrbit int; occupied map[int]bool }
  type placer struct { hosts []*host; cursor map[category]int } // round-robin per category
  func (p *placer) nextHost(cat) *host              // rotate: Primary,Close,Near,Far,wrap
  func (p *placer) place(h *host, target int) (int, bool) // collision/precluded resolve
  ```
  - **Host construction**: one `host` per present star. `minOrbit = surfaceOrbit+1` (precluded floor); `maxOrbit`: Primary=19; Close/Near/Far = `(starOrbitAroundPrimary - 3)` per B3 p.21 (so a Close star in orbit 2 hosts nothing; ≤ that is out of range). `hz = HZOrbit(host.Host)`.
  - **Rotate placement**: per category (GasGiant, then Belt), draw objects and assign to hosts round-robin: 1st→Primary, 2nd→Close, 3rd→Near, 4th→Far, then wrap. Skip hosts whose range is empty.
  - **Target orbit**: GG/Belt use P2 HZ-relative column ⇒ `target = host.hz + offset`; clamp/resolve.
  - **Collision/precluded resolution** (concretely): `place(h,target)`: if `target` is in-range, not precluded, and free ⇒ take it. Else spiral outward `target±1, target±2, …` (prefer outward, then inward, deterministic tie-break) to the nearest orbit that is in `[minOrbit,maxOrbit]`, non-precluded, and unoccupied; seat there and mark occupied. If the host has no free orbit at all ⇒ drop the object (P1 "ignore excess worlds").
  - **State**: `occupied` sets per host + per-category `cursor`; the shared `dice.Roller` is consumed in fixed order (category → rotation index → object) so a seed is reproducible.
  - Mainworld's seat (Phase 3) is pre-marked occupied before scheduling so GGs never collide with it; the satellite-parent GG counts as one of the GGs (Regina: GG1 already in orbit 4).
    **Reuse**: Phase 1 `HZOrbit`/`precluded`; Phase 3 `StarSystem.Orbits`.
    **Tests**: GG table golden (2D=7→S, 2D=2→M, 2D=5→Q); SGG→IG "every second" conversion; **Regina golden: GG1=S(LGG) in orbit 4 (the satellite parent), GG2=M(SGG) placed 2D=2→HZ-2→orbit 2, GG3=Q(SGG)→IG placed 2D=2→HZ+2 around the Far M-star**; a forced-collision case proving the spiral resolver picks the nearest free orbit; a precluded-orbit case; an overflow case that drops excess objects.
    **Ships**: fully detailed, concretely-orbited gas giants and belts; the reusable scheduler.

### Phase 5 — Other worlds: detail + placement (scheduler reuse) · effort **L**

**Scope**: create and place the remaining `Worlds` count as detailed worlds filling free orbits.
**Transcribes**: Other-worlds partial UWPs + world-type-by-1D tables (3i, B3 p.29), P2 World1/World2 columns (3h).
**Code** — `internal/systemgen/otherworld.go`:

- `type worldType int` (Inferno, InnerWorld, BigWorld, Stormworld, RadWorld, Hospitable, Worldlet, Iceworld, Planetoids).
- `rollWorldType(r, region)` per the 1D tables; `region` from the target orbit vs host HZ (Inner ≤HZ-2, Hospitable HZ-1..+1, Outer ≥HZ+2).
- `buildOtherWorld(r, worldType, mwPop) worldgen.World` — apply each type's partial-UWP formula, force spaceport (not starport), clamp `Pop = min(pop, mwPop-1)`.
- Extend the scheduler: after GGs/belts, place `remaining = Worlds - placed` other worlds, rotate per star, orbit from **World1 absolute column** (last world from **World2**); resolve collisions/precluded via the Phase-4 resolver; drop overflow.
  **Reuse**: `worldgen.Generate` for the full-UWP types (Hospitable/BigWorld/etc.); `worldgen.TradeClassifications` + Phase-2 `ClimateCodes` to tag each placed world's TCs.
  **Tests**: each world-type formula at its edges (e.g. Planetoids = St000PGL-T; Inferno = YSB0000-0, Siz 6+1D); region selection; **Regina golden: world 5 of 8 → World1 2D=5 → orbit 4 → Satellite of the orbit-4 GG** (per the p.22 worked example); World2 used for the last world; overflow dropped.
  **Ships**: complete inner/outer world roster with concrete orbits and UWPs.

### Phase 6 — Satellites · effort **M** (defer-able)

**Scope**: satellites for gas giants and worlds.
**Transcribes**: satellite counts (3j, B3 p.29 "S"), satellite orbit tables (B3 p.24 2C, B1 p.33 7b1/7b2).
**Code** — `internal/systemgen/satellite.go`: `satelliteCount(r, kind)` (GG 1D-1 / Inner 1D-5 / Hospitable 1D-4 / Outer 1D-3; 0⇒Ring reroll, <0⇒none); `rollSatellite(r, parent)` (2D 7-/8+ Close/Far, Flux→orbit letter, UWP via the same Inner/Outer satellite 1D tables + Other-Worlds formulas); attach to each `Placed`.
**Reuse**: Phase-5 `buildOtherWorld`, orbit-letter tables.
**Tests**: count formulas incl. Ring/none edges; Close/Far split; a golden satellite chain (Regina's orbit-4 GG hosting the mainworld as a Far Satellite "Arr").
**Ships**: satellites incl. the mainworld-as-satellite chain.

### Phase 7 — Rendering + CLI · effort **S**

**Scope**: surface the new detail.
**Code**: extend `System.String` to list stars, each hosting star's orbit map (orbit #, AU, occupant UWP/GG type), belts, satellites — FillForm-H-style; keep the existing one-line `SecondSurvey` stable (still Regina's canonical line) and add a verbose multi-line dump. Update `cmd/systemgen` (`-n`, `-seed`) to print the detailed system.
**Tests**: golden render of a small seeded system; assert `SecondSurvey` one-liner is byte-identical to today's Regina golden (regression guard).
**Ships**: user-visible detailed systems via the CLI.

---

## 5. Explicit call-outs

**(a) Which phase unblocks catalog #1 (climate trade codes):** **Phase 2.** It rolls the primary's HZ orbit (Phase 1), the mainworld's HZ Variance (B3 p.24 2B), derives the final orbit and climate, and adds a pure `worldgen.ClimateCodes(finalOrbit, hzOrbit)` that emits `Tr/Tu/Fr` (+ `Tz` for orbit 0/1) — exactly the climate codes `worldgen/tradeclass.go` currently excludes. After Phase 2 those codes flow into `Mainworld.TradeCodes`. (The book's Chart B emits Tr/Tu/Fr, not the T5SS Ho/Co pair; follow the book and document the deviation.)

**(b) The placement scheduler (stateful hard part):** built in **Phase 4**, reused in **Phase 5**. Algorithm:

1. **Hosts**: one per present star (Primary, Close, Near, Far). Each host precomputes `hz = HZOrbit(star)`, `minOrbit = surfaceOrbit(star)+1` (precluded floor), `maxOrbit` (Primary 19; secondary = its own orbit-around-primary − 3, per B3 p.21 — a star too close hosts nothing), and an empty `occupied` orbit set. The mainworld's seat is pre-marked occupied.
2. **Rotate per category** in priority order (Gas Giants → Belts → Other Worlds, per B3 p.29 P1): assign object _i_ to host `i mod len(activeHosts)` (Primary, Close, Near, Far, wrap), skipping hosts with no available range.
3. **Target orbit**: GG/Belt read the P2 **HZ-relative** column ⇒ `target = host.hz + offset`; Other Worlds read the P2 **absolute** World1 column (last world = World2).
4. **Collision / precluded resolution** (B3 p.29 "adjust to an adjacent or closest possible orbit"): if `target` is in `[minOrbit,maxOrbit]`, non-precluded, and free ⇒ seat it. Otherwise spiral outward — `target+1, target-1, target+2, target-2, …` — to the nearest orbit satisfying all three constraints; seat and mark occupied. If the host is full ⇒ **drop** the object (P1 "if worlds exceed available orbits, ignore excess").
5. **Determinism**: state is the per-host `occupied` sets plus per-category round-robin cursors; the single `dice.Roller` is consumed in a fixed (category → rotation index → object) order, so any seed reproduces exactly. Concretely a `placer{ hosts []*host; cursor map[category]int }` with `nextHost(cat)` and `place(host,target)(orbit,ok)`.

## 6. Suggested PR sequence & sizing

1. Phase 1 (M) — geometry primitives. 2. **Phase 2 (M) — orbit+climate, unblocks #1.** 3. Phase 3 (S) — placed-object model. 4. **Phase 4 (L) — GG/belt detail + scheduler.** 5. Phase 5 (L) — other worlds. 6. Phase 6 (M) — satellites (defer-able). 7. Phase 7 (S) — render + CLI.
   Every phase is independently shippable, golden-locked against the Regina worked example (B3 pp.22-23) where the book gives worked numbers, and leaves `SecondSurvey`'s existing one-liner byte-stable.
