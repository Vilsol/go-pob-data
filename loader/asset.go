package loader

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/Vilsol/slox"
	"github.com/pkg/errors"
)

// ErrAssetNotFound is returned when the CDN has no such asset. The set of tables the
// game ships changes between versions, so this is not fatal for every caller.
var ErrAssetNotFound = errors.New("asset not found")

// loadAsset returns the bytes of url, preferring the cache over the network.
//
// A cache entry that exists but holds no bytes is treated as a miss: it can only
// have come from a writer that died mid-write, and decoding it would otherwise
// look like a legitimately empty asset.
func loadAsset(ctx context.Context, url string, assetCache AssetCache) ([]byte, bool, error) {
	if assetCache.Exists(url) {
		b, err := assetCache.Get(url)
		if err != nil {
			return nil, false, errors.Wrap(err, "failed to retrieve url from cache: "+url)
		}

		if len(b) > 0 {
			return b, true, nil
		}

		slox.From(ctx).Warn(
			"discarding empty cached asset",
			slog.String("url", url),
		)
	}

	b, err := fetchAsset(ctx, url, assetCache)
	if err != nil {
		return nil, false, err
	}

	return b, false, nil
}

// fetchAsset downloads url and stores it in the cache, replacing any entry
// already held for that key.
func fetchAsset(ctx context.Context, url string, assetCache AssetCache) ([]byte, error) {
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

	// An error page is not an asset. Returning before the read keeps it out of the
	// cache, where it would decode as an empty table on every subsequent run.
	if response.StatusCode == http.StatusNotFound {
		return nil, errors.Wrap(ErrAssetNotFound, url)
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status %d fetching url: %s", response.StatusCode, url)
	}

	b, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	// Caching is best-effort: the caller already has valid bytes either way.
	if err := assetCache.Set(url, b); err != nil {
		slox.From(ctx).Warn(
			"failed to store asset in cache",
			slog.String("url", url),
			slog.Any("error", err),
		)
	}

	return b, nil
}

// assetExists asks the CDN whether url exists, without downloading it.
func assetExists(ctx context.Context, url string) (bool, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return false, errors.Wrap(err, "failed to create request")
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return false, errors.Wrap(err, "failed to fetch url: "+url)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	}

	return false, errors.Errorf("unexpected status %d probing url: %s", response.StatusCode, url)
}
