package main

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"testing/fstest"
	"unicode/utf16"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/loader"
)

// inTempDir runs the rest of the test in a scratch directory, since the extractors write data/ under the
// working directory.
func inTempDir(t *testing.T) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
}

func writeRows(t *testing.T, version string, table string, rows any) {
	t.Helper()
	path := filepath.Join("data", version, "raw", table+".json.gz")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := gzip.NewWriter(f)
	if err := json.NewEncoder(w).Encode(rows); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func utf16File(text string) *fstest.MapFile {
	b := []byte{0xff, 0xfe}
	for _, u := range utf16.Encode([]rune(text)) {
		b = append(b, byte(u), byte(u>>8))
	}
	return &fstest.MapFile{Data: b}
}

func TestTextFilesByGame(t *testing.T) {
	poe1, varieties1 := textFiles(extractor.PoE1)
	if !slices.Equal(poe1, textFilesToExport) || varieties1 != monsterVarieties {
		t.Error("poe1 must publish textFilesToExport and monsterVarieties unchanged")
	}

	poe2, varieties2 := textFiles(extractor.PoE2)
	if slices.Contains(poe2, "Metadata/StatDescriptions/skillpopup_stat_filters.txt") {
		t.Error("poe2 does not ship skillpopup_stat_filters.txt")
	}
	if len(poe2) != len(textFilesToExport)-1 || varieties2 != poe2MonsterVarieties || varieties2 == "" {
		t.Error("poe2 publishes the rest of textFilesToExport and its own monster list")
	}
}

// A template PoE2 names but does not ship is skipped, as PoB-PoE2 does; for PoE1 it fails the run.
func TestExtractTemplatesMissing(t *testing.T) {
	bundles := fstest.MapFS{
		"metadata/monsters/a.ot": utf16File("extends \"Metadata/Monsters/B\"\n"),
	}

	for _, c := range []struct {
		game    extractor.Game
		wantErr bool
	}{{extractor.PoE1, true}, {extractor.PoE2, false}} {
		t.Run(string(c.game), func(t *testing.T) {
			inTempDir(t)
			err := extractTemplates(bundles, c.game, "v", []string{"Metadata/Monsters/A"}, ".ot")
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, want an error: %v", err, c.wantErr)
			}
			if _, err := os.Stat(filepath.Join("data", "v", loader.FilePath("Metadata/Monsters/A.ot"))); err != nil {
				t.Errorf("A.ot was not published: %v", err)
			}
		})
	}
}

// PoE2's images are the illustrations its Characters table names, once each, and not the PoE1 set.
func TestImagesFor(t *testing.T) {
	images, err := imagesFor(extractor.PoE1, "3.29")
	if err != nil || !slices.Equal(images, imagesToExport) {
		t.Fatalf("poe1 must export imagesToExport unchanged: %v", err)
	}

	inTempDir(t)
	writeRows(t, "0.5", "Characters", []map[string]string{
		{"PassiveTreeImage": "Art/2DArt/BaseClassIllustrations/WarriorBaseIllustration.dds"},
		{"PassiveTreeImage": ""},
		{"PassiveTreeImage": "Art/2DArt/BaseClassIllustrations/MonkBaseIllustration.dds"},
		{"PassiveTreeImage": "Art/2DArt/BaseClassIllustrations/WarriorBaseIllustration.dds"},
	})
	images, err = imagesFor(extractor.PoE2, "0.5")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"Art/2DArt/BaseClassIllustrations/WarriorBaseIllustration.dds",
		"Art/2DArt/BaseClassIllustrations/MonkBaseIllustration.dds",
	}
	if !slices.Equal(images, want) {
		t.Errorf("images = %v, want %v", images, want)
	}
}

// A monster PoB-PoE2's templates name that the game no longer has is skipped for PoE2 only.
func TestExtractTextFilesMissingVariety(t *testing.T) {
	prev := poe2MonsterVarieties
	poe2MonsterVarieties = "Metadata/Monsters/Gone\nMetadata/Monsters/Here\n"
	t.Cleanup(func() { poe2MonsterVarieties = prev })

	inTempDir(t)
	writeRows(t, "0.5", "BaseItemTypes", []map[string]string{{"InheritsFrom": "Metadata/Items/Item"}})
	writeRows(t, "0.5", "MonsterVarieties", []map[string]string{
		{"Id": "Metadata/Monsters/Here", "BaseMonsterTypeIndex": "Metadata/Monsters/Here"},
	})
	bundles := fstest.MapFS{
		"metadata/characters/character.ot": utf16File(""),
		"metadata/monsters/monster.ot":     utf16File(""),
		"art/uiimages1.txt":                utf16File(""),
		"metadata/items/item.it":           utf16File(""),
		"metadata/monsters/here.ot":        utf16File("extends \"Metadata/Monsters/Monster\"\n"),
	}

	if err := extractTextFiles(bundles, extractor.PoE2, "0.5"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join("data", "0.5", loader.FilePath("Metadata/Monsters/Here.ot"))); err != nil {
		t.Errorf("Here.ot was not published: %v", err)
	}
}
