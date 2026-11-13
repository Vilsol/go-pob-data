package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/tinylib/msgp/msgp"

	"github.com/Vilsol/go-pob-data/loader"
)

// Every monster monster_varieties.txt names has its .ot template published, and so has every ancestor
// along its extends lines, wherever its MonsterVarieties row is extracted with the list.
func TestMonsterTemplatesPublished(t *testing.T) {
	const version = "3.29"

	varieties, err := readRows[struct {
		ID                   string `json:"Id"`
		BaseMonsterTypeIndex string `json:"BaseMonsterTypeIndex"`
	}](version, "MonsterVarieties")
	if err != nil {
		t.Fatal(err)
	}
	objectTypes := make(map[string]string, len(varieties))
	for _, variety := range varieties {
		objectTypes[variety.ID] = variety.BaseMonsterTypeIndex
	}

	ot := readFileBundle(t, filepath.Join("data", version), ".ot")

	checked := map[string]bool{"nothing": true}
	var check func(template string)
	check = func(template string) {
		if checked[template] {
			return
		}
		checked[template] = true

		text, ok := ot[template+".ot"]
		if !ok {
			t.Errorf("%s.ot is not published", template)
			return
		}
		for _, m := range extendsRegex.FindAllSubmatch(text, -1) {
			check(string(m[1]))
		}
	}

	for _, id := range strings.Fields(monsterVarieties) {
		objectType, ok := objectTypes[id]
		if !ok {
			t.Errorf("%s is not a monster variety", id)
			continue
		}
		check(objectType)
	}
}

// readFileBundle decodes versionDir's only ext bundle; nil when it has none.
func readFileBundle(t *testing.T, versionDir string, ext string) map[string][]byte {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(versionDir, loader.AssetPath("files", strings.TrimPrefix(ext, ".")+"-*")))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		return nil
	}
	if len(paths) > 1 {
		t.Fatalf("%s has more than one %s bundle: %v", versionDir, ext, paths)
	}

	b, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	if bytes.HasPrefix(b, []byte("version https://git-lfs")) {
		t.Skip("git lfs pointer, run git lfs pull")
	}
	b, err = io.ReadAll(brotli.NewReader(bytes.NewReader(b)))
	if err != nil {
		t.Fatal(err)
	}

	n, b, err := msgp.ReadMapHeaderBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	files := make(map[string][]byte, n)
	for range n {
		var path string
		if path, b, err = msgp.ReadStringBytes(b); err != nil {
			t.Fatal(err)
		}
		if files[path], b, err = msgp.ReadBytesBytes(b, nil); err != nil {
			t.Fatal(err)
		}
	}
	if len(b) > 0 {
		t.Fatalf("%s: %d trailing bytes", paths[0], len(b))
	}

	if want := loader.FileBundlePath(ext, loader.EncodeFileBundle(files)); filepath.Base(paths[0]) != want {
		t.Errorf("%s is named for other contents than %s", paths[0], want)
	}
	return files
}

// Every version's game text files are published once, in the bundle of their extension and not one at a
// time; the bundle is named for its contents.
func TestFileBundlesOnly(t *testing.T) {
	dirs, err := filepath.Glob("data/*")
	if err != nil {
		t.Fatal(err)
	}

	for _, versionDir := range dirs {
		t.Run(filepath.Base(versionDir), func(t *testing.T) {
			for _, ext := range []string{".ot", ".it", ".txt"} {
				for path := range readFileBundle(t, versionDir, ext) {
					if !strings.HasSuffix(path, ext) {
						t.Errorf("%s is in the %s bundle", path, ext)
					}
					if _, err := os.Stat(filepath.Join(versionDir, loader.FilePath(path))); err == nil {
						t.Errorf("%s is published one at a time as well as in its bundle", path)
					}
				}
			}
		})
	}
}
