package loader

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"maps"
	"slices"
	"strings"

	"github.com/Vilsol/slox"
	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
)

// EncodeFileBundle is files, game path to UTF-8 text, as a msgpack map with sorted keys.
func EncodeFileBundle(files map[string][]byte) []byte {
	b := msgp.AppendMapHeader(nil, uint32(len(files)))
	for _, path := range slices.Sorted(maps.Keys(files)) {
		b = msgp.AppendString(b, path)
		b = msgp.AppendBytes(b, files[path])
	}
	return b
}

// FileBundlePath is where the bundle of every published file with extension ext (".ot") lives under its
// version directory, named by the hash of its EncodeFileBundle bytes: "files.ot-1a2b3c4d.msgpack.br".
func FileBundlePath(ext string, encoded []byte) string {
	sum := sha256.Sum256(encoded)
	return AssetPath("files", strings.TrimPrefix(ext, ".")+"-"+hex.EncodeToString(sum[:4]))
}

// LoadFileBundle loads every game text file with extension ext (".ot", ".it", ".txt") that version
// publishes, by game path, in one asset. A version whose manifest names no such bundle is ErrAssetNotFound.
func LoadFileBundle(ctx context.Context, version string, ext string, assetCache AssetCache) (map[string][]byte, error) {
	manifest, err := LoadManifest(ctx, version)
	if err != nil {
		return nil, err
	}

	path, ok := manifest.FileBundles[ext]
	if !ok {
		return nil, errors.Wrap(ErrAssetNotFound, "no "+ext+" file bundle in "+version)
	}

	url := cdnBase + version + "/" + path

	slox.From(ctx).Debug(
		"loading file bundle",
		slog.String("version", version),
		slog.String("path", path),
	)

	b, fromCache, err := loadAsset(ctx, url, assetCache)
	if err != nil {
		return nil, err
	}

	files, err := decodeFileBundle(b)
	if err != nil && fromCache {
		slox.From(ctx).Warn(
			"discarding unreadable cached asset",
			slog.String("url", url),
			slog.Any("error", err),
		)

		b, err = fetchAsset(ctx, url, assetCache)
		if err != nil {
			return nil, err
		}

		files, err = decodeFileBundle(b)
	}

	return files, errors.Wrap(err, "failed to decode: "+url)
}

func decodeFileBundle(compressed []byte) (map[string][]byte, error) {
	b, err := io.ReadAll(brotli.NewReader(bytes.NewReader(compressed)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to decompress")
	}

	n, b, err := msgp.ReadMapHeaderBytes(b)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read bundle header")
	}

	files := make(map[string][]byte, n)
	for range n {
		var path string
		path, b, err = msgp.ReadStringBytes(b)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read bundle path")
		}

		var text []byte
		text, b, err = msgp.ReadBytesZC(b)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read "+path)
		}
		files[path] = text
	}

	if len(b) > 0 {
		return nil, errors.Errorf("%d trailing bytes", len(b))
	}

	return files, nil
}
