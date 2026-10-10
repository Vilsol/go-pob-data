package loader

import (
	"bytes"
	"context"
	"io"
	"log/slog"

	"github.com/Vilsol/slox"
	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
)

// FilePath is where a game text file such as "Metadata/Characters/Character.ot" lives under its version
// directory.
func FilePath(path string) string {
	return "files/" + path + ".br"
}

// FileURL is the URL LoadFile reads path of version from.
func FileURL(version string, path string) string {
	return cdnBase + version + "/" + FilePath(path)
}

// LoadFile loads a game text file, such as "Metadata/Monsters/Monster.ot" or "art/uiimages1.txt", as
// UTF-8 text. The game ships them as UTF-16.
//
// Returns data from cache if found
func LoadFile(ctx context.Context, version string, path string, assetCache AssetCache) ([]byte, error) {
	url := FileURL(version, path)

	slox.From(ctx).Debug(
		"loading file asset",
		slog.String("version", version),
		slog.String("path", path),
	)

	b, fromCache, err := loadAsset(ctx, url, assetCache)
	if err != nil {
		return nil, err
	}

	text, err := io.ReadAll(brotli.NewReader(bytes.NewReader(b)))
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

		text, err = io.ReadAll(brotli.NewReader(bytes.NewReader(b)))
	}

	return text, errors.Wrap(err, "failed to decompress: "+url)
}
