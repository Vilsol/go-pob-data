package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"

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

	checked := map[string]bool{"nothing": true}
	var check func(template string)
	check = func(template string) {
		if checked[template] {
			return
		}
		checked[template] = true

		b, err := os.ReadFile(filepath.Join("data", version, loader.FilePath(template+".ot")))
		if err != nil {
			t.Errorf("%s.ot is not published", template)
			return
		}
		if bytes.HasPrefix(b, []byte("version https://git-lfs")) {
			t.Skip("git lfs pointer, run git lfs pull")
		}

		text, err := io.ReadAll(brotli.NewReader(bytes.NewReader(b)))
		if err != nil {
			t.Fatal(err)
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
