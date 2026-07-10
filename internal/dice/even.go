package dice

// Even distributions contort D6 results into ranges that a single D6 cannot
// produce evenly. Each is rolled as two dice — one selecting a row band, the
// other a column offset — and read off the Even Distribution tables in the
// Dice Appendix (p. 259).

// EvenDist1to9 returns an evenly distributed value in 1..9. It is most often
// used for the population multiplier of a world, where 0 and 10 are not
// meaningful outcomes.
func (r *Roller) EvenDist1to9() int {
	row, col := r.d6(), r.d6()
	// Rows 1-3 and 4-6 repeat the same three bands (1-3, 4-6, 7-9); each
	// column triple (1-3, 4-6) repeats an offset of 1-3.
	band := ((row - 1) % 3) * 3
	offset := ((col - 1) % 3) + 1
	return band + offset
}

// EvenDist0to9 returns an evenly distributed value in 0..9 — the equivalent of
// a decimal die (D10). Rows 1-5 map to the pairs (0,1)(2,3)(4,5)(6,7)(8,9) by
// column half; a row of 6 is rerolled.
func (r *Roller) EvenDist0to9() int {
	for {
		row, col := r.d6(), r.d6()
		if row == 6 {
			continue // "rr": reroll
		}
		v := (row - 1) * 2
		if col > 3 {
			v++
		}
		return v
	}
}

// EvenDist1to10 returns an evenly distributed value in 1..10, reusing the 0..9
// table with a result of 0 substituted by 10 (Dice Appendix, p. 259).
func (r *Roller) EvenDist1to10() int {
	if v := r.EvenDist0to9(); v != 0 {
		return v
	}
	return 10
}
