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

const (
	// staleAbyssal: the export predates a patch that reworked these Abyssal bloodline nodes. The
	// generated text is what PoB's newer 3.29.1 export shows.
	staleAbyssal = "export predates the patch that reworked this Abyssal node"
	// staleReflect: no live description file has the export's "... from your Hits cannot be
	// Reflected"; the live text is PoB's newer 3.29.1 export's "Prevent +X% of Reflected ...".
	staleReflect = "export text is from before the reflect wording changed"
	// swappedOrder: the export lists two adjacent descriptions of the override file (critical
	// strike chance/multiplier and zombies/spectres in passive_skill_stat_descriptions.txt,
	// Inquisitor's two aura lines in passive_skill_aura_stat_descriptions.txt) in the opposite of
	// their file order, while every other pair follows file order. No rule in the live files
	// reproduces it; PoB's 3.29.1 export has the same order.
	swappedOrder = "export swaps two adjacent descriptions against the live file order"
)

// knownMismatches are export values the generator doesn't reproduce, keyed by "<field> <id>", with the reason.
var knownMismatches = map[string]string{
	"nodes.stats 16287":        staleAbyssal,
	"nodes.stats 19208":        staleAbyssal,
	"nodes.reminderText 19208": staleAbyssal,
	"nodes.stats 28609":        staleAbyssal,

	"nodes.masteryEffects.stats 12169[0]": staleReflect,
	"nodes.masteryEffects.stats 12239[0]": staleReflect,
	"nodes.masteryEffects.stats 12503[0]": staleReflect,
	"nodes.masteryEffects.stats 12873[0]": staleReflect,
	"nodes.masteryEffects.stats 13387[1]": staleReflect,
	"nodes.masteryEffects.stats 17127[0]": staleReflect,
	"nodes.masteryEffects.stats 20675[0]": staleReflect,
	"nodes.masteryEffects.stats 31818[0]": staleReflect,
	"nodes.masteryEffects.stats 33657[0]": staleReflect,
	"nodes.masteryEffects.stats 34317[1]": staleReflect,
	"nodes.masteryEffects.stats 42792[0]": staleReflect,
	"nodes.masteryEffects.stats 44298[1]": staleReflect,
	"nodes.masteryEffects.stats 48660[1]": staleReflect,
	"nodes.masteryEffects.stats 52220[0]": staleReflect,
	"nodes.masteryEffects.stats 5348[1]":  staleReflect,
	"nodes.masteryEffects.stats 5726[1]":  staleReflect,
	"nodes.masteryEffects.stats 64042[0]": staleReflect,

	"nodes.stats 6064":  swappedOrder,
	"nodes.stats 9014":  swappedOrder,
	"nodes.stats 28754": swappedOrder,
	"nodes.stats 30471": swappedOrder,
	"nodes.stats 19897": swappedOrder,
	"nodes.stats 40059": swappedOrder,
}

// exportPath is the frozen skilltree-export data.json of a version; data/<version>/tree now holds our own output.
func exportPath(short string) string {
	return filepath.Join("treegen", "testdata", "export-"+short+".json.gz")
}

func generateLive(t *testing.T) (*Tree, string) {
	t.Helper()
	short := testdata.ShortGameVersion()
	if _, err := os.Stat(filepath.Join("data", short, "schema.min.json")); err != nil {
		t.Skipf("no committed schema for live version %s", short)
	}

	loader, err := extractor.GetBundleLoader(extractor.NewWebFS(extractor.PoE1, testdata.GameVersion()), extractor.IsNewHash(extractor.PoE1, short))
	testza.AssertNoError(t, err)

	tree, err := Generate(loader, short)
	testza.AssertNoError(t, err)
	return tree, short
}

