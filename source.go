package main

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/oriath-net/pogo/poefs/ggpk"
	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/sidecar"
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
	if err != nil && os.Getenv(sidecarEnv) != "" {
		return ggfs, nil
	}
	if err != nil {
		_ = f.Close()
		return nil, errors.Wrapf(err, "%s has no %s (a preload ships its bundles without the index until launch)", path, bundleIndex)
	}
	_ = index.Close()

	return ggfs, nil
}

// sidecarEnv names a sidecar index (cmd/sidecar) to serve the bundles by when the source has no bundle index;
// sidecarMinEnv is the lowest confidence served, medium by default.
const (
	sidecarEnv    = "GOPOB_SIDECAR"
	sidecarMinEnv = "GOPOB_SIDECAR_MIN"
)

func bundleLoader(source fs.FS, game extractor.Game, gameVersion string) (fs.FS, error) {
	loader, err := sidecarOrIndexLoader(source, game, gameVersion)
	if err != nil {
		return nil, err
	}
	if logPath := os.Getenv(openLogEnv); logPath != "" {
		f, err := os.Create(logPath)
		if err != nil {
			return nil, errors.Wrap(err, "could not create the open log")
		}
		return &openLog{FS: loader, w: f}, nil
	}
	return loader, nil
}

// openLogEnv names a file to record every path the extractor opens, and whether it opened: the input list a
// sidecar index has to cover.
const openLogEnv = "GOPOB_OPENLOG"

type openLog struct {
	fs.FS
	w  *os.File
	mu sync.Mutex
}

func (l *openLog) Open(name string) (fs.File, error) {
	f, err := l.FS.Open(name)
	l.mu.Lock()
	defer l.mu.Unlock()
	status := "ok"
	if err != nil {
		status = "missing"
	}
	fmt.Fprintf(l.w, "%s\t%s\n", status, name)
	return f, err //nolint:wrapcheck // a pass-through fs.FS returns the lower FS's error as is
}

func sidecarOrIndexLoader(source fs.FS, game extractor.Game, gameVersion string) (fs.FS, error) {
	path := os.Getenv(sidecarEnv)
	if path == "" {
		loader, err := extractor.GetBundleLoader(source, IsNewHash(game, gameVersion))
		return loader, errors.Wrap(err, "could not initialize bundle loader")
	}
	minConf := sidecar.Medium
	if v := os.Getenv(sidecarMinEnv); v != "" {
		c, err := sidecar.ParseConfidence(v)
		if err != nil {
			return nil, err
		}
		minConf = c
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "could not open sidecar index")
	}
	defer f.Close()
	entries, err := sidecar.ReadEntries(f)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read sidecar index %s", path)
	}
	slog.Warn("serving bundles by a sidecar index, not the game's", "sidecar", path, "min", minConf.String())
	return sidecar.NewFS(source, entries, minConf), nil
}
