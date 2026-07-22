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

The extraction has three obvious OCR errors corrected from context:
`Truthfullness` becomes `Truthfulness`, `Pro table` becomes `Profitable`, and
`Per dy` (in the out-of-scope word bank) reads `Perfidy`.
