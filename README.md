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

$ go run ./cmd/chargen -career scout -n 2 -seed 42
547948  age 30  homeworld D665656-7  Scout: 3 terms, MusteredOut  [Actor-1 Animals-1 Comms-2 Computer-6 Driver-1 Gunner-3 JOT-2 Language-5 Language/Galanglic-2 Pilot-4 Trader-1]  benefits: Forbidden Knowledge
659378  age 22  homeworld C6439BB-A  Scout: 1 terms, MusteredOut  [Biologics-1 Computer-1 Gunner-3 JOT-2 Language-0 Language/Galanglic-1 Pilot-1 Steward-1 Streetwise-1]
```

Homeworld skills come from the character's homeworld (a generated world): one
skill per Trade Classification (Book 1 p. 56), so the worldgen and chargen
engines feed each other.

The mainworld line is the full world record — UWP, trade classifications, the
`{Ix}(Ex)[Cx]` Extensions, nobility, bases, and travel zone.

The engine is faithful to the rules and validated against the books' own worked
examples — e.g. worldgen reproduces the canonical Regina profile `A788899-C` and
its full record `A788899-C Ph Pa Ri {+4}(D7E+4)[9C6D] BcCeF NS -`.
Chargen runs the character lifecycle (homeworld skills, then qualification,
four-year terms with Risk & Reward and aging, skill eligibility, and mustering
out) for the Scout; education, the other careers, and per-world orbital detail
(systemgen) are the next stages.

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
