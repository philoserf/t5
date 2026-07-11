# T5 Automation Catalog

A rank-ordered catalog of every Traveller5 rules system that could be captured in code,
built from a full survey of the three core rulebooks. Each entry notes **what** it does,
**where** it lives in the rules, **why** it is automatable, what it **depends on**, and a
rough **effort** (S/M/L).

Page numbers are the printed page markers (e.g. "p. 28"); B1/B2/B3 = Core Rules Book 1
(Characters & Combat), Book 2 (Starships), Book 3 (Worlds & Adventures).

## Already built

`dice` (full engine: rolls, Flux/Good/Bad Flux, even distributions, roll-low Check/Resolve,
notation parser) · `ehex` · `uwp` · `worldgen` (mainworld UWP + UWP-determinable trade
classifications) · `systemgen` (stars with spectral type/decimal/size, orbital placement,
gas giants, belts, world count, mainworld) · `chargen` (six-characteristic UPP + Check
Characteristic).

## How this is ranked

Rank is a composite of **leverage** (does it complete or unblock the project's core
generators / does much depend on it), **readiness** (are its dependencies already built),
and **inverse effort/risk** (small, unambiguous, image-verifiable transcription beats
large dense grids). Foundational shared primitives rank high even when unglamorous because
they unblock whole tiers below them. Within each tier, items are ordered by that composite.

## Critical path (build order at a glance)

`finish trade classifications → PBG + Ix/Ex/Cx → world census line` completes worldgen.
In parallel, `RangeBand + Difficulty/UTF task layer` unlocks the entire play side (tasks,
senses, personals, combat). `chargen careers` is the single highest-value large piece.
Everything in Tiers 5–6 is independent content-generation that can follow in any order.

---

## Tier 1 — Finish the world & character census (ready, high-value, small)

These leverage what's already built, are mostly pure lookups/formulas over the existing
UWP/star data, and close out the "generate a canonical world/system line" goal.

**1. Remaining trade classifications** — _B3 p. 26 (Chart D); B2 pp. 204–207_
· extends worldgen · **S**
The classifier already covers the 21 UWP-determinable codes, including the economic (Pi, In,
Po, Pr, Ri, Ag, Na, Pa) and population (Lo, Ni, Ph, Hi) ones. Genuinely missing are the
Gov=0/Law=0 pair **Di/Ba** (Ba also needs starport E/X) — pure digit-set membership, a couple
more `tcRules` rows — and the climate/orbit/context codes (Fr, Ho, Co, Tr, Tu, Tz, Fa, Sa, Lk,
Cy, Cp/Cs/Cx…) that need systemgen HZ-orbit or region context. The UWP-determinable core that
Ix/Nobility/Ex depend on is done.

**2. Population multiplier digit / PBG** — _B3 pp. 24–25 (Chart C, PBG)_ · extends worldgen +
systemgen · **S**
Pop significant digit = Even-distribution 1–9 (if Pop>0) concatenated with the belt and
gas-giant counts systemgen already produces → the three-digit PBG. `EvenDist1to9` is already
in the engine; this is a roll plus string assembly. It's the literal next worldgen step.

