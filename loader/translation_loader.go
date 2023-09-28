package loader

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Vilsol/go-pob-data/raw"
	"github.com/Vilsol/slox"
	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
	"io"
	"log/slog"
	"net/http"
)

const cdnTranslationBase = "https://go-pob-data.pages.dev/data/%s/stat_translations/%s/%s.msgpack.br"

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

	var b []byte
	if assetCache.Exists(url) {
		get, err := assetCache.Get(url)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve url from cache: "+url)
		}
		b = get
	} else {
		slox.From(ctx).Debug(
			"fetching",
			slog.String("url", url),
		)

		request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create request")
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch url: "+url)
		}
		defer response.Body.Close()

		b, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read response body")
		}

		defer func() {
			_ = assetCache.Set(url, b)
		}()
	}

	unzipStream := brotli.NewReader(bytes.NewReader(b))

	r := msgp.NewReader(unzipStream)

	var out raw.TranslationFile
	if err := out.DecodeMsg(r); err != nil {
		return nil, errors.Wrap(err, "failed to read message")
	}

	return &out, nil
}
