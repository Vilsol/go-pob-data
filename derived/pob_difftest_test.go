package derived

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/Vilsol/go-pob-data/testdata"
)

// The difftests compare each table with the file Path of Building's export scripts wrote from the same
// game data: PoB's src/Data, read by luajit from the checkout POB_DIR names (default ../../PathOfBuilding),
// against pobVersion's committed data. Without either they are skipped.

// pobVersion is the game version the PoB checkout's src/Data was exported from.
const pobVersion = "3.29"

func pobData(t *testing.T, file string, mode ...string) []byte {
	t.Helper()
	dir := os.Getenv("POB_DIR")
	if dir == "" {
		dir = filepath.Join("..", "..", "PathOfBuilding")
	}
	path := filepath.Join(dir, "src", "Data", file)
	if _, err := os.Stat(path); err != nil {
		t.Skipf("no PoB checkout (POB_DIR): %v", err)
	}
	luajit, err := exec.LookPath("luajit")
	if err != nil {
		t.Skip("no luajit")
	}
	out, err := exec.Command(luajit, append([]string{filepath.Join("testdata", "pobdump.lua"), path}, mode...)...).Output() //nolint:gosec // the test's own script
	if err != nil {
		t.Fatalf("pobdump %s: %v", file, err)
	}
	return out
}

func pobDecode[T any](t *testing.T, file string, mode ...string) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(pobData(t, file, mode...), &v); err != nil {
		t.Fatalf("%s: %v", file, err)
	}
	return v
}

func committedSource(t *testing.T) source {
	t.Helper()
	testdata.ServeCommittedData(t)
	return source{ctx: context.Background(), version: pobVersion, assetCache: noCache{}}
}

type noCache struct{}

func (noCache) Get(string) ([]byte, error) { return nil, nil }
func (noCache) Set(string, []byte) error   { return nil }
func (noCache) Exists(string) bool         { return false }

// diffKeys reports every key of want missing from got, of got missing from want, and of both whose
// values differ, at most limit of each.
func diffKeys[V any](t *testing.T, what string, got, want map[string]V, equal func(a, b V) bool) {
	t.Helper()
	const limit = 10
	var missing, extra, differ []string
	for _, k := range slices.Sorted(maps.Keys(want)) {
		g, ok := got[k]
		switch {
		case !ok:
			missing = append(missing, k)
		case !equal(g, want[k]):
			differ = append(differ, fmt.Sprintf("%q: got %+v, PoB %+v", k, g, want[k]))
		}
	}
	for _, k := range slices.Sorted(maps.Keys(got)) {
		if _, ok := want[k]; !ok {
			extra = append(extra, k)
		}
	}
	for _, d := range []struct {
		name  string
		items []string
	}{{"missing", missing}, {"not in PoB", extra}, {"different", differ}} {
		if len(d.items) > 0 {
			t.Errorf("%s: %d keys %s, e.g. %q", what, len(d.items), d.name, d.items[:min(limit, len(d.items))])
		}
	}
	t.Logf("%s: %d keys, %d in PoB", what, len(got), len(want))
}

type pobScalability struct {
	Formats    []string `json:"formats"`
	IsScalable bool     `json:"isScalable"`
}

func TestModScalabilityMatchesPoB(t *testing.T) {
	want := pobDecode[map[string][]pobScalability](t, "ModScalability.lua")
	rows, err := buildModScalability(committedSource(t))
	if err != nil {
		t.Fatal(err)
	}
	got := map[string][]pobScalability{}
	for _, row := range rows {
		entries := make([]pobScalability, len(row.Entries))
		for i, e := range row.Entries {
			entries[i] = pobScalability(e)
		}
		got[row.Line] = entries
	}
	diffKeys(t, "ModScalability", got, want, func(a, b []pobScalability) bool {
		return len(a) == len(b) && (len(a) == 0 || reflect.DeepEqual(a, b))
	})
}

func TestItemModStatOrderMatchesPoB(t *testing.T) {
	want := pobDecode[struct {
		Exact      map[string]float64 `json:"exact"`
		Normalised map[string]float64 `json:"normalised"`
	}](t, "ModItemExclusive.lua", "uniqueModStatOrder")
	rows, err := buildItemModStatOrder(committedSource(t))
	if err != nil {
		t.Fatal(err)
	}
	exact, normalised := map[string]float64{}, map[string]float64{}
	for _, row := range rows {
		if row.Exact != 0 {
			exact[row.Line] = row.Exact
		}
		if row.Normalised != 0 {
			normalised[row.Line] = row.Normalised
		}
	}
	equal := func(a, b float64) bool { return a == b }
	diffKeys(t, "exact", exact, want.Exact, equal)
	diffKeys(t, "normalised", normalised, want.Normalised, equal)
}
