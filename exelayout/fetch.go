package exelayout

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/extractor"
)

// exeName is the client on the patch CDN; the small PathOfExile_x64.exe beside it is only a launcher.
const exeName = "PathOfExile.exe"

// cacheName is the layout cached beside the patch's downloaded bundles.
const cacheName = "exe-layout.json"

// Load returns patch's layout, from the bundle cache when an earlier run analyzed it, else by downloading
// the client from the patch CDN (kept in the bundle cache too) and analyzing it.
func Load(game extractor.Game, patch string) (*Layout, error) {
	web := extractor.NewWebFS(game, patch)
	cachePath := ""
	if web.CacheDir != "" {
		cachePath = filepath.Join(web.CacheDir, cacheName)
		if l, err := readCached(cachePath); err == nil {
			return l, nil
		}
	}

	exe, err := fetch(web)
	if err != nil {
		return nil, err
	}
	l, err := Analyze(bytes.NewReader(exe))
	if err != nil {
		return nil, err
	}
	l.Patch = patch

	if cachePath != "" {
		if err := writeCached(cachePath, l); err != nil {
			slog.Warn("could not cache the exe layout", slog.String("path", cachePath), slog.Any("err", err))
		}
	}
	return l, nil
}

func fetch(web *extractor.WebFS) ([]byte, error) {
	f, err := web.Open(exeName)
	if err != nil {
		return nil, errors.Wrap(err, "opening the client")
	}
	wf, ok := f.(*extractor.WebFile)
	if !ok {
		return nil, fmt.Errorf("unexpected %T", f)
	}
	return readAll(wf)
}

// readAll loads wf, turning extractor.Fetch's panic on a failed download into an error.
func readAll(wf *extractor.WebFile) ([]byte, error) {
	var failed any
	func() {
		defer func() { failed = recover() }()
		_, _ = wf.Read(nil)
	}()
	if failed != nil {
		return nil, fmt.Errorf("downloading %s%s: %v", wf.BasePath, wf.Name, failed)
	}
	return wf.CachedFile, nil
}

func readCached(path string) (*Layout, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "reading cached layout")
	}
	var l Layout
	if err := json.Unmarshal(b, &l); err != nil {
		return nil, errors.Wrap(err, "decoding cached layout")
	}
	if l.Version != formatVersion {
		return nil, errors.New("cached layout is from an older analysis")
	}
	return &l, nil
}

func writeCached(path string, l *Layout) error {
	b, err := json.Marshal(l)
	if err != nil {
		return errors.Wrap(err, "encoding layout")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return errors.Wrap(err, "creating cache dir")
	}
	return errors.Wrap(os.WriteFile(path, b, 0o644), "writing layout")
}
