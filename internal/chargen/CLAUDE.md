# chargen

Character creation (Book 1, Characteristics pp. 47+, careers pp. 63-79, Master Chargen
Checklist p. 72). Generates the six-characteristic UPP (Str/Dex/End/Int/Edu/Soc, each 2D,
eHex) at age 18, offers `Check`, and `AgingCheck` (Book 1 p. 89: `2D < LifeStage`, physical
from 34 / mental from 66, zero-cascade to illness/death). `GenerateCareered` runs the
checklist lifecycle, then serves additional careers while `Policy.NextCareer` supplies them
(multi-career via `serveCareer`; the CLI's `-career a,b,c` sequence): UPP → homeworld skills
(`ApplyHomeworldSkills`, one skill per Trade Classification, Book 1 p. 56 — the homeworld is a
`worldgen.World` input) → optional education (`educate`, Book 1 pp. 59-60: remedial ED5, then
either a vocational Trade School (`attendTradeSchool`, one year → a `theTrades` Major +2, no
Minor/Edu-bump/degree, chosen via `Policy.ChooseTradeSchool`) or the best-qualifying academic
program — College or University (undergraduate) then, for a BA-holder electing it
(`Policy.PursueGraduateSchool`), the post-graduate Masters → Professors ladder — all one
parameterized `academicProgram` (years, `awardsMajor`, `awardsMinor`, Edu-or-degree prereq,
grad-Edu, degree) through the shared apply/pass-fail/waiver `attendAcademic`: undergraduates
get Major+1 per pass and Minor+1 per 2 passes with a BA + Edu bump; Masters and Professors use
that same merged Provides cell; golden-locked to the book's Eneri Dinsha College
example `9AB58A`) → Begin → term loop → muster-out.

The remaining p.60 institutions are explicit opt-in operations through `AttendInstitution`:
Service Academies, Medical and Law School, Honors, OTC/NOTC, Flight School, and Command College.
`EducationRecord` preserves formal status (graduation, Honors, commission, Flight branch) without
pretending that pre-career completion is already a served career. `AvailableSkills` transcribes
all seven columns of the p.60 matrix and returns a sorted copy. Medical School and Marine Academy,
both printed as `M`, are separate typed columns. Institution selection, academy
service, and waiver use remain policy decisions; assigned ANM School and Command College timing
remain with the career that assigns them.

Begin (`beginCareer`, Book 1 p. 63) rolls to qualify; the first career Retries a failed Begin
once, later careers do not, and a character refused by every chosen career falls back to the
auto-begin Citizen life (T5 has no draft — no one ends up careerless). Education is gated on
`Policy.PursueEducation`, so a no-education policy leaves any dice trace (e.g. the golden
Scout's) untouched.

The term engine (`career.go`) is career-agnostic with pluggable seams (`ControllingCharMode`
Rotating/Fixed — under FixedControllingChar the policy picks one CC that serves the whole career, `AdvanceRule`
RollLow/RollHigh, `Qualification` char-set, `ContinueRule` fixed/char/UseCC/TermsMod, `BenefitDM`
selecting the muster Benefit-column die modifier (`MusterDM`: Terms/OfficerRank/Rank/FameHalf),
and the rank ladders `EnlistedRanks`/`OfficerRanks` +
`Commission`/`EnlistedPromote`/`OfficerPromote` rules for armed-forces careers); a `Policy` (with
`DefaultPolicy`) supplies every player choice so generation is deterministic and testable.

The rank step (`resolveRank`) runs after Risk & Reward for a surviving armed-forces character: an
enlisted soldier rolls Commission (success → officer track) else Enlisted Promotion, an officer
rolls Officer Promotion; promotion targets are raised by the summed p.70 table mods of the
character's `Medals` — earned on a held Risk (an XS) as well as a passed Reward — and **not** by
`WoundBadges`, a book conflict resolved against the Eneri Dinsha worked example (p.72) and
documented at `promoted`. Each rank grants its automatic skill.

Careers are data, each a file + hand-traced golden: `ScoutCareer` (`scout.go`, p. 79),
`RogueCareer` (`rogue.go`, p. 84 — FixedControllingChar), `SoldierCareer` (`soldier.go`, p. 82 — the first
ranked career), `MarineCareer` (`marine.go`, p. 86), `SpacerCareer` (`spacer.go`, p. 81 — the
naval career, whose Rating ladder uses the engine's EnlistedPromote), `AgentCareer` (`agent.go`,
p. 83 — a rankless `Term UndercoverTerm` whose `awardUndercover` borrows one skill from a rolled
career's grid (`undercoverAssignment` + the `CareerByID` registry) each term, adds the
Successful-Mission skills on a held Risk, and earns a Commendation on a Reward
[`RewardCommendation`, `DMCommends`]; Continue eases with terms served via
`ContinueRule.TermsMod`), `CitizenCareer` (`citizen.go`, p. 78 — an `AutoBegin` career whose
`Term CitizenTerm` (`runCitizenTerm`) replaces Risk & Reward with a benign roll that grants a
Job/Hobby skill and never injures), `EntertainerCareer` (`entertainer.go`, p. 77 — a `Term FameTerm`
whose `runFameTerm` shifts `Character.Fame` by a Flux roll, granting Talent +1 and two extra
skills on a rise, and Continues vs Fame via `ContinueRule.UseFame`), `CraftsmanCareer`
(`craftsman.go`, p. 75 — a `Term CraftsmanTerm` career whose `runCraftsmanTerm` attempts a Masterpiece
from Master Points [CC + Craftsman skill + `skill.Set.TopLevels`], raises the Craftsman skill each
term, and Continues vs Craftsman×2 via `ContinueRule.UseSkill`), `ScholarCareer` (`scholar.go`,
p. 76 — standard Risk & Reward where a Reward is a Publication [`RewardKind`], with a single rank
ladder [`resolveRank` skips Commission when there is no officer track] and Publication-boosted
promotion/continue [`PromotionRule.PubsMod`, `ContinueRule.PubsMod`]), `FunctionaryCareer`
(`functionary.go`, p. 87 — a `Term PoliticsTerm` career whose `runPoliticsTerm` is two unmodified
rolls: a failed Risk ends the career as a job loss [`MusteredOut` from the term, handled in
`RunCareer`], a Reward success is a promotion), `NobleCareer` (`noble.go`, p. 85 — a
`Term IntrigueTerm` career whose `runIntrigueTerm` risks Exile and offers Elevation [a roll-high
check vs Soc that raises Soc and awards a Land Grant]; the Noble's rank is their Social Standing
via `NobleTitle`), and `MerchantCareer` (`merchant.go`, p. 80 — standard Risk & Reward where a
Reward is escalating Ship Shares [`RewardShipShares`, the Nth reward = N shares], with a dual
Rating/Officer rank track). All 13 careers are now implemented.

The Academic grid column uses `AwardMajor` / `AwardMinor` cells that raise the character's College
Major/Minor (lost if uneducated, per the page footnote); `DefaultPolicy.ChooseSkillColumn` is
character-aware, so a graduate specializes in the Academic column while an uneducated Scout falls
through to Courier.

Deferred: Tra-based Apprenticeship/Mentor/Training Course and career assignment of ANM School or
Command College, the Scout's
Courier/Explorer duty and R&R reward, the Rogue's Scheme mechanic, the armed-forces
Branch/Operations R&R mods and commission/promotion skill eligibility, and each career's
documented flavor deferrals (in its own file header). See the per-career `.go` files for the exact
deferred pieces.

## Settled rulings — do not re-open

- **Branch column and Branch-change conflicts (`branch.go`, `BranchOps`).** The Spacer's
  NAVAL BRANCH (p.81) prints Officer and Enlisted columns that disagree on four of its
  eight rolls; every character enters a career enlisted (Book 1 p.64), so the Enlisted
  column is read at career start (`branchFor`) and the Officer column only from Commission
  (`commissionBranch`). WHEN Branch may change is a genuine conflict between two pages,
  resolved in favor of the general rule: the career pages say "Enlisted may select a new
  Branch upon Promotion" (p.81), but p.66 states the rule for the Armed Forces entire — "A
  non-officer character may change (reselect or reroll) Branch at the end of each Term...
  An Officer may not change Branch." The engine follows p.66 (end of every term, not only a
  promoted one) because p.81 is a career-page shorthand for the general section's wider
  rule. The choice itself is the policy's (`Policy.RerollBranch`,
  `Policy.RerollBranchOnCommission`), defaulting to keeping the current Branch — keeping
  rolls no die, so a default character's dice stream is unchanged from before this rule
  existed.
