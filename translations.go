package main

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	st "github.com/Vilsol/go-pob-data/stat_translations"
)

var translations = []string{
	"Metadata/StatDescriptions/stat_descriptions.txt",
	"Metadata/StatDescriptions/passive_skill_aura_stat_descriptions.txt",
	"Metadata/StatDescriptions/passive_skill_stat_descriptions.txt",
}

func extractTranslations(loader fs.FS, gameVersion string) error {
	for _, translation := range translations {
		parser := st.NewTranslationParser(loader)
		if err := parser.ParseFile(strings.ToLower(translation)); err != nil {
			return errors.Wrap(err, "failed to parse translation file")
		}
		if err := parser.SaveTo(filepath.Join("data", gameVersion, "stat_translations"), strings.Split(filepath.Base(translation), ".")[0]); err != nil {
			return errors.Wrap(err, "failed to save translation files")
		}
	}

	return nil
}
