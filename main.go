package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/Vilsol/go-pob-data/extractor"

	_ "github.com/andybalholm/brotli"
)

//go:generate msgp -v -file ./raw
//go:generate msgp -v -file ./poe2
//go:generate msgp -v -file ./poe2/pobdata
//go:generate msgp -v -file ./derived

// cdnPrefix selects the patch CDN instead of a local install, e.g. cdn:3.29.3.3.
const cdnPrefix = "cdn:"

// skipMissing is -skip-missing.
var skipMissing bool

func main() {
	treeOnly := flag.Bool("tree-only", false, "only regenerate the passive tree")
	gameName := flag.String("game", string(extractor.PoE1), "game to extract: poe1 or poe2")
	latest := flag.Bool("latest", false, "print the game's live patch and exit")
	flag.BoolVar(&skipMissing, "skip-missing", false, "log and skip inputs the source lacks instead of failing: for a partial source such as a preload read through a sidecar index")
	tradeStats := flag.Bool("trade-stats", false, "only fetch the trade site's stats list into data/<game_version>; the site serves only the live league's")
	checkExeLayout := flag.Bool("check-exe", false, "only compare the tables the client executable (from cdn:<patch>) loads against data/<game_version>/schema.min.json; fails when an exported table disagrees")
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

	if *tradeStats {
		if flag.NArg() < 1 {
			slog.Error("please provide game version")
			os.Exit(1)
			return
		}
		if err := extractTradeStats(game, flag.Arg(0)); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
			return
		}
		if err := generateMeta(); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		return
	}

	if *checkExeLayout {
		if flag.NArg() < 2 {
			slog.Error("please provide cdn:<patch> and game version")
			os.Exit(1)
			return
		}
		if err := checkExe(game, flag.Arg(0), flag.Arg(1)); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		return
	}

	if flag.NArg() < 1 {
		slog.Error("please provide path to the game directory, a Content.ggpk or cdn:<patch>")
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
	source, err := openSource(game, gamePath)
	if err != nil {
		return err
	}

	extractor.LoadParser(game, gameVersion)
	loader, err := bundleLoader(source, game, gameVersion)
	if err != nil {
		return err
	}

	if !treeOnly {
		slog.Info("extracting raw data")
		if err := extractRawData(loader, game, gameVersion); err != nil {
			return err
		}
	}

	slog.Info("generating tree data")
	if err := generateTreeData(loader, game, gameVersion); err != nil {
		if !skipMissing {
			return err
		}
		slog.Error("tree not generated", slog.String("error", err.Error()))
	}

	if treeOnly {
		return nil
	}

	slog.Info("extracting translations")
	if err := extractTranslations(loader, game, gameVersion); err != nil {
		if !skipMissing {
			return err
		}
		slog.Error("translations not extracted", slog.String("error", err.Error()))
	}

	slog.Info("extracting text files")
	if err := extractTextFiles(loader, game, gameVersion); err != nil {
		return err
	}

	slog.Info("generating meta")
	return generateMeta()
}
