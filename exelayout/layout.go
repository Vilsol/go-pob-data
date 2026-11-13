// Package exelayout reads every .dat table's row layout out of the PoE client executable: each table's
// loader checks that its row count times the row size lands on the 0xBB marker, then calls a function
// that fixes up the row's strings, keys, rows and lists. Comparing that against dat-schema catches
// columns the schema has wrong before they decode as garbage.
package exelayout

import (
	"io"
	"sort"
	"strings"
)

// Pointer is a column the client fixes up after loading the rows.
type Pointer struct {
	Kind   Kind   `json:"kind"`
	Target string `json:"target,omitempty"`
	Offset int    `json:"offset"`
	// Element is a list's element size, when the client scales by it.
	Element int `json:"element,omitempty"`
	// Proven is set when the key's resolving helper scales by Target's own row size.
	Proven bool `json:"proven,omitempty"`
}

// Table is one table's layout as the client loads it.
type Table struct {
	Name     string    `json:"name"`
	Pointers []Pointer `json:"pointers,omitempty"`
	RowSize  int       `json:"rowSize"`
	// Walked is set when the fixup function's row loop was found, so Pointers is the full set.
	Walked bool `json:"walked,omitempty"`
}

// Layout is every table a client loads.
type Layout struct {
	Patch   string  `json:"patch"`
	Tables  []Table `json:"tables"`
	Version int     `json:"version"`
}

// formatVersion changes whenever the analysis would produce a different Layout, invalidating cached ones.
const formatVersion = 1

// Table returns the table named name, ignoring case.
func (l *Layout) Table(name string) (Table, bool) {
	for _, t := range l.Tables {
		if strings.EqualFold(t.Name, name) {
			return t, true
		}
	}
	return Table{}, false
}

// Analyze reads the layout of every table the executable in r loads.
func Analyze(r io.ReaderAt) (*Layout, error) {
	img, err := parseImage(r)
	if err != nil {
		return nil, err
	}
	names := img.rowNames()
	starts := map[uint32]bool{}
	for _, cs := range names {
		for _, c := range cs {
			starts[c.rva] = true
		}
	}
	refs := img.references(starts)

	type found struct {
		Table
		fixup uint32
	}
	var tables []*found
	seen := map[string]bool{}
	for _, cs := range names {
		if t, size, fixup, ok := img.loaderFor(cs, refs); ok && !seen[strings.ToLower(t)] {
			seen[strings.ToLower(t)] = true
			tables = append(tables, &found{Table{Name: t, RowSize: size}, fixup})
		}
	}

	sizes := make(map[string]int, len(tables))
	for _, t := range tables {
		sizes[strings.ToLower(t.Name)] = t.RowSize
	}
	fx := newFixups(img, sizes)
	out := &Layout{Version: formatVersion, Tables: make([]Table, 0, len(tables))}
	for _, t := range tables {
		if t.fixup != 0 {
			if b, e, ok := img.function(t.fixup); ok && b == t.fixup {
				t.Pointers, t.Walked = fx.pointers(img.decode(b, e), t.RowSize)
			}
		}
		out.Tables = append(out.Tables, t.Table)
	}
	sort.Slice(out.Tables, func(a, b int) bool { return out.Tables[a].Name < out.Tables[b].Name })
	return out, nil
}

// loaderFor finds the loader among the functions referencing one "<Table>_Row" string, trying its longest name first.
func (img *image) loaderFor(cs []candidate, refs map[uint32][]uint32) (string, int, uint32, bool) {
	for _, c := range cs {
		for _, at := range refs[c.rva] {
			b, e, ok := img.function(at)
			if !ok {
				continue
			}
			if size, fixup, ok := loaderPattern(img.decode(b, e)); ok {
				return c.name, size, fixup, true
			}
		}
	}
	return "", 0, 0, false
}
