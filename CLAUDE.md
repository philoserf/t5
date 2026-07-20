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
- `internal/ehex/` — Traveller extended-hex digits (0-9, A-Z omitting I and O). `Digit`
  encodes, `ParseDigit` decodes. Every UWP characteristic is an eHex value.
- `internal/uwp/` — the `Profile` type and its `String` in StSAHPGL-T form (e.g. `A788899-C`).
  `Profile.IsBelt` names the one Size digit that is a **code, not a dimension**: `BeltSize` (0) means
  a field of asteroids, so every rule reading Size as a measurement must resolve it first. Reading it
  as a dimension has shipped three times (#213, #200, #309); ask `IsBelt` rather than comparing to 0.
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
- `internal/systemgen/` — star system creation (Book 3 pp. 16-17, 28): the stars (spectral
  type/decimal/size via the p. 28 table, transcribed from a rendered image since the dense
  grid does not survive text extraction), gas giants, planetoid belts, world count, and the
  mainworld via `worldgen`. `classify` is the pure table lookup; `rollStar`/`Generate` roll and
  compose. Secondary stars (Close/Near/Far) are placed in orbit bands. The mainworld is a full
  `worldgen.World` (UWP + all derived data); `System.SecondSurvey(hex, name, allegiance)` renders
  the canonical one-line record and `String` shows it with PBG. `placeOrbits` lays the full orbit
  map (mainworld/gas giants/belts/other worlds in concrete orbits, rotate-per-star); `rollSatellites`
  gives every placed body its moons — each a real satellite with a type (`satelliteType`, the p.29
  Satellites tables) and UWP, capped to its parent's size with a double-planet flag at equal size
  (Book 3 p.21), or a Ring. `rollMoon` is the **single** moon-assembly path (both the satellite pass
  and the orbit map's gas-giant-captured world go through it, so neither their dice order nor the
  tables they read can drift — the Outer Worlds and Outer Satellites tables disagree at 1D=4, and a
  captured world is created as a satellite, so it is typed as one). `satelliteParent` names the body
  an orbit's moons belong to, which is **not** always the orbit's Kind: when the mainworld is itself
  a satellite, p.21 puts a gas giant (or a `GenerateHostWorld` BigWorld, floored at the mainworld's
  own Size) in its orbit, so the orbit's moons are counted and capped for that parent and render as
  the mainworld's _sibling_ moons. Any parent that has a UWP is classified by `satelliteBody`, the
  **one** read that answers both halves at once — the count rule it takes and the cap it imposes —
  because when those were separate decisions only the cap resolved the asteroid-belt code, so a belt
  mainworld was capped as a belt but _counted_ as a world and rolled phantom moons (#309).
  Orbit letters are orbit names, so `satelliteOrbits` keeps them
  unique per parent, nudging a duplicate to the nearest free letter without touching the Flux roll
  (p.29, "adjust to an adjacent or the closest possible orbit"). The size cap is applied
  **inside** generation via `worldgen.GenerateSatelliteWorld` — Atmosphere is
  Flux+Size and Hydrographics is Flux+Atmosphere, so capping Size after the roll would leave a
  profile describing the larger world and break the World Creation chart's own structural rules
  ("If Siz=0, Atm=0", "If Siz <2, Hyd =0", p.24). Capping in place consumes identical dice, so it
  re-derives rather than re-rolls.
- `internal/chargen/` — character creation (Book 1, Characteristics pp. 47+, careers pp. 63-79,
  Master Chargen Checklist p. 72). Generates the six-characteristic UPP (Str/Dex/End/Int/Edu/Soc,
  each 2D, eHex) at age 18, offers `Check`, and `AgingCheck` (Book 1 p. 89: `2D < LifeStage`,
  physical from 34 / mental from 66, zero-cascade to illness/death). `GenerateCareered` runs the
  checklist lifecycle, then serves additional careers while `Policy.NextCareer` supplies them
  (multi-career via `serveCareer`; the CLI's `-career a,b,c` sequence): UPP → homeworld skills (`ApplyHomeworldSkills`, one skill per Trade
  Classification, Book 1 p. 56 — the homeworld is a `worldgen.World` input) → optional education
  (`educate`, Book 1 pp. 59-60: remedial ED5, then either a vocational Trade School
  (`attendTradeSchool`, one year → a `theTrades` Major +2, no Minor/Edu-bump/degree, chosen via
  `Policy.ChooseTradeSchool`) or the best-qualifying academic program — College or University
  (undergraduate) then, for a BA-holder electing it (`Policy.PursueGraduateSchool`), the post-graduate
  Masters → Professors ladder — all one parameterized `academicProgram` (years, `awardsMajor`,
  `awardsMinor`, Edu-or-degree prereq, grad-Edu, degree) through the shared apply/pass-fail/waiver
  `attendAcademic`: undergraduates get Major+1 per pass and Minor+1 per 2 passes with a BA + Edu bump,
  the Masters raises only the Minor (MA, Edu 9), Professors neither (Edu 12); golden-locked to the
  book's Eneri Dinsha College example `9AB58A`) → Begin → term loop →
  muster-out. Begin (`beginCareer`, Book 1 p. 63) rolls to qualify; the first career Retries a failed
  Begin once, later careers do not, and a character refused by every chosen career falls back to the
  auto-begin Citizen life (T5 has no draft — no one ends up careerless). Education is gated on `Policy.PursueEducation`, so a
  no-education policy leaves any dice trace (e.g. the golden Scout's) untouched. The term engine
  (`career.go`) is career-agnostic with pluggable seams (`CCMode` Rotate/Fixed — under FixedCC the
  policy picks one CC that serves the whole career, `AdvanceRule` RollLow/RollHigh, `Qualification`
  char-set, `ContinueRule` fixed/char/UseCC/TermsMod, `BenefitDM` selecting the muster Benefit-column
  die modifier (`MusterDM`: Terms/OfficerRank/Rank/FameHalf), and
  the rank ladders `EnlistedRanks`/`OfficerRanks` + `Commission`/`EnlistedPromote`/`OfficerPromote`
  rules for armed-forces careers); a `Policy` (with `DefaultPolicy`) supplies every player choice so
  generation is deterministic and testable. The rank step (`resolveRank`) runs after Risk & Reward
  for a surviving armed-forces character: an enlisted soldier rolls Commission (success → officer
  track) else Enlisted Promotion, an officer rolls Officer Promotion; promotion targets are raised
  by the summed p.70 table mods of the character's `Medals` — earned on a held Risk (an XS) as
  well as a passed Reward — and **not** by `WoundBadges`, a book conflict resolved against the
  Eneri Dinsha worked example (p.72) and documented at `promoted`. Each rank grants its automatic
  skill. Careers are data, each a file + hand-traced golden: `ScoutCareer` (`scout.go`, p. 79),
  `RogueCareer` (`rogue.go`, p. 84 — FixedCC), `SoldierCareer` (`soldier.go`, p. 82 — the first
  ranked career), `MarineCareer` (`marine.go`, p. 86), `SpacerCareer` (`spacer.go`, p. 81 — the
  naval career, whose Rating ladder uses the engine's EnlistedPromote), `AgentCareer`
  (`agent.go`, p. 83 — a rankless `UndercoverCareer` whose `awardUndercover` borrows one skill from
  a rolled career's grid (`undercoverAssignment` + the `CareerByID` registry) each term, adds the
  Successful-Mission skills on a held Risk, and earns a Commendation on a Reward [`RewardCommendation`,
  `DMCommends`]; Continue eases with terms served via `ContinueRule.TermsMod`), `CitizenCareer` (`citizen.go`, p. 78 — an `AutoBegin` career whose
  `CitizenLife` term (`runCitizenTerm`) replaces Risk & Reward with a benign roll that grants a
  Job/Hobby skill and never injures), `EntertainerCareer` (`entertainer.go`, p. 77 — a
  `FameCareer` whose `runFameTerm` shifts `Character.Fame` by a Flux roll, granting Talent +1 and
  two extra skills on a rise, and Continues vs Fame via `ContinueRule.UseFame`), `CraftsmanCareer` (`craftsman.go`, p. 75 — a `Masterpiece` career whose `runCraftsmanTerm`
  attempts a Masterpiece from Master Points [CC + Craftsman skill + `skill.Set.TopLevels`], raises
  the Craftsman skill each term, and Continues vs Craftsman×2 via `ContinueRule.UseSkill`), `ScholarCareer` (`scholar.go`, p. 76 — standard Risk & Reward where a Reward is a Publication
  [`RewardKind`], with a single rank ladder [`resolveRank` skips Commission when there is no
  officer track] and Publication-boosted promotion/continue [`PromotionRule.PubsMod`,
  `ContinueRule.PubsMod`]), `FunctionaryCareer` (`functionary.go`, p. 87 — an `OfficePolitics`
  career whose `runPoliticsTerm` is two unmodified rolls: a failed Risk ends the career as a job
  loss [`MusteredOut` from the term, handled in `RunCareer`], a Reward success is a promotion), `NobleCareer` (`noble.go`, p. 85 — a `ReturnIntrigue` career whose `runIntrigueTerm` risks Exile
  and offers Elevation [a roll-high check vs Soc that raises Soc and awards a Land Grant]; the
  Noble's rank is their Social Standing via `NobleTitle`), and `MerchantCareer` (`merchant.go`,
  p. 80 — standard Risk & Reward where a Reward is escalating Ship Shares [`RewardShipShares`, the
  Nth reward = N shares], with a dual Rating/Officer rank track). All 13 careers are now
  implemented. The Academic grid column uses `AwardMajor`
  / `AwardMinor` cells that raise the character's College Major/Minor (lost if uneducated, per the
  page footnote); `DefaultPolicy.ChooseSkillColumn` is character-aware, so a graduate specializes
  in the Academic column while an uneducated Scout falls through to Courier. Deferred: the rest of
  the education institutions (Trade School, higher/military), the Scout's Courier/Explorer duty and
  R&R reward, the Rogue's Scheme mechanic, the armed-forces Branch/Operations R&R mods and
  commission/promotion skill eligibility, and each career's documented flavor deferrals (in its own
  file header). See the per-career `.go` files for the exact deferred pieces.
- `internal/sophont/` — the Sophont Creation System (Book 3 pp. 217-239), the **core spine** that
  makes chargen work for aliens. `Generate` composes Flux-driven sub-generators into a `Species`
  template: a plausible homeworld (**reuses `worldgen`**, rerolled to Atm 2-9 / Pop 7+, rather than
  a parallel world generator — the book blesses the substitution), the evolutionary `Environment`
  (terrain → Environ DM, locomotion, ecological niche; p.227), the six-slot characteristic profile
  (`CharSpec{Name, Dice}` — C1=Str/C4=Int fixed, C2/C3/C5/C6 rolled) and its `GeneticProfile`
  string (p.228; `RollValue` applies the ≥4D "Rolling Higher" scaling `12+(n-2)D`), `Size` (weighted
  physical-dice total × multiplier, Human=72) and a closed-form `Height` that reproduces the pp.236-237
  grids, the `LifeCycle` (per-stage durations → lifespan, Human=74; p.231), the `Gender` structure +
  determination table (p.230), and — only when C6 is Caste — the `Caste` generation table (p.229,
  resolving the `=Gender`/`=Special`/Unique substitutions). Gender and caste each carry a
  `map[string]Difference` (the C1-dice/C1-C5-flat adjustments a member takes on at assignment).
  Golden-locked to the **Human reference** (`SDEIES`, all-2D, Size 72, lifespan 74) and the p.218/219
  **Ay Flux-0 fixtures** — the book prints no dice-traced worked species, so dense-table interiors are
  locked cell-by-cell in `tables_test.go` instead. The chargen bridge is `chargen.GenerateSophont`
  (`internal/chargen/sophont.go`): it rolls an individual's six characteristics per the species' die
  counts, assigns a gender (and caste, if any) by a 2D roll on the species tables, and applies their
  `Difference`s (no upper cap — an 8D Str reaches 48, so `Character.UPP` now renders out-of-eHex-range
  values as `?` via `ehex.Format`). Deferred until a consumer needs them: the physical/flavor tier
  (senses, body structure, manipulators, special abilities, size-BFP body form, uniques, psionics),
  the caste/gender life-cycle sub-mechanics (shift, assignment timing, caste-gender relation), the
  Skilled-caste skill lists (Chart 12), sophont career service, and species-driven aging.
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
- `internal/shipgen/` — Adventure Class Ship design (Book 2 pp. 30-95). Ship design is
  **deterministic**, not rolled: the designer chooses tonnage/TL/config/structure/armor/drives
  and `Design(spec) ShipSpec` computes the derived performance. `Design` is total — never an
  error; infeasibility (over-budget, underpowered plant, TL-capped drive) is reported in
  `Ship.Problems`. The core insight: the p.78 Z1 Drive Potential grid is a clean formula,
  `drivePotential = min(2*driveOrd/hullOrd, 9)` (`DriveForPotential` is the Z2 inverse). Hull/
  drive letters are the eHex letters (A-Z, no I/O) as an ordinal 1..24 (Hull A=100t … Z=2400t),
  distinct from an eHex value. `Ship.QSP()` renders the compact profile (`S-AL22`, the ship's
  UWP analog); golden-locked to the Murphy Scout and Beowulf. Costs are plain int Cr.
  **Armament** (`weapon.go`/`weapon_data.go`, `defense.go`, `missile.go`, `mounts.go`) is the same
  shape: Book 2 p.83 carries the whole weapon-design system as six tables (devices, TL stage
  effects, mounts, space/world range effects), and p.174 prints the same six for defenses — so one
  rule, `install`, scales both. The stage shifts the device's TL and prices it; the range shifts the
  TL again and scales the **mount's** tonnage and cost ("Range Effects apply to the Mount but not
  the Weapon" — a weapon's own tonnage is zero, so the mount is what takes up room). The mount also
  supplies the damage dice and the attack Mod, and the two mount tables run **opposite ways**: a
  Single Turret attacks at −2 (p.83) and defends at +1 (p.174) — a bigger mount aims worse but
  defends better. `DesignWeapon` (23 models), `DesignDefense` (10 screens, plus the nine weapons
  p.174 allows in the Anti-Missile Defensive Fire mode), and `DesignMissile` (size 1–7 × warhead ×
  guidance, each constraining the others) render the book's `LongName` — a weapon's UWP analog.
  `MinTL`/`MaxTL` bound the designable Tech Level (0..21, Book 2 p.51 — 21 is the design system's
  own ceiling, distinct from the TL-15 Imperial shipyard limit). `Design` does not enforce them (it
  is total); they exist for callers taking a TL from outside, since a negative one renders a ship
  card whose TL field is not an eHex value at all.
  `Design` mounts them against the hull: one HardPoint per 100t, or three FirmPoints instead
  (sub-ton mounts only), with the Bolt-In needing neither — which is why `Tonnage` is fixed-point
  hundredths. Golden-locked to the p.167 and p.176 catalogs, every row. Gotchas: the book divides by
  three by multiplying by **0.33** (a 200t Main at Vlong is "66 tons"), so range multipliers are
  hundredths — do not "fix" this to exact division; and the drive stage table and the weapon one
  stay **separate tables that happen to agree**, because each is printed both ways in the book and
  each is settled by its own worked examples (**Modified** costs /2 on both sides: pp.104/127/134/190
  for drives against the x1 of pp.63/76, pp.83/225/226/251 for weapons against the p.279 appendix).
  The drive side is a **majority** reading, not a clean one: p.48's sample-ship notes work Modified at
  x1, note 14 saying "same pricing per ton" in prose. Four printings and two self-reconciling worked
  columns outweigh two printings and two notes — but the book does not agree with itself, so do not
  re-open this on finding p.48 (#300 was mis-resolved twice that way). The cell is asserted in
  `TestDesignDriveStageCatalogP127`, which is what stops it drifting back.
  Drive stage tonnage **rounds up** and there is no tonnage floor — p.77's "no drive may be smaller
  than the Drive-A of the class" is a floor on the size **letter**, which is the only reading under
  which the worked tables' seven sub-Drive-A rows reproduce. Book conflicts are resolved against the
  design tables and documented at the point of transcription; the p.127 and p.134 stage columns are
  golden-locked cell by cell. Deferred: sensors, crew/accommodations, Quality,
  Batteries, world-surface defenses, and the pp.168–169 interference grid.
- `internal/trade/` — the Trade & Commerce pricing engine (Book 2 pp. 209-221). Speculative cargo
  is bought at a source world for its `Cost` (Cr3,000 base + per-value-class cost mods + Cr100/TL)
  and sold at a market world for a fraction/multiple of its `Price` (Cr5,000 base + source→market
  match mods, scaled 10%/TL-difference), realized through the `ActualValue` table (Flux→40–400%,
  with the capped Broker DM). Pure int-Cr, no dice in the value math; `CargoID` renders the p.221
  identity (e.g. `8-De Hi In Na Po Cr1,800`). Golden-locked to the Free Trader Beowulf journey.
  `shipping.go` adds premium passage pricing, passenger/freight availability + rates, and the Broker
  table (starport gating + commission + `NetSale`). `goods.go`/`goods_data.go` add the Random Trade
  Goods chart (12 TC-keyed columns, 1D type→1D good, Imbalance recursion, Trade Good Detail prefix;
  golden-locked to the Zivije/Knorbes examples). `contracts.go` adds Trader estimation
  (`EstimateActualValue`), the OTO/STS and accelerated-delivery surcharges, and the long-term
  mail-contract bid table. #21 is complete.
- `internal/shipcombat/` — the Space Combat resolution engine (Book 2 pp. 193-204). The Space
  Weapon, Missile, and Defensive Fire tasks are roll-low over `task.ResolveDice` (range-band or 5D
  dice; targets `weaponTL+C+S+K+mods`, `missileTL+guidance+mods`, `defenseTL−attackTL+mountMod`);
  plus `HitCompartment` (Flux + targeting), `Penetrate` (layered armor), the L1 damage-location
  table, damage/diagnosis `Severity`, the missile `MassiveExplosion` proximity table, and movement
  (`Agility`, `RammingHits`, the p.200 range-change grid). The ShipCard compartment model
  (`HullLocations`, p.95 Table H — compartments/span/subcompartments by hull ordinal) backs hit
  location and `SubCompartmentsKnockedOut` damage spread; the missile and weapons-task Massive
  Explosion tables are both present. The tasks still take primitives (TLs/mods/AV/compartment
  numbers) because that is what the book's tables are — but nobody has to invent them: `ship.go`
  bridges the two packages, so `Attack` takes a designed `shipgen.Weapon`, `Defend` a `Defense`,
  `AttackWithMissile` a round, and `ArmorLayers`/`Card`/`ShipAgility` a `Ship`. **A generated ship
  can fight.** Golden-locked to the Murphy/Gryphon, Vanguard/Antares, Joshua, and Vigilant worked
  examples. Note `Armor.AV` is the **per-layer** value, not a total (`ArmorLayers` repeats it, and
  must not divide by the layer count), and `ShipAgility` is capped by `Hull.MaxG`, not just the
  drive. Out of scope: the p.201 interference/clustering tactical options.
- `internal/sectorgen/` + `internal/survey/` + `internal/route/` — interstellar mapping (Book 3
  pp. 12-15, 21, 27-28): the 32×40/8×10 hex grid with CCRR/A-P coordinates, `Hex.Distance` (parsec
  jump distance via even-q offset→cube), and density system-presence rolls. `survey.Sector` is the
  one entry point: it composes them with `systemgen` into detailed Second Survey records — the coarse
  map flags constrain generation (`systemgen.GenerateForMap`: gas-giant symbol→≥1 giants / none→0,
  asteroid symbol→Size-0 belt mainworld) so preview and detail agree — then
  marks sector (Cs) and subsector (Cp) capitals, lays **trade routes** (`route.Build` — a pure,
  dice-free graph linking Ix≥4 worlds within J-4, bridging distant ones through intermediate worlds),
  and sites Scout Way Stations (~1/50 pc of route, bumping Ix). A sector is always surveyed whole —
  its systems share one dice stream, and capitals/routes/way stations need the whole region — so
  every `cmd/sectorgen` view _selects_ from one survey and they agree on what sits in a hex:
  the default lists one subsector, `-sector` lists all of it with the routes, `-hex CCRR` prints
  one system's sheet.
  `Record.Sheet` (`survey/sheet.go`) is the deep renderer: the one-line Second Survey record shows
  only the mainworld — ~94% of what the generators compute (the stellar family, the orbit map, every
  secondary world and moon with its own UWP) plus the mainworld's port facilities, native status, and
  Resource Units have no other renderer. `cmd/sectorgen -hex CCRR` prints one system's full sheet.
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
  through. `internal/cli`'s own tests are an external `package cli_test` — `clitest` imports `cli`,
  so testing from inside would cycle — and they drive `clitest.Command` like any `cmd`, with a
  stand-in `rollgen` whose `Main` dispatches on an env var to the fatal/roll/quiet child.

When adding a generator, transcribe the rule tables/formulas from `docs/reference/` and lock
them with a golden test built from a worked example in the books.

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
