# Senses

This package implements Traveller5 sense actions and the six fixed-format
sense IDs from Book 1 pp. 186–199.

Canonical IDs are `V-00-BBB`, `H-00-FSVR`, `S-00-S`, `T-00-S`, `A-00-A`, and
`P-00-TP`. Constants are two decimal digits; the remaining fields are uppercase
vision-band codes or eHex digits. `Parse` and the text marshaling methods are
strict and round-trip only canonical forms. A Constant of zero means the sense
is absent; its detail fields remain encoded so the ID retains its fixed shape.

Per-sense accessors decode only data explicitly carried by the ID. Narrative
effects, referee descriptions, species-generation rolls, characteristic scents,
and language behavior remain with their owning systems rather than being
inferred here.

## Settled ruling — human Hearing detail is 9382

Book 1 prints two values for the human Hearing ID, and `H-16-9382` is the
settled reading; do not re-litigate it from the p.193 sidebar alone. The
evidence is 3-vs-1 for 9382: p.192's format box ("Human= 16 | 9 3 8 2"),
p.192's Eneri Dinsha worked example `H-16-9382`, and the character-sheet
printing all say 9382, while only p.193's "Calculating What Sounds Can Be
Heard" sidebar derives 9392 (Voice=9 = 512 Hz). Per the repo's book-conflict
convention, worked examples outrank derived cells — count both sides before
reopening.