**3. Importance Extension {Ix}** — _B3 pp. 18, 27 (Chart E)_ · extends worldgen · **S**
Signed integer = additive DM table over starport, TL, trade classes (Pa/Ag/Hi/In/Ri), Pop,
and bases; "Important" at +4. Deterministic, no dice. Its trade-code inputs are already built.
Feeds Nobility, capitals, trade routes, and Book 2 mail contracts. **(Implemented — PR #10.)**

**4. Economic Extension (Ex) + Resource Units** — _B3 pp. 18, 27_ · extends worldgen · **M**
`(R L I ±E)` + RU. Resources = 2D (+GG+Belts if TL≥8, both from systemgen); Labor = Pop−1;
Infrastructure branches on Pop band using Ix; Efficiency = Flux; RU = R·L·I·E straight (the
printed rule has no "0→1" substitution). R/L/I floor at 0; Efficiency may be negative.
**(Implemented — PR #10.)**

**5. Cultural Extension [Cx]** — _B3 pp. 18, 27_ · extends worldgen · **S**
`[HASS]` — four independent Flux formulas over Pop/Ix/TL with a clamp-to-1 rule, ehex-encoded.
Watch the book's Heterogeneity-vs-Homogeneity naming slip and the Strangeness = "Flux+5"
(chart) vs "2D−2" (example) discrepancy — trust the chart.

**6. Bases · Nobility · Travel Zones · Native Status · Allegiance** — _B3 pp. 19, 24, 28
(Chart F/G)_ · extends worldgen · **S each**
A cluster of small per-world attributes: Bases (2D vs starport-class thresholds), Nobility
(trade-class/Ix → noble-code string), Travel Zone (Gov+Law and Pop thresholds → G/A/R),
Native Status (Pop×Atm×TL → status label), Allegiance (enumerated code, referee-supplied,
default Im). Each is a lookup or a threshold predicate over data already in hand.

**7. Starport / spaceport facilities & fuel** — _B2 p. 24; B3 p. 24_ · extends worldgen · **S**
Starport letter → yards/repair/downport/highport/fuel-type + refuel time (2D/4D); non-MW
spaceports (1D, roll = Pop−1D → F/G/H/Y); highport presence predicate. Pure lookups off the
existing starport field.

**8. Second Survey line formatter** — _B3 pp. 16, 23_ · extends worldgen · **S**
Serialize everything above into the canonical `Hex Name UWP TC {Ix}(Ex)[Cx] N B Z PBG W A
Stellar` line (e.g. Regina's full record). Pure string assembly; a golden test against the
book's Regina line validates the whole Tier-1 stack.

---

## Tier 2 — Foundational play engines (high leverage, unblock Tiers 4–5)

**9. RangeBand ladder** — _B1 pp. 24–29 (+ 43, 186, 200)_ · new shared primitive · **M**
Two-way Range↔distance↔descriptor lookup on both the R= (world, 0–9) and S= (space, 0–13)
scales plus lettered special bands, with the S = R−5 conversion and sub-band interpolation.
The single most-reused new utility in B1 — Senses, Combat, Personals, Size, and travel-time
all sit on it. Build first among the play primitives.

**10. Difficulty / Universal Task Format layer** — _B1 pp. 120–131_ · extends dice · **S–M**
Generalizes the existing Check/Resolve: a Difficulty ladder (Easy 1D … Beyond Impossible 8D,
with Hasty/Cautious ±1 column) computes the dice pool and a Target Number = Char+Skill+ΣMods,
then delegates to the roll-low resolver. Adds Cooperative/Opposed/Uncertain/Spectacular
task modes. Everything in Tiers 4 (combat, senses, personals) is expressed in these terms.

**11. Expose individual die faces + Many-Dice / Good-Bad Flux** — _B1 pp. 259–261_ · extends
dice · **S**
Spectacular results (three 1s/6s), Genetics (first die = gene), and Uncertain tasks all need
the roller to surface individual dice, not just the sum. Many-Dice fast methods for 11D+
(reuse-10, 2D-subsample, ×3.5, 3.5-Flux) matter for nuclear/WMD damage. Good/Bad Flux already
exist in the engine — just confirm and wire in.

**12. Money · Value · Cost scales** — _B1 pp. 20, 36–37, 44–45_ · new primitive · **S**
An integer credit type (avoid float), the Value tier↔credits log scale, production-cost
divisors, and Flux-driven supply/demand & quality multipliers. Foundational for every economy
system (land grants, trade, ship costs, muster-out).

**13. Imperial Calendar date type** — _B1 pp. 262–263_ · new primitive · **S**
365-day year, named weeks + Holiday, day-of-year↔weekday, birthdate/age arithmetic. Small
`calendar` package that chargen birthdates, aging ticks, and time-degradation effects all use.

---

## Tier 3 — The big generators (highest value, large effort)

**14. Chargen careers** — _B1 pp. 63–99_ · extends chargen · **L** (the flagship piece)
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
· **L**
Habitable-zone orbit by spectral type/size, orbital distances & drive limits, mainworld orbit
placement + climate (which unblocks the climate trade classifications), gas-giant detailing,
satellites, secondary/"other" worlds (each a per-type partial-UWP formula), and the
rotate-placement scheduler that assigns every world/GG/belt to a concrete orbit with
collision/precluded-orbit resolution. The placement scheduler is the stateful hard part; the
rest is lookups + per-type UWP formulas reusing worldgen.

**16. Starship design generator** — _B2 pp. 30–95, 101–135, 188–192_ · new `shipgen` package · **L**
The largest self-contained system in the set: hull tons/cost/config, structure & armor layers,
drives (TL availability, stage effects, tonnage/EP/cost, the Potential formula, nexi, fuel),
the shared four-table mount/stage/range build pipeline for sensors/weapons/defenses, power-plant
matching, consoles/computers/crew, quality (Demand/Comfort/Ergonomics), payload as the
tonnage-budget residual, QSP encode/decode, and the ShipCard compartment layout. Mostly
closed-form formulas + lookups + a budget-closure constraint; the worked-example ships (Murphy
Scout, Beowulf, Kinunir) are ready-made golden fixtures.

**17. Sophont (species) creation** — _B3 pp. 215–246_ · extends chargen · **L**
A Flux-driven pipeline producing a non-human species template (homeworld, environment/niche,
characteristics + genetic profile, caste & gender structures, life stages, senses, body
structure, special abilities, size/height, uniques/metamorphosis, psionics, TL cap) that
chargen then consumes. ~14 sub-generators, many sharing tables with BeastMaker (niche/body)
and the `geom` size machinery. This is what makes chargen work for aliens.

---

## Tier 4 — Play/simulation subsystems (build on Tier 2)

**18. Personal combat** — _B1 pp. 200–227_ · new `combat` package · **M–L**
Round state machine (initiative, move/attack/damage phases, range-band movement), ranged/melee/
impact attack resolution, and the damage system: type×select multiplier, the 26-effect table
mapping each effect to which of 8 protection channels it checks, armor subtraction, knockdown,
and hit cascade across characteristics. Plus explosion/WMD/nuclear/environmental proximity
tables and the battle-damage aftermath chain. All lookups + formulas over the effect vocabulary.

**19. The Senses** — _B1 pp. 186–199; B3 senses_ · new `senses` package · **M**
Generic sense-check (nD = Range, roll-low vs Constant + Benchmark + Mods) plus fixed-format
sense-ID codecs (V-16-RGB, H-16-…) and per-sense systems (vision spectrum, hearing octaves,
smell UOP matching, touch, perception, awareness). Mostly string codecs + static tables on the
RangeBand primitive; introduces the first timed status-effect (special sounds).

**20. The Personals (social interaction)** — _B1 pp. 180–185_ · new `personals` package · **M–L**
Purpose (1D–4D) + Strategy×Tactic matrix + the five Laws + camaraderie counters → Target
Number, roll-low. Deterministic over a large, detail-sensitive Strategy×Tactic compatibility
matrix; quick-NPC generator included. Violence-fail hooks into combat.

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
· **S–M**
System-presence rolls by stellar density, per-hex gas-giant/belt flags, and the fixed
32×40 / 8×10 hex geometry with CCRR coordinates and A–P subsector partition. Feeds systemgen and
provides the multi-world context that capitals/trade-routes/Cy-classifications need.

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

**34. Travel-time / kinematics / orbital distances** — _B1 pp. 31–35_ · extends systemgen · **M**
Orbit→AU/km/light-time, constant-accel travel time, and light-delay. Closed-form kinematics;
verify overlap with systemgen's orbital data before rebuilding.

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
