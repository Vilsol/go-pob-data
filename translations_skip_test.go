package main

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/raw"
	st "github.com/Vilsol/go-pob-data/stat_translations"
)

// a listed file the game does not ship is skipped; one it ships as .csd counts as shipped; a shipped file whose
// include is missing still fails
func TestStatDescriptionShipped(t *testing.T) {
	loader := fstest.MapFS{
		"metadata/statdescriptions/a.txt": {},
		"metadata/statdescriptions/b.csd": {},
	}
	for file, want := range map[string]bool{
		"metadata/statdescriptions/a.txt":       true,
		"metadata/statdescriptions/b.txt":       true,
		"metadata/statdescriptions/missing.txt": false,
	} {
		if got := statDescriptionShipped(loader, file); got != want {
			t.Errorf("%s: got %v, want %v", file, got, want)
		}
	}
}

func TestExtractTranslationsRefusesWhenNothingShips(t *testing.T) {
	if len(st.Files) == 0 {
		t.Skip("no PoE1 stat description list")
	}
	err := extractTranslations(fstest.MapFS{"metadata/other.txt": {}}, extractor.PoE1, "test")
	if err == nil || !strings.Contains(err.Error(), "no stat description file") {
		t.Fatalf("err = %v, want a refusal", err)
	}
}

// a language without a description in any file is not published; English always is
func TestDropUndescribedLanguages(t *testing.T) {
	file := func(n int) *raw.TranslationFile {
		return &raw.TranslationFile{Descriptors: make([]*raw.StatTranslation, n)}
	}
	parsed := map[string]map[string]*raw.TranslationFile{
		"en": {"a": file(0)},
		"de": {"a": file(0), "b": file(1)},
		"tr": {"a": file(0), "b": file(0)},
	}
	dropUndescribedLanguages(parsed)
	if _, ok := parsed["tr"]; ok {
		t.Error("tr is kept")
	}
	for _, lang := range []string{"en", "de"} {
		if _, ok := parsed[lang]; !ok {
			t.Errorf("%s is dropped", lang)
		}
	}
}
