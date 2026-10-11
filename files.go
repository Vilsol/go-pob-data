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

// extractTextFiles publishes textFilesToExport and the item templates (.it) bases inherit from: every
// base's InheritsFrom and, through their extends lines, its ancestors. Bases read their tags from them.
func extractTextFiles(bundles fs.FS, gameVersion string) error {
	for _, file := range textFilesToExport {
		if _, err := extractTextFile(bundles, gameVersion, file); err != nil {
			return err
		}
	}

	queue, err := baseInheritsFrom(gameVersion)
	if err != nil {
		return err
	}

	seen := map[string]bool{"nothing": true}
	for len(queue) > 0 {
		template := queue[0]
		queue = queue[1:]
		if seen[template] {
			continue
		}
		seen[template] = true

		text, err := extractTextFile(bundles, gameVersion, template+".it")
		if err != nil {
			return err
		}
		for _, m := range extendsRegex.FindAllSubmatch(text, -1) {
			queue = append(queue, string(m[1]))
		}
	}

	return nil
}

// baseInheritsFrom reads the InheritsFrom of every base from the BaseItemTypes extractRawData wrote.
func baseInheritsFrom(gameVersion string) ([]string, error) {
	f, err := os.Open(filepath.Join("data", gameVersion, "raw", "BaseItemTypes.json.gz"))
	if err != nil {
		return nil, errors.Wrap(err, "could not open BaseItemTypes")
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, errors.Wrap(err, "could not decompress BaseItemTypes")
	}

	var bases []struct {
		InheritsFrom string `json:"InheritsFrom"`
	}
	if err := json.NewDecoder(r).Decode(&bases); err != nil {
		return nil, errors.Wrap(err, "could not decode BaseItemTypes")
	}

	out := make([]string, 0, len(bases))
	for _, base := range bases {
		if base.InheritsFrom != "" {
			out = append(out, base.InheritsFrom)
		}
	}
	slices.Sort(out)
	return slices.Compact(out), nil
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
