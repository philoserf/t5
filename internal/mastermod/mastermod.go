// Package mastermod provides the named random-reference tables from Traveller5
// Book 1 pp. 264-269.
package mastermod

import (
	"fmt"
	"slices"
	"strings"
)

// Table is a named contiguous die-total lookup. Dice documents the source roll
// notation; Rows[0] corresponds to Minimum.
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

// Lookup returns the row for a die total. Substitutions replace tokens written
// as <Name> deterministically; missing substitutions leave the source token
// intact. Out-of-range totals return false.
func (t Table) Lookup(total int, substitutions map[string]string) (string, bool) {
	if !t.Valid() {
		return "", false
	}

	index := total - t.Minimum
	if len(t.Rolls) != 0 {
		index, _ = slices.BinarySearch(t.Rolls, total)
		if index >= len(t.Rolls) || t.Rolls[index] != total {
			return "", false
		}
	} else if total < t.Minimum || total > t.Maximum() {
		return "", false
	}

	row := t.Rows[index]

	keys := make([]string, 0, len(substitutions))
	for key := range substitutions {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	for _, key := range keys {
		row = strings.ReplaceAll(row, "<"+key+">", substitutions[key])
	}

	return row, true
}

var registry = map[string]Table{}

// Get returns a named Master Mod table.
func Get(name string) (Table, bool) {
	t, ok := registry[name]

	return t, ok
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
	return Table{Name: name, Dice: dice, Minimum: rolls[0], Rolls: rolls, Rows: rows}
}
