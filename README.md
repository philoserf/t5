# t5

Multipurpose Traveller5 (T5) workspace: Go code, extracted rules reference, and
worldbuilding synthesis for fiction.

Go module: `github.com/philoserf/t5`.

## Layout

- `docs/pdf/` — source rulebooks (T5 Core Rules Books 1–3 + Read Me), git-ignored
- `docs/reference/` — text extracted from those PDFs, git-ignored (see below)
- `internal/` — the Go engine (`dice`, and generators built on it)

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
