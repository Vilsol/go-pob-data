package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"golang.org/x/text/encoding/unicode"

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

var extendsRegex = regexp.MustCompile(`(?m)^extends "([^"]+)"`)

// monsterVarieties are the monsters PoB's minion and spectre export templates name, one MonsterVarieties
// Id per line (mise run monster-varieties); their .ot templates are published.
//
//go:embed monster_varieties.txt
var monsterVarieties string

// extractTextFiles publishes textFilesToExport and the templates rows inherit from, with their ancestors
// through their extends lines: the item templates (.it) of every base's InheritsFrom, where bases read
// their tags, and the monster templates (.ot) of monsterVarieties' BaseMonsterTypeIndex, where those
// monsters read their base stats.
func extractTextFiles(bundles fs.FS, gameVersion string) error {
	for _, file := range textFilesToExport {
		if _, err := extractTextFile(bundles, gameVersion, file); err != nil {
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
	if err := extractTemplates(bundles, gameVersion, itemRoots, ".it"); err != nil {
		return err
	}

	varieties, err := readRows[struct {
		ID                   string `json:"Id"`
		BaseMonsterTypeIndex string `json:"BaseMonsterTypeIndex"`
	}](gameVersion, "MonsterVarieties")
	if err != nil {
		return err
	}
	objectTypes := make(map[string]string, len(varieties))
	for _, variety := range varieties {
		objectTypes[variety.ID] = variety.BaseMonsterTypeIndex
	}
	monsterRoots := make([]string, 0)
	for _, id := range strings.Fields(monsterVarieties) {
		objectType, ok := objectTypes[id]
		if !ok {
			return errors.New("monster_varieties.txt names a variety the game does not have: " + id)
		}
		monsterRoots = append(monsterRoots, objectType)
	}
	return extractTemplates(bundles, gameVersion, monsterRoots, ".ot")
}

// extractTemplates publishes the templates of roots and of their ancestors, up to "nothing" or a template
// textFilesToExport already publishes (Metadata/Monsters/Monster).
func extractTemplates(bundles fs.FS, gameVersion string, roots []string, ext string) error {
	seen := map[string]bool{"": true, "nothing": true}
	for _, file := range textFilesToExport {
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

	f, err := bundles.Open(strings.ToLower(file))
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
