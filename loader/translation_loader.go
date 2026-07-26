package loader

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/Vilsol/slox"
	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"

	"github.com/Vilsol/go-pob-data/raw"
)

const cdnTranslationBase = "https://data.pob.gg/data/%s/stat_translations/%s/%s.msgpack.br"

// LoadTranslation loads a raw brotli-compressed json dump from remote source
//
// Returns data from cache if found
func LoadTranslation(ctx context.Context, version string, language string, name string, assetCache AssetCache) (*raw.TranslationFile, error) {
	url := fmt.Sprintf(cdnTranslationBase, version, language, name)

	slox.From(ctx).Debug(
		"loading translation asset",
		slog.String("version", version),
		slog.String("name", name),
		slog.String("language", language),
	)

	b, fromCache, err := loadAsset(ctx, url, assetCache)
	if err != nil {
		return nil, err
	}

	var out raw.TranslationFile

	err = out.DecodeMsg(msgp.NewReader(brotli.NewReader(bytes.NewReader(b))))
	if err != nil && fromCache {
		// Nothing is returned to the caller until the whole file decodes, so a
		// cached asset that fails to decode is safe to throw away and fetch again.
		slox.From(ctx).Warn(
			"discarding unreadable cached asset",
			slog.String("url", url),
			slog.Any("error", err),
		)

		b, err = fetchAsset(ctx, url, assetCache)
		if err != nil {
			return nil, err
		}

		out = raw.TranslationFile{}
		err = out.DecodeMsg(msgp.NewReader(brotli.NewReader(bytes.NewReader(b))))
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to read message: "+url)
	}

	return &out, nil
}
