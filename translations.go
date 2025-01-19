package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/extractor"
	st "github.com/Vilsol/go-pob-data/stat_translations"
)

var translations = []string{
	"Metadata/StatDescriptions/stat_descriptions.txt",
	"Metadata/StatDescriptions/passive_skill_aura_stat_descriptions.txt",
	"Metadata/StatDescriptions/passive_skill_stat_descriptions.txt",
}

func extractTranslations(gamePath string, gameVersion string) error {
	if _, err := os.Stat(filepath.Join(gamePath, "Bundles2", "_.index.bin")); err != nil {
		return errors.Wrap(err, "could not find _.index.bin")
	}

	extractor.LoadParser(gameVersion)
	loader, err := extractor.GetBundleLoader(os.DirFS(gamePath), IsNewHash(gameVersion))
	if err != nil {
		return errors.Wrap(err, "could not initialize bundle loader")
	}

	for _, translation := range translations {
		parser := st.NewTranslationParser(loader)
		if err := parser.ParseFile(translation); err != nil {
			return errors.Wrap(err, "failed to parse translation file")
		}
		if err := parser.SaveTo(filepath.Join("data", gameVersion, "stat_translations"), strings.Split(filepath.Base(translation), ".")[0]); err != nil {
			return errors.Wrap(err, "failed to save translation files")
		}
	}

	return nil
}
