# EPIC adventures

`epic` implements the reusable structure on Book 3 pp. 274-279. `Scaffold`
returns four ordered acts, each with five independently completable scene slots,
followed by the climactic final resolution, whose “scenes as necessary” list is
initially empty. It intentionally does not invent
cast, background, synopsis, scene contents, or resources: those are authored or
supplied by the other generators named by the book's “There's More...” hooks.

`Generate` adds one unannounced session theme by consuming exactly two dice.
The first die selects a column and the second a row in the 6-by-6 Typical Mods
table on p. 279; `Theme` exposes that lookup without randomness. The much larger
“Other Themes” word bank has no 1D-by-1D selection procedure printed on the page
and is outside issue #182's explicit 1D×1D scope.

Three spellings are corrected from context, in two distinct categories.
`Pro table` → `Profitable` and `Per dy` → `Perfidy` (the latter in the
out-of-scope word bank) are extraction artifacts — dropped "fi" ligatures;
the rendered page prints both words correctly. `Truthfullness` →
`Truthfulness` is different: the book itself prints the double-l, in both
copies of the table (Book 3 p.279 and Book 1 p.267), so this one is a
deliberate normalization of the book's own typo, not an OCR fix.
`internal/mastermod` carries the same Themes table and makes the opposite
call, preserving the printed spelling — both are settled; do not "align"
one to the other.
