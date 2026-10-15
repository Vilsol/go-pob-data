package main

import (
	"flag"
	"fmt"
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
	gameName := flag.String("game", string(extractor.PoE1), "game to extract: poe1 or poe2")
	latest := flag.Bool("latest", false, "print the game's live patch and exit")
	flag.Parse()

	game, err := extractor.ParseGame(*gameName)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
		return
	}

	if *latest {
		patch, err := extractor.LatestPatch(game)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
			return
		}
		fmt.Println(patch)
		return
	}

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

	err = run(game, flag.Arg(0), flag.Arg(1), *treeOnly)
	// ImageMagick can't be initialized again once terminated, so it's only terminated here.
	imagick.Terminate()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
		return
	}
}

func run(game extractor.Game, gamePath string, gameVersion string, treeOnly bool) error {
	var source fs.FS
	if patch, ok := strings.CutPrefix(gamePath, cdnPrefix); ok {
		source = extractor.NewWebFS(game, patch)
	} else {
		if _, err := os.Stat(filepath.Join(gamePath, "Bundles2", "_.index.bin")); err != nil {
			return errors.Wrap(err, "could not find _.index.bin")
		}
		source = os.DirFS(gamePath)
	}

	extractor.LoadParser(game, gameVersion)
	loader, err := extractor.GetBundleLoader(source, IsNewHash(game, gameVersion))
	if err != nil {
		return errors.Wrap(err, "could not initialize bundle loader")
	}

	if !treeOnly {
		slog.Info("extracting raw data")
		if err := extractRawData(loader, game, gameVersion); err != nil {
			return err
		}
	}

	// treegen reads the PoE1 passive skill graph and tables only.
	if game == extractor.PoE1 {
		slog.Info("generating tree data")
		if err := generateTreeData(loader, gameVersion); err != nil {
			return err
		}
	} else if treeOnly {
		return errors.New("no tree generator for " + string(game))
	}

	if treeOnly {
		return nil
	}

	slog.Info("extracting translations")
	if err := extractTranslations(loader, game, gameVersion); err != nil {
		return err
	}

	slog.Info("extracting text files")
	if err := extractTextFiles(loader, game, gameVersion); err != nil {
		return err
	}

	slog.Info("generating meta")
	return generateMeta()
}
