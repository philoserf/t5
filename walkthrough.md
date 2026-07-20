# T5 Walkthrough

_2026-07-20T13:16:20Z by Showboat 0.6.1_
<!-- showboat-id: a27465ad-36f6-4d95-828e-f3f52705c941 -->

## Overview

`t5` is a Go implementation of the **Traveller5** tabletop RPG rules: a set of
generators that roll up worlds, star systems, sectors, characters, alien
sophonts, starships, trade cargoes, and space combat — all from the same seeded
dice engine.

The repo holds three kinds of work in one place: the Go generators, a rules
reference extracted from the T5 core rulebooks, and worldbuilding notes. This
walkthrough covers the Go code.

The design has one shape, repeated at every tier:

1. **Transcribe the book's tables and formulas as pure functions** that take
   their die rolls as arguments, so each can be tested at its edges.
2. **A `Generate` function rolls in the book's checklist order** and composes
   those pure functions into a record.
3. **Lock the result with a golden test built from a worked example** printed in
   the rulebook.

Everything is layered on `internal/dice`. Nothing calls `math/rand` directly.

```bash
cat <<'TREE'
cmd/            six CLIs: worldgen systemgen chargen sectorgen shipgen sophont
internal/
  dice/         the T5 dice engine — the single source of randomness
  ehex/         extended-hex digits (0-9, A-Z minus I and O)
  uwp/          the Universal World Profile type (A788899-C)
  task/         Universal Task Format — the roll-low difficulty ladder
  skill/        skills and cascade knowledges
  calendar/     the 365-day Imperial Calendar
  rangeband/    the world/space range ladder
  worldgen/     mainworld UWP + trade codes + extensions
  systemgen/    stars, gas giants, belts, orbit map, satellites
  sectorgen/    the hex grid and stellar density
  survey/       sector survey + the deep per-system sheet
  route/        trade routes between worlds
  chargen/      character creation: UPP, education, 13 careers
  sophont/      the Sophont Creation System (aliens)
  shipgen/      Adventure Class Ship design and armament
  shipcombat/   space combat resolution
  trade/        speculative cargo, shipping, contracts
  senses/       sense Actions
  personals/    social interaction
  combat/       personal combat
  cli/          shared CLI conventions (flags, seeds, streams)
  clitest/      end-to-end harness all six CLIs share
TREE
```

```output
cmd/            six CLIs: worldgen systemgen chargen sectorgen shipgen sophont
internal/
  dice/         the T5 dice engine — the single source of randomness
  ehex/         extended-hex digits (0-9, A-Z minus I and O)
  uwp/          the Universal World Profile type (A788899-C)
  task/         Universal Task Format — the roll-low difficulty ladder
  skill/        skills and cascade knowledges
  calendar/     the 365-day Imperial Calendar
  rangeband/    the world/space range ladder
  worldgen/     mainworld UWP + trade codes + extensions
  systemgen/    stars, gas giants, belts, orbit map, satellites
  sectorgen/    the hex grid and stellar density
  survey/       sector survey + the deep per-system sheet
  route/        trade routes between worlds
  chargen/      character creation: UPP, education, 13 careers
  sophont/      the Sophont Creation System (aliens)
  shipgen/      Adventure Class Ship design and armament
  shipcombat/   space combat resolution
  trade/        speculative cargo, shipping, contracts
  senses/       sense Actions
  personals/    social interaction
  combat/       personal combat
  cli/          shared CLI conventions (flags, seeds, streams)
  clitest/      end-to-end harness all six CLIs share
```

### The dependency layering

