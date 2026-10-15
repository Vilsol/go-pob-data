package main

import (
	"io/fs"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/extractor"
	st "github.com/Vilsol/go-pob-data/stat_translations"
)

func extractTranslations(loader fs.FS, _ extractor.Game, gameVersion string) error {
	for _, name := range st.Files {
		parser := st.NewTranslationParser(loader)
		if err := parser.ParseFile("metadata/statdescriptions/" + name + ".txt"); err != nil {
			return errors.Wrap(err, "failed to parse translation file")
		}
		if err := parser.SaveTo(filepath.Join("data", gameVersion, "stat_translations"), name); err != nil {
			return errors.Wrap(err, "failed to save translation files")
		}
	}

	return nil
}
