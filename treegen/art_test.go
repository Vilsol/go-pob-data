package treegen

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/pkg/errors"
	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/testdata"
)

type sprite struct {
	// Coords are raw: the export has a name with lists for values.
	Coords   map[string]json.RawMessage `json:"coords"`
	Filename string                     `json:"filename"`
	W        int                        `json:"w"`
	H        int                        `json:"h"`
}

type treeArt struct {
	Sprites map[string]map[string]sprite `json:"sprites"`
	Nodes   map[string]struct {
		Icon              string `json:"icon"`
		InactiveIcon      string `json:"inactiveIcon"`
		ActiveIcon        string `json:"activeIcon"`
		ActiveEffectImage string `json:"activeEffectImage"`
		IsKeystone        bool   `json:"isKeystone"`
		IsNotable         bool   `json:"isNotable"`
		IsMastery         bool   `json:"isMastery"`
	} `json:"nodes"`
	ExtraImages     map[string]struct{ Image string } `json:"extraImages"`
	ImageZoomLevels []float64                         `json:"imageZoomLevels"`
}

// last is the sprite at the zoom level the frontend uses, falling back like it does.
func (t *treeArt) last(group string) (sprite, bool) {
	zooms := t.Sprites[group]
	if s, ok := zooms[jsonFloat(t.ImageZoomLevels[len(t.ImageZoomLevels)-1])]; ok {
		return s, true
	}
	for _, s := range zooms {
		return s, true
	}
	return sprite{}, false
}

func jsonFloat(f float64) string {
	b, _ := json.Marshal(f)
	return string(b)
}

func readFiles(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	files := make(map[string][]byte)
	testza.AssertNoError(t, filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		files[rel], err = os.ReadFile(p)
		return errors.Wrap(err, "failed to read output")
	}))
	return files
}

func TestWriteArt(t *testing.T) {
	short := testdata.ShortGameVersion()
	if _, err := os.Stat(filepath.Join("data", short, "schema.min.json")); err != nil {
		t.Skipf("no committed schema for live version %s", short)
	}
	loader, err := extractor.GetBundleLoader(extractor.NewWebFS(extractor.PoE1, testdata.GameVersion()), extractor.IsNewHash(extractor.PoE1, short))
	testza.AssertNoError(t, err)

	dirs := []string{t.TempDir(), t.TempDir()}
	for _, dir := range dirs {
		_, err := Write(loader, short, dir)
		testza.AssertNoError(t, err)
	}

	first, second := readFiles(t, dirs[0]), readFiles(t, dirs[1])
	testza.AssertEqual(t, len(first), len(second))
	for name, b := range first {
		testza.AssertTrue(t, bytes.Equal(b, second[name]), "two runs wrote different "+name)
	}

	var ours treeArt
	testza.AssertNoError(t, json.Unmarshal(first["data.json"], &ours))

	t.Run("sheets", func(t *testing.T) {
		mw := imagick.NewMagickWand()
		defer mw.Destroy()
		for group, zooms := range ours.Sprites {
			for _, s := range zooms {
				u, err := url.Parse(s.Filename)
				testza.AssertNoError(t, err)
				b, ok := first[filepath.Join("assets", path.Base(u.Path))]
				if !ok {
					t.Errorf("%s: no sheet %s", group, s.Filename)
					continue
				}
				mw.Clear()
				testza.AssertNoError(t, mw.PingImageBlob(b))
				testza.AssertEqual(t, []int{s.W, s.H}, []int{int(mw.GetImageWidth()), int(mw.GetImageHeight())}, group)
				for name, raw := range s.Coords {
					var c struct{ X, Y, W, H int }
					testza.AssertNoError(t, json.Unmarshal(raw, &c))
					inside := c.X >= 0 && c.Y >= 0 && c.W > 0 && c.H > 0 && c.X+c.W <= s.W && c.Y+c.H <= s.H
					testza.AssertTrue(t, inside, group+": "+name+" is outside its sheet")
				}
			}
		}
	})

	t.Run("node icons", func(t *testing.T) {
		has := func(group, name string) bool {
			s, ok := ours.last(group)
			_, found := s.Coords[name]
			return ok && found
		}
		for id, n := range ours.Nodes {
			var groups []string
			switch {
			case n.IsMastery:
				groups = []string{"mastery"}
			case n.IsKeystone:
				groups = []string{"keystoneActive", "keystoneInactive"}
			case n.IsNotable:
				groups = []string{"notableActive", "notableInactive"}
			default:
				groups = []string{"normalActive", "normalInactive"}
			}
			for _, g := range groups {
				testza.AssertTrue(t, n.Icon == "" || has(g, n.Icon), id+": no "+g+" "+n.Icon)
			}
			testza.AssertTrue(t, n.InactiveIcon == "" || has("masteryInactive", n.InactiveIcon) && has("masteryConnected", n.InactiveIcon), id+": no inactive mastery "+n.InactiveIcon)
			testza.AssertTrue(t, n.ActiveIcon == "" || has("masteryActiveSelected", n.ActiveIcon), id+": no active mastery "+n.ActiveIcon)
			testza.AssertTrue(t, n.ActiveEffectImage == "" || has("masteryActiveEffect", n.ActiveEffectImage), id+": no mastery effect "+n.ActiveEffectImage)
		}
	})

	t.Run("export coverage", func(t *testing.T) {
		f, err := os.Open(exportPath(short))
		if err != nil {
			t.Skipf("no committed export for %s", short)
		}
		defer f.Close()
		gz, err := gzip.NewReader(f)
		testza.AssertNoError(t, err)
		var export treeArt
		testza.AssertNoError(t, json.NewDecoder(gz).Decode(&export))

		for group := range export.Sprites {
			exp, _ := export.last(group)
			got, ok := ours.last(group)
			if !ok {
				t.Errorf("no sprite group %s", group)
				continue
			}
			for name := range exp.Coords {
				_, ok := got.Coords[name]
				testza.AssertTrue(t, ok, group+": no "+name)
			}
		}
		for key, img := range export.ExtraImages {
			testza.AssertEqual(t, img.Image, ours.ExtraImages[key].Image, "extra image "+key)
		}
	})
}
