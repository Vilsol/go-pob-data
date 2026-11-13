package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"golang.org/x/text/encoding/unicode"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/loader"

	_ "embed"
)

// textFilesToExport are published as they are, transcoded from the game's UTF-16 to UTF-8. The .ot files
// are not parsed: their sections nest and hold free-form values, so a typed form would either drop
// content or be a tree no simpler than the text.
var textFilesToExport = []string{
	"Metadata/Characters/Character.ot",
	"Metadata/Monsters/Monster.ot",
	"Metadata/StatDescriptions/skillpopup_stat_filters.txt",
	"art/uiimages1.txt",
}

// poe2TextFilesToExport is textFilesToExport less skillpopup_stat_filters.txt, which PoE2 does not ship.
var poe2TextFilesToExport = []string{
	"Metadata/Characters/Character.ot",
	"Metadata/Monsters/Monster.ot",
	"art/uiimages1.txt",
}

var extendsRegex = regexp.MustCompile(`(?m)^extends "([^"]+)"`)

// monsterVarieties are the monsters PoB's minion and spectre export templates name, one MonsterVarieties
// Id per line (mise run monster-varieties); their .ot templates are published.
//
//go:embed monster_varieties.txt
var monsterVarieties string

// poe2MonsterVarieties is monsterVarieties from PoB-PoE2's templates (mise run monster-varieties <pob> poe2).
//
//go:embed monster_varieties_poe2.txt
var poe2MonsterVarieties string

// textFiles are game's textFilesToExport and monsterVarieties.
func textFiles(game extractor.Game) ([]string, string) {
	if game == extractor.PoE2 {
		return poe2TextFilesToExport, poe2MonsterVarieties
	}
	return textFilesToExport, monsterVarieties
}

// extractTextFiles publishes textFilesToExport and the templates rows inherit from, with their ancestors
// through their extends lines: the item templates (.it) of every base's InheritsFrom, where bases read
// their tags, and the monster templates (.ot) of monsterVarieties' BaseMonsterTypeIndex, where those
// monsters read their base stats.
func extractTextFiles(bundles fs.FS, game extractor.Game, gameVersion string) error {
	files, varieties := textFiles(game)
	for _, file := range files {
		if _, err := extractTextFile(bundles, gameVersion, file); err != nil {
			if skipMissing && errors.Is(err, fs.ErrNotExist) {
				slog.Warn("skipping a missing file", slog.String("path", file))
				continue
			}
			return err
		}
	}

	bases, err := readRows[struct {
		InheritsFrom string `json:"InheritsFrom"`
	}](gameVersion, "BaseItemTypes")
	if err != nil {
		return err
	}
	itemRoots := make([]string, 0, len(bases))
	for _, base := range bases {
		itemRoots = append(itemRoots, base.InheritsFrom)
	}
	if err := extractTemplates(bundles, game, gameVersion, itemRoots, ".it"); err != nil {
		return err
	}

	rows, err := readRows[struct {
		ID                   string `json:"Id"`
		BaseMonsterTypeIndex string `json:"BaseMonsterTypeIndex"`
	}](gameVersion, "MonsterVarieties")
	if err != nil {
		return err
	}
	objectTypes := make(map[string]string, len(rows))
	for _, variety := range rows {
		objectTypes[variety.ID] = variety.BaseMonsterTypeIndex
	}
	monsterRoots := make([]string, 0)
	for _, id := range strings.Fields(varieties) {
		objectType, ok := objectTypes[id]
		// PoB-PoE2's templates follow its own patch, so PoE2 skips a variety the game no longer has.
		if !ok && (game == extractor.PoE2 || skipMissing) {
			slog.Warn("skipping a monster variety the game does not have", slog.String("id", id))
			continue
		}
		if !ok {
			return errors.New("monster_varieties.txt names a variety the game does not have: " + id)
		}
		monsterRoots = append(monsterRoots, objectType)
	}
	if err := extractTemplates(bundles, game, gameVersion, monsterRoots, ".ot"); err != nil {
		return err
	}

	return writeFileBundle(filepath.Join("data", gameVersion), ".ot")
}

// fileBundle is the LoadFileBundle asset of every file under versionDir's files/ with extension ext, at
// its loader.FileBundlePath; false when there is none.
func fileBundle(versionDir string, ext string) (string, []byte, bool, error) {
	filesDir := filepath.Join(versionDir, "files")
	files := map[string][]byte{}
	err := filepath.WalkDir(filesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ext+".br") {
			return err
		}
		rel, err := filepath.Rel(filesDir, path)
		if err != nil {
			return errors.Wrap(err, "failed resolving file path")
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "could not read "+path)
		}
		if bytes.HasPrefix(b, []byte("version https://git-lfs")) {
			return errors.New(path + " is a git lfs pointer, run git lfs pull")
		}
		text, err := io.ReadAll(brotli.NewReader(bytes.NewReader(b)))
		if err != nil {
			return errors.Wrap(err, "could not decompress "+path)
		}
		files[strings.TrimSuffix(filepath.ToSlash(rel), ".br")] = text
		return nil
	})
	if os.IsNotExist(err) || len(files) == 0 {
		return "", nil, false, nil
	}
	if err != nil {
		return "", nil, false, errors.Wrap(err, "failed listing "+ext+" files")
	}

	encoded := loader.EncodeFileBundle(files)
	var buf bytes.Buffer
	w := brotli.NewWriter(&buf)
	if _, err := w.Write(encoded); err != nil {
		return "", nil, false, errors.Wrap(err, "could not compress the "+ext+" bundle")
	}
	if err := w.Close(); err != nil {
		return "", nil, false, errors.Wrap(err, "could not compress the "+ext+" bundle")
	}
	return loader.FileBundlePath(ext, encoded), buf.Bytes(), true, nil
}

