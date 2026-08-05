# Universal Task Format

This package owns the Traveller5 task rules in Book 1 pp. 120-131. `Difficulty`
is a ladder index; convert it to a dice count with `Dice`, `Hasty`,
`ExtraHasty`, or `Cautious`. `Resolve` and `ResolveDice` are the ordinary
roll-low paths and impose the chapter's Spectacular override.

## Special task modes

- Cooperation is target construction. `CooperativeTargetSkill` combines one
  characteristic with every participant's applicable skill; the characteristic
  form does the converse. Participant limits and whether a physical
  characteristic may be pooled are stated by the individual task/referee.
- `OpposedWinner` selects the unique lowest successful raw roll. A tie has no
  winner, matching the p. 130 dogfight example. `OpposedLoser` implements the
  extended-round variant: the unique highest failed roll is eliminated.
- `ResolveUncertainDice` rolls the player's visible dice before the referee's
  hidden dice, substitutes 3 for each hidden face in the apparent outcome, and
  retains the actual outcome for the referee. Visible and invisible
  Spectaculars are deliberately distinguished. `UncertaintyDice` applies the
  Hasty/Extra-Hasty increase; Cautious is represented by zero harder levels.
- `ResolveArcane` refuses a non-owner before drawing dice. Ownership is supplied
  as a fact by the character/adventure layer; an owned task otherwise resolves
  normally and may fail.

## Other explicit rules

`ThisIsHardDice`, the phantom Skill-3/Characteristic-7 assets, JOT substitution,
analog-characteristic substitution, and certification rank are small pure
helpers. They do not silently alter `Resolve`; callers know whether a task uses
a default skill, lacks an asset, or is a test.

`EvaluateMishap` implements the p. 131 failure roll exactly as printed: Flux plus
Reliability for an ordinary mishap, or Flux plus Safety for Dangerous and
Destructive tasks, triggers when negative. It classifies the consequence but
does not roll location, diagnosis, injury, or equipment damage: those use the
separate L/S/D/IA/FA tables and cross into the combat/equipment systems.

The p. 131 extracted prose describes the negative Flux-plus-Safety value as a
“Severity” and then compares it to the original positive Difficulty, without a
defined conversion onto the 1D-8D severity ladder. Do not invent a sign or
absolute-value conversion here; expose `MishapCheck.Total` so a future ruling
based on clearer source can add it without changing the trigger API.