func TestGenerateDeterministic(t *testing.T) {
	first, short := generateLive(t)
	a, err := first.Marshal()
	testza.AssertNoError(t, err)

	loader, err := extractor.GetBundleLoader(extractor.NewWebFS(extractor.PoE1, testdata.GameVersion()), extractor.IsNewHash(extractor.PoE1, short))
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

// TestDiffAgainstExport compares the generated tree with the frozen skilltree-export data.json
// field by field and fails on any mismatch without an entry in knownMismatches.
func TestDiffAgainstExport(t *testing.T) {
	tree, short := generateLive(t)

	exportFile, err := os.Open(exportPath(short))
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

	undescribed := make(map[string][]string)
	for _, n := range tree.Nodes.List {
		undescribed[nodeID(n.Skill)] = n.Undescribed
	}

	hit := make(map[string]bool)
	for _, field := range sortedKeys(d.mismatches) {
		ms := d.mismatches[field]
		slices.SortFunc(ms, func(a, b mismatch) int { return strings.Compare(a.id, b.id) })

		fmt.Fprintf(&report, "\n%s: %d mismatches\n", field, len(ms))
		for _, m := range ms {
			key := field + " " + m.id
			reason, known := knownMismatches[key]
			hit[key] = true
			if !known {
				reason = "UNEXPLAINED"
				t.Fail()
			}
			if u := undescribed[m.id]; field == "nodes.stats" && len(u) > 0 {
				reason += fmt.Sprintf(", stats without a description: %v", u)
			}
			fmt.Fprintf(&report, "  %s (%s): generated=%s export=%s\n", m.id, reason, compact(m.gen), compact(m.exp))
		}
	}
	for _, key := range sortedKeys(knownMismatches) {
		if !hit[key] {
			t.Errorf("known mismatch %q no longer happens", key)
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

// TestNodePositions checks that every node is either fully placed in a group that lists it or has
// no position at all, and that nothing references a missing or positionless node.
func TestNodePositions(t *testing.T) {
	tree, _ := generateLive(t)
	b, err := tree.Marshal()
	testza.AssertNoError(t, err)
	gen := decode(t, bytes.NewReader(b))
	nodes, groups := gen["nodes"].(map[string]any), gen["groups"].(map[string]any)

	positionKeys := []string{"group", "orbit", "orbitIndex", "out", "in"}
	positioned := make(map[string]bool)
	for id, n := range nodes {
		node := n.(map[string]any)
		present := 0
		for _, k := range positionKeys {
			if _, ok := node[k]; ok {
				present++
			}
		}
		switch present {
		case 0:
			continue
		case len(positionKeys):
			positioned[id] = true
		default:
			t.Errorf("node %s has a partial position", id)
			continue
		}
		if id == "root" {
			continue
		}
		group, ok := groups[fmt.Sprint(node["group"])].(map[string]any)
		if !ok || !slices.Contains(group["nodes"].([]any), any(id)) {
			t.Errorf("node %s is not listed by its group %v", id, node["group"])
		}
	}

	ref := func(from, id string) {
		if !positioned[id] {
			t.Errorf("%s references %s, which is missing or has no position", from, id)
		}
	}
	for id, n := range nodes {
		node := n.(map[string]any)
		for _, k := range []string{"out", "in"} {
			edges, _ := node[k].([]any)
			for _, e := range edges {
				ref("node "+id+" "+k, e.(string))
			}
		}
		if ej, ok := node["expansionJewel"].(map[string]any); ok {
			ref("node "+id+" expansionJewel.proxy", ej["proxy"].(string))
			if parent, ok := ej["parent"].(string); ok {
				ref("node "+id+" expansionJewel.parent", parent)
			}
		}
	}
	for id, g := range groups {
		for _, n := range g.(map[string]any)["nodes"].([]any) {
			ref("group "+id, n.(string))
		}
	}
	for _, s := range gen["jewelSlots"].([]any) {
		ref("jewelSlots", fmt.Sprint(s))
	}

	// Forbidden Flame/Flesh and Forbidden Tattoo ascendancy passives outside the graph.
	for _, id := range []string{"27602", "57568", "52435", "42469", "18054", "19355", "5029"} {
		node, ok := nodes[id].(map[string]any)
		if !ok || positioned[id] || node["ascendancyName"] == nil {
			t.Errorf("node %s should be an out-of-graph ascendancy node", id)
		}
	}
}
