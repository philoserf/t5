# Personals

This package implements Book 1 pp. 181-185. The existing low-level `Resolve`
accepts precomputed Strategy and Tactic values and remains available unchanged.
`ResolveSelection` is the checked P1-grid path: it obtains the Strategy value,
validates the Strategy/Tactic pairing, applies the printed multiplier or Mod,
and resolves the Personal without changing dice order.

The p.184 P1 grid has four distinct cell meanings:

- blank: compatible, neutral multiplier x1;
- `no`: incompatible;
- `x2` or `x3`: Strategy multiplier;
- signed numbers in Common Interests, Common Enemies, and Pain: additive Mod.

`TacticEffectFor` preserves those distinctions. The star on selected Pain cells
is reported as `InflictsPain`. Both Insult and Pain make the Personal violent;
if resolution fails, `PersonalResult.Fight` is the explicit handoff for a caller
to begin combat. This package does not import or orchestrate combat.

`GenerateQuickNPC` rolls four secret 2D base values in the printed order:
Carouse, Query, Persuade, Command. `QuickNPC.Check` rolls the normal Purpose dice
against that value plus caller-supplied Laws and Mods. Generation consumes
exactly eight dice and checking consumes only the Purpose dice.

The dense grid was transcribed from a rendered image of printed p.184 rather
than relying on text-column extraction. `TestTacticMatrixExhaustive` asserts all
320 cells independently.
