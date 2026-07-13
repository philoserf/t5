# t5

Multipurpose Traveller5 (T5) workspace: Go code, extracted rules reference, and
worldbuilding synthesis for fiction.

Go module: `github.com/philoserf/t5`.

## Layout

- `internal/` — the Go engine: `dice`, `ehex`, `uwp`, and the generators
- `cmd/` — command-line front ends for the generators
- `docs/pdf/` — source rulebooks (T5 Core Rules Books 1–3 + Read Me), git-ignored
- `docs/reference/` — text extracted from those PDFs, git-ignored (see below)

## Generators

All generators share a single seedable dice engine (`internal/dice`), so any run
is reproducible with `-seed`. Each command takes `-n` (count) and `-seed`.

```sh
go run ./cmd/worldgen  -n 5 -seed 42            # mainworld Universal World Profiles
go run ./cmd/systemgen -n 5 -seed 42            # star systems (stars, gas giants, belts, mainworld)
go run ./cmd/chargen   -n 5 -seed 42            # character UPPs
go run ./cmd/chargen   -career scout -n 5 -seed 42  # careered characters (qualify, terms, skills, muster-out)
```

Example output:

```
$ go run ./cmd/worldgen -n 3 -seed 42
D665656-7  Ga Ni Ag Ri
C7A5958-A  Fl Hi In
B160113-B  De Lo

$ go run ./cmd/systemgen -n 1 -seed 7
Primary: K5 V
Primary Companion: M0 VI
Worlds: 6  PBG: 901
Mainworld: E643231-6 Lo Po {-3}(810-3)[1164] B - -

$ go run ./cmd/chargen -career scout -n 2 -seed 7
Scout — age 22, mustered out after 1 term
  UPP            785399
  Homeworld      B6667B8-9   Ga Ag Ri
  Education      BA — Psychology (major), Robotics (minor)
  Skills         Actor-1, Animals-1, Biologics-3, Psychology-6, Robotics-5, Trader-1
  Benefits       Ship Share

Age 18 — did not qualify for a career
  UPP            5764A3
  Homeworld      B87A663-8   Wa Ni
  Education      BA — Psychology (major), Robotics (minor)
  Skills         Driver-1, Psychology-4, Robotics-2, Seafarer-1
```

A character's homeworld (a generated world) grants one skill per Trade
Classification (Book 1 p. 56), and college — when they qualify — grants a Major
and Minor and raises Edu (Book 1 p. 60), so the worldgen and chargen engines
feed each other before a career even begins.

The mainworld line is the full world record — UWP, trade classifications, the
`{Ix}(Ex)[Cx]` Extensions, nobility, bases, and travel zone.

The engine is faithful to the rules and validated against the books' own worked
examples — e.g. worldgen reproduces the canonical Regina profile `A788899-C` and
its full record `A788899-C Ph Pa Ri {+4}(D7E+4)[9C6D] BcCeF NS -`.
Chargen runs the character lifecycle (homeworld skills, college with a Major and
Minor, then career qualification, four-year terms with Risk & Reward and aging,
skill eligibility, and mustering out) for all thirteen careers, each selected
with its own `-career` value: the Scout, the fixed-characteristic Rogue, the
rankless Agent, the auto-enrolling Citizen (benign Citizen Life), the Fame-driven
Entertainer, the Masterpiece-making Craftsman, the single-ladder Scholar
(Research and Publications), the Office-Politics Functionary, the Noble (Return &
Intrigue, Elevation and Land Grants), the dual-track Merchant (Rating/Officer,
Ship Shares), and three ranked armed-forces careers — the Soldier, Marine, and
Spacer — whose enlisted/officer ladders, Commissions, and Medal-boosted
promotions exercise the rank engine (e.g. `-career soldier`). A character may
serve **several careers in sequence** — `-career scout,merchant,noble` — aging
and accumulating skills, benefits, and rank across each. The rest of the
education institutions and per-world orbital detail (systemgen) are the next
stages.

## Development

```sh
task            # run the tests
task check      # gofmt, vet, test — the pre-commit gate
task extract    # regenerate docs/reference/ from the PDFs
```

## Rulebooks

The T5 PDFs are **not committed** (git-ignored) — they are Marc Miller / Far Future
Enterprises material and not ours to redistribute. To work with the reference and
extraction pipeline, obtain them yourself and drop them in `docs/pdf/`:

- `Traveller5 Core Rules Book 1 Characters and Combat.pdf`
- `Traveller5 Core Rules Book 2 Starships.pdf`
- `Traveller5 Core Rules Book 3 Worlds and Adventures.pdf`
- `Traveller5 Read Me.pdf`

Then `task extract` runs `pdftotext` over them into `docs/reference/*.txt` for local
reading and reference. That extracted text is also git-ignored for the same copyright
reason — it is a local, regenerable derivation, not something we redistribute. The Go
generators encode only the rules' _mechanics_ (formulas and small tables), hand-authored
from the reference and validated against the books' worked examples.
