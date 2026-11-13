package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/oriath-net/pogo/poefs/ggpk"
	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/extractor"
)

const (
	bundleIndex = "Bundles2/_.index.bin"
	ggpkName    = "Content.ggpk"
)

// openSource opens cdn:<patch>, a directory with Bundles2/, a Content.ggpk file, or a directory holding one.
func openSource(game extractor.Game, gamePath string) (fs.FS, error) {
	if patch, ok := strings.CutPrefix(gamePath, cdnPrefix); ok {
		return extractor.NewWebFS(game, patch), nil
	}

	info, err := os.Stat(gamePath)
	if err != nil {
		return nil, errors.Wrap(err, "could not open game path")
	}

	if !info.IsDir() {
		return openGGPK(gamePath)
	}

	if _, err := os.Stat(filepath.Join(gamePath, filepath.FromSlash(bundleIndex))); err == nil {
		return os.DirFS(gamePath), nil
	}

	if _, err := os.Stat(filepath.Join(gamePath, ggpkName)); err == nil {
		return openGGPK(filepath.Join(gamePath, ggpkName))
	}

	return nil, fmt.Errorf("could not find %s or %s in %s", bundleIndex, ggpkName, gamePath)
}

// openGGPK keeps the file open for the life of the process; the loaders read it by random access.
func openGGPK(path string) (fs.FS, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "could not open ggpk")
	}

	ggfs, err := ggpk.NewLoader(f)
	if err != nil {
		_ = f.Close()
		return nil, errors.Wrapf(err, "could not read ggpk %s", path)
	}

	if !ggfs.UseBundles() {
		_ = f.Close()
		return nil, fmt.Errorf("%s predates bundles (3.11.2) and is not supported", path)
	}

	index, err := ggfs.Open(bundleIndex)
	if err != nil {
		_ = f.Close()
		return nil, errors.Wrapf(err, "%s has no %s", path, bundleIndex)
	}
	_ = index.Close()

	return ggfs, nil
}
