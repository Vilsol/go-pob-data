package extractor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/MarvinJWendt/testza"
)

func refitFixture(columns []string, rows int, cell func(column string, row int) string) *RefitTable {
	t := &RefitTable{Cells: map[string][]string{}, Kinds: map[string]string{}, Columns: columns}
	for r := range rows {
		t.IDs = append(t.IDs, fmt.Sprintf("Row%d", r))
		for _, c := range columns {
			v := cell(c, r)
			t.Cells[c] = append(t.Cells[c], v)
			if k := jsonKind(json.RawMessage(v)); k != "null" {
				t.Kinds[c] = k
			}
		}
	}
	return t
}

func TestRefitMovedAndLost(t *testing.T) {
	const v1, v2, v3, v4 = "Var1", "Var2", "Var3", "Var4"
	thrice := func(r int) string { return fmt.Sprint(r * 3) }
	week := func(r int) string { return fmt.Sprint(r % 7) }
	fixture := func(rows int, columns map[string]func(int) string, order ...string) *RefitTable {
		return refitFixture(order, rows, func(c string, r int) string {
			if f := columns[c]; f != nil {
				return f(r)
			}
			return "0"
		})
	}

	older := fixture(50, map[string]func(int) string{
		v1: thrice,
		v2: week,
		v3: func(r int) string { return fmt.Sprint(r * 11) },
	}, "Id", v1, v2, v3, "Flag")
	// A column inserted before Var1 renames every unnamed column after it; Var3's values are gone, but nothing
	// agrees better with them than the new Var3.
	newer := fixture(60, map[string]func(int) string{
		v2: thrice,
		v3: week,
		v4: func(r int) string { return fmt.Sprint(r%2 == 0) },
	}, "Id", "Header", v2, v3, v4)

	got := map[string]RefitFinding{}
	for _, f := range Refit(older, newer) {
		got[f.Column+"/"+f.Kind] = f
	}
	testza.AssertEqual(t, v2, got[v1+"/moved"].To)
	testza.AssertEqual(t, v3, got[v2+"/moved"].To)
	testza.AssertEqual(t, "changed", got[v3+"/changed"].Kind)
	testza.AssertEqual(t, 3, len(got))
}

func TestRefitStable(t *testing.T) {
	cell := func(c string, r int) string {
		if c == "Count" {
			return fmt.Sprint(r)
		}
		return fmt.Sprint(r % 2)
	}
	columns := []string{"Count", "Copy", "Twin"}
	older := refitFixture(columns, 30, cell)
	newer := refitFixture(columns, 30, cell)
	testza.AssertEqual(t, 0, len(Refit(older, newer)))
}

func TestRefitRetyped(t *testing.T) {
	columns := []string{"Flag"}
	older := refitFixture(columns, 30, func(string, int) string { return "0" })
	newer := refitFixture(columns, 30, func(string, int) string { return "false" })
	f := Refit(older, newer)
	testza.AssertEqual(t, 1, len(f))
	testza.AssertEqual(t, "retyped", f[0].Kind)
}

// The 3.30 AtlasNode gained MapNodeHeader at column 32, so every unnamed column after it took the next name and
// Var33 and Var35 turned from numbers into booleans (fixed by hand in raw.AtlasNode); Var32's values are now
// MapNodeHeader's.
func TestRefitAtlasNode330(t *testing.T) {
	older, err := LoadRefitTables(dataPath, "3.29", "AtlasNode")
	testza.AssertNoError(t, err)
	newer, err := LoadRefitTables(dataPath, "3.30", "AtlasNode")
	testza.AssertNoError(t, err)

	got := map[string]bool{}
	for _, f := range Refit(older["AtlasNode"], newer["AtlasNode"]) {
		t.Log(f)
		got[f.Column+"/"+f.Kind+"/"+f.To] = true
	}
	testza.AssertTrue(t, got["Var33/retyped/number -> bool"])
	testza.AssertTrue(t, got["Var35/retyped/number -> bool"])
	testza.AssertTrue(t, got["Var32/moved/MapNodeHeader"])
}

// TestRefitReport compares the published versions REFIT_FROM and REFIT_TO (e.g. 3.29 and 3.30) and logs,
// per table, every column whose values did not stay under its name.
func TestRefitReport(t *testing.T) {
	from, to := os.Getenv("REFIT_FROM"), os.Getenv("REFIT_TO")
	if from == "" || to == "" {
		t.Skip("REFIT_FROM and REFIT_TO not set")
	}
	older, err := LoadRefitTables(dataPath, from)
	testza.AssertNoError(t, err)
	newer, err := LoadRefitTables(dataPath, to)
	testza.AssertNoError(t, err)

	names := make([]string, 0, len(newer))
	for n := range newer {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		if older[name] == nil {
			continue
		}
		t.Run(name, func(t *testing.T) {
			for _, f := range Refit(older[name], newer[name]) {
				t.Log(f)
			}
		})
	}
}

func TestLoadRefitTablesResolvesKeys(t *testing.T) {
	if _, err := os.Stat(filepath.Join(dataPath, "3.30", "raw", "AtlasNode.json.gz")); err != nil {
		t.Skip("no committed 3.30 tables")
	}
	tables, err := LoadRefitTables(dataPath, "3.30", "AtlasNode")
	testza.AssertNoError(t, err)
	// Connections point at AtlasNode rows by key; they read as those rows' IDs.
	testza.AssertContains(t, tables["AtlasNode"].Cells["Connections"][5], "#[Map")
}
