# systemgen

Star system creation (Book 3 pp. 16-17, 28): the stars (spectral type/decimal/size via the p. 28
table, transcribed from a rendered image since the dense grid does not survive text extraction),
gas giants, planetoid belts, world count, and the mainworld via `worldgen`. `classify` is the pure
table lookup; `rollStar`/`Generate` roll and compose. Secondary stars (Close/Near/Far) are placed
in orbit bands. The mainworld is a full `worldgen.World` (UWP + all derived data);
`System.SecondSurvey(hex, name, allegiance)` renders the canonical one-line record and `String`
shows it with PBG.

`placeOrbits` lays the full orbit map (mainworld/gas giants/belts/other worlds in concrete orbits,
rotate-per-star); `rollSatellites` gives every placed body its moons — each a real satellite with a
type (`satelliteType`, the p.29 Satellites tables) and UWP, capped to its parent's size with a
double-planet flag at equal size (Book 3 p.21), or a Ring.

`rollMoon` is the **single** moon-assembly path (both the satellite pass and the orbit map's
gas-giant-captured world go through it, so neither their dice order nor the tables they read can
drift — the Outer Worlds and Outer Satellites tables disagree at 1D=4, and a captured world is
created as a satellite, so it is typed as one). `satelliteParent` names the body an orbit's moons
belong to, which is **not** always the orbit's Kind: when the mainworld is itself a satellite, p.21
puts a gas giant (or a `GenerateHostWorld` BigWorld, floored at the mainworld's own Size) in its
orbit, so the orbit's moons are counted and capped for that parent and render as the mainworld's
_sibling_ moons. Any parent that has a UWP is classified by `satelliteBody`, the **one** read that
answers both halves at once — the count rule it takes and the cap it imposes — because when those
were separate decisions only the cap resolved the asteroid-belt code, so a belt mainworld was
capped as a belt but _counted_ as a world and rolled phantom moons (#309).

Orbit letters are orbit names, so `satelliteOrbits` keeps them unique per parent, nudging a
duplicate to the nearest free letter without touching the Flux roll (p.29, "adjust to an adjacent
or the closest possible orbit").

The size cap is applied **inside** generation via `worldgen.GenerateSatelliteWorld` — Atmosphere is
Flux+Size and Hydrographics is Flux+Atmosphere, so capping Size after the roll would leave a
profile describing the larger world and break the World Creation chart's own structural rules ("If
Siz=0, Atm=0", "If Siz <2, Hyd =0", p.24). Capping in place consumes identical dice, so it
re-derives rather than re-rolls.
