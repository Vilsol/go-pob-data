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
// monsters read their base stats. They are published only in one bundle per extension.
func extractTextFiles(bundles fs.FS, game extractor.Game, gameVersion string) error {
	files, varieties := textFiles(game)
	texts := map[string][]byte{}
	for _, file := range files {
		text, err := extractTextFile(bundles, file)
		if err != nil {
			if skipMissing && errors.Is(err, fs.ErrNotExist) {
				slog.Warn("skipping a missing file", slog.String("path", file))
				continue
			}
			return err
		}
		texts[file] = text
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
	if err := extractTemplates(bundles, game, itemRoots, ".it", texts); err != nil {
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
	if err := extractTemplates(bundles, game, monsterRoots, ".ot", texts); err != nil {
		return err
	}

	return writeFileBundles(filepath.Join("data", gameVersion), texts)
}

// fileBundle is the LoadFileBundle asset of files, game path to UTF-8 text, at its loader.FileBundlePath.
func fileBundle(files map[string][]byte, ext string) (string, []byte, error) {
	encoded := loader.EncodeFileBundle(files)
	var buf bytes.Buffer
	w := brotli.NewWriter(&buf)
	if _, err := w.Write(encoded); err != nil {
		return "", nil, errors.Wrap(err, "could not compress the "+ext+" bundle")
	}
	if err := w.Close(); err != nil {
		return "", nil, errors.Wrap(err, "could not compress the "+ext+" bundle")
	}
	return loader.FileBundlePath(ext, encoded), buf.Bytes(), nil
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

// writeFileBundles writes the bundle of files' every extension to versionDir and removes the bundles they
// replace, and the copies of files an earlier extraction published one at a time.
func writeFileBundles(versionDir string, files map[string][]byte) error {
	byExt := map[string]map[string][]byte{}
	for file, text := range files {
		ext := path.Ext(file)
		if byExt[ext] == nil {
			byExt[ext] = map[string][]byte{}
		}
		byExt[ext][file] = text
	}

	for ext, group := range byExt {
		bundlePath, b, err := fileBundle(group, ext)
		if err != nil {
			return err
		}
		stale, err := staleFileBundles(versionDir, ext, bundlePath)
		if err != nil {
			return err
		}
		for file := range group {
			stale = append(stale, filepath.Join(versionDir, loader.FilePath(file)))
		}
		for _, old := range stale {
			if err := os.Remove(old); err != nil && !os.IsNotExist(err) {
				return errors.Wrap(err, "could not remove "+old)
			}
		}
		if err := os.WriteFile(filepath.Join(versionDir, bundlePath), b, 0o644); err != nil {
			return errors.Wrap(err, "could not write the "+ext+" bundle")
		}
	}

	return nil
}

// extractTemplates publishes the templates of roots and of their ancestors, up to "nothing" or a template
// textFiles already publishes (Metadata/Monsters/Monster), into texts.
func extractTemplates(bundles fs.FS, game extractor.Game, roots []string, ext string, texts map[string][]byte) error {
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

		text, err := extractTextFile(bundles, template+ext)
		// PoE2 names templates it does not ship (BloodClan); PoB-PoE2 skips them too (minions.lua:55-57).
		if errors.Is(err, fs.ErrNotExist) && (game == extractor.PoE2 || skipMissing) {
			slog.Warn("skipping a template the game does not have", slog.String("path", template+ext))
			continue
		}
		if err != nil {
			return err
		}
		texts[template+ext] = text
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

// extractTextFile reads file as UTF-8.
func extractTextFile(bundles fs.FS, file string) ([]byte, error) {
	slog.Info("extracting", slog.String("path", file))

	f, err := openTextFile(bundles, file)
	if err != nil {
		return nil, errors.Wrap(err, "could not open "+file)
	}

	text, err := io.ReadAll(unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder().Reader(f))
	return text, errors.Wrap(err, "could not decode "+file)
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
