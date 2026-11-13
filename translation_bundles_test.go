package main

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

func versionsWithTranslations(t *testing.T) []string {
	t.Helper()
	dirs, err := filepath.Glob("data/*/stat_translations")
	if err != nil {
		t.Fatal(err)
	}
	return dirs
}

// Every version has one current bundle per language, and its per-file translation assets are exactly the
// ones go-pob reads one at a time, each the same file as its bundle entry. -update builds the bundles from
// the per-file assets and prunes the rest.
func TestTranslationBundlesCurrent(t *testing.T) {
	for _, dir := range versionsWithTranslations(t) {
		versionDir := filepath.Dir(dir)
		t.Run(filepath.Base(versionDir), func(t *testing.T) {
			if *update {
				if err := bundleTranslations(versionDir); err != nil {
					t.Fatal(err)
				}
			}

			perFile, err := readPerFileTranslations(versionDir)
			if err != nil {
				t.Fatal(err)
			}
			bundled, err := readTranslationBundles(versionDir)
			if err != nil {
				t.Fatal(err)
			}
			if len(bundled) == 0 {
				if len(perFile) > 0 {
					t.Errorf("per-file translations but no bundles: run with -update")
				}
				return
			}

			for lang, files := range bundled {
				path, want, err := translationBundle(lang, files)
				if err != nil {
					t.Fatal(err)
				}
				got, err := os.ReadFile(filepath.Join(versionDir, path))
				if err != nil || !bytes.Equal(got, want) {
					t.Errorf("%s is missing or stale: run with -update", path)
				}
			}

			if b, err := os.ReadFile(filepath.Join(versionDir, "manifest.json")); err == nil {
				var manifest loader.Manifest
				if err := json.Unmarshal(b, &manifest); err != nil {
					t.Fatal(err)
				}
				for lang, files := range bundled {
					for _, name := range manifest.StatTranslations {
						if _, ok := files[name]; !ok {
							t.Errorf("the %s bundle lacks %s, which manifest.json lists", lang, name)
						}
					}
				}
			} else if !os.IsNotExist(err) {
				t.Fatal(err)
			}

			names := perFileTranslations(versionDir, bundled["en"])
			if !slices.Contains(names, "stat_descriptions") {
				t.Errorf("en/stat_descriptions is not published per file")
			}
			keep := keptTranslationAssets(versionDir, names)
			err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() || filepath.Dir(path) == dir {
					return err
				}
				rel, err := filepath.Rel(versionDir, path)
				if err != nil {
					t.Fatal(err)
				}
				if !keep[filepath.ToSlash(rel)] {
					t.Errorf("%s is not read per file: run with -update", rel)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}

			if got := slices.Sorted(maps.Keys(perFile["en"])); !slices.Equal(got, names) {
				t.Errorf("per-file en translations %v, want %v", got, names)
			}
			for name, file := range perFile["en"] {
				if !reflect.DeepEqual(file, bundled["en"][name]) {
					t.Errorf("en/%s differs from its bundle entry", name)
				}
			}
		})
	}
}

