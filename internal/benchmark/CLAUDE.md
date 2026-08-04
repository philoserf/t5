# benchmark

Pure shared benchmark scales from Book 1 pp. 36-43: Fame/Danger/Threat and Risk, impact speed
and damage, Hot-and-Cold temperature damage, insulation, and object Size. Tables preserve the
book's deliberately irregular benchmark values rather than interpolating them.

The package never rolls dice. Callers supply Flux results to `Risk` and `RandomSize`. Distances
for object Size are normalized to millimeters, impact mass is displacement tons, and temperature
damage is Hits per one-minute round. Invalid bounded lookups return `ok=false`; formula functions
return zero for non-positive physical inputs.

- **`SpeedEntry.KPH` 0 above Speed 16 means the book prints no kph, not zero speed.** Table 11a
  leaves the kph cell blank for Speeds 17-20 (only Hits are printed there); Speed 0 is the one
  genuine 0 kph. The zero value stands in for the blank cell — no sentinel, since nothing in the
  package distinguishes "unprinted" from zero elsewhere either.
- **The 11b damage-dice column (1D…40D, 100D-300D) is deliberately omitted.** The package never
  rolls; `Temperature.Hits` carries the printed flat Hits value, which is what a non-rolling
  consumer needs. Do not re-add the dice column without a consumer that actually rolls it — this
  is a settled omission, not an oversight.
