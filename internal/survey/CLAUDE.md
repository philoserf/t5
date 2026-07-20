# survey (with sectorgen and route)

Interstellar mapping (Book 3 pp. 12-15, 21, 27-28), spanning `internal/sectorgen/`,
`internal/survey/`, and `internal/route/`: the 32×40/8×10 hex grid with CCRR/A-P coordinates,
`Hex.Distance` (parsec jump distance via even-q offset→cube), and density system-presence rolls.

`survey.Sector` is the one entry point: it composes them with `systemgen` into detailed Second
Survey records — the coarse map flags constrain generation (`systemgen.GenerateForMap`: gas-giant
symbol→≥1 giants / none→0, asteroid symbol→Size-0 belt mainworld) so preview and detail agree —
then marks sector (Cs) and subsector (Cp) capitals, lays **trade routes** (`route.Build` — a pure,
dice-free graph linking Ix≥4 worlds within J-4, bridging distant ones through intermediate worlds),
and sites Scout Way Stations (~1/50 pc of route, bumping Ix).

A sector is always surveyed whole — its systems share one dice stream, and capitals/routes/way
stations need the whole region — so every `cmd/sectorgen` view _selects_ from one survey and they
agree on what sits in a hex: the default lists one subsector, `-sector` lists all of it with the
routes, `-hex CCRR` prints one system's sheet.

`Record.Sheet` (`survey/sheet.go`) is the deep renderer: the one-line Second Survey record shows
only the mainworld — ~94% of what the generators compute (the stellar family, the orbit map, every
secondary world and moon with its own UWP) plus the mainworld's port facilities, native status, and
Resource Units have no other renderer. `cmd/sectorgen -hex CCRR` prints one system's full sheet.
