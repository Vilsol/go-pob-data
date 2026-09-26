package treegen

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/testdata"
)

// The extractor reads schemas from data/ relative to the working directory.
func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

// knownMismatches are export fields the generator can't reproduce yet, with the reason.
var knownMismatches = map[string]string{
	"nodes.reminderText":                "phase 3: the export merges stat description reminder text into ReminderText rows",
	"nodes.stats":                       "phase 3: stat text rendering",
	"nodes.masteryEffects.stats":        "phase 3: stat text rendering",
	"nodes.masteryEffects.reminderText": "phase 3: stat description reminder text",
}

func generateLive(t *testing.T) (*Tree, string) {
	t.Helper()
	short := testdata.ShortGameVersion()
	if _, err := os.Stat(filepath.Join("data", short, "schema.min.json")); err != nil {
		t.Skipf("no committed schema for live version %s", short)
	}

	loader, err := extractor.GetBundleLoader(extractor.NewWebFS(testdata.GameVersion()), extractor.IsNewHash(short))
	testza.AssertNoError(t, err)

	tree, err := Generate(loader, short)
	testza.AssertNoError(t, err)
	return tree, short
}

func TestGenerateDeterministic(t *testing.T) {
	first, short := generateLive(t)
	a, err := first.Marshal()
	testza.AssertNoError(t, err)

	loader, err := extractor.GetBundleLoader(extractor.NewWebFS(testdata.GameVersion()), extractor.IsNewHash(short))
	testza.AssertNoError(t, err)
	second, err := Generate(loader, short)
	testza.AssertNoError(t, err)
	b, err := second.Marshal()
	testza.AssertNoError(t, err)

	testza.AssertTrue(t, bytes.Equal(a, b), "two generations differ")
}

type mismatch struct {
	gen, exp any
	id       string
}

type differ struct {
	matched    map[string]int
	total      map[string]int
	mismatches map[string][]mismatch
}

func (d *differ) compare(field, id string, gen, exp any) {
	// Equal-length lists of objects are diffed per element field, e.g. nodes.masteryEffects.effect.
	genList, genOk := gen.([]any)
	expList, expOk := exp.([]any)
	if genOk && expOk && len(genList) == len(expList) && len(genList) > 0 {
		if _, isObj := expList[0].(map[string]any); isObj {
			for i := range genList {
				g, _ := genList[i].(map[string]any)
				e, _ := expList[i].(map[string]any)
				d.compareObjects(field, fmt.Sprintf("%s[%d]", id, i), g, e)
			}
			return
		}
	}

	d.total[field]++
	if reflect.DeepEqual(gen, exp) {
		d.matched[field]++
		return
	}
	d.mismatches[field] = append(d.mismatches[field], mismatch{id: id, gen: gen, exp: exp})
}

// compareObjects diffs every key present on either side, as prefix.<key>.
func (d *differ) compareObjects(prefix, id string, gen, exp map[string]any) {
	keys := make(map[string]bool)
	for k := range gen {
		keys[k] = true
	}
	for k := range exp {
		keys[k] = true
	}
	for k := range keys {
		d.compare(prefix+"."+k, id, gen[k], exp[k])
	}
}

func decode(t *testing.T, r io.Reader) map[string]any {
	t.Helper()
	var out map[string]any
	testza.AssertNoError(t, json.NewDecoder(r).Decode(&out))
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// TestDiffAgainstExport compares the generated tree with the committed skilltree-export data.json
// field by field and fails on any mismatch without an entry in knownMismatches.
func TestDiffAgainstExport(t *testing.T) {
	tree, short := generateLive(t)

	exportFile, err := os.Open(filepath.Join("data", short, "tree", "data.json.gz"))
	if err != nil {
		t.Skipf("no committed export for %s", short)
	}
	defer exportFile.Close()
	gz, err := gzip.NewReader(exportFile)
	testza.AssertNoError(t, err)
	exp := decode(t, gz)

	b, err := tree.Marshal()
	testza.AssertNoError(t, err)
	gen := decode(t, bytes.NewReader(b))

	d := &differ{matched: map[string]int{}, total: map[string]int{}, mismatches: map[string][]mismatch{}}

	genNodes, expNodes := gen["nodes"].(map[string]any), exp["nodes"].(map[string]any)
	for id, n := range expNodes {
		if _, ok := genNodes[id]; !ok {
			d.compare("nodes", id, nil, "present")
			continue
		}
		d.compareObjects("nodes", id, genNodes[id].(map[string]any), n.(map[string]any))
	}
	// The export lacks the out-of-graph ascendancy nodes; any other node it lacks is a mismatch.
	generatedOnly := make([]string, 0)
	for id, n := range genNodes {
		if _, ok := expNodes[id]; ok {
			continue
		}
		node := n.(map[string]any)
		_, positioned := node["group"]
		if _, ascendancy := node["ascendancyName"]; positioned || !ascendancy {
			d.compare("nodes", id, "present", nil)
			continue
		}
		generatedOnly = append(generatedOnly, id)
	}
	sort.Strings(generatedOnly)

	genGroups, expGroups := gen["groups"].(map[string]any), exp["groups"].(map[string]any)
	for id, g := range expGroups {
		gg, _ := genGroups[id].(map[string]any)
		d.compareObjects("groups", id, gg, g.(map[string]any))
	}
	for id := range genGroups {
		if _, ok := expGroups[id]; !ok {
			d.compare("groups", id, "present", nil)
		}
	}

	d.compareObjects("constants", "", gen["constants"].(map[string]any), exp["constants"].(map[string]any))
	d.compareObjects("points", "", gen["points"].(map[string]any), exp["points"].(map[string]any))

	for _, key := range []string{"tree", "jewelSlots", "min_x", "min_y", "max_x", "max_y"} {
		d.compare(key, "", gen[key], exp[key])
	}
	for _, key := range []string{"classes", "alternate_ascendancies"} {
		genList, _ := gen[key].([]any)
		expList, _ := exp[key].([]any)
		for i := range max(len(genList), len(expList)) {
			var g, e map[string]any
			if i < len(genList) {
				g = genList[i].(map[string]any)
			}
			if i < len(expList) {
				e = expList[i].(map[string]any)
			}
			d.compareObjects(key, fmt.Sprint(i), g, e)
		}
	}

	var report strings.Builder
	fmt.Fprintf(&report, "generated out-of-graph nodes the export lacks: %d\n", len(generatedOnly))
	for _, id := range generatedOnly {
		fmt.Fprintf(&report, "  %s: %s\n", id, compact(genNodes[id]))
	}
	fmt.Fprintf(&report, "%-40s %s\n", "field", "matched/total")
	for _, field := range sortedKeys(d.total) {
		fmt.Fprintf(&report, "%-40s %d/%d\n", field, d.matched[field], d.total[field])
	}

	for _, field := range sortedKeys(d.mismatches) {
		ms := d.mismatches[field]
		slices.SortFunc(ms, func(a, b mismatch) int { return strings.Compare(a.id, b.id) })

		reason, known := knownMismatches[field]
		if !known {
			reason = "UNEXPLAINED"
			t.Fail()
		}
		fmt.Fprintf(&report, "\n%s: %d mismatches (%s)\n", field, len(ms), reason)
		for _, m := range ms {
			fmt.Fprintf(&report, "  %s: generated=%s export=%s\n", m.id, compact(m.gen), compact(m.exp))
		}
	}

	t.Log("\n" + report.String())
}

func compact(v any) string {
	b, _ := json.Marshal(v)
	if len(b) > 300 {
		return string(b[:300]) + "..."
	}
	return string(b)
}
