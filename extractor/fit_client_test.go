package extractor

import (
	"io"
	"os"
	"sort"
	"strings"
	"testing"
)

// fitClientTables reads every schema table a cached client holds. FIT_PATCH names the patch (e.g. 3.30.0.1.2)
// whose bundles are already in the bundle cache; the patch CDN serves only the live patch.
func fitClientTables(t *testing.T) map[string][]byte {
	t.Helper()
	patch := os.Getenv("FIT_PATCH")
	if patch == "" {
		t.Skip("FIT_PATCH not set")
	}
	short := strings.Join(strings.Split(patch, ".")[:2], ".")
	LoadSchema(PoE1, short)

	loader, err := GetBundleLoader(NewWebFS(PoE1, patch), IsNewHash(PoE1, short))
	if err != nil {
		t.Fatal(err)
	}

	out := map[string][]byte{}
	for name := range tableMap {
		f, err := loader.Open("data/" + strings.ToLower(name) + ".datc64")
		if err != nil {
			continue
		}
		b, err := io.ReadAll(f)
		if err != nil {
			continue
		}
		out[name] = b
	}
	return out
}

func sortedNames(m map[string][]byte) []string {
	names := make([]string, 0, len(m))
	for n := range m {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// TestFitClient runs the fit check over every table of a real client whose schema matches its row width.
func TestFitClient(t *testing.T) {
	tables := fitClientTables(t)

	checked, widthOff := 0, 0
	for _, name := range sortedNames(tables) {
		b := tables[name]
		filename := name + ".datc64"
		if checkRowWidth(b, filename) != nil {
			widthOff++
			continue
		}
		checked++
		if err := checkFit(b, filename); err != nil {
			t.Error(err)
		}
		for c, s := range fitShares(b, filename) {
			if s >= 0.05 {
				t.Logf("near miss: %s.%s %.0f%%", name, c, s*100)
			}
		}
	}
	t.Logf("%d tables checked, %d skipped for a row width mismatch", checked, widthOff)
}

// TestFitClientShifted rotates each table's columns by one (same row width, every column moved) and counts the
// tables the fit check refuses; a table with no string, list or key columns has nothing to misread.
func TestFitClientShifted(t *testing.T) {
	tables := fitClientTables(t)

	caught, missed := 0, []string{}
	for _, name := range sortedNames(tables) {
		b := tables[name]
		filename := name + ".datc64"
		if checkRowWidth(b, filename) != nil {
			continue
		}
		orig := tableMap[name]
		if len(orig.Columns) < 2 {
			continue
		}
		shifted := orig
		shifted.Columns = append([]Column{orig.Columns[len(orig.Columns)-1]}, orig.Columns[:len(orig.Columns)-1]...)
		tableMap[name] = shifted
		err := checkFit(b, filename)
		tableMap[name] = orig
		if err != nil {
			caught++
		} else {
			missed = append(missed, name)
		}
	}
	t.Logf("shifted: %d caught, %d missed: %v", caught, len(missed), missed)
}
