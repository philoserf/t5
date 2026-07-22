# t5

A [Traveller5](https://www.farfuture.net/) (T5) toolkit in Go: a faithful, seedable
implementation of the T5 generators — worlds, star systems, characters, alien species,
sectors, and starships — alongside the extracted rules reference and worldbuilding notes that
draw on them.

Go module: `github.com/philoserf/t5`. Standard library only.

Jump to what you need:

- [**Run the generators**](#run-the-generators) — the command-line tools.
- [**Build on the engine**](#build-on-the-engine) — the Go packages.
- [**Understand & contribute**](#understand--contribute) — design, layout, and how to help.

Requires [Go](https://go.dev/) 1.26+ (`go run` needs nothing else). Contributors can install
the full toolchain — go, go-task, golangci-lint, poppler — with `task deps` (Homebrew).

---

## Run the generators

Six command-line generators. Every run is reproducible: pass `-seed` and you get the same
result every time; omit it for a fresh random run. Generators write their records to
**stdout** and everything else — notes, errors — to **stderr**, so a piped record stream stays
clean.

```sh
go run ./cmd/worldgen  -n 5 -seed 42                      # mainworld Universal World Profiles
go run ./cmd/systemgen -n 1 -seed 7                       # full star systems (stars, giants, belts, orbits, moons)
go run ./cmd/chargen   -career scout,merchant -seed 7     # characters through their careers
go run ./cmd/sophont   -seed 42                           # alien species templates
go run ./cmd/sectorgen -seed 42                           # a surveyed subsector (worlds, capitals, trade routes)
go run ./cmd/shipgen   -hull A -tl 12 -config S -maneuver A -jump A \
                       -weapon "beamlaser:T1:orbit" -defense blackglobe   # designed starships
```

Shared flags: `-n` (count) and `-seed`. Command-specific ones:

- **chargen** — `-career name[,name…]`: run one career or a sequence (aging and accumulating
  skills across each); omit for a bare UPP.
- **sectorgen** — `-density D`, `-subsector L`, `-hex CCRR`: the coarse density, one subsector
  listing, or one system's full sheet.
- **shipgen** — `-hull -tl -config -structure -armor -maneuver -jump -power -mission -weapon
-defense`: the design is deterministic (no dice); infeasibility is reported, not fatal.
- **sophont** — `-char`: also roll an individual of the generated species.

### Sample output

```
$ go run ./cmd/worldgen -n 3 -seed 42
D665656-7  Ga Ni Ag Ri
C7A5958-A  Fl Hi In
B160113-B  De Lo

$ go run ./cmd/chargen -career scout -seed 7
Scout — age 27, mustered out after 1 term
  UPP            785399
  Homeworld      B6667B8-9   Ga Ag Ri
  Education      BA — Psychology (major), Robotics (minor)
  Skills         Actor-1, Animals-1, Athlete-1, Biologist-1, Comms-2, Fighter-1, Hostile Environ-1, JOT-1, Medic-1, Psychology-4, Robotics-2, Trader-1
  Benefits       Ship Share

$ go run ./cmd/sectorgen -seed 42 | head -3
0101 Maesavo E410100-7 Lo Co {-3}(500-2)[1139] B - - 233 17 Im K8 V BD K4 VI
0104 Helo B120557-F De Ni Po Ho {+1}(E42+1)[566J] B - - 820 10 Im G5 V K5 VI
0107 Magi A200440-E Va Ni Co Tz {+1}(F32+2)[653J] B - - 932 14 Im M2 V BD

$ go run ./cmd/shipgen -hull A -tl 12 -config S -maneuver A -jump A -weapon "beamlaser:T1:orbit" -defense blackglobe
Ship  X-AS22
Hull:    A 100t Streamlined · TL-12 · Frame-and-Plate
Drives:  Maneuver-A 2G · Jump-A J-2 · Power-A
Armor:   1 layers AV-12
Weapons: Standard Orbit Single Turret Beam Laser-11 Mod=+0. 2 tons. MCr1.1. Hits= 1D. R=08.
```

Those record lines are the canonical T5 "Second Survey" forms — a world line carries its UWP,
trade classifications, the `{Ix}(Ex)[Cx]` Extensions, bases, and travel zone; a ship line is
the compact QSP profile. The generators feed each other: chargen composes worldgen (the
homeworld grants one skill per trade classification) and education before a career even begins.

---

## Build on the engine

The generators are a thin shell over a Go library. The packages live under `internal/`, so
they are consumed **from within this module** — you build on the engine by adding a `cmd/` tool
or an `internal/` package, not by importing it from another project.

One rule governs the whole thing: **randomness is injected, never global.** A `*dice.Roller` is
the single source of dice — hand it a seed for reproducibility, or a scripted die sequence for
exact, deterministic tests.

```go
import (
	"fmt"

	"github.com/philoserf/t5/internal/dice"
	"github.com/philoserf/t5/internal/worldgen"
)

r := dice.NewWithSeed(42)   // reproducible; dice.New() = random; dice.NewScripted(3, 5) = fixed rolls
uwp := worldgen.Generate(r) // a mainworld Universal World Profile
fmt.Println(uwp)            // "D665656-7"
```

Entry points, each taking a `*dice.Roller` (except ship design, which is pure):

| Package     | Call                                             | Produces                                            |
| ----------- | ------------------------------------------------ | --------------------------------------------------- |
| `worldgen`  | `Generate(r)` · `GenerateWorld(r, …)`            | a UWP · a full world record                         |
| `systemgen` | `Generate(r)`                                    | a star system (stars, giants, belts, orbits, moons) |
| `chargen`   | `GenerateCareered(r, policy, homeworld, career)` | a character through their careers                   |
| `sophont`   | `Generate(r)` · `chargen.GenerateSophont(…)`     | an alien species · an individual of it              |
| `shipgen`   | `Design(spec)`                                   | a designed ship (deterministic)                     |
| `vehiclegen`| `Design(spec)` · `Generate(r)`                   | a ground vehicle, flyer, or watercraft              |
| `survey`    | `Sector(…)`                                      | a surveyed sector with capitals and trade routes    |

The layers, bottom to top:

- **`dice`** — the T5 dice engine (Flux, roll-low checks/resolves, Many-Dice, chart notation),
  the one source of randomness.
- **Primitives** — `ehex` (extended-hex digits), `uwp` (the `Profile`), `skill`, `calendar`,
  `task` (the Universal Task profile), `rangeband`.
- **Generators** — `worldgen`, `systemgen`, `chargen`, `sophont`, `shipgen`, `vehiclegen`, and the mapping
  stack (`sectorgen` + `survey` + `route`).
- **Play & economy** — `senses`, `personals`, `combat`, `shipcombat`, and `trade` build on the
  task engine.

Determinism is the contract at every layer: the same seed (or the same scripted rolls) yields
the same output — which is exactly what makes the golden tests possible.

---

## Understand & contribute

This repo holds three kinds of work in one place:

- **Go code** — the generators above, faithful to the T5 rules.
- **Rules reference** — the core rulebooks extracted from PDF to text (see below).
- **Worldbuilding** — setting notes and fiction that draw on the generators.

### Design

Each generator **transcribes the rulebook's own tables and formulas** as pure functions that
take their rolls as arguments, then locks the result with a **golden test** built from a worked
example in the books — e.g. worldgen reproduces the canonical Regina record,
`A788899-C Ph Pa Ri {+4}(D7E+4)[9C6D] BcCeF NS -`. (The book prints Regina with two more codes,
`An` and `Cp` — an Ancient Site and a Subsector Capital — both referee-assigned, so the
generated line is the generatable _subset_ of the book's, not a copy of it.) The house style is
YAGNI: structure arrives when real content fills it.

### Layout

- `internal/` — the engine and generators (module-local)
- `cmd/` — the six command-line front ends
- `docs/pdf/` — the source rulebooks, **git-ignored** (see below)
- `docs/reference/` — text extracted from those PDFs, git-ignored and regenerable

### Working on it

```sh
task              # run the tests
task check        # golangci-lint + tests — the pre-commit gate (also runs in CI)
task lint         # golangci-lint run
task fmt          # gofumpt + goimports + golines
task audit        # non-gating report of what the globally-disabled linters would flag
task extract      # regenerate docs/reference/ from the PDFs
task deps         # install the toolchain from the Brewfile
```

The lint config (`.golangci.yml`) is deliberately aggressive — `default: all`, disabling only
where a linter fights the repo's transcribed-table design (magic numbers, data globals, terse
dice idioms). `task audit` reports what those exceptions hide, for periodic review. To add a
generator: transcribe its tables from `docs/reference/`, build it on a `*dice.Roller`, and lock
it with a golden test from a worked example. The backlog lives in **GitHub issues** (the
"Triage and Tracking" project) — one issue per unstarted generator and per deferred piece.

### Rulebooks

The T5 PDFs are **not committed** (git-ignored) — they are Marc Miller / Far Future Enterprises
material, not ours to redistribute. To use the reference and extraction pipeline, obtain them
yourself and drop them in `docs/pdf/`:

- `Traveller5 Core Rules Book 1 Characters and Combat.pdf`
- `Traveller5 Core Rules Book 2 Starships.pdf`
- `Traveller5 Core Rules Book 3 Worlds and Adventures.pdf`
- `Traveller5 Read Me.pdf`

Then `task extract` runs `pdftotext` over them into `docs/reference/*.txt` for local reading.
That extracted text is git-ignored too, for the same copyright reason — a local, regenerable
derivation. The Go generators encode only the rules' _mechanics_ (formulas and small tables),
hand-authored from the reference and validated against the books' worked examples.