The packages stack strictly. `dice` sits at the bottom and imports nothing from
the repo; every tier above builds on the one below.

    cmd/*  →  cli  →  dice
       ↓
    worldgen → uwp → ehex
       ↓
    systemgen → worldgen
       ↓
    sectorgen / survey / route

Character, ship, trade, and combat form a parallel stack on the same base:
`chargen → skill + task + sophont`, `shipcombat → shipgen`, `trade → uwp`.

Two things to hold onto before reading further, because they explain most of
the odd-looking code:

- **The dice stream is an observable.** Two runs with the same seed must
  produce byte-identical records. So generators often roll a value, then throw
  it away, rather than skipping the roll — skipping would shift everything
  drawn afterwards.
- **Panic on programmer error, `?` on display.** Strict paths (`ehex.Digit`,
  `Difficulty.Dice`) panic on out-of-domain input, because returning a
  plausible value would hide the bug. Display paths (`ehex.Format`,
  `Difficulty.String`) render `?` instead, because a `String` method must not
  crash.

```bash
cat go.mod
```

```output
module github.com/philoserf/t5

go 1.26.5
```

No third-party dependencies at all — the whole thing is the standard library.

---

## 1. `internal/dice` — the engine everything sits on

Traveller uses only six-sided dice. `nD` means "roll n D6 and sum them". On top
of that the books define Flux (`D-D`, ranging −5..+5), its Good and Bad
variants, the half-die, and several "even distributions" that contort D6
results into ranges like 1-9.

A `Roller` is the single source of die rolls. Note that the die itself is held
as a **function**, which is what makes every generator in the repo testable.

```bash
sed -n '20,35p' internal/dice/dice.go
```

```output
// seeded) or NewWithSeed (deterministic, for tests and reproducible worlds).
// The zero value is not usable.
type Roller struct {
	// d6 returns a single die result in 1..6. Held as a func so tests can
	// substitute a scripted sequence.
	d6 func() int

	// seed is the value the generator was built from, kept so that even an
	// auto-seeded Roller can name it; seeded is false for a Roller drawing from
	// a caller-supplied sequence, which has no seed to name.
	seed   uint64
	seeded bool
}

// New returns a Roller seeded from the runtime's random source. The drawn seed
// is kept and reported by Seed, so output from an auto-seeded Roller stays
```

There are three constructors. `New` draws a fresh seed _and keeps it_ — never
draws and discards — so even an unseeded run can be replayed afterwards.
`NewWithSeed` is the deterministic one. `NewSource` takes an arbitrary
`func() int`, which is how tests inject a scripted die.

The scripted roller is the repo's central testing device, and it is
deliberately unforgiving:

```bash
sed -n '78,102p' internal/dice/dice.go
```

```output
		if f < 1 || f > 6 {
			panic(fmt.Sprintf("dice: NewScripted face %d at index %d is not a die face (want 1..6)", f, i))
		}
	}

	i := 0

	return NewSource(func() int {
		if i >= len(faces) {
			panic(fmt.Sprintf("dice: NewScripted script exhausted after %d faces; "+
				"the test consumed more dice than it scripted, so the script no longer "+
				"describes the rolls being made — extend it to cover every roll", len(faces)))
		}

		v := faces[i]
		i++

		return v
	})
}

// Seed reports the seed the Roller was built from and whether it has one. A
// Roller from NewSource or NewScripted draws from a supplied sequence rather
// than a seeded generator, so it reports ok false.
func (r *Roller) Seed() (uint64, bool) {
```

**Exact, not cyclic.** A script that runs out panics instead of wrapping around.
That is what makes a green test suite evidence that the dice stream is
unchanged: a test must enumerate _every_ die the code under test draws, so if a
generator starts consuming one more roll, the test fails loudly rather than
being quietly served recycled faces.

The primitives themselves are small. Flux is the T5 signature roll:

```bash
grep -n -A4 'func (r \*Roller) Flux\|func (r \*Roller) GoodFlux\|func (r \*Roller) HalfDie' internal/dice/dice.go
```

```output
143:func (r *Roller) Flux() int {
144-	return r.d6() - r.d6() //nolint:staticcheck // SA4000: d6 is stateful; two distinct rolls
145-}
146-
147-// GoodFlux rolls two dice and subtracts the smaller from the larger, ranging
--
149:func (r *Roller) GoodFlux() int {
150-	a, b := r.d6(), r.d6()
151-
152-	return max(a, b) - min(a, b)
153-}
--
169:func (r *Roller) HalfDie() int {
170-	return (r.d6() + 1) / 2
171-}
```

### The roll-low check

T5 resolves actions by rolling _under_ a target. Two kinds of adjustment exist
and the book is careful to distinguish them — a **Mod** adjusts the target
(higher is easier, it is an asset), a **DM** adjusts the roll (higher is
harder).

```bash
sed -n '25,36p' internal/dice/check.go
```

```output
// defaults to DefaultCheckDice. See Book 1, "Mods Versus DMs" (p. 19).
type Check struct {
	Dice   int
	Target int
	Mod    int
	DM     int
}

// A CheckResult reports the outcome of resolving a Check.
//
// Success is the plain arithmetic outcome, Total <= Target. It is NOT the Book 1
// p. 127 Spectacular override: the book calls that a property of a *task* result
```

```bash
sed -n '75,91p' internal/dice/check.go
```

```output
	roll := 0
	for _, f := range faces {
		roll += f
	}

	total := roll + c.DM
	target := c.Target + c.Mod

	return CheckResult{
		Roll:    roll,
		Total:   total,
		Target:  target,
		Success: total <= target,
		Effect:  target - total,
		Faces:   faces,
	}
}
```

`Resolve` keeps the individual `Faces`, not just the sum, because Book 1 p.127
defines a _Spectacular_ result off the raw dice: three or more ones is a
Spectacular Success, three or more sixes a Spectacular Failure, and both at
once (only reachable on 6D+) is "Spectacularly Interesting".

```bash
sed -n '48,68p' internal/dice/spectacular.go
```

```output

	for _, f := range faces {
		switch f {
		case 1:
			ones++
		case 6:
			sixes++
		}
	}

	switch {
	case ones >= 3 && sixes >= 3:
		return SpectacularlyInteresting
	case sixes >= 3:
		return SpectacularFailure
	case ones >= 3:
		return SpectacularSuccess
	default:
		return NotSpectacular
	}
}
```

### The `dice` / `task` split

Here is the first architectural decision worth stopping on. `dice` **classifies**
a Spectacular roll but does not **act** on it. `CheckResult.Success` in `dice`
is plain arithmetic, `Total <= Target`, and nothing more.

The p.127 override lives one layer up, in `internal/task`, which owns Book 1
pp.120-131. The reasoning is that the book states the rule about _tasks_ —
"Sometimes the task result is Spectacular" — not about dice. So `dice` keeps the
dice observation and `task` keeps the consequence. A caller rolling a
`dice.Check` that isn't a task gets arithmetic, with no opt-out flag to
remember.

---

## 2. `internal/task` — the Universal Task Format

`task` provides the difficulty ladder: Easy (1D) through Beyond Impossible (8D).
More dice make a low roll harder, so higher difficulty is genuinely harder.

```bash
sed -n '17,33p' internal/task/task.go
```

```output
type Difficulty int

// Task difficulties (Book 1 p. 120).
const (
	Easy Difficulty = iota
	Average
	Difficult
	Formidable
	Staggering
	Hopeless
	Impossible
	BeyondImpossible
)

var difficultyNames = [...]string{
	"Easy", "Average", "Difficult", "Formidable",
	"Staggering", "Hopeless", "Impossible", "Beyond Impossible",
```

There is a naming trap here the code guards explicitly. A `Difficulty` is a
ladder **index** (`Average` is 1); a `dice.Check{Dice: …}` is a **count** (an
Average check rolls 2D). The two numbering schemes disagree, so `dice` refuses
to name any difficulty at all — it exports exactly one count, `DefaultCheckDice`
— and the conversion is `Difficulty.Dice()`, which panics off-ladder rather
than returning a count that would silently resolve as an ordinary check:

```bash
sed -n '48,56p' internal/task/task.go
```

```output
	if d < Easy || d > BeyondImpossible {
		panic(fmt.Sprintf("task: difficulty %d out of range %d..%d",
			int(d), int(Easy), int(BeyondImpossible)))
	}

	return int(d) + 1
}

// Hasty returns the dice count when rushing the task: one level harder (+1D).
```

And the override itself — the whole reason `task` exists as a separate layer:

```bash
sed -n '115,141p' internal/task/task.go
```

```output
// on 6D or more) is deliberately left alone: the book describes it as "a
// situation involving both Spectacular Success and Spectacular Failure (and a
// sign that the referee should make [the] situation a rousing, interesting
// event)" and never states whether the task itself succeeds. That is a referee
// call, so the arithmetic outcome stands and the caller can detect the case via
// CheckResult.Spectacular.
func applySpectacular(res dice.CheckResult) dice.CheckResult {
	switch res.Spectacular() {
	case dice.SpectacularSuccess:
		res.Success = true
	case dice.SpectacularFailure:
		res.Success = false
	case dice.NotSpectacular, dice.SpectacularlyInteresting:
		// Arithmetic stands, per the comment above.
	}

	return res
}

func sum(xs []int) int {
	total := 0
	for _, x := range xs {
		total += x
	}

	return total
}
```

---

## 3. `internal/ehex` and `internal/uwp` — the notation

Traveller writes numbers 0-33 as a single character: 0-9, then A-Z with **I and
O omitted** (they read too easily as 1 and 0). A world Size of 10 prints as
`A`, a Tech Level of 12 as `C`.

```bash
sed -n '15,21p' internal/ehex/ehex.go
```

```output
// its digit. I and O are omitted, following standard Traveller usage.
const Alphabet = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// Max is the largest value a single eHex digit can represent.
const Max = len(Alphabet) - 1

// Digit returns the eHex character for v, which must be in 0..Max. It panics
```

`Digit` panics out of range (generators clamp their outputs, so a bad value is a
bug); `Format` returns `?` for the display paths. The same strict/total pairing
as `Difficulty.Dice` and `Difficulty.String`.

Eight eHex digits make a **Universal World Profile** — the compact StSAHPGL-T
summary that is the atom of Traveller worldbuilding. Regina is `A788899-C`:
Starport A, Size 7, Atmosphere 8, Hydrographics 8, Population 8, Government 9,
Law 9, Tech Level C (12).

```bash
sed -n '18,29p' internal/uwp/uwp.go
```

```output
// does — because a Profile alone does not say which kind of world it describes.
type Profile struct {
	Starport      byte
	Size          int
	Atmosphere    int
	Hydrographics int
	Population    int
	Government    int
	Law           int
	TechLevel     int
}

```

### One digit that is a code, not a dimension

Size 0 renders an **asteroid belt** — a field of rubble with no diameter. But
whether a Size-0 body actually _is_ a belt depends on what the body is, and that
subtlety is the whole point. A **mainworld** with Size 0 is a belt (Book 3 p.16,
"determined when World Size is generated"). A **secondary** Size-0 world usually
is not: a Worldlet is a genuinely tiny solid world that renders the same
`St000...`, and only a `Planetoids` body is a belt. A `Profile` alone cannot tell
the two apart — so it must not try:

```bash
sed -n '30,49p' internal/uwp/uwp.go
```

```output
// BeltSize is the Size digit an asteroid belt renders with. It is a *code*, not
// a dimension: a belt has no diameter, so the digit means "field of asteroids".
//
// Whether Size 0 MEANS a belt depends on the body, and that is why the reader is
// not a Profile method:
//
//   - A mainworld with Size 0 is a belt — Book 3 p.16, "determined when World
//     Size is generated" — whether the sector map forced it or the 2D-2 roll came
//     up 0. That fact is carried as worldgen.World.Belt.
//   - A SECONDARY world with Size 0 is usually NOT a belt: a Worldlet rolls a tiny
//     solid world that renders the same St000..., and only a Planetoids body is a
//     belt. That fact is its worldgen.OtherWorldType (IsBelt).
//
// So a Profile alone cannot answer "is this a belt" — the mainworld and the
// Worldlet share the profile — which is why uwp.Profile.IsBelt was removed in #328
// (it was Size == BeltSize, right for a mainworld but wrong for a Worldlet, and it
// stamped phantom As-with-moons on Size-0 Worldlets, #324/#213/#200/#309). Ask the
// body fact instead. BeltSize remains for its two genuine dimension uses: the
// belt's rendered digit, and the smallest Size a satellite cap floors against.
const BeltSize = 0
```

---

## 4. `internal/worldgen` — the generator archetype

This is the package to read first if you want to understand the shape of every
other generator. It builds a mainworld from Book 3's world-creation checklist.

Each characteristic is a small formula over the dice and the characteristics
already rolled. Crucially, **the formulas are pure functions taking their rolls
as arguments** — so they can be tested at their edges without a roller — and
`Generate` does all the rolling, in the book's checklist order:

```bash
sed -n '35,58p' internal/worldgen/worldgen.go
```

```output
	sp := starport(r.Dice(2))

	size := rollSize(r)
	if belt {
		size = 0
	}

	atm := atmosphere(r.Flux(), size)
	hyd := hydrographics(r.Flux(), atm, size)
	pop := rollPopulation(r)
	gov := government(r.Flux(), pop)
	lawLevel := law(r.Flux(), gov)

	return uwp.Profile{
		Starport:      sp,
		Size:          size,
		Atmosphere:    atm,
		Hydrographics: hyd,
		Population:    pop,
		Government:    gov,
		Law:           lawLevel,
		TechLevel:     techLevel(r.Die(), sp, size, atm, hyd, pop, gov),
	}
}
```

Note the `belt` branch: when an asteroid mainworld is forced, the size roll is
**still made and then superseded**. That is the dice-stream discipline in
miniature — skipping the roll would shift every later draw and produce a
different world from the same seed.

The formulas themselves are direct transcriptions of the book's tables:

```bash
sed -n '60,86p' internal/worldgen/worldgen.go
```

```output
// starport maps a 2D roll to a starport quality letter (Book 3 table 1).
func starport(twoD int) byte {
	switch {
	case twoD <= 4:
		return 'A'
	case twoD <= 6:
		return 'B'
	case twoD <= 8:
		return 'C'
	case twoD == 9:
		return 'D'
	case twoD <= 11:
		return 'E'
	default:
		return 'X'
	}
}

// rollSize is 2D-2; a result of 10 is rerolled as 1D+9 for the largest worlds.
func rollSize(r *dice.Roller) int {
	if s := r.Dice(2) - 2; s != 10 {
		return s
	}

	return r.Die() + 9
}

```

Each characteristic cascades from the previous one: Atmosphere is Flux+Size,
Hydrographics is Flux+Atmosphere (with a −4 for thin or very dense air),
Government is Flux+Population, Law is Flux+Government. Tech Level takes a
modifier from the starport and from every single characteristic.

```bash
grep -n -A12 '^func atmosphere\|^func hydrographics' internal/worldgen/worldgen.go
```

```output
90:func atmosphere(flux, size int) int {
91-	return sizedAtmosphere(flux+size, size)
92-}
93-
94-// hydrographics is Flux+Atmosphere, zero for worlds smaller than size 2, with a
95-// -4 modifier for thin or very dense atmospheres, clamped to A. As with
96-// atmosphere, the size rule and the clamp live in sizedHydrographics.
97:func hydrographics(flux, atm, size int) int {
98-	h := flux + atm
99-	if atm < 2 || atm > 9 {
100-		h -= 4
101-	}
102-
103-	return sizedHydrographics(h, size)
104-}
105-
106-// rollPopulation is 2D-2; a result of 10 is rerolled as 2D+3.
107-func rollPopulation(r *dice.Roller) int {
108-	if p := r.Dice(2) - 2; p != 10 {
109-		return p
```

### The golden test

This is how every generator in the repo is locked down. Book 3 prints a worked
example — Regina — so the test feeds the book's exact die rolls through a
scripted roller and demands the book's exact answer:

```bash
sed -n '136,152p' internal/worldgen/worldgen_test.go
```

```output
// TestGenerateRegina feeds the exact die rolls from Book 3's Regina worked
// example, in checklist order, and asserts the canonical UWP A788899-C.
func TestGenerateRegina(t *testing.T) {
	rolls := []int{
		2, 2, // Starport 2D = 4 -> A
		4, 5, // Size 2D = 9 -> 7
		2, 1, // Atmosphere Flux = +1 -> 8
		3, 3, // Hydrographics Flux = 0 -> 8
		5, 5, // Population 2D = 10 -> 8
		2, 1, // Government Flux = +1 -> 9
		3, 3, // Law Flux = 0 -> 9
		6, // Tech Level 1D = 6, +6 (Starport A) -> 12
	}
	if got := Generate(dice.NewScripted(rolls...)).String(); got != "A788899-C" {
		t.Fatalf("Generate (Regina) = %q, want A788899-C", got)
	}
}
```

### From a Profile to a World

A bare UWP is not the whole record. `World` adds everything Book 3 Charts C-F
derive from it: trade classifications, the `{Ix}(Ex)[Cx]` Extensions, nobility,
bases, travel zone, and native status.

```bash
sed -n '13,30p' internal/worldgen/world.go
```

```output
// derived data from Book 3 Charts C-F — trade classifications, the {Ix}(Ex)[Cx]
// Extensions, nobility, bases, travel zone, native status, and the population
// multiplier digit.
type World struct { //nolint:recvcheck // deliberate value-reader / pointer-mutator split
	Profile         uwp.Profile
	TradeCodes      []string
	Importance      int
	Economic        Economic
	Cultural        Cultural
	Nobility        string
	NavalBase       bool
	ScoutBase       bool
	NavalDepot      bool
	WayStation      bool
	Zone            byte
	NativeStatus    string
	PopulationDigit int
}
```

`generateWorld` composes it, and the ordering is load-bearing — bases are rolled
before Importance because Importance depends on them:

```bash
grep -n -A20 'func generateWorld' internal/worldgen/world.go
```

```output
58:func generateWorld(r *dice.Roller, gasGiants, belts int, isCapital, belt bool) World {
59-	p := generate(r, belt)
60-	tcs := TradeClassificationsWithContext(p, WorldContext{IsMainworld: true})
61-	naval, scout := RollBases(r, p.Starport)
62-	ix := Importance(p, tcs, naval, scout, false)
63-
64-	return World{
65-		Profile:         p,
66-		TradeCodes:      tcs,
67-		Importance:      ix,
68-		Economic:        RollEconomic(r, p, ix, gasGiants, belts),
69-		Cultural:        RollCultural(r, p, ix),
70-		Nobility:        Nobility(tcs, ix, isCapital),
71-		NavalBase:       naval,
72-		ScoutBase:       scout,
73-		Zone:            TravelZone(p),
74-		NativeStatus:    NativeStatus(p),
75-		PopulationDigit: PopulationDigit(r, p.Population),
76-	}
77-}
78-
```

### The positional record

`SecondSurvey` renders the world's line. The full Regina record is
`A788899-C Ph Pa Ri {+4}(D7E+4)[9C6D] BcCeF NS -` — UWP, trade codes,
extensions, nobility, bases, zone.

```bash
grep -n -A12 'func (w World) SecondSurvey' internal/worldgen/world.go
```

```output
168:func (w World) SecondSurvey() string {
169-	fields := []string{
170-		w.Profile.String(),
171-		dashIfEmpty(strings.Join(OrderTradeCodes(w.TradeCodes), " ")),
172-		w.Extensions(),
173-		w.Nobility,
174-		dashIfEmpty(w.bases()),
175-		dashIfEmpty(w.zone()),
176-	}
177-
178-	return strings.Join(fields, " ")
179-}
180-
```

The trailing `-` is not decoration. This record is **positional**, so every
empty field is dashed rather than dropped. A world matching no trade code at all
is a real thing (`C539700-8` is one — its atmosphere, hydrographics and
population between them exclude every rule), and collapsing its TC column would
shift the extensions, nobility, bases and zone one place left. A reader, or a
parser, would then read its extensions as its trade codes.

Let's actually run it:

```bash
go run ./cmd/worldgen -n 5 -seed 42
```

```output
D665656-7  Ga Ni Ag Ri
C7A5958-A  Fl Hi In
B160113-B  De Lo
E621896-6  He Ph Na Pi Po
D555236-6  Lo
```

---

## 5. `internal/systemgen` — the system around the world

A mainworld does not float alone. `systemgen` rolls the stars, the gas giants,
the planetoid belts, the world count, the orbit map, and the moons.

The `System` struct shows the shape of a T5 stellar family: a Primary that
always exists, plus Close, Near, and Far secondaries and a companion for each,
present only when rolled — hence pointers.

```bash
sed -n '22,45p' internal/systemgen/systemgen.go
```

```output

// A System is a generated star system. The Primary always exists; the other
// stars are present only when rolled, so they are pointers that are nil when
// absent. Each of Primary, Close, Near, and Far may have a companion.
type System struct { //nolint:recvcheck // deliberate value-reader / pointer-mutator split
	Primary          Star
	PrimaryCompanion *Star
	Close            *Star
	CloseCompanion   *Star
	Near             *Star
	NearCompanion    *Star
	Far              *Star
	FarCompanion     *Star

	// Orbit numbers of the secondary stars around the Primary (Book 3 p. 28),
	// valid only when the corresponding star is present. A companion orbits
	// inside its own star's orbit and has no separate number; the Primary's
	// companion sits inside Orbit 0.
	CloseOrbit int
	NearOrbit  int
	FarOrbit   int

	GasGiants int
	Belts     int
```

`generate` runs the checklist. Gas giants and belts come first because the
mainworld's Economic Extension and PBG digits need them; the stars come next;
then the mainworld is placed relative to the primary's habitable zone:

```bash
sed -n '99,145p' internal/systemgen/systemgen.go
```

```output

func generate(r *dice.Roller, gg ggConstraint, asteroidMainworld bool) System {
	var s System
	// Gas giants and belts are needed to detail the mainworld (the Economic
	// Extension and the PBG use them), so roll them before the mainworld.
	s.GasGiants, s.Giants = giantsFor(r, gg)
	s.Belts = belts(r)
	// Capital status needs region context (sectorgen), so default to false.
	if asteroidMainworld {
		s.Mainworld = worldgen.GenerateBeltWorld(r, s.GasGiants, s.Belts, false)
	} else {
		s.Mainworld = worldgen.GenerateWorld(r, s.GasGiants, s.Belts, false)
	}

	var primaryType, primarySize int

	s.Primary, primaryType, primarySize = rollStar(r, true, 0, 0)

	// With the primary known, place the mainworld relative to its habitable zone
	// and tag its climate / satellite trade codes (Book 3 p.24).
	s.MainworldOrbit, s.MainworldSatellite = placeMainworld(r, s.Primary, &s.Mainworld)

	// Every non-primary star — secondaries and companions alike — derives from
	// the primary's Flux values (Book 3 p. 28). present rolls one optional star,
	// returning nil unless its presence Flux clears the threshold.
	present := func() *Star {
		if r.Flux() < starPresent {
			return nil
		}

		star, _, _ := rollStar(r, false, primaryType, primarySize)

		return &star
	}

	s.PrimaryCompanion = present()
	s.Close = present()
	s.Near = present()

	s.Far = present()
	if s.Close != nil {
		s.CloseCompanion = present()
	}

	if s.Near != nil {
		s.NearCompanion = present()
	}
```

### Dice-stream discipline, in full

`GenerateForMap` is the interesting variant. A coarse sector map already says
whether a hex has a gas giant and whether its mainworld is an asteroid belt, so
detail generation has to _honor a constraint it did not roll_.

The naive fix — skip the gas-giant roll when the map already decided — would
shift every subsequent draw and give a different mainworld from the same seed.
So instead the count is rolled, the giants are detailed, and only then is the
constraint applied:

```bash
sed -n '215,228p' internal/systemgen/systemgen.go
```

```output
func giantsFor(r *dice.Roller, gg ggConstraint) (int, []GasGiant) {
	rolled := gasGiants(r)
	detailed := rollGasGiants(r, rolled)
	count := clampGasGiants(rolled, gg)

	switch {
	case count < rolled: // ggAbsent: rolled, detailed, discarded
		detailed = detailed[:count]
	case count > rolled: // ggPresent over a rolled 0: the unalignable extra roll
		detailed = append(detailed, rollGasGiants(r, count-rolled)...)
	}

	return count, detailed
}
```

A `ggAbsent` system still rolls its giants' 2D sizes and throws them away, so
nothing drawn afterwards moves. One case genuinely cannot align — `ggPresent`
over a rolled 0, where the demanded giant has no roll in the unconstrained
stream to re-derive from — and the doc comment says so out loud, then adds a
warning about the arithmetic:

```bash
sed -n '205,214p' internal/systemgen/systemgen.go
```

```output
// The one case that cannot align is ggPresent over a rolled 0 — the giant the
// constraint demands has no roll in the unconstrained stream to re-derive from,
// so its 2D is drawn last, after the aligned segment. Everything drawn before it
// still matches; belts onward is shifted by that one roll.
//
// That case is common, not a corner: gasGiants is max(2D/2-2, 0) in INTEGER
// division, so it returns 0 for 2D of 2, 3, 4 and 5 — 5/2 is 2, not 2.5. Ten
// outcomes in thirty-six, about 28%, roughly one system in 3.6. Reading the
// predicate as 2D <= 4 and the frequency as one in six is the natural mistake,
// and it was made twice while this was written.
```

That last paragraph is worth reading twice. `max(2D/2-2, 0)` is **integer**
division, so 2D of 2, 3, 4 _and_ 5 all give 0 — `5/2` is 2, not 2.5. Ten
outcomes in thirty-six, about 28%, not the one-in-six you'd guess from reading
the predicate as `2D <= 4`. The comment records that the mistake was made twice
while the function was being written.

You will meet this register throughout the repo: comments that adjudicate the
_rules_ rather than explain the code, usually anchored to a named test.

Here is a whole system:

```bash
go run ./cmd/systemgen -seed 7
```

```output
Primary: M6 V
Orbits:
        0: Mainworld (1 moon: Dee Y69A000-0 dp)
        1: Gas Giant T LGG (2 moons: Eff YAD4035-5; Aitch Y891002-6)
        2: Worldlet G000686-4 Va Ni Na
        3: Belt
        4: Iceworld Y541210-3 He Lo Po Fr
        5: Iceworld Y646121-6 Lo Fr (1 moon: Ring)
        6: Big World YEDA334-6 Oc Lo (4 moons: Ring; Tee Y423200-5; Ess Y524001-4; Yu Y8B3356-7)
        7: Iceworld Y59A003-9 Wa Fr (3 moons: Tee G10087A-7; Ee Y566000-6 dp; Gee Y5D3321-8 dp)
        8: Belt
        9: Iceworld Y110201-B Lo (3 moons: Yu F100676-9 dp; Eff Y120321-A dp; Pee Y000122-7)
        10: Iceworld Y160135-6 De Lo
        13: Iceworld Y566100-3 Lo Fr (3 moons: Gee Y554586-7 dp; Eff Y522000-8 dp; Cee Y433000-0)
Worlds: 12  PBG: 821
Mainworld: C65397A-A Hi Po Tz {+2}(B89-2)[9B79] BE - -
```

---

## 6. Mapping — `sectorgen`, `survey`, `route`

These three are best read as one composition. `sectorgen` owns the geometry and
the coarse map, `route` is a pure dice-free graph algorithm, and `survey.Sector`
is the single entry point that composes both with `systemgen`.

### The geometry

A sector is 32 columns × 40 rows of hexes, named `CCRR` (`0436` is column 4,
row 36). Distance uses the standard even-q offset → cube conversion, and it is
the whole reason `route` can ask "is this within Jump-4":

```bash
sed -n '85,95p' internal/sectorgen/sectorgen.go
```

```output
// the Traveller map (Book 3 p.12). Traveller hexes are flat-topped and arranged
// in columns with even columns shifted half a hex down ("even-q" offset); the
// distance is the cube-coordinate distance after converting from that offset.
func (h Hex) Distance(o Hex) int {
	ax, az := h.Col, h.Row-(h.Col+(h.Col&1))/2
	bx, bz := o.Col, o.Row-(o.Col+(o.Col&1))/2
	ay, by := -ax-az, -bx-bz

	return (max(ax-bx, bx-ax) + max(ay-by, by-ay) + max(az-bz, bz-az)) / 2
}

```

Note the strict/total split again, applied to a _filing_ decision. `Hex.String`
returns `"????"` for an off-map hex because it is display. `Hex.Subsector`
panics, because a plausible-but-wrong letter would silently misfile a world into
a subsector it doesn't belong to.

### Density, and overriding the printed book

Whether a hex holds a star system at all is a density roll. One table drives the
name, the dice count, and the threshold — and its comment is the clearest
example in the repo of the code deliberately disagreeing with the printed
rulebook:

```bash
sed -n '150,170p' internal/sectorgen/sectorgen.go
```

```output
// densityInfo is the single table of each density's display name and its
// system-presence check: roll `dice` D6 and a system is present at or under
// `threshold` (Book 3 p.13). Core uses 2D <= 10, not the row's literal "11 or
// less": that text conflicts with every other Core figure the book prints — the
// stated 91% density, the Per-Sector 1170/1280 (91.4%), and the Count-Off [12]
// (~92%) all triangulate to 2D <= 10 (33/36 = 91.7%), while 2D <= 11 is 97.2%.
// The "11" reads as a typo for "10"; we follow the corroborated 91%.
var densityInfo = [...]struct {
	name            string
	dice, threshold int
}{
	ExtraGalactic: {"Extra Galactic", 3, 3},
	Rift:          {"Rift", 2, 2},
	Sparse:        {"Sparse", 1, 1},
	Scattered:     {"Scattered", 1, 2},
	Standard:      {"Standard", 1, 3},
	Dense:         {"Dense", 1, 4},
	Cluster:       {"Cluster", 1, 5},
	Core:          {"Core", 2, 10},
}

```

### Why a sector is atomic

`GenerateSector` walks all 1280 hexes in column-major order and rolls contents
for each present system:

```bash
sed -n '236,258p' internal/sectorgen/sectorgen.go
```

```output
}

// GenerateSector rolls all 1280 hexes of a sector at the given density and
// returns the populated ones in column-major CCRR order (Book 3 pp.12-13).
//
// A sector is the only unit that can be rolled. Rolling a sub-region on its own
// would consume a different run of the dice, so its hexes would hold different
// systems than the same coordinates do in a sector — see survey.Survey.Subsector,
// which selects a subsector from a surveyed sector instead.
func GenerateSector(r *dice.Roller, d Density) []StellarHex {
	var systems []StellarHex

	for col := 1; col <= Columns; col++ {
		for row := 1; row <= Rows; row++ {
			if SystemPresent(r, d) {
				systems = append(systems, rollContents(r, Hex{col, row}))
			}
		}
	}

	return systems
}
```

This is the premise the whole `survey` design rests on: a subsector rolled _on
its own_ would consume a different run of dice, so the same coordinates would
hold different worlds. Every CLI view therefore **selects from** one whole
survey rather than generating a region directly.

### `survey.Sector` — the composition

Twenty-odd lines that name the entire mapping tier:

```bash
sed -n '62,86p' internal/survey/survey.go
```

```output
func Sector(r *dice.Roller, d sectorgen.Density) Survey {
	hexes := sectorgen.GenerateSector(r, d)

	records := make([]Record, len(hexes))
	for i, h := range hexes {
		records[i] = Record{
			Hex:    h.Hex,
			Name:   worldName(r),
			System: systemgen.GenerateForMap(r, h.GasGiant, h.AsteroidMainworld),
		}
	}
	// Capitals and Naval Depots are placed from base Importance, then routes, then
	// Way Stations (which bump Importance). Depots rank before Way Stations so they
	// see the same base Importance the capitals did — a depot and a capital agree on
	// which world is most Important, rather than the depot ranking on a +1 a Way
	// Station happened to add first. A Way Station's +1 does not re-trigger route,
	// capital, or depot selection.
	markSectorCapitals(records)
	placeNavalDepots(records)
	links := route.Build(worldsOf(records), route.DefaultJump)
	placeWayStations(records, links)

	return Survey{Records: records, Routes: links}
}

```

The coarse map's per-hex flags are handed to `systemgen.GenerateForMap` — this
is the constraint plumbing from the previous section, so the quick preview and
the detailed system always agree about what sits in a hex.

The pass ordering is the single most important decision here, and it is stated
as prose rather than left implicit. Capitals and depots rank on **base**
Importance; way stations run last precisely because their `+1` must not
retroactively change who won capital or depot selection.

### `internal/route` — trade routes

`route` takes nothing but hexes and Importance values, and returns links. No
dice at all. `Build` sorts by CCRR immediately, so adjacency order — and
therefore BFS tie-breaking — is deterministic:

```bash
sed -n '43,52p' internal/route/route.go
```

```output
func Build(worlds []World, maxJump int) []Link { //nolint:gocognit,cyclop // trade-route graph; irreducibly branchy
	if maxJump <= 0 {
		maxJump = DefaultJump
	}

	// Work in a stable CCRR order so adjacency (and thus BFS) is deterministic.
	sorted := append([]World(nil), worlds...)
	sort.Slice(sorted, func(i, j int) bool { return before(sorted[i].Hex, sorted[j].Hex) })

	// adj[i] lists the indices within maxJump of world i, in ascending (CCRR)
```

Undirected links are canonicalized on the way into the set, which is what lets
a plain `map[Link]bool` deduplicate regardless of which endpoint discovered the
link:

```bash
sed -n '158,167p' internal/route/route.go
```

```output
// orderedLink builds a Link with its endpoints in CCRR order so undirected links
// de-duplicate regardless of which end they were found from.
func orderedLink(a, b sectorgen.Hex) Link {
	if before(b, a) {
		a, b = b, a
	}

	return Link{From: a, To: b, Jump: a.Distance(b)}
}

```

And the package's boundary is stated in `ExpectedTraffic`, which enumerates
every adjustment the book prints and then hands all of them to the caller —
because they need dice, and `route` does not roll:

```bash
sed -n '169,194p' internal/route/route.go
```

```output
// the given Importance (Book 3 p.27 Expected Ship Traffic table), clamped to the
// table's [-3, +5] range. All three adjustments printed under the table are the
// caller's: "Plus Flux" — actual arrivals are the returned value plus a Flux
// roll — and the two empire shifts ("For a Busy Empire, next row higher. For a
// Rural Empire, next row lower.", ±1 row). This package is deterministic and
// rolls no dice, so it returns the bare row value.
func ExpectedTraffic(ix int) int {
	switch {
	case ix >= 5:
		return 1000
	case ix == 4:
		return 100
	case ix == 3:
		return 30
	case ix == 2:
		return 20
	case ix == 1:
		return 10
	case ix == 0:
		return 2
	case ix == -1:
		return 1
	default:
		return 0
	}
}
```

### The views

`Survey.Subsector` is a **filter**, not a generator — the concrete payoff of
"survey whole, select part":

```bash
sed -n '106,124p' internal/survey/survey.go
```

```output
//
// A subsector with no star system selects nothing, and so does a letter outside
// A-P — callers taking a letter from a user should reject it first with
// sectorgen.ParseSubsector rather than read an empty result as an empty region.
func (s Survey) Subsector(letter byte) []Record {
	if letter >= 'a' && letter <= 'z' {
		letter -= 'a' - 'A'
	}

	var out []Record

	for _, rec := range s.Records {
		if rec.Hex.Subsector() == letter {
			out = append(out, rec)
		}
	}

	return out
}
```

Three views select from one survey. Here is the default subsector listing, six
worlds' worth:

```bash
go run ./cmd/sectorgen -seed 42 -subsector A 2>/dev/null | head -6
```

```output
0101 Maesavo E410100-7 Lo Co {-3}(500-2)[1139] B - - 233 17 Im K8 V BD K4 VI
0104 Helo B120557-F De Ni Po Ho {+1}(E42+1)[566J] B - - 820 10 Im G5 V K5 VI
0107 Magi A200440-E Va Ni Co Tz {+1}(F32+2)[653J] B - - 932 14 Im M2 V BD
0108 Hipa B978003-B Tz {+1}(200+0)[0000] B N - 000 6 Im K D M9 VI M2 VI
0109 Cutuhy B7997B8-7 Sa Pi Co {0}(86B-2)[A773] BD - - 212 8 Im F7 II G5 V M0 V
0110 Bijeri D110359-8 Lo {-3}(D20-2)[214C] B - - 131 15 Im F3 IV K0 VI K D F8 VI
```

And the deep per-system sheet for one hex — the detail a one-line survey record
cannot carry:

And the deep per-system sheet for one hex — the detail a one-line survey record
cannot carry. Note that the renderer only _lays out_; every piece of meaning is
delegated back to the package that owns it (`worldgen.OrderTradeCodes`,
`route.ExpectedTraffic`, `mw.Extensions()`):

```bash
go run ./cmd/sectorgen -seed 42 -hex 0109 2>/dev/null | head -24
```

```output
0109  Cutuhy
────────────────────────────────────────────────────────────────
  Mainworld   B7997B8-7  Sa Pi Co
  Extensions  {0}(86B-2)[A773]   RU -1056
  Traffic     ~2 ships/week
  Nobility    BD
  Bases       none
  Travel Zone Green
  Natives     Natives
  Starport    B — Good
              builds Spacecraft · repairs: Overhaul · fuel: Refined+Unrefined (2D hours) · downport

  Stars
    Primary            F7 II
    Close              G5 V  (Orbit 3)
    Far                M0 V  (Orbit 17)

  Orbits — 8 worlds · PBG 212 · 2 gas giants · 1 belt
    [Primary]
      2  Inferno        YBB0000-0  He
      9  Belt
   * 10  Mainworld      B7997B8-7  — moon of Gas Giant T LGG
          · sibling moon Ee    Storm World  H686664-1 Ga Lk Ni Ag Ri Co Tu  (close orbit)
          · sibling moon Tee   Inferno      YCB0000-0 He Sa Co  (far orbit)
```

---

## 7. `internal/chargen` — character creation

The biggest package in the repo (~4700 lines), and the one that most resembles
an engine with pluggable data rather than a transcription.

A character starts as six 2D rolls — the UPP, Strength/Dexterity/Endurance/
Intelligence/Education/Social — at age 18:

```bash
sed -n '90,99p' internal/chargen/chargen.go
```

```output
// Generate rolls a character's six characteristics, each 2D, in UPP order, at
// the starting age.
func Generate(r *dice.Roller) Character {
	c := Character{Age: startingAge}
	for i := range c.scores {
		c.scores[i] = r.Dice(2)
	}

	return c
}
```

`GenerateCareered` runs Book 1's full checklist, and reads almost literally as
the book's numbered steps: homeworld skills, education, then careers until the
policy stops offering them, with a Citizen fallback for a character every career
refuses:

```bash
sed -n '447,476p' internal/chargen/career.go
```

```output
func GenerateCareered(r *dice.Roller, p Policy, homeworld worldgen.World, career Career) Character {
	c := Generate(r)
	c.Homeworld = homeworld
	ApplyHomeworldSkills(&c, homeworld, p)

	if p.PursueEducation(c) {
		educate(r, p, &c)
	}

	entered := serveCareer(r, p, &c, career)
	// A character may serve more than one career (Book 1), so long as they live.
	for !c.Dead {
		next, ok := p.NextCareer(c)
		if !ok {
			break
		}

		if serveCareer(r, p, &c, next) {
			entered = true
		}
	}

	if !entered && !c.Dead {
		serveCareer(r, p, &c, CitizenCareer) // fall back to the auto-begin Citizen life
	}

	computeEntitlements(&c)

	return c
}
```

### Careers are data, not subclasses

All thirteen careers are values of one `Career` struct. Rule variants that a
career needs are `bool` flags on that struct, and `runTerm` dispatches on them —
there is no interface hierarchy:

```bash
sed -n '302,325p' internal/chargen/career.go
```

```output
type Career struct {
	ID               CareerID
	Name             string
	Qualify          Qualification
	CCMode           CCMode
	ControllingChars []Characteristic
	Continue         ContinueRule
	EligPerTerm      int        // number of skill rolls a surviving term grants
	BenefitDM        MusterDM   // die modifier the muster Benefit column adds (Money always adds +Terms)
	AutoBegin        bool       // the career is entered automatically, with no qualify roll (Citizen)
	CitizenLife      bool       // the term uses benign Citizen Life instead of Risk & Reward (Citizen)
	FameCareer       bool       // the term resolves Fame/Talent instead of Risk & Reward (Entertainer)
	Masterpiece      bool       // the term attempts a Masterpiece instead of Risk & Reward (Craftsman)
	OfficePolitics   bool       // the term resolves Office Politics instead of Risk & Reward (Functionary)
	ReturnIntrigue   bool       // the term resolves Return & Intrigue instead of Risk & Reward (Noble)
	ScoutDuty        bool       // the term picks Courier (no R&R, 4 skills) or Explorer (R&R, EligPerTerm skills) (Scout)
	SchemeCareer     bool       // the term masterminds a Rogue Scheme instead of Risk & Reward (Rogue)
	AutoFailOn12     bool       // a natural 12 always fails, whatever the target (Rogue; see autoFails)
	UndercoverCareer bool       // the term runs an Undercover Assignment alongside Risk & Reward (Agent)
	RewardKind       RewardKind // what a successful Reward roll earns
	Skills           SkillGrid
	MusterOut        MusterTable

	// AutoBenefits are the mustering-out awards a career grants unconditionally,
```

The other important structural idea is the split between durable history and
transient scratch. `CareerRecord` is what the character keeps; `careerRun` is
per-career working state that is deliberately kept _out_ of `Character`, so a
character carries no half-finished career state between careers.

The term loop is the engine's heartbeat — term, age four years, aging check,
outcome dispatch, continue:

```bash
sed -n '597,626p' internal/chargen/career.go
```

```output
	for {
		run.terms = rec.Terms // terms served before this one, for the Rogue's "Mod +Terms"
		outcome := runTerm(r, p, c, run, career)
		rec.Terms++
		c.Age += termYears
		AgingCheck(r, c) // no-op before age 34; may set c.Dead

		if outcome == Died || c.Dead {
			rec.Outcome = Died

			break
		}

		if outcome == Disabled {
			rec.Outcome = Disabled

			break
		}

		if outcome == MusteredOut {
			rec.Outcome = MusteredOut // a term forced the career to end (Office Politics job loss)

			break
		}

		// Keep the record's Branch in step with the run before any policy hook reads
		// it. rec is passed by value to RerollBranch and Continue, and assigning it
		// only at muster-out (below) left both seeing "" for the whole career — so a
		// policy meaning "keep Flight" rerolled every term, drawing an extra die and
		// landing the character somewhere he never chose.
```

That comment is documenting a live bug that was fixed: because `rec` is passed
_by value_ to the policy hooks, assigning `rec.Branch` only at muster-out left
both hooks seeing `""` for the whole career — so a policy meaning "keep Flight"
rerolled every term, drew an extra die, and landed the character somewhere they
never chose.

### The Policy seam

Everywhere the rules say "the player chooses", `chargen` asks a `Policy`. That
is what keeps the engine deterministic and testable: choices come from the
policy, dice come from the roller, and the two never mix.

`ApplyHomeworldSkills` is the cleanest example — it grants one skill per
homeworld trade classification and **rolls nothing at all**, so a homeworld
composes in without perturbing any dice sequence:

```bash
sed -n '66,84p' internal/chargen/homeworld.go
```

```output
// ApplyHomeworldSkills grants a character their homeworld skills: one per Trade
// Classification of the world. Industrial (In) and Rich (Ri) worlds grant a
// player-chosen One Trade or One Art; the rest are fixed. No dice are rolled —
// choices come from the policy — so a homeworld composes into generation without
// disturbing any dice sequence.
func ApplyHomeworldSkills(c *Character, world worldgen.World, p Policy) {
	for _, code := range world.TradeCodes {
		switch code {
		case "In":
			c.Skills.Raise(p.ChooseSkill(*c, theTrades), 1)
		case "Ri":
			c.Skills.Raise(p.ChooseSkill(*c, oneArt), 1)
		default:
			for _, s := range homeworldSkill[code] {
				c.Skills.Raise(s, 1)
			}
		}
	}
}
```

### `internal/skill` — the cascade rule

Skills are a pure inventory: two maps, one for plain skills and cascade parents,
one for the knowledges hanging under a parent. The two stack when a task is
performed — Engineer-1 plus J-Drive-1 resolves J-Drive tasks at 2.

The signature rule is the career progression, and it fits in seven lines:

```bash
sed -n '155,167p' internal/skill/skill.go
```

```output

// GrantCascade applies one career award of a cascade parent's knowledge, per the
// Knowledge-Knowledge-Skill progression: the first two awards raise the
// knowledge (to 1, then 2), and further awards raise the parent skill (1, 2, …).
// A player who would rather keep advancing the knowledge past 2 calls
// RaiseKnowledge directly.
func (s *Set) GrantCascade(parent, knowledge string) {
	if s.KnowledgeLevel(parent, knowledge) < 2 {
		s.RaiseKnowledge(parent, knowledge, 1)
	} else {
		s.Raise(parent, 1)
	}
}
```

---

## 8. `internal/sophont` — making chargen work for aliens

The Sophont Creation System builds a _species_, not an individual. A `Species`
is essentially the book's "Sophont Creation Card": homeworld, environment, six
characteristic **specs**, size, life cycle, gender, and optionally caste.

The linchpin is that a species stores **how many dice each characteristic rolls**
— not a value. That is what makes the bridge to `chargen` a one-line
substitution.

```bash
sed -n '44,65p' internal/sophont/sophont.go
```

```output
func Generate(r *dice.Roller) Species {
	home := plausibleHomeworld(r)
	env := rollEnvironment(r, home.Profile)
	chars, gp := rollCharacteristics(r, env)
	gender := rollGender(r, chars[4].Name == Ins)

	s := Species{
		Homeworld:      home,
		Environment:    env,
		Chars:          chars,
		GeneticProfile: gp,
		Size:           Size(chars),
		LifeCycle:      rollLifeCycle(r),
		Gender:         gender,
	}
	if chars[5].Name == Cas { // C6 is Caste
		caste := rollCaste(r, gender)
		s.Caste = &caste
	}

	return s
}
```

Note `plausibleHomeworld`: rather than write a parallel world generator for
sophont cradles, it rerolls `worldgen` until it gets an Atmosphere 2-9,
Population 7+ world. Reuse over reimplementation, at the cost of some dice.

Large dice pools would run away, so Book 3's "Rolling Higher Characteristics"
rule caps them — 6D becomes 12+4D:

```bash
sed -n '188,198p' internal/sophont/characteristics.go
```

```output
// RollValue rolls a characteristic value for the given die count, applying the
// "Rolling Higher Characteristics" rule (Book 3 p.228): 1-3 dice roll normally;
// 4+ dice roll 12 + (dice-2)D so a large pool stays bounded (6D = 12+4D). Used
// when an individual of the species is generated.
func RollValue(r *dice.Roller, numDice int) int {
	if numDice <= 3 {
		return r.Dice(numDice)
	}

	return 12 + r.Dice(numDice-2)
}
```

### The bridge

`chargen.GenerateSophont` substitutes for step 1 of character creation only. It
rolls the species' own dice counts instead of six flat 2D, then rolls gender and
caste:

```bash
sed -n '28,43p' internal/chargen/sophont.go
```

```output
func GenerateSophont(r *dice.Roller, species sophont.Species) Character {
	c := Character{Age: startingAge, Homeworld: species.Homeworld}
	// A slot with 0 dice carries no characteristic value — the Caste C6 of a
	// caste species, whose caste comes from the table roll below. RollValue rolls
	// nothing for it, leaving the score at its 0 sentinel and the stream intact.
	for i := range c.scores {
		c.scores[i] = sophont.RollValue(r, species.Chars[i].Dice)
	}

	c.Gender = rollTrait(r, &c, species.Gender.Table, species.Gender.Differences)
	if species.Caste != nil {
		c.Caste = rollTrait(r, &c, species.Caste.Table, species.Caste.Differences)
	}

	return c
}
```

Two subtleties worth noticing. A **0-dice slot** — the Caste C6 of a caste
species, whose value comes from the caste table rather than a roll — draws
nothing, leaving the score at its 0 sentinel _and the dice stream intact_. And
because the six characteristic rolls come first in the same order, a
Human-equivalent species reproduces plain `Generate` exactly.

One downstream consequence did land: sophont scores are uncapped, so
`Character.UPP` had to become total and render out-of-range values as `?` rather
than panicking.

Running the character generator — a bare UPP, then a full career:

```bash
go run ./cmd/chargen -seed 3 2>/dev/null; echo '--'; go run ./cmd/chargen -seed 3 -career scout,merchant 2>/dev/null
```

```output
7763B6
--
Scout, then Merchant — age 37
  UPP            965778
  Homeworld      C532954-9   Hi Na Po
  Education      Psychology (major)
  Service        Scout: 3 terms, mustered out
                 Merchant: 1 term, mustered out as Fourth Officer
  Skills         Actor-1, Animals-1, Artist-1, Astrogation-1, Athlete-2, Biologics-1, Biologist-1, Broker-1, Comms-2, Craftsman-1, Diplomat-1, Gunner-3, JOT-1, Language-0, Language/Galanglic-1, Medic-2, Pilot-1, Psychology-1, Steward-2, Streetwise-2, Survey-2, Survival-1, Teacher-1, Trader-1, Vacc Suit-3
  Reputation     Fame 2
  Discoveries    2
  Land grants    2
  Ship shares    1
  Wound badges   1
  Benefits       Forbidden Knowledge, Forbidden Knowledge

```

---

## 9. `internal/shipgen` — ship design, deterministic

The starship tier breaks the pattern in one respect: **design rolls no dice**.
`Design(spec) Ship` is a total function — spec in, ship out, no errors, no
randomness. Book 2's Adventure Class Ship rules are arithmetic, not tables.

```bash
sed -n '182,200p' internal/shipgen/shipgen.go
```

```output
type ShipSpec struct {
	Name        string // display name, e.g. "Murphy-class Scout"
	Mission     string // QSP mission code, e.g. "S"
	TL          int
	HullLetter  int // ordinal 1..24
	Tons        int // 0 = nominal (HullLetter*100)
	Config      Config
	Structure   Structure
	ArmorLayers int

	Maneuver *DriveSpec
	Jump     *DriveSpec
	Power    *DriveSpec

	Weapons  []WeaponSpec
	Defenses []DefenseSpec

	FuelScoop, FuelPurifier bool
}
```

`Design` walks the book's checklist: hull → spec validation → drives → feasibility
cross-checks → fuel (which needs hull tons, so it runs after the drives) → armor
→ armament → mount points → tonnage → budget → cost.

The package's central discovery is that the book's irregular-looking drive
Potential grid (Table Z1) is actually one line of arithmetic:

```bash
sed -n '38,44p' internal/shipgen/drive.go
```

```output
func drivePotential(driveOrd, hullOrd, effPct int) int {
	if driveOrd < 1 || hullOrd < 1 {
		return 0
	}

	return min(2*driveOrd*effPct/(hullOrd*100), 9)
}
```

Because `Design` is total, infeasibility is **reported, not refused**. A ship
with drives and no power plant still gets built; it just carries a problem:

```bash
sed -n '123,135p' internal/shipgen/design.go
```

```output
	switch {
	case powered > 0 && ship.Power == nil:
		problems = append(problems, "drives require a power plant")
	case ship.Power != nil && ship.Power.Potential < powered:
		problems = append(
			problems,
			fmt.Sprintf(
				"power plant potential %d is below the drives it feeds (%d)",
				ship.Power.Potential,
				powered,
			),
		)
	}
```

That creates an accounting hazard, and one tiny function is the answer to it.
A component that failed to design must claim no mount point and spend no
tonnage — otherwise the ship gets a second, false complaint on top of the real
one. All three accountings route through it, so they cannot drift apart:

```bash
sed -n '20,29p' internal/shipgen/mounts.go
```

```output
// aboard reports whether a designed component is actually carried. One that failed
// to design is not: it claims no mount point and it spends no tonnage. Counting it
// either way would invent a second complaint on top of the real problem that
// stopped it being built — two complaints for one mistake, one of them false.
//
// All three accountings of the same component list go through this — mount points,
// tonnage, and cost — so they cannot drift into disagreeing about which components
// exist. Defense.installed asks it too, so a refused component does not render a
// full line of tonnage and cost for something the ship does not carry.
func aboard(problems []string) bool { return len(problems) == 0 }
```

### One rule, two builders

Weapons and defenses scale by the same rule, so it lives in exactly one place.
The **stage** shifts TL and multiplies the _device_ cost; the **range** shifts TL
again and scales the _mount's_ tonnage and cost. The two builders differ in what
they do with the resulting Mod and what they refuse to build — not in this
arithmetic:

```bash
sed -n '339,351p' internal/shipgen/weapon.go
```

```output
func install(
	deviceTL, deviceCost, mountTons, mountCost int,
	rng Range,
	stage Stage,
) (int, Tonnage, int, int) {
	r := rangeData[rng]
	st := stageCostData[stageIndex(stage)]

	return deviceTL + r.tlMod + stageTL(stage),
		Tonnage(mountTons * r.tons),
		deviceCost*st.num/st.den + mountCost*r.cost/100,
		r.band
}
```

Let's design one:

```bash
go run ./cmd/shipgen -hull A -tl 12 -config L -structure shell -maneuver A -jump A -weapon beamlaser:T1:orbit -defense blackglobe 2>/dev/null
```

```output
Ship  X-AL22
Hull:    A 100t Lifting Body · TL-12 · Shell
Drives:  Maneuver-A 2G · Jump-A J-2 · Power-A
Armor:   1 layers AV-6
Weapons: Standard Orbit Single Turret Beam Laser-11 Mod=+0. 2 tons. MCr1.1. Hits= 1D. R=08.
Defenses: Standard Vdistant Bolt-In Black Globe-16 Mod=+3. 3 tons. MCr13. R=07. (Electronic).
Fuel:    22t · Cost MCr 48.2 · Payload 56t
! Black Globe-16 is TL-16, above the ship's TL-12
```

---

## 10. `internal/trade` — the pricing engine

Trade is two independent quantities joined by a die roll. **Cost** at the source
is `3000 + TL*100 + Σ modifiers`. **Price** at the market is a base plus
Cr1,000 for every source trade class the market matches, scaled 10% per Tech
Level the source exceeds the market:

```bash
sed -n '118,134p' internal/trade/trade.go
```

```output
func Price(sourceTL, marketTL int, sourceTCs, marketTCs []string) int {
	p := basePrice
	// A class with no priceMatch row yields the zero value, whose nil markets make
	// the inner loop a no-op — no presence check needed.
	for _, code := range ValueClasses(sourceTCs) {
		m := priceMatch[code]
		for _, mc := range m.markets {
			if slices.Contains(marketTCs, mc) {
				p += m.per
			}
		}
	}

	p += (sourceTL - marketTL) * p / 10

	return max(p, 0)
}
```

Only then do dice enter, and they enter exactly once. Note that the broker's DM
folds into the Flux **before** the table lookup, not after:

```bash
sed -n '139,149p' internal/trade/trade.go
```

```output
func ActualValuePercent(flux, brokerSkill int) int {
	e := clamp(flux+BrokerDM(brokerSkill), -5, 8)

	return actualValue[e+5]
}

// SellingPrice is the realized per-ton sale price: the expected Price times the
// Actual Value percentage (Book 2 p.221).
func SellingPrice(price, flux, brokerSkill int) int {
	return price * ActualValuePercent(flux, brokerSkill) / 100
}
```

That sets up the engine's central trade-off: a better broker both raises the
Actual Value roll and takes a bigger cut, at 5% per point of DM.

```bash
sed -n '12,20p' internal/trade/brokers.go
```

```output
func BrokerDM(brokerSkill int) int {
	return min(max((brokerSkill+1)/2, 0), 4)
}

// BrokerCommissionPercent is the cut a broker of the given skill takes from a
// sale (Book 2 p.220 Brokers): 5% per point of Broker DM.
func BrokerCommissionPercent(brokerSkill int) int {
	return BrokerDM(brokerSkill) * 5
}
```

`trade` is the one tier not yet wired to a CLI. It connects to `shipgen` only
conceptually: the tonnage its freight and passages would fill is
`Ship.Tonnage.Payload`, the residual the design engine deliberately leaves
unmodelled. That is the seam where a future campaign layer joins them.

---

## 11. `internal/shipcombat` — where the packages meet

Space combat is two deliberately-stacked layers. The **lower layer** is
primitives over plain ints, because that is what the book's tables are: to-hit,
defensive fire, hit location, penetration, damage. Armor is a slice of per-layer
values that damage grinds through in turn:

```bash
sed -n '26,36p' internal/shipcombat/damage.go
```

```output
func Penetrate(damage int, layerAV []int) (bool, int) {
	for _, av := range layerAV {
		if damage < av {
			return false, 0
		}

		damage -= av
	}

	return damage >= 1, damage
}
```

The **upper layer** is the bridge — the only place `shipcombat` imports
`shipgen`, and the thesis of the whole starship tier: _a generated ship can
fight._ `Attack` is three lines, because a designed weapon already carries its
own TL and Mod; the gunner supplies the rest:

```bash
sed -n '28,30p' internal/shipcombat/ship.go
```

```output
func Attack(r *dice.Roller, w shipgen.Weapon, rangeBands, csk int, mods ...int) dice.CheckResult {
	return ResolveSpaceWeapon(r, rangeBands, w.TL, csk, append([]int{w.Mod}, mods...)...)
}
```

`Defend` is where the two packages have to agree about something subtle. A
defense the designer _refused_ to build is not aboard — `shipgen` charged the
ship nothing for it and spent no tonnage. It must not then intercept anything,
or a refused screen becomes a free one:

```bash
sed -n '170,190p' internal/shipcombat/ship.go
```

```output
// A defense the designer refused is not aboard: shipgen already charges the ship
// nothing for it, spends no tonnage on it, and renders it as a bare name. It must
// not then intercept anything, or a refused screen becomes a free one — the ship
// gets the protection precisely because it never paid for it.
func Defend(r *dice.Roller, d shipgen.Defense, attackTL int) dice.CheckResult {
	if !d.Installed() {
		// No hardware, no interception. Resolve against an impossible target so the
		// result reads as a clean miss rather than a special case for every caller.
		return ResolveDefensiveFire(r, 0, attackTL, noDefenseMod)
	}

	return ResolveDefensiveFire(r, d.TL, attackTL, d.Mod)
}

// noDefenseMod is the modifier for defensive fire that has nothing to fire with:
// large enough that no roll succeeds, so an absent defense never intercepts.
const noDefenseMod = -100

// ArmorLayers is a designed ship's armor as Penetrate wants it: one Armor Value
// per layer (Book 2 p.75). Damage grinds through the layers in turn, so the shape
// matters — four layers of AV-6 are not one layer of AV-24.
```

And `ArmorLayers` carries a fixed bug in its comment, which is worth quoting
because it shows exactly what "the layer shape matters" means:

```bash
sed -n '192,200p' internal/shipcombat/ship.go
```

```output
// Armor.AV is already the per-layer value: "identical layers are not summed on the
// record" (shipgen/fuelarmor.go). Dividing it by the layer count — which this once
// did — armored every multi-layer ship at a fraction of its real protection, and
// the more layers it bought the thinner each one became.
func ArmorLayers(s shipgen.Ship) []int {
	if s.Armor.Layers <= 0 {
		return nil
	}

```

---

## 12. `internal/cli` — the command contract

All six commands follow one convention, and it exists so a command's stdout can
be piped as data:

- **Generated records go to stdout. Everything else goes to stderr.**
- Bad input is `cli.Fatalf` — exit 2, the code `flag` itself uses.
- A true-but-empty result is `cli.Notef` — exit 0, still off stdout, so a piped
  record stream stays clean.

A command that printed "unknown density" onto its record stream and exited 0
would be indistinguishable from one that worked.

### Every run is reproducible

`cli.Roller` draws the fresh seed itself when `-seed` is omitted and reports it
on stderr, so a run worth keeping can always be replayed. (The seed is different
every run, so it is masked here to keep this document verifiable.)

```bash
go run ./cmd/worldgen -n 2 2>&1 >/dev/null | sed 's/seed [0-9]*/seed <freshly drawn>/'
```

```output
worldgen: seed <freshly drawn>
```

The subtle part is _when_ the seed is reported. `Roller` returns a `reportSeed`
function rather than printing at construction, because `Roller` only parses —
every per-command check (an unknown density, a bad hull letter) happens after it
returns. Printing at construction announced a seed and then died on bad input,
naming a run that generated nothing:

```bash
sed -n '133,150p' internal/cli/cli.go
```

```output
	// decided here; deferring only WHEN it is said, not what.
	//
	// The guard is the idempotence the doc promises, and it is load-bearing:
	// chargen and shipgen each call reportSeed from two places, and today only
	// avoid saying the seed twice because control flow happens to return between
	// them. A command that grows a third path should not have to know that.
	return r, func() {
		if reported {
			return
		}

		reported = true

		Notef(seedNote, fresh)
	}
}

