package main

import (
	"flag"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/Vilsol/go-pob-data/extractor"

	_ "github.com/andybalholm/brotli"
)

//go:generate msgp -v -file ./raw

// cdnPrefix selects the patch CDN instead of a local install, e.g. cdn:3.29.3.3.
const cdnPrefix = "cdn:"

func main() {
	treeOnly := flag.Bool("tree-only", false, "only regenerate the passive tree")
	flag.Parse()

	if flag.NArg() < 1 {
		slog.Error("please provide path to the game directory or cdn:<patch>")
		os.Exit(1)
		return
	}

	if flag.NArg() < 2 {
		slog.Error("please provide game version")
		os.Exit(1)
		return
	}

	err := run(flag.Arg(0), flag.Arg(1), *treeOnly)
	// ImageMagick can't be initialized again once terminated, so it's only terminated here.
	imagick.Terminate()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
		return
	}
}

func run(gamePath string, gameVersion string, treeOnly bool) error {
	var source fs.FS
	if patch, ok := strings.CutPrefix(gamePath, cdnPrefix); ok {
		source = extractor.NewWebFS(patch)
	} else {
		if _, err := os.Stat(filepath.Join(gamePath, "Bundles2", "_.index.bin")); err != nil {
			return errors.Wrap(err, "could not find _.index.bin")
		}
		source = os.DirFS(gamePath)
	}

	extractor.LoadParser(gameVersion)
	loader, err := extractor.GetBundleLoader(source, IsNewHash(gameVersion))
	if err != nil {
		return errors.Wrap(err, "could not initialize bundle loader")
	}

	if !treeOnly {
		slog.Info("extracting raw data")
		if err := extractRawData(loader, gameVersion); err != nil {
			return err
		}
	}

	slog.Info("generating tree data")
	if err := generateTreeData(loader, gameVersion); err != nil {
		return err
	}

	if treeOnly {
		return nil
	}

	slog.Info("extracting translations")
	if err := extractTranslations(loader, gameVersion); err != nil {
		return err
	}

	slog.Info("extracting text files")
	if err := extractTextFiles(loader, gameVersion); err != nil {
		return err
	}

	slog.Info("generating meta")
	return generateMeta()
}
