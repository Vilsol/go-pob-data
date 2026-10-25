package main

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/loader"
)

type Metadata struct {
	Versions []string `json:"versions"`
}

func generateMeta() error {
	dir, err := os.ReadDir("data")
	if err != nil {
		return errors.Wrap(err, "failed reading data directory")
	}

	metadata := &Metadata{}
	for _, entry := range dir {
		if entry.IsDir() {
			metadata.Versions = append(metadata.Versions, entry.Name())

			manifest, err := manifestJSON(filepath.Join("data", entry.Name()))
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join("data", entry.Name(), "manifest.json"), manifest, 0o644); err != nil {
				return errors.Wrap(err, "failed writing manifest")
			}
		}
	}

	bs, err := json.Marshal(metadata)
	if err != nil {
		return errors.Wrap(err, "failed marshaling metadata")
	}

	return errors.Wrap(os.WriteFile("data/meta.json", bs, 0o644), "failed writing metadata")
}

// manifestJSON lists what versionDir publishes. Every table and translation has a json.gz, whichever
// tuple layouts it also has.
func manifestJSON(versionDir string) ([]byte, error) {
	manifest := loader.Manifest{
		Tables:           []string{},
		StatTranslations: []string{},
		Languages:        []string{},
		Files:            []string{},
	}

	tables, err := filepath.Glob(filepath.Join(versionDir, "raw", "*.json.gz"))
	if err != nil {
		return nil, errors.Wrap(err, "failed listing tables")
	}
	for _, table := range tables {
		manifest.Tables = append(manifest.Tables, strings.TrimSuffix(filepath.Base(table), ".json.gz"))
	}

	tuples, err := filepath.Glob(filepath.Join(versionDir, "raw", "*.t-*.msgpack.br"))
	if err != nil {
		return nil, errors.Wrap(err, "failed listing tuple assets")
	}
	manifest.JSONOnly = len(tables) > 0 && len(tuples) == 0

	languages, err := os.ReadDir(filepath.Join(versionDir, "stat_translations"))
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrap(err, "failed listing languages")
	}
	for _, language := range languages {
		if language.IsDir() {
			manifest.Languages = append(manifest.Languages, language.Name())
		}
	}

	// PoE2 nests its per-skill descriptions under specific_skill_stat_descriptions/.
	enDir := filepath.Join(versionDir, "stat_translations", "en")
	err = filepath.WalkDir(enDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".json.gz") {
			return err
		}
		rel, err := filepath.Rel(enDir, path)
		if err != nil {
			return errors.Wrap(err, "failed resolving translation path")
		}
		manifest.StatTranslations = append(manifest.StatTranslations, strings.TrimSuffix(filepath.ToSlash(rel), ".json.gz"))
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrap(err, "failed listing translations")
	}
	slices.Sort(manifest.StatTranslations)

	filesDir := filepath.Join(versionDir, "files")
	err = filepath.WalkDir(filesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(filesDir, path)
		if err != nil {
			return errors.Wrap(err, "failed resolving file path")
		}
		manifest.Files = append(manifest.Files, strings.TrimSuffix(filepath.ToSlash(rel), ".br"))
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrap(err, "failed listing files")
	}

	slices.Sort(manifest.Files)

	b, err := json.MarshalIndent(manifest, "", "\t")
	return append(b, '\n'), errors.Wrap(err, "failed marshaling manifest")
}

func IsNewHash(game extractor.Game, gameVersion string) bool {
	return extractor.IsNewHash(game, gameVersion)
}

func GetFileExtension(game extractor.Game, gameVersion string) string {
	return extractor.GetFileExtension(game, gameVersion)
}