// Released go-pob versions (v0.1.0 to v0.7.0, on go-pob-data v0.5.6 to v0.5.17) read English stat
// descriptions one file at a time in the t-c976d1d7 layout, and 3.29's stat_descriptions in the .min layout
// of its field list: these must stay published for as long as those versions are in use.
func TestReleasedPerFileTranslations(t *testing.T) {
	const released = "t-c976d1d7"
	releasedMin := map[string]string{"3.29": "min-ad0f83aa"}

	total := 0
	for _, dir := range versionsWithTranslations(t) {
		versionDir := filepath.Dir(dir)
		version := filepath.Base(versionDir)
		bundled, err := readTranslationBundles(versionDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(bundled) == 0 {
			continue
		}

		for _, name := range perFileTranslations(versionDir, bundled["en"]) {
			asset := filepath.Join(dir, "en", filepath.FromSlash(name))
			paths := []string{loader.AssetPath(asset, released)}
			if name == "stat_descriptions" && releasedMin[version] != "" {
				paths = append(paths, loader.AssetPath(asset, releasedMin[version]))
			}
			for _, path := range paths {
				total++
				if _, err := os.Stat(path); err != nil {
					t.Errorf("%s is gone, but released go-pob versions read it", path)
				}
			}
		}
	}
	t.Logf("%d per-file translation assets kept for released go-pob versions", total)
	if total == 0 {
		t.Fatal("no per-file translation assets found")
	}
}

// writeTranslations bundles every language, keeps per file only what the roots reach, and drops the rest.
func TestWriteTranslations(t *testing.T) {
	versionDir := filepath.Join(t.TempDir(), "3.99")
	stale := filepath.Join(versionDir, "stat_translations", "de", "stat_descriptions.json.gz")
	if err := os.MkdirAll(filepath.Dir(stale), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stale, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	file := func(includes ...string) *raw.TranslationFile {
		return &raw.TranslationFile{Descriptors: []*raw.StatTranslation{}, Includes: includes}
	}
	translations := map[string]map[string]*raw.TranslationFile{
		"en": {
			"stat_descriptions":                  file(),
			"gem_stat_descriptions":              file("Metadata/StatDescriptions/stat_descriptions.txt"),
			"skill_stat_descriptions":            file("Metadata/StatDescriptions/active_skill_gem_stat_descriptions.txt"),
			"atlas_stat_descriptions":            file(),
			"active_skill_gem_stat_descriptions": file("Metadata/StatDescriptions/gem_stat_descriptions.txt"),
		},
		"de": {"stat_descriptions": file(), "atlas_stat_descriptions": file()},
	}
	if err := writeTranslations(versionDir, translations); err != nil {
		t.Fatal(err)
	}

	bundled, err := readTranslationBundles(versionDir)
	if err != nil {
		t.Fatal(err)
	}
	encode := func(files map[string]map[string]*raw.TranslationFile) map[string][]byte {
		out := map[string][]byte{}
		for lang, byName := range files {
			b, err := loader.EncodeTranslationBundle(byName)
			if err != nil {
				t.Fatal(err)
			}
			out[lang] = b
		}
		return out
	}
	if got, want := encode(bundled), encode(translations); !reflect.DeepEqual(got, want) {
		t.Errorf("bundles hold %v, want %v", slices.Sorted(maps.Keys(bundled)), slices.Sorted(maps.Keys(translations)))
	}

	perFile, err := readPerFileTranslations(versionDir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"active_skill_gem_stat_descriptions", "gem_stat_descriptions", "skill_stat_descriptions", "stat_descriptions"}
	if got := slices.Sorted(maps.Keys(perFile["en"])); !slices.Equal(got, want) {
		t.Errorf("per-file en %v, want %v", got, want)
	}
	if _, ok := perFile["de"]; ok {
		t.Errorf("de is published per file")
	}
	if _, err := os.Stat(filepath.Dir(stale)); !os.IsNotExist(err) {
		t.Errorf("%s is left behind", filepath.Dir(stale))
	}

	// A second write with changed content replaces the bundle rather than adding one.
	translations["de"]["stat_descriptions"] = file("Metadata/StatDescriptions/atlas_stat_descriptions.txt")
	if err := writeTranslations(versionDir, translations); err != nil {
		t.Fatal(err)
	}
	if _, err := readTranslationBundles(versionDir); err != nil {
		t.Fatal(err)
	}
}

// A new version's directory holds nothing yet.
func TestWriteTranslationsNewVersion(t *testing.T) {
	versionDir := filepath.Join(t.TempDir(), "3.99")
	translations := map[string]map[string]*raw.TranslationFile{
		"de": {"stat_descriptions": {Descriptors: []*raw.StatTranslation{}}},
	}
	if err := writeTranslations(versionDir, translations); err != nil {
		t.Fatal(err)
	}
	if bundles, err := translationBundles(versionDir); err != nil || len(bundles["de"]) != 1 {
		t.Fatalf("bundles %v, %v", bundles, err)
	}
}

// Rebuilding a version whose per-file assets are pruned keeps every name its bundles already hold, and a
// per-file asset replaces its bundle entry.
func TestBundleTranslationsKeepsPrunedNames(t *testing.T) {
	versionDir := filepath.Join(t.TempDir(), "3.99")
	if err := os.MkdirAll(filepath.Join(versionDir, "stat_translations"), 0o755); err != nil {
		t.Fatal(err)
	}
	file := func(includes ...string) *raw.TranslationFile {
		return &raw.TranslationFile{Descriptors: []*raw.StatTranslation{}, Includes: includes}
	}
	full := map[string]map[string]*raw.TranslationFile{
		"en": {
			"stat_descriptions":       file(),
			"atlas_stat_descriptions": file(),
			"map_stat_descriptions":   file(),
		},
		"de": {"stat_descriptions": file(), "atlas_stat_descriptions": file()},
	}
	if err := writeTranslations(versionDir, full); err != nil {
		t.Fatal(err)
	}
	perFile, err := readPerFileTranslations(versionDir)
	if err != nil {
		t.Fatal(err)
	}
	if got := slices.Sorted(maps.Keys(perFile["en"])); !slices.Equal(got, []string{"stat_descriptions"}) {
		t.Fatalf("per-file en %v, want only stat_descriptions", got)
	}

	changed := file("Metadata/StatDescriptions/atlas_stat_descriptions.txt")
	path := filepath.Join(versionDir, "stat_translations", "en", "stat_descriptions")
	if err := writeTranslationAsset(loader.AssetPath(path, loader.TranslationBundleVariant()), changed); err != nil {
		t.Fatal(err)
	}

	if err := bundleTranslations(versionDir); err != nil {
		t.Fatal(err)
	}
	bundled, err := readTranslationBundles(versionDir)
	if err != nil {
		t.Fatal(err)
	}
	for lang, files := range full {
		if got, want := slices.Sorted(maps.Keys(bundled[lang])), slices.Sorted(maps.Keys(files)); !slices.Equal(got, want) {
			t.Errorf("%s bundle holds %v, want %v", lang, got, want)
		}
	}
	got, err := bundled["en"]["stat_descriptions"].MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}
	want, err := changed.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("en/stat_descriptions is not its per-file asset")
	}
}
