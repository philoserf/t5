# kinematics

Pure Traveller distance, signal-delay, and constant-acceleration calculations from Book 1
pp. 31-35. The package uses the rules' rounded working constants (150 million km/AU, 300,000
km/s, and 10 m/s² per G) so its results match the printed charts to their hour rounding rather
than scientific ephemerides: the functions return exact durations (e.g. 21.52h) where the charts
truncate to whole hours (21h), and the tests treat chart cells as ±1h goldens.

`KilometersFromAU` converts astronomical distance, `LightDelay` gives one-way signal time,
`ImpactTime` accelerates from rest through the endpoint, and `StartStopTime` accelerates to the
midpoint then decelerates to rest. Calculations are pure; non-positive distance or acceleration
returns zero.