// staleFileBundles are versionDir's ext bundles other than keep.
func staleFileBundles(versionDir string, ext string, keep string) ([]string, error) {
	paths, err := filepath.Glob(filepath.Join(versionDir, loader.AssetPath("files", strings.TrimPrefix(ext, ".")+"-*")))
	if err != nil {
		return nil, errors.Wrap(err, "failed listing file bundles")
	}
	return slices.DeleteFunc(paths, func(path string) bool {
		return path == filepath.Join(versionDir, keep)
	}), nil
}

// writeFileBundle writes versionDir's ext bundle and removes the bundles it replaces.
func writeFileBundle(versionDir string, ext string) error {
	path, b, ok, err := fileBundle(versionDir, ext)
	if err != nil || !ok {
		return err
	}
	stale, err := staleFileBundles(versionDir, ext, path)
	if err != nil {
		return err
	}
	for _, old := range stale {
		if err := os.Remove(old); err != nil {
			return errors.Wrap(err, "could not remove "+old)
		}
	}
	return errors.Wrap(os.WriteFile(filepath.Join(versionDir, path), b, 0o644), "could not write the "+ext+" bundle")
}

// extractTemplates publishes the templates of roots and of their ancestors, up to "nothing" or a template
// textFiles already publishes (Metadata/Monsters/Monster).
func extractTemplates(bundles fs.FS, game extractor.Game, gameVersion string, roots []string, ext string) error {
	files, _ := textFiles(game)
	seen := map[string]bool{"": true, "nothing": true}
	for _, file := range files {
		seen[strings.TrimSuffix(file, ext)] = true
	}

	queue := slices.Clone(roots)
	slices.Sort(queue)
	for len(queue) > 0 {
		template := queue[0]
		queue = queue[1:]
		if seen[template] {
			continue
		}
		seen[template] = true

		text, err := extractTextFile(bundles, gameVersion, template+ext)
		// PoE2 names templates it does not ship (BloodClan); PoB-PoE2 skips them too (minions.lua:55-57).
		if errors.Is(err, fs.ErrNotExist) && (game == extractor.PoE2 || skipMissing) {
			slog.Warn("skipping a template the game does not have", slog.String("path", template+ext))
			continue
		}
		if err != nil {
			return err
		}
		for _, m := range extendsRegex.FindAllSubmatch(text, -1) {
			queue = append(queue, string(m[1]))
		}
	}

	return nil
}

// readRows decodes the rows of a table extractRawData wrote.
func readRows[T any](gameVersion string, table string) ([]T, error) {
	f, err := os.Open(filepath.Join("data", gameVersion, "raw", table+".json.gz"))
	if err != nil {
		return nil, errors.Wrap(err, "could not open "+table)
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, errors.Wrap(err, "could not decompress "+table)
	}

	var rows []T
	return rows, errors.Wrap(json.NewDecoder(r).Decode(&rows), "could not decode "+table)
}

// extractTextFile publishes file as UTF-8 and returns its text.
func extractTextFile(bundles fs.FS, gameVersion string, file string) ([]byte, error) {
	slog.Info("extracting", slog.String("path", file))

	f, err := openTextFile(bundles, file)
	if err != nil {
		return nil, errors.Wrap(err, "could not open "+file)
	}

	text, err := io.ReadAll(unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder().Reader(f))
	if err != nil {
		return nil, errors.Wrap(err, "could not decode "+file)
	}

	var buf bytes.Buffer
	w := brotli.NewWriter(&buf)
	if _, err := w.Write(text); err != nil {
		return nil, errors.Wrap(err, "could not compress "+file)
	}
	if err := w.Close(); err != nil {
		return nil, errors.Wrap(err, "could not compress "+file)
	}

	outPath := filepath.Join("data", gameVersion, loader.FilePath(file))
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return nil, errors.Wrap(err, "could not create directory tree")
	}
	if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
		return nil, errors.Wrap(err, "could not write "+outPath)
	}

	return text, nil
}

// openTextFile opens file, or under Metadata/StatDescriptions the same file with the other of .txt and .csd: 3.30
// ships its stat files as .csd where earlier versions had .txt. The file is published under the name asked for.
func openTextFile(bundles fs.FS, file string) (fs.File, error) {
	name := strings.ToLower(file)
	f, err := bundles.Open(name)
	if err == nil || !errors.Is(err, fs.ErrNotExist) || !strings.HasPrefix(name, "metadata/statdescriptions/") {
		return f, err //nolint:wrapcheck
	}
	ext := path.Ext(name)
	other := map[string]string{".txt": ".csd", ".csd": ".txt"}[ext]
	if other == "" {
		return nil, err //nolint:wrapcheck
	}
	f, otherErr := bundles.Open(strings.TrimSuffix(name, ext) + other)
	if otherErr != nil {
		return nil, err //nolint:wrapcheck
	}
	return f, nil
}
