package main

import (
	"io/fs"
	"log/slog"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/extractor"
	st "github.com/Vilsol/go-pob-data/stat_translations"
)

func extractTranslations(loader fs.FS, game extractor.Game, gameVersion string) error {
	dir, ext, names := "metadata/statdescriptions/", ".txt", st.Files
	if game == extractor.PoE2 {
		dir, ext, names = "data/statdescriptions/", ".csd", slices.Clone(st.PoE2Files)
		specific, err := statDescriptionsUnder(loader, dir, st.SpecificSkillDir, ext)
		if err != nil {
			return err
		}
		names = append(names, specific...)
	}

	saved := 0
	for _, name := range names {
		if !statDescriptionShipped(loader, dir+name+ext) {
			// 3.30 dropped some of the listed files (tincture, village, variable_duration); a file the game ships
			// whose include is missing still fails below
			slog.Warn("skipping a stat description file the game does not ship", slog.String("path", dir+name+ext))
			continue
		}
		parser := st.NewTranslationParser(loader)
		if err := parser.ParseFile(dir + name + ext); err != nil {
			return errors.Wrap(err, "failed to parse translation file")
		}
		if err := parser.SaveTo(filepath.Join("data", gameVersion, "stat_translations"), name); err != nil {
			return errors.Wrap(err, "failed to save translation files")
		}
		saved++
	}
	if saved == 0 {
		return errors.New("no stat description file the list names is in the game")
	}

	return nil
}

// statDescriptionShipped reports whether the game has file, under its own extension or, for a PoE1 .txt, as the
// .csd 3.30 ships instead.
func statDescriptionShipped(loader fs.FS, file string) bool {
	if _, err := fs.Stat(loader, file); err == nil {
		return true
	}
	if path.Ext(file) != ".txt" {
		return false
	}
	_, err := fs.Stat(loader, strings.TrimSuffix(file, ".txt")+".csd")
	return err == nil
}

// statDescriptionsUnder lists the ext files under dir+sub, recursively, by name relative to dir without ext.
func statDescriptionsUnder(loader fs.FS, dir string, sub string, ext string) ([]string, error) {
	var names []string
	err := fs.WalkDir(loader, path.Join(dir, sub), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path.Ext(p) != ext {
			return err
		}
		names = append(names, strings.TrimSuffix(strings.TrimPrefix(p, dir), ext))
		return nil
	})
	if err == nil && len(names) == 0 {
		err = errors.New("no stat descriptions under " + dir + sub)
	}
	return names, errors.Wrap(err, "failed to list stat descriptions")
}
