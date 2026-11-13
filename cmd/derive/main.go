// Command derive writes the derived package's assets of every PoE1 version from its committed tables and
// stat descriptions, replacing each table's asset of the same layout. With no arguments it derives every
// version under the data directory. Regenerate the manifests afterwards.
//
//	go run ./cmd/derive [version...]
package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/derived"
	"github.com/Vilsol/go-pob-data/poe2/pobdata"
)

func main() {
	dataDir := flag.String("data", "data", "the data directory")
	flag.Parse()

	versions := flag.Args()
	if len(versions) == 0 {
		dirs, err := os.ReadDir(*dataDir)
		if err != nil {
			slog.Error("failed listing versions", slog.Any("error", err))
			os.Exit(1)
		}
		for _, dir := range dirs {
			// 0.x is PoE2, whose tables the poe package does not read.
			if dir.IsDir() && strings.HasPrefix(dir.Name(), "3.") {
				versions = append(versions, dir.Name())
			}
		}
	}

	abs, err := filepath.Abs(*dataDir)
	if err != nil {
		slog.Error("failed resolving the data directory", slog.Any("error", err))
		os.Exit(1)
	}
	http.DefaultClient.Transport = localCDN{files: http.NewFileTransport(http.Dir(abs))}

	for _, version := range versions {
		if err := derive(*dataDir, version); err != nil {
			slog.Error("derivation failed", slog.String("version", version), slog.Any("error", err))
			os.Exit(1)
		}
	}
}

func derive(dataDir string, version string) error {
	assets, err := derived.Build(context.Background(), version, noCache{})
	if err != nil {
		return err //nolint:wrapcheck
	}
	if len(assets) == 0 {
		slog.Warn("no table's inputs are published", slog.String("version", version))
	}

	for _, asset := range assets {
		path := filepath.Join(dataDir, version, filepath.FromSlash(asset.Path))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return errors.Wrap(err, "failed creating the derived directory")
		}

		// The previous asset of this table and layout is replaced, not kept beside it.
		prefix := asset.Path[:strings.LastIndex(asset.Path, "-")]
		stale, err := filepath.Glob(filepath.Join(dataDir, version, filepath.FromSlash(prefix)) + "-*.msgpack.br")
		if err != nil {
			return errors.Wrap(err, "failed listing previous assets")
		}
		for _, old := range stale {
			if old != path {
				if err := os.Remove(old); err != nil {
					return errors.Wrap(err, "failed removing a previous asset")
				}
			}
		}

		compressed, err := pobdata.Compress(asset.Encoded)
		if err != nil {
			return err //nolint:wrapcheck
		}
		if err := os.WriteFile(path, compressed, 0o644); err != nil { //nolint:gosec // published data
			return errors.Wrap(err, "failed writing asset")
		}
		slog.Info("derived", slog.String("path", path), slog.Int("rows", asset.Rows))
	}
	return nil
}

type noCache struct{}

func (noCache) Get(string) ([]byte, error) { return nil, nil }
func (noCache) Set(string, []byte) error   { return nil }
func (noCache) Exists(string) bool         { return false }

const cdnBase = "https://data.pob.gg/data/"

// localCDN serves data.pob.gg from the data directory.
type localCDN struct {
	files http.RoundTripper
}

func (c localCDN) RoundTrip(req *http.Request) (*http.Response, error) {
	path, ok := strings.CutPrefix(req.URL.String(), cdnBase)
	if !ok {
		return nil, errors.Errorf("%s is not on the data CDN", req.URL)
	}
	local := req.Clone(req.Context())
	local.URL = &url.URL{Scheme: "file", Path: "/" + path}
	return c.files.RoundTrip(local) //nolint:wrapcheck
}
