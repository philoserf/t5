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
system presence) · `survey` (detailed sector → full systems + capitals + trade routes) ·
`route` (trade-route graph) · `shipgen` (ACS ship design: hull/drives/fuel/armor + QSP, and the
armament — weapons, mounts, defenses, missiles) ·
`trade` (the p.221 Cargo ID / Cost / Price / Actual Value pricing engine) · `shipcombat` (space
combat resolution: weapon/missile/defensive tasks, penetration, hit location, damage, movement).

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

The worldgen/systemgen/chargen census is **done and audited against the book page by page**:
every UWP- and orbit-determinable trade classification (including the Da/Pz/Fo zone codes, the
Ho/Co climate pair, and satellite codes on every moon), PBG, Ix/Ex/Cx, bases (with Naval Depots),
the complete port-facilities picture (exotic fuels, local fuel, beltports), the world census line,
chargen's 13 careers, and the full multi-star system map all ship. The audit corrected real bugs —
swapped capital codes, a collapsing survey column — and two docs that described working code as
broken. The shared play primitives are now in place too — the **RangeBand ladder** (#9)
and the Difficulty/UTF task layer (#10) — so the play side (senses, personals, combat) is
unblocked. **Starship design (#16) is done**, armament and all, and `shipcombat` (#22) consumes it:
a generated ship can be flown into a fight. The largest open pieces are now `Sophont creation`
(#17, the last big Tier-3 generator) and the Tier-5 content makers, which are independent and can
follow in any order.

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
`TradeClassificationsWithContext` (PR #84). The **Ho/Co** (Hot/Cold) pair was added in the Tier-1
audit: they depend on the ORBIT ALONE (Chart D gives them no UWP columns), which makes them
strictly more determinable than the Tr/Tu/Fr the classifier already emitted. They had been omitted
on the stated grounds that "this edition's Chart B has no Ho/Co pair" — false, and the excuse was
self-inconsistent besides, since the same chart header covers Lk and Tz, which *are* emitted.
Climate codes now reach non-mainworlds too: Chart D is for "the Mainworld **and other worlds in
the system**," and only the mainworld had been getting them.

Every UWP- and orbit-determinable Chart D code is now generated. The referee-assigned
Politicals/Specials (Cp/Cs/Cx/Cy/Mr/Fo/Pz/Da/Ab/An) remain out of scope per Chart D's own header
("Politicals and Specials are assigned by Referee (not generated)") — with the caveat that **Da/Pz
are in fact derivable** from Zone + Pop (B3 p.19: an Amber world with Pop≤6 is Dangerous, Pop≥7 is
Puzzling), so that pair is a genuine remaining gap rather than a referee call. `survey` assigns the
capital codes it can (Cp/Cs) from the region it surveys.

**2. Population multiplier digit / PBG** — _B3 pp. 24–25 (Chart C, PBG)_ · extends worldgen +
systemgen · **S** · ✅ **done**
Pop significant digit = Even-distribution 1–9 (if Pop>0) concatenated with the belt and
gas-giant counts systemgen already produces → the three-digit PBG. `EvenDist1to9` is already
in the engine; this is a roll plus string assembly. It's the literal next worldgen step.

**3. Importance Extension {Ix}** — _B3 pp. 18, 27 (Chart E)_ · extends worldgen · **S** · ✅ **done**
Signed integer = additive DM table over starport, TL, trade classes (**Ag/Hi/In/Ri** — Pa is
*not* on Chart E), Pop, and bases; "Important" at +4. Deterministic, no dice. Feeds Nobility,
capitals, trade routes, and Book 2 mail contracts. **(Implemented — PR #10.)**
_Book conflict:_ the p.22 Regina example counts Pre-Ag **and** omits the Naval+Scout bonus that
Regina plainly earns, reaching +4 by two errors cancelling. The code follows Chart E (p.27), which
reaches the same +4 by a different route — so **Regina cannot discriminate between the two
readings**, and the golden is weaker than it looks. It does catch the single most likely mistake
(adding Pa alone yields +5).

**4. Economic Extension (Ex) + Resource Units** — _B3 pp. 18, 27_ · extends worldgen · **M** · ✅ **done**
`(R L I ±E)` + RU. Resources = 2D (+GG+Belts if TL≥8, both from systemgen); Labor = Pop−1;
Infrastructure branches on Pop band using Ix (Pop 0 → 0; 1-3 → Ix; 4-6 → 1D+Ix; 7+ → 2D+Ix);
Efficiency = Flux. R/L/I floor at 0; Efficiency may be negative, and a negative Efficiency makes
RU negative. **RU = R·L·I·E with the zero substitution: "if any value = 0, use 1 instead (to avoid
multiplying by zero)"** — the book prints that rule on both p.18 and p.27, and the code implements
it. (This entry previously claimed the opposite. It was wrong, and the code was right — a reader
trusting it would have "fixed" working code into a bug.) **(Implemented — PR #10.)**

**5. Cultural Extension [Cx]** — _B3 pp. 18, 27_ · extends worldgen · **S** · ✅ **done**
`[HASS]` — four independent Flux formulas over Pop/Ix/TL with a clamp-to-1 rule, ehex-encoded.
Watch the book's Heterogeneity-vs-Homogeneity naming slip and the Strangeness = "Flux+5"
(chart) vs "2D−2" (example) discrepancy — trust the chart.

**6. Bases · Nobility · Travel Zones · Native Status · Allegiance** — _B3 pp. 19, 24, 28
(Chart F/G)_ · extends worldgen · **S each** · ✅ **done**
A cluster of small per-world attributes. **Bases** (2D vs starport-class thresholds, plus the
region-scoped Naval **Depot** — 1 per 1000 worlds, sited in `survey` like a Way Station);
**Nobility** (trade-class/Ix → noble-code string, golden-locked to Regina's `BcCeF`); **Native
Status** (Pop×Atm×TL → status label, all twelve rows — an inhabited world is assumed TL 1+ per
Chart F); **Travel Zone** (G/A/R from Gov+Law and starport — *population plays no part*, contrary to
this entry's earlier wording, but it does drive the Da/Pz split, item 1); and **Allegiance**, which
is referee-imposed per B3 p.23's checklist and so is a validated code (`Allegiance`,
`ParseAllegiance`) defaulting to Imperial — a typed passthrough, not a table roll.

_Out of scope:_ Chart F's Military/Scientific/Diplomatic/Cultural bases, which the book calls
referee "exceptions" ("Other bases may be established as exceptions").

_Fixed in the Tier-1 audit:_ the **capital codes were swapped** — Chart D (p.26) is unambiguous
(Cp Subsector, Cs Sector, Cx Imperial), and `survey` had marked its sector capital `Cx` and each
subsector capital `Cs`, so every generated sector promoted all sixteen of its subsector capitals to
sector capitals and its own to the capital of the Imperium.

**7. Starport / spaceport facilities & fuel** — _B2 p. 24; B3 p. 24_ · extends worldgen · **S** · ✅ **done**
The starport/spaceport letters are generated (mainworld A–E/X and non-MW spaceports F/G/H/Y),
and `worldgen.PortFacilities` maps a world to its services: shipyard tier, heaviest repair,
hydrogen fuel-type, downport (full/beacon), refuel time (2D/4D hours), the population-gated A/B/C
highport, the **TL-gated exotic fuels** (Radioactives 8+, Collector 14+, Anti-Matter 18+ at A/B —
B3 p.24), the **local unrefined-fuel** fallback for a fuel-less port on a world with water or ice
(the p.24 ** note), and the **Beltport** that replaces a downport at an asteroid mainworld (B2 p.24).
It takes the whole `uwp.Profile` now, since those depend on TL, hydrographics, and size, not just
the class. Golden-tested against every class.

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
drop (PR #88). Satellites carry real bodies — each moon gets a type (the p.29 Satellites tables)
and UWP, capped to its parent's size with a double-planet flag at equal size (Book 3 p.21); a
non-mainworld world whose target orbit is held by a gas giant becomes that giant's moon rather
than being nudged aside. `System.String` and `cmd/systemgen` render the full multi-star orbit map.

**16. Starship design generator** — _B2 pp. 30–95, 101–135, 188–192_ · new `shipgen` package · **L** · ✅ **done**
Built (`internal/shipgen` + `cmd/shipgen`): the deterministic ACS design engine — `Design(spec)
Ship` (never errors; infeasibility is reported in `Ship.Problems`) composing hull (tons/cost/
config/structure/armor/hardpoints/over-tonnage), drives (the **Drive Potential** formula — the
p.78 Z1 grid is exactly `min(⌊2·drive/hull⌋, 9)`, with the Z2 inverse, TL availability, and the
11 stage effects), fuel (`P·hull/10` jump, `P·hull/100` ops), and armor. Renders the QSP
(`S-AL22`) and a ship card; a thin `Generate(r)` rolls a random feasible ship. Golden-locked to
the Murphy Scout (S-AL22 end to end) and Beowulf (overtonnage).

**Armament** (p.83 for weapons, p.174 for defenses — each page carries the whole system as six
tables: the devices, the TL stage effects, the mounts, and the space/world range effects). One
rule, `install`, scales both: the stage shifts the device's TL and prices it, the range shifts the
TL again and scales the **mount's** tonnage and cost ("Range Effects apply to the Mount but not the
Weapon" — a weapon's own tonnage is zero; the mount is what takes up room). The mount also supplies
the damage dice and the attack Mod, and the weapon and defense mount tables run **opposite ways**: a
Single Turret attacks at −2 and defends at +1. `DesignWeapon` (23 models), `DesignDefense` (10
screens, plus the nine weapons p.174 allows in the Anti-Missile Defensive Fire mode), and
`DesignMissile` (sizes 1–7 × warhead × guidance, each constraining the others) render the book's
`LongName` — the weapon's UWP analog. `Design` mounts them: one HardPoint per 100t, or three
FirmPoints instead (sub-ton mounts only), with the Bolt-In needing neither. Golden-locked to the
p.167 TYPICAL WEAPONS and p.176 IDENTIFYING DEFENSES catalogs, every row.

**Deferred:** sensors (pp.136–153), consoles/computers/crew/accommodations, Quality
(Demand/Comfort/Ergonomics), Batteries (p.171), world-surface defenses (p.187), the pp.168–169
weapon/defense-vs-range interference grid, QSP decode, and the non-jump interstellar drives
(Hop/Skip/NAFAL). Kinunir (stage-effect-heavy) awaits verification.

_Rules notes:_ the book divides by three by multiplying by **0.33** (a 200t Main mount at Vlong is
"66 tons"), so range multipliers are hundredths — being more precise than the book would be wrong.
Four book conflicts were resolved against the design tables and documented in comments: the Generic
stage Mod (p.156 says +1, p.83's table says 0), the defense Quad Turret cost (three worked rows need
MCr2.5, the table prints 1.5), the missile catalogs' EMP (one lower than the effects grid on the
same page), and p.155's catalog (does not derive from its own tables). The drive and weapon stage
tables also genuinely differ in one cell: **Modified** costs a drive ×1 but a weapon /2.

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

**21. Trade & commerce** — _B2 pp. 204–221_ · new `trade` package · **M–L** · ✅ **done**
Built (`internal/trade`): the p.221 pricing engine — the Cargo ID, source-world `Cost`
(Cr3,000 base + per-value-class cost mods + Cr100/TL), market-world `Price` (Cr5,000 base +
source→market match mods, ×10%/TL-difference effect), and the `ActualValue` table (Flux→40–400%
of Price with the capped Broker DM). Golden-locked end-to-end to the Free Trader Beowulf worked
journey (Cost/CargoID/Price/SellingPrice incl. the Broker-4 sale). Also the p.220 shipping layer
(`shipping.go`): premium passage pricing (High/Mid/Low by Passage Demand), passenger/freight
availability (Flux + Pop + skill/Liaison, freight ×(valueTCs+1)), freight/mail rates, and the
Broker table (starport availability + 5%/DM commission, `NetSale`). Also the Random Trade Goods
chart (`goods.go`/`goods_data.go`, pp.218-219): the 12 TC-keyed columns (~430 goods) with the
column-selection rules, the 1D type-block → 1D specific-good roll, Imbalance recursion to another
column, and the Trade Good Detail prefix (Industrial/Asteroid omit rules) — golden-locked to the
book's Zivije (Antibiotics) and Knorbes (Imbalance→Pelts) examples. Finally the edge rules
(`contracts.go`): Trader estimation (`EstimateActualValue` bounds the sale once one Flux die is
known), the OTO/STS and accelerated-delivery surcharges, and the long-term mail-contract bid table.
Complete — a ship can price, sell (brokered), fill its hold with passengers/freight, name each
cargo, and bid mail contracts.

**22. Starship combat** — _B2 pp. 193–204_ · new `shipcombat` package · **M–L** · ✅ **done**
Built (`internal/shipcombat`): the full space-combat resolution engine — the Space Weapon, Missile,
and Defensive Fire tasks (roll-low over `task.ResolveDice`: range-band/5D dice, TL+C+S+K /
DefenseTL−AttackTL+Mount targets), mount mods, missile guidance; `HitCompartment` (Flux + targeting)
over the **ShipCard compartment model** (`HullLocations`, the p.86 Table H of compartments/span/
subcompartments per hull); `Penetrate` vs layered armor; the L1 damage-location table; **damage
spread** (`SubCompartmentsKnockedOut`, 10/subcompartment, 60/compartment); damage/diagnosis
`Severity`; the missile `MassiveExplosion` proximity table **and** the p.196 weapons-task multiplier
table; and movement (`Agility`, `RammingHits` = compartments², the p.200 range-change grid).
Golden-locked to the book's Murphy-vs-Gryphon, Vanguard-vs-Antares, Antares-vs-Joshua, and Vigilant
worked examples. The per-weapon/defense stat catalogs and mount installation now live where they
belong — in `shipgen` (#16) — and `shipcombat/ship.go` bridges them: `Attack` takes a designed
`Weapon`, `Defend` a designed `Defense`, `AttackWithMissile` a designed round, and `ArmorLayers` /
`Card` / `ShipAgility` a designed `Ship`. A generated ship can fight. _Still out of scope:_ the
p.201 interference/clustering tactical options.

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
rolls. `survey` wires each stellar hex to a full `systemgen.System` (world names, subsector Cs
and sector Cx capitals) and `internal/route` lays the **trade routes** — a pure, dice-free graph
linking Important (Ix≥4) worlds within J-4, bridging distant ones through intermediate worlds
(Book 3 pp. 21, 27), plus `Hex.Distance` (parsec jump distance) and Scout Way Stations (~1/50 pc,
Book 3 p.28). `cmd/sectorgen -sector` renders the sector with routes and Expected-Ship-Traffic.
_Follow-on:_ the Cy owned-world link and the full T5SS `.sec` route metadata serialization.
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