- **Pre-career education cost, verified against the book's own worked example
  (`education.go`).** ED5 alone is free ("no time required", p.60 chart and p.61). The
  book's Eneri Dinsha walkthrough (p.61) prices the rest out: he enters at 18, fails his
  College application, is admitted on Waiver, and completes four years — 18 + 1 + 4 = age
  23, the age the book prints. That confirms both cost rules the code applies: a failed
  application costs the year whether or not a Waiver then rescues it (Waiver attempts
  themselves are free), and each Success spends one year of the program's Duration (p.62's
  training example: "he rolls 8 and fails. A year passes").
- **`Character.Check`'s numDice is a raw dice count, not a `task.Difficulty`
  (`chargen.go`).** A Check's dice count is fixed by how the characteristic was generated,
  not chosen as a judgment of difficulty: "Non-Humans: If the Characteristic checked was
  generated with other than 2D, check Characteristic with the number of Dice used to
  generate it" (Book 1 p.47) — a sophont whose Str is 5D checks Str on 5D. The Book 1 p.120
  Difficulty ladder does name every dice count (Easy 1D through BeyondImpossible 8D,
  Staggering being 5D), but calling that sophont's routine Strength check "Staggering"
  would misdescribe it, and a `task.Difficulty` parameter would invite exactly that
  confusion. Callers working from the ladder convert explicitly with
  `task.Difficulty.Dice()` instead.

The sophont bridge (`sophont.go`) is documented in `internal/sophont/CLAUDE.md`.
