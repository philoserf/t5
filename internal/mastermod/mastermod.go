// Package mastermod provides the named random-reference tables from Traveller5
// Book 1 pp. 264-269.
//
// Rows are book-literal text; the seven "<none>" entries transcribe the
// printed instruction "Treat blank entries as <none>." (Book 1 p.267).
//
// Deliberate exclusions (do not re-add): Barrier Height/Width/Depth (blank
// cells in the appendix, Book 1 p.268), Scene Mods (a formula, not a die
// table, p.268), and the unrolled Imperiallines/Hortalez MegaCorporation rows
// (printed without 2x1D keys, p.269). See CLAUDE.md.
package mastermod

import (
	"fmt"
	"slices"
	"strings"
)

// Table is a named die-total lookup. Dice documents the source roll notation.
// With Rolls empty the keys are contiguous and Rows[0] corresponds to Minimum;
// a non-empty Rolls lists each row's key explicitly, for sparse or
// non-contiguous source columns.
type Table struct {
	Name    string
	Dice    string
	Minimum int
	Rows    []string
	Rolls   []int // optional explicit keys for sparse/non-contiguous source columns
}

// Maximum returns the largest accepted total.
func (t Table) Maximum() int {
	if len(t.Rolls) != 0 {
		return t.Rolls[len(t.Rolls)-1]
	}

	return t.Minimum + len(t.Rows) - 1
}

// Valid reports whether a table has a name, roll notation, and at least one
// non-placeholder row. Blank source cells are omitted when tables are built.
func (t Table) Valid() bool {
	if strings.TrimSpace(t.Name) == "" || strings.TrimSpace(t.Dice) == "" || len(t.Rows) == 0 {
		return false
	}

	for _, row := range t.Rows {
		if strings.TrimSpace(row) == "" {
			return false
		}
	}

	if len(t.Rolls) != 0 {
		if len(t.Rolls) != len(t.Rows) || !slices.IsSorted(t.Rolls) {
			return false
		}

		for i := 1; i < len(t.Rolls); i++ {
			if t.Rolls[i] == t.Rolls[i-1] {
				return false
			}
		}
	}

	return true
}

// Lookup returns the row for a die total. Out-of-range totals (including the
// gaps of a sparse table) return false.
func (t Table) Lookup(total int) (string, bool) {
	if len(t.Rolls) != 0 {
		index, ok := slices.BinarySearch(t.Rolls, total)
		if !ok {
			return "", false
		}

		return t.Rows[index], true
	}

	if total < t.Minimum || total > t.Maximum() {
		return "", false
	}

	return t.Rows[total-t.Minimum], true
}

var registry = map[string]Table{}

// Get returns a named Master Mod table. The returned Table is a deep copy;
// mutating its Rows or Rolls cannot affect the registry.
func Get(name string) (Table, bool) {
	t, ok := registry[name]
	if !ok {
		return Table{}, false
	}

	t.Rows = slices.Clone(t.Rows)
	t.Rolls = slices.Clone(t.Rolls)

	return t, true
}

// Names returns all registered names in stable alphabetical order.
func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}

	slices.Sort(names)

	return names
}

func register(tables ...Table) {
	for _, table := range tables {
		if !table.Valid() {
			panic(fmt.Sprintf("mastermod: invalid table %q", table.Name))
		}

		if _, exists := registry[table.Name]; exists {
			panic(fmt.Sprintf("mastermod: duplicate table %q", table.Name))
		}

		registry[table.Name] = table
	}
}

func table(name, dice string, minimum int, rows ...string) Table {
	return Table{Name: name, Dice: dice, Minimum: minimum, Rows: rows}
}

func sparse(name, dice string, rolls []int, rows ...string) Table {
	if len(rolls) == 0 {
		panic(fmt.Sprintf("mastermod: sparse table %q has no rolls", name))
	}

	return Table{Name: name, Dice: dice, Minimum: rolls[0], Rolls: rolls, Rows: rows}
}
