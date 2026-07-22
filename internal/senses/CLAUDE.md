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