// SeededRoller defines the shared -n and -seed flags (naming the item in the -n
```

### A flag a path cannot honor is bad input

`RejectUnusable` is the other half of the contract. `shipgen` without `-hull`
rolls a random ship that reads none of the design flags, so `-tl 99` used to
produce a TL-14 ship at exit 0. Worse, `sectorgen`'s `-hex` view ignored
`-subsector`, so `-hex 0436 -subsector Q` printed a hex at exit 0 — while that
same `-subsector Q` is rejected outright on the default path.

```bash
sed -n '186,208p' internal/cli/cli.go
```

```output
// It asks flag.Visit — which walks only what was actually set — rather than
// comparing values against defaults, because most defaults are legal values a
// caller may well type. "-tl 0" is a real Tech Level and "-armor 1" is the hull's
// integral layer; a value test would read both as unset, and "-tl 0" alongside a
// random ship is exactly the discarded input this exists to catch.
//
// Call it before reporting the seed, so a run rejected here never names a seed
// for records it did not generate.
func RejectUnusable(what string, usable ...string) {
	var unusable []string

	flag.Visit(func(f *flag.Flag) {
		if slices.Contains(usable, f.Name) || slices.Contains(sharedFlags, f.Name) {
			return
		}

		// A bool explicitly set false asks for its path NOT to be taken, which is
		// agreement rather than conflict: "-hex 0436 -sector=false" is an unambiguous
		// command line, and a script building "-sector=$want" should not die on it.
		// Only a bool set TRUE competes for the path.
		if bv, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && bv.IsBoolFlag() &&
			f.Value.String() == "false" {
			return
```

Two details. "Was it set" is asked of `flag.Visit`, never of the value — `-tl 0`
is a real Tech Level and `-armor 1` the hull's integral layer, so a
default-comparison would wave real input through as unset. And a bool explicitly
set _false_ asks for its path **not** to be taken, which is agreement rather
than conflict, so a script building `-sector=$want` doesn't die on it.

The naming runs the useful way round: the caller lists the flags its path
_reads_, so a flag added later is covered the day it is added rather than the
day someone remembers to list it.

```bash
go build -o /tmp/sectorgen-demo ./cmd/sectorgen && /tmp/sectorgen-demo -hex 0436 -subsector Q; echo "exit $?"
```

```output
sectorgen-demo: -subsector cannot be combined with -hex
exit 2
```

---

## 13. `internal/clitest` — testing the contract end-to-end

`main` calls `os.Exit`, so a command cannot be exercised inside the test
process. Each case therefore **re-executes the test binary as a child** that
runs `main` instead of the tests. All six commands need the same child, so it
lives here once:

```bash
sed -n '57,66p' internal/clitest/clitest.go
```

```output
func (c Command) TestMain(m *testing.M) {
	if os.Getenv(childEnv) == "1" {
		c.runChild(os.Args[1:]) // never returns
	}

	os.Exit(m.Run())
}

// Run executes the command in a subprocess with the given command line and
// returns its streams and exit code. No -seed is passed unless the caller passes
```

The interception happens before `m.Run`, which is what frees argv: nothing has
parsed `os.Args` yet, so the child's command line rides it and the environment
variable is only a marker. Passing args out of band would have needed an
encoding with a separator that cannot occur inside a real flag value — which
rules out space and comma (several commands take list-valued flags like
`-weapon beamlaser:T1:orbit,sandcaster`) and NUL (`os/exec` refuses it).

A command wires it up in two lines and then asserts the whole contract in one
call:

    var shipgen = clitest.Command{Name: "shipgen", Main: main}

    func TestMain(m *testing.M) { shipgen.TestMain(m) }

    // ...
    shipgen.Run(t, "-tl", "-5").AssertRejected(t)

### Assertions that fail closed

This is the most carefully-argued code in the repo, and the reason is that
`AssertRejected` makes a **negative** assertion — "this run named no seed" —
which is the shape that fails open. A matcher that drifted off the real wording
would stop matching and wave a leaked seed through.

The answer is that `cli` owns the wording, and the matcher is _built from_ the
format string the writer uses, so the two cannot drift apart:

```bash
sed -n '20,26p' internal/cli/cli.go; sed -n '45,47p' internal/cli/cli.go
```

```output
// FailureCode is the exit status for bad input. It is flag's own usage-error
// code, so a rejected flag value and a rejected flag both exit the same way.
const FailureCode = 2

// seedNote is the Notef format Roller reports a drawn seed with, and the single
// definition of what a seed line says. Nothing else spells the wording out:
// seedLine below is BUILT from this constant, so the seed WORDING cannot drift
// loudly on the positive side rather than quietly on the negative one.
var seedLine = regexp.MustCompile(`(?m)^[^\s:]+: ` +
	strings.ReplaceAll(regexp.QuoteMeta(seedNote), `%d`, `(\d+)`) + `$`)
```

The seed line is also matched as a **whole line** (`^cmd: seed \d+$`) rather
than as the substring `"seed "`, because `flag`'s own usage text describes the
`-seed` flag and would otherwise read as a seed report on every run `flag`
rejects.

`AssertRejected` also checks for a panic, and the comment explains why that
clause is not redundant:

```bash
sed -n '145,157p' internal/clitest/clitest.go
```

```output
		t.Errorf("named a seed for a run that generated nothing: %q", r.Stderr)
	}

	// An unrecovered panic ALSO exits 2, writes to stderr, writes nothing to
	// stdout, and names no seed — so it satisfies every clause above and a crash
	// on a bad-input path would read as a clean rejection. That is the fail-open
	// shape this repo has now found three times; here it would hide the very
	// crashes these tables exist to catch.
	if strings.Contains(r.Stderr, "panic:") {
		t.Errorf("child panicked rather than rejecting the input: %q", r.Stderr)
	}
}

```

And `AssertReportedSeed` deliberately runs **two** stdout leak checks that look
redundant and are not — they fail in opposite directions, and neither subsumes
the other:

```bash
sed -n '172,186p' internal/clitest/clitest.go
```

```output
	if strings.TrimSpace(r.Stdout) == "" {
		t.Error("nothing on stdout; a successful run must print its records there")
	}

	// Two checks, because they fail in opposite directions and neither subsumes
	// the other. cli.HasSeedReport is the canonical Notef line, routed through the
	// owner of the wording so a reword cannot leave a stale matcher behind. The
	// broad substring catches a seed reaching stdout by any OTHER path — a record
	// header like `fmt.Println("# seed", n)`, or a sheet field — which the anchored
	// form cannot see.
	//
	// Replacing the substring with the anchored matcher alone was a mistake made in
	// this very wave: it closed a drift hole and opened a coverage hole, in the one
	// assertion whose whole job is catching a leak onto the record stream.
	if cli.HasSeedReport(r.Stdout) || strings.Contains(strings.ToLower(r.Stdout), "seed") {
```

The anchored matcher cannot see a leak that arrives by some other path (a header
like `fmt.Println(\"# seed\", n)`, a sheet field); the substring cannot survive a
reword. Collapsing them to the anchored form alone looks like a tidy-up and is a
coverage regression — a mistake that shipped once and was caught in review.

---

## 14. Running it

The whole thing is driven by `task` (go-task):

```bash
task --list
```

```output
task: Available tasks for this project:
* audit:         Report what the globally-disabled linters would flag — non-gating, for periodic review of the blind spots those exceptions create
* build:         Build all packages
* check:         Lint (incl. format + vet) and tests — the pre-commit gate
* cover:         Run the tests with a coverage summary
* default:       Run the tests
* deps:          Install tooling from the Brewfile
* extract:       Extract the local rulebook PDFs to text reference (git-ignored, not distributed)
* fmt:           Format the code (gofumpt + goimports + golines)
* lint:          Lint the code
* test:          Run the tests
* tidy:          Tidy the module files
* vet:           Vet the code
```

`task check` is the pre-commit gate: `golangci-lint run` (which subsumes format
and vet) plus `go test ./...`.

The test suite is large and its pass count is a reasonable proxy for coverage of
the transcribed rules:

```bash
grep -rc '^func Test' --include='*_test.go' . | awk -F: '{s+=$2} END {print s" test functions across the repo"}'
```

```output
599 test functions across the repo
```

---

## Closing notes — what to carry away

Five ideas recur at every tier, and knowing them makes unfamiliar code in this
repo read quickly:

1. **Pure formula + rolling composer.** Rules become functions of their die
   rolls; a `Generate` composes them in the book's checklist order. Test the
   formulas at their edges, lock the composer with a golden test from a printed
   worked example.

2. **The dice stream is an observable.** Same seed, same records — so a
   constraint supersedes a roll rather than skipping it, and a generator that
   changes what it draws is a breaking change. `dice.NewScripted` panics rather
   than wrapping precisely so that a green suite is _evidence_ about the stream.

3. **Strict paths panic, display paths render `?`.** `ehex.Digit` vs
   `ehex.Format`, `Difficulty.Dice` vs `Difficulty.String`, `Hex.Subsector` vs
   `Hex.String`. The rule is whether a wrong-but-plausible value would hide a bug
   or merely look odd.

4. **A rule lives in the layer that owns the book pages.** `dice` classifies a
   Spectacular roll; `task` applies it, because p.127 is a statement about
   _tasks_. `route` computes expected traffic but hands every dice-driven
   adjustment back to its caller.

5. **Beware assertions that pass when the mechanism is absent.** This is the
   repo's signature defect class, and the comments name it explicitly: the
   panic clause in `AssertRejected`, the paired stdout leak checks, the
   derived-not-restated seed matcher.

A last observation about the comments. They are unusually dense, and mostly they
are not explaining the code — they are **adjudicating the rules**, recording
which printing of a self-contradicting table was followed and why, usually tied
to a named golden test. Read them as the reasoning that produced the code, not
as a gloss on it.
