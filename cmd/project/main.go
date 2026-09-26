// Command project writes the .min assets of every asset in a version's field list, from the committed
// full tuple assets. With no arguments it projects every version that has a list.
//
//	go run ./cmd/project [version...]
package main

import (
	"bytes"
	"flag"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"

	"github.com/Vilsol/go-pob-data/fields"
	"github.com/Vilsol/go-pob-data/loader"
)

func main() {
	dataDir := flag.String("data", "data", "the data directory")
	flag.Parse()

	versions := flag.Args()
	if len(versions) == 0 {
		versions = fields.Versions()
	}

	for _, version := range versions {
		if err := project(*dataDir, version); err != nil {
			slog.Error("projection failed", slog.String("version", version), slog.Any("error", err))
			os.Exit(1)
		}
	}
}

func project(dataDir string, version string) error {
	list, ok := fields.For(version)
	if !ok {
		return errors.Errorf("no field list for %s", version)
	}

	for _, asset := range slices.Sorted(maps.Keys(list.Assets)) {
		t, _ := fields.RowType(asset)
		fullPath := filepath.Join(dataDir, version, loader.AssetPath(asset, loader.TupleVariant(reflect.New(t).Interface())))
		full, err := os.ReadFile(fullPath)
		if os.IsNotExist(err) {
			// The loader reads a table the version does not ship (no json) as empty, .min or not.
			if _, jsonErr := os.Stat(filepath.Join(dataDir, version, asset+".json.gz")); os.IsNotExist(jsonErr) && strings.HasPrefix(asset, "raw/") {
				slog.Warn("not in this version, the loader reads it as empty", slog.String("version", version), slog.String("asset", asset))
				continue
			}
			return errors.Errorf("%s is in the %s list, but %s does not exist", asset, version, fullPath)
		}
		if err != nil {
			return errors.Wrap(err, "failed to read full asset")
		}
		if bytes.HasPrefix(full, []byte("version https://git-lfs")) {
			return errors.Errorf("%s is a git lfs pointer, run git lfs pull", fullPath)
		}

		minAsset, err := fields.Project(asset, list.Assets[asset].Keep, full)
		if err != nil {
			return err
		}

		if strings.HasPrefix(asset, "raw/") {
			fullRows, minRows := rowCount(full), rowCount(minAsset)
			if fullRows != minRows {
				return errors.Errorf("%s: the .min has %d rows, the full asset %d", asset, minRows, fullRows)
			}
		}

		variant, _ := list.MinVariant(asset)
		minPath := filepath.Join(dataDir, version, loader.AssetPath(asset, variant))
		if err := os.WriteFile(minPath, minAsset, 0o644); err != nil {
			return errors.Wrap(err, "failed to write .min asset")
		}

		slog.Info("projected", slog.String("path", minPath), slog.Int("full", len(full)), slog.Int("min", len(minAsset)))
	}

	return nil
}

// rowCount is the row count of a raw asset, -1 if it has none.
func rowCount(asset []byte) int64 {
	n, err := msgp.NewReader(brotli.NewReader(bytes.NewReader(asset))).ReadInt64()
	if err != nil {
		return -1
	}
	return n
}
