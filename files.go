package main

import (
	"bytes"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
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

func extractTextFiles(bundles fs.FS, gameVersion string) error {
	for _, file := range textFilesToExport {
		slog.Info("extracting", slog.String("path", file))

		f, err := bundles.Open(strings.ToLower(file))
		if err != nil {
			return errors.Wrap(err, "could not open "+file)
		}

		text, err := io.ReadAll(unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder().Reader(f))
		if err != nil {
			return errors.Wrap(err, "could not decode "+file)
		}

		var buf bytes.Buffer
		w := brotli.NewWriter(&buf)
		if _, err := w.Write(text); err != nil {
			return errors.Wrap(err, "could not compress "+file)
		}
		if err := w.Close(); err != nil {
			return errors.Wrap(err, "could not compress "+file)
		}

		outPath := filepath.Join("data", gameVersion, loader.FilePath(file))
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return errors.Wrap(err, "could not create directory tree")
		}
		if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
			return errors.Wrap(err, "could not write "+outPath)
		}
	}

	return nil
}
