# T5 Automation Catalog

A rank-ordered catalog of every Traveller5 rules system that could be captured in code,
built from a full survey of the three core rulebooks. Each entry notes **what** it does,
**where** it lives in the rules, **why** it is automatable, what it **depends on**, and a
rough **effort** (S/M/L).

Page numbers are the printed page markers (e.g. "p. 28"); B1/B2/B3 = Core Rules Book 1
(Characters & Combat), Book 2 (Starships), Book 3 (Worlds & Adventures).

## Already built

`dice` (full engine: rolls, Flux/Good/Bad Flux, even distributions, roll-low Check/Resolve,
notation parser, individual die faces + Spectacular + Many-Dice) · `ehex` · `uwp` ·
`worldgen` (mainworld UWP + all UWP/context-determinable trade classifications + secondary
"other" worlds) · `systemgen` (full star family, gas-giant detail, belts, mainworld with
orbit/climate/satellite, detailed "other" worlds, per-world satellites, and a concrete
multi-star orbit map placing every world/GG/belt round-robin across all hosting stars, plus
port facilities/fuel) · `chargen` (six-characteristic UPP + Check Characteristic + the 13
careers) · `task` (Difficulty/UTF resolve) · `calendar` · `rangeband` (world/space range ladder) ·
`senses` · `personals` · `combat` (the play tier, core mechanics) · `sectorgen` (hex map +
system presence) · `survey` (detailed sector → full systems + capitals) · `shipgen` (ACS ship
design: hull/drives/Drive-Potential/fuel/armor + QSP).

## Status legend

Each item is tagged ✅ **done**, 🟡 **partial** (core built, noted gaps remain), or ⬜ **not
started**. Tier 1 (the world/character/system census) is complete. Everything in Tiers 4–6
(starships, sophonts, play systems, content makers) is ⬜ **not started** unless tagged
otherwise below.

## How this is ranked

Rank is a composite of **leverage** (does it complete or unblock the project's core
generators / does much depend on it), **readiness** (are its dependencies already built),
and **inverse effort/risk** (small, unambiguous, image-verifiable transcription beats
large dense grids). Foundational shared primitives rank high even when unglamorous because
they unblock whole tiers below them. Within each tier, items are ordered by that composite.

## Critical path (build order at a glance)

