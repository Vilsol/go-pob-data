package treegen

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/extractor"
)

// pobTreePath is PathOfBuilding-PoE2's src/TreeData/<version>/tree.json, frozen without its sprite
// sheet fields (assets, ddsCoords), which the generator does not produce.
func pobTreePath(short string) string {
	return filepath.Join("treegen", "testdata", "pob-tree-"+short+".json.gz")
}

// knownPoE2Mismatches are PoB-PoE2 values the generator doesn't reproduce, keyed by "<field> <id>".
var knownPoE2Mismatches = map[string]string{
	// PoB's statdesc.lua ignores the add_one handler the game applies.
	"nodes.stats 4367": "PoB ignores the add_one index handler",
	// PoB's statdesc.lua matches gem_quality variants outside gem quality.
	"nodes.stats 37778": "PoB renders a gem_quality variant",
}

func readGzipJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	f, err := os.Open(path)
	testza.AssertNoError(t, err)
	defer f.Close()
	r, err := gzip.NewReader(f)
	testza.AssertNoError(t, err)
	b, err := io.ReadAll(r)
	testza.AssertNoError(t, err)
	var v map[string]any
	testza.AssertNoError(t, json.Unmarshal(b, &v))
	return v
}

func jsonClose(a, b any) bool {
	switch av := a.(type) {
	case float64:
		bv, ok := b.(float64)
		return ok && (av == bv || math.Abs(av-bv) <= 1e-9*math.Max(math.Abs(av), math.Abs(bv)))
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !jsonClose(av[i], bv[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			w, ok := bv[k]
			if !ok || !jsonClose(v, w) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}

// diffPoE2Trees lists every "<field> <id>" where got and want disagree: per node field for nodes,
// per index for groups and classes, and per key elsewhere.
func diffPoE2Trees(got, want map[string]any) []string {
	var diffs []string
	keys := func(m ...map[string]any) []string {
		var out []string
		for _, mm := range m {
			for k := range mm {
				if !slices.Contains(out, k) {
					out = append(out, k)
				}
			}
		}
		sort.Strings(out)
		return out
	}
	for _, k := range keys(got, want) {
		switch k {
		case "nodes":
			gn, _ := got[k].(map[string]any)
			wn, _ := want[k].(map[string]any)
			for _, id := range keys(gn, wn) {
				g, _ := gn[id].(map[string]any)
				w, _ := wn[id].(map[string]any)
				if g == nil || w == nil {
					diffs = append(diffs, "nodes "+id)
					continue
				}
				for _, f := range keys(g, w) {
					if !jsonClose(g[f], w[f]) {
						diffs = append(diffs, "nodes."+f+" "+id)
					}
				}
			}
		case "groups", "classes":
			gl, _ := got[k].([]any)
			wl, _ := want[k].([]any)
			for i := range max(len(gl), len(wl)) {
				if i >= len(gl) || i >= len(wl) || !jsonClose(gl[i], wl[i]) {
					diffs = append(diffs, fmt.Sprintf("%s %d", k, i+1))
				}
			}
		default:
			if !jsonClose(got[k], want[k]) {
				diffs = append(diffs, k)
			}
		}
	}
	return diffs
}

// sortGroupOrbits sorts each group's orbits: PoB-PoE2 lists them in LuaJIT's pairs() order
// (passivetree.lua:1065-1067), the generator in ascending order.
func sortGroupOrbits(tree map[string]any) {
	groups, _ := tree["groups"].([]any)
	for _, g := range groups {
		if g, ok := g.(map[string]any); ok {
			if orbits, ok := g["orbits"].([]any); ok {
				slices.SortFunc(orbits, func(a, b any) int { return int(a.(float64) - b.(float64)) })
			}
		}
	}
}

func TestDiffPoE2Trees(t *testing.T) {
	want := map[string]any{
		"min_x":  1.0,
		"groups": []any{nil, map[string]any{"x": 2.0}},
		"nodes":  map[string]any{"1": map[string]any{"stats": []any{"a"}, "orbit": 1.0}, "2": map[string]any{}},
	}
	testza.AssertEqual(t, []string(nil), diffPoE2Trees(want, want))

	got := map[string]any{
		"min_x":  1.0 + 1e-12,
		"groups": []any{nil, map[string]any{"x": 2.1}},
		"nodes":  map[string]any{"1": map[string]any{"stats": []any{"b"}, "orbit": 1.0}, "3": map[string]any{}},
		"extra":  true,
	}
	testza.AssertEqual(t, []string{"extra", "groups 2", "nodes.stats 1", "nodes 2", "nodes 3"}, diffPoE2Trees(got, want))
}

func TestRoundTo(t *testing.T) {
	// Lua's floor(x * 100 + 0.5) rounds halves up, not away from zero.
	testza.AssertEqual(t, -10648.12, roundTo(-10648.125, 2))
	testza.AssertEqual(t, 10648.13, roundTo(10648.125, 2))
	testza.AssertEqual(t, 1.5, roundTo(1.5, 2))
}

func TestCalcOrbitAngles(t *testing.T) {
	a := calcOrbitAngles(12)
	testza.AssertLen(t, a, 13, "uniform orbits repeat the full turn")
	testza.AssertEqual(t, 0.0, a[0])
	testza.AssertTrue(t, math.Abs(a[12]-2*math.Pi) < 1e-12)
	testza.AssertLen(t, calcOrbitAngles(16), 16)
	testza.AssertTrue(t, math.Abs(calcOrbitAngles(16)[2]-math.Pi/4) < 1e-12)
	testza.AssertLen(t, calcOrbitAngles(1), 2)
}

func TestEscapeAndSanitise(t *testing.T) {
	testza.AssertEqual(t, "15% increased chance to Shock", escapeGGGString("15% increased chance to [Shock]"))
	testza.AssertEqual(t, "Feuer-Schaden", escapeGGGString("[Fire|Feuer]-[Damage|Schaden]"))
	testza.AssertEqual(t, "Fire", escapeGGGString("<rgb(255,0,0)>{Fire}"))

	testza.AssertEqual(t, "1 Helmet socket", sanitiseText("• 1 Helmet socket"))
	testza.AssertEqual(t, "a-b", sanitiseText("a–b"))
	testza.AssertEqual(t, "Mjolner", sanitiseText("Mjölner"))
	testza.AssertEqual(t, "x??", sanitiseText("xé"), "an unnamed character becomes one ? per byte")
	testza.AssertEqual(t, "ab", sanitiseText("a<b<c>>b"))
	testza.AssertEqual(t, "plain [x]", sanitiseText("plain [x]"))
}

func generatePoE2Live(t *testing.T) (*Tree2, string) {
	t.Helper()
	patch, err := extractor.LatestPatch(extractor.PoE2)
	if err != nil {
		t.Skipf("no live poe2 patch: %v", err)
	}
	short := strings.Join(strings.Split(patch, ".")[:2], ".")
	if _, err := os.Stat(pobTreePath(short)); err != nil {
		t.Skipf("no frozen PoB-PoE2 tree for live version %s", short)
	}

	loader, err := extractor.GetBundleLoader(extractor.NewWebFS(extractor.PoE2, patch), extractor.IsNewHash(extractor.PoE2, short))
	testza.AssertNoError(t, err)

	tree, err := GeneratePoE2(loader, short)
	testza.AssertNoError(t, err)
	return tree, short
}

func TestPoE2TreeAgainstPoB(t *testing.T) {
	tree, short := generatePoE2Live(t)

	b, err := tree.Marshal()
	testza.AssertNoError(t, err)
	var got map[string]any
	testza.AssertNoError(t, json.Unmarshal(b, &got))

	want := readGzipJSON(t, pobTreePath(short))
	sortGroupOrbits(got)
	sortGroupOrbits(want)
	diffs := diffPoE2Trees(got, want)
	for _, d := range diffs {
		if _, ok := knownPoE2Mismatches[d]; !ok {
			t.Errorf("%s disagrees with PoB-PoE2", d)
		}
	}
	for k := range knownPoE2Mismatches {
		testza.AssertTrue(t, slices.Contains(diffs, k), "known mismatch %s now agrees", k)
	}

	committed := filepath.Join("data", short, "tree", "data.json.gz")
	f, err := os.Open(committed)
	testza.AssertNoError(t, err)
	defer f.Close()
	r, err := gzip.NewReader(f)
	testza.AssertNoError(t, err)
	committedJSON, err := io.ReadAll(r)
	testza.AssertNoError(t, err)
	// The committed tree also carries the sprite sheets' coordinates, which GeneratePoE2 leaves out.
	var committedTree map[string]any
	testza.AssertNoError(t, json.Unmarshal(committedJSON, &committedTree))
	testza.AssertNotNil(t, committedTree["sprites"], "%s has no sprites", committed)
	delete(committedTree, "sprites")
	delete(committedTree, "imageZoomLevels")
	var generated map[string]any
	testza.AssertNoError(t, json.Unmarshal(b, &generated))
	testza.AssertTrue(t, reflect.DeepEqual(generated, committedTree), "%s is stale; regenerate it with -game poe2 -tree-only", committed)
}

func TestSolidRuns(t *testing.T) {
	// One row: a speckled shadow, a solid arc with a faint edge, a lone speckle, and a second arc.
	alphas := []uint8{0, 20, 0, 30, 60, 255, 255, 40, 0, 0, 20, 0, 200, 0}
	img := image.NewNRGBA(image.Rect(0, 0, len(alphas), 1))
	for x, a := range alphas {
		img.SetNRGBA(x, 0, color.NRGBA{A: a})
	}
	testza.AssertEqual(t, [][2]int{{3, 8}, {12, 13}}, solidRuns(img, image.Pt(1, 0), image.Pt(0, 0), len(alphas)))
	testza.AssertLen(t, opaqueRuns(img, image.Pt(1, 0), image.Pt(0, 0), len(alphas)), 4, "the speckles split opaqueRuns")
}

func TestLineCutterKeepsItsMaskRegion(t *testing.T) {
	// Two nested quarter arcs around the bottom right corner, each with a faint shadow, and a mask
	// lighting a band around each.
	const n = 24
	tex := image.NewNRGBA(image.Rect(0, 0, n, n))
	mask := image.NewNRGBA(image.Rect(0, 0, n, n))
	radius := func(x, y int) float64 { return math.Hypot(float64(n-1-x), float64(n-1-y)) }
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			for _, r := range []float64{8, 18} {
				d := math.Abs(radius(x, y) - r)
				if d <= 1 {
					tex.SetNRGBA(x, y, color.NRGBA{R: 200, A: 255})
				} else if d <= 2.5 {
					tex.SetNRGBA(x, y, color.NRGBA{A: 30})
				}
				if d <= 3 {
					mask.SetNRGBA(x, y, color.NRGBA{G: 255, B: 255, A: 255})
				}
			}
			if mask.NRGBAAt(x, y).A == 0 {
				mask.SetNRGBA(x, y, color.NRGBA{A: 255})
			}
		}
	}

	owners := maskOwners(mask)
	s := &sheet{}
	cut := lineCutter{art: &art{img: tex}, owners: owners, used: make(map[int]string)}
	arcs := solidRuns(tex, image.Pt(1, 0), image.Pt(0, n-1), n)
	testza.AssertLen(t, arcs, 2)
	for i, arc := range arcs {
		testza.AssertNoError(t, cut.add(s, strconv.Itoa(i), image.Rect(arc[0], arc[0], n, n), image.Pt(1, 0), image.Pt(arc[0], n-1), arc[1]-arc[0]))
	}

	outer := s.entries[0].art.img
	off := arcs[0][0]
	region := owners[(n-1)*n+(arcs[0][0]+arcs[0][1])/2]
	drawn := 0
	for y := off; y < n; y++ {
		for x := off; x < n; x++ {
			a := outer.NRGBAAt(x-off, y-off).A
			if owners[y*n+x] != region {
				testza.AssertEqual(t, uint8(0), a, "pixel %d,%d of the outer arc's sprite is outside its mask region", x, y)
			} else if a == tex.NRGBAAt(x, y).A && a > 0 {
				drawn++
			}
		}
	}
	testza.AssertTrue(t, drawn > 0, "the outer arc's sprite keeps its own pixels")
	testza.AssertNotNil(t, cut.add(s, "again", image.Rect(0, 0, n, n), image.Pt(1, 0), image.Pt(arcs[1][0], n-1), arcs[1][1]-arcs[1][0]), "two lines may not share a mask region")
}

// TestLineSpritesHoldOneArc reads the committed PoE2 line sheet: along each orbit sprite's bottom
// row and right column, which every nested arc of the curves textures crosses, only its own arc
// may be solid.
func TestLineSpritesHoldOneArc(t *testing.T) {
	var tree struct {
		Sprites map[string]map[string]struct {
			Coords   map[string]Coord `json:"coords"`
			Filename string           `json:"filename"`
		} `json:"sprites"`
	}
	b, err := json.Marshal(readGzipJSON(t, filepath.Join("data", "0.5", "tree", "data.json.gz")))
	testza.AssertNoError(t, err)
	testza.AssertNoError(t, json.Unmarshal(b, &tree))
	testza.AssertLen(t, tree.Sprites["line"], 1)
	for _, sp := range tree.Sprites["line"] {
		f, err := os.Open(filepath.Join("data", "0.5", "tree", "assets", path.Base(sp.Filename)))
		testza.AssertNoError(t, err)
		decoded, err := png.Decode(f)
		f.Close()
		if err != nil {
			t.Fatalf("%s: %v (an LFS pointer if git lfs did not fetch it)", sp.Filename, err)
		}
		img := image.NewNRGBA(decoded.Bounds())
		draw.Draw(img, img.Rect, decoded, image.Point{}, draw.Src)

		orbits := 0
		for name, c := range sp.Coords {
			if !strings.Contains(name, "Orbit") {
				continue
			}
			orbits++
			bottom := solidRuns(img, image.Pt(1, 0), image.Pt(c.X, c.Y+c.H-1), c.W)
			right := solidRuns(img, image.Pt(0, 1), image.Pt(c.X+c.W-1, c.Y), c.H)
			testza.AssertLen(t, bottom, 1, "%s crosses its bottom row %d times", name, len(bottom))
			testza.AssertLen(t, right, 1, "%s crosses its right column %d times", name, len(right))
		}
		testza.AssertEqual(t, 81, orbits)
	}
}
