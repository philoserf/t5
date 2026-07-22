# benchmark

Pure shared benchmark scales from Book 1 pp. 36-43: Fame/Danger/Threat and Risk, impact speed
and damage, Hot-and-Cold temperature damage, insulation, and object Size. Tables preserve the
book's deliberately irregular benchmark values rather than interpolating them.

The package never rolls dice. Callers supply Flux results to `Risk` and `RandomSize`. Distances
for object Size are normalized to millimeters, impact mass is displacement tons, and temperature
damage is Hits per one-minute round. Invalid bounded lookups return `ok=false`; formula functions
return zero for non-positive physical inputs.