The worldgen/systemgen/chargen census is **done**: trade classifications, PBG, Ix/Ex/Cx,
the world census line, port facilities, chargen's 13 careers, and the full multi-star system
map all ship. The shared play primitives are now in place too — the **RangeBand ladder** (#9)
and the Difficulty/UTF task layer (#10) — so the play side (senses, personals, combat) is
unblocked. `Starship design` (#16) is the largest open piece. Everything in Tiers 5–6 is
independent content-generation that can follow in any order.

---

## Tier 1 — Finish the world & character census (ready, high-value, small)

These leverage what's already built, are mostly pure lookups/formulas over the existing
UWP/star data, and close out the "generate a canonical world/system line" goal.

**1. Remaining trade classifications** — _B3 p. 26 (Chart D); B2 pp. 204–207_
· extends worldgen · **S** · ✅ **done**
Complete. Beyond the original 21 UWP codes, the classifier now emits the Gov=0/Law=0 pair
**Di/Ba** and **Px/Re** (do-now, PR #76); the HZ-orbit **climate** codes Tr/Tu/Fr/Tz
(`ClimateCodes`, PR #78 — this edition's Chart B has no Ho/Co pair); the satellite codes
**Sa/Lk** (PR #79); and the context-dependent Secondary codes **Fa/Mi/Pe** via
`TradeClassificationsWithContext` (PR #84). Every UWP- and context-determinable Chart D code
is generated; only the referee-assigned Politicals/Specials (Cp/Cs/Cx/Cy/Mr/Fo/Pz/Da/Ab/An)
remain, and those are intentionally out of scope (Chart D: "assigned by Referee").

**2. Population multiplier digit / PBG** — _B3 pp. 24–25 (Chart C, PBG)_ · extends worldgen +
systemgen · **S** · ✅ **done**
Pop significant digit = Even-distribution 1–9 (if Pop>0) concatenated with the belt and
gas-giant counts systemgen already produces → the three-digit PBG. `EvenDist1to9` is already
in the engine; this is a roll plus string assembly. It's the literal next worldgen step.

**3. Importance Extension {Ix}** — _B3 pp. 18, 27 (Chart E)_ · extends worldgen · **S** · ✅ **done**
Signed integer = additive DM table over starport, TL, trade classes (Pa/Ag/Hi/In/Ri), Pop,
and bases; "Important" at +4. Deterministic, no dice. Its trade-code inputs are already built.
Feeds Nobility, capitals, trade routes, and Book 2 mail contracts. **(Implemented — PR #10.)**

**4. Economic Extension (Ex) + Resource Units** — _B3 pp. 18, 27_ · extends worldgen · **M** · ✅ **done**
`(R L I ±E)` + RU. Resources = 2D (+GG+Belts if TL≥8, both from systemgen); Labor = Pop−1;
Infrastructure branches on Pop band using Ix; Efficiency = Flux; RU = R·L·I·E straight (the
printed rule has no "0→1" substitution). R/L/I floor at 0; Efficiency may be negative.
**(Implemented — PR #10.)**

**5. Cultural Extension [Cx]** — _B3 pp. 18, 27_ · extends worldgen · **S** · ✅ **done**
`[HASS]` — four independent Flux formulas over Pop/Ix/TL with a clamp-to-1 rule, ehex-encoded.
Watch the book's Heterogeneity-vs-Homogeneity naming slip and the Strangeness = "Flux+5"
(chart) vs "2D−2" (example) discrepancy — trust the chart.

**6. Bases · Nobility · Travel Zones · Native Status · Allegiance** — _B3 pp. 19, 24, 28
(Chart F/G)_ · extends worldgen · **S each** · ✅ **done** (Allegiance is referee-supplied, default Im)
A cluster of small per-world attributes: Bases (2D vs starport-class thresholds), Nobility
(trade-class/Ix → noble-code string), Travel Zone (Gov+Law and Pop thresholds → G/A/R),
Native Status (Pop×Atm×TL → status label), Allegiance (enumerated code, referee-supplied,
default Im). Each is a lookup or a threshold predicate over data already in hand.

**7. Starport / spaceport facilities & fuel** — _B2 p. 24; B3 p. 24_ · extends worldgen · **S** · ✅ **done**
The starport/spaceport letters are generated (mainworld A–E/X and non-MW spaceports F/G/H/Y),
and `worldgen.PortFacilities` maps each class to its services (Book 2 p.24): shipyard tier,
heaviest repair, hydrogen fuel-type, downport (full/beacon), refuel time (2D/4D hours), and the
population-gated A/B/C highport. Golden-tested against every class. _Deferred:_ the TL-gated
exotic fuels (radioactives/anti-matter/collector, B3 p.24), outside this item's fuel-type scope.

**8. Second Survey line formatter** — _B3 pp. 16, 23_ · extends worldgen · **S** · ✅ **done**
Serialize everything above into the canonical `Hex Name UWP TC {Ix}(Ex)[Cx] N B Z PBG W A
Stellar` line (e.g. Regina's full record). Pure string assembly; a golden test against the
book's Regina line validates the whole Tier-1 stack.

---

## Tier 2 — Foundational play engines (high leverage, unblock Tiers 4–5)

**9. RangeBand ladder** — _B1 pp. 24–29 (+ 43, 186, 200)_ · new shared primitive · **M** · ✅ **done**
`internal/rangeband`: two-way Range↔distance↔descriptor lookup on both the R= (world, 0–9) and
S= (space, 0–13) scales plus the lettered bands (Contact/Reading/Talking, Boarding) and the
space-combat band letters (F1/F2/SR/AR/LR/DS), with the S = R−5 conversion and log sub-band
interpolation. The most-reused new utility in B1 — Senses, Combat, Personals, and travel-time
will sit on it. Tables golden-locked to the p.24/p.29 charts.

**10. Difficulty / Universal Task Format layer** — _B1 pp. 120–131_ · extends dice · **S–M** · 🟡 **partial**
Built (`internal/task`): the Difficulty ladder (Easy 1D … Beyond Impossible 8D) with
Hasty/Cautious pace over `task.Resolve`/`ResolveDice` (Target = Char+Skill+ΣMods → roll-low),
and Spectacular via the dice engine. **Not built:** the Cooperative/Opposed/Uncertain task
modes. Everything in Tier 4 (combat, senses, personals) is expressed in these terms.

**11. Expose individual die faces + Many-Dice / Good-Bad Flux** — _B1 pp. 259–261_ · extends
dice · **S** · ✅ **done** (PR #75)
Complete. `DiceFaces(n)` surfaces individual dice; `CheckResult.Faces` + `Spectacular()`
classify three-1s/6s results; the four Many-Dice fast methods for 11D+ (reuse-10,
2D-subsample, ×3.5, 3.5-Flux) are implemented and golden-tested. Good/Bad Flux already
existed in the engine.

**12. Money · Value · Cost scales** — _B1 pp. 20, 36–37, 44–45_ · new primitive · **S** · ⬜ **not started**
An integer credit type (avoid float), the Value tier↔credits log scale, production-cost
divisors, and Flux-driven supply/demand & quality multipliers. Foundational for every economy
system (land grants, trade, ship costs, muster-out).

**13. Imperial Calendar date type** — _B1 pp. 262–263_ · new primitive · **S** · ✅ **done**
365-day year, named weeks + Holiday, day-of-year↔weekday, birthdate/age arithmetic. Small
`calendar` package that chargen birthdates, aging ticks, and time-degradation effects all use.

---

## Tier 3 — The big generators (highest value, large effort)

**14. Chargen careers** — _B1 pp. 63–99_ · extends chargen · **L** (the flagship piece) · ✅ **done**
The full career life-cycle: 2D career selection, per-term Controlling-Characteristic rotation,
the Risk & Reward roll, injury/wound/disabled/dead consequences, branch/operations/medals,
promotion/commission (incl. the roll-high Noble Elevation), skill-eligibility rolls into the
13 per-career skill grids, continue/aging triggers, and mustering-out (benefit tables, reroll,
pensions, characteristic improvement, land grants). Every mechanic is a Check or table roll;
the effort is the dense per-career data (13 careers × ~4 tables each) — transcribe from images
and golden-test against a worked character. Prerequisites: homeworld skills (B1 p. 56, extends
worldgen→chargen), education engine (B1 pp. 58–62), aging (B1 pp. 88–89), the skills/knowledges
data model (B1 pp. 132–171).

**15. Systemgen per-world detailing & placement** — _B3 pp. 20–29 (Chart G)_ · extends systemgen
· **L** · ✅ **done**
Complete. HZ orbit by spectral type/size + orbital distances (`HZOrbit`/`OrbitAU`, PR #77);
mainworld orbit + climate (PR #78, unblocked #1's climate codes); mainworld satellite
(PR #79); gas-giant detail — size/diameter/skim-gravity/class + every-second-SGG→IG
(PR #80); the P2 chart + Book 1 p.31 sub-orbit floors (PR #81); the placement engine that
assigns the mainworld, gas giants, belts, and other worlds to concrete orbits with
duplicate/precluded collision resolution (PR #82); secondary/"other" world detailing —
per-type partial-UWP generation + zone-based type selection + context trade codes
(PRs #83, #84); corrected HZ zone boundaries (PR #87); per-world satellites (PR #86); and
**"Rotate Placement Per Star"** — round-robin placement across the Primary/Close/Near/Far
hosts, each with its own habitable zone, sub-orbit floor, and Orbit N-3 range, with overflow
drop (PR #88). `System.String` and `cmd/systemgen` render the full multi-star orbit map.
_Minor documented deferral:_ a non-mainworld world landing on a gas-giant orbit is nudged to
a free orbit rather than becoming that giant's moon.

**16. Starship design generator** — _B2 pp. 30–95, 101–135, 188–192_ · new `shipgen` package · **L** · 🟡 **core done**
Built (`internal/shipgen` + `cmd/shipgen`): the deterministic ACS design engine — `Design(spec)
Ship` (never errors; infeasibility is reported in `Ship.Problems`) composing hull (tons/cost/
config/structure/armor/hardpoints/over-tonnage), drives (the **Drive Potential** formula — the
p.78 Z1 grid is exactly `min(⌊2·drive/hull⌋, 9)`, with the Z2 inverse, TL availability, and the
11 stage effects), fuel (`P·hull/10` jump, `P·hull/100` ops), and armor. Renders the QSP
(`S-AL22`) and a ship card; a thin `Generate(r)` rolls a random feasible ship. Golden-locked to
the Murphy Scout (S-AL22 end to end) and Beowulf (overtonnage). **Deferred:** weapons/sensors/
defenses (the mount/stage/range grids, B2 pp.153–192), consoles/computers/crew/accommodations,
Quality (Demand/Comfort/Ergonomics), the ShipCard compartment layout, QSP decode, and the
non-jump interstellar drives (Hop/Skip/NAFAL). Kinunir (stage-effect-heavy) awaits verification.

**17. Sophont (species) creation** — _B3 pp. 215–246_ · extends chargen · **L**
A Flux-driven pipeline producing a non-human species template (homeworld, environment/niche,
characteristics + genetic profile, caste & gender structures, life stages, senses, body
structure, special abilities, size/height, uniques/metamorphosis, psionics, TL cap) that
chargen then consumes. ~14 sub-generators, many sharing tables with BeastMaker (niche/body)
and the `geom` size machinery. This is what makes chargen work for aliens.

---

## Tier 4 — Play/simulation subsystems (build on Tier 2)

**18. Personal combat** — _B1 pp. 200–227_ · new `combat` package · **M–L** · 🟡 **partial**
Built (`internal/combat`): the Move-Attack-Damage round as roll-low Tasks — the three Combat
Numbers (Shooting/Melee/Impact), `TargetSize` (Size−Range + stance), the Ranged (R=Range dice,
This-Is-Hard +1D), Melee (2D vs AMN−DMN), and Impact (2D vs C2−Speed, Speed² damage) attacks,
and damage resolution (`Absorb` vs armor, `Wound` cascading Str/Dex/End). Golden-locked to the
p.202/203 Roberto-Landor and Cayne-Corbett examples. **Not built:** attack modes
(Aimed/Standard/SnapFire), the weapon effect codes + Pen/Protection nuances, Knockdown/KO/Quick
Kill, and the after-battle recovery pass.

**19. The Senses** — _B1 pp. 186–199; B3 senses_ · new `senses` package · **M** · 🟡 **partial**
Built (`internal/senses`): the generic sense Action — the six senses with human Constants
(Vision/Hearing 16, Smell 10, Touch 6; Awareness/Perception non-human), `NoticeAtRange`
(nD = Range, roll-low vs Constant + (Size−Range) + Mods) and `NoticeInContact` (2D), plus
`RangeBand(meters)` off the rangeband primitive. Golden-locked to the p.190 cargo-mover example.
**Not built:** the fixed-format sense-ID codecs (V-16-RGB spectrum, H-16 octaves, smell UOP
matching) and per-sense detail systems, and the special-sound timed status-effect.

**20. The Personals (social interaction)** — _B1 pp. 180–185_ · new `personals` package · **M–L** · 🟡 **partial**
Built (`internal/personals`): the roll-low Personal — Purpose→dice (Carouse 1D … Command 4D),
per-purpose Strategy base values (`StrategyValue`), the Five Laws mod table (`LawMod`), `Resolve`
(Target = Strategy×Tactic + Law + Mods), the Camaraderie counter, and the Brazen/Urgent/Repeat
mods. Golden-locked to the p.184 example. **Not built:** the full Strategy×Tactic compatibility
matrix (which tactics apply per strategy and their ×2/×3 multipliers — dense p.183 grid), the
quick-NPC generator, and the Violence-fail→combat hook.

**21. Trade & commerce** — _B2 pp. 204–221_ · new `trade` package · **M–L**
Cargo IDs, random trade goods (two 1D rolls into ~19 TC-keyed columns with imbalance recursion),
passenger/freight availability (Flux + Pop + skill/TC DMs), premium pricing, broker hire, and
the buy/sell Actual-Value loop (Flux → 40–400% with capped broker/trader DMs). Consumes worldgen
TCs/TL + chargen skills; the multi-page goods tables are the transcription load.

**22. Starship combat** — _B2 pp. 193–204_ · new `shipcombat` package · **M–L**
Agility-ordered rounds, the space-weapon / missile / defensive-fire tasks, Flux hit-location on
the ShipCard, penetration vs layered armor, damage spread across compartments, massive-explosion
proximity, and ranging movement. Depends on shipgen's compartment model.

---

## Tier 5 — World / creature / equipment content generators (independent)

**23. World surface mapping** — _B3 pp. 36–56_ · new `worldmap` package · **L**
The icosahedral geodesic hex model (20 triangles, per-Size hex counts, pents, edge wrap) plus
the 30-step terrain-layout algorithm keyed on UWP/Ex/trade-classes and the World→Terrain→Local→
Single drill-down populator. High worldbuilding value; the geodesic geometry is the hard part.
Includes gas-giant and cylindrical-habitat map variants and the terrain/speed/altitude/depth/
horizon data tables.

**24. BeastMaker & encounters** — _B3 pp. 247–269_ · new `beastgen` package · **M**
Per-world/terrain encounter tables of fully-generated animals (niche, quantity, size/strength
with grav/atm DMs, locomotion, speed, natural weapons, reactions, body structure, edibility,
training) plus non-animal event tables. Shares niche/body/geom tables with Sophont creation.

**25. VehicleMaker** — _B3 pp. 132–158_ · new `vehiclegen` package · **L**
Column-modifier design system over ground/flyer/watercraft charts → tonnage/speed/load/armor/
protections/cost + derived values (Beastpower, speed↔kph, endurance→range, occupants). Shares
the speed/Beastpower tables with BeastMaker.

**26. ThingMaker (equipment)** — _B3 pp. 159–173_ · new `thinggen` package · **M**
Function + TL + size + profile + density + construction → dimensions/volume/mass/8-channel
protection/damage/range-effects/power/cost. Mostly deterministic formulas; the shared `geom`
size/profile/density machinery (used by Thing, Beast, and Sophont) should be built once here.

**27. GunMaker / ArmorMaker (equipmentgen)** — _B1 pp. 234–235; B3 pp. 95–131_ · new package · **M–L**
Chained category→type→descriptor→burden→stage rolls producing a weapon/armor designation and
stats (same shape as chargen career tables) + availability tables. Produces the effect-coded
weapons and 8-stat armor that personal combat consumes; a static catalog of pre-made items is
the data alternative.

**28. Sector / subsector generation + hex addressing** — _B3 pp. 12–15_ · new `sectorgen` package
· **S–M** · ✅ **done**
Built (`internal/sectorgen` + `cmd/sectorgen`): the 32×40 / 8×10 hex geometry with CCRR
coordinates and the A–P subsector partition (`Hex.String`/`Hex.Subsector`), the eight stellar
densities with their system-presence rolls (Book 3 p.13), and per-hex contents (gas giant on
2D≤8, asteroid-belt mainworld on a natural 2). `GenerateSector`/`GenerateSubsector` populate a
region; the CLI prints a subsector map. Golden-locked to the geometry and the density/contents
rolls. _Follow-on:_ wiring each stellar hex to a full `systemgen.System` (world names,
capitals, allegiance, trade routes, the Cy owned-world link, the T5SS `.sec` line).
---

## Tier 6 — Support primitives, niche generators, and reference data

**29. Genetics · offspring inheritance · clones · synthetics · chimera** — _B1 pp. 102–119_ ·
extends chargen · **M** (chimera **L**, needs the Sophont card)
DNA encoding + 81 genetic profiles, gene inheritance with dominance/mutation/litter/pedigree,
and the sophontoid/clone/chimera pipelines (characteristic-point formulas, cost, markings,
accelerated aging). Depends on the dice engine exposing individual die faces (#11).

**30. Psionics (testing + action resolution)** — _B3 pp. 198–214_ · extends chargen · **M**
Psi×6 aptitude point distribution and per-action task checks with additive/quadratic size/mass
penalties, sense benchmarks, and drug/attention modifiers. Pairs with the Sophont psionics
generator.

**31. Benchmark scales** — _B1 pp. 36–43_ · shared primitives · **S each**
Fame/Danger/Threat (0–36), Risk (three Flux), Impact (Speed²·tons), Hot-and-Cold (temperature↔
Hits), Insulation, and the Size benchmark. Small tables reused by combat, worldgen hazards, and
chargen Fame.

**32. Master Mod tables + generic table engine** — _B1 pp. 264–269_ · new `mastermod` data package
· **L** (data), **S** (engine)
~90 named Nd→row referee tables (Visibility, Attitude, Crime, Emotion, Sensor Responses, etc.)
on one trivial data-driven lookup engine with placeholder substitution. Also powers the chargen
background tables (educational institutions, life motivations, secrets, organizations, important
events — B1 pp. 92–97) and the career/muster-out tables. The engine is small; the data is bulk.

**33. EPIC adventure scaffolds + themes** — _B3 pp. 270–279_ · standalone/aggregator · **S–M**
A fixed 4-act/5-scene/climax template plus a 1D×1D themes roller; naturally an aggregator that
stitches together outputs of every other generator ("there's more…" hooks).

**34. Travel-time / kinematics / orbital distances** — _B1 pp. 31–35_ · extends systemgen · **M** · 🟡 **partial**
The orbit→AU table (`systemgen.OrbitAU`, Book 1 p.31) is built as part of #15. **Not built:**
orbit→km/light-time conversions, constant-accel travel time, and light-delay — closed-form
kinematics on top of the existing orbital data.

**35. Weapon / armor / equipment / ship static catalogs** — _B1 pp. 234–249; B2 worked ships;
B3 equipment catalog_ · reference data · **S–M**
Structured pre-generated item and ship data — best sourced via scripted extraction, used both as
content and as golden fixtures for the corresponding makers.

---

## Notes on overlaps and shared machinery

- **Trade classifications** appear in both B2 (p. 205) and B3 (p. 26) — one system; extend the
  existing worldgen classifier, adding the climate codes once systemgen supplies HZ-relative orbit.
- **Orbital distances / habitable zones / satellite orbits / drive limits** appear in both B1
  (pp. 31–33) and B3 (pp. 20, 28) — consolidate into systemgen; verify against what's built.
- **Senses** span B1 (play side) and B3 Sophont (species-definition side) — one `senses` model
  serves both; **native status** likewise (B3 worlds and B3 sophont homeworld share the table).
- **`geom`** (Size / Profile / Density / Volume / Mass) recurs verbatim in ThingMaker, BeastMaker,
  and Sophont creation — build once.
- **Shared build-first primitives** that many entries depend on: RangeBand (#9), the Difficulty/UTF
  layer (#10), individual-die exposure (#11), Money/Value (#12), the Calendar (#13), and the generic
  dice→table-with-placeholders engine (#32). Doing these early makes the tiers above cheaper.

## Method that works

For any dense grid (stellar tables, career skill grids, trade-goods columns, effect tables):
render the page to an image and transcribe cell-by-cell — `pdftotext` shifts rows — then lock the
system with a golden test built from a worked example in the book (as worldgen does with Regina).
Ship each system as its own reviewed PR.
